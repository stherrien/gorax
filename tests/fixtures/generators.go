package fixtures

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// WorkflowDefinition generates a test workflow definition
type WorkflowDefinition struct {
	Name        string
	Description string
	NodeTypes   []string
	EdgeCount   int
}

// GenerateWorkflowDefinition generates a valid workflow definition
func GenerateWorkflowDefinition(opts WorkflowDefinition) json.RawMessage {
	if opts.Name == "" {
		opts.Name = "Test Workflow " + RandomString(8)
	}
	if opts.Description == "" {
		opts.Description = "Generated test workflow"
	}
	if len(opts.NodeTypes) == 0 {
		opts.NodeTypes = []string{"trigger", "action"}
	}

	nodes := make([]map[string]interface{}, 0, len(opts.NodeTypes))
	for i, nodeType := range opts.NodeTypes {
		node := map[string]interface{}{
			"id":   fmt.Sprintf("node-%d", i+1),
			"type": nodeType,
			"data": map[string]interface{}{
				"name": fmt.Sprintf("%s Node %d", nodeType, i+1),
			},
			"position": map[string]interface{}{
				"x": i * 200,
				"y": 100,
			},
		}
		nodes = append(nodes, node)
	}

	// Generate edges
	edges := make([]map[string]interface{}, 0, opts.EdgeCount)
	for i := 0; i < opts.EdgeCount && i < len(nodes)-1; i++ {
		edge := map[string]interface{}{
			"id":     fmt.Sprintf("edge-%d", i+1),
			"source": fmt.Sprintf("node-%d", i+1),
			"target": fmt.Sprintf("node-%d", i+2),
		}
		edges = append(edges, edge)
	}

	definition := map[string]interface{}{
		"nodes": nodes,
		"edges": edges,
	}

	jsonData, _ := json.Marshal(definition)
	return json.RawMessage(jsonData)
}

// UserData represents test user data
type UserData struct {
	Email string
	Name  string
	Role  string
}

// GenerateUser generates test user data
func GenerateUser() UserData {
	id := RandomString(6)
	return UserData{
		Email: fmt.Sprintf("user-%s@test.com", id),
		Name:  fmt.Sprintf("Test User %s", id),
		Role:  "user",
	}
}

// GenerateAdmin generates test admin user data
func GenerateAdmin() UserData {
	id := RandomString(6)
	return UserData{
		Email: fmt.Sprintf("admin-%s@test.com", id),
		Name:  fmt.Sprintf("Admin User %s", id),
		Role:  "admin",
	}
}

// TenantData represents test tenant data
type TenantData struct {
	Name string
	Slug string
}

// GenerateTenant generates test tenant data
func GenerateTenant() TenantData {
	id := RandomString(6)
	return TenantData{
		Name: fmt.Sprintf("Test Tenant %s", id),
		Slug: fmt.Sprintf("tenant-%s", id),
	}
}

// WorkflowTemplateData represents test workflow template data
type WorkflowTemplateData struct {
	Name        string
	Description string
	Category    string
	Tags        []string
	Definition  json.RawMessage
	Version     string
}

// GenerateWorkflowTemplate generates test workflow template data
func GenerateWorkflowTemplate() WorkflowTemplateData {
	id := RandomString(6)
	return WorkflowTemplateData{
		Name:        fmt.Sprintf("Template %s", id),
		Description: fmt.Sprintf("Test template for integration testing - %s", id),
		Category:    "automation",
		Tags:        []string{"test", "automation"},
		Definition: GenerateWorkflowDefinition(WorkflowDefinition{
			NodeTypes: []string{"trigger", "action"},
			EdgeCount: 1,
		}),
		Version: "1.0.0",
	}
}

// CredentialData represents test credential data
type CredentialData struct {
	Name        string
	Type        string
	Description string
	Value       map[string]string
}

// GenerateCredential generates test credential data
func GenerateCredential(credType string) CredentialData {
	id := RandomString(6)

	var value map[string]string
	switch credType {
	case "api_key":
		value = map[string]string{
			"api_key": RandomString(32),
		}
	case "oauth2":
		value = map[string]string{
			"access_token":  RandomString(64),
			"refresh_token": RandomString(64),
			"expires_at":    time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		}
	case "basic_auth":
		value = map[string]string{
			"username": fmt.Sprintf("user-%s", id),
			"password": RandomString(16),
		}
	default:
		value = map[string]string{
			"key": RandomString(32),
		}
	}

	return CredentialData{
		Name:        fmt.Sprintf("%s-cred-%s", credType, id),
		Type:        credType,
		Description: fmt.Sprintf("Test %s credential", credType),
		Value:       value,
	}
}

