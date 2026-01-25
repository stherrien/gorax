/**
 * Node schema definitions for dynamic configuration form generation
 * Each schema defines the fields, validation, and UI for a node type
 */

import type { NodeSchema, FieldSchema } from '../types/workflow'

// ============================================================================
// Field Templates (reusable field definitions)
// ============================================================================

const LABEL_FIELD: FieldSchema = {
  name: 'label',
  label: 'Name',
  type: 'text',
  required: true,
  placeholder: 'Enter node name',
  description: 'A descriptive name for this node',
}

const DESCRIPTION_FIELD: FieldSchema = {
  name: 'description',
  label: 'Description',
  type: 'textarea',
  placeholder: 'Optional description',
  description: 'Additional details about this node',
}

const CREDENTIAL_FIELD: FieldSchema = {
  name: 'credentialId',
  label: 'Credential',
  type: 'credential',
  description: 'Select a stored credential',
}

// ============================================================================
// Trigger Schemas
// ============================================================================

const WEBHOOK_TRIGGER_SCHEMA: NodeSchema = {
  type: 'webhook',
  label: 'Webhook',
  description: 'Trigger workflow via HTTP request',
  icon: '🔗',
  category: 'trigger',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'path',
      label: 'Path',
      type: 'text',
      placeholder: '/my-webhook',
      description: 'URL path for the webhook endpoint',
    },
    {
      name: 'method',
      label: 'HTTP Method',
      type: 'select',
      defaultValue: 'POST',
      options: [
        { value: 'GET', label: 'GET' },
        { value: 'POST', label: 'POST' },
        { value: 'PUT', label: 'PUT' },
        { value: 'DELETE', label: 'DELETE' },
        { value: 'PATCH', label: 'PATCH' },
      ],
    },
    {
      name: 'authType',
      label: 'Authentication',
      type: 'select',
      defaultValue: 'none',
      options: [
        { value: 'none', label: 'None' },
        { value: 'basic', label: 'Basic Auth' },
        { value: 'signature', label: 'HMAC Signature' },
        { value: 'api_key', label: 'API Key' },
      ],
      description: 'How to authenticate incoming requests',
    },
    {
      name: 'secret',
      label: 'Secret',
      type: 'text',
      placeholder: 'Enter secret key',
      description: 'Secret for signature verification',
      dependsOn: {
        field: 'authType',
        value: 'signature',
      },
    },
    {
      name: 'priority',
      label: 'Priority',
      type: 'number',
      defaultValue: 1,
      description: 'Execution priority (1-10, higher = more urgent)',
      validation: {
        min: 1,
        max: 10,
      },
    },
  ],
  outputs: 1,
}

const SCHEDULE_TRIGGER_SCHEMA: NodeSchema = {
  type: 'schedule',
  label: 'Schedule',
  description: 'Trigger workflow on a schedule',
  icon: '⏰',
  category: 'trigger',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'cron',
      label: 'Cron Expression',
      type: 'text',
      required: true,
      placeholder: '0 * * * *',
      description: 'Standard cron expression (minute, hour, day, month, weekday)',
    },
    {
      name: 'timezone',
      label: 'Timezone',
      type: 'text',
      defaultValue: 'UTC',
      placeholder: 'America/New_York',
      description: 'IANA timezone for schedule evaluation',
    },
  ],
  outputs: 1,
}

const MANUAL_TRIGGER_SCHEMA: NodeSchema = {
  type: 'manual',
  label: 'Manual',
  description: 'Trigger workflow manually',
  icon: '👆',
  category: 'trigger',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'inputSchema',
      label: 'Input Schema',
      type: 'json',
      placeholder: '{"type": "object", "properties": {}}',
      description: 'JSON Schema for manual trigger input',
    },
  ],
  outputs: 1,
}

// ============================================================================
// Action Schemas
// ============================================================================

