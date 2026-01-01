import { useState } from 'react'
import { useSearchParams } from 'react-router-dom'

type TabId = 'getting-started' | 'features' | 'api' | 'quick-ref'

export default function Documentation() {
  const [searchParams, setSearchParams] = useSearchParams()
  const activeTab = (searchParams.get('tab') as TabId) || 'getting-started'
  const [searchQuery, setSearchQuery] = useState('')

  const setActiveTab = (tab: TabId) => {
    setSearchParams({ tab })
  }

  const tabs = [
    { id: 'getting-started' as TabId, label: 'Getting Started' },
    { id: 'features' as TabId, label: 'Features' },
    { id: 'api' as TabId, label: 'API Reference' },
    { id: 'quick-ref' as TabId, label: 'Quick Reference' },
  ]

  return (
    <div className="max-w-6xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-white">Documentation</h1>
        <div className="relative">
          <input
            type="text"
            placeholder="Search docs..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-64 px-4 py-2 bg-gray-800 text-white border border-gray-700 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
        </div>
      </div>

      {/* Tabs */}
      <div className="flex space-x-1 mb-6 bg-gray-800 p-1 rounded-lg">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`flex-1 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              activeTab === tab.id
                ? 'bg-primary-600 text-white'
                : 'text-gray-400 hover:text-white hover:bg-gray-700'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Content */}
      <div className="bg-gray-800 rounded-lg p-6">
        {activeTab === 'getting-started' && <GettingStartedContent />}
        {activeTab === 'features' && <FeaturesContent />}
        {activeTab === 'api' && <ApiReferenceContent />}
        {activeTab === 'quick-ref' && <QuickReferenceContent />}
      </div>
    </div>
  )
}

function GettingStartedContent() {
  return (
    <div className="space-y-8">
      <Section title="Welcome to Gorax">
        <p className="text-gray-300 mb-4">
          Gorax is a powerful workflow automation platform that lets you create, manage, and monitor
          automated workflows with ease. Whether you're integrating APIs, processing data, or
          automating business processes, Gorax provides the tools you need.
        </p>
        <h4 className="text-white font-medium mb-2">Key Features</h4>
        <ul className="list-disc list-inside text-gray-300 space-y-1">
          <li>Visual workflow editor with drag-and-drop canvas</li>
          <li>Multiple trigger types: webhooks, schedules, and manual</li>
          <li>Built-in actions for HTTP requests, data transformation, and conditions</li>
          <li>Secure credential management with encryption</li>
          <li>Real-time execution monitoring and logging</li>
          <li>AI-powered workflow builder</li>
          <li>Marketplace for sharing and discovering templates</li>
        </ul>
      </Section>

      <Section title="Quick Start Tutorial">
        <div className="space-y-6">
          <Step number={1} title="Create a New Workflow">
            <p className="text-gray-300 mb-2">
              Click "New Workflow" in the header or go to Workflows → New Workflow.
              Give your workflow a name and description.
            </p>
          </Step>

          <Step number={2} title="Add a Trigger">
            <p className="text-gray-300 mb-2">
              Every workflow starts with a trigger. Drag a trigger node from the palette:
            </p>
            <ul className="list-disc list-inside text-gray-300 ml-4 space-y-1">
              <li><strong>Webhook Trigger</strong> - Triggered by HTTP requests</li>
              <li><strong>Schedule Trigger</strong> - Triggered by cron schedule</li>
              <li><strong>Manual Trigger</strong> - Triggered manually via UI or API</li>
            </ul>
          </Step>

          <Step number={3} title="Add Actions">
            <p className="text-gray-300 mb-2">
              Connect actions to your trigger to define what happens:
            </p>
            <ul className="list-disc list-inside text-gray-300 ml-4 space-y-1">
              <li><strong>HTTP Request</strong> - Make API calls to external services</li>
              <li><strong>Transform</strong> - Manipulate and transform data</li>
              <li><strong>Condition</strong> - Branch based on conditions</li>
              <li><strong>Loop</strong> - Iterate over arrays</li>
            </ul>
          </Step>

          <Step number={4} title="Configure and Test">
            <p className="text-gray-300 mb-2">
              Click on any node to configure it in the property panel. Use "Dry Run" to test
              your workflow without affecting external systems.
            </p>
          </Step>

          <Step number={5} title="Activate and Monitor">
            <p className="text-gray-300">
              Save your workflow and set it to "Active". Monitor executions in the Executions page
              to see real-time status and logs.
            </p>
          </Step>
        </div>
      </Section>

      <Section title="Core Concepts">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <ConceptCard
            title="Workflows"
            description="A workflow is a series of connected nodes that define an automated process. Workflows have a trigger and one or more actions."
          />
          <ConceptCard
            title="Nodes"
            description="Nodes are the building blocks of workflows. Each node represents a step: triggers, actions, conditions, or loops."
          />
          <ConceptCard
            title="Edges"
            description="Edges connect nodes together, defining the flow of execution. Data flows from source to target nodes."
          />
          <ConceptCard
            title="Executions"
            description="An execution is a single run of a workflow. It tracks status, timing, and outputs for each step."
          />
          <ConceptCard
            title="Credentials"
            description="Credentials securely store API keys, tokens, and passwords for use in workflows."
          />
          <ConceptCard
            title="Templates"
            description="Templates are pre-built workflows you can install from the marketplace or save from your own workflows."
          />
        </div>
      </Section>
    </div>
  )
}

function FeaturesContent() {
  const [activeFeature, setActiveFeature] = useState('workflows')

  const features = [
    { id: 'workflows', name: 'Workflows' },
    { id: 'webhooks', name: 'Webhooks' },
    { id: 'schedules', name: 'Schedules' },
    { id: 'executions', name: 'Executions' },
    { id: 'credentials', name: 'Credentials' },
    { id: 'ai-builder', name: 'AI Builder' },
    { id: 'marketplace', name: 'Marketplace' },
    { id: 'analytics', name: 'Analytics' },
  ]

  return (
    <div className="flex gap-6">
      {/* Feature Sidebar */}
      <div className="w-48 flex-shrink-0">
        <nav className="space-y-1">
          {features.map((feature) => (
            <button
              key={feature.id}
              onClick={() => setActiveFeature(feature.id)}
              className={`w-full text-left px-3 py-2 rounded-md text-sm transition-colors ${
                activeFeature === feature.id
                  ? 'bg-gray-700 text-white'
                  : 'text-gray-400 hover:text-white hover:bg-gray-700/50'
              }`}
            >
              {feature.name}
            </button>
          ))}
        </nav>
      </div>

      {/* Feature Content */}
      <div className="flex-1 min-w-0">
        {activeFeature === 'workflows' && <WorkflowsDocs />}
        {activeFeature === 'webhooks' && <WebhooksDocs />}
        {activeFeature === 'schedules' && <SchedulesDocs />}
        {activeFeature === 'executions' && <ExecutionsDocs />}
        {activeFeature === 'credentials' && <CredentialsDocs />}
        {activeFeature === 'ai-builder' && <AIBuilderDocs />}
        {activeFeature === 'marketplace' && <MarketplaceDocs />}
        {activeFeature === 'analytics' && <AnalyticsDocs />}
      </div>
    </div>
  )
}

function WorkflowsDocs() {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold text-white">Workflows</h3>
      <p className="text-gray-300">
        Workflows are the core of Gorax. Each workflow defines an automated process with triggers and actions.
      </p>

      <h4 className="text-lg font-medium text-white mt-6">Creating Workflows</h4>
      <p className="text-gray-300 mb-2">
        Navigate to Workflows → New Workflow to create a new workflow. The visual editor lets you:
      </p>
      <ul className="list-disc list-inside text-gray-300 space-y-1">
        <li>Drag nodes from the palette onto the canvas</li>
        <li>Connect nodes by dragging from output to input handles</li>
        <li>Configure nodes using the property panel on the right</li>
        <li>Save and version your workflows</li>
      </ul>

      <h4 className="text-lg font-medium text-white mt-6">Node Types</h4>
      <div className="space-y-3">
        <NodeTypeCard name="Trigger Nodes" description="Start the workflow: Webhook, Schedule, or Manual triggers" color="green" />
        <NodeTypeCard name="Action Nodes" description="Perform operations: HTTP Request, Transform, Send Email, etc." color="blue" />
        <NodeTypeCard name="Condition Nodes" description="Branch based on conditions with if/else logic" color="yellow" />
        <NodeTypeCard name="Loop Nodes" description="Iterate over arrays and process items" color="purple" />
        <NodeTypeCard name="Parallel Nodes" description="Execute multiple branches simultaneously" color="cyan" />
      </div>

      <h4 className="text-lg font-medium text-white mt-6">Version History</h4>
      <p className="text-gray-300">
        Every save creates a new version. Access version history from the workflow editor to view,
        compare, or restore previous versions.
      </p>
    </div>
  )
}

function WebhooksDocs() {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold text-white">Webhooks</h3>
      <p className="text-gray-300">
        Webhooks allow external services to trigger your workflows via HTTP requests.
      </p>

      <h4 className="text-lg font-medium text-white mt-6">Creating Webhooks</h4>
      <p className="text-gray-300 mb-2">
        Go to Webhooks → Create Webhook. Configure:
      </p>
      <ul className="list-disc list-inside text-gray-300 space-y-1">
        <li><strong>Name</strong> - A descriptive name for the webhook</li>
        <li><strong>Workflow</strong> - The workflow to trigger</li>
        <li><strong>Authentication</strong> - None, Signature, Basic Auth, or API Key</li>
        <li><strong>Filters</strong> - Conditions to filter incoming requests</li>
      </ul>

      <h4 className="text-lg font-medium text-white mt-6">Authentication Types</h4>
      <div className="space-y-3">
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">None</h5>
          <p className="text-gray-400 text-sm">No authentication required (use with caution)</p>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">Signature</h5>
          <p className="text-gray-400 text-sm">HMAC-SHA256 signature verification in X-Signature header</p>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">Basic Auth</h5>
          <p className="text-gray-400 text-sm">Username/password in Authorization header</p>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">API Key</h5>
          <p className="text-gray-400 text-sm">API key in X-API-Key header or query parameter</p>
        </div>
      </div>

      <h4 className="text-lg font-medium text-white mt-6">Event History</h4>
      <p className="text-gray-300">
        View all incoming webhook requests in the event history. You can inspect payloads,
        headers, and replay events to re-trigger workflows.
      </p>
    </div>
  )
}

function SchedulesDocs() {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold text-white">Schedules</h3>
      <p className="text-gray-300">
        Schedules let you run workflows automatically at specified times using cron expressions.
      </p>

      <h4 className="text-lg font-medium text-white mt-6">Creating Schedules</h4>
      <p className="text-gray-300 mb-2">
        Go to Schedules → Create Schedule. Configure:
      </p>
      <ul className="list-disc list-inside text-gray-300 space-y-1">
        <li><strong>Workflow</strong> - The workflow to execute</li>
        <li><strong>Cron Expression</strong> - When to run (e.g., "0 9 * * *" for 9 AM daily)</li>
        <li><strong>Timezone</strong> - The timezone for the schedule</li>
        <li><strong>Enabled</strong> - Whether the schedule is active</li>
      </ul>

      <h4 className="text-lg font-medium text-white mt-6">Cron Expression Format</h4>
      <CodeBlock code={`┌───────────── minute (0 - 59)
│ ┌─────────── hour (0 - 23)
│ │ ┌───────── day of month (1 - 31)
│ │ │ ┌─────── month (1 - 12)
│ │ │ │ ┌───── day of week (0 - 6) (Sunday = 0)
│ │ │ │ │
* * * * *`} />

      <h4 className="text-lg font-medium text-white mt-6">Common Examples</h4>
      <div className="space-y-2">
        <CronExample expression="* * * * *" description="Every minute" />
        <CronExample expression="0 * * * *" description="Every hour" />
        <CronExample expression="0 9 * * *" description="Every day at 9 AM" />
        <CronExample expression="0 9 * * 1-5" description="Weekdays at 9 AM" />
        <CronExample expression="0 0 1 * *" description="First day of month at midnight" />
        <CronExample expression="*/15 * * * *" description="Every 15 minutes" />
      </div>
    </div>
  )
}

function ExecutionsDocs() {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold text-white">Executions</h3>
      <p className="text-gray-300">
        Executions show the history and status of all workflow runs.
      </p>

      <h4 className="text-lg font-medium text-white mt-6">Execution Statuses</h4>
      <div className="space-y-2">
        <StatusBadge status="queued" description="Waiting to be processed" />
        <StatusBadge status="running" description="Currently executing" />
        <StatusBadge status="completed" description="Finished successfully" />
        <StatusBadge status="failed" description="Encountered an error" />
        <StatusBadge status="cancelled" description="Manually cancelled" />
        <StatusBadge status="timeout" description="Exceeded time limit" />
      </div>

      <h4 className="text-lg font-medium text-white mt-6">Viewing Execution Details</h4>
      <p className="text-gray-300 mb-2">
        Click on any execution to see:
      </p>
      <ul className="list-disc list-inside text-gray-300 space-y-1">
        <li>Step-by-step execution trace</li>
        <li>Input and output data for each node</li>
        <li>Timing information</li>
        <li>Error messages and stack traces</li>
      </ul>

      <h4 className="text-lg font-medium text-white mt-6">Actions</h4>
      <ul className="list-disc list-inside text-gray-300 space-y-1">
        <li><strong>Cancel</strong> - Stop a running execution</li>
        <li><strong>Retry</strong> - Re-run a failed execution</li>
        <li><strong>Export Logs</strong> - Download execution logs</li>
      </ul>
    </div>
  )
}

function CredentialsDocs() {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold text-white">Credentials</h3>
      <p className="text-gray-300">
        Credentials securely store sensitive data like API keys, passwords, and tokens.
      </p>

      <h4 className="text-lg font-medium text-white mt-6">Credential Types</h4>
      <div className="space-y-3">
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">API Key</h5>
          <p className="text-gray-400 text-sm">Store API keys for external services</p>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">OAuth2</h5>
          <p className="text-gray-400 text-sm">OAuth2 client credentials with automatic token refresh</p>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">Basic Auth</h5>
          <p className="text-gray-400 text-sm">Username and password pairs</p>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">Bearer Token</h5>
          <p className="text-gray-400 text-sm">JWT or other bearer tokens</p>
        </div>
      </div>

      <h4 className="text-lg font-medium text-white mt-6">Security</h4>
      <ul className="list-disc list-inside text-gray-300 space-y-1">
        <li>All credentials are encrypted at rest using AES-256-GCM</li>
        <li>Credential values are never exposed in logs or API responses</li>
        <li>Access is limited to workflows in the same tenant</li>
        <li>Credentials can be rotated without updating workflows</li>
      </ul>

      <h4 className="text-lg font-medium text-white mt-6">Using Credentials in Workflows</h4>
      <p className="text-gray-300">
        Reference credentials in node configurations using the credential selector.
        The system will automatically inject the decrypted value at runtime.
      </p>
    </div>
  )
}

function AIBuilderDocs() {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold text-white">AI Workflow Builder</h3>
      <p className="text-gray-300">
        Use natural language to create workflows and leverage AI actions within your automation pipelines.
      </p>

      {/* Setup Section */}
      <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-lg p-4">
        <h4 className="text-yellow-400 font-medium mb-2">⚡ Setup Required</h4>
        <p className="text-gray-300 text-sm">
          AI features require configuration. Set the following environment variables before using:
        </p>
      </div>

      <h4 className="text-lg font-medium text-white mt-6">Environment Configuration</h4>
      <p className="text-gray-300 mb-3">
        Configure these environment variables in your <code className="text-primary-400">.env</code> file:
      </p>
      <CodeBlock code={`# Enable AI Builder
AI_BUILDER_ENABLED=true

# Choose provider: openai, anthropic, or bedrock
AI_BUILDER_PROVIDER=openai

# Provider API Key (for OpenAI or Anthropic)
AI_BUILDER_API_KEY=sk-your-api-key-here

# Model selection (optional, defaults shown)
AI_BUILDER_MODEL=gpt-4
AI_BUILDER_MAX_TOKENS=4096
AI_BUILDER_TEMPERATURE=0.7`} />

      <h4 className="text-lg font-medium text-white mt-6">Supported Providers</h4>
      <div className="space-y-3">
        <div className="bg-gray-900 p-4 rounded-lg border-l-4 border-l-green-500">
          <h5 className="text-white font-medium">OpenAI</h5>
          <p className="text-gray-400 text-sm mt-1">GPT-4, GPT-4 Turbo, GPT-3.5 Turbo</p>
          <CodeBlock code={`AI_BUILDER_PROVIDER=openai
AI_BUILDER_API_KEY=sk-...
AI_BUILDER_MODEL=gpt-4  # or gpt-4-turbo, gpt-3.5-turbo`} />
        </div>
        <div className="bg-gray-900 p-4 rounded-lg border-l-4 border-l-orange-500">
          <h5 className="text-white font-medium">Anthropic (Claude)</h5>
          <p className="text-gray-400 text-sm mt-1">Claude 3 Opus, Claude 3 Sonnet, Claude 3 Haiku</p>
          <CodeBlock code={`AI_BUILDER_PROVIDER=anthropic
AI_BUILDER_API_KEY=sk-ant-...
AI_BUILDER_MODEL=claude-3-sonnet-20240229`} />
        </div>
        <div className="bg-gray-900 p-4 rounded-lg border-l-4 border-l-yellow-500">
          <h5 className="text-white font-medium">AWS Bedrock</h5>
          <p className="text-gray-400 text-sm mt-1">Claude, Titan, and other models via AWS</p>
          <CodeBlock code={`AI_BUILDER_PROVIDER=bedrock
# Uses AWS credentials from environment or IAM role
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=AKIA...
AWS_SECRET_ACCESS_KEY=...
AI_BUILDER_MODEL=anthropic.claude-3-sonnet-20240229-v1:0`} />
        </div>
      </div>

      <h4 className="text-lg font-medium text-white mt-6">AI Builder - Natural Language Workflows</h4>
      <p className="text-gray-300 mb-3">
        Navigate to <strong>AI Builder</strong> in the sidebar to create workflows using natural language.
      </p>
      <ol className="list-decimal list-inside text-gray-300 space-y-2">
        <li>Describe what you want to automate in plain English</li>
        <li>The AI analyzes your request and generates a workflow</li>
        <li>Preview the generated workflow in the visual editor</li>
        <li>Apply the workflow or continue refining with more prompts</li>
      </ol>

      <h4 className="text-lg font-medium text-white mt-6">Example Prompts</h4>
      <div className="space-y-2">
        <div className="bg-gray-900 p-3 rounded-lg">
          <p className="text-gray-300 text-sm italic">
            "Create a workflow that fetches data from an API every hour and sends a Slack notification if the response contains errors"
          </p>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <p className="text-gray-300 text-sm italic">
            "Build a webhook handler that validates incoming JSON, transforms the data, and posts it to another API"
          </p>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <p className="text-gray-300 text-sm italic">
            "Make a workflow that processes customer feedback, uses AI to classify sentiment, and routes to appropriate teams"
          </p>
        </div>
      </div>

      <h4 className="text-lg font-medium text-white mt-6">AI Actions in Workflows</h4>
      <p className="text-gray-300 mb-3">
        In addition to the AI Builder, you can add AI-powered action nodes to your workflows:
      </p>
      <div className="space-y-3">
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">Chat Completion</h5>
          <p className="text-gray-400 text-sm">Generate text responses, answer questions, or have conversations with LLMs</p>
          <div className="mt-2">
            <span className="text-xs text-gray-500">Use cases: </span>
            <span className="text-xs text-gray-400">Customer support responses, content generation, Q&A</span>
          </div>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">Text Classification</h5>
          <p className="text-gray-400 text-sm">Categorize text into predefined labels or categories</p>
          <div className="mt-2">
            <span className="text-xs text-gray-500">Use cases: </span>
            <span className="text-xs text-gray-400">Sentiment analysis, ticket routing, content moderation</span>
          </div>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">Entity Extraction</h5>
          <p className="text-gray-400 text-sm">Extract structured data from unstructured text (names, dates, amounts, etc.)</p>
          <div className="mt-2">
            <span className="text-xs text-gray-500">Use cases: </span>
            <span className="text-xs text-gray-400">Invoice processing, lead extraction, data parsing</span>
          </div>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">Summarization</h5>
          <p className="text-gray-400 text-sm">Generate concise summaries of long documents or text</p>
          <div className="mt-2">
            <span className="text-xs text-gray-500">Use cases: </span>
            <span className="text-xs text-gray-400">Meeting notes, article digests, report summaries</span>
          </div>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">Embeddings</h5>
          <p className="text-gray-400 text-sm">Convert text into vector embeddings for semantic search and similarity</p>
          <div className="mt-2">
            <span className="text-xs text-gray-500">Use cases: </span>
            <span className="text-xs text-gray-400">Semantic search, document clustering, recommendations</span>
          </div>
        </div>
      </div>

      <h4 className="text-lg font-medium text-white mt-6">Using AI Actions in Workflows</h4>
      <p className="text-gray-300 mb-3">
        To add AI actions to your workflow:
      </p>
      <ol className="list-decimal list-inside text-gray-300 space-y-2">
        <li>Drag an "AI Action" node from the palette onto the canvas</li>
        <li>Select the action type (Chat Completion, Classification, etc.)</li>
        <li>Choose the AI provider credential from your saved credentials</li>
        <li>Configure the prompt or input template using data from previous nodes</li>
        <li>Connect the output to subsequent nodes in your workflow</li>
      </ol>

      <h4 className="text-lg font-medium text-white mt-6">Setting Up AI Credentials</h4>
      <p className="text-gray-300 mb-3">
        For AI actions in workflows, store your API keys securely in Credentials:
      </p>
      <ol className="list-decimal list-inside text-gray-300 space-y-2">
        <li>Go to <strong>Credentials → Create Credential</strong></li>
        <li>Select type: <strong>API Key</strong></li>
        <li>Name it descriptively (e.g., "OpenAI Production")</li>
        <li>Paste your API key and save</li>
        <li>Reference this credential in AI action nodes</li>
      </ol>

      <h4 className="text-lg font-medium text-white mt-6">Best Practices</h4>
      <ul className="list-disc list-inside text-gray-300 space-y-1">
        <li>Be specific about triggers, conditions, and expected outputs</li>
        <li>Test with small inputs before processing large batches</li>
        <li>Set appropriate token limits to control costs</li>
        <li>Use lower temperature (0.1-0.3) for consistent, factual outputs</li>
        <li>Use higher temperature (0.7-1.0) for creative or varied outputs</li>
        <li>Store API keys in Credentials, never hardcode them</li>
        <li>Monitor AI action execution times - they can be slower than HTTP requests</li>
        <li>Use error handling nodes after AI actions to handle API failures gracefully</li>
      </ul>

      <h4 className="text-lg font-medium text-white mt-6">Troubleshooting</h4>
      <div className="space-y-3">
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-red-400 font-medium text-sm">AI Builder not available</h5>
          <p className="text-gray-400 text-sm">Ensure <code className="text-primary-400">AI_BUILDER_ENABLED=true</code> and restart the server</p>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-red-400 font-medium text-sm">Authentication errors</h5>
          <p className="text-gray-400 text-sm">Verify your API key is correct and has not expired. Check provider dashboard for key status.</p>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-red-400 font-medium text-sm">Rate limit exceeded</h5>
          <p className="text-gray-400 text-sm">Add delays between AI action calls or upgrade your API plan. Consider using a smaller model for high-volume tasks.</p>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-red-400 font-medium text-sm">Bedrock access denied</h5>
          <p className="text-gray-400 text-sm">Ensure your AWS IAM role has <code className="text-primary-400">bedrock:InvokeModel</code> permission and the model is enabled in your region.</p>
        </div>
      </div>
    </div>
  )
}

