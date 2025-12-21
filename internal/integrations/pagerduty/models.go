package pagerduty

import "fmt"

// CreateIncidentConfig represents configuration for CreateIncident action
type CreateIncidentConfig struct {
	Title       string `json:"title"`
	Service     string `json:"service"` // Service ID
	Urgency     string `json:"urgency"` // "high" or "low"
	Body        string `json:"body,omitempty"`
	IncidentKey string `json:"incident_key,omitempty"`
}

// Validate validates CreateIncidentConfig
func (c *CreateIncidentConfig) Validate() error {
	if c.Title == "" {
		return fmt.Errorf("title is required")
	}
	if c.Service == "" {
		return fmt.Errorf("service is required")
	}
	if c.Urgency != "" && c.Urgency != "high" && c.Urgency != "low" {
		return fmt.Errorf("urgency must be 'high' or 'low'")
	}
	return nil
}

// AcknowledgeIncidentConfig represents configuration for AcknowledgeIncident action
type AcknowledgeIncidentConfig struct {
	IncidentID string `json:"incident_id"`
}

// Validate validates AcknowledgeIncidentConfig
func (c *AcknowledgeIncidentConfig) Validate() error {
	if c.IncidentID == "" {
		return fmt.Errorf("incident_id is required")
	}
	return nil
}

// ResolveIncidentConfig represents configuration for ResolveIncident action
type ResolveIncidentConfig struct {
	IncidentID string `json:"incident_id"`
}

// Validate validates ResolveIncidentConfig
func (c *ResolveIncidentConfig) Validate() error {
	if c.IncidentID == "" {
		return fmt.Errorf("incident_id is required")
	}
	return nil
}

// AddNoteConfig represents configuration for AddNote action
type AddNoteConfig struct {
	IncidentID string `json:"incident_id"`
	Content    string `json:"content"`
}

// Validate validates AddNoteConfig
func (c *AddNoteConfig) Validate() error {
	if c.IncidentID == "" {
		return fmt.Errorf("incident_id is required")
	}
	if c.Content == "" {
		return fmt.Errorf("content is required")
	}
	return nil
}

// Incident represents a PagerDuty incident
type Incident struct {
	ID             string `json:"id"`
	IncidentNumber int    `json:"incident_number"`
	Title          string `json:"title"`
	Status         string `json:"status"`
	Urgency        string `json:"urgency"`
	HTMLURL        string `json:"html_url"`
}

// Note represents a PagerDuty incident note
type Note struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

// ErrorResponse represents a PagerDuty error response
type ErrorResponse struct {
	Error struct {
		Message string   `json:"message"`
		Code    int      `json:"code"`
		Errors  []string `json:"errors,omitempty"`
	} `json:"error"`
}
