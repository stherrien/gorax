import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface User {
  id: string
  email: string
  name?: {
    first?: string
    last?: string
  }
  tenantId: string
}

interface AuthState {
  user: User | null
  sessionToken: string | null
  isAuthenticated: boolean
  isLoading: boolean

  setUser: (user: User | null) => void
  setSessionToken: (token: string | null) => void
  setLoading: (loading: boolean) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      sessionToken: null,
      isAuthenticated: false,
      isLoading: true,

      setUser: (user) => set({
        user,
        isAuthenticated: user !== null,
        isLoading: false,
      }),

      setSessionToken: (token) => set({ sessionToken: token }),

      setLoading: (loading) => set({ isLoading: loading }),

      logout: () => set({
        user: null,
        sessionToken: null,
        isAuthenticated: false,
        isLoading: false,
      }),
    }),
    {
      name: 'gorax-auth',
      partialize: (state) => ({
        sessionToken: state.sessionToken,
      }),
    }
  )
)
