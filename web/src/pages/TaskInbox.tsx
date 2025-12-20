import React, { useState } from 'react';
import { useTasks } from '../hooks/useTasks';
import { TaskCard } from '../components/tasks/TaskCard';
import { ApprovalDialog } from '../components/tasks/ApprovalDialog';
import { HumanTask } from '../api/tasks';

type FilterTab = 'all' | 'pending' | 'completed';

export const TaskInbox: React.FC = () => {
  const [activeTab, setActiveTab] = useState<FilterTab>('pending');
  const [selectedTask, setSelectedTask] = useState<HumanTask | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const statusFilter =
    activeTab === 'pending'
      ? 'pending'
      : activeTab === 'completed'
      ? undefined
      : undefined;

  const { data, isLoading, error, refetch } = useTasks({
    status: statusFilter,
    limit: 50,
  });

  const handleTaskClick = (task: HumanTask) => {
    if (task.status === 'pending') {
      setSelectedTask(task);
      setIsDialogOpen(true);
    }
  };

  const handleCloseDialog = () => {
    setIsDialogOpen(false);
    setSelectedTask(null);
    refetch();
  };

  const tabs: { key: FilterTab; label: string }[] = [
    { key: 'pending', label: 'Pending' },
    { key: 'completed', label: 'Completed' },
    { key: 'all', label: 'All Tasks' },
  ];

  const filteredTasks = data?.tasks.filter((task) => {
    if (activeTab === 'pending') return task.status === 'pending';
    if (activeTab === 'completed')
      return task.status !== 'pending' && task.status !== 'cancelled';
    return true;
  });

  const pendingCount =
    data?.tasks.filter((task) => task.status === 'pending').length || 0;

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">Task Inbox</h1>
              <p className="mt-1 text-sm text-gray-600">
                Manage your approval and input tasks
              </p>
            </div>
            <div className="flex items-center gap-4">
              <div className="bg-blue-100 text-blue-800 px-4 py-2 rounded-full font-medium">
                {pendingCount} Pending
              </div>
              <button
                onClick={() => refetch()}
                className="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors"
                title="Refresh"
              >
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
                    d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                  />
                </svg>
              </button>
            </div>
          </div>

          {/* Tabs */}
          <div className="mt-6 flex space-x-8 border-b border-gray-200">
            {tabs.map((tab) => (
              <button
                key={tab.key}
                onClick={() => setActiveTab(tab.key)}
                className={`pb-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                  activeTab === tab.key
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {isLoading && (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500" />
          </div>
        )}

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4">
            <div className="flex items-center">
              <svg
                className="w-5 h-5 text-red-500 mr-2"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              <p className="text-red-800">
                Failed to load tasks. Please try again.
              </p>
            </div>
          </div>
        )}

        {!isLoading && !error && filteredTasks && filteredTasks.length === 0 && (
          <div className="text-center py-12">
            <svg
              className="mx-auto h-12 w-12 text-gray-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
            <h3 className="mt-2 text-sm font-medium text-gray-900">
              No tasks found
            </h3>
            <p className="mt-1 text-sm text-gray-500">
              {activeTab === 'pending'
                ? "You don't have any pending tasks."
                : 'No tasks to display.'}
            </p>
          </div>
        )}

        {!isLoading && !error && filteredTasks && filteredTasks.length > 0 && (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {filteredTasks.map((task) => (
              <TaskCard
                key={task.id}
                task={task}
                onClick={() => handleTaskClick(task)}
              />
            ))}
          </div>
        )}
      </div>

      {/* Approval Dialog */}
      {selectedTask && (
        <ApprovalDialog
          task={selectedTask}
          isOpen={isDialogOpen}
          onClose={handleCloseDialog}
        />
      )}
    </div>
  );
};

export default TaskInbox;