function MarketplaceDocs() {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold text-white">Marketplace</h3>
      <p className="text-gray-300">
        Discover and share workflow templates with the community.
      </p>

      <h4 className="text-lg font-medium text-white mt-6">Browsing Templates</h4>
      <ul className="list-disc list-inside text-gray-300 space-y-1">
        <li>Filter by category: Integrations, Data Processing, Notifications, etc.</li>
        <li>Search by keyword or tag</li>
        <li>Sort by popularity, rating, or recency</li>
        <li>Preview template structure before installing</li>
      </ul>

      <h4 className="text-lg font-medium text-white mt-6">Installing Templates</h4>
      <ol className="list-decimal list-inside text-gray-300 space-y-1">
        <li>Click "Install" on any template</li>
        <li>Provide a name for your new workflow</li>
        <li>The template is copied to your workflows</li>
        <li>Configure credentials and settings as needed</li>
      </ol>

      <h4 className="text-lg font-medium text-white mt-6">Publishing Templates</h4>
      <p className="text-gray-300 mb-2">
        Share your workflows with the community:
      </p>
      <ol className="list-decimal list-inside text-gray-300 space-y-1">
        <li>Open your workflow in the editor</li>
        <li>Click "Save as Template" → "Publish to Marketplace"</li>
        <li>Add a description, category, and tags</li>
        <li>Submit for review</li>
      </ol>
    </div>
  )
}