const HTTP_ACTION_SCHEMA: NodeSchema = {
  type: 'http',
  label: 'HTTP Request',
  description: 'Make HTTP requests to external APIs',
  icon: '🌐',
  category: 'action',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'url',
      label: 'URL',
      type: 'expression',
      required: true,
      placeholder: 'https://api.example.com/endpoint',
      description: 'Target URL (supports expressions)',
      validation: {
        pattern: '^(https?://|\\{\\{)',
      },
    },
    {
      name: 'method',
      label: 'Method',
      type: 'select',
      defaultValue: 'GET',
      options: [
        { value: 'GET', label: 'GET' },
        { value: 'POST', label: 'POST' },
        { value: 'PUT', label: 'PUT' },
        { value: 'DELETE', label: 'DELETE' },
        { value: 'PATCH', label: 'PATCH' },
      ],
    },
    {
      name: 'headers',
      label: 'Headers',
      type: 'json',
      placeholder: '{"Content-Type": "application/json"}',
      description: 'Request headers as JSON object',
    },
    {
      name: 'body',
      label: 'Body',
      type: 'expression',
      placeholder: '{"key": "{{steps.previous.output}}"}',
      description: 'Request body (supports expressions)',
      dependsOn: {
        field: 'method',
        value: 'GET',
      },
    },
    {
      name: 'timeout',
      label: 'Timeout (seconds)',
      type: 'number',
      defaultValue: 30,
      description: 'Request timeout in seconds',
      validation: {
        min: 1,
        max: 300,
      },
    },
    CREDENTIAL_FIELD,
  ],
  inputs: 1,
  outputs: 1,
}

const TRANSFORM_ACTION_SCHEMA: NodeSchema = {
  type: 'transform',
  label: 'Transform',
  description: 'Transform and manipulate data',
  icon: '🔄',
  category: 'action',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'expression',
      label: 'Expression',
      type: 'expression',
      required: true,
      placeholder: '{{steps.previous.output.data}}',
      description: 'Data transformation expression',
    },
    {
      name: 'mapping',
      label: 'Field Mapping',
      type: 'json',
      placeholder: '{"newField": "{{input.oldField}}"}',
      description: 'Map input fields to output fields',
    },
  ],
  inputs: 1,
  outputs: 1,
}

const SCRIPT_ACTION_SCHEMA: NodeSchema = {
  type: 'script',
  label: 'Run Script',
  description: 'Execute custom JavaScript code',
  icon: '📜',
  category: 'action',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'language',
      label: 'Language',
      type: 'select',
      defaultValue: 'javascript',
      options: [
        { value: 'javascript', label: 'JavaScript' },
      ],
    },
    {
      name: 'code',
      label: 'Code',
      type: 'textarea',
      required: true,
      placeholder: '// Access input via `input` variable\nreturn { result: input.data }',
      description: 'Script to execute',
    },
  ],
  inputs: 1,
  outputs: 1,
}

const EMAIL_ACTION_SCHEMA: NodeSchema = {
  type: 'email',
  label: 'Send Email',
  description: 'Send email notifications',
  icon: '📧',
  category: 'action',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'to',
      label: 'To',
      type: 'expression',
      required: true,
      placeholder: 'user@example.com',
      description: 'Recipient email address(es)',
    },
    {
      name: 'subject',
      label: 'Subject',
      type: 'expression',
      required: true,
      placeholder: 'Notification: {{trigger.data.event}}',
      description: 'Email subject line',
    },
    {
      name: 'bodyTemplate',
      label: 'Body',
      type: 'textarea',
      required: true,
      placeholder: 'Hello,\n\nThis is a notification about {{trigger.data.event}}.',
      description: 'Email body (supports expressions)',
    },
    CREDENTIAL_FIELD,
  ],
  inputs: 1,
  outputs: 1,
}

// ============================================================================
// Slack Action Schemas
// ============================================================================

