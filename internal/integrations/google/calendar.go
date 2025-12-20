package google

import (
	"context"
	"fmt"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	calendarScope = "https://www.googleapis.com/auth/calendar"
)

// CalendarCreateAction implements the Google Calendar Create Event action
type CalendarCreateAction struct {
	credentialService credential.Service
	baseURL           string
}

// CalendarCreateConfig defines the configuration for creating an event
type CalendarCreateConfig struct {
	CalendarID  string   `json:"calendar_id"` // Usually "primary"
	Summary     string   `json:"summary"`
	Description string   `json:"description,omitempty"`
	Location    string   `json:"location,omitempty"`
	StartTime   string   `json:"start_time"` // RFC3339 format
	EndTime     string   `json:"end_time"`   // RFC3339 format
	TimeZone    string   `json:"time_zone,omitempty"`
	Attendees   []string `json:"attendees,omitempty"` // Email addresses
}

// CalendarEventResult represents a calendar event
type CalendarEventResult struct {
	EventID     string   `json:"event_id"`
	Summary     string   `json:"summary"`
	StartTime   string   `json:"start_time"`
	EndTime     string   `json:"end_time"`
	HtmlLink    string   `json:"html_link"`
	Attendees   []string `json:"attendees,omitempty"`
}

// Validate validates the Calendar create configuration
func (c *CalendarCreateConfig) Validate() error {
	if c.CalendarID == "" {
		return fmt.Errorf("calendar_id is required")
	}
	if c.Summary == "" {
		return fmt.Errorf("summary is required")
	}
	if c.StartTime == "" {
		return fmt.Errorf("start_time is required")
	}
	if c.EndTime == "" {
		return fmt.Errorf("end_time is required")
	}
	return nil
}

// NewCalendarCreateAction creates a new Calendar create action
func NewCalendarCreateAction(credentialService credential.Service) *CalendarCreateAction {
	return &CalendarCreateAction{
		credentialService: credentialService,
	}
}

