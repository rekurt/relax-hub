import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import App from '../App'
import { useAuthStore } from '@/stores/auth'

vi.mock('@/api/generated/auth/auth', () => ({
  getAuthMe: vi.fn().mockResolvedValue({ success: false }),
}))

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetMyBathhouses: vi.fn().mockReturnValue({
    data: { data: [] },
    isLoading: false,
  }),
  useGetBathhouses: vi.fn().mockReturnValue({
    data: { data: [], success: true },
    isLoading: false,
  }),
  getMyBathhouses: vi.fn().mockResolvedValue({ success: true, data: [] }),
}))

vi.mock('@/api/generated/cities/cities', () => ({
  useGetCities: vi.fn().mockReturnValue({
    data: { data: [] },
    isLoading: false,
  }),
}))

vi.mock('@/api/generated/favorites/favorites', () => ({
  usePostBathhousesIdFavorite: vi.fn().mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  }),
}))

vi.mock('@/api/generated/notifications/notifications', () => ({
  useGetMyNotificationsUnreadCount: vi.fn().mockReturnValue({
    data: { data: { unread_count: 0 } },
    isLoading: false,
  }),
  useGetMyNotifications: vi.fn().mockReturnValue({
    data: { data: [] },
    isLoading: false,
  }),
  usePatchMyNotificationsIdRead: vi.fn().mockReturnValue({ mutate: vi.fn(), isPending: false }),
  usePatchMyNotificationsReadAll: vi.fn().mockReturnValue({ mutate: vi.fn(), isPending: false }),
}))

vi.mock('@/api/generated/chat/chat', () => ({
  useGetMyUnreadMessagesCount: vi.fn().mockReturnValue({
    data: { data: { unread_count: 0 } },
    isLoading: false,
  }),
}))

vi.mock('@/api/generated/bookings/bookings', () => ({
  useGetBookings: vi.fn().mockReturnValue({
    data: { data: [], meta: { page: 1, page_size: 10, total_count: 0, total_pages: 0 } },
    isLoading: false,
  }),
  usePatchBookingsIdCancel: vi.fn().mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  }),
}))

function renderWithProviders(ui: React.ReactElement, { route = '/' } = {}) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <MemoryRouter initialEntries={[route]}>{ui}</MemoryRouter>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

describe('App', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: null,
      token: null,
      isLoading: false,
      isAuthenticated: false,
    })
  })

  it('renders public home on root route', () => {
    renderWithProviders(<App />, { route: '/' })
    expect(screen.getByText('Быстрое бронирование')).toBeInTheDocument()
  })

  it('redirects to login when opening protected route unauthenticated', () => {
    renderWithProviders(<App />, { route: '/dashboard' })
    expect(screen.getByText('Вход в личный кабинет')).toBeInTheDocument()
  })

  it('renders login page on /login route', () => {
    renderWithProviders(<App />, { route: '/login' })
    expect(screen.getByText('Вход в личный кабинет')).toBeInTheDocument()
  })

  it('renders register page on /register route', () => {
    renderWithProviders(<App />, { route: '/register' })
    expect(screen.getByText('Регистрация клиента')).toBeInTheDocument()
  })

  it('renders dashboard when authenticated as owner', () => {
    useAuthStore.setState({
      isAuthenticated: true,
      isLoading: false,
      user: { id: '1', role: 'owner', email: 'test@test.com' },
      token: 'jwt-token',
    })
    const { container } = renderWithProviders(<App />, { route: '/dashboard' })
    const mainContent = container.querySelector('.ant-layout-content')
    expect(mainContent).toBeTruthy()
    expect(mainContent!.textContent).toContain('Дашборд')
  })

  it('renders protected content when authenticated as representative', () => {
    useAuthStore.setState({
      isAuthenticated: true,
      isLoading: false,
      user: { id: '2', role: 'representative', email: 'rep@test.com' },
      token: 'jwt-token',
    })
    const { container } = renderWithProviders(<App />, { route: '/dashboard' })
    const mainContent = container.querySelector('.ant-layout-content')
    expect(mainContent!.textContent).toContain('Дашборд')
  })

  it('redirects client role from protected owner route to catalog', () => {
    useAuthStore.setState({
      isAuthenticated: true,
      isLoading: false,
      user: { id: '3', role: 'client', email: 'client@test.com' },
      token: 'jwt-token',
    })
    const { container } = renderWithProviders(<App />, { route: '/dashboard' })
    const mainContent = container.querySelector('.ant-layout-content')
    expect(mainContent!.textContent).toContain('Поиск бань')
  })

  it('shows spinner while loading auth', () => {
    useAuthStore.setState({ isLoading: true, token: 'jwt-token' })
    const { container } = renderWithProviders(<App />, { route: '/dashboard' })
    expect(container.querySelector('.ant-spin')).toBeTruthy()
  })
})