const SLACK_SEND_MESSAGE_SCHEMA: NodeSchema = {
  type: 'slack_send_message',
  label: 'Slack: Send Message',
  description: 'Send a message to a Slack channel',
  icon: '💬',
  category: 'action',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'channel',
      label: 'Channel ID',
      type: 'expression',
      required: true,
      placeholder: 'C01234567',
      description: 'Slack channel ID',
    },
    {
      name: 'text',
      label: 'Message Text',
      type: 'expression',
      placeholder: 'Hello from workflow!',
      description: 'Plain text message (fallback for blocks)',
    },
    {
      name: 'blocks',
      label: 'Block Kit Blocks',
      type: 'json',
      placeholder: '[{"type": "section", "text": {"type": "mrkdwn", "text": "Hello!"}}]',
      description: 'Slack Block Kit JSON',
    },
    CREDENTIAL_FIELD,
  ],
  inputs: 1,
  outputs: 1,
}

const SLACK_SEND_DM_SCHEMA: NodeSchema = {
  type: 'slack_send_dm',
  label: 'Slack: Send DM',
  description: 'Send a direct message to a Slack user',
  icon: '✉️',
  category: 'action',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'user',
      label: 'User',
      type: 'expression',
      required: true,
      placeholder: 'U01234567 or user@example.com',
      description: 'User ID or email',
    },
    {
      name: 'text',
      label: 'Message Text',
      type: 'expression',
      placeholder: 'Hello!',
      description: 'Plain text message',
    },
    {
      name: 'blocks',
      label: 'Block Kit Blocks',
      type: 'json',
      placeholder: '[{"type": "section", "text": {"type": "mrkdwn", "text": "Hello!"}}]',
      description: 'Slack Block Kit JSON',
    },
    CREDENTIAL_FIELD,
  ],
  inputs: 1,
  outputs: 1,
}

const SLACK_UPDATE_MESSAGE_SCHEMA: NodeSchema = {
  type: 'slack_update_message',
  label: 'Slack: Update Message',
  description: 'Update an existing Slack message',
  icon: '✏️',
  category: 'action',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'channel',
      label: 'Channel ID',
      type: 'expression',
      required: true,
      placeholder: 'C01234567',
      description: 'Channel containing the message',
    },
    {
      name: 'ts',
      label: 'Message Timestamp',
      type: 'expression',
      required: true,
      placeholder: '{{steps.send_message.output.ts}}',
      description: 'Timestamp of message to update',
    },
    {
      name: 'text',
      label: 'New Text',
      type: 'expression',
      placeholder: 'Updated message',
      description: 'Updated message text',
    },
    {
      name: 'blocks',
      label: 'Block Kit Blocks',
      type: 'json',
      description: 'Updated Block Kit JSON',
    },
    CREDENTIAL_FIELD,
  ],
  inputs: 1,
  outputs: 1,
}

const SLACK_ADD_REACTION_SCHEMA: NodeSchema = {
  type: 'slack_add_reaction',
  label: 'Slack: Add Reaction',
  description: 'Add an emoji reaction to a message',
  icon: '👍',
  category: 'action',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'channel',
      label: 'Channel ID',
      type: 'expression',
      required: true,
      placeholder: 'C01234567',
      description: 'Channel containing the message',
    },
    {
      name: 'timestamp',
      label: 'Message Timestamp',
      type: 'expression',
      required: true,
      placeholder: '{{steps.send_message.output.ts}}',
      description: 'Timestamp of message',
    },
    {
      name: 'emoji',
      label: 'Emoji',
      type: 'text',
      required: true,
      placeholder: 'thumbsup',
      description: 'Emoji name (without colons)',
    },
    CREDENTIAL_FIELD,
  ],
  inputs: 1,
  outputs: 1,
}

// ============================================================================
// AI Action Schemas
// ============================================================================

