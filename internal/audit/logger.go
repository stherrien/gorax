package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserContext contains user information extracted from request context
type UserContext struct {
	TenantID  string
	UserID    string
	UserEmail string
	SessionID string
	IPAddress string
	UserAgent string
}

// LogUserAction logs a user-related audit event
func (s *Service) LogUserAction(
	ctx context.Context,
	userCtx UserContext,
	eventType EventType,
	action string,
	userID string,
	userName string,
	status Status,
	errorMsg string,
	details map[string]any,
) error {
	event := &AuditEvent{
		ID:           uuid.New().String(),
		TenantID:     userCtx.TenantID,
		UserID:       userCtx.UserID,
		UserEmail:    userCtx.UserEmail,
		SessionID:    userCtx.SessionID,
		Category:     CategoryUserManagement,
		EventType:    eventType,
		Action:       action,
		ResourceType: string(ResourceTypeUser),
		ResourceID:   userID,
		ResourceName: userName,
		IPAddress:    userCtx.IPAddress,
		UserAgent:    userCtx.UserAgent,
		Severity:     determineSeverityForAction(eventType, status),
		Status:       status,
		ErrorMessage: errorMsg,
		Details:      details,
		Metadata:     make(map[string]any),
		CreatedAt:    time.Now(),
	}

	return s.LogEvent(ctx, event)
}

// LogWorkflowAction logs a workflow-related audit event
func (s *Service) LogWorkflowAction(
	ctx context.Context,
	userCtx UserContext,
	eventType EventType,
	action string,
	workflowID string,
	workflowName string,
	status Status,
	errorMsg string,
	oldValues map[string]any,
	newValues map[string]any,
	details map[string]any,
) error {
	event := &AuditEvent{
		ID:           uuid.New().String(),
		TenantID:     userCtx.TenantID,
		UserID:       userCtx.UserID,
		UserEmail:    userCtx.UserEmail,
		SessionID:    userCtx.SessionID,
		WorkflowID:   &workflowID,
		Category:     CategoryWorkflow,
		EventType:    eventType,
		Action:       action,
		ResourceType: string(ResourceTypeWorkflow),
		ResourceID:   workflowID,
		ResourceName: workflowName,
		IPAddress:    userCtx.IPAddress,
		UserAgent:    userCtx.UserAgent,
		Severity:     determineSeverityForAction(eventType, status),
		Status:       status,
		ErrorMessage: errorMsg,
		OldValues:    oldValues,
		NewValues:    newValues,
		Details:      details,
		Metadata:     make(map[string]any),
		CreatedAt:    time.Now(),
	}

	return s.LogEvent(ctx, event)
}

// LogExecutionAction logs a workflow execution-related audit event
func (s *Service) LogExecutionAction(
	ctx context.Context,
	userCtx UserContext,
	eventType EventType,
	action string,
	workflowID string,
	executionID string,
	executionName string,
	status Status,
	errorMsg string,
	details map[string]any,
) error {
	event := &AuditEvent{
		ID:           uuid.New().String(),
		TenantID:     userCtx.TenantID,
		UserID:       userCtx.UserID,
		UserEmail:    userCtx.UserEmail,
		SessionID:    userCtx.SessionID,
		WorkflowID:   &workflowID,
		ExecutionID:  &executionID,
		Category:     CategoryWorkflow,
		EventType:    eventType,
		Action:       action,
		ResourceType: string(ResourceTypeExecution),
		ResourceID:   executionID,
		ResourceName: executionName,
		IPAddress:    userCtx.IPAddress,
		UserAgent:    userCtx.UserAgent,
		Severity:     determineSeverityForAction(eventType, status),
		Status:       status,
		ErrorMessage: errorMsg,
		Details:      details,
		Metadata:     make(map[string]any),
		CreatedAt:    time.Now(),
	}

	return s.LogEvent(ctx, event)
}

