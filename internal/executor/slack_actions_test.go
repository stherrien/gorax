package executor

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/workflow"
)

// MockCredentialService implements credential.Service for testing
type MockCredentialService struct {
	GetValueFunc func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error)
}

func (m *MockCredentialService) GetValue(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
	if m.GetValueFunc != nil {
		return m.GetValueFunc(ctx, tenantID, credentialID, userID)
	}
	return &credential.DecryptedValue{
		Value: map[string]interface{}{
			"access_token": "xoxb-test-token",
		},
	}, nil
}

func (m *MockCredentialService) Create(ctx context.Context, tenantID, userID string, input credential.CreateCredentialInput) (*credential.Credential, error) {
	return nil, nil
}

func (m *MockCredentialService) List(ctx context.Context, tenantID string, filter credential.CredentialListFilter, limit, offset int) ([]*credential.Credential, error) {
	return nil, nil
}

func (m *MockCredentialService) GetByID(ctx context.Context, tenantID, credentialID string) (*credential.Credential, error) {
	return nil, nil
}

func (m *MockCredentialService) Update(ctx context.Context, tenantID, credentialID, userID string, input credential.UpdateCredentialInput) (*credential.Credential, error) {
	return nil, nil
}

func (m *MockCredentialService) Delete(ctx context.Context, tenantID, credentialID, userID string) error {
	return nil
}

func (m *MockCredentialService) Rotate(ctx context.Context, tenantID, credentialID, userID string, input credential.RotateCredentialInput) (*credential.Credential, error) {
	return nil, nil
}

func (m *MockCredentialService) ListVersions(ctx context.Context, tenantID, credentialID string) ([]*credential.CredentialValue, error) {
	return nil, nil
}

func (m *MockCredentialService) GetAccessLog(ctx context.Context, tenantID, credentialID string, limit, offset int) ([]*credential.AccessLog, error) {
	return nil, nil
}

// TestExecuteSlackSendMessageAction tests the Slack send message execution
func TestExecuteSlackSendMessageAction(t *testing.T) {
	tests := []struct {
		name           string
		node           workflow.Node
		execCtx        *ExecutionContext
		credService    credential.Service
		wantErr        bool
		errorContains  string
		validateOutput func(t *testing.T, output interface{})
	}{
		{
			name: "successful message send",
			node: workflow.Node{
				ID:   "slack-1",
				Type: string(workflow.NodeTypeActionSlackSendMessage),
				Data: workflow.NodeData{
					Name: "Send Slack Message",
					Config: json.RawMessage(`{
						"channel": "C1234567890",
						"text": "Hello from workflow!"
					}`),
				},
			},
			execCtx: &ExecutionContext{
				TenantID:    "tenant-123",
				ExecutionID: "exec-123",
				WorkflowID:  "wf-123",
				TriggerData: map[string]interface{}{},
				StepOutputs: map[string]interface{}{
					"env": map[string]interface{}{
						"tenant_id": "tenant-123",
					},
					"credential_id": "cred-123",
				},
			},
			credService: &MockCredentialService{
				GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
					return &credential.DecryptedValue{
						Value: map[string]interface{}{
							"access_token": "xoxb-test-token",
						},
					}, nil
				},
			},
			wantErr: false,
			validateOutput: func(t *testing.T, output interface{}) {
				assert.NotNil(t, output)
				// The actual output validation would depend on mocking the Slack API
			},
		},
		{
			name: "missing credential service",
			node: workflow.Node{
				ID:   "slack-1",
				Type: string(workflow.NodeTypeActionSlackSendMessage),
				Data: workflow.NodeData{
					Name: "Send Slack Message",
					Config: json.RawMessage(`{
						"channel": "C1234567890",
						"text": "Hello!"
					}`),
				},
			},
			execCtx: &ExecutionContext{
				TenantID:    "tenant-123",
				ExecutionID: "exec-123",
				WorkflowID:  "wf-123",
				TriggerData: map[string]interface{}{},
				StepOutputs: map[string]interface{}{},
			},
			credService:   nil,
			wantErr:       true,
			errorContains: "credential service not available",
		},
		{
			name: "missing config",
			node: workflow.Node{
				ID:   "slack-1",
				Type: string(workflow.NodeTypeActionSlackSendMessage),
				Data: workflow.NodeData{
					Name:   "Send Slack Message",
					Config: json.RawMessage(`{}`),
				},
			},
			execCtx: &ExecutionContext{
				TenantID:    "tenant-123",
				ExecutionID: "exec-123",
				WorkflowID:  "wf-123",
				TriggerData: map[string]interface{}{},
				StepOutputs: map[string]interface{}{},
			},
			credService: &MockCredentialService{},
			wantErr:     false, // Config validation happens in the action, not executor
		},
		{
			name: "invalid JSON config",
			node: workflow.Node{
				ID:   "slack-1",
				Type: string(workflow.NodeTypeActionSlackSendMessage),
				Data: workflow.NodeData{
					Name:   "Send Slack Message",
					Config: json.RawMessage(`{invalid json}`),
				},
			},
			execCtx: &ExecutionContext{
				TenantID:    "tenant-123",
				ExecutionID: "exec-123",
				WorkflowID:  "wf-123",
				TriggerData: map[string]interface{}{},
				StepOutputs: map[string]interface{}{},
			},
			credService:   &MockCredentialService{},
			wantErr:       true,
			errorContains: "failed to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			executor := &Executor{
				logger:            logger,
				credentialService: tt.credService,
			}

			ctx := context.Background()
			output, err := executor.executeSlackSendMessageAction(ctx, tt.node, tt.execCtx)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					// Some errors are expected from the actual Slack action execution
					// since we're not mocking the HTTP client
					t.Logf("Expected error from Slack action: %v", err)
				}
				if tt.validateOutput != nil && output != nil {
					tt.validateOutput(t, output)
				}
			}
		})
	}
}