function AnalyticsDocs() {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold text-white">Analytics</h3>
      <p className="text-gray-300">
        Monitor workflow performance and identify issues with analytics.
      </p>

      <h4 className="text-lg font-medium text-white mt-6">Available Metrics</h4>
      <div className="space-y-3">
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">Execution Trends</h5>
          <p className="text-gray-400 text-sm">Track executions over time, see success/failure rates</p>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">Duration Stats</h5>
          <p className="text-gray-400 text-sm">Average, P50, P90, P99 execution times by workflow</p>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">Top Failures</h5>
          <p className="text-gray-400 text-sm">Workflows with the most failures and error previews</p>
        </div>
        <div className="bg-gray-900 p-3 rounded-lg">
          <h5 className="text-white font-medium">Trigger Breakdown</h5>
          <p className="text-gray-400 text-sm">Distribution of executions by trigger type</p>
        </div>
      </div>

      <h4 className="text-lg font-medium text-white mt-6">Time Range Selection</h4>
      <p className="text-gray-300">
        Use the time range selector to view analytics for the last 7, 30, or 90 days.
        All charts and metrics update based on your selection.
      </p>
    </div>
  )
}

function ApiReferenceContent() {
  const [activeEndpoint, setActiveEndpoint] = useState('workflows')

  const endpoints = [
    { id: 'auth', name: 'Authentication' },
    { id: 'workflows', name: 'Workflows' },
    { id: 'webhooks', name: 'Webhooks' },
    { id: 'schedules', name: 'Schedules' },
    { id: 'executions', name: 'Executions' },
    { id: 'credentials', name: 'Credentials' },
  ]

  return (
    <div className="flex gap-6">
      {/* Endpoint Sidebar */}
      <div className="w-48 flex-shrink-0">
        <nav className="space-y-1">
          {endpoints.map((endpoint) => (
            <button
              key={endpoint.id}
              onClick={() => setActiveEndpoint(endpoint.id)}
              className={`w-full text-left px-3 py-2 rounded-md text-sm transition-colors ${
                activeEndpoint === endpoint.id
                  ? 'bg-gray-700 text-white'
                  : 'text-gray-400 hover:text-white hover:bg-gray-700/50'
              }`}
            >
              {endpoint.name}
            </button>
          ))}
        </nav>
      </div>

      {/* Endpoint Content */}
      <div className="flex-1 min-w-0">
        {activeEndpoint === 'auth' && <AuthApiDocs />}
        {activeEndpoint === 'workflows' && <WorkflowsApiDocs />}
        {activeEndpoint === 'webhooks' && <WebhooksApiDocs />}
        {activeEndpoint === 'schedules' && <SchedulesApiDocs />}
        {activeEndpoint === 'executions' && <ExecutionsApiDocs />}
        {activeEndpoint === 'credentials' && <CredentialsApiDocs />}
      </div>
    </div>
  )
}

