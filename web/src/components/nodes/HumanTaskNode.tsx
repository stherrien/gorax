import React from 'react';
import { Handle, Position, NodeProps } from 'reactflow';

interface HumanTaskNodeData {
  label: string;
  task_type?: 'approval' | 'input' | 'review';
  title?: string;
  assignees?: string[];
  due_date?: string;
  config?: {
    timeout?: string;
    on_timeout?: string;
  };
}

export const HumanTaskNode: React.FC<NodeProps<HumanTaskNodeData>> = ({
  data,
  selected,
}) => {
  const getTaskTypeIcon = () => {
    switch (data.task_type) {
      case 'approval':
        return (
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        );
      case 'input':
        return (
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
            />
          </svg>
        );
      case 'review':
        return (
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
            />
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
            />
          </svg>
        );
      default:
        return (
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
            />
          </svg>
        );
    }
  };

  const getTaskTypeLabel = () => {
    switch (data.task_type) {
      case 'approval':
        return 'Approval';
      case 'input':
        return 'Input';
      case 'review':
        return 'Review';
      default:
        return 'Human Task';
    }
  };

  return (
    <div
      className={`bg-white rounded-lg border-2 shadow-md min-w-[200px] ${
        selected ? 'border-blue-500' : 'border-purple-300'
      }`}
    >
      <Handle type="target" position={Position.Top} className="w-3 h-3" />

      {/* Header */}
      <div className="bg-purple-500 text-white px-4 py-2 rounded-t-lg flex items-center gap-2">
        {getTaskTypeIcon()}
        <span className="font-semibold text-sm">{getTaskTypeLabel()}</span>
      </div>

      {/* Body */}
      <div className="px-4 py-3">
        <div className="font-medium text-gray-900 mb-2">
          {data.title || data.label}
        </div>

        {data.assignees && data.assignees.length > 0 && (
          <div className="flex items-center gap-2 text-xs text-gray-600 mb-2">
            <svg
              className="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
              />
            </svg>
            <span className="truncate">
              {data.assignees.length} assignee{data.assignees.length !== 1 ? 's' : ''}
            </span>
          </div>
        )}

        {data.due_date && (
          <div className="flex items-center gap-2 text-xs text-gray-600 mb-2">
            <svg
              className="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>Due: {data.due_date}</span>
          </div>
        )}

        {data.config?.timeout && (
          <div className="flex items-center gap-2 text-xs text-gray-600">
            <svg
              className="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>Timeout: {data.config.timeout}</span>
          </div>
        )}
      </div>

      {/* Footer */}
      {data.config?.on_timeout && (
        <div className="px-4 py-2 bg-gray-50 rounded-b-lg border-t border-gray-200">
          <div className="text-xs text-gray-600">
            On timeout: <span className="font-medium">{data.config.on_timeout}</span>
          </div>
        </div>
      )}

      <Handle type="source" position={Position.Bottom} className="w-3 h-3" />
    </div>
  );
};

export default HumanTaskNode;
