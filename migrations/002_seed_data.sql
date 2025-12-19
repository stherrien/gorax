-- Seed data for development

-- Create a default tenant for development
INSERT INTO tenants (id, name, subdomain, status, tier, settings, quotas)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Development Tenant',
    'dev',
    'active',
    'professional',
    '{"default_timezone": "UTC", "webhook_secret": "dev-secret-key"}',
    '{"max_workflows": 50, "max_executions_per_day": 5000, "max_concurrent_executions": 10}'
);

-- Create a development user
INSERT INTO users (id, tenant_id, kratos_identity_id, email, role, status)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000001',
    'dev-kratos-id',
    'dev@example.com',
    'admin',
    'active'
);

-- Create a sample workflow
INSERT INTO workflows (id, tenant_id, name, description, definition, status, version, created_by)
VALUES (
    'b0000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'Hello World Workflow',
    'A simple workflow that makes an HTTP request',
    '{
        "nodes": [
            {
                "id": "trigger-1",
                "type": "trigger:webhook",
                "name": "Webhook Trigger",
                "position": {"x": 100, "y": 100},
                "config": {
                    "path": "/hello",
                    "auth_type": "none"
                }
            },
            {
                "id": "http-1",
                "type": "action:http",
                "name": "HTTP Request",
                "position": {"x": 100, "y": 250},
                "config": {
                    "method": "GET",
                    "url": "https://httpbin.org/get",
                    "headers": {}
                }
            },
            {
                "id": "transform-1",
                "type": "action:transform",
                "name": "Extract Data",
                "position": {"x": 100, "y": 400},
                "config": {
                    "mapping": {
                        "origin": "steps.http-1.body.origin",
                        "timestamp": "steps.http-1.headers.Date"
                    }
                }
            }
        ],
        "edges": [
            {"id": "e1", "source": "trigger-1", "target": "http-1"},
            {"id": "e2", "source": "http-1", "target": "transform-1"}
        ]
    }',
    'active',
    1,
    '00000000-0000-0000-0000-000000000002'
);

-- Create a webhook for the sample workflow
INSERT INTO webhooks (id, tenant_id, workflow_id, node_id, path, secret, auth_type, enabled)
VALUES (
    'c0000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'b0000000-0000-0000-0000-000000000001',
    'trigger-1',
    'b0000000-0000-0000-0000-000000000001/c0000000-0000-0000-0000-000000000001',
    'webhook-secret-123',
    'none',
    true
);
