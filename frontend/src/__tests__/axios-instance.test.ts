import { axiosInstance } from '../api/axios-instance'
import { useAuthStore } from '@/stores/auth'
import { AUTH_TOKEN_KEY } from '@/lib/constants'
import { isPublicPathname } from '@/lib/public-route'

describe('axiosInstance', () => {
  beforeEach(() => {
    localStorage.clear()
    useAuthStore.setState({
      user: null,
      token: null,
      isLoading: false,
      isAuthenticated: false,
    })
    window.history.replaceState({}, '', '/')
  })

  it('has correct baseURL', () => {
    expect(axiosInstance.defaults.baseURL).toBe('/api/v1')
  })

  it('attaches Authorization header when token exists', () => {
    localStorage.setItem('bani_token', 'test-jwt-token')

    type InterceptorHandler = {
      fulfilled: (config: { headers: Record<string, string> }) => { headers: Record<string, string> }
    }
    const handlers = (
      axiosInstance.interceptors.request as unknown as { handlers: InterceptorHandler[] }
    ).handlers
    const config = handlers[0]!.fulfilled({ headers: {} as Record<string, string> })

    expect(config.headers.Authorization).toBe('Bearer test-jwt-token')

    localStorage.removeItem('bani_token')
  })

  it('does not attach Authorization header when no token', () => {
    localStorage.removeItem('bani_token')

    type InterceptorHandler = {
      fulfilled: (config: { headers: Record<string, string> }) => { headers: Record<string, string> }
    }
    const handlers = (
      axiosInstance.interceptors.request as unknown as { handlers: InterceptorHandler[] }
    ).handlers
    const config = handlers[0]!.fulfilled({ headers: {} as Record<string, string> })

    expect(config.headers.Authorization).toBeUndefined()
  })

  it('classifies public paths explicitly', () => {
    expect(isPublicPathname('/')).toBe(true)
    expect(isPublicPathname('/catalog')).toBe(true)
    expect(isPublicPathname('/bathhouses/par-master')).toBe(true)
    expect(isPublicPathname('/checkout')).toBe(true)
    expect(isPublicPathname('/client/bookings')).toBe(false)
    expect(isPublicPathname('/dashboard')).toBe(false)
  })

  it('logs out on 401 but keeps guest on public routes', async () => {
    localStorage.setItem(AUTH_TOKEN_KEY, 'expired-token')
    useAuthStore.getState().setAuth('expired-token', {
      id: 'u1',
      role: 'client',
      email: 'client@example.com',
    })
    window.history.replaceState({}, '', '/catalog')

    type ResponseInterceptorHandler = {
      rejected: (error: { response?: { status?: number } }) => Promise<never>
    }
    const handlers = (
      axiosInstance.interceptors.response as unknown as { handlers: ResponseInterceptorHandler[] }
    ).handlers

    await expect(handlers[0]!.rejected({ response: { status: 401 } })).rejects.toEqual({ response: { status: 401 } })

    expect(useAuthStore.getState().isAuthenticated).toBe(false)
    expect(window.location.pathname).toBe('/catalog')
  })

  it('redirects to login on 401 for protected routes', async () => {
    localStorage.setItem(AUTH_TOKEN_KEY, 'expired-token')
    useAuthStore.getState().setAuth('expired-token', {
      id: 'u2',
      role: 'owner',
      email: 'owner@example.com',
    })
    window.history.replaceState({}, '', '/dashboard')

    type ResponseInterceptorHandler = {
      rejected: (error: { response?: { status?: number } }) => Promise<never>
    }
    const handlers = (
      axiosInstance.interceptors.response as unknown as { handlers: ResponseInterceptorHandler[] }
    ).handlers

    await expect(handlers[0]!.rejected({ response: { status: 401 } })).rejects.toEqual({ response: { status: 401 } })

    expect(useAuthStore.getState().isAuthenticated).toBe(false)
    expect(window.location.pathname).toBe('/login')
  })
})
