package aibuilder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConversationStatus(t *testing.T) {
	tests := []struct {
		name   string
		status ConversationStatus
		valid  bool
	}{
		{"active status", ConversationStatusActive, true},
		{"completed status", ConversationStatusCompleted, true},
		{"abandoned status", ConversationStatusAbandoned, true},
		{"invalid status", ConversationStatus("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestMessageRole(t *testing.T) {
	tests := []struct {
		name  string
		role  MessageRole
		valid bool
	}{
		{"user role", MessageRoleUser, true},
		{"assistant role", MessageRoleAssistant, true},
		{"system role", MessageRoleSystem, true},
		{"invalid role", MessageRole("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.IsValid()
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestBuildRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request BuildRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: BuildRequest{
				Description: "Create a workflow that sends a Slack message when a webhook is received",
			},
			wantErr: false,
		},
		{
			name: "valid request with context",
			request: BuildRequest{
				Description: "Send email when webhook triggers",
				Context: &BuildContext{
					AvailableCredentials: []string{"slack", "email"},
					AvailableIntegrations: []string{"slack", "email"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid request with constraints",
			request: BuildRequest{
				Description: "Simple webhook handler",
				Constraints: &BuildConstraints{
					MaxNodes:     10,
					AllowedTypes: []string{"trigger:webhook", "action:http"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty description",
			request: BuildRequest{
				Description: "",
			},
			wantErr: true,
			errMsg:  "description is required",
		},
		{
			name: "description too short",
			request: BuildRequest{
				Description: "hi",
			},
			wantErr: true,
			errMsg:  "description must be at least 10 characters",
		},
		{
			name: "invalid max nodes",
			request: BuildRequest{
				Description: "Create a workflow",
				Constraints: &BuildConstraints{
					MaxNodes: -1,
				},
			},
			wantErr: true,
			errMsg:  "max_nodes must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBuildResult_HasWarnings(t *testing.T) {
	tests := []struct {
		name     string
		result   BuildResult
		expected bool
	}{
		{
			name:     "no warnings",
			result:   BuildResult{Warnings: nil},
			expected: false,
		},
		{
			name:     "empty warnings",
			result:   BuildResult{Warnings: []string{}},
			expected: false,
		},
		{
			name:     "has warnings",
			result:   BuildResult{Warnings: []string{"Missing credential"}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.HasWarnings())
		})
	}
}

func TestBuildResult_HasSuggestions(t *testing.T) {
	tests := []struct {
		name     string
		result   BuildResult
		expected bool
	}{
		{
			name:     "no suggestions",
			result:   BuildResult{Suggestions: nil},
			expected: false,
		},
		{
			name:     "empty suggestions",
			result:   BuildResult{Suggestions: []string{}},
			expected: false,
		},
		{
			name:     "has suggestions",
			result:   BuildResult{Suggestions: []string{"Add error handling"}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.HasSuggestions())
		})
	}
}

func TestGeneratedWorkflow_Validate(t *testing.T) {
	tests := []struct {
		name     string
		workflow GeneratedWorkflow
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid workflow",
			workflow: GeneratedWorkflow{
				Name:        "My Workflow",
				Description: "A test workflow",
				Definition: &WorkflowDefinition{
					Nodes: []GeneratedNode{
						{ID: "node1", Type: "trigger:webhook", Name: "Webhook Trigger"},
					},
					Edges: []GeneratedEdge{},
				},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			workflow: GeneratedWorkflow{
				Name: "",
				Definition: &WorkflowDefinition{
					Nodes: []GeneratedNode{{ID: "node1", Type: "trigger:webhook", Name: "Trigger"}},
				},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "nil definition",
			workflow: GeneratedWorkflow{
				Name:       "My Workflow",
				Definition: nil,
			},
			wantErr: true,
			errMsg:  "definition is required",
		},
		{
			name: "empty nodes",
			workflow: GeneratedWorkflow{
				Name: "My Workflow",
				Definition: &WorkflowDefinition{
					Nodes: []GeneratedNode{},
				},
			},
			wantErr: true,
			errMsg:  "at least one node is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.workflow.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConversation_AddMessage(t *testing.T) {
	conv := &Conversation{
		ID:       "conv-123",
		TenantID: "tenant-1",
		UserID:   "user-1",
		Status:   ConversationStatusActive,
		Messages: []ConversationMessage{},
	}

	// Add user message
	msg := conv.AddMessage(MessageRoleUser, "Create a workflow")

	require.Len(t, conv.Messages, 1)
	assert.Equal(t, MessageRoleUser, msg.Role)
	assert.Equal(t, "Create a workflow", msg.Content)
	assert.NotEmpty(t, msg.ID)
	assert.False(t, msg.CreatedAt.IsZero())
}

func TestConversation_GetLastMessage(t *testing.T) {
	t.Run("empty messages", func(t *testing.T) {
		conv := &Conversation{Messages: []ConversationMessage{}}
		msg := conv.GetLastMessage()
		assert.Nil(t, msg)
	})

	t.Run("with messages", func(t *testing.T) {
		conv := &Conversation{
			Messages: []ConversationMessage{
				{ID: "msg1", Content: "First"},
				{ID: "msg2", Content: "Second"},
			},
		}
		msg := conv.GetLastMessage()
		require.NotNil(t, msg)
		assert.Equal(t, "msg2", msg.ID)
		assert.Equal(t, "Second", msg.Content)
	})
}

func TestConversation_GetMessagesByRole(t *testing.T) {
	conv := &Conversation{
		Messages: []ConversationMessage{
			{ID: "msg1", Role: MessageRoleUser, Content: "User 1"},
			{ID: "msg2", Role: MessageRoleAssistant, Content: "Assistant 1"},
			{ID: "msg3", Role: MessageRoleUser, Content: "User 2"},
		},
	}

	userMessages := conv.GetMessagesByRole(MessageRoleUser)
	assert.Len(t, userMessages, 2)

	assistantMessages := conv.GetMessagesByRole(MessageRoleAssistant)
	assert.Len(t, assistantMessages, 1)
}

func TestConversation_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   ConversationStatus
		expected bool
	}{
		{"active", ConversationStatusActive, true},
		{"completed", ConversationStatusCompleted, false},
		{"abandoned", ConversationStatusAbandoned, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := &Conversation{Status: tt.status}
			assert.Equal(t, tt.expected, conv.IsActive())
		})
	}
}

func TestConversation_Complete(t *testing.T) {
	conv := &Conversation{
		Status: ConversationStatusActive,
	}

	conv.Complete()

	assert.Equal(t, ConversationStatusCompleted, conv.Status)
	assert.False(t, conv.UpdatedAt.IsZero())
}

func TestConversation_Abandon(t *testing.T) {
	conv := &Conversation{
		Status: ConversationStatusActive,
	}

	conv.Abandon()

	assert.Equal(t, ConversationStatusAbandoned, conv.Status)
	assert.False(t, conv.UpdatedAt.IsZero())
}

func TestConversation_SetCurrentWorkflow(t *testing.T) {
	conv := &Conversation{
		Status: ConversationStatusActive,
	}

	workflow := &GeneratedWorkflow{
		Name:        "Test Workflow",
		Description: "A test",
		Definition: &WorkflowDefinition{
			Nodes: []GeneratedNode{
				{ID: "node1", Type: "trigger:webhook", Name: "Trigger"},
			},
		},
	}

	conv.SetCurrentWorkflow(workflow)

	require.NotNil(t, conv.CurrentWorkflow)
	assert.Equal(t, "Test Workflow", conv.CurrentWorkflow.Name)
}

func TestNewConversation(t *testing.T) {
	conv := NewConversation("tenant-123", "user-456")

	assert.NotEmpty(t, conv.ID)
	assert.Equal(t, "tenant-123", conv.TenantID)
	assert.Equal(t, "user-456", conv.UserID)
	assert.Equal(t, ConversationStatusActive, conv.Status)
	assert.Empty(t, conv.Messages)
	assert.False(t, conv.CreatedAt.IsZero())
	assert.False(t, conv.UpdatedAt.IsZero())
}

func TestNewBuildResult(t *testing.T) {
	workflow := &GeneratedWorkflow{
		Name: "Test",
		Definition: &WorkflowDefinition{
			Nodes: []GeneratedNode{{ID: "n1", Type: "trigger:webhook", Name: "T"}},
		},
	}

	result := NewBuildResult("conv-123", workflow, "Here's your workflow")

	assert.Equal(t, "conv-123", result.ConversationID)
	assert.Equal(t, workflow, result.Workflow)
	assert.Equal(t, "Here's your workflow", result.Explanation)
	assert.Empty(t, result.Warnings)
	assert.Empty(t, result.Suggestions)
}

func TestRefineRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request RefineRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: RefineRequest{
				ConversationID: "conv-123",
				Message:        "Add error handling to the workflow",
			},
			wantErr: false,
		},
		{
			name: "empty conversation ID",
			request: RefineRequest{
				ConversationID: "",
				Message:        "Add error handling",
			},
			wantErr: true,
			errMsg:  "conversation_id is required",
		},
		{
			name: "empty message",
			request: RefineRequest{
				ConversationID: "conv-123",
				Message:        "",
			},
			wantErr: true,
			errMsg:  "message is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGeneratedNode_Validate(t *testing.T) {
	tests := []struct {
		name    string
		node    GeneratedNode
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid node",
			node: GeneratedNode{
				ID:   "node1",
				Type: "trigger:webhook",
				Name: "Webhook Trigger",
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			node: GeneratedNode{
				ID:   "",
				Type: "trigger:webhook",
				Name: "Trigger",
			},
			wantErr: true,
			errMsg:  "node id is required",
		},
		{
			name: "empty type",
			node: GeneratedNode{
				ID:   "node1",
				Type: "",
				Name: "Trigger",
			},
			wantErr: true,
			errMsg:  "node type is required",
		},
		{
			name: "empty name",
			node: GeneratedNode{
				ID:   "node1",
				Type: "trigger:webhook",
				Name: "",
			},
			wantErr: true,
			errMsg:  "node name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.node.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGeneratedEdge_Validate(t *testing.T) {
	tests := []struct {
		name    string
		edge    GeneratedEdge
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid edge",
			edge: GeneratedEdge{
				ID:     "edge1",
				Source: "node1",
				Target: "node2",
			},
			wantErr: false,
		},
		{
			name: "valid edge with label",
			edge: GeneratedEdge{
				ID:     "edge1",
				Source: "node1",
				Target: "node2",
				Label:  "true",
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			edge: GeneratedEdge{
				ID:     "",
				Source: "node1",
				Target: "node2",
			},
			wantErr: true,
			errMsg:  "edge id is required",
		},
		{
			name: "empty source",
			edge: GeneratedEdge{
				ID:     "edge1",
				Source: "",
				Target: "node2",
			},
			wantErr: true,
			errMsg:  "edge source is required",
		},
		{
			name: "empty target",
			edge: GeneratedEdge{
				ID:     "edge1",
				Source: "node1",
				Target: "",
			},
			wantErr: true,
			errMsg:  "edge target is required",
		},
		{
			name: "self-referencing edge",
			edge: GeneratedEdge{
				ID:     "edge1",
				Source: "node1",
				Target: "node1",
			},
			wantErr: true,
			errMsg:  "edge cannot reference the same node",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.edge.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWorkflowDefinition_Validate(t *testing.T) {
	tests := []struct {
		name    string
		def     WorkflowDefinition
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid definition",
			def: WorkflowDefinition{
				Nodes: []GeneratedNode{
					{ID: "node1", Type: "trigger:webhook", Name: "Trigger"},
					{ID: "node2", Type: "action:http", Name: "HTTP Call"},
				},
				Edges: []GeneratedEdge{
					{ID: "edge1", Source: "node1", Target: "node2"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty nodes",
			def: WorkflowDefinition{
				Nodes: []GeneratedNode{},
			},
			wantErr: true,
			errMsg:  "at least one node is required",
		},
		{
			name: "invalid node",
			def: WorkflowDefinition{
				Nodes: []GeneratedNode{
					{ID: "", Type: "trigger:webhook", Name: "Trigger"},
				},
			},
			wantErr: true,
			errMsg:  "node id is required",
		},
		{
			name: "invalid edge",
			def: WorkflowDefinition{
				Nodes: []GeneratedNode{
					{ID: "node1", Type: "trigger:webhook", Name: "Trigger"},
				},
				Edges: []GeneratedEdge{
					{ID: "", Source: "node1", Target: "node2"},
				},
			},
			wantErr: true,
			errMsg:  "edge id is required",
		},
		{
			name: "edge references non-existent source",
			def: WorkflowDefinition{
				Nodes: []GeneratedNode{
					{ID: "node1", Type: "trigger:webhook", Name: "Trigger"},
				},
				Edges: []GeneratedEdge{
					{ID: "edge1", Source: "node99", Target: "node1"},
				},
			},
			wantErr: true,
			errMsg:  "edge source node not found",
		},
		{
			name: "edge references non-existent target",
			def: WorkflowDefinition{
				Nodes: []GeneratedNode{
					{ID: "node1", Type: "trigger:webhook", Name: "Trigger"},
				},
				Edges: []GeneratedEdge{
					{ID: "edge1", Source: "node1", Target: "node99"},
				},
			},
			wantErr: true,
			errMsg:  "edge target node not found",
		},
		{
			name: "duplicate node IDs",
			def: WorkflowDefinition{
				Nodes: []GeneratedNode{
					{ID: "node1", Type: "trigger:webhook", Name: "Trigger"},
					{ID: "node1", Type: "action:http", Name: "HTTP Call"},
				},
			},
			wantErr: true,
			errMsg:  "duplicate node id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.def.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConversationMessage_UpdateTime(t *testing.T) {
	msg := ConversationMessage{
		ID:        "msg-1",
		Role:      MessageRoleUser,
		Content:   "Test",
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}

	before := msg.CreatedAt
	time.Sleep(10 * time.Millisecond)

	// Messages are immutable after creation - just verify timestamp
	assert.Equal(t, before, msg.CreatedAt)
}

func TestBuildContext_HasCredential(t *testing.T) {
	ctx := &BuildContext{
		AvailableCredentials: []string{"slack", "email", "github"},
	}

	assert.True(t, ctx.HasCredential("slack"))
	assert.True(t, ctx.HasCredential("email"))
	assert.False(t, ctx.HasCredential("jira"))
}

func TestBuildContext_HasIntegration(t *testing.T) {
	ctx := &BuildContext{
		AvailableIntegrations: []string{"slack", "email", "github"},
	}

	assert.True(t, ctx.HasIntegration("slack"))
	assert.True(t, ctx.HasIntegration("email"))
	assert.False(t, ctx.HasIntegration("jira"))
}

func TestBuildConstraints_IsTypeAllowed(t *testing.T) {
	t.Run("no restrictions", func(t *testing.T) {
		constraints := &BuildConstraints{}
		assert.True(t, constraints.IsTypeAllowed("trigger:webhook"))
		assert.True(t, constraints.IsTypeAllowed("action:http"))
	})

	t.Run("with allowed types", func(t *testing.T) {
		constraints := &BuildConstraints{
			AllowedTypes: []string{"trigger:webhook", "action:http"},
		}
		assert.True(t, constraints.IsTypeAllowed("trigger:webhook"))
		assert.True(t, constraints.IsTypeAllowed("action:http"))
		assert.False(t, constraints.IsTypeAllowed("action:email"))
	})
}
