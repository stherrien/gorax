import React from 'react';
import { Handle, Position, NodeProps } from 'reactflow';

interface GitHubNodeData {
  action: 'create_issue' | 'create_pr_comment' | 'add_label';
  config: {
    owner?: string;
    repo?: string;
    title?: string;
    number?: number;
    labels?: string[];
  };
}

const GitHubNode: React.FC<NodeProps<GitHubNodeData>> = ({ data, selected }) => {
  const getActionLabel = () => {
    switch (data.action) {
      case 'create_issue':
        return 'Create Issue';
      case 'create_pr_comment':
        return 'Create PR Comment';
      case 'add_label':
        return 'Add Label';
      default:
        return 'GitHub Action';
    }
  };

  const getConfigSummary = () => {
    const repoName = data.config.owner && data.config.repo
      ? `${data.config.owner}/${data.config.repo}`
      : 'No repo set';

    switch (data.action) {
      case 'create_issue':
        return repoName;
      case 'create_pr_comment':
        return data.config.number
          ? `${repoName}#${data.config.number}`
          : repoName;
      case 'add_label':
        return data.config.labels && data.config.labels.length > 0
          ? data.config.labels.join(', ')
          : 'No labels set';
      default:
        return repoName;
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
        <div className="flex h-8 w-8 items-center justify-center rounded bg-gray-100">
          <svg
            className="h-5 w-5 text-gray-800"
            fill="currentColor"
            viewBox="0 0 24 24"
          >
            <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z" />
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

export default GitHubNode;