function AuthApiDocs() {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold text-white">Authentication</h3>
      <p className="text-gray-300">
        All API requests require authentication using one of the following methods.
      </p>

      <h4 className="text-lg font-medium text-white mt-6">API Key Authentication</h4>
      <p className="text-gray-300 mb-2">
        Include your API key in the Authorization header:
      </p>
      <CodeBlock code={`Authorization: Bearer YOUR_API_KEY`} />

      <h4 className="text-lg font-medium text-white mt-6">Tenant ID Header</h4>
      <p className="text-gray-300 mb-2">
        For development, include the tenant ID header:
      </p>
      <CodeBlock code={`X-Tenant-ID: YOUR_TENANT_ID`} />

      <h4 className="text-lg font-medium text-white mt-6">Base URL</h4>
      <CodeBlock code={`https://api.gorax.io/api/v1`} />
    </div>
  )
}

function WorkflowsApiDocs() {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold text-white">Workflows API</h3>

      <ApiEndpoint
        method="GET"
        path="/api/v1/workflows"
        description="List all workflows"
        params={[
          { name: 'limit', type: 'number', description: 'Max results (default: 20)' },
          { name: 'offset', type: 'number', description: 'Pagination offset' },
          { name: 'status', type: 'string', description: 'Filter by status (active, draft, inactive)' },
        ]}
      />

      <ApiEndpoint
        method="GET"
        path="/api/v1/workflows/:id"
        description="Get a workflow by ID"
      />

      <ApiEndpoint
        method="POST"
        path="/api/v1/workflows"
        description="Create a new workflow"
        body={`{
  "name": "My Workflow",
  "description": "A sample workflow",
  "definition": {
    "nodes": [...],
    "edges": [...]
  }
}`}
      />

      <ApiEndpoint
        method="PUT"
        path="/api/v1/workflows/:id"
        description="Update a workflow"
      />

      <ApiEndpoint
        method="DELETE"
        path="/api/v1/workflows/:id"
        description="Delete a workflow"
      />

      <ApiEndpoint
        method="POST"
        path="/api/v1/workflows/:id/execute"
        description="Execute a workflow"
        body={`{
  "input": {
    "key": "value"
  }
}`}
      />
    </div>
  )
}

