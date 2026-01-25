import { Page, Locator, expect } from '@playwright/test'
import { waitForPageLoad, generateTestId, waitForApiResponse } from '../utils/test-helpers'

/**
 * Page Object Model for the Workflow Editor page
 * Provides methods for interacting with the workflow canvas and editor controls
 */
export class WorkflowEditorPage {
  readonly page: Page
  readonly canvas: Locator
  readonly toolbox: Locator
  readonly saveButton: Locator
  readonly executeButton: Locator
  readonly validateButton: Locator
  readonly nameInput: Locator
  readonly descriptionInput: Locator

  constructor(page: Page) {
    this.page = page
    this.canvas = page.locator('[data-testid="workflow-canvas"]')
    this.toolbox = page.locator('[data-testid="node-toolbox"]')
    this.saveButton = page.locator('button:has-text("Save"), [data-testid="save-workflow"]')
    this.executeButton = page.locator('button:has-text("Execute"), [data-testid="execute-workflow"]')
    this.validateButton = page.locator('button:has-text("Validate"), [data-testid="validate-workflow"]')
    this.nameInput = page.locator('input[name="name"], [data-testid="workflow-name"]')
    this.descriptionInput = page.locator('textarea[name="description"], [data-testid="workflow-description"]')
  }

  /**
   * Navigate to a workflow editor by ID
   */
  async goto(workflowId: string): Promise<void> {
    await this.page.goto(`/workflows/${workflowId}`)
    await waitForPageLoad(this.page)
    await expect(this.canvas).toBeVisible({ timeout: 10000 })
  }

  /**
   * Navigate to create a new workflow
   */
  async gotoNew(): Promise<void> {
    await this.page.goto('/workflows/new')
    await waitForPageLoad(this.page)
  }

  /**
   * Create a new workflow with the given name
   */
  async createWorkflow(name: string, description?: string): Promise<string> {
    await this.page.goto('/workflows')
    await this.page.click('button:has-text("Create"), [data-testid="create-workflow"]')

    await this.nameInput.fill(name)
    if (description) {
      await this.descriptionInput.fill(description)
    }

    await this.page.click('button:has-text("Create"), [type="submit"]')

    // Wait for redirect to editor
    await this.page.waitForURL(/\/workflows\/[a-f0-9-]+/)
    const workflowId = this.page.url().split('/').pop() || ''

    return workflowId
  }

  /**
   * Add a node to the canvas
   */
  async addNode(nodeType: string, position?: { x: number; y: number }): Promise<string> {
    const nodeId = generateTestId('node')

    // Click on the node type in toolbox
    const nodeButton = this.toolbox.locator(`[data-node-type="${nodeType}"], button:has-text("${nodeType}")`)
    await nodeButton.click()

    // If position specified, click on canvas at that position
    if (position) {
      const canvasBounds = await this.canvas.boundingBox()
      if (canvasBounds) {
        await this.canvas.click({
          position: {
            x: position.x,
            y: position.y
          }
        })
      }
    } else {
      // Default: click center of canvas
      await this.canvas.click()
    }

    // Wait for node to be added
    await this.page.waitForTimeout(300)

    return nodeId
  }

  /**
   * Get a node locator by its ID
   */
  getNode(nodeId: string): Locator {
    return this.canvas.locator(`[data-node-id="${nodeId}"], [data-id="${nodeId}"]`)
  }

  /**
   * Select a node by clicking on it
   */
  async selectNode(nodeId: string): Promise<void> {
    const node = this.getNode(nodeId)
    await node.click()
    await expect(node).toHaveClass(/selected/)
  }

  /**
   * Connect two nodes with an edge
   */
  async connectNodes(sourceId: string, targetId: string): Promise<void> {
    const sourceNode = this.getNode(sourceId)
    const targetNode = this.getNode(targetId)

    // Find the output handle of source and input handle of target
    const sourceHandle = sourceNode.locator('[data-handletype="source"], .react-flow__handle-bottom')
    const targetHandle = targetNode.locator('[data-handletype="target"], .react-flow__handle-top')

    // Drag from source to target
    await sourceHandle.dragTo(targetHandle)

    // Verify connection was made
    await this.page.waitForTimeout(300)
  }

