// Package integrations provides implementations of external service integrations
// and a registration function to add them to the integration registry.
package integrations

import (
	"log/slog"

	"github.com/gorax/gorax/internal/integration"
)

// RegisterAll registers all available integrations with the given registry.
// This function should be called during application initialization.
func RegisterAll(registry *integration.Registry, logger *slog.Logger) error {
	integrations := []integration.Integration{
		NewSlackIntegration(logger),
		NewGitHubIntegration(logger),
		NewJiraIntegration(logger),
		NewPagerDutyIntegration(logger),
		NewAdvancedHTTPIntegration(logger),
	}

	for _, integ := range integrations {
		if err := registry.Register(integ); err != nil {
			return err
		}
	}

	return nil
}

// RegisterAllGlobal registers all available integrations with the global registry.
func RegisterAllGlobal(logger *slog.Logger) error {
	return RegisterAll(integration.GlobalRegistry(), logger)
}

// MustRegisterAll registers all integrations and panics on error.
func MustRegisterAll(registry *integration.Registry, logger *slog.Logger) {
	if err := RegisterAll(registry, logger); err != nil {
		panic("failed to register integrations: " + err.Error())
	}
}

// GetIntegrationNames returns the names of all available integrations.
func GetIntegrationNames() []string {
	return []string{
		"slack",
		"github",
		"jira",
		"pagerduty",
		"http_advanced",
	}
}

// IntegrationInfo provides metadata about available integrations.
type IntegrationInfo struct {
	Name        string
	DisplayName string
	Category    string
	Description string
}

// GetIntegrationInfo returns information about all available integrations.
func GetIntegrationInfo() []IntegrationInfo {
	return []IntegrationInfo{
		{
			Name:        "slack",
			DisplayName: "Slack",
			Category:    "messaging",
			Description: "Send messages, manage channels, and interact with Slack workspaces",
		},
		{
			Name:        "github",
			DisplayName: "GitHub",
			Category:    "version_control",
			Description: "Manage repositories, issues, pull requests, and webhooks on GitHub",
		},
		{
			Name:        "jira",
			DisplayName: "Jira",
			Category:    "project_management",
			Description: "Manage issues, projects, and workflows in Jira",
		},
		{
			Name:        "pagerduty",
			DisplayName: "PagerDuty",
			Category:    "incident_management",
			Description: "Manage incidents, services, and on-call schedules in PagerDuty",
		},
		{
			Name:        "http_advanced",
			DisplayName: "Advanced HTTP",
			Category:    "networking",
			Description: "Advanced HTTP requests with comprehensive authentication, templating, and response parsing",
		},
	}
}