// LogCredentialAction logs a credential-related audit event
func (s *Service) LogCredentialAction(
	ctx context.Context,
	userCtx UserContext,
	eventType EventType,
	action string,
	credentialID string,
	credentialName string,
	status Status,
	errorMsg string,
	details map[string]any,
) error {
	event := &AuditEvent{
		ID:           uuid.New().String(),
		TenantID:     userCtx.TenantID,
		UserID:       userCtx.UserID,
		UserEmail:    userCtx.UserEmail,
		SessionID:    userCtx.SessionID,
		Category:     CategoryCredential,
		EventType:    eventType,
		Action:       action,
		ResourceType: string(ResourceTypeCredential),
		ResourceID:   credentialID,
		ResourceName: credentialName,
		IPAddress:    userCtx.IPAddress,
		UserAgent:    userCtx.UserAgent,
		Severity:     determineSeverityForCredential(eventType, status),
		Status:       status,
		ErrorMessage: errorMsg,
		Details:      details,
		Metadata:     make(map[string]any),
		CreatedAt:    time.Now(),
	}

	return s.LogEvent(ctx, event)
}

// LogPermissionAction logs a permission change audit event
func (s *Service) LogPermissionAction(
	ctx context.Context,
	userCtx UserContext,
	eventType EventType,
	action string,
	resourceType ResourceType,
	resourceID string,
	resourceName string,
	targetUserID string,
	permission string,
	status Status,
	errorMsg string,
) error {
	event := &AuditEvent{
		ID:           uuid.New().String(),
		TenantID:     userCtx.TenantID,
		UserID:       userCtx.UserID,
		UserEmail:    userCtx.UserEmail,
		SessionID:    userCtx.SessionID,
		Category:     CategoryAuthorization,
		EventType:    eventType,
		Action:       action,
		ResourceType: string(resourceType),
		ResourceID:   resourceID,
		ResourceName: resourceName,
		IPAddress:    userCtx.IPAddress,
		UserAgent:    userCtx.UserAgent,
		Severity:     SeverityWarning,
		Status:       status,
		ErrorMessage: errorMsg,
		Details: map[string]any{
			"target_user_id": targetUserID,
			"permission":     permission,
		},
		Metadata:  make(map[string]any),
		CreatedAt: time.Now(),
	}

	return s.LogEvent(ctx, event)
}

// LogAuthenticationAction logs an authentication-related audit event
func (s *Service) LogAuthenticationAction(
	ctx context.Context,
	userCtx UserContext,
	eventType EventType,
	action string,
	status Status,
	errorMsg string,
	details map[string]any,
) error {
	severity := SeverityInfo
	if status == StatusFailure {
		severity = SeverityCritical
	}

	event := &AuditEvent{
		ID:           uuid.New().String(),
		TenantID:     userCtx.TenantID,
		UserID:       userCtx.UserID,
		UserEmail:    userCtx.UserEmail,
		SessionID:    userCtx.SessionID,
		Category:     CategoryAuthentication,
		EventType:    eventType,
		Action:       action,
		ResourceType: string(ResourceTypeUser),
		ResourceID:   userCtx.UserID,
		ResourceName: userCtx.UserEmail,
		IPAddress:    userCtx.IPAddress,
		UserAgent:    userCtx.UserAgent,
		Severity:     severity,
		Status:       status,
		ErrorMessage: errorMsg,
		Details:      details,
		Metadata:     make(map[string]any),
		CreatedAt:    time.Now(),
	}

	return s.LogEvent(ctx, event)
}

