import { useState, useEffect, useCallback } from 'react'
import { templateAPI } from '../api/templates'
import type {
  Template,
  TemplateListParams,
  CreateTemplateInput,
  UpdateTemplateInput,
  CreateFromWorkflowInput,
  InstantiateTemplateInput,
  InstantiateTemplateResult,
} from '../api/templates'

/**
 * Hook to fetch and manage list of templates
 */
export function useTemplates(params?: TemplateListParams) {
  const [templates, setTemplates] = useState<Template[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchTemplates = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await templateAPI.list(params)
      setTemplates(data)
    } catch (err) {
      setError(err as Error)
      setTemplates([])
    } finally {
      setLoading(false)
    }
  }, [params])

  useEffect(() => {
    fetchTemplates()
  }, [fetchTemplates])

  return {
    templates,
    loading,
    error,
    refetch: fetchTemplates,
  }
}

/**
 * Hook to fetch a single template by ID
 */
export function useTemplate(id: string | null) {
  const [template, setTemplate] = useState<Template | null>(null)
  const [loading, setLoading] = useState(!!id)
  const [error, setError] = useState<Error | null>(null)

  const fetchTemplate = useCallback(async () => {
    if (!id) {
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const data = await templateAPI.get(id)
      setTemplate(data)
    } catch (err) {
      setError(err as Error)
      setTemplate(null)
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    fetchTemplate()
  }, [fetchTemplate])

  return {
    template,
    loading,
    error,
    refetch: fetchTemplate,
  }
}

/**
 * Hook for template CRUD mutations
 */
export function useTemplateMutations() {
  const [creating, setCreating] = useState(false)
  const [updating, setUpdating] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [instantiating, setInstantiating] = useState(false)

  const createTemplate = async (input: CreateTemplateInput): Promise<Template> => {
    try {
      setCreating(true)
      const template = await templateAPI.create(input)
      return template
    } finally {
      setCreating(false)
    }
  }

  const updateTemplate = async (
    id: string,
    updates: UpdateTemplateInput
  ): Promise<void> => {
    try {
      setUpdating(true)
      await templateAPI.update(id, updates)
    } finally {
      setUpdating(false)
    }
  }

  const deleteTemplate = async (id: string): Promise<void> => {
    try {
      setDeleting(true)
      await templateAPI.delete(id)
    } finally {
      setDeleting(false)
    }
  }

  const createFromWorkflow = async (
    workflowId: string,
    input: CreateFromWorkflowInput
  ): Promise<Template> => {
    try {
      setCreating(true)
      const template = await templateAPI.createFromWorkflow(workflowId, input)
      return template
    } finally {
      setCreating(false)
    }
  }

  const instantiateTemplate = async (
    templateId: string,
    input: InstantiateTemplateInput
  ): Promise<InstantiateTemplateResult> => {
    try {
      setInstantiating(true)
      const result = await templateAPI.instantiate(templateId, input)
      return result
    } finally {
      setInstantiating(false)
    }
  }

  return {
    createTemplate,
    updateTemplate,
    deleteTemplate,
    createFromWorkflow,
    instantiateTemplate,
    creating,
    updating,
    deleting,
    instantiating,
  }
}