function WebhooksApiDocs() {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold text-white">Webhooks API</h3>

      <ApiEndpoint
        method="GET"
        path="/api/v1/webhooks"
        description="List all webhooks"
      />

      <ApiEndpoint
        method="GET"
        path="/api/v1/webhooks/:id"
        description="Get a webhook by ID"
      />

      <ApiEndpoint
        method="POST"
        path="/api/v1/webhooks"
        description="Create a new webhook"
        body={`{
  "name": "My Webhook",
  "workflow_id": "uuid",
  "auth_type": "signature",
  "enabled": true
}`}
      />

      <ApiEndpoint
        method="PUT"
        path="/api/v1/webhooks/:id"
        description="Update a webhook"
      />

      <ApiEndpoint
        method="DELETE"
        path="/api/v1/webhooks/:id"
        description="Delete a webhook"
      />

      <ApiEndpoint
        method="POST"
        path="/api/v1/webhooks/:id/regenerate-secret"
        description="Regenerate webhook secret"
      />
    </div>
  )
}

function SchedulesApiDocs() {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold text-white">Schedules API</h3>

      <ApiEndpoint
        method="GET"
        path="/api/v1/schedules"
        description="List all schedules"
      />

      <ApiEndpoint
        method="GET"
        path="/api/v1/schedules/:id"
        description="Get a schedule by ID"
      />

      <ApiEndpoint
        method="POST"
        path="/api/v1/workflows/:id/schedules"
        description="Create a new schedule for a workflow"
        body={`{
  "cron_expression": "0 9 * * *",
  "timezone": "America/New_York",
  "enabled": true
}`}
      />

      <ApiEndpoint
        method="PUT"
        path="/api/v1/schedules/:id"
        description="Update a schedule"
      />

      <ApiEndpoint
        method="DELETE"
        path="/api/v1/schedules/:id"
        description="Delete a schedule"
      />

      <ApiEndpoint
        method="POST"
        path="/api/v1/schedules/parse-cron"
        description="Parse a cron expression"
        body={`{
  "expression": "0 9 * * 1-5"
}`}
      />
    </div>
  )
}

