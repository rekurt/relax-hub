import { renderHook, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useDeviceToken } from '@/lib/useDeviceToken'

const mockMutateAsync = vi.fn()
const mockDeleteMutateAsync = vi.fn()

vi.mock('@/api/generated/device-tokens/device-tokens', () => ({
  usePostDeviceTokens: () => ({ mutateAsync: mockMutateAsync }),
  useDeleteDeviceTokensId: () => ({ mutateAsync: mockDeleteMutateAsync }),
}))

let mockIsAuthenticated = false
vi.mock('@/stores/auth', () => ({
  useAuthStore: (selector: (s: Record<string, unknown>) => unknown) =>
    selector({ isAuthenticated: mockIsAuthenticated }),
}))

function wrapper({ children }: { children: React.ReactNode }) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
}

describe('useDeviceToken', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    mockIsAuthenticated = false
  })

  it('registerToken calls API and stores token ID', async () => {
    mockMutateAsync.mockResolvedValue({ data: { id: 'device-token-123' } })

    const { result } = renderHook(() => useDeviceToken(), { wrapper })

    await act(async () => {
      await result.current.registerToken('fcm-token-abc')
    })

    expect(mockMutateAsync).toHaveBeenCalledWith({
      data: { token: 'fcm-token-abc', platform: 'web' },
    })
    expect(localStorage.getItem('rh_device_token_id')).toBe('device-token-123')
  })

  it('unregisterToken calls delete API and clears storage', async () => {
    localStorage.setItem('rh_device_token_id', 'device-token-456')
    mockDeleteMutateAsync.mockResolvedValue({})

    const { result } = renderHook(() => useDeviceToken(), { wrapper })

    await act(async () => {
      await result.current.unregisterToken()
    })

    expect(mockDeleteMutateAsync).toHaveBeenCalledWith({ id: 'device-token-456' })
    expect(localStorage.getItem('rh_device_token_id')).toBeNull()
  })

  it('unregisterToken does nothing if no token in storage', async () => {
    const { result } = renderHook(() => useDeviceToken(), { wrapper })

    await act(async () => {
      await result.current.unregisterToken()
    })

    expect(mockDeleteMutateAsync).not.toHaveBeenCalled()
  })

  it('registerToken handles API failure gracefully', async () => {
    mockMutateAsync.mockRejectedValue(new Error('Network error'))

    const { result } = renderHook(() => useDeviceToken(), { wrapper })

    await act(async () => {
      await result.current.registerToken('fcm-token-xyz')
    })

    expect(localStorage.getItem('rh_device_token_id')).toBeNull()
  })

  it('unregisterToken clears storage even on API failure', async () => {
    localStorage.setItem('rh_device_token_id', 'device-token-789')
    mockDeleteMutateAsync.mockRejectedValue(new Error('Network error'))

    const { result } = renderHook(() => useDeviceToken(), { wrapper })

    await act(async () => {
      await result.current.unregisterToken()
    })

    expect(localStorage.getItem('rh_device_token_id')).toBeNull()
  })
})
