package template

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// GetBuiltinTemplates returns all built-in workflow templates
func GetBuiltinTemplates() []*Template {
	now := time.Now()
	templates := []*Template{
		// DevOps Templates
		createCICDNotificationTemplate(now),
		createDeploymentApprovalTemplate(now),
		createInfraAlertHandlerTemplate(now),
		createGitHubPRAutomationTemplate(now),
		createKubernetesDeploymentMonitorTemplate(now),
		createContainerRegistryCleanupTemplate(now),

		// Business Process Templates
		createCustomerOnboardingTemplate(now),
		createApprovalRequestTemplate(now),
		createDocumentProcessingTemplate(now),

		// Integration Templates
		createSlackToJiraTemplate(now),
		createEmailToTaskTemplate(now),
		createCalendarReminderTemplate(now),
		createSlackAlertNotificationTemplate(now),
		createMultiStepUserOnboardingTemplate(now),
		createMultiStepApprovalWorkflowTemplate(now),
		createAPIOrchestrationWorkflowTemplate(now),
		createSalesforceLeadSyncTemplate(now),
		createServiceNowIncidentTemplate(now),

		// Data Processing Templates
		createDataSyncTemplate(now),
		createReportGenerationTemplate(now),
		createDataValidationTemplate(now),
		createDataBackupWorkflowTemplate(now),
		createScheduledReportGenerationTemplate(now),
		createDataETLSyncWorkflowTemplate(now),
		createScheduledReportingWorkflowTemplate(now),

		// Monitoring Templates
		createErrorMonitoringWorkflowTemplate(now),
		createAPIHealthCheckWorkflowTemplate(now),
		createErrorNotificationWorkflowTemplate(now),

		// Security Templates
		createSecurityAlertTriageTemplate(now),
		createVulnerabilityResponseTemplate(now),
		createAccessReviewAutomationTemplate(now),
	}

	return templates
}

// GetTemplateByName returns a template by name, or nil if not found
func GetTemplateByName(name string) *Template {
	templates := GetBuiltinTemplates()
	for _, tmpl := range templates {
		if tmpl.Name == name {
			return tmpl
		}
	}
	return nil
}

// GetTemplatesByCategory returns all templates in a category
func GetTemplatesByCategory(category string) []*Template {
	templates := GetBuiltinTemplates()
	result := make([]*Template, 0)
	for _, tmpl := range templates {
		if tmpl.Category == category {
			result = append(result, tmpl)
		}
	}
	return result
}

// GetTemplatesByTag returns all templates with a specific tag
func GetTemplatesByTag(tag string) []*Template {
	templates := GetBuiltinTemplates()
	result := make([]*Template, 0)
	for _, tmpl := range templates {
		for _, t := range tmpl.Tags {
			if t == tag {
				result = append(result, tmpl)
				break
			}
		}
	}
	return result
}

// DevOps Templates

func createCICDNotificationTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "GitHub Webhook",
					"config": map[string]interface{}{
						"path":      "/webhooks/github-cicd",
						"auth_type": "signature",
						"secret":    "${env.GITHUB_WEBHOOK_SECRET}",
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Extract Pipeline Data",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"repo":       "${trigger.repository.name}",
							"branch":     "${trigger.ref}",
							"status":     "${trigger.status}",
							"commit":     "${trigger.head_commit.message}",
							"author":     "${trigger.head_commit.author.name}",
							"build_url":  "${trigger.check_suite.url}",
							"conclusion": "${trigger.check_suite.conclusion}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "condition-1",
				"type": "control:if",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Check Status",
					"config": map[string]interface{}{
						"condition":   "${steps.transform-1.conclusion} == 'failure'",
						"description": "Only notify on failures",
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 700,
					"y": 50,
				},
				"data": map[string]interface{}{
					"name": "Send Failure Notification",
					"config": map[string]interface{}{
						"channel": "#deployments",
						"message": "🔴 Pipeline Failed\nRepo: ${steps.transform-1.repo}\nBranch: ${steps.transform-1.branch}\nCommit: ${steps.transform-1.commit}\nAuthor: ${steps.transform-1.author}\nURL: ${steps.transform-1.build_url}",
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-2",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 700,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Send Success Notification",
					"config": map[string]interface{}{
						"channel": "#deployments",
						"message": "✅ Pipeline Succeeded\nRepo: ${steps.transform-1.repo}\nBranch: ${steps.transform-1.branch}\nCommit: ${steps.transform-1.commit}",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "transform-1",
				"target": "condition-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "condition-1",
				"target": "slack-1",
				"label":  "true",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "condition-1",
				"target": "slack-2",
				"label":  "false",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "CI/CD Pipeline Notification",
		Description: "Receive GitHub webhook events and send Slack notifications on pipeline success or failure. Includes conditional logic to send different messages based on build status.",
		Category:    string(CategoryDevOps),
		Definition:  defJSON,
		Tags:        []string{"cicd", "github", "slack", "notification", "devops"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createDeploymentApprovalTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Deployment Request",
					"config": map[string]interface{}{
						"path":      "/webhooks/deployment-request",
						"auth_type": "api_key",
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Request Approval",
					"config": map[string]interface{}{
						"channel": "#deployments",
						"message": "🚀 Deployment Approval Required\nEnvironment: ${trigger.environment}\nVersion: ${trigger.version}\nRequested by: ${trigger.requester}\n\nReply with 'approve' or 'reject'",
					},
				},
			},
			map[string]interface{}{
				"id":   "delay-1",
				"type": "control:delay",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Wait for Response",
					"config": map[string]interface{}{
						"duration": "5m",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Trigger Deployment",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${trigger.deployment_url}",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${env.DEPLOYMENT_TOKEN}",
							"Content-Type":  "application/json",
						},
						"body": map[string]interface{}{
							"environment": "${trigger.environment}",
							"version":     "${trigger.version}",
							"approved_by": "${steps.slack-1.response.user}",
						},
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "slack-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "slack-1",
				"target": "delay-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "delay-1",
				"target": "http-1",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Deployment Approval Workflow",
		Description: "Request deployment approval via Slack, wait for response, and trigger deployment upon approval. Includes timeout handling and notification to deployment system.",
		Category:    string(CategoryDevOps),
		Definition:  defJSON,
		Tags:        []string{"deployment", "approval", "slack", "devops", "automation"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createInfraAlertHandlerTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Infrastructure Alert",
					"config": map[string]interface{}{
						"path":      "/webhooks/infra-alert",
						"auth_type": "signature",
					},
				},
			},
			map[string]interface{}{
				"id":   "condition-1",
				"type": "control:if",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Check Severity",
					"config": map[string]interface{}{
						"condition":   "${trigger.severity} == 'critical'",
						"description": "Only auto-remediate critical alerts",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 500,
					"y": 50,
				},
				"data": map[string]interface{}{
					"name": "Auto Remediation",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.REMEDIATION_API}/remediate",
						"body": map[string]interface{}{
							"alert_id": "${trigger.id}",
							"action":   "restart_service",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 700,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Notify Team",
					"config": map[string]interface{}{
						"channel": "#alerts",
						"message": "⚠️ Infrastructure Alert\nSeverity: ${trigger.severity}\nService: ${trigger.service}\nMessage: ${trigger.message}\nRemediation: ${steps.http-1.status}",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "condition-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "condition-1",
				"target": "http-1",
				"label":  "true",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "http-1",
				"target": "slack-1",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "condition-1",
				"target": "slack-1",
				"label":  "false",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Infrastructure Alert Handler",
		Description: "Handle infrastructure monitoring alerts with automatic remediation for critical issues and team notifications. Integrates with monitoring systems and incident management.",
		Category:    string(CategoryDevOps),
		Definition:  defJSON,
		Tags:        []string{"monitoring", "alerts", "remediation", "devops", "infrastructure"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Business Process Templates

func createCustomerOnboardingTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "New Customer Signup",
					"config": map[string]interface{}{
						"path": "/webhooks/customer-signup",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Create Account",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.CRM_API}/accounts",
						"body": map[string]interface{}{
							"name":  "${trigger.company_name}",
							"email": "${trigger.email}",
							"plan":  "${trigger.plan}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "email-1",
				"type": "action:email",
				"position": map[string]interface{}{
					"x": 300,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Send Welcome Email",
					"config": map[string]interface{}{
						"to":      "${trigger.email}",
						"subject": "Welcome to Our Platform!",
						"body":    "Hi ${trigger.first_name}, welcome to our platform! Your account has been created.",
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 500,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Notify Sales Team",
					"config": map[string]interface{}{
						"channel": "#sales",
						"message": "🎉 New customer signup!\nCompany: ${trigger.company_name}\nPlan: ${trigger.plan}\nEmail: ${trigger.email}",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "http-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "trigger-1",
				"target": "email-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "http-1",
				"target": "slack-1",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Customer Onboarding Workflow",
		Description: "Automate customer onboarding by creating CRM accounts, sending welcome emails, and notifying the sales team. Streamlines the signup process with parallel actions.",
		Category:    string(CategoryIntegration),
		Definition:  defJSON,
		Tags:        []string{"onboarding", "business", "crm", "email", "sales"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createApprovalRequestTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Approval Request",
					"config": map[string]interface{}{
						"path": "/webhooks/approval-request",
					},
				},
			},
			map[string]interface{}{
				"id":   "email-1",
				"type": "action:email",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Send Approval Email",
					"config": map[string]interface{}{
						"to":      "${trigger.approver_email}",
						"subject": "Approval Required: ${trigger.request_title}",
						"body":    "Please review and approve: ${trigger.description}\n\nClick to approve: ${trigger.approval_url}",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Update Status",
					"config": map[string]interface{}{
						"method": "PATCH",
						"url":    "${trigger.callback_url}",
						"body": map[string]interface{}{
							"status":    "pending_approval",
							"sent_to":   "${trigger.approver_email}",
							"sent_at":   "${now()}",
							"workflow":  "approval_workflow",
							"requestId": "${trigger.request_id}",
						},
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "email-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "email-1",
				"target": "http-1",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Approval Request Workflow",
		Description: "Send approval request emails to designated approvers and update request status in your system. Tracks approval workflow state and maintains audit trail.",
		Category:    string(CategoryIntegration),
		Definition:  defJSON,
		Tags:        []string{"approval", "business", "email", "workflow"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createDocumentProcessingTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Document Upload",
					"config": map[string]interface{}{
						"path": "/webhooks/document-upload",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Extract Text",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.OCR_API}/extract",
						"body": map[string]interface{}{
							"document_url": "${trigger.document_url}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Parse Data",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"invoice_number": "${steps.http-1.extracted.invoice_number}",
							"amount":         "${steps.http-1.extracted.total_amount}",
							"date":           "${steps.http-1.extracted.date}",
							"vendor":         "${steps.http-1.extracted.vendor_name}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "http-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Store in Database",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.API_URL}/invoices",
						"body": map[string]interface{}{
							"invoice_number": "${steps.transform-1.invoice_number}",
							"amount":         "${steps.transform-1.amount}",
							"date":           "${steps.transform-1.date}",
							"vendor":         "${steps.transform-1.vendor}",
							"document_url":   "${trigger.document_url}",
						},
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "http-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "http-1",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "transform-1",
				"target": "http-2",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Document Processing Pipeline",
		Description: "Process uploaded documents with OCR extraction, data transformation, and database storage. Ideal for invoice processing, form handling, and document digitization.",
		Category:    string(CategoryDataOps),
		Definition:  defJSON,
		Tags:        []string{"document", "ocr", "processing", "business", "automation"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Integration Templates

func createSlackToJiraTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Slack Message",
					"config": map[string]interface{}{
						"path":      "/webhooks/slack-command",
						"auth_type": "signature",
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Parse Command",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"summary":     "${trigger.text}",
							"reporter":    "${trigger.user_name}",
							"channel":     "${trigger.channel_name}",
							"description": "Created from Slack by ${trigger.user_name}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Create Jira Ticket",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.JIRA_URL}/rest/api/3/issue",
						"headers": map[string]interface{}{
							"Authorization": "Basic ${env.JIRA_API_TOKEN}",
							"Content-Type":  "application/json",
						},
						"body": map[string]interface{}{
							"fields": map[string]interface{}{
								"project": map[string]interface{}{
									"key": "${env.JIRA_PROJECT_KEY}",
								},
								"summary":     "${steps.transform-1.summary}",
								"description": "${steps.transform-1.description}",
								"issuetype": map[string]interface{}{
									"name": "Task",
								},
							},
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 700,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Confirm Creation",
					"config": map[string]interface{}{
						"channel": "${trigger.channel_id}",
						"message": "✅ Jira ticket created: ${steps.http-1.response.key}\nURL: ${env.JIRA_URL}/browse/${steps.http-1.response.key}",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "transform-1",
				"target": "http-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "http-1",
				"target": "slack-1",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Slack to Jira Ticket Creator",
		Description: "Create Jira tickets directly from Slack messages using slash commands. Automatically populates ticket fields and confirms creation in Slack.",
		Category:    string(CategoryIntegration),
		Definition:  defJSON,
		Tags:        []string{"slack", "jira", "integration", "ticket", "automation"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createEmailToTaskTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Incoming Email",
					"config": map[string]interface{}{
						"path": "/webhooks/email",
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Extract Task Info",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"title":       "${trigger.subject}",
							"description": "${trigger.body}",
							"assignee":    "${trigger.from}",
							"priority":    "normal",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Create Task",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.TASK_API}/tasks",
						"body": map[string]interface{}{
							"title":       "${steps.transform-1.title}",
							"description": "${steps.transform-1.description}",
							"assignee":    "${steps.transform-1.assignee}",
							"priority":    "${steps.transform-1.priority}",
							"source":      "email",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "email-1",
				"type": "action:email",
				"position": map[string]interface{}{
					"x": 700,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Send Confirmation",
					"config": map[string]interface{}{
						"to":      "${trigger.from}",
						"subject": "Task Created: ${steps.transform-1.title}",
						"body":    "Your task has been created successfully.\nTask ID: ${steps.http-1.response.id}",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "transform-1",
				"target": "http-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "http-1",
				"target": "email-1",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Email to Task Converter",
		Description: "Convert incoming emails into tasks in your task management system. Automatically extracts information and sends confirmation to the sender.",
		Category:    string(CategoryIntegration),
		Definition:  defJSON,
		Tags:        []string{"email", "task", "integration", "automation"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createCalendarReminderTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:schedule",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Daily Check",
					"config": map[string]interface{}{
						"cron":     "0 9 * * *",
						"timezone": "America/New_York",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Fetch Today's Events",
					"config": map[string]interface{}{
						"method": "GET",
						"url":    "${env.CALENDAR_API}/events/today",
					},
				},
			},
			map[string]interface{}{
				"id":   "loop-1",
				"type": "control:loop",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Process Events",
					"config": map[string]interface{}{
						"source":         "${steps.http-1.response.events}",
						"item_variable":  "event",
						"index_variable": "index",
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_dm",
				"position": map[string]interface{}{
					"x": 700,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Send Reminder",
					"config": map[string]interface{}{
						"user":    "${event.attendee_slack_id}",
						"message": "📅 Reminder: ${event.title}\nTime: ${event.start_time}\nLocation: ${event.location}",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "http-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "http-1",
				"target": "loop-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "loop-1",
				"target": "slack-1",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Calendar Event Reminder",
		Description: "Send daily Slack reminders for calendar events. Fetches today's events and sends personalized reminders to each attendee.",
		Category:    string(CategoryIntegration),
		Definition:  defJSON,
		Tags:        []string{"calendar", "slack", "reminder", "schedule", "automation"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Data Processing Templates

func createDataSyncTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:schedule",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Hourly Sync",
					"config": map[string]interface{}{
						"cron": "0 * * * *",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Fetch from Source API",
					"config": map[string]interface{}{
						"method": "GET",
						"url":    "${env.SOURCE_API_URL}/data",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${env.SOURCE_API_KEY}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Transform Data",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"records":   "${steps.http-1.response.data}",
							"synced_at": "${now()}",
							"count":     "${len(steps.http-1.response.data)}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "http-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Sync to Database",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.TARGET_API_URL}/sync",
						"body": map[string]interface{}{
							"records":   "${steps.transform-1.records}",
							"synced_at": "${steps.transform-1.synced_at}",
						},
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "http-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "http-1",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "transform-1",
				"target": "http-2",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Data Sync Workflow",
		Description: "Periodically sync data from an external API to your database. Includes data transformation and scheduled execution.",
		Category:    string(CategoryDataOps),
		Definition:  defJSON,
		Tags:        []string{"sync", "database", "schedule", "api", "dataops"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createReportGenerationTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:schedule",
				"position": map[string]interface{}{
					"x": 100,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Weekly Report",
					"config": map[string]interface{}{
						"cron":     "0 9 * * 1",
						"timezone": "America/New_York",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Fetch Sales Data",
					"config": map[string]interface{}{
						"method": "GET",
						"url":    "${env.API_URL}/sales/weekly",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Fetch Customer Data",
					"config": map[string]interface{}{
						"method": "GET",
						"url":    "${env.API_URL}/customers/weekly",
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 500,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Compile Report",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"total_sales":   "${steps.http-1.response.total}",
							"new_customers": "${steps.http-2.response.count}",
							"average_order": "${steps.http-1.response.average}",
							"growth_rate":   "${steps.http-1.response.growth}",
							"report_week":   "${now()}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "email-1",
				"type": "action:email",
				"position": map[string]interface{}{
					"x": 700,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Email Report",
					"config": map[string]interface{}{
						"to":      "${env.REPORT_RECIPIENTS}",
						"subject": "Weekly Business Report - ${steps.transform-1.report_week}",
						"body":    "Weekly Report:\n\nTotal Sales: $${steps.transform-1.total_sales}\nNew Customers: ${steps.transform-1.new_customers}\nAverage Order: $${steps.transform-1.average_order}\nGrowth Rate: ${steps.transform-1.growth_rate}%",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "http-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "trigger-1",
				"target": "http-2",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "http-1",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "http-2",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e5",
				"source": "transform-1",
				"target": "email-1",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Report Generation Workflow",
		Description: "Generate and email weekly business reports by aggregating data from multiple sources. Includes parallel data fetching and report compilation.",
		Category:    string(CategoryDataOps),
		Definition:  defJSON,
		Tags:        []string{"report", "analytics", "email", "schedule", "dataops"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createDataValidationTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Data Import",
					"config": map[string]interface{}{
						"path": "/webhooks/data-import",
					},
				},
			},
			map[string]interface{}{
				"id":   "formula-1",
				"type": "action:formula",
				"position": map[string]interface{}{
					"x": 300,
					"y": 50,
				},
				"data": map[string]interface{}{
					"name": "Validate Email",
					"config": map[string]interface{}{
						"expression":      "REGEX_MATCH(trigger.email, '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$')",
						"output_variable": "email_valid",
					},
				},
			},
			map[string]interface{}{
				"id":   "formula-2",
				"type": "action:formula",
				"position": map[string]interface{}{
					"x": 300,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Validate Age",
					"config": map[string]interface{}{
						"expression":      "trigger.age >= 18 AND trigger.age <= 120",
						"output_variable": "age_valid",
					},
				},
			},
			map[string]interface{}{
				"id":   "condition-1",
				"type": "control:if",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Check All Valid",
					"config": map[string]interface{}{
						"condition":   "${steps.formula-1.email_valid} AND ${steps.formula-2.age_valid}",
						"description": "Ensure all validations pass",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 50,
				},
				"data": map[string]interface{}{
					"name": "Store Valid Data",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.API_URL}/data",
						"body":   "${trigger}",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Log Invalid Data",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.API_URL}/validation-errors",
						"body": map[string]interface{}{
							"data":          "${trigger}",
							"email_valid":   "${steps.formula-1.email_valid}",
							"age_valid":     "${steps.formula-2.age_valid}",
							"validation_at": "${now()}",
						},
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "formula-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "trigger-1",
				"target": "formula-2",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "formula-1",
				"target": "condition-1",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "formula-2",
				"target": "condition-1",
			},
			map[string]interface{}{
				"id":     "e5",
				"source": "condition-1",
				"target": "http-1",
				"label":  "true",
			},
			map[string]interface{}{
				"id":     "e6",
				"source": "condition-1",
				"target": "http-2",
				"label":  "false",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Data Validation Pipeline",
		Description: "Validate incoming data using multiple rules (email format, age range, etc.) and route to appropriate destinations based on validation results.",
		Category:    string(CategoryDataOps),
		Definition:  defJSON,
		Tags:        []string{"validation", "dataops", "quality", "conditional"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// New Templates

func createSlackAlertNotificationTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Alert Trigger",
					"config": map[string]interface{}{
						"path":      "/webhooks/alert",
						"auth_type": "api_key",
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Format Alert Message",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"severity":      "${trigger.severity}",
							"service":       "${trigger.service}",
							"message":       "${trigger.message}",
							"timestamp":     "${trigger.timestamp}",
							"environment":   "${trigger.environment}",
							"formatted_msg": "🚨 Alert: ${trigger.severity}\nService: ${trigger.service}\nEnvironment: ${trigger.environment}\nMessage: ${trigger.message}\nTime: ${trigger.timestamp}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "condition-1",
				"type": "control:if",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Check Severity Level",
					"config": map[string]interface{}{
						"condition":   "${steps.transform-1.severity} == 'critical' || ${steps.transform-1.severity} == 'high'",
						"description": "Alert for critical and high severity only",
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 700,
					"y": 50,
				},
				"data": map[string]interface{}{
					"name": "Send to Urgent Channel",
					"config": map[string]interface{}{
						"channel": "#alerts-urgent",
						"message": "${steps.transform-1.formatted_msg}",
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-2",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 700,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Send to General Channel",
					"config": map[string]interface{}{
						"channel": "#alerts",
						"message": "${steps.transform-1.formatted_msg}",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "transform-1",
				"target": "condition-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "condition-1",
				"target": "slack-1",
				"label":  "true",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "condition-1",
				"target": "slack-2",
				"label":  "false",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Slack Alert Notification",
		Description: "Send alert notifications to Slack channels based on severity level. Routes critical alerts to urgent channel and lower severity to general channel.",
		Category:    string(CategoryIntegration),
		Definition:  defJSON,
		Tags:        []string{"slack", "alert", "notification", "monitoring"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createGitHubPRAutomationTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "GitHub PR Event",
					"config": map[string]interface{}{
						"path":      "/webhooks/github-pr",
						"auth_type": "signature",
						"secret":    "${env.GITHUB_WEBHOOK_SECRET}",
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 300,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Extract PR Data",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"pr_number":   "${trigger.pull_request.number}",
							"pr_title":    "${trigger.pull_request.title}",
							"pr_body":     "${trigger.pull_request.body}",
							"author":      "${trigger.pull_request.user.login}",
							"repo":        "${trigger.repository.full_name}",
							"action":      "${trigger.action}",
							"files_count": "${len(trigger.pull_request.changed_files)}",
							"labels":      "${trigger.pull_request.labels}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "condition-1",
				"type": "control:if",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Check if New PR",
					"config": map[string]interface{}{
						"condition":   "${steps.transform-1.action} == 'opened'",
						"description": "Only process newly opened PRs",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 50,
				},
				"data": map[string]interface{}{
					"name": "Auto-label PR",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "https://api.github.com/repos/${steps.transform-1.repo}/issues/${steps.transform-1.pr_number}/labels",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${env.GITHUB_TOKEN}",
							"Accept":        "application/vnd.github.v3+json",
						},
						"body": map[string]interface{}{
							"labels": []string{"needs-review", "automated"},
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "http-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Add Welcome Comment",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "https://api.github.com/repos/${steps.transform-1.repo}/issues/${steps.transform-1.pr_number}/comments",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${env.GITHUB_TOKEN}",
							"Accept":        "application/vnd.github.v3+json",
						},
						"body": map[string]interface{}{
							"body": "Thank you for your contribution @${steps.transform-1.author}! 🎉\n\nA team member will review your PR soon. Make sure all tests pass and the code follows our guidelines.",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 900,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Notify Team",
					"config": map[string]interface{}{
						"channel": "#pull-requests",
						"message": "📝 New PR Opened\nRepo: ${steps.transform-1.repo}\nPR #${steps.transform-1.pr_number}: ${steps.transform-1.pr_title}\nAuthor: @${steps.transform-1.author}\nFiles changed: ${steps.transform-1.files_count}",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "transform-1",
				"target": "condition-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "condition-1",
				"target": "http-1",
				"label":  "true",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "http-1",
				"target": "http-2",
			},
			map[string]interface{}{
				"id":     "e5",
				"source": "http-2",
				"target": "slack-1",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "GitHub PR Automation",
		Description: "Automatically label new GitHub pull requests, add welcome comments, and notify team in Slack. Streamlines PR review process with automation.",
		Category:    string(CategoryDevOps),
		Definition:  defJSON,
		Tags:        []string{"github", "pr", "automation", "devops", "code-review"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createDataBackupWorkflowTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:schedule",
				"position": map[string]interface{}{
					"x": 100,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Daily Backup Schedule",
					"config": map[string]interface{}{
						"cron":     "0 2 * * *",
						"timezone": "UTC",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Trigger Database Backup",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.BACKUP_API_URL}/backup/database",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${env.BACKUP_API_TOKEN}",
							"Content-Type":  "application/json",
						},
						"body": map[string]interface{}{
							"database":   "${env.DATABASE_NAME}",
							"type":       "full",
							"encryption": true,
							"retention":  30,
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "http-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Backup Files to S3",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.BACKUP_API_URL}/backup/files",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${env.BACKUP_API_TOKEN}",
						},
						"body": map[string]interface{}{
							"source":      "${env.FILES_PATH}",
							"destination": "s3://${env.S3_BACKUP_BUCKET}/backups/${date()}",
							"compression": "gzip",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 500,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Compile Backup Report",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"db_backup_status":    "${steps.http-1.response.status}",
							"db_backup_size":      "${steps.http-1.response.size}",
							"files_backup_status": "${steps.http-2.response.status}",
							"files_count":         "${steps.http-2.response.files_count}",
							"timestamp":           "${now()}",
							"success":             "${steps.http-1.response.status == 'success' && steps.http-2.response.status == 'success'}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "condition-1",
				"type": "control:if",
				"position": map[string]interface{}{
					"x": 700,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Check Backup Success",
					"config": map[string]interface{}{
						"condition":   "${steps.transform-1.success}",
						"description": "Check if all backups completed successfully",
					},
				},
			},
			map[string]interface{}{
				"id":   "email-1",
				"type": "action:email",
				"position": map[string]interface{}{
					"x": 900,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Send Success Notification",
					"config": map[string]interface{}{
						"to":      "${env.ADMIN_EMAIL}",
						"subject": "✅ Daily Backup Completed Successfully",
						"body":    "Backup completed at ${steps.transform-1.timestamp}\n\nDatabase: ${steps.transform-1.db_backup_size}\nFiles: ${steps.transform-1.files_count} files backed up",
					},
				},
			},
			map[string]interface{}{
				"id":   "email-2",
				"type": "action:email",
				"position": map[string]interface{}{
					"x": 900,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Send Failure Alert",
					"config": map[string]interface{}{
						"to":      "${env.ADMIN_EMAIL}",
						"subject": "❌ URGENT: Backup Failed",
						"body":    "Backup failed at ${steps.transform-1.timestamp}\n\nDatabase status: ${steps.transform-1.db_backup_status}\nFiles status: ${steps.transform-1.files_backup_status}\n\nPlease investigate immediately!",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "http-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "trigger-1",
				"target": "http-2",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "http-1",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "http-2",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e5",
				"source": "transform-1",
				"target": "condition-1",
			},
			map[string]interface{}{
				"id":     "e6",
				"source": "condition-1",
				"target": "email-1",
				"label":  "true",
			},
			map[string]interface{}{
				"id":     "e7",
				"source": "condition-1",
				"target": "email-2",
				"label":  "false",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Data Backup Workflow",
		Description: "Automated daily backup workflow for databases and files with parallel execution, status monitoring, and email notifications on success or failure.",
		Category:    string(CategoryDataOps),
		Definition:  defJSON,
		Tags:        []string{"backup", "schedule", "dataops", "automation", "monitoring"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createErrorMonitoringWorkflowTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Error Event",
					"config": map[string]interface{}{
						"path":      "/webhooks/error",
						"auth_type": "api_key",
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 300,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Parse Error Details",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"error_type":    "${trigger.error.type}",
							"error_message": "${trigger.error.message}",
							"stack_trace":   "${trigger.error.stack_trace}",
							"service":       "${trigger.service}",
							"environment":   "${trigger.environment}",
							"user_id":       "${trigger.user_id}",
							"timestamp":     "${trigger.timestamp}",
							"severity":      "${trigger.severity || 'medium'}",
							"occurrences":   "${trigger.count || 1}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "condition-1",
				"type": "control:if",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Check if Critical Error",
					"config": map[string]interface{}{
						"condition":   "${steps.transform-1.severity} == 'critical' || ${steps.transform-1.occurrences} >= 10",
						"description": "Trigger alerts for critical errors or repeated occurrences",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 50,
				},
				"data": map[string]interface{}{
					"name": "Create PagerDuty Incident",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "https://api.pagerduty.com/incidents",
						"headers": map[string]interface{}{
							"Authorization": "Token token=${env.PAGERDUTY_TOKEN}",
							"Content-Type":  "application/json",
						},
						"body": map[string]interface{}{
							"incident": map[string]interface{}{
								"type":  "incident",
								"title": "Critical Error: ${steps.transform-1.error_type}",
								"service": map[string]interface{}{
									"id":   "${env.PAGERDUTY_SERVICE_ID}",
									"type": "service_reference",
								},
								"urgency": "high",
								"body": map[string]interface{}{
									"type":    "incident_body",
									"details": "Service: ${steps.transform-1.service}\nEnvironment: ${steps.transform-1.environment}\nError: ${steps.transform-1.error_message}\nOccurrences: ${steps.transform-1.occurrences}",
								},
							},
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 700,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Alert in Slack",
					"config": map[string]interface{}{
						"channel": "#errors-critical",
						"message": "🔥 Critical Error Detected\nService: ${steps.transform-1.service}\nEnvironment: ${steps.transform-1.environment}\nType: ${steps.transform-1.error_type}\nMessage: ${steps.transform-1.error_message}\nOccurrences: ${steps.transform-1.occurrences}\nTime: ${steps.transform-1.timestamp}",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Log Error",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.ERROR_LOGGING_API}/errors",
						"body": map[string]interface{}{
							"error_type":    "${steps.transform-1.error_type}",
							"error_message": "${steps.transform-1.error_message}",
							"stack_trace":   "${steps.transform-1.stack_trace}",
							"service":       "${steps.transform-1.service}",
							"environment":   "${steps.transform-1.environment}",
							"user_id":       "${steps.transform-1.user_id}",
							"timestamp":     "${steps.transform-1.timestamp}",
							"severity":      "${steps.transform-1.severity}",
						},
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "transform-1",
				"target": "condition-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "condition-1",
				"target": "http-1",
				"label":  "true",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "http-1",
				"target": "slack-1",
			},
			map[string]interface{}{
				"id":     "e5",
				"source": "condition-1",
				"target": "http-2",
				"label":  "false",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Error Monitoring Workflow",
		Description: "Monitor application errors and automatically create PagerDuty incidents for critical errors, send Slack alerts, and log all errors for analysis.",
		Category:    string(CategoryMonitoring),
		Definition:  defJSON,
		Tags:        []string{"error", "monitoring", "alert", "pagerduty", "incident"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createMultiStepUserOnboardingTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "New User Signup",
					"config": map[string]interface{}{
						"path": "/webhooks/user-signup",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Create User Account",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.USER_API_URL}/users",
						"body": map[string]interface{}{
							"email":      "${trigger.email}",
							"first_name": "${trigger.first_name}",
							"last_name":  "${trigger.last_name}",
							"plan":       "${trigger.plan}",
							"source":     "${trigger.source}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "email-1",
				"type": "action:email",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Send Welcome Email",
					"config": map[string]interface{}{
						"to":      "${trigger.email}",
						"subject": "Welcome to Our Platform! 🎉",
						"body":    "Hi ${trigger.first_name},\n\nWelcome aboard! We're excited to have you.\n\nYour account has been created and you can now log in.\n\nHere's what to do next:\n1. Complete your profile\n2. Set up your preferences\n3. Explore our features\n\nNeed help? Just reply to this email.",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 500,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Add to CRM",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.CRM_API_URL}/contacts",
						"body": map[string]interface{}{
							"email":           "${trigger.email}",
							"name":            "${trigger.first_name} ${trigger.last_name}",
							"plan":            "${trigger.plan}",
							"signup_date":     "${now()}",
							"lifecycle_stage": "customer",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 700,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Notify Sales Team",
					"config": map[string]interface{}{
						"channel": "#new-signups",
						"message": "🎊 New User Signed Up!\nName: ${trigger.first_name} ${trigger.last_name}\nEmail: ${trigger.email}\nPlan: ${trigger.plan}\nSource: ${trigger.source}",
					},
				},
			},
			map[string]interface{}{
				"id":   "delay-1",
				"type": "control:delay",
				"position": map[string]interface{}{
					"x": 900,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Wait 3 Days",
					"config": map[string]interface{}{
						"duration": "72h",
					},
				},
			},
			map[string]interface{}{
				"id":   "email-2",
				"type": "action:email",
				"position": map[string]interface{}{
					"x": 1100,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Send Follow-up Email",
					"config": map[string]interface{}{
						"to":      "${trigger.email}",
						"subject": "How's Your Experience So Far?",
						"body":    "Hi ${trigger.first_name},\n\nIt's been a few days since you joined! How are you finding the platform?\n\nWe'd love to hear your feedback and help you get the most out of your account.\n\nBest regards,\nThe Team",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "http-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "http-1",
				"target": "email-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "http-1",
				"target": "http-2",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "email-1",
				"target": "slack-1",
			},
			map[string]interface{}{
				"id":     "e5",
				"source": "http-2",
				"target": "slack-1",
			},
			map[string]interface{}{
				"id":     "e6",
				"source": "slack-1",
				"target": "delay-1",
			},
			map[string]interface{}{
				"id":     "e7",
				"source": "delay-1",
				"target": "email-2",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Multi-Step User Onboarding",
		Description: "Complete user onboarding workflow with account creation, welcome email, CRM integration, team notifications, and automated follow-up after 3 days.",
		Category:    string(CategoryIntegration),
		Definition:  defJSON,
		Tags:        []string{"onboarding", "user", "business", "email", "crm"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createAPIHealthCheckWorkflowTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:schedule",
				"position": map[string]interface{}{
					"x": 100,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Every 5 Minutes",
					"config": map[string]interface{}{
						"cron": "*/5 * * * *",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Check API Endpoint 1",
					"config": map[string]interface{}{
						"method":  "GET",
						"url":     "${env.API_URL_1}/health",
						"timeout": 10,
					},
				},
			},
			map[string]interface{}{
				"id":   "http-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Check API Endpoint 2",
					"config": map[string]interface{}{
						"method":  "GET",
						"url":     "${env.API_URL_2}/health",
						"timeout": 10,
					},
				},
			},
			map[string]interface{}{
				"id":   "http-3",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 300,
				},
				"data": map[string]interface{}{
					"name": "Check Database",
					"config": map[string]interface{}{
						"method":  "GET",
						"url":     "${env.DB_HEALTH_URL}/ping",
						"timeout": 10,
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 500,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Compile Health Status",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"api1_status":  "${steps.http-1.status_code}",
							"api1_healthy": "${steps.http-1.status_code == 200}",
							"api2_status":  "${steps.http-2.status_code}",
							"api2_healthy": "${steps.http-2.status_code == 200}",
							"db_status":    "${steps.http-3.status_code}",
							"db_healthy":   "${steps.http-3.status_code == 200}",
							"all_healthy":  "${steps.http-1.status_code == 200 && steps.http-2.status_code == 200 && steps.http-3.status_code == 200}",
							"timestamp":    "${now()}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "condition-1",
				"type": "control:if",
				"position": map[string]interface{}{
					"x": 700,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Check if Unhealthy",
					"config": map[string]interface{}{
						"condition":   "!${steps.transform-1.all_healthy}",
						"description": "Alert if any service is unhealthy",
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 900,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Alert Team",
					"config": map[string]interface{}{
						"channel": "#health-alerts",
						"message": "⚠️ Health Check Failed\nAPI 1: ${steps.transform-1.api1_healthy ? '✅' : '❌'} (${steps.transform-1.api1_status})\nAPI 2: ${steps.transform-1.api2_healthy ? '✅' : '❌'} (${steps.transform-1.api2_status})\nDatabase: ${steps.transform-1.db_healthy ? '✅' : '❌'} (${steps.transform-1.db_status})\nTime: ${steps.transform-1.timestamp}",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-4",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 900,
					"y": 250,
				},
				"data": map[string]interface{}{
					"name": "Log Health Status",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.MONITORING_API}/health-checks",
						"body": map[string]interface{}{
							"api1_healthy": "${steps.transform-1.api1_healthy}",
							"api2_healthy": "${steps.transform-1.api2_healthy}",
							"db_healthy":   "${steps.transform-1.db_healthy}",
							"all_healthy":  "${steps.transform-1.all_healthy}",
							"timestamp":    "${steps.transform-1.timestamp}",
						},
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "http-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "trigger-1",
				"target": "http-2",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "trigger-1",
				"target": "http-3",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "http-1",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e5",
				"source": "http-2",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e6",
				"source": "http-3",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e7",
				"source": "transform-1",
				"target": "condition-1",
			},
			map[string]interface{}{
				"id":     "e8",
				"source": "condition-1",
				"target": "slack-1",
				"label":  "true",
			},
			map[string]interface{}{
				"id":     "e9",
				"source": "condition-1",
				"target": "http-4",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "API Health Check Workflow",
		Description: "Periodic health checks for multiple API endpoints and databases. Monitors service availability, alerts team on failures, and logs all health status.",
		Category:    string(CategoryMonitoring),
		Definition:  defJSON,
		Tags:        []string{"health", "monitoring", "api", "schedule", "alert"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createScheduledReportGenerationTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:schedule",
				"position": map[string]interface{}{
					"x": 100,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Monthly Report Schedule",
					"config": map[string]interface{}{
						"cron":     "0 8 1 * *",
						"timezone": "America/New_York",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Fetch Revenue Data",
					"config": map[string]interface{}{
						"method": "GET",
						"url":    "${env.ANALYTICS_API_URL}/revenue/monthly",
						"params": map[string]interface{}{
							"start_date": "${dateAdd(now(), -1, 'month')}",
							"end_date":   "${now()}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "http-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Fetch User Metrics",
					"config": map[string]interface{}{
						"method": "GET",
						"url":    "${env.ANALYTICS_API_URL}/users/monthly",
						"params": map[string]interface{}{
							"start_date": "${dateAdd(now(), -1, 'month')}",
							"end_date":   "${now()}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "http-3",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 300,
				},
				"data": map[string]interface{}{
					"name": "Fetch Conversion Data",
					"config": map[string]interface{}{
						"method": "GET",
						"url":    "${env.ANALYTICS_API_URL}/conversions/monthly",
						"params": map[string]interface{}{
							"start_date": "${dateAdd(now(), -1, 'month')}",
							"end_date":   "${now()}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 500,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Compile Report Data",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"period":            "${dateFormat(dateAdd(now(), -1, 'month'), 'MMMM YYYY')}",
							"total_revenue":     "${steps.http-1.response.total}",
							"revenue_growth":    "${steps.http-1.response.growth_percentage}",
							"new_users":         "${steps.http-2.response.new_users}",
							"active_users":      "${steps.http-2.response.active_users}",
							"churn_rate":        "${steps.http-2.response.churn_rate}",
							"conversion_rate":   "${steps.http-3.response.rate}",
							"total_conversions": "${steps.http-3.response.total}",
							"generated_at":      "${now()}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "http-4",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Generate PDF Report",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.REPORT_GENERATOR_URL}/generate/pdf",
						"body": map[string]interface{}{
							"template": "monthly_business_report",
							"data":     "${steps.transform-1}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "email-1",
				"type": "action:email",
				"position": map[string]interface{}{
					"x": 900,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Email Report to Executives",
					"config": map[string]interface{}{
						"to":      "${env.EXECUTIVE_EMAILS}",
						"subject": "📊 Monthly Business Report - ${steps.transform-1.period}",
						"body":    "Dear Team,\n\nPlease find attached the monthly business report for ${steps.transform-1.period}.\n\nKey Highlights:\n• Total Revenue: $${steps.transform-1.total_revenue} (${steps.transform-1.revenue_growth}% growth)\n• New Users: ${steps.transform-1.new_users}\n• Active Users: ${steps.transform-1.active_users}\n• Conversion Rate: ${steps.transform-1.conversion_rate}%\n• Churn Rate: ${steps.transform-1.churn_rate}%\n\nFull report is attached.\n\nBest regards,\nAutomated Reporting System",
						"attachments": []map[string]interface{}{
							{
								"filename": "monthly_report_${steps.transform-1.period}.pdf",
								"url":      "${steps.http-4.response.pdf_url}",
							},
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 900,
					"y": 250,
				},
				"data": map[string]interface{}{
					"name": "Post Summary to Slack",
					"config": map[string]interface{}{
						"channel": "#business-metrics",
						"message": "📊 Monthly Report Generated - ${steps.transform-1.period}\n\n💰 Revenue: $${steps.transform-1.total_revenue} (${steps.transform-1.revenue_growth}% growth)\n👥 New Users: ${steps.transform-1.new_users}\n📈 Conversion Rate: ${steps.transform-1.conversion_rate}%\n\nFull report has been emailed to executives.",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "http-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "trigger-1",
				"target": "http-2",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "trigger-1",
				"target": "http-3",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "http-1",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e5",
				"source": "http-2",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e6",
				"source": "http-3",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e7",
				"source": "transform-1",
				"target": "http-4",
			},
			map[string]interface{}{
				"id":     "e8",
				"source": "http-4",
				"target": "email-1",
			},
			map[string]interface{}{
				"id":     "e9",
				"source": "email-1",
				"target": "slack-1",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Scheduled Report Generation",
		Description: "Automated monthly business report generation with parallel data fetching, PDF creation, email distribution to executives, and Slack notifications.",
		Category:    string(CategoryDataOps),
		Definition:  defJSON,
		Tags:        []string{"report", "schedule", "analytics", "email", "dataops"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Additional Templates

func createDataETLSyncWorkflowTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:schedule",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Scheduled ETL Trigger",
					"config": map[string]interface{}{
						"cron":        "0 2 * * *",
						"timezone":    "UTC",
						"description": "Run ETL sync at 2 AM daily",
					},
				},
			},
			map[string]interface{}{
				"id":   "extract-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Extract Data from Source",
					"config": map[string]interface{}{
						"method": "GET",
						"url":    "${env.SOURCE_DB_API}/extract",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${credentials.source_db_token}",
						},
						"query_params": map[string]interface{}{
							"since": "${trigger.last_run_time}",
							"limit": "1000",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Transform Data",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"records": "${steps.extract-1.data.map(item => ({ id: item.id, name: item.name.trim().toUpperCase(), email: item.email.toLowerCase(), created_at: item.timestamp }))}",
							"count":   "${steps.extract-1.data.length}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "validate-1",
				"type": "control:if",
				"position": map[string]interface{}{
					"x": 700,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Validate Data Quality",
					"config": map[string]interface{}{
						"condition":   "${steps.transform-1.count > 0}",
						"description": "Ensure we have data to load",
					},
				},
			},
			map[string]interface{}{
				"id":   "load-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 900,
					"y": 50,
				},
				"data": map[string]interface{}{
					"name": "Load Data to Target",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.TARGET_DB_API}/bulk-insert",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${credentials.target_db_token}",
							"Content-Type":  "application/json",
						},
						"body": map[string]interface{}{
							"records": "${steps.transform-1.records}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "notify-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 1100,
					"y": 50,
				},
				"data": map[string]interface{}{
					"name": "Notify Success",
					"config": map[string]interface{}{
						"channel": "#data-ops",
						"message": "✅ ETL Sync Completed\nRecords processed: ${steps.transform-1.count}\nStatus: ${steps.load-1.status}",
					},
				},
			},
			map[string]interface{}{
				"id":   "notify-2",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 900,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Notify No Data",
					"config": map[string]interface{}{
						"channel": "#data-ops",
						"message": "⚠️ ETL Sync: No new data to process",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "extract-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "extract-1",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "transform-1",
				"target": "validate-1",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "validate-1",
				"target": "load-1",
				"label":  "true",
			},
			map[string]interface{}{
				"id":     "e5",
				"source": "validate-1",
				"target": "notify-2",
				"label":  "false",
			},
			map[string]interface{}{
				"id":     "e6",
				"source": "load-1",
				"target": "notify-1",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Data ETL Sync Workflow",
		Description: "Scheduled ETL workflow for extracting data from source systems, transforming it according to business rules, and loading into target databases. Includes data validation and notification steps.",
		Category:    string(CategoryDataOps),
		Definition:  defJSON,
		Tags:        []string{"etl", "sync", "dataops", "schedule", "automation"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createScheduledReportingWorkflowTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:schedule",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Weekly Report Trigger",
					"config": map[string]interface{}{
						"cron":        "0 9 * * 1",
						"timezone":    "America/New_York",
						"description": "Generate report every Monday at 9 AM",
					},
				},
			},
			map[string]interface{}{
				"id":   "fetch-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Fetch Analytics Data",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.ANALYTICS_API}/query",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${credentials.analytics_token}",
						},
						"body": map[string]interface{}{
							"metrics":    []string{"revenue", "users", "conversions"},
							"start_date": "${trigger.last_week_start}",
							"end_date":   "${trigger.last_week_end}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Calculate Metrics",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"total_revenue":   "${steps.fetch-1.revenue.reduce((sum, v) => sum + v, 0)}",
							"avg_daily_users": "${steps.fetch-1.users.reduce((sum, v) => sum + v, 0) / 7}",
							"conversion_rate": "${(steps.fetch-1.conversions / steps.fetch-1.users.reduce((sum, v) => sum + v, 0)) * 100}",
							"week_period":     "${trigger.last_week_start + ' to ' + trigger.last_week_end}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "generate-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Generate PDF Report",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.REPORT_API}/generate",
						"body": map[string]interface{}{
							"template": "weekly_analytics",
							"data": map[string]interface{}{
								"metrics": "${steps.transform-1}",
							},
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "email-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 900,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Email Report",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.EMAIL_API}/send",
						"body": map[string]interface{}{
							"to":      []string{"executives@company.com"},
							"subject": "Weekly Analytics Report - ${steps.transform-1.week_period}",
							"body":    "Please find attached the weekly analytics report.\n\nRevenue: $${steps.transform-1.total_revenue}\nAvg Daily Users: ${steps.transform-1.avg_daily_users}\nConversion Rate: ${steps.transform-1.conversion_rate}%",
							"attachments": []map[string]interface{}{
								{
									"filename": "weekly_report.pdf",
									"content":  "${steps.generate-1.pdf_base64}",
								},
							},
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 1100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Slack Notification",
					"config": map[string]interface{}{
						"channel": "#analytics",
						"message": "📊 Weekly Report Generated\nPeriod: ${steps.transform-1.week_period}\nRevenue: $${steps.transform-1.total_revenue}\nReport sent to executives.",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "fetch-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "fetch-1",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "transform-1",
				"target": "generate-1",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "generate-1",
				"target": "email-1",
			},
			map[string]interface{}{
				"id":     "e5",
				"source": "email-1",
				"target": "slack-1",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Scheduled Reporting Workflow",
		Description: "Automated report generation workflow that fetches analytics data, calculates key metrics, generates PDF reports, and distributes via email. Configurable for daily, weekly, or monthly schedules.",
		Category:    string(CategoryDataOps),
		Definition:  defJSON,
		Tags:        []string{"report", "schedule", "analytics", "automation", "email"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createMultiStepApprovalWorkflowTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Approval Request Received",
					"config": map[string]interface{}{
						"path": "/webhooks/approval-request",
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Create Approval Record",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.API_URL}/approvals",
						"body": map[string]interface{}{
							"request_id":  "${trigger.request_id}",
							"requester":   "${trigger.requester}",
							"amount":      "${trigger.amount}",
							"description": "${trigger.description}",
							"status":      "pending_manager",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "email-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 500,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Notify Manager",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.EMAIL_API}/send",
						"body": map[string]interface{}{
							"to":      "${trigger.manager_email}",
							"subject": "Approval Required: ${trigger.description}",
							"body":    "A new approval request requires your attention.\n\nRequester: ${trigger.requester}\nAmount: $${trigger.amount}\nDescription: ${trigger.description}\n\nApproval Link: ${env.APP_URL}/approvals/${steps.http-1.id}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "wait-1",
				"type": "control:wait",
				"position": map[string]interface{}{
					"x": 700,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Wait for Manager Approval",
					"config": map[string]interface{}{
						"timeout":      "48h",
						"webhook_path": "/webhooks/approval/${steps.http-1.id}/manager",
					},
				},
			},
			map[string]interface{}{
				"id":   "condition-1",
				"type": "control:if",
				"position": map[string]interface{}{
					"x": 900,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Check Manager Decision",
					"config": map[string]interface{}{
						"condition": "${steps.wait-1.decision == 'approved'}",
					},
				},
			},
			map[string]interface{}{
				"id":   "email-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 1100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Notify Director",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.EMAIL_API}/send",
						"body": map[string]interface{}{
							"to":      "${trigger.director_email}",
							"subject": "Final Approval Required: ${trigger.description}",
							"body":    "Manager approved. Final approval required.\n\nRequester: ${trigger.requester}\nAmount: $${trigger.amount}\nManager: ${steps.wait-1.approved_by}\n\nApproval Link: ${env.APP_URL}/approvals/${steps.http-1.id}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "wait-2",
				"type": "control:wait",
				"position": map[string]interface{}{
					"x": 1300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Wait for Director Approval",
					"config": map[string]interface{}{
						"timeout":      "24h",
						"webhook_path": "/webhooks/approval/${steps.http-1.id}/director",
					},
				},
			},
			map[string]interface{}{
				"id":   "notify-approved",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 1500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Notify Fully Approved",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.EMAIL_API}/send",
						"body": map[string]interface{}{
							"to":      "${trigger.requester}",
							"subject": "Request Approved: ${trigger.description}",
							"body":    "Your request has been fully approved!\n\nAmount: $${trigger.amount}\nApproved by: ${steps.wait-1.approved_by}, ${steps.wait-2.approved_by}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "notify-rejected",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 1100,
					"y": 300,
				},
				"data": map[string]interface{}{
					"name": "Notify Rejected",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.EMAIL_API}/send",
						"body": map[string]interface{}{
							"to":      "${trigger.requester}",
							"subject": "Request Rejected: ${trigger.description}",
							"body":    "Your request has been rejected.\n\nAmount: $${trigger.amount}\nRejected by: ${steps.wait-1.rejected_by || steps.wait-2.rejected_by}\nReason: ${steps.wait-1.reason || steps.wait-2.reason}",
						},
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "http-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "http-1",
				"target": "email-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "email-1",
				"target": "wait-1",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "wait-1",
				"target": "condition-1",
			},
			map[string]interface{}{
				"id":     "e5",
				"source": "condition-1",
				"target": "email-2",
				"label":  "approved",
			},
			map[string]interface{}{
				"id":     "e6",
				"source": "condition-1",
				"target": "notify-rejected",
				"label":  "rejected",
			},
			map[string]interface{}{
				"id":     "e7",
				"source": "email-2",
				"target": "wait-2",
			},
			map[string]interface{}{
				"id":     "e8",
				"source": "wait-2",
				"target": "notify-approved",
				"label":  "approved",
			},
			map[string]interface{}{
				"id":     "e9",
				"source": "wait-2",
				"target": "notify-rejected",
				"label":  "rejected",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Multi-Step Approval Workflow",
		Description: "Comprehensive approval workflow with manager and director approval steps. Includes timeout handling, rejection paths, and automatic notifications at each stage. Ideal for expense approvals, purchase orders, or any multi-level authorization process.",
		Category:    string(CategoryIntegration),
		Definition:  defJSON,
		Tags:        []string{"approval", "business", "workflow", "automation", "multi-step"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createErrorNotificationWorkflowTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Error Event Received",
					"config": map[string]interface{}{
						"path": "/webhooks/error-event",
					},
				},
			},
			map[string]interface{}{
				"id":   "enrich-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 300,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Enrich Error Context",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"error_message": "${trigger.error}",
							"service":       "${trigger.service}",
							"environment":   "${trigger.environment}",
							"timestamp":     "${trigger.timestamp}",
							"stack_trace":   "${trigger.stack_trace}",
							"user_id":       "${trigger.user_id}",
							"request_id":    "${trigger.request_id}",
							"severity":      "${trigger.severity || 'medium'}",
							"error_hash":    "${trigger.error.split(':')[0]}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "http-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 500,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Log to Error Tracking",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.ERROR_TRACKING_API}/errors",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${credentials.error_tracking_token}",
						},
						"body": "${steps.enrich-1}",
					},
				},
			},
			map[string]interface{}{
				"id":   "condition-1",
				"type": "control:if",
				"position": map[string]interface{}{
					"x": 700,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Check Severity",
					"config": map[string]interface{}{
						"condition": "${steps.enrich-1.severity == 'critical' || steps.enrich-1.severity == 'high'}",
					},
				},
			},
			map[string]interface{}{
				"id":   "pagerduty-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 900,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Create PagerDuty Incident",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.PAGERDUTY_API}/incidents",
						"headers": map[string]interface{}{
							"Authorization": "Token ${credentials.pagerduty_token}",
						},
						"body": map[string]interface{}{
							"incident": map[string]interface{}{
								"type":    "incident",
								"title":   "${steps.enrich-1.service}: ${steps.enrich-1.error_message}",
								"service": "${env.PAGERDUTY_SERVICE_ID}",
								"urgency": "high",
								"body": map[string]interface{}{
									"type":    "incident_body",
									"details": "Environment: ${steps.enrich-1.environment}\nService: ${steps.enrich-1.service}\nError: ${steps.enrich-1.error_message}\nRequest ID: ${steps.enrich-1.request_id}",
								},
							},
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-critical",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 1100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Alert Team - Critical",
					"config": map[string]interface{}{
						"channel": "#incidents",
						"message": "🚨 CRITICAL ERROR ALERT\n\nService: ${steps.enrich-1.service}\nEnvironment: ${steps.enrich-1.environment}\nError: ${steps.enrich-1.error_message}\nRequest ID: ${steps.enrich-1.request_id}\nPagerDuty Incident: ${steps.pagerduty-1.incident_url}",
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-normal",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 900,
					"y": 300,
				},
				"data": map[string]interface{}{
					"name": "Log to Slack - Normal",
					"config": map[string]interface{}{
						"channel": "#errors",
						"message": "⚠️ Error Logged\n\nService: ${steps.enrich-1.service}\nEnvironment: ${steps.enrich-1.environment}\nError: ${steps.enrich-1.error_message}\nSeverity: ${steps.enrich-1.severity}",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "enrich-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "enrich-1",
				"target": "http-1",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "http-1",
				"target": "condition-1",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "condition-1",
				"target": "pagerduty-1",
				"label":  "critical",
			},
			map[string]interface{}{
				"id":     "e5",
				"source": "condition-1",
				"target": "slack-normal",
				"label":  "normal",
			},
			map[string]interface{}{
				"id":     "e6",
				"source": "pagerduty-1",
				"target": "slack-critical",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Error Notification Workflow",
		Description: "Intelligent error handling and notification workflow that enriches error context, logs to error tracking systems, and routes notifications based on severity. Critical errors trigger PagerDuty incidents and alert on-call engineers.",
		Category:    string(CategoryMonitoring),
		Definition:  defJSON,
		Tags:        []string{"error", "notification", "monitoring", "pagerduty", "incident"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createAPIOrchestrationWorkflowTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 250,
				},
				"data": map[string]interface{}{
					"name": "Request Received",
					"config": map[string]interface{}{
						"path": "/webhooks/orchestrate",
					},
				},
			},
			map[string]interface{}{
				"id":   "api-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Fetch User Profile",
					"config": map[string]interface{}{
						"method": "GET",
						"url":    "${env.USER_API}/users/${trigger.user_id}",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${credentials.user_api_token}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "api-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 250,
				},
				"data": map[string]interface{}{
					"name": "Fetch Account Details",
					"config": map[string]interface{}{
						"method": "GET",
						"url":    "${env.ACCOUNT_API}/accounts/${trigger.account_id}",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${credentials.account_api_token}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "api-3",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 350,
				},
				"data": map[string]interface{}{
					"name": "Fetch Transaction History",
					"config": map[string]interface{}{
						"method": "GET",
						"url":    "${env.TRANSACTION_API}/transactions",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${credentials.transaction_api_token}",
						},
						"query_params": map[string]interface{}{
							"user_id": "${trigger.user_id}",
							"limit":   "10",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 500,
					"y": 250,
				},
				"data": map[string]interface{}{
					"name": "Aggregate Data",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"user": map[string]interface{}{
								"id":    "${steps.api-1.id}",
								"name":  "${steps.api-1.name}",
								"email": "${steps.api-1.email}",
								"tier":  "${steps.api-1.tier}",
							},
							"account": map[string]interface{}{
								"id":      "${steps.api-2.id}",
								"balance": "${steps.api-2.balance}",
								"status":  "${steps.api-2.status}",
							},
							"transactions": map[string]interface{}{
								"recent":       "${steps.api-3.transactions}",
								"total_count":  "${steps.api-3.total_count}",
								"total_amount": "${steps.api-3.transactions.reduce((sum, t) => sum + t.amount, 0)}",
							},
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "api-4",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 250,
				},
				"data": map[string]interface{}{
					"name": "Calculate Risk Score",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.RISK_API}/calculate",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${credentials.risk_api_token}",
						},
						"body": map[string]interface{}{
							"user_id":      "${trigger.user_id}",
							"account_data": "${steps.transform-1.account}",
							"transaction_data": map[string]interface{}{
								"count":  "${steps.transform-1.transactions.total_count}",
								"amount": "${steps.transform-1.transactions.total_amount}",
							},
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "api-5",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 900,
					"y": 250,
				},
				"data": map[string]interface{}{
					"name": "Generate Recommendations",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.RECOMMENDATION_API}/generate",
						"body": map[string]interface{}{
							"user":         "${steps.transform-1.user}",
							"account":      "${steps.transform-1.account}",
							"risk_score":   "${steps.api-4.risk_score}",
							"transactions": "${steps.transform-1.transactions}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "response-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 1100,
					"y": 250,
				},
				"data": map[string]interface{}{
					"name": "Format Response",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"status": "success",
							"data": map[string]interface{}{
								"user":            "${steps.transform-1.user}",
								"account":         "${steps.transform-1.account}",
								"transactions":    "${steps.transform-1.transactions}",
								"risk_score":      "${steps.api-4.risk_score}",
								"risk_level":      "${steps.api-4.risk_level}",
								"recommendations": "${steps.api-5.recommendations}",
							},
						},
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"id":     "e1",
				"source": "trigger-1",
				"target": "api-1",
			},
			map[string]interface{}{
				"id":     "e2",
				"source": "trigger-1",
				"target": "api-2",
			},
			map[string]interface{}{
				"id":     "e3",
				"source": "trigger-1",
				"target": "api-3",
			},
			map[string]interface{}{
				"id":     "e4",
				"source": "api-1",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e5",
				"source": "api-2",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e6",
				"source": "api-3",
				"target": "transform-1",
			},
			map[string]interface{}{
				"id":     "e7",
				"source": "transform-1",
				"target": "api-4",
			},
			map[string]interface{}{
				"id":     "e8",
				"source": "api-4",
				"target": "api-5",
			},
			map[string]interface{}{
				"id":     "e9",
				"source": "api-5",
				"target": "response-1",
			},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "API Orchestration Workflow",
		Description: "Complex API orchestration that coordinates multiple microservices, aggregates responses, calculates derived metrics, and returns a unified response. Demonstrates parallel API calls, data transformation, and sequential processing for comprehensive data assembly.",
		Category:    string(CategoryIntegration),
		Definition:  defJSON,
		Tags:        []string{"api", "orchestration", "integration", "microservices", "aggregation"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// DevOps Templates (additional)

func createKubernetesDeploymentMonitorTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:schedule",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Check Every 5 Minutes",
					"config": map[string]interface{}{
						"cron": "*/5 * * * *",
					},
				},
			},
			map[string]interface{}{
				"id":   "api-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Get Deployments",
					"config": map[string]interface{}{
						"method":  "GET",
						"url":     "${env.K8S_API_URL}/apis/apps/v1/namespaces/${env.K8S_NAMESPACE}/deployments",
						"headers": map[string]interface{}{"Authorization": "Bearer ${credentials.k8s_token}"},
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Check Deployment Health",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"unhealthy_deployments": "filter(${steps.api-1.body.items}, d => d.status.readyReplicas < d.status.replicas)",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "condition-1",
				"type": "control:if",
				"position": map[string]interface{}{
					"x": 700,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Has Unhealthy?",
					"config": map[string]interface{}{
						"condition": "len(${steps.transform-1.unhealthy_deployments}) > 0",
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 900,
					"y": 50,
				},
				"data": map[string]interface{}{
					"name": "Alert Unhealthy Deployments",
					"config": map[string]interface{}{
						"channel": "#k8s-alerts",
						"message": "⚠️ Unhealthy Deployments Detected\n\nNamespace: ${env.K8S_NAMESPACE}\nUnhealthy: ${len(steps.transform-1.unhealthy_deployments)} deployment(s)\n\nCheck cluster status immediately.",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{"id": "e1", "source": "trigger-1", "target": "api-1"},
			map[string]interface{}{"id": "e2", "source": "api-1", "target": "transform-1"},
			map[string]interface{}{"id": "e3", "source": "transform-1", "target": "condition-1"},
			map[string]interface{}{"id": "e4", "source": "condition-1", "target": "slack-1", "label": "true"},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Kubernetes Deployment Monitor",
		Description: "Periodically checks Kubernetes deployments for unhealthy pods and sends Slack alerts when deployments have fewer ready replicas than expected.",
		Category:    string(CategoryDevOps),
		Definition:  defJSON,
		Tags:        []string{"kubernetes", "k8s", "monitoring", "deployment", "devops", "alerting"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createContainerRegistryCleanupTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:schedule",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Weekly Cleanup",
					"config": map[string]interface{}{
						"cron": "0 2 * * 0",
					},
				},
			},
			map[string]interface{}{
				"id":   "api-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "List Images",
					"config": map[string]interface{}{
						"method":  "GET",
						"url":     "${env.REGISTRY_URL}/v2/_catalog",
						"headers": map[string]interface{}{"Authorization": "Basic ${credentials.registry_auth}"},
					},
				},
			},
			map[string]interface{}{
				"id":   "loop-1",
				"type": "control:loop",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Process Each Repository",
					"config": map[string]interface{}{
						"items":    "${steps.api-1.body.repositories}",
						"variable": "repo",
					},
				},
			},
			map[string]interface{}{
				"id":   "api-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Get Tags",
					"config": map[string]interface{}{
						"method":  "GET",
						"url":     "${env.REGISTRY_URL}/v2/${loop.repo}/tags/list",
						"headers": map[string]interface{}{"Authorization": "Basic ${credentials.registry_auth}"},
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 900,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Filter Old Tags",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"old_tags": "filter(${steps.api-2.body.tags}, t => !startsWith(t, 'release-') && !contains(['latest', 'stable'], t))",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 1100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Report Cleanup",
					"config": map[string]interface{}{
						"channel": "#devops",
						"message": "🧹 Container Registry Cleanup Complete\n\nProcessed: ${len(steps.api-1.body.repositories)} repositories\nTags identified for cleanup: ${len(steps.transform-1.old_tags)}",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{"id": "e1", "source": "trigger-1", "target": "api-1"},
			map[string]interface{}{"id": "e2", "source": "api-1", "target": "loop-1"},
			map[string]interface{}{"id": "e3", "source": "loop-1", "target": "api-2"},
			map[string]interface{}{"id": "e4", "source": "api-2", "target": "transform-1"},
			map[string]interface{}{"id": "e5", "source": "transform-1", "target": "slack-1"},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Container Registry Cleanup",
		Description: "Weekly scheduled workflow that scans container registries for old or unused image tags and reports cleanup opportunities. Helps manage registry storage costs.",
		Category:    string(CategoryDevOps),
		Definition:  defJSON,
		Tags:        []string{"docker", "container", "registry", "cleanup", "devops", "automation"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Integration Templates (additional)

func createSalesforceLeadSyncTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "New Lead Webhook",
					"config": map[string]interface{}{
						"path":      "/webhooks/new-lead",
						"auth_type": "api_key",
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Map to Salesforce",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"FirstName":  "${trigger.first_name}",
							"LastName":   "${trigger.last_name}",
							"Email":      "${trigger.email}",
							"Company":    "${trigger.company}",
							"Phone":      "${trigger.phone}",
							"LeadSource": "Web",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "api-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Create Salesforce Lead",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.SALESFORCE_INSTANCE_URL}/services/data/v58.0/sobjects/Lead",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${credentials.salesforce_token}",
							"Content-Type":  "application/json",
						},
						"body": "${steps.transform-1}",
					},
				},
			},
			map[string]interface{}{
				"id":   "condition-1",
				"type": "control:if",
				"position": map[string]interface{}{
					"x": 700,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Check Success",
					"config": map[string]interface{}{
						"condition": "${steps.api-1.status} == 201",
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 900,
					"y": 50,
				},
				"data": map[string]interface{}{
					"name": "Notify Sales Team",
					"config": map[string]interface{}{
						"channel": "#sales-leads",
						"message": "🎯 New Lead Created in Salesforce\n\nName: ${trigger.first_name} ${trigger.last_name}\nCompany: ${trigger.company}\nEmail: ${trigger.email}\n\nSalesforce ID: ${steps.api-1.body.id}",
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-2",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 900,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Alert on Failure",
					"config": map[string]interface{}{
						"channel": "#sales-ops",
						"message": "❌ Failed to create Salesforce lead\n\nEmail: ${trigger.email}\nError: ${steps.api-1.body}",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{"id": "e1", "source": "trigger-1", "target": "transform-1"},
			map[string]interface{}{"id": "e2", "source": "transform-1", "target": "api-1"},
			map[string]interface{}{"id": "e3", "source": "api-1", "target": "condition-1"},
			map[string]interface{}{"id": "e4", "source": "condition-1", "target": "slack-1", "label": "true"},
			map[string]interface{}{"id": "e5", "source": "condition-1", "target": "slack-2", "label": "false"},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Salesforce Lead Sync",
		Description: "Automatically creates leads in Salesforce from webhook events and notifies the sales team. Includes error handling and failure notifications.",
		Category:    string(CategoryIntegration),
		Definition:  defJSON,
		Tags:        []string{"salesforce", "crm", "lead", "sync", "integration", "sales"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createServiceNowIncidentTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Alert Webhook",
					"config": map[string]interface{}{
						"path":      "/webhooks/create-incident",
						"auth_type": "signature",
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Map to ServiceNow",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"short_description": "${trigger.title}",
							"description":       "${trigger.description}",
							"urgency":           "${trigger.severity == 'critical' ? '1' : trigger.severity == 'high' ? '2' : '3'}",
							"impact":            "${trigger.severity == 'critical' ? '1' : trigger.severity == 'high' ? '2' : '3'}",
							"category":          "software",
							"caller_id":         "${env.SERVICENOW_CALLER_ID}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "api-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Create ServiceNow Incident",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.SERVICENOW_INSTANCE}/api/now/table/incident",
						"headers": map[string]interface{}{
							"Authorization": "Basic ${credentials.servicenow_auth}",
							"Content-Type":  "application/json",
						},
						"body": "${steps.transform-1}",
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-2",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 700,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Extract Incident Number",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"incident_number": "${steps.api-1.body.result.number}",
							"incident_url":    "${env.SERVICENOW_INSTANCE}/nav_to.do?uri=incident.do?sys_id=${steps.api-1.body.result.sys_id}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 900,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Notify Team",
					"config": map[string]interface{}{
						"channel": "#incidents",
						"message": "🎫 ServiceNow Incident Created\n\nIncident: ${steps.transform-2.incident_number}\nTitle: ${trigger.title}\nSeverity: ${trigger.severity}\n\nView: ${steps.transform-2.incident_url}",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{"id": "e1", "source": "trigger-1", "target": "transform-1"},
			map[string]interface{}{"id": "e2", "source": "transform-1", "target": "api-1"},
			map[string]interface{}{"id": "e3", "source": "api-1", "target": "transform-2"},
			map[string]interface{}{"id": "e4", "source": "transform-2", "target": "slack-1"},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "ServiceNow Incident Creation",
		Description: "Creates ServiceNow incidents from webhook alerts with automatic severity mapping and team notification via Slack.",
		Category:    string(CategoryIntegration),
		Definition:  defJSON,
		Tags:        []string{"servicenow", "itsm", "incident", "integration", "alerting"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Security Templates

func createSecurityAlertTriageTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Security Alert Webhook",
					"config": map[string]interface{}{
						"path":      "/webhooks/security-alert",
						"auth_type": "signature",
						"secret":    "${env.SECURITY_WEBHOOK_SECRET}",
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Enrich Alert Data",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"severity":       "${trigger.severity}",
							"source":         "${trigger.source}",
							"description":    "${trigger.description}",
							"affected_asset": "${trigger.asset}",
							"timestamp":      "${trigger.timestamp}",
							"is_critical":    "${trigger.severity == 'critical' || trigger.severity == 'high'}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "condition-1",
				"type": "control:if",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Is Critical?",
					"config": map[string]interface{}{
						"condition": "${steps.transform-1.is_critical} == true",
					},
				},
			},
			map[string]interface{}{
				"id":   "api-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 50,
				},
				"data": map[string]interface{}{
					"name": "Create PagerDuty Incident",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "https://events.pagerduty.com/v2/enqueue",
						"headers": map[string]interface{}{
							"Content-Type": "application/json",
						},
						"body": map[string]interface{}{
							"routing_key":  "${credentials.pagerduty_key}",
							"event_action": "trigger",
							"payload": map[string]interface{}{
								"summary":   "${steps.transform-1.description}",
								"severity":  "${steps.transform-1.severity}",
								"source":    "${steps.transform-1.source}",
								"component": "${steps.transform-1.affected_asset}",
							},
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "api-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Log to SIEM",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.SIEM_ENDPOINT}/api/v1/events",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${credentials.siem_token}",
							"Content-Type":  "application/json",
						},
						"body": "${trigger}",
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 900,
					"y": 50,
				},
				"data": map[string]interface{}{
					"name": "Alert Security Team",
					"config": map[string]interface{}{
						"channel": "#security-critical",
						"message": "🚨 CRITICAL SECURITY ALERT\n\nSeverity: ${steps.transform-1.severity}\nSource: ${steps.transform-1.source}\nAsset: ${steps.transform-1.affected_asset}\n\n${steps.transform-1.description}\n\nPagerDuty incident created. Immediate response required.",
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-2",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 900,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Log to Security Channel",
					"config": map[string]interface{}{
						"channel": "#security-alerts",
						"message": "🔔 Security Alert\n\nSeverity: ${steps.transform-1.severity}\nSource: ${steps.transform-1.source}\nAsset: ${steps.transform-1.affected_asset}\n\n${steps.transform-1.description}",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{"id": "e1", "source": "trigger-1", "target": "transform-1"},
			map[string]interface{}{"id": "e2", "source": "transform-1", "target": "condition-1"},
			map[string]interface{}{"id": "e3", "source": "condition-1", "target": "api-1", "label": "true"},
			map[string]interface{}{"id": "e4", "source": "api-1", "target": "slack-1"},
			map[string]interface{}{"id": "e5", "source": "condition-1", "target": "api-2", "label": "false"},
			map[string]interface{}{"id": "e6", "source": "api-2", "target": "slack-2"},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Security Alert Triage",
		Description: "Automatically triages security alerts by severity, creates PagerDuty incidents for critical alerts, logs to SIEM, and notifies the security team via Slack.",
		Category:    string(CategorySecurity),
		Definition:  defJSON,
		Tags:        []string{"security", "alerting", "triage", "pagerduty", "siem", "incident-response"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createVulnerabilityResponseTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:webhook",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Vulnerability Scanner Webhook",
					"config": map[string]interface{}{
						"path":      "/webhooks/vulnerability",
						"auth_type": "signature",
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Parse Vulnerability",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"cve_id":      "${trigger.cve_id}",
							"cvss_score":  "${trigger.cvss_score}",
							"severity":    "${trigger.cvss_score >= 9.0 ? 'critical' : trigger.cvss_score >= 7.0 ? 'high' : trigger.cvss_score >= 4.0 ? 'medium' : 'low'}",
							"affected":    "${trigger.affected_packages}",
							"description": "${trigger.description}",
							"remediation": "${trigger.fix_available ? trigger.fix_version : 'No fix available'}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "condition-1",
				"type": "control:if",
				"position": map[string]interface{}{
					"x": 500,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Critical/High?",
					"config": map[string]interface{}{
						"condition": "${steps.transform-1.cvss_score} >= 7.0",
					},
				},
			},
			map[string]interface{}{
				"id":   "api-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 50,
				},
				"data": map[string]interface{}{
					"name": "Create Jira Ticket",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.JIRA_URL}/rest/api/3/issue",
						"headers": map[string]interface{}{
							"Authorization": "Basic ${credentials.jira_auth}",
							"Content-Type":  "application/json",
						},
						"body": map[string]interface{}{
							"fields": map[string]interface{}{
								"project":     map[string]interface{}{"key": "${env.JIRA_SECURITY_PROJECT}"},
								"summary":     "[${steps.transform-1.severity}] ${steps.transform-1.cve_id}",
								"description": "CVE: ${steps.transform-1.cve_id}\nCVSS: ${steps.transform-1.cvss_score}\n\n${steps.transform-1.description}\n\nAffected: ${steps.transform-1.affected}\nRemediation: ${steps.transform-1.remediation}",
								"issuetype":   map[string]interface{}{"name": "Security Vulnerability"},
								"priority":    map[string]interface{}{"name": "${steps.transform-1.severity == 'critical' ? 'Highest' : 'High'}"},
							},
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 900,
					"y": 50,
				},
				"data": map[string]interface{}{
					"name": "Alert Security Team",
					"config": map[string]interface{}{
						"channel": "#security-vulnerabilities",
						"message": "🔴 ${steps.transform-1.severity} Vulnerability Detected\n\nCVE: ${steps.transform-1.cve_id}\nCVSS: ${steps.transform-1.cvss_score}\n\n${steps.transform-1.description}\n\nJira ticket created for tracking.",
					},
				},
			},
			map[string]interface{}{
				"id":   "api-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Log Low/Medium Vuln",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.VULNERABILITY_DB}/api/log",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${credentials.vuln_db_token}",
							"Content-Type":  "application/json",
						},
						"body": "${steps.transform-1}",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{"id": "e1", "source": "trigger-1", "target": "transform-1"},
			map[string]interface{}{"id": "e2", "source": "transform-1", "target": "condition-1"},
			map[string]interface{}{"id": "e3", "source": "condition-1", "target": "api-1", "label": "true"},
			map[string]interface{}{"id": "e4", "source": "api-1", "target": "slack-1"},
			map[string]interface{}{"id": "e5", "source": "condition-1", "target": "api-2", "label": "false"},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Vulnerability Response Automation",
		Description: "Processes vulnerability scanner findings, automatically creates Jira tickets for critical/high vulnerabilities, and notifies the security team. Lower severity vulnerabilities are logged for tracking.",
		Category:    string(CategorySecurity),
		Definition:  defJSON,
		Tags:        []string{"security", "vulnerability", "cve", "jira", "automation", "compliance"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func createAccessReviewAutomationTemplate(now time.Time) *Template {
	definition := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "trigger-1",
				"type": "trigger:schedule",
				"position": map[string]interface{}{
					"x": 100,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Monthly Access Review",
					"config": map[string]interface{}{
						"cron": "0 9 1 * *",
					},
				},
			},
			map[string]interface{}{
				"id":   "api-1",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 100,
				},
				"data": map[string]interface{}{
					"name": "Get Users with Elevated Access",
					"config": map[string]interface{}{
						"method": "GET",
						"url":    "${env.IAM_API}/api/v1/users?role=admin,superuser,privileged",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${credentials.iam_token}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "api-2",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 300,
					"y": 200,
				},
				"data": map[string]interface{}{
					"name": "Get Last Login Data",
					"config": map[string]interface{}{
						"method": "GET",
						"url":    "${env.IAM_API}/api/v1/audit/logins?days=90",
						"headers": map[string]interface{}{
							"Authorization": "Bearer ${credentials.iam_token}",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "transform-1",
				"type": "action:transform",
				"position": map[string]interface{}{
					"x": 500,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Identify Stale Accounts",
					"config": map[string]interface{}{
						"mapping": map[string]interface{}{
							"stale_accounts":  "filter(${steps.api-1.body.users}, u => !contains(map(${steps.api-2.body.logins}, l => l.user_id), u.id))",
							"review_required": "${len(steps.api-1.body.users)}",
							"stale_count":     "len(filter(${steps.api-1.body.users}, u => !contains(map(${steps.api-2.body.logins}, l => l.user_id), u.id)))",
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "api-3",
				"type": "action:http",
				"position": map[string]interface{}{
					"x": 700,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Create Access Review Ticket",
					"config": map[string]interface{}{
						"method": "POST",
						"url":    "${env.JIRA_URL}/rest/api/3/issue",
						"headers": map[string]interface{}{
							"Authorization": "Basic ${credentials.jira_auth}",
							"Content-Type":  "application/json",
						},
						"body": map[string]interface{}{
							"fields": map[string]interface{}{
								"project":     map[string]interface{}{"key": "${env.JIRA_SECURITY_PROJECT}"},
								"summary":     "Monthly Access Review - ${steps.transform-1.review_required} accounts",
								"description": "Monthly access review required.\n\nTotal accounts with elevated access: ${steps.transform-1.review_required}\nAccounts with no login in 90 days: ${steps.transform-1.stale_count}\n\nStale accounts requiring review:\n${join(map(steps.transform-1.stale_accounts, a => a.email), '\\n')}",
								"issuetype":   map[string]interface{}{"name": "Task"},
								"labels":      []string{"access-review", "security", "compliance"},
							},
						},
					},
				},
			},
			map[string]interface{}{
				"id":   "slack-1",
				"type": "slack:send_message",
				"position": map[string]interface{}{
					"x": 900,
					"y": 150,
				},
				"data": map[string]interface{}{
					"name": "Notify Security Team",
					"config": map[string]interface{}{
						"channel": "#security-compliance",
						"message": "📋 Monthly Access Review Initiated\n\nAccounts with elevated access: ${steps.transform-1.review_required}\nStale accounts (no login 90 days): ${steps.transform-1.stale_count}\n\nJira ticket created for tracking. Please complete review within 7 days.",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{"id": "e1", "source": "trigger-1", "target": "api-1"},
			map[string]interface{}{"id": "e2", "source": "trigger-1", "target": "api-2"},
			map[string]interface{}{"id": "e3", "source": "api-1", "target": "transform-1"},
			map[string]interface{}{"id": "e4", "source": "api-2", "target": "transform-1"},
			map[string]interface{}{"id": "e5", "source": "transform-1", "target": "api-3"},
			map[string]interface{}{"id": "e6", "source": "api-3", "target": "slack-1"},
		},
	}

	defJSON, _ := json.Marshal(definition) //nolint:errcheck // static template definition cannot fail

	return &Template{
		TenantID:    nil,
		Name:        "Access Review Automation",
		Description: "Monthly automated access review that identifies users with elevated privileges, detects stale accounts (no login in 90 days), creates compliance tickets, and notifies the security team.",
		Category:    string(CategorySecurity),
		Definition:  defJSON,
		Tags:        []string{"security", "access-review", "compliance", "iam", "audit", "automation"},
		IsPublic:    true,
		CreatedBy:   "system",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// SeedBuiltinTemplates seeds built-in templates into the database for a tenant
func SeedBuiltinTemplates(service *Service, tenantID string) error {
	templates := GetBuiltinTemplates()
	ctx := context.Background()

	for _, tmpl := range templates {
		input := CreateTemplateInput{
			Name:        tmpl.Name,
			Description: tmpl.Description,
			Category:    tmpl.Category,
			Definition:  tmpl.Definition,
			Tags:        tmpl.Tags,
			IsPublic:    tmpl.IsPublic,
		}

		_, err := service.CreateTemplate(ctx, tenantID, "system", input)
		if err != nil {
			// Skip if template already exists
			if strings.Contains(err.Error(), "already exists") {
				continue
			}
			return fmt.Errorf("failed to seed template %s: %w", tmpl.Name, err)
		}
	}

	return nil
}
