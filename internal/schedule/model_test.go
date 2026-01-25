package schedule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOverlapPolicy_Constants(t *testing.T) {
	// Verify overlap policy constants have correct string values
	assert.Equal(t, "skip", string(OverlapPolicySkip))
	assert.Equal(t, "queue", string(OverlapPolicyQueue))
	assert.Equal(t, "terminate", string(OverlapPolicyTerminate))
}

func TestOverlapPolicy_IsValid_AllPolicies(t *testing.T) {
	tests := []struct {
		name     string
		policy   OverlapPolicy
		expected bool
	}{
		{"skip is valid", OverlapPolicySkip, true},
		{"queue is valid", OverlapPolicyQueue, true},
		{"terminate is valid", OverlapPolicyTerminate, true},
		{"empty is invalid", "", false},
		{"random string is invalid", OverlapPolicy("random"), false},
		{"uppercase SKIP is invalid", OverlapPolicy("SKIP"), false},
		{"mixed case Skip is invalid", OverlapPolicy("Skip"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.policy.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSchedule_Fields(t *testing.T) {
	now := time.Now()
	nextRun := now.Add(1 * time.Hour)
	lastRun := now.Add(-1 * time.Hour)
	execID := "exec-123"
	runningID := "running-456"

	schedule := Schedule{
		ID:                 "sched-1",
		TenantID:           "tenant-1",
		WorkflowID:         "workflow-1",
		Name:               "Test Schedule",
		CronExpression:     "0 * * * *",
		Timezone:           "America/New_York",
		OverlapPolicy:      OverlapPolicySkip,
		Enabled:            true,
		NextRunAt:          &nextRun,
		LastRunAt:          &lastRun,
		LastExecutionID:    &execID,
		RunningExecutionID: &runningID,
		CreatedBy:          "user-1",
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	assert.Equal(t, "sched-1", schedule.ID)
	assert.Equal(t, "tenant-1", schedule.TenantID)
	assert.Equal(t, "workflow-1", schedule.WorkflowID)
	assert.Equal(t, "Test Schedule", schedule.Name)
	assert.Equal(t, "0 * * * *", schedule.CronExpression)
	assert.Equal(t, "America/New_York", schedule.Timezone)
	assert.Equal(t, OverlapPolicySkip, schedule.OverlapPolicy)
	assert.True(t, schedule.Enabled)
	assert.NotNil(t, schedule.NextRunAt)
	assert.NotNil(t, schedule.LastRunAt)
	assert.NotNil(t, schedule.LastExecutionID)
	assert.NotNil(t, schedule.RunningExecutionID)
}

func TestCreateScheduleInput_Fields(t *testing.T) {
	input := CreateScheduleInput{
		Name:           "Test Schedule",
		CronExpression: "*/5 * * * *",
		Timezone:       "UTC",
		OverlapPolicy:  OverlapPolicyQueue,
		Enabled:        true,
	}

	assert.Equal(t, "Test Schedule", input.Name)
	assert.Equal(t, "*/5 * * * *", input.CronExpression)
	assert.Equal(t, "UTC", input.Timezone)
	assert.Equal(t, OverlapPolicyQueue, input.OverlapPolicy)
	assert.True(t, input.Enabled)
}

func TestUpdateScheduleInput_Fields(t *testing.T) {
	name := "Updated Name"
	cron := "0 0 * * *"
	tz := "Europe/London"
	policy := OverlapPolicyTerminate
	enabled := false

	input := UpdateScheduleInput{
		Name:           &name,
		CronExpression: &cron,
		Timezone:       &tz,
		OverlapPolicy:  &policy,
		Enabled:        &enabled,
	}

	assert.NotNil(t, input.Name)
	assert.Equal(t, "Updated Name", *input.Name)
	assert.NotNil(t, input.CronExpression)
	assert.Equal(t, "0 0 * * *", *input.CronExpression)
	assert.NotNil(t, input.Timezone)
	assert.Equal(t, "Europe/London", *input.Timezone)
	assert.NotNil(t, input.OverlapPolicy)
	assert.Equal(t, OverlapPolicyTerminate, *input.OverlapPolicy)
	assert.NotNil(t, input.Enabled)
	assert.False(t, *input.Enabled)
}

func TestUpdateScheduleInput_PartialUpdate(t *testing.T) {
	// Test that partial updates work (only some fields set)
	name := "New Name"
	input := UpdateScheduleInput{
		Name: &name,
	}

	assert.NotNil(t, input.Name)
	assert.Nil(t, input.CronExpression)
	assert.Nil(t, input.Timezone)
	assert.Nil(t, input.OverlapPolicy)
	assert.Nil(t, input.Enabled)
}

func TestExecutionLog_Fields(t *testing.T) {
	now := time.Now()
	startedAt := now.Add(-10 * time.Minute)
	completedAt := now
	execID := "exec-789"
	errMsg := "test error"
	skipReason := "previous execution still running"

	log := ExecutionLog{
		ID:            "log-1",
		TenantID:      "tenant-1",
		ScheduleID:    "sched-1",
		ExecutionID:   &execID,
		Status:        ExecutionLogStatusCompleted,
		StartedAt:     &startedAt,
		CompletedAt:   &completedAt,
		ErrorMessage:  &errMsg,
		TriggerTime:   now.Add(-15 * time.Minute),
		SkippedReason: &skipReason,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	assert.Equal(t, "log-1", log.ID)
	assert.Equal(t, "tenant-1", log.TenantID)
	assert.Equal(t, "sched-1", log.ScheduleID)
	assert.NotNil(t, log.ExecutionID)
	assert.Equal(t, "exec-789", *log.ExecutionID)
	assert.Equal(t, ExecutionLogStatusCompleted, log.Status)
	assert.NotNil(t, log.StartedAt)
	assert.NotNil(t, log.CompletedAt)
	assert.NotNil(t, log.ErrorMessage)
	assert.NotNil(t, log.SkippedReason)
}

func TestExecutionLogStatus_Constants(t *testing.T) {
	// Verify all status constants
	statuses := []ExecutionLogStatus{
		ExecutionLogStatusPending,
		ExecutionLogStatusRunning,
		ExecutionLogStatusCompleted,
		ExecutionLogStatusFailed,
		ExecutionLogStatusSkipped,
		ExecutionLogStatusTerminated,
	}

	expectedStrings := []string{
		"pending",
		"running",
		"completed",
		"failed",
		"skipped",
		"terminated",
	}

	for i, status := range statuses {
		assert.Equal(t, expectedStrings[i], string(status))
	}
}

func TestExecutionLogListParams_Fields(t *testing.T) {
	status := ExecutionLogStatusFailed
	params := ExecutionLogListParams{
		ScheduleID: "sched-1",
		Status:     &status,
		Limit:      50,
		Offset:     10,
	}

	assert.Equal(t, "sched-1", params.ScheduleID)
	assert.NotNil(t, params.Status)
	assert.Equal(t, ExecutionLogStatusFailed, *params.Status)
	assert.Equal(t, 50, params.Limit)
	assert.Equal(t, 10, params.Offset)
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{Message: "test error message"}
	assert.Equal(t, "test error message", err.Error())
}

func TestScheduleWithWorkflow_EmbeddedFields(t *testing.T) {
	now := time.Now()
	schedule := ScheduleWithWorkflow{
		Schedule: Schedule{
			ID:             "sched-1",
			TenantID:       "tenant-1",
			WorkflowID:     "workflow-1",
			Name:           "Test Schedule",
			CronExpression: "0 * * * *",
			Timezone:       "UTC",
			OverlapPolicy:  OverlapPolicySkip,
			Enabled:        true,
			CreatedBy:      "user-1",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		WorkflowName:   "Test Workflow",
		WorkflowStatus: "active",
	}

	// Test embedded Schedule fields
	assert.Equal(t, "sched-1", schedule.ID)
	assert.Equal(t, "tenant-1", schedule.TenantID)
	assert.Equal(t, OverlapPolicySkip, schedule.OverlapPolicy)

	// Test ScheduleWithWorkflow specific fields
	assert.Equal(t, "Test Workflow", schedule.WorkflowName)
	assert.Equal(t, "active", schedule.WorkflowStatus)
}