const AI_CHAT_SCHEMA: NodeSchema = {
  type: 'ai_chat',
  label: 'AI: Chat Completion',
  description: 'Generate AI responses with LLM',
  icon: '🤖',
  category: 'ai',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'provider',
      label: 'Provider',
      type: 'select',
      required: true,
      defaultValue: 'openai',
      options: [
        { value: 'openai', label: 'OpenAI' },
        { value: 'anthropic', label: 'Anthropic' },
        { value: 'bedrock', label: 'AWS Bedrock' },
      ],
    },
    {
      name: 'model',
      label: 'Model',
      type: 'select',
      required: true,
      options: [
        { value: 'gpt-4', label: 'GPT-4' },
        { value: 'gpt-4-turbo', label: 'GPT-4 Turbo' },
        { value: 'gpt-3.5-turbo', label: 'GPT-3.5 Turbo' },
        { value: 'claude-3-opus', label: 'Claude 3 Opus' },
        { value: 'claude-3-sonnet', label: 'Claude 3 Sonnet' },
      ],
    },
    {
      name: 'systemPrompt',
      label: 'System Prompt',
      type: 'textarea',
      placeholder: 'You are a helpful assistant...',
      description: 'System message to set AI behavior',
    },
    {
      name: 'prompt',
      label: 'User Prompt',
      type: 'expression',
      required: true,
      placeholder: 'Analyze the following: {{trigger.data}}',
      description: 'The prompt to send to the AI',
    },
    {
      name: 'temperature',
      label: 'Temperature',
      type: 'number',
      defaultValue: 0.7,
      description: 'Randomness of responses (0-2)',
      validation: {
        min: 0,
        max: 2,
      },
    },
    {
      name: 'maxTokens',
      label: 'Max Tokens',
      type: 'number',
      defaultValue: 1024,
      description: 'Maximum tokens in response',
      validation: {
        min: 1,
        max: 128000,
      },
    },
    CREDENTIAL_FIELD,
  ],
  inputs: 1,
  outputs: 1,
}

const AI_SUMMARIZE_SCHEMA: NodeSchema = {
  type: 'ai_summarize',
  label: 'AI: Summarize',
  description: 'Summarize text using AI',
  icon: '📝',
  category: 'ai',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'provider',
      label: 'Provider',
      type: 'select',
      required: true,
      defaultValue: 'openai',
      options: [
        { value: 'openai', label: 'OpenAI' },
        { value: 'anthropic', label: 'Anthropic' },
      ],
    },
    {
      name: 'text',
      label: 'Text to Summarize',
      type: 'expression',
      required: true,
      placeholder: '{{trigger.data.content}}',
      description: 'The text to summarize',
    },
    {
      name: 'maxLength',
      label: 'Max Summary Length',
      type: 'number',
      defaultValue: 200,
      description: 'Target summary length in words',
    },
    {
      name: 'style',
      label: 'Summary Style',
      type: 'select',
      defaultValue: 'concise',
      options: [
        { value: 'concise', label: 'Concise' },
        { value: 'detailed', label: 'Detailed' },
        { value: 'bullet_points', label: 'Bullet Points' },
      ],
    },
    CREDENTIAL_FIELD,
  ],
  inputs: 1,
  outputs: 1,
}

const AI_CLASSIFY_SCHEMA: NodeSchema = {
  type: 'ai_classify',
  label: 'AI: Classify',
  description: 'Classify text into categories',
  icon: '🏷️',
  category: 'ai',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'provider',
      label: 'Provider',
      type: 'select',
      required: true,
      defaultValue: 'openai',
      options: [
        { value: 'openai', label: 'OpenAI' },
        { value: 'anthropic', label: 'Anthropic' },
      ],
    },
    {
      name: 'text',
      label: 'Text to Classify',
      type: 'expression',
      required: true,
      placeholder: '{{trigger.data.message}}',
      description: 'The text to classify',
    },
    {
      name: 'categories',
      label: 'Categories',
      type: 'multiselect',
      required: true,
      options: [
        { value: 'bug', label: 'Bug Report' },
        { value: 'feature', label: 'Feature Request' },
        { value: 'question', label: 'Question' },
        { value: 'feedback', label: 'Feedback' },
        { value: 'spam', label: 'Spam' },
      ],
      description: 'Categories to classify into',
    },
    {
      name: 'multiLabel',
      label: 'Allow Multiple Labels',
      type: 'boolean',
      defaultValue: false,
      description: 'Allow assigning multiple categories',
    },
    CREDENTIAL_FIELD,
  ],
  inputs: 1,
  outputs: 1,
}

