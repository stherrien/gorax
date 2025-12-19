package workflow

import (
	"encoding/json"
	"testing"
	"time"
)

// TestExecutionFilter_Validate tests validation of ExecutionFilter
func TestExecutionFilter_Validate(t *testing.T) {
	tests := []struct {
		name    string
		filter  ExecutionFilter
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty filter is valid",
			filter:  ExecutionFilter{},
			wantErr: false,
		},
		{
			name: "valid filter with all fields",
			filter: ExecutionFilter{
				WorkflowID:  "workflow-123",
				Status:      "completed",
				TriggerType: "webhook",
				StartDate:   timePtr(time.Now().Add(-24 * time.Hour)),
				EndDate:     timePtr(time.Now()),
			},
			wantErr: false,
		},
		{
			name: "valid filter with only workflow_id",
			filter: ExecutionFilter{
				WorkflowID: "workflow-123",
			},
			wantErr: false,
		},
		{
			name: "valid filter with only status",
			filter: ExecutionFilter{
				Status: "running",
			},
			wantErr: false,
		},
		{
			name: "valid filter with date range",
			filter: ExecutionFilter{
				StartDate: timePtr(time.Now().Add(-24 * time.Hour)),
				EndDate:   timePtr(time.Now()),
			},
			wantErr: false,
		},
		{
			name: "invalid: end_date before start_date",
			filter: ExecutionFilter{
				StartDate: timePtr(time.Now()),
				EndDate:   timePtr(time.Now().Add(-24 * time.Hour)),
			},
			wantErr: true,
			errMsg:  "end_date must be after start_date",
		},
		{
			name: "invalid: empty workflow_id",
			filter: ExecutionFilter{
				WorkflowID: "",
			},
			wantErr: false, // Empty string is treated as no filter
		},
		{
			name: "valid: start_date only",
			filter: ExecutionFilter{
				StartDate: timePtr(time.Now().Add(-24 * time.Hour)),
			},
			wantErr: false,
		},
		{
			name: "valid: end_date only",
			filter: ExecutionFilter{
				EndDate: timePtr(time.Now()),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecutionFilter.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("ExecutionFilter.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestPaginationCursor_Encode tests cursor encoding
func TestPaginationCursor_Encode(t *testing.T) {
	tests := []struct {
		name   string
		cursor PaginationCursor
	}{
		{
			name: "encode cursor with all fields",
			cursor: PaginationCursor{
				CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				ID:        "execution-123",
			},
		},
		{
			name: "encode cursor with special characters in ID",
			cursor: PaginationCursor{
				CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				ID:        "execution-with-special-chars-!@#$%",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := tt.cursor.Encode()
			if encoded == "" {
				t.Error("PaginationCursor.Encode() returned empty string")
			}
			// Encoded cursor should be base64
			if len(encoded) < 10 {
				t.Error("PaginationCursor.Encode() returned suspiciously short string")
			}
		})
	}
}

// TestPaginationCursor_Decode tests cursor decoding
func TestPaginationCursor_Decode(t *testing.T) {
	// Create a valid cursor and encode it
	originalCursor := PaginationCursor{
		CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		ID:        "execution-123",
	}
	encoded := originalCursor.Encode()

	tests := []struct {
		name       string
		encoded    string
		wantErr    bool
		validateFn func(*testing.T, PaginationCursor)
	}{
		{
			name:    "decode valid cursor",
			encoded: encoded,
			wantErr: false,
			validateFn: func(t *testing.T, cursor PaginationCursor) {
				if cursor.ID != originalCursor.ID {
					t.Errorf("Decoded ID = %v, want %v", cursor.ID, originalCursor.ID)
				}
				if !cursor.CreatedAt.Equal(originalCursor.CreatedAt) {
					t.Errorf("Decoded CreatedAt = %v, want %v", cursor.CreatedAt, originalCursor.CreatedAt)
				}
			},
		},
		{
			name:    "decode invalid base64",
			encoded: "not-valid-base64!@#$",
			wantErr: true,
		},
		{
			name:    "decode invalid json",
			encoded: "YWJjZGVmZw==", // "abcdefg" in base64
			wantErr: true,
		},
		{
			name:    "decode empty string",
			encoded: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor, err := DecodePaginationCursor(tt.encoded)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodePaginationCursor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validateFn != nil {
				tt.validateFn(t, cursor)
			}
		})
	}
}

// TestPaginationCursor_EncodeDecode tests round-trip encoding/decoding
func TestPaginationCursor_EncodeDecode(t *testing.T) {
	tests := []struct {
		name   string
		cursor PaginationCursor
	}{
		{
			name: "round-trip with standard cursor",
			cursor: PaginationCursor{
				CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				ID:        "execution-123",
			},
		},
		{
			name: "round-trip with recent timestamp",
			cursor: PaginationCursor{
				CreatedAt: time.Now(),
				ID:        "execution-456",
			},
		},
		{
			name: "round-trip with UUID",
			cursor: PaginationCursor{
				CreatedAt: time.Now(),
				ID:        "550e8400-e29b-41d4-a716-446655440000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := tt.cursor.Encode()
			decoded, err := DecodePaginationCursor(encoded)
			if err != nil {
				t.Fatalf("DecodePaginationCursor() error = %v", err)
			}
			if decoded.ID != tt.cursor.ID {
				t.Errorf("Round-trip ID = %v, want %v", decoded.ID, tt.cursor.ID)
			}
			// Compare timestamps with some tolerance for rounding
			if decoded.CreatedAt.Sub(tt.cursor.CreatedAt).Abs() > time.Microsecond {
				t.Errorf("Round-trip CreatedAt = %v, want %v", decoded.CreatedAt, tt.cursor.CreatedAt)
			}
		})
	}
}

// TestExecutionListResult_JSON tests JSON marshaling of ExecutionListResult
func TestExecutionListResult_JSON(t *testing.T) {
	now := time.Now()
	executions := []*Execution{
		{
			ID:          "exec-1",
			TenantID:    "tenant-1",
			WorkflowID:  "workflow-1",
			Status:      "completed",
			TriggerType: "webhook",
			CreatedAt:   now,
		},
		{
			ID:          "exec-2",
			TenantID:    "tenant-1",
			WorkflowID:  "workflow-1",
			Status:      "running",
			TriggerType: "schedule",
			CreatedAt:   now.Add(-1 * time.Hour),
		},
	}

	cursor := PaginationCursor{
		CreatedAt: now.Add(-1 * time.Hour),
		ID:        "exec-2",
	}

	result := ExecutionListResult{
		Data:       executions,
		Cursor:     cursor.Encode(),
		HasMore:    true,
		TotalCount: 10,
	}

	// Test marshaling
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Test unmarshaling
	var unmarshaled ExecutionListResult
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(unmarshaled.Data) != len(result.Data) {
		t.Errorf("Unmarshaled Data length = %v, want %v", len(unmarshaled.Data), len(result.Data))
	}
	if unmarshaled.Cursor != result.Cursor {
		t.Errorf("Unmarshaled Cursor = %v, want %v", unmarshaled.Cursor, result.Cursor)
	}
	if unmarshaled.HasMore != result.HasMore {
		t.Errorf("Unmarshaled HasMore = %v, want %v", unmarshaled.HasMore, result.HasMore)
	}
	if unmarshaled.TotalCount != result.TotalCount {
		t.Errorf("Unmarshaled TotalCount = %v, want %v", unmarshaled.TotalCount, result.TotalCount)
	}
}

// TestExecutionWithSteps_Structure tests ExecutionWithSteps structure
func TestExecutionWithSteps_Structure(t *testing.T) {
	now := time.Now()

	execution := &Execution{
		ID:          "exec-1",
		TenantID:    "tenant-1",
		WorkflowID:  "workflow-1",
		Status:      "completed",
		TriggerType: "webhook",
		CreatedAt:   now,
	}

	steps := []*StepExecution{
		{
			ID:          "step-1",
			ExecutionID: "exec-1",
			NodeID:      "node-1",
			NodeType:    "action:http",
			Status:      "completed",
		},
		{
			ID:          "step-2",
			ExecutionID: "exec-1",
			NodeID:      "node-2",
			NodeType:    "action:transform",
			Status:      "completed",
		},
	}

	result := ExecutionWithSteps{
		Execution: execution,
		Steps:     steps,
	}

	// Test that structure can be marshaled to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Test that structure can be unmarshaled from JSON
	var unmarshaled ExecutionWithSteps
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Execution.ID != execution.ID {
		t.Errorf("Unmarshaled Execution.ID = %v, want %v", unmarshaled.Execution.ID, execution.ID)
	}
	if len(unmarshaled.Steps) != len(steps) {
		t.Errorf("Unmarshaled Steps length = %v, want %v", len(unmarshaled.Steps), len(steps))
	}
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
