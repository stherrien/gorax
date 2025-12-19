import { Handle, Position } from '@xyflow/react';
import type { ActionNodeData } from '../../stores/workflowStore';

interface FormulaNodeProps {
  data: ActionNodeData & {
    config?: {
      expression?: string;
      output_variable?: string;
    };
  };
  selected?: boolean;
}

export default function FormulaNode({ data, selected }: FormulaNodeProps) {
  const expression = data.config?.expression || '';
  const truncatedExpression = expression.length > 30
    ? expression.substring(0, 30) + '...'
    : expression;

  return (
    <div
      className={`
        px-4 py-3 rounded-lg shadow-lg min-w-[180px]
        bg-gradient-to-br from-purple-500 to-indigo-600
        ${selected ? 'ring-2 ring-white ring-offset-2 ring-offset-gray-900' : ''}
      `}
      data-testid="formula-node"
    >
      {/* Input handle */}
      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 bg-white border-2 border-purple-500"
      />

      <div className="flex items-start space-x-2">
        <span className="text-lg">🔢</span>
        <div className="flex-1 min-w-0">
          <p className="text-white font-medium text-sm">
            {data.label || 'Formula'}
          </p>
          {expression && (
            <p
              className="text-white/80 text-xs font-mono mt-1 truncate"
              title={expression}
            >
              {truncatedExpression}
            </p>
          )}
          {!expression && (
            <p className="text-white/60 text-xs italic mt-1">
              No formula set
            </p>
          )}
        </div>
      </div>

      {/* Output handle */}
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-3 h-3 bg-white border-2 border-purple-500"
      />
    </div>
  );
}
