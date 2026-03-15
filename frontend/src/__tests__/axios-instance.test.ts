import { axiosInstance } from '../api/axios-instance'

describe('axiosInstance', () => {
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
})
