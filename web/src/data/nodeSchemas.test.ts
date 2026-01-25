/**
 * Tests for nodeSchemas data
 */

import { describe, it, expect } from 'vitest'
import {
  NODE_SCHEMAS,
  getNodeSchema,
  getSchemasByCategory,
  getAllNodeTypes,
} from './nodeSchemas'

describe('NODE_SCHEMAS', () => {
  it('should have schemas for all trigger types', () => {
    expect(NODE_SCHEMAS.webhook).toBeDefined()
    expect(NODE_SCHEMAS.schedule).toBeDefined()
    expect(NODE_SCHEMAS.manual).toBeDefined()
  })

  it('should have schemas for action types', () => {
    expect(NODE_SCHEMAS.http).toBeDefined()
    expect(NODE_SCHEMAS.transform).toBeDefined()
    expect(NODE_SCHEMAS.script).toBeDefined()
    expect(NODE_SCHEMAS.email).toBeDefined()
  })

  it('should have schemas for Slack actions', () => {
    expect(NODE_SCHEMAS.slack_send_message).toBeDefined()
    expect(NODE_SCHEMAS.slack_send_dm).toBeDefined()
    expect(NODE_SCHEMAS.slack_update_message).toBeDefined()
    expect(NODE_SCHEMAS.slack_add_reaction).toBeDefined()
  })

  it('should have schemas for AI actions', () => {
    expect(NODE_SCHEMAS.ai_chat).toBeDefined()
    expect(NODE_SCHEMAS.ai_summarize).toBeDefined()
    expect(NODE_SCHEMAS.ai_classify).toBeDefined()
    expect(NODE_SCHEMAS.ai_extract).toBeDefined()
    expect(NODE_SCHEMAS.ai_embed).toBeDefined()
  })

  it('should have schemas for control nodes', () => {
    expect(NODE_SCHEMAS.conditional).toBeDefined()
    expect(NODE_SCHEMAS.loop).toBeDefined()
    expect(NODE_SCHEMAS.parallel).toBeDefined()
    expect(NODE_SCHEMAS.delay).toBeDefined()
  })

  it('should have required fields in all schemas', () => {
    Object.entries(NODE_SCHEMAS).forEach(([type, schema]) => {
      expect(schema.type).toBe(type)
      expect(schema.label).toBeTruthy()
      expect(schema.description).toBeTruthy()
      expect(schema.icon).toBeTruthy()
      expect(schema.category).toMatch(/^(trigger|action|ai|control)$/)
      expect(Array.isArray(schema.fields)).toBe(true)
    })
  })

  it('should have label field in all schemas', () => {
    Object.values(NODE_SCHEMAS).forEach((schema) => {
      const labelField = schema.fields.find((f) => f.name === 'label')
      expect(labelField).toBeDefined()
      expect(labelField?.required).toBe(true)
    })
  })
})

describe('getNodeSchema', () => {
  it('should return schema for valid node type', () => {
    const schema = getNodeSchema('webhook')

    expect(schema).toBeDefined()
    expect(schema?.type).toBe('webhook')
    expect(schema?.category).toBe('trigger')
  })

  it('should return undefined for invalid node type', () => {
    const schema = getNodeSchema('nonexistent')

    expect(schema).toBeUndefined()
  })
})

describe('getSchemasByCategory', () => {
  it('should return all trigger schemas', () => {
    const triggers = getSchemasByCategory('trigger')

    expect(triggers.length).toBeGreaterThan(0)
    expect(triggers.every((s) => s.category === 'trigger')).toBe(true)
  })

  it('should return all action schemas', () => {
    const actions = getSchemasByCategory('action')

    expect(actions.length).toBeGreaterThan(0)
    expect(actions.every((s) => s.category === 'action')).toBe(true)
  })

  it('should return all AI schemas', () => {
    const aiSchemas = getSchemasByCategory('ai')

    expect(aiSchemas.length).toBeGreaterThan(0)
    expect(aiSchemas.every((s) => s.category === 'ai')).toBe(true)
  })

  it('should return all control schemas', () => {
    const controls = getSchemasByCategory('control')

    expect(controls.length).toBeGreaterThan(0)
    expect(controls.every((s) => s.category === 'control')).toBe(true)
  })
})

describe('getAllNodeTypes', () => {
  it('should return all node type keys', () => {
    const types = getAllNodeTypes()

    expect(types).toContain('webhook')
    expect(types).toContain('http')
    expect(types).toContain('ai_chat')
    expect(types).toContain('conditional')
    expect(types.length).toBe(Object.keys(NODE_SCHEMAS).length)
  })
})

describe('Schema field validation rules', () => {
  it('should have validation rules for URL fields', () => {
    const httpSchema = getNodeSchema('http')
    const urlField = httpSchema?.fields.find((f) => f.name === 'url')

    expect(urlField?.validation?.pattern).toBeDefined()
  })

  it('should have min/max validation for number fields', () => {
    const loopSchema = getNodeSchema('loop')
    const maxIterationsField = loopSchema?.fields.find((f) => f.name === 'maxIterations')

    expect(maxIterationsField?.validation?.min).toBeDefined()
    expect(maxIterationsField?.validation?.max).toBeDefined()
  })

  it('should have options for select fields', () => {
    const httpSchema = getNodeSchema('http')
    const methodField = httpSchema?.fields.find((f) => f.name === 'method')

    expect(methodField?.type).toBe('select')
    expect(methodField?.options).toBeDefined()
    expect(methodField?.options?.length).toBeGreaterThan(0)
  })
})

describe('Schema output configuration', () => {
  it('should have single output for action nodes', () => {
    const httpSchema = getNodeSchema('http')

    expect(httpSchema?.outputs).toBe(1)
  })

  it('should have multiple outputs for conditional node', () => {
    const conditionalSchema = getNodeSchema('conditional')

    expect(conditionalSchema?.outputs).toBe(2)
    expect(conditionalSchema?.outputLabels).toEqual(['True', 'False'])
  })

  it('should have multiple outputs for parallel node', () => {
    const parallelSchema = getNodeSchema('parallel')

    expect(parallelSchema?.outputs).toBe(3)
    expect(parallelSchema?.outputLabels?.length).toBe(3)
  })
})
