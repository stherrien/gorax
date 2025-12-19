import React from 'react'

interface SlackConfigPanelProps {
  nodeType: string
  formData: any
  onChange: (field: string, value: any) => void
  errors: Record<string, string>
}

export default function SlackConfigPanel({
  nodeType,
  formData,
  onChange,
  errors,
}: SlackConfigPanelProps) {
  return (
    <div className="space-y-4">
      {/* Send Message Configuration */}
      {nodeType === 'slack_send_message' && (
        <SlackSendMessageFields
          formData={formData}
          onChange={onChange}
          errors={errors}
        />
      )}

      {/* Send DM Configuration */}
      {nodeType === 'slack_send_dm' && (
        <SlackSendDMFields formData={formData} onChange={onChange} errors={errors} />
      )}

      {/* Update Message Configuration */}
      {nodeType === 'slack_update_message' && (
        <SlackUpdateMessageFields
          formData={formData}
          onChange={onChange}
          errors={errors}
        />
      )}

      {/* Add Reaction Configuration */}
      {nodeType === 'slack_add_reaction' && (
        <SlackAddReactionFields
          formData={formData}
          onChange={onChange}
          errors={errors}
        />
      )}
    </div>
  )
}

// Send Message Fields Component
function SlackSendMessageFields({
  formData,
  onChange,
  errors,
}: {
  formData: any
  onChange: (field: string, value: any) => void
  errors: Record<string, string>
}) {
  return (
    <>
      {/* Channel ID */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Channel ID <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          value={formData.channel || ''}
          onChange={(e) => onChange('channel', e.target.value)}
          placeholder="C1234567890 or {{trigger.channel}}"
          className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500 ${
            errors.channel ? 'border-red-500' : 'border-gray-300'
          }`}
        />
        {errors.channel && (
          <p className="text-red-500 text-xs mt-1">{errors.channel}</p>
        )}
        <p className="text-xs text-gray-500 mt-1">
          Slack channel ID (e.g., C1234567890). Supports template variables.
        </p>
      </div>

      {/* Message Text */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Message Text
        </label>
        <textarea
          value={formData.text || ''}
          onChange={(e) => onChange('text', e.target.value)}
          placeholder="Hello! Your workflow has completed."
          rows={4}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
        />
        <p className="text-xs text-gray-500 mt-1">
          Plain text or Markdown message. Leave empty if using blocks.
        </p>
      </div>

      {/* Blocks (JSON) */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Blocks (JSON)
        </label>
        <textarea
          value={formData.blocks || ''}
          onChange={(e) => onChange('blocks', e.target.value)}
          placeholder='[{"type": "section", "text": {"type": "mrkdwn", "text": "*Hello!*"}}]'
          rows={6}
          className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500 font-mono text-sm ${
            errors.blocks ? 'border-red-500' : 'border-gray-300'
          }`}
        />
        {errors.blocks && (
          <p className="text-red-500 text-xs mt-1">{errors.blocks}</p>
        )}
        <p className="text-xs text-gray-500 mt-1">
          Slack Block Kit JSON array. See{' '}
          <a
            href="https://api.slack.com/block-kit"
            target="_blank"
            rel="noopener noreferrer"
            className="text-blue-500 hover:underline"
          >
            Block Kit documentation
          </a>
          .
        </p>
      </div>

      {/* Thread TS (Optional) */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Thread Timestamp (Optional)
        </label>
        <input
          type="text"
          value={formData.thread_ts || ''}
          onChange={(e) => onChange('thread_ts', e.target.value)}
          placeholder="1503435956.000247 or {{steps.previous.timestamp}}"
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
        />
        <p className="text-xs text-gray-500 mt-1">
          Reply to a thread by providing the parent message timestamp.
        </p>
      </div>

      {/* Username (Optional) */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Bot Username (Optional)
        </label>
        <input
          type="text"
          value={formData.username || ''}
          onChange={(e) => onChange('username', e.target.value)}
          placeholder="Workflow Bot"
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
        />
        <p className="text-xs text-gray-500 mt-1">
          Custom username for the bot message.
        </p>
      </div>

      {/* Icon Emoji (Optional) */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Icon Emoji (Optional)
        </label>
        <input
          type="text"
          value={formData.icon_emoji || ''}
          onChange={(e) => onChange('icon_emoji', e.target.value)}
          placeholder=":robot_face:"
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
        />
        <p className="text-xs text-gray-500 mt-1">
          Emoji code for the bot icon (e.g., :robot_face:).
        </p>
      </div>
    </>
  )
}

// Send DM Fields Component
function SlackSendDMFields({
  formData,
  onChange,
  errors,
}: {
  formData: any
  onChange: (field: string, value: any) => void
  errors: Record<string, string>
}) {
  return (
    <>
      {/* User Email or ID */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          User Email or ID <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          value={formData.user || ''}
          onChange={(e) => onChange('user', e.target.value)}
          placeholder="user@example.com or U1234567890"
          className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500 ${
            errors.user ? 'border-red-500' : 'border-gray-300'
          }`}
        />
        {errors.user && <p className="text-red-500 text-xs mt-1">{errors.user}</p>}
        <p className="text-xs text-gray-500 mt-1">
          User email address or Slack user ID (U1234567890).
        </p>
      </div>

      {/* Message Text */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Message Text
        </label>
        <textarea
          value={formData.text || ''}
          onChange={(e) => onChange('text', e.target.value)}
          placeholder="Hello! This is a direct message from the workflow."
          rows={4}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
        />
        <p className="text-xs text-gray-500 mt-1">
          Plain text or Markdown message. Leave empty if using blocks.
        </p>
      </div>

      {/* Blocks (JSON) */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Blocks (JSON)
        </label>
        <textarea
          value={formData.blocks || ''}
          onChange={(e) => onChange('blocks', e.target.value)}
          placeholder='[{"type": "section", "text": {"type": "mrkdwn", "text": "*Important!*"}}]'
          rows={6}
          className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500 font-mono text-sm ${
            errors.blocks ? 'border-red-500' : 'border-gray-300'
          }`}
        />
        {errors.blocks && (
          <p className="text-red-500 text-xs mt-1">{errors.blocks}</p>
        )}
        <p className="text-xs text-gray-500 mt-1">
          Slack Block Kit JSON array for rich formatting.
        </p>
      </div>
    </>
  )
}

// Update Message Fields Component
function SlackUpdateMessageFields({
  formData,
  onChange,
  errors,
}: {
  formData: any
  onChange: (field: string, value: any) => void
  errors: Record<string, string>
}) {
  return (
    <>
      {/* Channel ID */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Channel ID <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          value={formData.channel || ''}
          onChange={(e) => onChange('channel', e.target.value)}
          placeholder="C1234567890 or {{steps.send.channel}}"
          className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500 ${
            errors.channel ? 'border-red-500' : 'border-gray-300'
          }`}
        />
        {errors.channel && (
          <p className="text-red-500 text-xs mt-1">{errors.channel}</p>
        )}
        <p className="text-xs text-gray-500 mt-1">
          Channel ID where the message was sent.
        </p>
      </div>

      {/* Message Timestamp */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Message Timestamp <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          value={formData.ts || ''}
          onChange={(e) => onChange('ts', e.target.value)}
          placeholder="1503435956.000247 or {{steps.send.timestamp}}"
          className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500 ${
            errors.ts ? 'border-red-500' : 'border-gray-300'
          }`}
        />
        {errors.ts && <p className="text-red-500 text-xs mt-1">{errors.ts}</p>}
        <p className="text-xs text-gray-500 mt-1">
          Timestamp of the message to update. Usually from a previous send action.
        </p>
      </div>

      {/* New Message Text */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Updated Message Text
        </label>
        <textarea
          value={formData.text || ''}
          onChange={(e) => onChange('text', e.target.value)}
          placeholder="✅ Workflow completed successfully!"
          rows={4}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
        />
        <p className="text-xs text-gray-500 mt-1">
          New text to replace the original message. Leave empty if using blocks.
        </p>
      </div>

      {/* Blocks (JSON) */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Updated Blocks (JSON)
        </label>
        <textarea
          value={formData.blocks || ''}
          onChange={(e) => onChange('blocks', e.target.value)}
          placeholder='[{"type": "section", "text": {"type": "mrkdwn", "text": "✅ *Done!*"}}]'
          rows={6}
          className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500 font-mono text-sm ${
            errors.blocks ? 'border-red-500' : 'border-gray-300'
          }`}
        />
        {errors.blocks && (
          <p className="text-red-500 text-xs mt-1">{errors.blocks}</p>
        )}
        <p className="text-xs text-gray-500 mt-1">
          New Block Kit JSON to replace the original blocks.
        </p>
      </div>
    </>
  )
}

// Add Reaction Fields Component
function SlackAddReactionFields({
  formData,
  onChange,
  errors,
}: {
  formData: any
  onChange: (field: string, value: any) => void
  errors: Record<string, string>
}) {
  // Common emoji options
  const commonEmojis = [
    { name: 'thumbsup', emoji: '👍' },
    { name: 'thumbsdown', emoji: '👎' },
    { name: 'white_check_mark', emoji: '✅' },
    { name: 'x', emoji: '❌' },
    { name: 'warning', emoji: '⚠️' },
    { name: 'eyes', emoji: '👀' },
    { name: 'rocket', emoji: '🚀' },
    { name: 'tada', emoji: '🎉' },
    { name: 'fire', emoji: '🔥' },
    { name: 'heart', emoji: '❤️' },
  ]

  return (
    <>
      {/* Channel ID */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Channel ID <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          value={formData.channel || ''}
          onChange={(e) => onChange('channel', e.target.value)}
          placeholder="C1234567890 or {{steps.send.channel}}"
          className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500 ${
            errors.channel ? 'border-red-500' : 'border-gray-300'
          }`}
        />
        {errors.channel && (
          <p className="text-red-500 text-xs mt-1">{errors.channel}</p>
        )}
        <p className="text-xs text-gray-500 mt-1">
          Channel ID where the message was sent.
        </p>
      </div>

      {/* Message Timestamp */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Message Timestamp <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          value={formData.timestamp || ''}
          onChange={(e) => onChange('timestamp', e.target.value)}
          placeholder="1503435956.000247 or {{steps.send.timestamp}}"
          className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500 ${
            errors.timestamp ? 'border-red-500' : 'border-gray-300'
          }`}
        />
        {errors.timestamp && (
          <p className="text-red-500 text-xs mt-1">{errors.timestamp}</p>
        )}
        <p className="text-xs text-gray-500 mt-1">
          Timestamp of the message to react to.
        </p>
      </div>

      {/* Emoji Name */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Emoji Name <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          value={formData.emoji || ''}
          onChange={(e) => onChange('emoji', e.target.value)}
          placeholder="thumbsup or :thumbsup:"
          className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500 ${
            errors.emoji ? 'border-red-500' : 'border-gray-300'
          }`}
        />
        {errors.emoji && <p className="text-red-500 text-xs mt-1">{errors.emoji}</p>}
        <p className="text-xs text-gray-500 mt-1">
          Emoji name without colons (e.g., thumbsup, tada, rocket).
        </p>
      </div>

      {/* Common Emoji Picker */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Quick Pick
        </label>
        <div className="grid grid-cols-5 gap-2">
          {commonEmojis.map((item) => (
            <button
              key={item.name}
              type="button"
              onClick={() => onChange('emoji', item.name)}
              className="px-3 py-2 border border-gray-300 rounded-md hover:bg-gray-50 text-2xl"
              title={item.name}
            >
              {item.emoji}
            </button>
          ))}
        </div>
        <p className="text-xs text-gray-500 mt-2">
          Click an emoji to use it, or type a custom emoji name above.
        </p>
      </div>
    </>
  )
}
