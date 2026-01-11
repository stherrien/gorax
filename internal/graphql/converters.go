package graphql

import (
	"encoding/json"
	"fmt"

	"github.com/gorax/gorax/internal/graphql/generated"
	"github.com/gorax/gorax/internal/schedule"
	"github.com/gorax/gorax/internal/template"
	"github.com/gorax/gorax/internal/webhook"
	"github.com/gorax/gorax/internal/workflow"
)

// Workflow converters
func toGraphQLWorkflow(w *workflow.Workflow) (*generated.Workflow, error) {
	if w == nil {
		return nil, nil
	}

	defStr := string(w.Definition)
	return &generated.Workflow{
		ID:          w.ID,
		TenantID:    w.TenantID,
		Name:        w.Name,
		Description: w.Description,
		Definition:  defStr,
		Status:      w.Status,
		Version:     w.Version,
		CreatedBy:   w.CreatedBy,
		CreatedAt:   w.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   w.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func toGraphQLWorkflows(workflows []*workflow.Workflow) ([]*generated.Workflow, error) {
	result := make([]*generated.Workflow, len(workflows))
	for i, w := range workflows {
		gw, err := toGraphQLWorkflow(w)
		if err != nil {
			return nil, err
		}
		result[i] = gw
	}
	return result, nil
}

// Execution converters
func toGraphQLExecution(e *workflow.Execution) (*generated.Execution, error) {
	if e == nil {
		return nil, nil
	}

	result := &generated.Execution{
		ID:              e.ID,
		TenantID:        e.TenantID,
		WorkflowID:      e.WorkflowID,
		WorkflowVersion: e.WorkflowVersion,
		Status:          e.Status,
		TriggerType:     e.TriggerType,
		ExecutionDepth:  e.ExecutionDepth,
		CreatedAt:       e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if e.TriggerData != nil {
		triggerDataStr := string(*e.TriggerData)
		result.TriggerData = &triggerDataStr
	}

	if e.OutputData != nil {
		outputDataStr := string(*e.OutputData)
		result.OutputData = &outputDataStr
	}

	if e.ErrorMessage != nil {
		result.ErrorMessage = e.ErrorMessage
	}

	if e.ParentExecutionID != nil {
		result.ParentExecutionID = e.ParentExecutionID
	}

	if e.StartedAt != nil {
		startedAtStr := e.StartedAt.Format("2006-01-02T15:04:05Z")
		result.StartedAt = &startedAtStr
	}

	if e.CompletedAt != nil {
		completedAtStr := e.CompletedAt.Format("2006-01-02T15:04:05Z")
		result.CompletedAt = &completedAtStr
	}

	return result, nil
}

func toGraphQLExecutions(executions []*workflow.Execution) ([]*generated.Execution, error) {
	result := make([]*generated.Execution, len(executions))
	for i, e := range executions {
		ge, err := toGraphQLExecution(e)
		if err != nil {
			return nil, err
		}
		result[i] = ge
	}
	return result, nil
}

func toGraphQLStepExecutions(steps []*workflow.StepExecution) ([]*generated.StepExecution, error) {
	result := make([]*generated.StepExecution, len(steps))
	for i, s := range steps {
		se := &generated.StepExecution{
			ID:          s.ID,
			ExecutionID: s.ExecutionID,
			NodeID:      s.NodeID,
			NodeType:    s.NodeType,
			Status:      s.Status,
			RetryCount:  s.RetryCount,
		}

		if s.InputData != nil {
			inputDataStr := string(*s.InputData)
			se.InputData = &inputDataStr
		}

		if s.OutputData != nil {
			outputDataStr := string(*s.OutputData)
			se.OutputData = &outputDataStr
		}

		if s.ErrorMessage != nil {
			se.ErrorMessage = s.ErrorMessage
		}

		if s.StartedAt != nil {
			startedAtStr := s.StartedAt.Format("2006-01-02T15:04:05Z")
			se.StartedAt = &startedAtStr
		}

		if s.CompletedAt != nil {
			completedAtStr := s.CompletedAt.Format("2006-01-02T15:04:05Z")
			se.CompletedAt = &completedAtStr
		}

		if s.DurationMs != nil {
			se.DurationMs = s.DurationMs
		}

		result[i] = se
	}
	return result, nil
}

// Webhook converters
func toGraphQLWebhook(w *webhook.Webhook) (*generated.Webhook, error) {
	if w == nil {
		return nil, nil
	}

	result := &generated.Webhook{
		ID:           w.ID,
		TenantID:     w.TenantID,
		WorkflowID:   w.WorkflowID,
		NodeID:       w.NodeID,
		Name:         w.Name,
		Path:         w.Path,
		Secret:       w.Secret,
		AuthType:     w.AuthType,
		Description:  w.Description,
		Priority:     w.Priority,
		Enabled:      w.Enabled,
		TriggerCount: w.TriggerCount,
		CreatedAt:    w.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    w.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if w.LastTriggeredAt != nil {
		lastTriggeredAtStr := w.LastTriggeredAt.Format("2006-01-02T15:04:05Z")
		result.LastTriggeredAt = &lastTriggeredAtStr
	}

	return result, nil
}

func toGraphQLWebhooks(webhooks []*webhook.Webhook) ([]*generated.Webhook, error) {
	result := make([]*generated.Webhook, len(webhooks))
	for i, w := range webhooks {
		gw, err := toGraphQLWebhook(w)
		if err != nil {
			return nil, err
		}
		result[i] = gw
	}
	return result, nil
}

func toGraphQLWebhookEvents(events []*webhook.WebhookEvent) ([]*generated.WebhookEvent, error) {
	result := make([]*generated.WebhookEvent, len(events))
	for i, e := range events {
		headersJSON, err := json.Marshal(e.RequestHeaders)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request headers: %w", err)
		}

		event := &generated.WebhookEvent{
			ID:             e.ID,
			TenantID:       e.TenantID,
			WebhookID:      e.WebhookID,
			RequestMethod:  e.RequestMethod,
			RequestHeaders: string(headersJSON),
			RequestBody:    string(e.RequestBody),
			Status:         string(e.Status),
			ReplayCount:    e.ReplayCount,
			CreatedAt:      e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}

		if e.ExecutionID != nil {
			event.ExecutionID = e.ExecutionID
		}

		if e.ResponseStatus != nil {
			event.ResponseStatus = e.ResponseStatus
		}

		if e.ProcessingTimeMs != nil {
			event.ProcessingTimeMs = e.ProcessingTimeMs
		}

		if e.ErrorMessage != nil {
			event.ErrorMessage = e.ErrorMessage
		}

		if e.FilteredReason != nil {
			event.FilteredReason = e.FilteredReason
		}

		if e.SourceEventID != nil {
			event.SourceEventID = e.SourceEventID
		}

		if e.Metadata != nil {
			metadataJSON, err := json.Marshal(e.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal metadata: %w", err)
			}
			metadataStr := string(metadataJSON)
			event.Metadata = &metadataStr
		}

		result[i] = event
	}
	return result, nil
}

// Schedule converters
func toGraphQLSchedule(s *schedule.Schedule) (*generated.Schedule, error) {
	if s == nil {
		return nil, nil
	}

	result := &generated.Schedule{
		ID:             s.ID,
		TenantID:       s.TenantID,
		WorkflowID:     s.WorkflowID,
		Name:           s.Name,
		CronExpression: s.CronExpression,
		Timezone:       s.Timezone,
		Enabled:        s.Enabled,
		CreatedBy:      s.CreatedBy,
		CreatedAt:      s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if s.NextRunAt != nil {
		nextRunAtStr := s.NextRunAt.Format("2006-01-02T15:04:05Z")
		result.NextRunAt = &nextRunAtStr
	}

	if s.LastRunAt != nil {
		lastRunAtStr := s.LastRunAt.Format("2006-01-02T15:04:05Z")
		result.LastRunAt = &lastRunAtStr
	}

	if s.LastExecutionID != nil {
		result.LastExecutionID = s.LastExecutionID
	}

	return result, nil
}

func toGraphQLSchedules(schedules []*schedule.Schedule) ([]*generated.Schedule, error) {
	result := make([]*generated.Schedule, len(schedules))
	for i, s := range schedules {
		gs, err := toGraphQLSchedule(s)
		if err != nil {
			return nil, err
		}
		result[i] = gs
	}
	return result, nil
}

// Template converters
func toGraphQLTemplate(t *template.Template) (*generated.Template, error) {
	if t == nil {
		return nil, nil
	}

	defStr := string(t.Definition)
	result := &generated.Template{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Category:    t.Category,
		Definition:  defStr,
		Tags:        t.Tags,
		IsPublic:    t.IsPublic,
		UsageCount:  t.UsageCount,
		CreatedBy:   t.CreatedBy,
		CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if t.TenantID != nil {
		result.TenantID = t.TenantID
	}

	return result, nil
}

func toGraphQLTemplates(templates []*template.Template) ([]*generated.Template, error) {
	result := make([]*generated.Template, len(templates))
	for i, t := range templates {
		gt, err := toGraphQLTemplate(t)
		if err != nil {
			return nil, err
		}
		result[i] = gt
	}
	return result, nil
}
