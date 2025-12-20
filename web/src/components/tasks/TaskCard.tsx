import React from 'react';
import { HumanTask } from '../../api/tasks';
import { formatDistanceToNow } from 'date-fns';

interface TaskCardProps {
  task: HumanTask;
  onClick?: () => void;
}

export const TaskCard: React.FC<TaskCardProps> = ({ task, onClick }) => {
  const isOverdue = task.due_date && new Date(task.due_date) < new Date();

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending':
        return 'bg-yellow-100 text-yellow-800';
      case 'approved':
        return 'bg-green-100 text-green-800';
      case 'rejected':
        return 'bg-red-100 text-red-800';
      case 'expired':
        return 'bg-gray-100 text-gray-800';
      case 'cancelled':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getTaskTypeLabel = (type: string) => {
    switch (type) {
      case 'approval':
        return 'Approval';
      case 'input':
        return 'Input';
      case 'review':
        return 'Review';
      default:
        return type;
    }
  };

  return (
    <div
      className={`border rounded-lg p-4 hover:shadow-md transition-shadow cursor-pointer ${
        isOverdue ? 'border-red-300 bg-red-50' : 'border-gray-200 bg-white'
      }`}
      onClick={onClick}
    >
      <div className="flex justify-between items-start mb-2">
        <div className="flex-1">
          <h3 className="text-lg font-semibold text-gray-900">{task.title}</h3>
          {task.description && (
            <p className="text-sm text-gray-600 mt-1">{task.description}</p>
          )}
        </div>
        <span
          className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(
            task.status
          )}`}
        >
          {task.status}
        </span>
      </div>

      <div className="flex items-center gap-4 text-sm text-gray-500 mt-3">
        <span className="inline-flex items-center">
          <svg
            className="w-4 h-4 mr-1"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
            />
          </svg>
          {getTaskTypeLabel(task.task_type)}
        </span>

        {task.due_date && (
          <span
            className={`inline-flex items-center ${
              isOverdue ? 'text-red-600 font-medium' : ''
            }`}
          >
            <svg
              className="w-4 h-4 mr-1"
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
            {isOverdue ? 'Overdue: ' : 'Due: '}
            {formatDistanceToNow(new Date(task.due_date), { addSuffix: true })}
          </span>
        )}

        <span className="inline-flex items-center">
          <svg
            className="w-4 h-4 mr-1"
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
          Created {formatDistanceToNow(new Date(task.created_at), { addSuffix: true })}
        </span>
      </div>

      {task.assignees_list && task.assignees_list.length > 0 && (
        <div className="mt-3 flex items-center">
          <span className="text-xs text-gray-500 mr-2">Assigned to:</span>
          <div className="flex -space-x-2">
            {task.assignees_list.slice(0, 3).map((assignee, index) => (
              <div
                key={index}
                className="w-6 h-6 rounded-full bg-blue-500 border-2 border-white flex items-center justify-center text-xs text-white font-medium"
                title={assignee}
              >
                {assignee.substring(0, 2).toUpperCase()}
              </div>
            ))}
            {task.assignees_list.length > 3 && (
              <div className="w-6 h-6 rounded-full bg-gray-400 border-2 border-white flex items-center justify-center text-xs text-white font-medium">
                +{task.assignees_list.length - 3}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default TaskCard;
