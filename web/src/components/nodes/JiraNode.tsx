import { Handle, Position } from '@xyflow/react';

interface JiraNodeData {
  action: 'create_issue' | 'update_issue' | 'add_comment' | 'transition_issue' | 'search_issues';
  config: {
    project?: string;
    issueType?: string;
    summary?: string;
    issueKey?: string;
    jql?: string;
  };
}

interface JiraNodeProps {
  id: string;
  data: JiraNodeData;
  selected?: boolean;
}

const JiraNode = ({ data, selected }: JiraNodeProps) => {
  const getActionLabel = () => {
    switch (data.action) {
      case 'create_issue':
        return 'Create Issue';
      case 'update_issue':
        return 'Update Issue';
      case 'add_comment':
        return 'Add Comment';
      case 'transition_issue':
        return 'Transition Issue';
      case 'search_issues':
        return 'Search Issues';
      default:
        return 'Jira Action';
    }
  };

  const getConfigSummary = () => {
    switch (data.action) {
      case 'create_issue':
        return data.config.project
          ? `${data.config.project} - ${data.config.issueType || 'Task'}`
          : 'No project set';
      case 'update_issue':
      case 'add_comment':
      case 'transition_issue':
        return data.config.issueKey || 'No issue key set';
      case 'search_issues':
        return data.config.jql ? 'JQL query set' : 'No query set';
      default:
        return '';
    }
  };

  return (
    <div
      className={`px-4 py-2 shadow-md rounded-md border-2 bg-white min-w-[200px] ${
        selected ? 'border-blue-500' : 'border-gray-300'
      }`}
    >
      <Handle type="target" position={Position.Top} className="w-3 h-3" />

      <div className="flex items-center gap-2">
        <div className="flex h-8 w-8 items-center justify-center rounded bg-blue-100">
          <svg
            className="h-5 w-5 text-blue-600"
            fill="currentColor"
            viewBox="0 0 24 24"
          >
            <path d="M11.571 11.513H0a5.218 5.218 0 0 0 5.232 5.215h2.13v2.057A5.215 5.215 0 0 0 12.575 24V12.518a1.005 1.005 0 0 0-1.005-1.005zm5.723-5.756H24a5.218 5.218 0 0 0-5.232-5.214h-2.13V2.6a5.215 5.215 0 0 0-5.213 5.213v11.482a1.005 1.005 0 0 0 1.005 1.005h5.723v-11.48a1.062 1.062 0 0 1 1.062-1.062z" />
          </svg>
        </div>

        <div className="flex-1">
          <div className="text-sm font-semibold text-gray-700">
            {getActionLabel()}
          </div>
          <div className="text-xs text-gray-500 truncate">
            {getConfigSummary()}
          </div>
        </div>
      </div>

      <Handle type="source" position={Position.Bottom} className="w-3 h-3" />
    </div>
  );
};

export default JiraNode;