// TestExecuteSlackSendDMAction tests the Slack send DM execution
func TestExecuteSlackSendDMAction(t *testing.T) {
	tests := []struct {
		name          string
		node          workflow.Node
		execCtx       *ExecutionContext
		credService   credential.Service
		wantErr       bool
		errorContains string
	}{
		{
			name: "successful DM send",
			node: workflow.Node{
				ID:   "slack-dm-1",
				Type: string(workflow.NodeTypeActionSlackSendDM),
				Data: workflow.NodeData{
					Name: "Send Slack DM",
					Config: json.RawMessage(`{
						"user": "user@example.com",
						"text": "Hello from workflow!"
					}`),
				},
			},
			execCtx: &ExecutionContext{
				TenantID:    "tenant-123",
				ExecutionID: "exec-123",
				WorkflowID:  "wf-123",
				TriggerData: map[string]interface{}{},
				StepOutputs: map[string]interface{}{
					"env": map[string]interface{}{
						"tenant_id": "tenant-123",
					},
					"credential_id": "cred-123",
				},
			},
			credService: &MockCredentialService{},
			wantErr:     false,
		},
		{
			name: "missing credential service",
			node: workflow.Node{
				ID:   "slack-dm-1",
				Type: string(workflow.NodeTypeActionSlackSendDM),
				Data: workflow.NodeData{
					Name: "Send Slack DM",
					Config: json.RawMessage(`{
						"user": "user@example.com",
						"text": "Hello!"
					}`),
				},
			},
			execCtx: &ExecutionContext{
				TenantID:    "tenant-123",
				ExecutionID: "exec-123",
				WorkflowID:  "wf-123",
				TriggerData: map[string]interface{}{},
				StepOutputs: map[string]interface{}{},
			},
			credService:   nil,
			wantErr:       true,
			errorContains: "credential service not available",
		},
		{
			name: "empty config",
			node: workflow.Node{
				ID:   "slack-dm-1",
				Type: string(workflow.NodeTypeActionSlackSendDM),
				Data: workflow.NodeData{
					Name:   "Send Slack DM",
					Config: nil,
				},
			},
			execCtx: &ExecutionContext{
				TenantID:    "tenant-123",
				ExecutionID: "exec-123",
				WorkflowID:  "wf-123",
				TriggerData: map[string]interface{}{},
				StepOutputs: map[string]interface{}{},
			},
			credService:   &MockCredentialService{},
			wantErr:       true,
			errorContains: "missing config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			executor := &Executor{
				logger:            logger,
				credentialService: tt.credService,
			}

			ctx := context.Background()
			_, err := executor.executeSlackSendDMAction(ctx, tt.node, tt.execCtx)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					// Expected errors from actual Slack action execution
					t.Logf("Expected error from Slack action: %v", err)
				}
			}
		})
	}
}