const AI_EXTRACT_SCHEMA: NodeSchema = {
  type: 'ai_extract',
  label: 'AI: Extract Entities',
  description: 'Extract named entities from text',
  icon: '🔍',
  category: 'ai',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'provider',
      label: 'Provider',
      type: 'select',
      required: true,
      defaultValue: 'openai',
      options: [
        { value: 'openai', label: 'OpenAI' },
        { value: 'anthropic', label: 'Anthropic' },
      ],
    },
    {
      name: 'text',
      label: 'Text to Analyze',
      type: 'expression',
      required: true,
      placeholder: '{{trigger.data.content}}',
      description: 'The text to extract entities from',
    },
    {
      name: 'entityTypes',
      label: 'Entity Types',
      type: 'multiselect',
      options: [
        { value: 'person', label: 'Person' },
        { value: 'organization', label: 'Organization' },
        { value: 'location', label: 'Location' },
        { value: 'date', label: 'Date' },
        { value: 'email', label: 'Email' },
        { value: 'phone', label: 'Phone Number' },
        { value: 'url', label: 'URL' },
      ],
      description: 'Types of entities to extract',
    },
    CREDENTIAL_FIELD,
  ],
  inputs: 1,
  outputs: 1,
}

const AI_EMBED_SCHEMA: NodeSchema = {
  type: 'ai_embed',
  label: 'AI: Generate Embeddings',
  description: 'Create vector embeddings for text',
  icon: '📊',
  category: 'ai',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'provider',
      label: 'Provider',
      type: 'select',
      required: true,
      defaultValue: 'openai',
      options: [
        { value: 'openai', label: 'OpenAI' },
      ],
    },
    {
      name: 'model',
      label: 'Model',
      type: 'select',
      required: true,
      defaultValue: 'text-embedding-3-small',
      options: [
        { value: 'text-embedding-3-small', label: 'text-embedding-3-small' },
        { value: 'text-embedding-3-large', label: 'text-embedding-3-large' },
        { value: 'text-embedding-ada-002', label: 'text-embedding-ada-002' },
      ],
    },
    {
      name: 'text',
      label: 'Text to Embed',
      type: 'expression',
      required: true,
      placeholder: '{{trigger.data.content}}',
      description: 'The text to generate embeddings for',
    },
    CREDENTIAL_FIELD,
  ],
  inputs: 1,
  outputs: 1,
}

// ============================================================================
// Control Flow Schemas
// ============================================================================

const CONDITIONAL_SCHEMA: NodeSchema = {
  type: 'conditional',
  label: 'Conditional',
  description: 'Branch based on conditions',
  icon: '🔀',
  category: 'control',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'condition',
      label: 'Condition',
      type: 'expression',
      required: true,
      placeholder: '{{steps.previous.output.status}} === "success"',
      description: 'Expression that evaluates to true or false',
    },
    {
      name: 'trueLabel',
      label: 'True Branch Label',
      type: 'text',
      defaultValue: 'Yes',
      description: 'Label for the true branch',
    },
    {
      name: 'falseLabel',
      label: 'False Branch Label',
      type: 'text',
      defaultValue: 'No',
      description: 'Label for the false branch',
    },
  ],
  inputs: 1,
  outputs: 2,
  outputLabels: ['True', 'False'],
}

const LOOP_SCHEMA: NodeSchema = {
  type: 'loop',
  label: 'Loop',
  description: 'Iterate over arrays',
  icon: '🔁',
  category: 'control',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'source',
      label: 'Source Array',
      type: 'expression',
      required: true,
      placeholder: '{{steps.previous.output.items}}',
      description: 'Array to iterate over',
    },
    {
      name: 'itemVariable',
      label: 'Item Variable',
      type: 'text',
      required: true,
      defaultValue: 'item',
      placeholder: 'item',
      description: 'Variable name for current item',
    },
    {
      name: 'indexVariable',
      label: 'Index Variable',
      type: 'text',
      defaultValue: 'index',
      placeholder: 'index',
      description: 'Variable name for current index',
    },
    {
      name: 'maxIterations',
      label: 'Max Iterations',
      type: 'number',
      defaultValue: 1000,
      description: 'Safety limit for iterations',
      validation: {
        min: 1,
        max: 10000,
      },
    },
    {
      name: 'onError',
      label: 'On Error',
      type: 'select',
      defaultValue: 'stop',
      options: [
        { value: 'stop', label: 'Stop Loop' },
        { value: 'continue', label: 'Continue to Next' },
        { value: 'skip', label: 'Skip Current Item' },
      ],
      description: 'Behavior when an iteration fails',
    },
  ],
  inputs: 1,
  outputs: 2,
  outputLabels: ['Loop Body', 'After Loop'],
}

