-- Schedule Templates Library
-- Provides pre-configured schedule patterns for common use cases

-- Schedule templates table
CREATE TABLE schedule_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    cron_expression VARCHAR(100) NOT NULL,
    timezone VARCHAR(100) DEFAULT 'UTC',
    tags TEXT[],
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_schedule_templates_category ON schedule_templates(category);
CREATE INDEX idx_schedule_templates_tags ON schedule_templates USING GIN(tags);
CREATE INDEX idx_schedule_templates_is_system ON schedule_templates(is_system);

-- Seed common schedule templates

-- Frequent schedules
INSERT INTO schedule_templates (name, description, category, cron_expression, timezone, is_system, tags) VALUES
('Every Hour', 'Runs at the start of every hour', 'frequent', '0 * * * *', 'UTC', true, ARRAY['hourly', 'frequent']),
('Every 6 Hours', 'Runs every 6 hours starting at midnight', 'frequent', '0 */6 * * *', 'UTC', true, ARRAY['frequent', 'periodic']),
('Every 4 Hours', 'Runs every 4 hours starting at midnight', 'frequent', '0 */4 * * *', 'UTC', true, ARRAY['frequent', 'periodic']),
('Every 2 Hours', 'Runs every 2 hours starting at midnight', 'frequent', '0 */2 * * *', 'UTC', true, ARRAY['frequent', 'periodic']),
('Every 30 Minutes', 'Runs twice per hour at :00 and :30', 'frequent', '*/30 * * * *', 'UTC', true, ARRAY['frequent', 'real-time']),
('Every 15 Minutes', 'Runs four times per hour', 'frequent', '*/15 * * * *', 'UTC', true, ARRAY['frequent', 'real-time']);

-- Daily schedules
INSERT INTO schedule_templates (name, description, category, cron_expression, timezone, is_system, tags) VALUES
('Daily at Midnight', 'Runs every day at midnight UTC', 'daily', '0 0 * * *', 'UTC', true, ARRAY['daily', 'overnight']),
('Daily at 9 AM', 'Runs every day at 9:00 AM UTC', 'daily', '0 9 * * *', 'UTC', true, ARRAY['daily', 'morning']),
('Daily at 6 PM', 'Runs every day at 6:00 PM UTC', 'daily', '0 18 * * *', 'UTC', true, ARRAY['daily', 'evening']),
('Daily at Noon', 'Runs every day at 12:00 PM UTC', 'daily', '0 12 * * *', 'UTC', true, ARRAY['daily', 'midday']),
('Daily at 2 AM', 'Runs every day at 2:00 AM UTC (good for maintenance)', 'daily', '0 2 * * *', 'UTC', true, ARRAY['daily', 'overnight', 'maintenance']);

-- Business hours schedules
INSERT INTO schedule_templates (name, description, category, cron_expression, timezone, is_system, tags) VALUES
('Weekdays at 9 AM', 'Runs Monday through Friday at 9:00 AM', 'business', '0 9 * * 1-5', 'UTC', true, ARRAY['weekday', 'business-hours', 'morning']),
('Weekdays at 6 PM', 'Runs Monday through Friday at 6:00 PM', 'business', '0 18 * * 1-5', 'UTC', true, ARRAY['weekday', 'business-hours', 'evening']),
('Business Hours Start', 'Runs weekdays at 8:00 AM', 'business', '0 8 * * 1-5', 'UTC', true, ARRAY['weekday', 'business-hours']),
('Business Hours End', 'Runs weekdays at 5:00 PM', 'business', '0 17 * * 1-5', 'UTC', true, ARRAY['weekday', 'business-hours']);

-- Weekly schedules
INSERT INTO schedule_templates (name, description, category, cron_expression, timezone, is_system, tags) VALUES
('Weekly on Monday', 'Runs every Monday at 9:00 AM', 'weekly', '0 9 * * 1', 'UTC', true, ARRAY['weekly', 'monday']),
('Weekly on Friday', 'Runs every Friday at 5:00 PM', 'weekly', '0 17 * * 5', 'UTC', true, ARRAY['weekly', 'friday']),
('Weekly on Sunday', 'Runs every Sunday at midnight', 'weekly', '0 0 * * 0', 'UTC', true, ARRAY['weekly', 'sunday', 'weekend']),
('Bi-Weekly on Monday', 'Runs every other Monday at 9:00 AM', 'weekly', '0 9 1-7,15-21 * 1', 'UTC', true, ARRAY['biweekly', 'monday']);

-- Monthly schedules
INSERT INTO schedule_templates (name, description, category, cron_expression, timezone, is_system, tags) VALUES
('Monthly on 1st', 'Runs on the 1st of every month at midnight', 'monthly', '0 0 1 * *', 'UTC', true, ARRAY['monthly', 'month-start']),
('Monthly on 15th', 'Runs on the 15th of every month at noon', 'monthly', '0 12 15 * *', 'UTC', true, ARRAY['monthly', 'mid-month']),
('Monthly Last Day', 'Runs on the last day of every month', 'monthly', '0 0 28-31 * *', 'UTC', true, ARRAY['monthly', 'month-end']),
('Quarterly', 'Runs on the first day of Jan, Apr, Jul, Oct', 'monthly', '0 0 1 1,4,7,10 *', 'UTC', true, ARRAY['quarterly']);

