import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import type { Node, Edge } from '@xyflow/react'
import WorkflowCanvas from '../components/canvas/WorkflowCanvas'
import NodePalette from '../components/canvas/NodePalette'
import PropertyPanel from '../components/canvas/PropertyPanel'
import VersionHistory from '../components/workflow/VersionHistory'
import { TemplateBrowser, SaveAsTemplate } from '../components/templates'
import { useWorkflow, useWorkflowMutations } from '../hooks/useWorkflows'
import { useTemplateMutations } from '../hooks/useTemplates'
import { workflowAPI, type DryRunResult } from '../api/workflows'
import type { Template } from '../api/templates'
import {
  serializeWorkflowForBackend,
  deserializeWorkflowFromBackend,
  type FrontendNode,
  type BackendWorkflowDefinition,
} from '../utils/nodeTypeMapper'
import {
  WorkflowErrorBoundary,
  CanvasErrorBoundary,
  PanelErrorBoundary,
} from '../components/ErrorBoundary/WorkflowErrorBoundary'
import { useWorkflowRecovery, useAutoSaveWorkflow } from '../hooks/useWorkflowRecovery'
import { errorLogger } from '../services/errorLogger'

export default function WorkflowEditor() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNewWorkflow = id === 'new'

  // For existing workflows, use the ID directly - backend will validate
  const workflowId = !isNewWorkflow && id ? id : null

  // Load existing workflow if editing
  const { workflow, loading, error } = useWorkflow(workflowId)
  const { createWorkflow, updateWorkflow, creating, updating } = useWorkflowMutations()

  // Form state
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [nodes, setNodes] = useState<Node[]>([])
  const [edges, setEdges] = useState<Edge[]>([])
  const [selectedNode, setSelectedNode] = useState<Node | null>(null)
  const [dataLoaded, setDataLoaded] = useState(isNewWorkflow) // New workflows are "loaded" immediately
  const [validationError, setValidationError] = useState<string | null>(null)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [saveSuccess, setSaveSuccess] = useState<string | null>(null)
  const [showVersionHistory, setShowVersionHistory] = useState(false)
  const [showTemplateBrowser, setShowTemplateBrowser] = useState(false)
  const [showSaveAsTemplate, setShowSaveAsTemplate] = useState(false)
  const [showDryRunResults, setShowDryRunResults] = useState(false)
  const [dryRunResult, setDryRunResult] = useState<DryRunResult | null>(null)
  const [dryRunLoading, setDryRunLoading] = useState(false)
  const [dryRunError, setDryRunError] = useState<string | null>(null)

  const { instantiateTemplate } = useTemplateMutations()

  // Workflow recovery
  const {
    lastBackupTime,
    recoverFromBackup,
    clearBackup,
    getBackup,
  } = useWorkflowRecovery(workflowId)

  const [showRecoveryPrompt, setShowRecoveryPrompt] = useState(false)
  const [recoveryChecked, setRecoveryChecked] = useState(false)

  // Auto-save workflow state
  const { lastSaved: autoSaveTime } = useAutoSaveWorkflow(
    workflowId ?? 'new',
    nodes,
    edges,
    name,
    description,
    { enabled: dataLoaded && nodes.length > 0 }
  )

  // Check for recovery backup on mount
  useEffect(() => {
    if (workflowId && !recoveryChecked && dataLoaded) {
      const backup = getBackup(workflowId)
      if (backup && backup.timestamp > (workflow?.updatedAt ? new Date(workflow.updatedAt).getTime() : 0)) {
        setShowRecoveryPrompt(true)
      }
      setRecoveryChecked(true)
    }
  }, [workflowId, recoveryChecked, dataLoaded, getBackup, workflow?.updatedAt])

  // Handle recovery from backup
  const handleRecoverFromBackup = useCallback(() => {
    if (!workflowId) return

    const snapshot = recoverFromBackup(workflowId)
    if (snapshot) {
      setNodes(snapshot.nodes)
      setEdges(snapshot.edges)
      setName(snapshot.name)
      setDescription(snapshot.description)
      setSaveSuccess('Workflow recovered from backup')
      setTimeout(() => setSaveSuccess(null), 3000)
    }
    setShowRecoveryPrompt(false)
  }, [workflowId, recoverFromBackup])

  // Dismiss recovery prompt
  const handleDismissRecovery = useCallback(() => {
    if (workflowId) {
      clearBackup(workflowId)
    }
    setShowRecoveryPrompt(false)
  }, [workflowId, clearBackup])

  // Error handler for error boundaries
  const handleBoundaryError = useCallback((error: Error, errorInfo: React.ErrorInfo) => {
    errorLogger.logBoundaryError(
      error,
      errorInfo,
      'WorkflowEditor',
      { workflowId: workflowId ?? undefined }
    )
  }, [workflowId])

  // Load workflow data when editing
  useEffect(() => {
    if (workflow && !isNewWorkflow) {
      console.log('[WorkflowEditor] Loading workflow:', workflow.id, workflow.name)
      console.log('[WorkflowEditor] Raw definition:', workflow.definition)

      setName(workflow.name)
      setDescription(workflow.description || '')

      // Deserialize backend format to frontend format
      // Converts "trigger:webhook" -> { type: "trigger", data: { nodeType: "webhook" } }
      if (workflow.definition) {
        const backendDef = workflow.definition as unknown as BackendWorkflowDefinition
        console.log('[WorkflowEditor] Backend definition nodes:', backendDef.nodes)
        const frontendDef = deserializeWorkflowFromBackend(backendDef)
        console.log('[WorkflowEditor] Frontend definition nodes:', frontendDef.nodes)
        setNodes(frontendDef.nodes as Node[])
        setEdges(frontendDef.edges as Edge[])
      } else {
        console.log('[WorkflowEditor] No definition found, setting empty arrays')
        setNodes([])
        setEdges([])
      }
      // Mark data as loaded after setting state
      setDataLoaded(true)
    }
  }, [workflow, isNewWorkflow])

  const handleAddNode = (nodeData: { type: string; nodeType: string }) => {
    const newNode: Node = {
      id: `node-${Date.now()}`,
      type: nodeData.type,
      position: { x: 250, y: 100 },
      data: {
        nodeType: nodeData.nodeType,
        label: `New ${nodeData.nodeType}`,
      },
    }
    setNodes((prev) => [...prev, newNode])
  }

  const handleCanvasChange = (workflow: { nodes: Node[]; edges: Edge[] }) => {
    setNodes(workflow.nodes)
    setEdges(workflow.edges)
  }

  const handleNodeSelect = (node: Node | null) => {
    setSelectedNode(node)
  }

  const handleNodeUpdate = (nodeId: string, data: any) => {
    setNodes((prev) =>
      prev.map((node) => (node.id === nodeId ? { ...node, data: { ...node.data, ...data } } : node))
    )
  }

  const handleSave = async () => {
    // Validate
    if (!name || name.trim() === '') {
      setValidationError('Workflow name is required')
      return
    }

    setValidationError(null)
    setSaveError(null)
    setSaveSuccess(null)

    // Convert frontend nodes to backend format using the type mapper
    const frontendDefinition = {
      nodes: nodes.map((node) => ({
        id: node.id,
        type: node.type,
        position: node.position,
        data: node.data || {},
      })) as FrontendNode[],
      edges: edges.map((edge) => ({
        id: edge.id,
        source: edge.source,
        target: edge.target,
        sourceHandle: edge.sourceHandle || undefined,
        targetHandle: edge.targetHandle || undefined,
      })),
    }

    console.log('[WorkflowEditor] Saving - frontend nodes:', frontendDefinition.nodes)

    // Serialize to backend format (converts "trigger" + "webhook" -> "trigger:webhook")
    const backendDefinition = serializeWorkflowForBackend(frontendDefinition)

    console.log('[WorkflowEditor] Saving - backend nodes:', backendDefinition.nodes)

    const workflowData = {
      name,
      description,
      definition: backendDefinition,
    }

    console.log('[WorkflowEditor] Saving workflow data:', workflowData)

    try {
      if (isNewWorkflow) {
        const newWorkflow = await createWorkflow(workflowData)
        navigate(`/workflows/${newWorkflow.id}`)
      } else if (workflowId) {
        await updateWorkflow(workflowId, workflowData)
        setSaveSuccess('Workflow saved successfully')
        setTimeout(() => setSaveSuccess(null), 3000)
      } else {
        setSaveError('Cannot save workflow without ID')
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to save workflow'
      setSaveError(errorMessage)
    }
  }

  const handleTestWorkflow = async () => {
    if (isNewWorkflow || !workflowId) {
      setDryRunError('Please save the workflow before testing')
      setTimeout(() => setDryRunError(null), 3000)
      return
    }

    setDryRunLoading(true)
    setDryRunError(null)

    try {
      const result = await workflowAPI.dryRun(workflowId, {})
      setDryRunResult(result)
      setShowDryRunResults(true)
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to test workflow'
      setDryRunError(errorMessage)
      setTimeout(() => setDryRunError(null), 5000)
    } finally {
      setDryRunLoading(false)
    }
  }

  const handleVersionRestore = (version: number) => {
    setSaveSuccess(`Workflow restored to version ${version}`)
    setShowVersionHistory(false)
    // Reload workflow data
    window.location.reload()
  }

  const handleSelectTemplate = async (template: Template) => {
    try {
      const result = await instantiateTemplate(template.id, {
        workflowName: name || template.name
      })

      // Deserialize backend format to frontend format
      const backendDef = result.definition as unknown as BackendWorkflowDefinition
      const frontendDef = deserializeWorkflowFromBackend(backendDef)
      setNodes(frontendDef.nodes as Node[])
      setEdges(frontendDef.edges as Edge[])
      if (!name) {
        setName(result.workflowName)
      }
      setShowTemplateBrowser(false)
      setSaveSuccess('Template loaded successfully')
      setTimeout(() => setSaveSuccess(null), 3000)
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load template'
      setSaveError(errorMessage)
    }
  }

  const handleSaveAsTemplateSuccess = () => {
    setShowSaveAsTemplate(false)
    setSaveSuccess('Template saved successfully')
    setTimeout(() => setSaveSuccess(null), 3000)
  }

  // Loading state - wait for both the API call AND the data to be processed
  if ((loading || !dataLoaded) && !isNewWorkflow) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-900">
        <div className="text-white text-xl">Loading workflow...</div>
      </div>
    )
  }

  // Error state
  if (error && !isNewWorkflow) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-900">
        <div className="text-center">
          <div className="text-red-400 text-xl mb-4">Failed to load workflow</div>
          <div className="text-gray-400">{error.message}</div>
          <Link
            to="/workflows"
            className="mt-4 inline-block px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700"
          >
            Back to Workflows
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="h-screen bg-gray-900 flex flex-col">
      {/* Header */}
      <div className="bg-gray-800 border-b border-gray-700 px-6 py-4">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-4">
            <Link
              to="/workflows"
              className="text-gray-400 hover:text-white transition-colors"
              aria-label="Back to workflows"
            >
              ← Back to Workflows
            </Link>
            <h1 className="text-2xl font-bold text-white">
              {isNewWorkflow ? 'New Workflow' : 'Edit Workflow'}
            </h1>
            {!isNewWorkflow && workflow && (
              <span className="text-sm text-gray-400">
                v{workflow.version}
              </span>
            )}
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setShowTemplateBrowser(true)}
              className="px-4 py-2 text-gray-300 hover:text-white border border-gray-600 rounded-lg font-medium hover:border-gray-500 transition-colors"
            >
              Browse Templates
            </button>
            {!isNewWorkflow && (
              <>
                <button
                  onClick={() => setShowSaveAsTemplate(true)}
                  className="px-4 py-2 text-gray-300 hover:text-white border border-gray-600 rounded-lg font-medium hover:border-gray-500 transition-colors"
                >
                  Save as Template
                </button>
                <button
                  onClick={() => setShowVersionHistory(!showVersionHistory)}
                  className="px-4 py-2 text-gray-300 hover:text-white border border-gray-600 rounded-lg font-medium hover:border-gray-500 transition-colors"
                >
                  Version History
                </button>
              </>
            )}
            {!isNewWorkflow && (
              <button
                onClick={handleTestWorkflow}
                disabled={dryRunLoading}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {dryRunLoading ? 'Testing...' : 'Test Workflow'}
              </button>
            )}
            <button
              onClick={handleSave}
              disabled={creating || updating}
              className="px-6 py-2 bg-primary-600 text-white rounded-lg font-medium hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {creating || updating ? 'Saving...' : 'Save Workflow'}
            </button>
          </div>
        </div>

        {/* Workflow Metadata */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label htmlFor="workflow-name" className="block text-sm font-medium text-gray-300 mb-2">
              Workflow Name *
            </label>
            <input
              id="workflow-name"
              type="text"
              value={name}
              onChange={(e) => {
                setName(e.target.value)
                if (validationError) setValidationError(null)
              }}
              placeholder="Enter workflow name"
              className="w-full px-4 py-2 bg-gray-700 text-white rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
            {validationError && <div className="mt-1 text-sm text-red-400">{validationError}</div>}
          </div>
          <div>
            <label htmlFor="workflow-description" className="block text-sm font-medium text-gray-300 mb-2">
              Description
            </label>
            <textarea
              id="workflow-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Enter workflow description"
              rows={1}
              className="w-full px-4 py-2 bg-gray-700 text-white rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 resize-none"
            />
          </div>
        </div>

        {/* Save Error Message */}
        {saveError && (
          <div className="mt-4 p-3 bg-red-900/20 border border-red-500/30 text-red-400 text-sm rounded">
            {saveError}
          </div>
        )}

        {/* Save Success Message */}
        {saveSuccess && (
          <div className="mt-4 p-3 bg-green-900/20 border border-green-500/30 text-green-400 text-sm rounded">
            {saveSuccess}
          </div>
        )}

        {/* Dry Run Error Message */}
        {dryRunError && (
          <div className="mt-4 p-3 bg-red-900/20 border border-red-500/30 text-red-400 text-sm rounded">
            {dryRunError}
          </div>
        )}
      </div>

      {/* Recovery Prompt */}
      {showRecoveryPrompt && (
        <div className="bg-blue-900/30 border-b border-blue-500/30 px-6 py-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <svg className="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-blue-300 text-sm">
                A backup of unsaved changes was found from {lastBackupTime?.toLocaleString()}.
              </span>
            </div>
            <div className="flex items-center space-x-2">
              <button
                onClick={handleRecoverFromBackup}
                className="px-3 py-1 bg-blue-600 text-white rounded text-sm font-medium hover:bg-blue-700 transition-colors"
              >
                Recover
              </button>
              <button
                onClick={handleDismissRecovery}
                className="px-3 py-1 bg-gray-700 text-gray-300 rounded text-sm font-medium hover:bg-gray-600 transition-colors"
              >
                Dismiss
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Auto-save indicator */}
      {autoSaveTime && (
        <div className="absolute bottom-4 left-4 text-xs text-gray-500 z-10">
          Auto-saved at {autoSaveTime.toLocaleTimeString()}
        </div>
      )}

      {/* Editor Layout */}
      <div className="flex-1 flex overflow-hidden">
        {/* Node Palette */}
        <PanelErrorBoundary
          componentName="NodePalette"
          onError={handleBoundaryError}
          title="Node Palette Error"
        >
          <NodePalette onAddNode={handleAddNode} />
        </PanelErrorBoundary>

        {/* Canvas */}
        <div className="flex-1 relative">
          <CanvasErrorBoundary
            workflowId={workflowId ?? undefined}
            onError={handleBoundaryError}
          >
            <WorkflowCanvas
              key={`${workflow?.id || 'new'}-v${workflow?.version || 0}`}
              initialNodes={nodes}
              initialEdges={edges}
              onChange={handleCanvasChange}
              onNodeSelect={handleNodeSelect}
              onSave={handleSave}
            />
          </CanvasErrorBoundary>
        </div>

        {/* Property Panel */}
        <PanelErrorBoundary
          componentName="PropertyPanel"
          onError={handleBoundaryError}
          title="Property Panel Error"
        >
          <PropertyPanel
            node={selectedNode}
            onUpdate={handleNodeUpdate}
            onClose={() => setSelectedNode(null)}
            onSave={handleSave}
            isSaving={creating || updating}
          />
        </PanelErrorBoundary>

        {/* Template Browser Modal */}
        {showTemplateBrowser && (
          <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center z-20">
            <div className="bg-gray-800 rounded-lg shadow-xl max-w-6xl w-full max-h-[90vh] overflow-hidden">
              <WorkflowErrorBoundary
                componentName="TemplateBrowser"
                onError={handleBoundaryError}
                fallbackType="inline"
              >
                <TemplateBrowser
                  onSelectTemplate={handleSelectTemplate}
                  onClose={() => setShowTemplateBrowser(false)}
                />
              </WorkflowErrorBoundary>
            </div>
          </div>
        )}

        {/* Save as Template Modal */}
        {showSaveAsTemplate && !isNewWorkflow && workflow && (
          <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center z-20">
            <div className="bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-auto">
              <WorkflowErrorBoundary
                componentName="SaveAsTemplate"
                onError={handleBoundaryError}
                fallbackType="inline"
              >
                <SaveAsTemplate
                  workflowId={workflow.id}
                  workflowName={name}
                  definition={serializeWorkflowForBackend({
                    nodes: nodes.map(node => ({
                      id: node.id,
                      type: node.type,
                      position: node.position,
                      data: node.data || {},
                    })) as FrontendNode[],
                    edges: edges.map(edge => ({
                      id: edge.id,
                      source: edge.source,
                      target: edge.target,
                      sourceHandle: edge.sourceHandle ?? undefined,
                      targetHandle: edge.targetHandle ?? undefined,
                      label: edge.label as string | undefined,
                    })),
                  })}
                  onSuccess={handleSaveAsTemplateSuccess}
                  onCancel={() => setShowSaveAsTemplate(false)}
                />
              </WorkflowErrorBoundary>
            </div>
          </div>
        )}

        {/* Version History Slide-out */}
        {showVersionHistory && !isNewWorkflow && workflow && (
          <div className="absolute right-0 top-0 bottom-0 w-96 bg-gray-800 border-l border-gray-700 shadow-xl z-10">
            <WorkflowErrorBoundary
              componentName="VersionHistory"
              onError={handleBoundaryError}
              fallbackType="panel"
            >
              <VersionHistory
                workflowId={workflow.id}
                currentVersion={workflow.version}
                onRestore={handleVersionRestore}
                onClose={() => setShowVersionHistory(false)}
              />
            </WorkflowErrorBoundary>
          </div>
        )}

        {/* Dry Run Results Modal */}
        {showDryRunResults && dryRunResult && (
          <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center z-20">
            <div className="bg-gray-800 rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-hidden">
              <div className="p-6 border-b border-gray-700">
                <div className="flex items-center justify-between">
                  <h2 className="text-xl font-bold text-white">Workflow Test Results</h2>
                  <button
                    onClick={() => setShowDryRunResults(false)}
                    className="text-gray-400 hover:text-white transition-colors"
                  >
                    ✕
                  </button>
                </div>
              </div>

              <div className="p-6 overflow-y-auto max-h-[calc(90vh-120px)]">
                {/* Validation Status */}
                <div className={`p-4 rounded-lg mb-6 ${dryRunResult.valid ? 'bg-green-900/20 border border-green-500/30' : 'bg-red-900/20 border border-red-500/30'}`}>
                  <div className="flex items-center space-x-2">
                    <span className="text-lg">
                      {dryRunResult.valid ? '✓' : '✗'}
                    </span>
                    <span className={`font-semibold ${dryRunResult.valid ? 'text-green-400' : 'text-red-400'}`}>
                      {dryRunResult.valid ? 'Workflow is valid' : 'Workflow has errors'}
                    </span>
                  </div>
                </div>

                {/* Execution Order */}
                <div className="mb-6">
                  <h3 className="text-lg font-semibold text-white mb-3">Execution Order</h3>
                  <div className="flex flex-wrap gap-2">
                    {dryRunResult.executionOrder.map((nodeId, index) => (
                      <div key={nodeId} className="flex items-center">
                        <div className="px-3 py-1 bg-gray-700 text-gray-200 rounded-lg text-sm">
                          {index + 1}. {nodeId}
                        </div>
                        {index < dryRunResult.executionOrder.length - 1 && (
                          <span className="mx-2 text-gray-500">→</span>
                        )}
                      </div>
                    ))}
                  </div>
                </div>

                {/* Errors */}
                {dryRunResult.errors.length > 0 && (
                  <div className="mb-6">
                    <h3 className="text-lg font-semibold text-red-400 mb-3">Errors ({dryRunResult.errors.length})</h3>
                    <div className="space-y-2">
                      {dryRunResult.errors.map((error, index) => (
                        <div key={index} className="p-3 bg-red-900/20 border border-red-500/30 rounded-lg">
                          <div className="flex items-start space-x-2">
                            <span className="text-red-400 font-bold">•</span>
                            <div className="flex-1">
                              <div className="text-red-400 font-medium">
                                Node: {error.nodeId || 'Workflow'} {error.field && `(${error.field})`}
                              </div>
                              <div className="text-red-300 text-sm mt-1">{error.message}</div>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* Warnings */}
                {dryRunResult.warnings.length > 0 && (
                  <div className="mb-6">
                    <h3 className="text-lg font-semibold text-yellow-400 mb-3">Warnings ({dryRunResult.warnings.length})</h3>
                    <div className="space-y-2">
                      {dryRunResult.warnings.map((warning, index) => (
                        <div key={index} className="p-3 bg-yellow-900/20 border border-yellow-500/30 rounded-lg">
                          <div className="flex items-start space-x-2">
                            <span className="text-yellow-400 font-bold">⚠</span>
                            <div className="flex-1">
                              <div className="text-yellow-400 font-medium">Node: {warning.nodeId}</div>
                              <div className="text-yellow-300 text-sm mt-1">{warning.message}</div>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* Variable Mapping */}
                <div>
                  <h3 className="text-lg font-semibold text-white mb-3">Variable Mapping</h3>
                  <div className="bg-gray-900/50 rounded-lg p-4">
                    <div className="space-y-2">
                      {Object.entries(dryRunResult.variableMapping).map(([variable, source]) => (
                        <div key={variable} className="flex items-center justify-between text-sm">
                          <span className="text-gray-300 font-mono">{variable}</span>
                          <span className="text-gray-500">→</span>
                          <span className="text-gray-400">{source}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>

              <div className="p-6 border-t border-gray-700 flex justify-end">
                <button
                  onClick={() => setShowDryRunResults(false)}
                  className="px-6 py-2 bg-primary-600 text-white rounded-lg font-medium hover:bg-primary-700 transition-colors"
                >
                  Close
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
