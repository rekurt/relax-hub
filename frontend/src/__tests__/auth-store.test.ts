import { useAuthStore } from '@/stores/auth'

const mockGetAuthMe = vi.fn()
vi.mock('@/api/generated/auth/auth', () => ({
  getAuthMe: (...args: unknown[]) => mockGetAuthMe(...args),
}))

const mockLocalStorage = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value }),
    removeItem: vi.fn((key: string) => { delete store[key] }),
    clear: () => { store = {} },
  }
})()

Object.defineProperty(window, 'localStorage', { value: mockLocalStorage })

describe('useAuthStore', () => {
  beforeEach(() => {
    mockLocalStorage.clear()
    mockGetAuthMe.mockReset()
    useAuthStore.setState({
      user: null,
      token: null,
      isLoading: false,
      isAuthenticated: false,
    })
  })

  it('initial state is unauthenticated', () => {
    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(false)
    expect(state.user).toBeNull()
    expect(state.token).toBeNull()
  })

  it('setAuth stores token and user', () => {
    const user = { id: '1', email: 'test@test.com', name: 'Test', role: 'owner' }
    useAuthStore.getState().setAuth('jwt-token', user)

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(true)
    expect(state.token).toBe('jwt-token')
    expect(state.user).toEqual(user)
    expect(mockLocalStorage.setItem).toHaveBeenCalledWith('bani_token', 'jwt-token')
  })

  it('logout clears state and localStorage', () => {
    const user = { id: '1', email: 'test@test.com', role: 'owner' }
    useAuthStore.getState().setAuth('jwt-token', user)
    useAuthStore.getState().logout()

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(false)
    expect(state.token).toBeNull()
    expect(state.user).toBeNull()
    expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('bani_token')
  })

  it('loadProfile fetches user when token exists', async () => {
    const user = { id: '1', email: 'test@test.com', name: 'Test', role: 'owner' }
    mockGetAuthMe.mockResolvedValue({ success: true, data: user })
    useAuthStore.setState({ token: 'jwt-token', isLoading: true })

    await useAuthStore.getState().loadProfile()

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(true)
    expect(state.user).toEqual(user)
    expect(state.isLoading).toBe(false)
  })

  it('loadProfile logs out on failed response', async () => {
    mockGetAuthMe.mockResolvedValue({ success: false })
    useAuthStore.setState({ token: 'bad-token', isLoading: true })

    await useAuthStore.getState().loadProfile()

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(false)
    expect(state.user).toBeNull()
  })

  it('loadProfile logs out on network error', async () => {
    mockGetAuthMe.mockRejectedValue(new Error('Network error'))
    useAuthStore.setState({ token: 'bad-token', isLoading: true })

    await useAuthStore.getState().loadProfile()

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(false)
  })

  it('loadProfile does nothing when no token', async () => {
    useAuthStore.setState({ token: null, isLoading: true })

    await useAuthStore.getState().loadProfile()

    expect(mockGetAuthMe).not.toHaveBeenCalled()
    expect(useAuthStore.getState().isLoading).toBe(false)
  })
})
