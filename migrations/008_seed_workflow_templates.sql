-- Seed default workflow templates
-- These are public templates available to all tenants

-- Security Scan Pipeline Template
INSERT INTO workflow_templates (
    name,
    description,
    category,
    definition,
    tags,
    is_public,
    created_by
) VALUES (
    'Security Scan Pipeline',
    'Automated security scanning workflow that triggers on code push and runs vulnerability checks',
    'security',
    '{
      "nodes": [
        {
          "id": "node-1",
          "type": "trigger:webhook",
          "position": {"x": 100, "y": 100},
          "data": {
            "name": "Code Push Trigger",
            "config": {
              "path": "/security-scan",
              "auth_type": "signature"
            }
          }
        },
        {
          "id": "node-2",
          "type": "action:http",
          "position": {"x": 300, "y": 100},
          "data": {
            "name": "Run Dependency Scan",
            "config": {
              "method": "POST",
              "url": "https://security-scanner.example.com/scan",
              "headers": {"Content-Type": "application/json"}
            }
          }
        },
        {
          "id": "node-3",
          "type": "control:if",
          "position": {"x": 500, "y": 100},
          "data": {
            "name": "Check Vulnerabilities",
            "config": {
              "condition": "${steps.node2.output.vulnerabilities.length > 0}"
            }
          }
        },
        {
          "id": "node-4",
          "type": "action:email",
          "position": {"x": 700, "y": 50},
          "data": {
            "name": "Send Alert",
            "config": {
              "to": "security@example.com",
              "subject": "Security Vulnerabilities Detected"
            }
          }
        }
      ],
      "edges": [
        {"id": "e1", "source": "node-1", "target": "node-2"},
        {"id": "e2", "source": "node-2", "target": "node-3"},
        {"id": "e3", "source": "node-3", "target": "node-4", "label": "true"}
      ]
    }'::jsonb,
    ARRAY['security', 'automation', 'scanning'],
    true,
    '00000000-0000-0000-0000-000000000000'
);

-- API Integration Template
INSERT INTO workflow_templates (
    name,
    description,
    category,
    definition,
    tags,
    is_public,
    created_by
) VALUES (
    'API Data Integration',
    'Fetch data from external API, transform it, and send to another service',
    'integration',
    '{
      "nodes": [
        {
          "id": "node-1",
          "type": "trigger:schedule",
          "position": {"x": 100, "y": 100},
          "data": {
            "name": "Hourly Schedule",
            "config": {
              "cron": "0 * * * *",
              "timezone": "UTC"
            }
          }
        },
        {
          "id": "node-2",
          "type": "action:http",
          "position": {"x": 300, "y": 100},
          "data": {
            "name": "Fetch External Data",
            "config": {
              "method": "GET",
              "url": "https://api.example.com/data"
            }
          }
        },
        {
          "id": "node-3",
          "type": "action:transform",
          "position": {"x": 500, "y": 100},
          "data": {
            "name": "Transform Data",
            "config": {
              "mapping": {
                "id": "${item.external_id}",
                "name": "${item.display_name}",
                "timestamp": "${now()}"
              }
            }
          }
        },
        {
          "id": "node-4",
          "type": "action:http",
          "position": {"x": 700, "y": 100},
          "data": {
            "name": "Send to Destination",
            "config": {
              "method": "POST",
              "url": "https://destination.example.com/ingest"
            }
          }
        }
      ],
      "edges": [
        {"id": "e1", "source": "node-1", "target": "node-2"},
        {"id": "e2", "source": "node-2", "target": "node-3"},
        {"id": "e3", "source": "node-3", "target": "node-4"}
      ]
    }'::jsonb,
    ARRAY['integration', 'api', 'etl'],
    true,
    '00000000-0000-0000-0000-000000000000'
);