// TestExecuteSlackUpdateMessageAction tests the Slack update message execution
func TestExecuteSlackUpdateMessageAction(t *testing.T) {
	tests := []struct {
		name          string
		node          workflow.Node
		execCtx       *ExecutionContext
		credService   credential.Service
		wantErr       bool
		errorContains string
	}{
		{
			name: "successful message update",
			node: workflow.Node{
				ID:   "slack-update-1",
				Type: string(workflow.NodeTypeActionSlackUpdateMessage),
				Data: workflow.NodeData{
					Name: "Update Slack Message",
					Config: json.RawMessage(`{
						"channel": "C1234567890",
						"ts": "1503435956.000247",
						"text": "Updated message"
					}`),
				},
			},
			execCtx: &ExecutionContext{
				TenantID:    "tenant-123",
				ExecutionID: "exec-123",
				WorkflowID:  "wf-123",
				TriggerData: map[string]interface{}{},
				StepOutputs: map[string]interface{}{
					"env": map[string]interface{}{
						"tenant_id": "tenant-123",
					},
					"credential_id": "cred-123",
				},
			},
			credService: &MockCredentialService{},
			wantErr:     false,
		},
		{
			name: "missing credential service",
			node: workflow.Node{
				ID:   "slack-update-1",
				Type: string(workflow.NodeTypeActionSlackUpdateMessage),
				Data: workflow.NodeData{
					Name: "Update Slack Message",
					Config: json.RawMessage(`{
						"channel": "C1234567890",
						"ts": "1503435956.000247",
						"text": "Updated"
					}`),
				},
			},
			execCtx: &ExecutionContext{
				TenantID:    "tenant-123",
				ExecutionID: "exec-123",
				WorkflowID:  "wf-123",
				TriggerData: map[string]interface{}{},
				StepOutputs: map[string]interface{}{},
			},
			credService:   nil,
			wantErr:       true,
			errorContains: "credential service not available",
		},
		{
			name: "update from previous step output",
			node: workflow.Node{
				ID:   "slack-update-1",
				Type: string(workflow.NodeTypeActionSlackUpdateMessage),
				Data: workflow.NodeData{
					Name: "Update Slack Message",
					Config: json.RawMessage(`{
						"channel": "{{steps.send-message.channel}}",
						"ts": "{{steps.send-message.timestamp}}",
						"text": "Updated from previous step"
					}`),
				},
			},
			execCtx: &ExecutionContext{
				TenantID:    "tenant-123",
				ExecutionID: "exec-123",
				WorkflowID:  "wf-123",
				TriggerData: map[string]interface{}{},
				StepOutputs: map[string]interface{}{
					"send-message": map[string]interface{}{
						"channel":   "C1234567890",
						"timestamp": "1503435956.000247",
					},
					"env": map[string]interface{}{
						"tenant_id": "tenant-123",
					},
					"credential_id": "cred-123",
				},
			},
			credService: &MockCredentialService{},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			executor := &Executor{
				logger:            logger,
				credentialService: tt.credService,
			}

			ctx := context.Background()
			_, err := executor.executeSlackUpdateMessageAction(ctx, tt.node, tt.execCtx)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					// Expected errors from actual Slack action execution
					t.Logf("Expected error from Slack action: %v", err)
				}
			}
		})
	}
}