const PARALLEL_SCHEMA: NodeSchema = {
  type: 'parallel',
  label: 'Parallel',
  description: 'Execute branches concurrently',
  icon: '⚡',
  category: 'control',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'errorStrategy',
      label: 'Error Strategy',
      type: 'select',
      defaultValue: 'fail_fast',
      options: [
        { value: 'fail_fast', label: 'Fail Fast', description: 'Stop all on first error' },
        { value: 'continue_on_error', label: 'Continue on Error', description: 'Continue other branches' },
        { value: 'wait_all', label: 'Wait All', description: 'Wait for all to complete' },
      ],
      description: 'How to handle errors in parallel branches',
    },
    {
      name: 'maxConcurrency',
      label: 'Max Concurrency',
      type: 'number',
      defaultValue: 0,
      description: '0 = unlimited, otherwise max parallel executions',
      validation: {
        min: 0,
        max: 100,
      },
    },
  ],
  inputs: 1,
  outputs: 3,
  outputLabels: ['Branch 1', 'Branch 2', 'Branch 3'],
}

const DELAY_SCHEMA: NodeSchema = {
  type: 'delay',
  label: 'Delay',
  description: 'Wait for a specified time',
  icon: '⏸️',
  category: 'control',
  fields: [
    LABEL_FIELD,
    DESCRIPTION_FIELD,
    {
      name: 'duration',
      label: 'Duration (seconds)',
      type: 'number',
      required: true,
      defaultValue: 5,
      description: 'Time to wait in seconds',
      validation: {
        min: 1,
        max: 86400,
      },
    },
  ],
  inputs: 1,
  outputs: 1,
}

// ============================================================================
// Schema Registry
// ============================================================================

export const NODE_SCHEMAS: Record<string, NodeSchema> = {
  // Triggers
  webhook: WEBHOOK_TRIGGER_SCHEMA,
  schedule: SCHEDULE_TRIGGER_SCHEMA,
  manual: MANUAL_TRIGGER_SCHEMA,

  // Actions
  http: HTTP_ACTION_SCHEMA,
  transform: TRANSFORM_ACTION_SCHEMA,
  script: SCRIPT_ACTION_SCHEMA,
  email: EMAIL_ACTION_SCHEMA,
  slack_send_message: SLACK_SEND_MESSAGE_SCHEMA,
  slack_send_dm: SLACK_SEND_DM_SCHEMA,
  slack_update_message: SLACK_UPDATE_MESSAGE_SCHEMA,
  slack_add_reaction: SLACK_ADD_REACTION_SCHEMA,

  // AI
  ai_chat: AI_CHAT_SCHEMA,
  ai_summarize: AI_SUMMARIZE_SCHEMA,
  ai_classify: AI_CLASSIFY_SCHEMA,
  ai_extract: AI_EXTRACT_SCHEMA,
  ai_embed: AI_EMBED_SCHEMA,

  // Control
  conditional: CONDITIONAL_SCHEMA,
  loop: LOOP_SCHEMA,
  parallel: PARALLEL_SCHEMA,
  delay: DELAY_SCHEMA,
}

/**
 * Get schema for a node type
 */
export function getNodeSchema(nodeType: string): NodeSchema | undefined {
  return NODE_SCHEMAS[nodeType]
}

/**
 * Get all schemas by category
 */
export function getSchemasByCategory(category: NodeSchema['category']): NodeSchema[] {
  return Object.values(NODE_SCHEMAS).filter((schema) => schema.category === category)
}

/**
 * Get all available node types
 */
export function getAllNodeTypes(): string[] {
  return Object.keys(NODE_SCHEMAS)
}