// Execute implements the Action interface
func (a *CalendarCreateAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	config, ok := input.Config.(CalendarCreateConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected CalendarCreateConfig")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	tenantID, err := extractString(input.Context, "env.tenant_id")
	if err != nil {
		return nil, fmt.Errorf("tenant_id is required in context: %w", err)
	}

	credentialID, err := extractString(input.Context, "credential_id")
	if err != nil {
		return nil, fmt.Errorf("credential_id is required in context: %w", err)
	}

	decryptedCred, err := a.credentialService.GetValue(ctx, tenantID, credentialID, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	token, err := createOAuth2Token(decryptedCred.Value)
	if err != nil {
		return nil, err
	}

	var calendarService *calendar.Service
	if a.baseURL != "" {
		calendarService, err = calendar.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(a.baseURL))
	} else {
		calendarService, err = calendar.NewService(ctx, createOAuth2Client(ctx, token))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Calendar service: %w", err)
	}

	// Create event
	event := &calendar.Event{
		Summary:     config.Summary,
		Description: config.Description,
		Location:    config.Location,
		Start: &calendar.EventDateTime{
			DateTime: config.StartTime,
			TimeZone: config.TimeZone,
		},
		End: &calendar.EventDateTime{
			DateTime: config.EndTime,
			TimeZone: config.TimeZone,
		},
	}

	// Add attendees
	if len(config.Attendees) > 0 {
		attendees := make([]*calendar.EventAttendee, len(config.Attendees))
		for i, email := range config.Attendees {
			attendees[i] = &calendar.EventAttendee{
				Email: email,
			}
		}
		event.Attendees = attendees
	}

	createdEvent, err := calendarService.Events.Insert(config.CalendarID, event).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	// Build result
	attendeeEmails := make([]string, len(createdEvent.Attendees))
	for i, a := range createdEvent.Attendees {
		attendeeEmails[i] = a.Email
	}

	result := &CalendarEventResult{
		EventID:   createdEvent.Id,
		Summary:   createdEvent.Summary,
		StartTime: createdEvent.Start.DateTime,
		EndTime:   createdEvent.End.DateTime,
		HtmlLink:  createdEvent.HtmlLink,
		Attendees: attendeeEmails,
	}

	output := actions.NewActionOutput(result)
	output.WithMetadata("event_id", createdEvent.Id)
	output.WithMetadata("html_link", createdEvent.HtmlLink)

	return output, nil
}

// CalendarListAction implements the Google Calendar List Events action
type CalendarListAction struct {
	credentialService credential.Service
	baseURL           string
}

// CalendarListConfig defines the configuration for listing events
type CalendarListConfig struct {
	CalendarID string `json:"calendar_id"` // Usually "primary"
	TimeMin    string `json:"time_min,omitempty"` // RFC3339 format
	TimeMax    string `json:"time_max,omitempty"` // RFC3339 format
	MaxResults int64  `json:"max_results,omitempty"` // Default 10
}

// CalendarListResult represents the result of listing events
type CalendarListResult struct {
	Events []CalendarEventResult `json:"events"`
	Count  int                   `json:"count"`
}

// Validate validates the Calendar list configuration
func (c *CalendarListConfig) Validate() error {
	if c.CalendarID == "" {
		return fmt.Errorf("calendar_id is required")
	}
	return nil
}

// NewCalendarListAction creates a new Calendar list action
func NewCalendarListAction(credentialService credential.Service) *CalendarListAction {
	return &CalendarListAction{
		credentialService: credentialService,
	}
}

// Execute implements the Action interface
func (a *CalendarListAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	config, ok := input.Config.(CalendarListConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected CalendarListConfig")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	if config.MaxResults == 0 {
		config.MaxResults = 10
	}

	tenantID, err := extractString(input.Context, "env.tenant_id")
	if err != nil {
		return nil, fmt.Errorf("tenant_id is required in context: %w", err)
	}

	credentialID, err := extractString(input.Context, "credential_id")
	if err != nil {
		return nil, fmt.Errorf("credential_id is required in context: %w", err)
	}

	decryptedCred, err := a.credentialService.GetValue(ctx, tenantID, credentialID, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	token, err := createOAuth2Token(decryptedCred.Value)
	if err != nil {
		return nil, err
	}

	var calendarService *calendar.Service
	if a.baseURL != "" {
		calendarService, err = calendar.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(a.baseURL))
	} else {
		calendarService, err = calendar.NewService(ctx, createOAuth2Client(ctx, token))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Calendar service: %w", err)
	}

	// Build list call
	listCall := calendarService.Events.List(config.CalendarID).
		MaxResults(config.MaxResults).
		SingleEvents(true).
		OrderBy("startTime")

	if config.TimeMin != "" {
		listCall = listCall.TimeMin(config.TimeMin)
	}
	if config.TimeMax != "" {
		listCall = listCall.TimeMax(config.TimeMax)
	}

	events, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	// Convert to result format
	results := make([]CalendarEventResult, 0, len(events.Items))
	for _, item := range events.Items {
		attendeeEmails := make([]string, len(item.Attendees))
		for i, a := range item.Attendees {
			attendeeEmails[i] = a.Email
		}

		startTime := item.Start.DateTime
		if startTime == "" {
			startTime = item.Start.Date
		}
		endTime := item.End.DateTime
		if endTime == "" {
			endTime = item.End.Date
		}

		results = append(results, CalendarEventResult{
			EventID:   item.Id,
			Summary:   item.Summary,
			StartTime: startTime,
			EndTime:   endTime,
			HtmlLink:  item.HtmlLink,
			Attendees: attendeeEmails,
		})
	}

	result := &CalendarListResult{
		Events: results,
		Count:  len(results),
	}

	output := actions.NewActionOutput(result)
	output.WithMetadata("count", len(results))

	return output, nil
}

// CalendarDeleteAction implements the Google Calendar Delete Event action
type CalendarDeleteAction struct {
	credentialService credential.Service
	baseURL           string
}

// CalendarDeleteConfig defines the configuration for deleting an event
type CalendarDeleteConfig struct {
	CalendarID string `json:"calendar_id"` // Usually "primary"
	EventID    string `json:"event_id"`
}

// CalendarDeleteResult represents the result of deleting an event
type CalendarDeleteResult struct {
	EventID string `json:"event_id"`
	Deleted bool   `json:"deleted"`
}

// Validate validates the Calendar delete configuration
func (c *CalendarDeleteConfig) Validate() error {
	if c.CalendarID == "" {
		return fmt.Errorf("calendar_id is required")
	}
	if c.EventID == "" {
		return fmt.Errorf("event_id is required")
	}
	return nil
}

// NewCalendarDeleteAction creates a new Calendar delete action
func NewCalendarDeleteAction(credentialService credential.Service) *CalendarDeleteAction {
	return &CalendarDeleteAction{
		credentialService: credentialService,
	}
}

// Execute implements the Action interface
func (a *CalendarDeleteAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	config, ok := input.Config.(CalendarDeleteConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected CalendarDeleteConfig")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	tenantID, err := extractString(input.Context, "env.tenant_id")
	if err != nil {
		return nil, fmt.Errorf("tenant_id is required in context: %w", err)
	}

	credentialID, err := extractString(input.Context, "credential_id")
	if err != nil {
		return nil, fmt.Errorf("credential_id is required in context: %w", err)
	}

	decryptedCred, err := a.credentialService.GetValue(ctx, tenantID, credentialID, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	token, err := createOAuth2Token(decryptedCred.Value)
	if err != nil {
		return nil, err
	}

	var calendarService *calendar.Service
	if a.baseURL != "" {
		calendarService, err = calendar.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(a.baseURL))
	} else {
		calendarService, err = calendar.NewService(ctx, createOAuth2Client(ctx, token))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Calendar service: %w", err)
	}

	// Delete event
	err = calendarService.Events.Delete(config.CalendarID, config.EventID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to delete event: %w", err)
	}

	result := &CalendarDeleteResult{
		EventID: config.EventID,
		Deleted: true,
	}

	output := actions.NewActionOutput(result)
	output.WithMetadata("event_id", config.EventID)
	output.WithMetadata("deleted", true)

	return output, nil
}
