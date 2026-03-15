import { create } from 'zustand'
import { getAuthMe } from '@/api/generated/auth/auth'
import type { InternalHandlerUserResponse } from '@/api/generated/model'
import { AUTH_TOKEN_KEY } from '@/lib/constants'

export interface AuthState {
  user: InternalHandlerUserResponse | null
  token: string | null
  isLoading: boolean
  isAuthenticated: boolean
  setAuth: (token: string, user: InternalHandlerUserResponse) => void
  logout: () => void
  loadProfile: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  token: localStorage.getItem(AUTH_TOKEN_KEY),
  isLoading: !!localStorage.getItem(AUTH_TOKEN_KEY),
  isAuthenticated: false,

  setAuth: (token, user) => {
    localStorage.setItem(AUTH_TOKEN_KEY, token)
    set({ token, user, isAuthenticated: true, isLoading: false })
  },

  logout: () => {
    localStorage.removeItem(AUTH_TOKEN_KEY)
    set({ token: null, user: null, isAuthenticated: false, isLoading: false })
  },

  loadProfile: async () => {
    const { token } = get()
    if (!token) {
      set({ isLoading: false })
      return
    }
    try {
      const response = await getAuthMe()
      if (response.success && response.data) {
        set({
          user: response.data,
          isAuthenticated: true,
          isLoading: false,
        })
      } else {
        get().logout()
      }
    } catch {
      get().logout()
    }
  },
}))
