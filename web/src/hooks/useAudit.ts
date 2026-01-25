import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { auditAPI } from '../api/audit'
import type {
  QueryFilter,
  TimeRange,
  ExportFormat,
  AuditEvent,
  AuditStats,
  QueryResponse,
} from '../types/audit'

/**
 * Hook to query audit events with pagination
 */
export function useAuditEvents(filter: QueryFilter, enabled = true) {
  return useQuery<QueryResponse, Error>({
    queryKey: ['audit-events', filter],
    queryFn: () => auditAPI.queryAuditEvents(filter),
    enabled,
    staleTime: 30000, // 30 seconds
    refetchInterval: 60000, // Auto-refetch every 60 seconds
  })
}

/**
 * Hook to get a single audit event by ID
 */
export function useAuditEvent(eventId: string, enabled = true) {
  return useQuery<AuditEvent, Error>({
    queryKey: ['audit-event', eventId],
    queryFn: () => auditAPI.getAuditEvent(eventId),
    enabled: enabled && !!eventId,
    staleTime: 300000, // 5 minutes (audit events don't change)
  })
}

/**
 * Hook to get audit statistics for a time range
 */
export function useAuditStats(timeRange: TimeRange, enabled = true) {
  return useQuery<AuditStats, Error>({
    queryKey: ['audit-stats', timeRange],
    queryFn: () => auditAPI.getAuditStats(timeRange),
    enabled,
    staleTime: 60000, // 1 minute
    refetchInterval: 120000, // Auto-refetch every 2 minutes
  })
}

/**
 * Hook to export audit events
 */
export function useExportAudit() {
  const queryClient = useQueryClient()

  return useMutation<
    Blob,
    Error,
    { filter: QueryFilter; format: ExportFormat; filename?: string }
  >({
    mutationFn: ({ filter, format }) => auditAPI.exportAuditEvents(filter, format),
    onSuccess: (blob, variables) => {
      // Create download link
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url

      // Set filename based on format and timestamp
      const timestamp = new Date().toISOString().split('T')[0]
      const ext = variables.format === 'csv' ? 'csv' : 'json'
      link.download = variables.filename || `audit-logs-${timestamp}.${ext}`

      document.body.appendChild(link)
      link.click()

      // Cleanup
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)

      // Invalidate queries to ensure fresh data after export
      queryClient.invalidateQueries({ queryKey: ['audit-events'] })
    },
  })
}

/**
 * Hook to get available audit categories
 */
export function useAuditCategories() {
  return {
    data: auditAPI.getCategories(),
    isLoading: false,
    error: null,
  }
}

/**
 * Hook to get available audit event types
 */
export function useAuditEventTypes() {
  return {
    data: auditAPI.getEventTypes(),
    isLoading: false,
    error: null,
  }
}

/**
 * Hook to get available severity levels
 */
export function useAuditSeverities() {
  return {
    data: auditAPI.getSeverities(),
    isLoading: false,
    error: null,
  }
}

/**
 * Hook to get available status values
 */
export function useAuditStatuses() {
  return {
    data: auditAPI.getStatuses(),
    isLoading: false,
    error: null,
  }
}

/**
 * Hook for real-time audit log updates (optional WebSocket integration)
 * This is a placeholder for future real-time functionality
 */
export function useAuditRealtime() {
  const queryClient = useQueryClient()

  // TODO: Implement WebSocket connection for real-time updates
  // For now, just return a function to manually refresh
  const refresh = () => {
    queryClient.invalidateQueries({ queryKey: ['audit-events'] })
    queryClient.invalidateQueries({ queryKey: ['audit-stats'] })
  }

  return {
    refresh,
    isConnected: false,
    lastUpdate: null,
  }
}
