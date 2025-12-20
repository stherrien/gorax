export { default as SlackNode } from './SlackNode';
export { default as JiraNode } from './JiraNode';
export { default as GitHubNode } from './GitHubNode';
export { default as PagerDutyNode } from './PagerDutyNode';

// Node type definitions for React Flow
export const nodeTypes = {
  slack: SlackNode,
  jira: JiraNode,
  github: GitHubNode,
  pagerduty: PagerDutyNode,
};
