import { create } from 'zustand'
import { getAuthMe } from '@/api/generated/auth/auth'
import type { InternalHandlerUserResponse } from '@/api/generated/model'
import { AUTH_TOKEN_KEY } from '@/lib/constants'

export function getRoleHomePath(role?: string): string {
  switch (role) {
    case 'client':
      return '/catalog'
    case 'owner':
      return '/dashboard'
    case 'representative':
      return '/dashboard'
    case 'admin':
      return '/admin'
    default:
      return '/'
  }
}

export interface AuthState {
  user: InternalHandlerUserResponse | null
  token: string | null
  isLoading: boolean
  isAuthenticated: boolean
  // Bumped each time the API rejects with admin_2fa_required. A bridge
  // component watches this and renders a one-time toast; the persistent
  // banner reads `user.two_fa_method` directly.
  notice2faRequiredAt: number
  setAuth: (token: string, user: InternalHandlerUserResponse) => void
  setUser: (patch: Partial<InternalHandlerUserResponse>) => void
  logout: () => void
  loadProfile: () => Promise<void>
  bumpNotice2FARequired: () => void
}

const NOTICE_2FA_COOLDOWN_MS = 5_000

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  token: localStorage.getItem(AUTH_TOKEN_KEY),
  isLoading: !!localStorage.getItem(AUTH_TOKEN_KEY),
  isAuthenticated: false,
  notice2faRequiredAt: 0,

  setAuth: (token, user) => {
    localStorage.setItem(AUTH_TOKEN_KEY, token)
    set({ token, user, isAuthenticated: true, isLoading: false })
  },

  setUser: (patch) => {
    const current = get().user
    if (!current) return
    set({ user: { ...current, ...patch } })
  },

  logout: () => {
    localStorage.removeItem(AUTH_TOKEN_KEY)
    set({ token: null, user: null, isAuthenticated: false, isLoading: false, notice2faRequiredAt: 0 })
  },

  bumpNotice2FARequired: () => {
    const now = Date.now()
    if (now - get().notice2faRequiredAt < NOTICE_2FA_COOLDOWN_MS) return
    set({ notice2faRequiredAt: now })
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