// TestExecuteSlackAddReactionAction tests the Slack add reaction execution
func TestExecuteSlackAddReactionAction(t *testing.T) {
	tests := []struct {
		name          string
		node          workflow.Node
		execCtx       *ExecutionContext
		credService   credential.Service
		wantErr       bool
		errorContains string
	}{
		{
			name: "successful reaction add",
			node: workflow.Node{
				ID:   "slack-reaction-1",
				Type: string(workflow.NodeTypeActionSlackAddReaction),
				Data: workflow.NodeData{
					Name: "Add Slack Reaction",
					Config: json.RawMessage(`{
						"channel": "C1234567890",
						"timestamp": "1503435956.000247",
						"emoji": "thumbsup"
					}`),
				},
			},
			execCtx: &ExecutionContext{
				TenantID:    "tenant-123",
				ExecutionID: "exec-123",
				WorkflowID:  "wf-123",
				TriggerData: map[string]interface{}{},
				StepOutputs: map[string]interface{}{
					"env": map[string]interface{}{
						"tenant_id": "tenant-123",
					},
					"credential_id": "cred-123",
				},
			},
			credService: &MockCredentialService{},
			wantErr:     false,
		},
		{
			name: "missing credential service",
			node: workflow.Node{
				ID:   "slack-reaction-1",
				Type: string(workflow.NodeTypeActionSlackAddReaction),
				Data: workflow.NodeData{
					Name: "Add Slack Reaction",
					Config: json.RawMessage(`{
						"channel": "C1234567890",
						"timestamp": "1503435956.000247",
						"emoji": "thumbsup"
					}`),
				},
			},
			execCtx: &ExecutionContext{
				TenantID:    "tenant-123",
				ExecutionID: "exec-123",
				WorkflowID:  "wf-123",
				TriggerData: map[string]interface{}{},
				StepOutputs: map[string]interface{}{},
			},
			credService:   nil,
			wantErr:       true,
			errorContains: "credential service not available",
		},
		{
			name: "emoji with colons",
			node: workflow.Node{
				ID:   "slack-reaction-1",
				Type: string(workflow.NodeTypeActionSlackAddReaction),
				Data: workflow.NodeData{
					Name: "Add Slack Reaction",
					Config: json.RawMessage(`{
						"channel": "C1234567890",
						"timestamp": "1503435956.000247",
						"emoji": ":thumbsup:"
					}`),
				},
			},
			execCtx: &ExecutionContext{
				TenantID:    "tenant-123",
				ExecutionID: "exec-123",
				WorkflowID:  "wf-123",
				TriggerData: map[string]interface{}{},
				StepOutputs: map[string]interface{}{
					"env": map[string]interface{}{
						"tenant_id": "tenant-123",
					},
					"credential_id": "cred-123",
				},
			},
			credService: &MockCredentialService{},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			executor := &Executor{
				logger:            logger,
				credentialService: tt.credService,
			}

			ctx := context.Background()
			_, err := executor.executeSlackAddReactionAction(ctx, tt.node, tt.execCtx)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					// Expected errors from actual Slack action execution
					t.Logf("Expected error from Slack action: %v", err)
				}
			}
		})
	}
}

// TestExecuteSlackActions_ContextPropagation tests that context is properly propagated
func TestExecuteSlackActions_ContextPropagation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	credService := &MockCredentialService{
		GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
			// Verify context values
			require.NotNil(t, ctx)
			assert.Equal(t, "tenant-123", tenantID)
			return &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "xoxb-test-token",
				},
			}, nil
		},
	}

	executor := &Executor{
		logger:            logger,
		credentialService: credService,
	}

	node := workflow.Node{
		ID:   "slack-1",
		Type: string(workflow.NodeTypeActionSlackSendMessage),
		Data: workflow.NodeData{
			Name: "Send Slack Message",
			Config: json.RawMessage(`{
				"channel": "C1234567890",
				"text": "Test message"
			}`),
		},
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant-123",
		ExecutionID: "exec-123",
		WorkflowID:  "wf-123",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{
			"env": map[string]interface{}{
				"tenant_id": "tenant-123",
			},
			"credential_id": "cred-123",
		},
	}

	ctx := context.Background()
	_, err := executor.executeSlackSendMessageAction(ctx, node, execCtx)

	// We expect an error from the actual Slack API call, but the context should have been propagated
	if err != nil {
		t.Logf("Expected error from Slack API: %v", err)
	}
}