-- Scheduled Report Template
INSERT INTO workflow_templates (
    name,
    description,
    category,
    definition,
    tags,
    is_public,
    created_by
) VALUES (
    'Daily Scheduled Report',
    'Generate and email a daily report with data aggregation',
    'monitoring',
    '{
      "nodes": [
        {
          "id": "node-1",
          "type": "trigger:schedule",
          "position": {"x": 100, "y": 100},
          "data": {
            "name": "Daily at 9 AM",
            "config": {
              "cron": "0 9 * * *",
              "timezone": "America/New_York"
            }
          }
        },
        {
          "id": "node-2",
          "type": "action:http",
          "position": {"x": 300, "y": 100},
          "data": {
            "name": "Fetch Metrics",
            "config": {
              "method": "GET",
              "url": "https://metrics.example.com/daily"
            }
          }
        },
        {
          "id": "node-3",
          "type": "action:formula",
          "position": {"x": 500, "y": 100},
          "data": {
            "name": "Calculate Summary",
            "config": {
              "expression": "sum(${steps.node2.output.metrics.values})",
              "output_variable": "total"
            }
          }
        },
        {
          "id": "node-4",
          "type": "action:email",
          "position": {"x": 700, "y": 100},
          "data": {
            "name": "Email Report",
            "config": {
              "to": "team@example.com",
              "subject": "Daily Metrics Report",
              "body": "Total: ${steps.node3.output.total}"
            }
          }
        }
      ],
      "edges": [
        {"id": "e1", "source": "node-1", "target": "node-2"},
        {"id": "e2", "source": "node-2", "target": "node-3"},
        {"id": "e3", "source": "node-3", "target": "node-4"}
      ]
    }'::jsonb,
    ARRAY['monitoring', 'reporting', 'scheduled'],
    true,
    '00000000-0000-0000-0000-000000000000'
);

-- Slack Notification Template
INSERT INTO workflow_templates (
    name,
    description,
    category,
    definition,
    tags,
    is_public,
    created_by
) VALUES (
    'Slack Alert Workflow',
    'Webhook-triggered workflow that sends formatted alerts to Slack',
    'integration',
    '{
      "nodes": [
        {
          "id": "node-1",
          "type": "trigger:webhook",
          "position": {"x": 100, "y": 100},
          "data": {
            "name": "Alert Webhook",
            "config": {
              "path": "/alerts",
              "auth_type": "signature"
            }
          }
        },
        {
          "id": "node-2",
          "type": "action:transform",
          "position": {"x": 300, "y": 100},
          "data": {
            "name": "Format Message",
            "config": {
              "mapping": {
                "text": "Alert: ${trigger.severity} - ${trigger.message}",
                "channel": "#alerts"
              }
            }
          }
        },
        {
          "id": "node-3",
          "type": "slack:send_message",
          "position": {"x": 500, "y": 100},
          "data": {
            "name": "Post to Slack",
            "config": {
              "channel": "${steps.node2.output.channel}",
              "text": "${steps.node2.output.text}"
            }
          }
        }
      ],
      "edges": [
        {"id": "e1", "source": "node-1", "target": "node-2"},
        {"id": "e2", "source": "node-2", "target": "node-3"}
      ]
    }'::jsonb,
    ARRAY['slack', 'notification', 'alert'],
    true,
    '00000000-0000-0000-0000-000000000000'
);

-- Data Processing Pipeline Template
INSERT INTO workflow_templates (
    name,
    description,
    category,
    definition,
    tags,
    is_public,
    created_by
) VALUES (
    'Batch Data Processing',
    'Process batches of data with parallel execution and error handling',
    'dataops',
    '{
      "nodes": [
        {
          "id": "node-1",
          "type": "trigger:webhook",
          "position": {"x": 100, "y": 100},
          "data": {
            "name": "Data Batch Trigger",
            "config": {
              "path": "/process-batch",
              "auth_type": "api_key"
            }
          }
        },
        {
          "id": "node-2",
          "type": "control:loop",
          "position": {"x": 300, "y": 100},
          "data": {
            "name": "Process Each Item",
            "config": {
              "source": "${trigger.items}",
              "item_variable": "item",
              "max_iterations": 1000
            }
          }
        },
        {
          "id": "node-3",
          "type": "action:transform",
          "position": {"x": 500, "y": 100},
          "data": {
            "name": "Transform Item",
            "config": {
              "mapping": {
                "processed": "true",
                "data": "${item.data}",
                "timestamp": "${now()}"
              }
            }
          }
        },
        {
          "id": "node-4",
          "type": "action:http",
          "position": {"x": 700, "y": 100},
          "data": {
            "name": "Save Result",
            "config": {
              "method": "POST",
              "url": "https://storage.example.com/save"
            }
          }
        }
      ],
      "edges": [
        {"id": "e1", "source": "node-1", "target": "node-2"},
        {"id": "e2", "source": "node-2", "target": "node-3"},
        {"id": "e3", "source": "node-3", "target": "node-4"}
      ]
    }'::jsonb,
    ARRAY['dataops', 'batch', 'etl'],
    true,
    '00000000-0000-0000-0000-000000000000'
);