function ExecutionsApiDocs() {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold text-white">Executions API</h3>

      <ApiEndpoint
        method="GET"
        path="/api/v1/executions"
        description="List executions"
        params={[
          { name: 'workflow_id', type: 'string', description: 'Filter by workflow' },
          { name: 'status', type: 'string', description: 'Filter by status' },
          { name: 'limit', type: 'number', description: 'Max results' },
          { name: 'offset', type: 'number', description: 'Pagination offset' },
        ]}
      />

      <ApiEndpoint
        method="GET"
        path="/api/v1/executions/:id"
        description="Get execution details"
      />

      <ApiEndpoint
        method="GET"
        path="/api/v1/executions/:id/steps"
        description="Get execution steps"
      />

      <ApiEndpoint
        method="POST"
        path="/api/v1/executions/:id/cancel"
        description="Cancel a running execution"
      />

      <ApiEndpoint
        method="POST"
        path="/api/v1/executions/:id/retry"
        description="Retry a failed execution"
      />
    </div>
  )
}

function CredentialsApiDocs() {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold text-white">Credentials API</h3>

      <ApiEndpoint
        method="GET"
        path="/api/v1/credentials"
        description="List credentials (metadata only)"
      />

      <ApiEndpoint
        method="GET"
        path="/api/v1/credentials/:id"
        description="Get credential metadata"
      />

      <ApiEndpoint
        method="POST"
        path="/api/v1/credentials"
        description="Create a new credential"
        body={`{
  "name": "My API Key",
  "type": "api_key",
  "data": {
    "key": "sk_live_..."
  }
}`}
      />

      <ApiEndpoint
        method="PUT"
        path="/api/v1/credentials/:id"
        description="Update a credential"
      />

      <ApiEndpoint
        method="DELETE"
        path="/api/v1/credentials/:id"
        description="Delete a credential"
      />

      <ApiEndpoint
        method="POST"
        path="/api/v1/credentials/:id/rotate"
        description="Rotate credential value"
      />
    </div>
  )
}