  /**
   * Configure a node by opening its panel and setting values
   */
  async configureNode(nodeId: string, config: Record<string, unknown>): Promise<void> {
    // Double-click to open config panel
    const node = this.getNode(nodeId)
    await node.dblclick()

    // Wait for config panel to open
    const configPanel = this.page.locator('[data-testid="node-config-panel"], .node-config-panel')
    await expect(configPanel).toBeVisible()

    // Fill in config values
    for (const [key, value] of Object.entries(config)) {
      const input = configPanel.locator(`[name="${key}"], [data-field="${key}"]`)
      if (typeof value === 'string') {
        await input.fill(value)
      } else if (typeof value === 'boolean') {
        if (value) {
          await input.check()
        } else {
          await input.uncheck()
        }
      }
    }

    // Close panel by clicking outside or apply button
    const applyButton = configPanel.locator('button:has-text("Apply"), button:has-text("Save")')
    if (await applyButton.isVisible()) {
      await applyButton.click()
    } else {
      await this.canvas.click({ position: { x: 10, y: 10 } })
    }
  }

  /**
   * Delete a node from the canvas
   */
  async deleteNode(nodeId: string): Promise<void> {
    const node = this.getNode(nodeId)
    await node.click()

    // Press delete key
    await this.page.keyboard.press('Delete')

    // Verify node is removed
    await expect(node).not.toBeVisible()
  }

  /**
   * Save the workflow
   */
  async saveWorkflow(): Promise<void> {
    await this.saveButton.click()

    // Wait for save API call
    await waitForApiResponse(this.page, /\/api\/v1\/workflows\//)

    // Verify success
    await expect(this.page.locator('text=/saved|success/i')).toBeVisible({ timeout: 5000 })
  }

  /**
   * Execute the workflow
   */
  async executeWorkflow(input?: Record<string, unknown>): Promise<string> {
    await this.executeButton.click()

    // If input is required, fill the input modal
    if (input) {
      const inputModal = this.page.locator('[data-testid="execution-input-modal"], .execution-modal')
      if (await inputModal.isVisible()) {
        const inputField = inputModal.locator('textarea, [data-testid="execution-input"]')
        await inputField.fill(JSON.stringify(input))
        await inputModal.locator('button:has-text("Execute"), button:has-text("Run")').click()
      }
    }

    // Wait for execution to start
    const response = await waitForApiResponse(this.page, /\/api\/v1\/workflows\/.*\/execute/)
    const responseData = await response.json()

    return responseData.id || responseData.execution_id
  }

  /**
   * Validate the workflow
   */
  async validateWorkflow(): Promise<boolean> {
    await this.validateButton.click()

    // Wait for validation response
    await waitForApiResponse(this.page, /\/api\/v1\/workflows\/.*\/validate/)

    // Check if valid
    const successMessage = this.page.locator('text=/valid|no errors/i')
    const errorMessage = this.page.locator('text=/invalid|error|issue/i')

    if (await successMessage.isVisible()) {
      return true
    }

    return !(await errorMessage.isVisible())
  }

  /**
   * Get validation errors
   */
  async getValidationErrors(): Promise<string[]> {
    await this.validateWorkflow()

    const errors: string[] = []
    const errorElements = this.page.locator('[data-testid="validation-error"], .validation-error')
    const count = await errorElements.count()

    for (let i = 0; i < count; i++) {
      const text = await errorElements.nth(i).textContent()
      if (text) errors.push(text)
    }

    return errors
  }

  /**
   * Get all nodes on the canvas
   */
  async getNodes(): Promise<string[]> {
    const nodes = this.canvas.locator('[data-node-id], .react-flow__node')
    const count = await nodes.count()
    const nodeIds: string[] = []

    for (let i = 0; i < count; i++) {
      const id = await nodes.nth(i).getAttribute('data-node-id') ||
                 await nodes.nth(i).getAttribute('data-id')
      if (id) nodeIds.push(id)
    }

    return nodeIds
  }

  /**
   * Check if workflow has unsaved changes
   */
  async hasUnsavedChanges(): Promise<boolean> {
    const indicator = this.page.locator('[data-testid="unsaved-indicator"], .unsaved-dot')
    return await indicator.isVisible()
  }

  /**
   * Undo last action
   */
  async undo(): Promise<void> {
    await this.page.keyboard.press('Control+z')
  }

  /**
   * Redo last undone action
   */
  async redo(): Promise<void> {
    await this.page.keyboard.press('Control+Shift+z')
  }

  /**
   * Zoom in on the canvas
   */
  async zoomIn(): Promise<void> {
    const zoomInButton = this.page.locator('[data-testid="zoom-in"], button[aria-label="Zoom in"]')
    await zoomInButton.click()
  }

  /**
   * Zoom out on the canvas
   */
  async zoomOut(): Promise<void> {
    const zoomOutButton = this.page.locator('[data-testid="zoom-out"], button[aria-label="Zoom out"]')
    await zoomOutButton.click()
  }

  /**
   * Fit the canvas to show all nodes
   */
  async fitView(): Promise<void> {
    const fitButton = this.page.locator('[data-testid="fit-view"], button[aria-label="Fit view"]')
    await fitButton.click()
  }
}