// WebhookData represents test webhook data
type WebhookData struct {
	Name        string
	Description string
	Secret      string
}

// GenerateWebhook generates test webhook data
func GenerateWebhook() WebhookData {
	id := RandomString(6)
	return WebhookData{
		Name:        fmt.Sprintf("Webhook %s", id),
		Description: fmt.Sprintf("Test webhook for integration testing - %s", id),
		Secret:      RandomString(32),
	}
}

// ExecutionData represents test execution data
type ExecutionData struct {
	Status      string
	TriggerType string
	TriggerData map[string]interface{}
	StartedAt   time.Time
	CompletedAt *time.Time
}

// GenerateExecution generates test execution data
func GenerateExecution(status string) ExecutionData {
	startedAt := time.Now().Add(-5 * time.Minute)
	var completedAt *time.Time

	if status == "completed" || status == "failed" {
		t := startedAt.Add(2 * time.Minute)
		completedAt = &t
	}

	return ExecutionData{
		Status:      status,
		TriggerType: "manual",
		TriggerData: map[string]interface{}{
			"user_id": uuid.New().String(),
		},
		StartedAt:   startedAt,
		CompletedAt: completedAt,
	}
}

// AuditLogData represents test audit log data
type AuditLogData struct {
	EventType    string
	ResourceType string
	ResourceID   string
	Action       string
	Metadata     map[string]interface{}
	CreatedAt    time.Time
}

// GenerateAuditLog generates test audit log data
func GenerateAuditLog(action string) AuditLogData {
	resourceID := uuid.New().String()
	return AuditLogData{
		EventType:    fmt.Sprintf("workflow.%s", action),
		ResourceType: "workflow",
		ResourceID:   resourceID,
		Action:       action,
		Metadata: map[string]interface{}{
			"ip_address": "127.0.0.1",
			"user_agent": "Test Agent",
		},
		CreatedAt: time.Now(),
	}
}

// OAuthConnectionData represents test OAuth connection data
type OAuthConnectionData struct {
	ProviderKey      string
	ProviderUserID   string
	ProviderUsername string
	ProviderEmail    string
	AccessToken      string
	RefreshToken     string
	TokenExpiry      time.Time
	Scopes           []string
}

// GenerateOAuthConnection generates test OAuth connection data
func GenerateOAuthConnection(providerKey string) OAuthConnectionData {
	id := RandomString(6)
	return OAuthConnectionData{
		ProviderKey:      providerKey,
		ProviderUserID:   fmt.Sprintf("provider-user-%s", id),
		ProviderUsername: fmt.Sprintf("user_%s", id),
		ProviderEmail:    fmt.Sprintf("user-%s@provider.com", id),
		AccessToken:      RandomString(64),
		RefreshToken:     RandomString(64),
		TokenExpiry:      time.Now().Add(1 * time.Hour),
		Scopes:           []string{"read", "write"},
	}
}

// ReviewData represents test review data
type ReviewData struct {
	Rating  int
	Comment string
}

// GenerateReview generates test review data
func GenerateReview(rating int) ReviewData {
	comments := map[int]string{
		1: "Not useful at all",
		2: "Needs significant improvement",
		3: "It's okay, works as expected",
		4: "Very good, with minor issues",
		5: "Excellent! Exactly what I needed",
	}

	comment, ok := comments[rating]
	if !ok {
		comment = "Test review"
	}

	return ReviewData{
		Rating:  rating,
		Comment: comment,
	}
}

// Utility functions

// RandomString generates a random string of specified length
func RandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("failed to generate random string: %v", err))
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

// RandomInt generates a random integer between min and max (inclusive)
func RandomInt(min, max int) int {
	b := make([]byte, 1)
	rand.Read(b)
	return min + int(b[0])%(max-min+1)
}

// RandomEmail generates a random email address
func RandomEmail() string {
	return fmt.Sprintf("test-%s@example.com", RandomString(8))
}

// RandomURL generates a random URL
func RandomURL() string {
	return fmt.Sprintf("https://example.com/%s", RandomString(10))
}

// PastTime returns a time in the past by the specified duration
func PastTime(d time.Duration) time.Time {
	return time.Now().Add(-d)
}

// FutureTime returns a time in the future by the specified duration
func FutureTime(d time.Duration) time.Time {
	return time.Now().Add(d)
}