// LogDataAccessAction logs a data access audit event
func (s *Service) LogDataAccessAction(
	ctx context.Context,
	userCtx UserContext,
	action string,
	resourceType ResourceType,
	resourceID string,
	resourceName string,
	status Status,
	details map[string]any,
) error {
	event := &AuditEvent{
		ID:           uuid.New().String(),
		TenantID:     userCtx.TenantID,
		UserID:       userCtx.UserID,
		UserEmail:    userCtx.UserEmail,
		SessionID:    userCtx.SessionID,
		Category:     CategoryDataAccess,
		EventType:    EventTypeRead,
		Action:       action,
		ResourceType: string(resourceType),
		ResourceID:   resourceID,
		ResourceName: resourceName,
		IPAddress:    userCtx.IPAddress,
		UserAgent:    userCtx.UserAgent,
		Severity:     SeverityInfo,
		Status:       status,
		Details:      details,
		Metadata:     make(map[string]any),
		CreatedAt:    time.Now(),
	}

	return s.LogEvent(ctx, event)
}

// LogIntegrationAction logs an integration-related audit event
func (s *Service) LogIntegrationAction(
	ctx context.Context,
	userCtx UserContext,
	eventType EventType,
	action string,
	integrationID string,
	integrationName string,
	status Status,
	errorMsg string,
	details map[string]any,
) error {
	event := &AuditEvent{
		ID:           uuid.New().String(),
		TenantID:     userCtx.TenantID,
		UserID:       userCtx.UserID,
		UserEmail:    userCtx.UserEmail,
		SessionID:    userCtx.SessionID,
		Category:     CategoryIntegration,
		EventType:    eventType,
		Action:       action,
		ResourceType: string(ResourceTypeIntegration),
		ResourceID:   integrationID,
		ResourceName: integrationName,
		IPAddress:    userCtx.IPAddress,
		UserAgent:    userCtx.UserAgent,
		Severity:     determineSeverityForAction(eventType, status),
		Status:       status,
		ErrorMessage: errorMsg,
		Details:      details,
		Metadata:     make(map[string]any),
		CreatedAt:    time.Now(),
	}

	return s.LogEvent(ctx, event)
}

// LogConfigurationAction logs a configuration change audit event
func (s *Service) LogConfigurationAction(
	ctx context.Context,
	userCtx UserContext,
	eventType EventType,
	action string,
	configType string,
	configID string,
	configName string,
	oldValues map[string]any,
	newValues map[string]any,
	status Status,
	errorMsg string,
) error {
	event := &AuditEvent{
		ID:           uuid.New().String(),
		TenantID:     userCtx.TenantID,
		UserID:       userCtx.UserID,
		UserEmail:    userCtx.UserEmail,
		SessionID:    userCtx.SessionID,
		Category:     CategoryConfiguration,
		EventType:    eventType,
		Action:       action,
		ResourceType: configType,
		ResourceID:   configID,
		ResourceName: configName,
		IPAddress:    userCtx.IPAddress,
		UserAgent:    userCtx.UserAgent,
		Severity:     SeverityWarning,
		Status:       status,
		ErrorMessage: errorMsg,
		OldValues:    oldValues,
		NewValues:    newValues,
		Metadata:     make(map[string]any),
		CreatedAt:    time.Now(),
	}

	return s.LogEvent(ctx, event)
}

// determineSeverityForAction determines severity based on event type and status
func determineSeverityForAction(eventType EventType, status Status) Severity {
	if status == StatusFailure {
		return SeverityError
	}

	switch eventType {
	case EventTypeDelete:
		return SeverityWarning
	case EventTypeFailed:
		return SeverityError
	case EventTypePermissionChange, EventTypeGrant, EventTypeRevoke:
		return SeverityWarning
	default:
		return SeverityInfo
	}
}

// determineSeverityForCredential determines severity for credential actions
func determineSeverityForCredential(eventType EventType, status Status) Severity {
	if status == StatusFailure {
		return SeverityError
	}

	switch eventType {
	case EventTypeAccess, EventTypeRead:
		return SeverityWarning
	case EventTypeDelete:
		return SeverityCritical
	case EventTypeUpdate, EventTypeCreate:
		return SeverityWarning
	default:
		return SeverityInfo
	}
}