function QuickReferenceContent() {
  return (
    <div className="space-y-8">
      <Section title="Cron Expression Quick Reference">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-700">
                <th className="text-left py-2 text-gray-400">Expression</th>
                <th className="text-left py-2 text-gray-400">Description</th>
              </tr>
            </thead>
            <tbody className="text-gray-300">
              <tr className="border-b border-gray-700/50">
                <td className="py-2 font-mono">* * * * *</td>
                <td className="py-2">Every minute</td>
              </tr>
              <tr className="border-b border-gray-700/50">
                <td className="py-2 font-mono">*/5 * * * *</td>
                <td className="py-2">Every 5 minutes</td>
              </tr>
              <tr className="border-b border-gray-700/50">
                <td className="py-2 font-mono">0 * * * *</td>
                <td className="py-2">Every hour</td>
              </tr>
              <tr className="border-b border-gray-700/50">
                <td className="py-2 font-mono">0 0 * * *</td>
                <td className="py-2">Every day at midnight</td>
              </tr>
              <tr className="border-b border-gray-700/50">
                <td className="py-2 font-mono">0 9 * * *</td>
                <td className="py-2">Every day at 9 AM</td>
              </tr>
              <tr className="border-b border-gray-700/50">
                <td className="py-2 font-mono">0 9 * * 1-5</td>
                <td className="py-2">Weekdays at 9 AM</td>
              </tr>
              <tr className="border-b border-gray-700/50">
                <td className="py-2 font-mono">0 0 1 * *</td>
                <td className="py-2">First day of month</td>
              </tr>
              <tr>
                <td className="py-2 font-mono">0 0 * * 0</td>
                <td className="py-2">Every Sunday at midnight</td>
              </tr>
            </tbody>
          </table>
        </div>
      </Section>

      <Section title="Filter Operators">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <OperatorCard operator="equals" example='{"field": "status", "op": "equals", "value": "active"}' />
          <OperatorCard operator="not_equals" example='{"field": "type", "op": "not_equals", "value": "test"}' />
          <OperatorCard operator="contains" example='{"field": "name", "op": "contains", "value": "prod"}' />
          <OperatorCard operator="starts_with" example='{"field": "id", "op": "starts_with", "value": "wf-"}' />
          <OperatorCard operator="ends_with" example='{"field": "email", "op": "ends_with", "value": "@company.com"}' />
          <OperatorCard operator="regex" example='{"field": "code", "op": "regex", "value": "^[A-Z]{3}$"}' />
          <OperatorCard operator="gt" example='{"field": "count", "op": "gt", "value": 10}' />
          <OperatorCard operator="lt" example='{"field": "priority", "op": "lt", "value": 5}' />
          <OperatorCard operator="exists" example='{"field": "metadata.tag", "op": "exists"}' />
          <OperatorCard operator="in" example='{"field": "status", "op": "in", "value": ["new", "pending"]}' />
        </div>
      </Section>

      <Section title="Execution Statuses">
        <div className="flex flex-wrap gap-2">
          <span className="px-3 py-1 rounded-full text-sm bg-yellow-500/20 text-yellow-400">queued</span>
          <span className="px-3 py-1 rounded-full text-sm bg-blue-500/20 text-blue-400">running</span>
          <span className="px-3 py-1 rounded-full text-sm bg-green-500/20 text-green-400">completed</span>
          <span className="px-3 py-1 rounded-full text-sm bg-red-500/20 text-red-400">failed</span>
          <span className="px-3 py-1 rounded-full text-sm bg-gray-500/20 text-gray-400">cancelled</span>
          <span className="px-3 py-1 rounded-full text-sm bg-orange-500/20 text-orange-400">timeout</span>
        </div>
      </Section>

      <Section title="HTTP Methods">
        <div className="flex flex-wrap gap-2">
          <span className="px-3 py-1 rounded text-sm bg-green-500/20 text-green-400 font-mono">GET</span>
          <span className="px-3 py-1 rounded text-sm bg-blue-500/20 text-blue-400 font-mono">POST</span>
          <span className="px-3 py-1 rounded text-sm bg-yellow-500/20 text-yellow-400 font-mono">PUT</span>
          <span className="px-3 py-1 rounded text-sm bg-orange-500/20 text-orange-400 font-mono">PATCH</span>
          <span className="px-3 py-1 rounded text-sm bg-red-500/20 text-red-400 font-mono">DELETE</span>
        </div>
      </Section>

      <Section title="Common Headers">
        <CodeBlock code={`Content-Type: application/json
Authorization: Bearer YOUR_API_KEY
X-Tenant-ID: YOUR_TENANT_ID
X-Request-ID: unique-request-id`} />
      </Section>
    </div>
  )
}

