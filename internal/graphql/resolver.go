package graphql

import (
	"log/slog"

	"github.com/gorax/gorax/internal/schedule"
	"github.com/gorax/gorax/internal/template"
	"github.com/gorax/gorax/internal/webhook"
	"github.com/gorax/gorax/internal/websocket"
	"github.com/gorax/gorax/internal/workflow"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

// Resolver holds service dependencies for GraphQL resolvers
type Resolver struct {
	WorkflowService *workflow.Service
	WebhookService  *webhook.Service
	ScheduleService *schedule.Service
	TemplateService *template.Service
	WebSocketHub    *websocket.Hub
	Logger          *slog.Logger
}
