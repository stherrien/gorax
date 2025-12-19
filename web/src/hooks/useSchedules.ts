import { useState, useEffect, useCallback } from 'react'
import { scheduleAPI } from '../api/schedules'
import type {
  Schedule,
  ScheduleListParams,
  ScheduleCreateInput,
  ScheduleUpdateInput,
} from '../api/schedules'

/**
 * Hook to fetch and manage list of schedules
 */
export function useSchedules(params?: ScheduleListParams) {
  const [schedules, setSchedules] = useState<Schedule[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchSchedules = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await scheduleAPI.list(params)
      setSchedules(response.schedules)
      setTotal(response.total)
    } catch (err) {
      setError(err as Error)
      setSchedules([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [params])

  useEffect(() => {
    fetchSchedules()
  }, [fetchSchedules])

  return {
    schedules,
    total,
    loading,
    error,
    refetch: fetchSchedules,
  }
}

/**
 * Hook to fetch a single schedule by ID
 */
export function useSchedule(id: string | null) {
  const [schedule, setSchedule] = useState<Schedule | null>(null)
  const [loading, setLoading] = useState(!!id)
  const [error, setError] = useState<Error | null>(null)

  const fetchSchedule = useCallback(async () => {
    if (!id) {
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const data = await scheduleAPI.get(id)
      setSchedule(data)
    } catch (err) {
      setError(err as Error)
      setSchedule(null)
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    fetchSchedule()
  }, [fetchSchedule])

  return {
    schedule,
    loading,
    error,
    refetch: fetchSchedule,
  }
}

/**
 * Hook for schedule CRUD mutations
 */
export function useScheduleMutations() {
  const [creating, setCreating] = useState(false)
  const [updating, setUpdating] = useState(false)
  const [deleting, setDeleting] = useState(false)

  const createSchedule = async (
    workflowId: string,
    input: ScheduleCreateInput
  ): Promise<Schedule> => {
    try {
      setCreating(true)
      const schedule = await scheduleAPI.create(workflowId, input)
      return schedule
    } finally {
      setCreating(false)
    }
  }

  const updateSchedule = async (
    id: string,
    updates: ScheduleUpdateInput
  ): Promise<Schedule> => {
    try {
      setUpdating(true)
      const schedule = await scheduleAPI.update(id, updates)
      return schedule
    } finally {
      setUpdating(false)
    }
  }

  const deleteSchedule = async (id: string): Promise<void> => {
    try {
      setDeleting(true)
      await scheduleAPI.delete(id)
    } finally {
      setDeleting(false)
    }
  }

  const toggleSchedule = async (id: string, enabled: boolean): Promise<Schedule> => {
    try {
      setUpdating(true)
      const schedule = await scheduleAPI.toggle(id, enabled)
      return schedule
    } finally {
      setUpdating(false)
    }
  }

  return {
    createSchedule,
    updateSchedule,
    deleteSchedule,
    toggleSchedule,
    creating,
    updating,
    deleting,
  }
}
