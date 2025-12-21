import { Handle, Position } from '@xyflow/react';

interface PagerDutyNodeData {
  action: 'create_incident' | 'acknowledge_incident' | 'resolve_incident' | 'add_note';
  config: {
    title?: string;
    service?: string;
    urgency?: 'high' | 'low';
    incidentId?: string;
    content?: string;
  };
}

interface PagerDutyNodeProps {
  id: string;
  data: PagerDutyNodeData;
  selected?: boolean;
}

const PagerDutyNode = ({ data, selected }: PagerDutyNodeProps) => {
  const getActionLabel = () => {
    switch (data.action) {
      case 'create_incident':
        return 'Create Incident';
      case 'acknowledge_incident':
        return 'Acknowledge Incident';
      case 'resolve_incident':
        return 'Resolve Incident';
      case 'add_note':
        return 'Add Note';
      default:
        return 'PagerDuty Action';
    }
  };

  const getConfigSummary = () => {
    switch (data.action) {
      case 'create_incident': {
        const urgency = data.config.urgency ? ` (${data.config.urgency})` : '';
        return data.config.service
          ? `Service: ${data.config.service}${urgency}`
          : 'No service set';
      }
      case 'acknowledge_incident':
      case 'resolve_incident':
      case 'add_note':
        return data.config.incidentId || 'No incident ID set';
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
        <div className="flex h-8 w-8 items-center justify-center rounded bg-green-100">
          <svg
            className="h-5 w-5 text-green-600"
            fill="currentColor"
            viewBox="0 0 24 24"
          >
            <path d="M16.965 1.18C15.085.164 13.769 0 12.262 0h-.524C10.231 0 8.915.164 7.035 1.18 3.24 3.223 1.322 5.194.511 7.662c-.82 2.483-.582 4.557.21 6.97l.007.016c.662 1.891 2.318 4.337 3.952 6.139C6.897 23.294 9.343 24 11.999 24c2.656 0 5.102-.706 7.32-3.213 1.633-1.802 3.29-4.248 3.951-6.139l.007-.016c.792-2.413 1.03-4.487.21-6.97-.81-2.468-2.729-4.439-6.522-6.482zM12 19.5c-4.136 0-7.5-3.364-7.5-7.5S7.864 4.5 12 4.5s7.5 3.364 7.5 7.5-3.364 7.5-7.5 7.5zm0-13.125c-3.107 0-5.625 2.518-5.625 5.625S8.893 17.625 12 17.625 17.625 15.107 17.625 12 15.107 6.375 12 6.375z" />
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

export default PagerDutyNode;