-- Compliance schedules
INSERT INTO schedule_templates (name, description, category, cron_expression, timezone, is_system, tags) VALUES
('SOC2 Daily Scan', 'Daily SOC2 compliance check at 2:00 AM', 'compliance', '0 2 * * *', 'UTC', true, ARRAY['compliance', 'soc2', 'security', 'daily']),
('FedRAMP Weekly Review', 'Weekly FedRAMP compliance review on Monday', 'compliance', '0 6 * * 1', 'UTC', true, ARRAY['compliance', 'fedramp', 'weekly']),
('HIPAA Audit Monthly', 'Monthly HIPAA audit on the 1st at 3:00 AM', 'compliance', '0 3 1 * *', 'UTC', true, ARRAY['compliance', 'hipaa', 'monthly']),
('PCI-DSS Daily Check', 'Daily PCI-DSS compliance check', 'compliance', '0 3 * * *', 'UTC', true, ARRAY['compliance', 'pci-dss', 'security', 'daily']),
('GDPR Weekly Audit', 'Weekly GDPR compliance audit on Sunday', 'compliance', '0 4 * * 0', 'UTC', true, ARRAY['compliance', 'gdpr', 'weekly']),
('ISO 27001 Monthly Report', 'Monthly ISO 27001 compliance report', 'compliance', '0 5 1 * *', 'UTC', true, ARRAY['compliance', 'iso-27001', 'monthly']);

-- Security schedules
INSERT INTO schedule_templates (name, description, category, cron_expression, timezone, is_system, tags) VALUES
('Wiz Hourly Scan', 'Hourly vulnerability scan via Wiz', 'security', '0 * * * *', 'UTC', true, ARRAY['security', 'vulnerability', 'wiz', 'hourly']),
('Qualys Daily Scan', 'Daily Qualys vulnerability assessment at 1:00 AM', 'security', '0 1 * * *', 'UTC', true, ARRAY['security', 'vulnerability', 'qualys', 'daily']),
('Tenable Weekly Scan', 'Weekly Tenable security scan on Saturday', 'security', '0 2 * * 6', 'UTC', true, ARRAY['security', 'vulnerability', 'tenable', 'weekly']),
('Security Posture Daily', 'Daily security posture assessment', 'security', '0 5 * * *', 'UTC', true, ARRAY['security', 'posture', 'daily']),
('Threat Intelligence Hourly', 'Hourly threat intelligence feed update', 'security', '0 * * * *', 'UTC', true, ARRAY['security', 'threat-intel', 'hourly']);

-- Sync and integration schedules
INSERT INTO schedule_templates (name, description, category, cron_expression, timezone, is_system, tags) VALUES
('Asset Sync Every 6 Hours', 'Sync assets every 6 hours', 'sync', '0 */6 * * *', 'UTC', true, ARRAY['sync', 'assets', 'periodic']),
('User Directory Sync Daily', 'Daily user directory synchronization at 3:00 AM', 'sync', '0 3 * * *', 'UTC', true, ARRAY['sync', 'users', 'daily']),
('Config Backup Hourly', 'Hourly configuration backup', 'sync', '0 * * * *', 'UTC', true, ARRAY['sync', 'backup', 'hourly']),
('Database Sync Every 4 Hours', 'Database synchronization every 4 hours', 'sync', '0 */4 * * *', 'UTC', true, ARRAY['sync', 'database', 'periodic']),
('Log Aggregation Every 15 Min', 'Aggregate logs every 15 minutes', 'sync', '*/15 * * * *', 'UTC', true, ARRAY['sync', 'logs', 'frequent']);

-- Monitoring schedules
INSERT INTO schedule_templates (name, description, category, cron_expression, timezone, is_system, tags) VALUES
('Health Check Every 5 Minutes', 'System health check every 5 minutes', 'monitoring', '*/5 * * * *', 'UTC', true, ARRAY['monitoring', 'health-check', 'frequent']),
('Metrics Collection Hourly', 'Collect system metrics every hour', 'monitoring', '0 * * * *', 'UTC', true, ARRAY['monitoring', 'metrics', 'hourly']),
('Daily Report Generation', 'Generate daily monitoring report at 7:00 AM', 'monitoring', '0 7 * * *', 'UTC', true, ARRAY['monitoring', 'reports', 'daily']),
('Weekly Summary Report', 'Generate weekly summary on Monday at 8:00 AM', 'monitoring', '0 8 * * 1', 'UTC', true, ARRAY['monitoring', 'reports', 'weekly']);
