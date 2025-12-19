import { create } from 'zustand'
import { credentialAPI } from '../api/credentials'
import type {
  Credential,
  CredentialCreateInput,
  CredentialUpdateInput,
  CredentialListParams,
  CredentialValue,
  CredentialTestResult,
} from '../api/credentials'

interface CredentialState {
  // State
  credentials: Credential[]
  selectedCredential: Credential | null
  loading: boolean
  error: string | null

  // Actions
  fetchCredentials: (params?: CredentialListParams) => Promise<void>
  fetchCredential: (id: string) => Promise<void>
  createCredential: (credential: CredentialCreateInput) => Promise<void>
  updateCredential: (id: string, updates: CredentialUpdateInput) => Promise<void>
  deleteCredential: (id: string) => Promise<void>
  rotateCredential: (id: string, value: CredentialValue) => Promise<void>
  testCredential: (id: string) => Promise<CredentialTestResult>
  selectCredential: (id: string | null) => void
  setCredentials: (credentials: Credential[]) => void
  clearError: () => void
  reset: () => void
}

export const useCredentialStore = create<CredentialState>((set, get) => ({
  // Initial state
  credentials: [],
  selectedCredential: null,
  loading: false,
  error: null,

  // Fetch all credentials
  fetchCredentials: async (params?: CredentialListParams) => {
    set({ loading: true, error: null })
    try {
      const response = await credentialAPI.list(params)
      set({ credentials: response.credentials, loading: false })
    } catch (error: any) {
      set({ error: error.message, loading: false })
    }
  },

  // Fetch single credential
  fetchCredential: async (id: string) => {
    set({ loading: true, error: null })
    try {
      const credential = await credentialAPI.get(id)
      set({ selectedCredential: credential, loading: false })
    } catch (error: any) {
      set({ error: error.message, loading: false })
    }
  },

  // Create new credential
  createCredential: async (credential: CredentialCreateInput) => {
    set({ loading: true, error: null })
    try {
      await credentialAPI.create(credential)
      // Refresh list after creation
      await get().fetchCredentials()
    } catch (error: any) {
      set({ error: error.message, loading: false })
    }
  },

  // Update credential
  updateCredential: async (id: string, updates: CredentialUpdateInput) => {
    set({ loading: true, error: null })
    try {
      await credentialAPI.update(id, updates)
      // Refresh list after update
      await get().fetchCredentials()
    } catch (error: any) {
      set({ error: error.message, loading: false })
    }
  },

  // Delete credential
  deleteCredential: async (id: string) => {
    set({ loading: true, error: null })
    try {
      await credentialAPI.delete(id)
      // Refresh list after deletion
      await get().fetchCredentials()
    } catch (error: any) {
      set({ error: error.message, loading: false })
    }
  },

  // Rotate credential value
  rotateCredential: async (id: string, value: CredentialValue) => {
    set({ loading: true, error: null })
    try {
      await credentialAPI.rotate(id, value)
      // Refresh list after rotation
      await get().fetchCredentials()
    } catch (error: any) {
      set({ error: error.message, loading: false })
    }
  },

  // Test credential
  testCredential: async (id: string) => {
    set({ loading: true, error: null })
    try {
      const result = await credentialAPI.test(id)
      set({ loading: false })
      return result
    } catch (error: any) {
      set({ error: error.message, loading: false })
      throw error
    }
  },

  // Select credential
  selectCredential: (id: string | null) => {
    if (id === null) {
      set({ selectedCredential: null })
      return
    }
    const credential = get().credentials.find((c) => c.id === id)
    set({ selectedCredential: credential || null })
  },

  // Set credentials directly
  setCredentials: (credentials: Credential[]) => {
    set({ credentials })
  },

  // Clear error
  clearError: () => {
    set({ error: null })
  },

  // Reset store
  reset: () => {
    set({
      credentials: [],
      selectedCredential: null,
      loading: false,
      error: null,
    })
  },
}))
