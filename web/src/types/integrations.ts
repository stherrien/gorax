// Slack Integration Types
export interface SlackConfig {
  action: 'send_message' | 'send_dm' | 'add_reaction' | 'update_message';
  channel?: string;
  text?: string;
  blocks?: any[];
  user?: string;
  emoji?: string;
  timestamp?: string;
  username?: string;
  iconEmoji?: string;
}

export interface SlackSendMessageConfig {
  channel: string;
  text?: string;
  blocks?: any[];
  threadTs?: string;
  replyBroadcast?: boolean;
  unfurlLinks?: boolean;
  unfurlMedia?: boolean;
  iconEmoji?: string;
  username?: string;
}

export interface SlackSendDMConfig {
  user: string;
  text?: string;
  blocks?: any[];
}

export interface SlackAddReactionConfig {
  channel: string;
  timestamp: string;
  emoji: string;
}

export interface SlackUpdateMessageConfig {
  channel: string;
  timestamp: string;
  text?: string;
  blocks?: any[];
}

// Jira Integration Types
export interface JiraConfig {
  action: 'create_issue' | 'update_issue' | 'add_comment' | 'transition_issue' | 'search_issues';
  project?: string;
  issueType?: string;
  summary?: string;
  description?: string;
  issueKey?: string;
  fields?: Record<string, any>;
  body?: string;
  transitionName?: string;
  jql?: string;
  maxResults?: number;
}

export interface JiraCreateIssueConfig {
  project: string;
  issueType: string;
  summary: string;
  description?: string;
  priority?: string;
  assignee?: string;
  labels?: string[];
  components?: string[];
}

export interface JiraUpdateIssueConfig {
  issueKey: string;
  fields: Record<string, any>;
}

export interface JiraAddCommentConfig {
  issueKey: string;
  body: string;
}

export interface JiraTransitionIssueConfig {
  issueKey: string;
  transitionName: string;
}

export interface JiraSearchIssuesConfig {
  jql: string;
  maxResults?: number;
  startAt?: number;
}

// GitHub Integration Types
export interface GitHubConfig {
  action: 'create_issue' | 'create_pr_comment' | 'add_label';
  owner?: string;
  repo?: string;
  title?: string;
  body?: string;
  number?: number;
  labels?: string[];
}

export interface GitHubCreateIssueConfig {
  owner: string;
  repo: string;
  title: string;
  body?: string;
  labels?: string[];
}

export interface GitHubCreatePRCommentConfig {
  owner: string;
  repo: string;
  number: number;
  body: string;
}

export interface GitHubAddLabelConfig {
  owner: string;
  repo: string;
  number: number;
  labels: string[];
}

// PagerDuty Integration Types
export interface PagerDutyConfig {
  action: 'create_incident' | 'acknowledge_incident' | 'resolve_incident' | 'add_note';
  title?: string;
  service?: string;
  urgency?: 'high' | 'low';
  body?: string;
  incidentKey?: string;
  incidentId?: string;
  content?: string;
}

export interface PagerDutyCreateIncidentConfig {
  title: string;
  service: string;
  urgency?: 'high' | 'low';
  body?: string;
  incidentKey?: string;
}

export interface PagerDutyAcknowledgeIncidentConfig {
  incidentId: string;
}

export interface PagerDutyResolveIncidentConfig {
  incidentId: string;
}

export interface PagerDutyAddNoteConfig {
  incidentId: string;
  content: string;
}

// Integration Action Registry
export interface IntegrationAction {
  id: string;
  name: string;
  description: string;
  integration: 'slack' | 'jira' | 'github' | 'pagerduty';
  configSchema: Record<string, any>;
}

