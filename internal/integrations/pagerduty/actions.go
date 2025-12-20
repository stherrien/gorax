package pagerduty

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorax/gorax/internal/integrations"
)

// CreateIncidentAction implements the pagerduty:create_incident action
type CreateIncidentAction struct {
	apiKey string
	email  string
}

// NewCreateIncidentAction creates a new CreateIncident action
func NewCreateIncidentAction(apiKey, email string) *CreateIncidentAction {
	return &CreateIncidentAction{apiKey: apiKey, email: email}
}

// Execute implements the Action interface
func (a *CreateIncidentAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var incidentConfig CreateIncidentConfig
	if err := json.Unmarshal(configJSON, &incidentConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := incidentConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewClient(a.apiKey, a.email)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	incident, err := client.CreateIncident(ctx, incidentConfig.Title, incidentConfig.Service, incidentConfig.Urgency, incidentConfig.Body, incidentConfig.IncidentKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}

	return map[string]interface{}{
		"id":              incident.ID,
		"incident_number": incident.IncidentNumber,
		"title":           incident.Title,
		"status":          incident.Status,
		"urgency":         incident.Urgency,
		"url":             incident.HTMLURL,
	}, nil
}

// Validate implements the Action interface
func (a *CreateIncidentAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var incidentConfig CreateIncidentConfig
	if err := json.Unmarshal(configJSON, &incidentConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return incidentConfig.Validate()
}

// Name implements the Action interface
func (a *CreateIncidentAction) Name() string {
	return "pagerduty:create_incident"
}

// Description implements the Action interface
func (a *CreateIncidentAction) Description() string {
	return "Create a new PagerDuty incident"
}

// AcknowledgeIncidentAction implements the pagerduty:acknowledge_incident action
type AcknowledgeIncidentAction struct {
	apiKey string
	email  string
}

// NewAcknowledgeIncidentAction creates a new AcknowledgeIncident action
func NewAcknowledgeIncidentAction(apiKey, email string) *AcknowledgeIncidentAction {
	return &AcknowledgeIncidentAction{apiKey: apiKey, email: email}
}

// Execute implements the Action interface
func (a *AcknowledgeIncidentAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var ackConfig AcknowledgeIncidentConfig
	if err := json.Unmarshal(configJSON, &ackConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := ackConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewClient(a.apiKey, a.email)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.AcknowledgeIncident(ctx, ackConfig.IncidentID); err != nil {
		return nil, fmt.Errorf("failed to acknowledge incident: %w", err)
	}

	return map[string]interface{}{
		"success":     true,
		"incident_id": ackConfig.IncidentID,
		"status":      "acknowledged",
	}, nil
}

// Validate implements the Action interface
func (a *AcknowledgeIncidentAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var ackConfig AcknowledgeIncidentConfig
	if err := json.Unmarshal(configJSON, &ackConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return ackConfig.Validate()
}

// Name implements the Action interface
func (a *AcknowledgeIncidentAction) Name() string {
	return "pagerduty:acknowledge_incident"
}

// Description implements the Action interface
func (a *AcknowledgeIncidentAction) Description() string {
	return "Acknowledge a PagerDuty incident"
}

// ResolveIncidentAction implements the pagerduty:resolve_incident action
type ResolveIncidentAction struct {
	apiKey string
	email  string
}

// NewResolveIncidentAction creates a new ResolveIncident action
func NewResolveIncidentAction(apiKey, email string) *ResolveIncidentAction {
	return &ResolveIncidentAction{apiKey: apiKey, email: email}
}

// Execute implements the Action interface
func (a *ResolveIncidentAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var resolveConfig ResolveIncidentConfig
	if err := json.Unmarshal(configJSON, &resolveConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := resolveConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewClient(a.apiKey, a.email)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.ResolveIncident(ctx, resolveConfig.IncidentID); err != nil {
		return nil, fmt.Errorf("failed to resolve incident: %w", err)
	}

	return map[string]interface{}{
		"success":     true,
		"incident_id": resolveConfig.IncidentID,
		"status":      "resolved",
	}, nil
}

// Validate implements the Action interface
func (a *ResolveIncidentAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var resolveConfig ResolveIncidentConfig
	if err := json.Unmarshal(configJSON, &resolveConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return resolveConfig.Validate()
}

// Name implements the Action interface
func (a *ResolveIncidentAction) Name() string {
	return "pagerduty:resolve_incident"
}

// Description implements the Action interface
func (a *ResolveIncidentAction) Description() string {
	return "Resolve a PagerDuty incident"
}

// AddNoteAction implements the pagerduty:add_note action
type AddNoteAction struct {
	apiKey string
	email  string
}

// NewAddNoteAction creates a new AddNote action
func NewAddNoteAction(apiKey, email string) *AddNoteAction {
	return &AddNoteAction{apiKey: apiKey, email: email}
}

// Execute implements the Action interface
func (a *AddNoteAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var noteConfig AddNoteConfig
	if err := json.Unmarshal(configJSON, &noteConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := noteConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewClient(a.apiKey, a.email)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	note, err := client.AddNote(ctx, noteConfig.IncidentID, noteConfig.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to add note: %w", err)
	}

	return map[string]interface{}{
		"id":      note.ID,
		"content": note.Content,
	}, nil
}

// Validate implements the Action interface
func (a *AddNoteAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var noteConfig AddNoteConfig
	if err := json.Unmarshal(configJSON, &noteConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return noteConfig.Validate()
}

// Name implements the Action interface
func (a *AddNoteAction) Name() string {
	return "pagerduty:add_note"
}

// Description implements the Action interface
func (a *AddNoteAction) Description() string {
	return "Add a note to a PagerDuty incident"
}

// Ensure all actions implement the Action interface
var (
	_ integrations.Action = (*CreateIncidentAction)(nil)
	_ integrations.Action = (*AcknowledgeIncidentAction)(nil)
	_ integrations.Action = (*ResolveIncidentAction)(nil)
	_ integrations.Action = (*AddNoteAction)(nil)
)