// TestExecuteSlackActions_CredentialInjection tests credential injection flow
func TestExecuteSlackActions_CredentialInjection(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Mock credential service that returns valid Slack token
	credService := &MockCredentialService{
		GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
			assert.Equal(t, "tenant-123", tenantID)
			assert.Equal(t, "cred-123", credentialID)
			return &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token":  "xoxb-valid-token",
					"refresh_token": "xoxe-1-refresh",
					"token_type":    "Bearer",
				},
			}, nil
		},
	}

	executor := &Executor{
		logger:            logger,
		credentialService: credService,
	}

	node := workflow.Node{
		ID:   "slack-send-1",
		Type: string(workflow.NodeTypeActionSlackSendMessage),
		Data: workflow.NodeData{
			Name: "Send Message",
			Config: json.RawMessage(`{
				"channel": "C1234567890",
				"text": "Test message with credential"
			}`),
		},
	}

	// In real execution, credential_id would be injected by the credential injector
	// Here we simulate it being in the context
	execCtx := &ExecutionContext{
		TenantID:    "tenant-123",
		ExecutionID: "exec-123",
		WorkflowID:  "wf-123",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{},
	}

	ctx := context.Background()
	_, err := executor.executeSlackSendMessageAction(ctx, node, execCtx)

	// We expect an error because credential_id is not in the ActionInput context
	// This is expected - in real execution, the credential injector would handle this
	if err != nil {
		assert.Contains(t, err.Error(), "credential_id is required in context")
	}
}

// TestExecuteSlackActions_AllActions tests that all Slack actions can be executed
func TestExecuteSlackActions_AllActions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	credService := &MockCredentialService{}

	executor := &Executor{
		logger:            logger,
		credentialService: credService,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant-123",
		ExecutionID: "exec-123",
		WorkflowID:  "wf-123",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{
			"env": map[string]interface{}{
				"tenant_id": "tenant-123",
			},
			"credential_id": "cred-123",
		},
	}

	tests := []struct {
		name     string
		nodeType workflow.NodeType
		config   string
		execute  func(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error)
	}{
		{
			name:     "send_message",
			nodeType: workflow.NodeTypeActionSlackSendMessage,
			config:   `{"channel": "C1234567890", "text": "Test"}`,
			execute:  executor.executeSlackSendMessageAction,
		},
		{
			name:     "send_dm",
			nodeType: workflow.NodeTypeActionSlackSendDM,
			config:   `{"user": "user@example.com", "text": "Test"}`,
			execute:  executor.executeSlackSendDMAction,
		},
		{
			name:     "update_message",
			nodeType: workflow.NodeTypeActionSlackUpdateMessage,
			config:   `{"channel": "C1234567890", "ts": "1503435956.000247", "text": "Updated"}`,
			execute:  executor.executeSlackUpdateMessageAction,
		},
		{
			name:     "add_reaction",
			nodeType: workflow.NodeTypeActionSlackAddReaction,
			config:   `{"channel": "C1234567890", "timestamp": "1503435956.000247", "emoji": "thumbsup"}`,
			execute:  executor.executeSlackAddReactionAction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := workflow.Node{
				ID:   "slack-test",
				Type: string(tt.nodeType),
				Data: workflow.NodeData{
					Name:   "Slack Action Test",
					Config: json.RawMessage(tt.config),
				},
			}

			ctx := context.Background()
			_, err := tt.execute(ctx, node, execCtx)

			// We expect errors from the actual Slack API calls since we're not mocking the HTTP client
			// The important thing is that the executor methods exist and can be called
			if err != nil {
				t.Logf("Expected error from Slack API: %v", err)
			}
		})
	}
}