export const integrationActions: IntegrationAction[] = [
  // Slack Actions
  {
    id: 'slack:send_message',
    name: 'Send Message',
    description: 'Send a message to a Slack channel',
    integration: 'slack',
    configSchema: {
      channel: { type: 'string', required: true },
      text: { type: 'string', required: false },
      blocks: { type: 'array', required: false },
    },
  },
  {
    id: 'slack:send_dm',
    name: 'Send Direct Message',
    description: 'Send a direct message to a Slack user',
    integration: 'slack',
    configSchema: {
      user: { type: 'string', required: true },
      text: { type: 'string', required: false },
      blocks: { type: 'array', required: false },
    },
  },
  {
    id: 'slack:add_reaction',
    name: 'Add Reaction',
    description: 'Add an emoji reaction to a Slack message',
    integration: 'slack',
    configSchema: {
      channel: { type: 'string', required: true },
      timestamp: { type: 'string', required: true },
      emoji: { type: 'string', required: true },
    },
  },
  {
    id: 'slack:update_message',
    name: 'Update Message',
    description: 'Update an existing Slack message',
    integration: 'slack',
    configSchema: {
      channel: { type: 'string', required: true },
      timestamp: { type: 'string', required: true },
      text: { type: 'string', required: false },
      blocks: { type: 'array', required: false },
    },
  },
  // Jira Actions
  {
    id: 'jira:create_issue',
    name: 'Create Issue',
    description: 'Create a new Jira issue',
    integration: 'jira',
    configSchema: {
      project: { type: 'string', required: true },
      issueType: { type: 'string', required: true },
      summary: { type: 'string', required: true },
      description: { type: 'string', required: false },
    },
  },
  {
    id: 'jira:update_issue',
    name: 'Update Issue',
    description: 'Update an existing Jira issue',
    integration: 'jira',
    configSchema: {
      issueKey: { type: 'string', required: true },
      fields: { type: 'object', required: true },
    },
  },
  {
    id: 'jira:add_comment',
    name: 'Add Comment',
    description: 'Add a comment to a Jira issue',
    integration: 'jira',
    configSchema: {
      issueKey: { type: 'string', required: true },
      body: { type: 'string', required: true },
    },
  },
  {
    id: 'jira:transition_issue',
    name: 'Transition Issue',
    description: 'Change the status of a Jira issue',
    integration: 'jira',
    configSchema: {
      issueKey: { type: 'string', required: true },
      transitionName: { type: 'string', required: true },
    },
  },
  {
    id: 'jira:search_issues',
    name: 'Search Issues',
    description: 'Search for Jira issues using JQL',
    integration: 'jira',
    configSchema: {
      jql: { type: 'string', required: true },
      maxResults: { type: 'number', required: false },
    },
  },
  // GitHub Actions
  {
    id: 'github:create_issue',
    name: 'Create Issue',
    description: 'Create a new GitHub issue',
    integration: 'github',
    configSchema: {
      owner: { type: 'string', required: true },
      repo: { type: 'string', required: true },
      title: { type: 'string', required: true },
      body: { type: 'string', required: false },
      labels: { type: 'array', required: false },
    },
  },
  {
    id: 'github:create_pr_comment',
    name: 'Create PR Comment',
    description: 'Create a comment on a GitHub pull request',
    integration: 'github',
    configSchema: {
      owner: { type: 'string', required: true },
      repo: { type: 'string', required: true },
      number: { type: 'number', required: true },
      body: { type: 'string', required: true },
    },
  },
  {
    id: 'github:add_label',
    name: 'Add Label',
    description: 'Add labels to a GitHub issue or PR',
    integration: 'github',
    configSchema: {
      owner: { type: 'string', required: true },
      repo: { type: 'string', required: true },
      number: { type: 'number', required: true },
      labels: { type: 'array', required: true },
    },
  },
  // PagerDuty Actions
  {
    id: 'pagerduty:create_incident',
    name: 'Create Incident',
    description: 'Create a new PagerDuty incident',
    integration: 'pagerduty',
    configSchema: {
      title: { type: 'string', required: true },
      service: { type: 'string', required: true },
      urgency: { type: 'string', required: false },
      body: { type: 'string', required: false },
    },
  },
  {
    id: 'pagerduty:acknowledge_incident',
    name: 'Acknowledge Incident',
    description: 'Acknowledge a PagerDuty incident',
    integration: 'pagerduty',
    configSchema: {
      incidentId: { type: 'string', required: true },
    },
  },
  {
    id: 'pagerduty:resolve_incident',
    name: 'Resolve Incident',
    description: 'Resolve a PagerDuty incident',
    integration: 'pagerduty',
    configSchema: {
      incidentId: { type: 'string', required: true },
    },
  },
  {
    id: 'pagerduty:add_note',
    name: 'Add Note',
    description: 'Add a note to a PagerDuty incident',
    integration: 'pagerduty',
    configSchema: {
      incidentId: { type: 'string', required: true },
      content: { type: 'string', required: true },
    },
  },
];
