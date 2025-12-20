import React from 'react';
import { Handle, Position, NodeProps } from 'reactflow';

interface SlackNodeData {
  action: 'send_message' | 'send_dm' | 'add_reaction' | 'update_message';
  config: {
    channel?: string;
    text?: string;
    user?: string;
    emoji?: string;
    timestamp?: string;
  };
}

const SlackNode: React.FC<NodeProps<SlackNodeData>> = ({ data, selected }) => {
  const getActionLabel = () => {
    switch (data.action) {
      case 'send_message':
        return 'Send Message';
      case 'send_dm':
        return 'Send DM';
      case 'add_reaction':
        return 'Add Reaction';
      case 'update_message':
        return 'Update Message';
      default:
        return 'Slack Action';
    }
  };

  const getConfigSummary = () => {
    switch (data.action) {
      case 'send_message':
        return data.config.channel || 'No channel set';
      case 'send_dm':
        return data.config.user || 'No user set';
      case 'add_reaction':
        return data.config.emoji || 'No emoji set';
      case 'update_message':
        return 'Update existing message';
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
        <div className="flex h-8 w-8 items-center justify-center rounded bg-purple-100">
          <svg
            className="h-5 w-5 text-purple-600"
            fill="currentColor"
            viewBox="0 0 24 24"
          >
            <path d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zM6.313 15.165a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313zM8.834 5.042a2.528 2.528 0 0 1-2.521-2.52A2.528 2.528 0 0 1 8.834 0a2.528 2.528 0 0 1 2.521 2.522v2.52H8.834zM8.834 6.313a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521H2.522A2.528 2.528 0 0 1 0 8.834a2.528 2.528 0 0 1 2.522-2.521h6.312zM18.956 8.834a2.528 2.528 0 0 1 2.522-2.521A2.528 2.528 0 0 1 24 8.834a2.528 2.528 0 0 1-2.522 2.521h-2.522V8.834zM17.688 8.834a2.528 2.528 0 0 1-2.523 2.521 2.527 2.527 0 0 1-2.52-2.521V2.522A2.527 2.527 0 0 1 15.165 0a2.528 2.528 0 0 1 2.523 2.522v6.312zM15.165 18.956a2.528 2.528 0 0 1 2.523 2.522A2.528 2.528 0 0 1 15.165 24a2.527 2.527 0 0 1-2.52-2.522v-2.522h2.52zM15.165 17.688a2.527 2.527 0 0 1-2.52-2.523 2.526 2.526 0 0 1 2.52-2.52h6.313A2.527 2.527 0 0 1 24 15.165a2.528 2.528 0 0 1-2.522 2.523h-6.313z" />
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

export default SlackNode;