// Helper Components

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section>
      <h3 className="text-xl font-semibold text-white mb-4">{title}</h3>
      {children}
    </section>
  )
}

function Step({ number, title, children }: { number: number; title: string; children: React.ReactNode }) {
  return (
    <div className="flex gap-4">
      <div className="flex-shrink-0 w-8 h-8 rounded-full bg-primary-600 text-white flex items-center justify-center font-bold text-sm">
        {number}
      </div>
      <div className="flex-1">
        <h4 className="text-white font-medium mb-2">{title}</h4>
        {children}
      </div>
    </div>
  )
}

function ConceptCard({ title, description }: { title: string; description: string }) {
  return (
    <div className="bg-gray-900 p-4 rounded-lg">
      <h4 className="text-white font-medium mb-1">{title}</h4>
      <p className="text-gray-400 text-sm">{description}</p>
    </div>
  )
}

function NodeTypeCard({ name, description, color }: { name: string; description: string; color: string }) {
  const colorClasses: Record<string, string> = {
    green: 'border-l-green-500',
    blue: 'border-l-blue-500',
    yellow: 'border-l-yellow-500',
    purple: 'border-l-purple-500',
    cyan: 'border-l-cyan-500',
  }

  return (
    <div className={`bg-gray-900 p-3 rounded-lg border-l-4 ${colorClasses[color]}`}>
      <h5 className="text-white font-medium">{name}</h5>
      <p className="text-gray-400 text-sm">{description}</p>
    </div>
  )
}

function CodeBlock({ code }: { code: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = () => {
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="relative">
      <pre className="bg-gray-900 border border-gray-700 rounded-lg p-4 font-mono text-sm text-gray-300 overflow-x-auto">
        {code}
      </pre>
      <button
        onClick={handleCopy}
        className="absolute top-2 right-2 px-2 py-1 text-xs bg-gray-700 text-gray-300 rounded hover:bg-gray-600 transition-colors"
      >
        {copied ? 'Copied!' : 'Copy'}
      </button>
    </div>
  )
}

function CronExample({ expression, description }: { expression: string; description: string }) {
  return (
    <div className="flex items-center justify-between bg-gray-900 px-3 py-2 rounded-lg">
      <code className="text-gray-300 font-mono text-sm">{expression}</code>
      <span className="text-gray-400 text-sm">{description}</span>
    </div>
  )
}

function StatusBadge({ status, description }: { status: string; description: string }) {
  const colors: Record<string, string> = {
    queued: 'bg-yellow-500/20 text-yellow-400',
    running: 'bg-blue-500/20 text-blue-400',
    completed: 'bg-green-500/20 text-green-400',
    failed: 'bg-red-500/20 text-red-400',
    cancelled: 'bg-gray-500/20 text-gray-400',
    timeout: 'bg-orange-500/20 text-orange-400',
  }

  return (
    <div className="flex items-center gap-3">
      <span className={`px-3 py-1 rounded-full text-sm ${colors[status]}`}>{status}</span>
      <span className="text-gray-400 text-sm">{description}</span>
    </div>
  )
}

interface ApiEndpointProps {
  method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  path: string
  description: string
  params?: { name: string; type: string; description: string }[]
  body?: string
}

function ApiEndpoint({ method, path, description, params, body }: ApiEndpointProps) {
  const methodColors: Record<string, string> = {
    GET: 'bg-green-500/20 text-green-400',
    POST: 'bg-blue-500/20 text-blue-400',
    PUT: 'bg-yellow-500/20 text-yellow-400',
    PATCH: 'bg-orange-500/20 text-orange-400',
    DELETE: 'bg-red-500/20 text-red-400',
  }

  return (
    <div className="bg-gray-900 rounded-lg p-4 mb-4">
      <div className="flex items-center gap-3 mb-2">
        <span className={`px-2 py-1 rounded text-xs font-bold ${methodColors[method]}`}>
          {method}
        </span>
        <code className="text-gray-300 font-mono text-sm">{path}</code>
      </div>
      <p className="text-gray-400 text-sm mb-3">{description}</p>

      {params && params.length > 0 && (
        <div className="mt-3">
          <h5 className="text-gray-400 text-xs uppercase mb-2">Query Parameters</h5>
          <div className="space-y-1">
            {params.map((param) => (
              <div key={param.name} className="flex items-start gap-2 text-sm">
                <code className="text-primary-400">{param.name}</code>
                <span className="text-gray-500">({param.type})</span>
                <span className="text-gray-400">{param.description}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {body && (
        <div className="mt-3">
          <h5 className="text-gray-400 text-xs uppercase mb-2">Request Body</h5>
          <pre className="bg-gray-950 rounded p-3 text-xs font-mono text-gray-300 overflow-x-auto">
            {body}
          </pre>
        </div>
      )}
    </div>
  )
}

function OperatorCard({ operator, example }: { operator: string; example: string }) {
  return (
    <div className="bg-gray-900 p-3 rounded-lg">
      <code className="text-primary-400 font-bold">{operator}</code>
      <pre className="text-gray-400 text-xs mt-1 overflow-x-auto">{example}</pre>
    </div>
  )
}
