import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import AppRouter from '@/router'
import { useAuthStore } from '@/stores/auth'

vi.mock('@/api/generated/auth/auth', () => ({
  getAuthMe: vi.fn().mockResolvedValue({ success: false }),
  usePutAuthMe: vi.fn().mockReturnValue({ mutate: vi.fn(), isPending: false }),
  usePostAuthMeAvatar: vi.fn().mockReturnValue({ mutate: vi.fn(), isPending: false }),
  useDeleteAuthMeAvatar: vi.fn().mockReturnValue({ mutate: vi.fn(), isPending: false }),
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
  useGetMyNotificationPreferences: vi.fn().mockReturnValue({
    data: { data: {} },
    isLoading: false,
  }),
  usePutMyNotificationPreferences: vi.fn().mockReturnValue({ mutate: vi.fn(), isPending: false }),
}))

vi.mock('@/api/generated/oauth/oauth', () => ({
  useGetAuthMeSocialAccounts: vi.fn().mockReturnValue({
    data: { data: [] },
    isLoading: false,
  }),
  useDeleteAuthLinkProvider: vi.fn().mockReturnValue({ mutate: vi.fn(), isPending: false }),
}))

vi.mock('@/api/generated/cities/cities', () => ({
  useGetCities: vi.fn().mockReturnValue({
    data: { data: [] },
    isLoading: false,
  }),
}))

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetMyBathhouses: vi.fn().mockReturnValue({
    data: { data: [{ id: 'b1', name: 'Тестовая баня' }] },
    isLoading: false,
  }),
  getMyBathhouses: vi.fn(),
}))

function renderRouter(route = '/') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <MemoryRouter initialEntries={[route]}>
          <AppRouter />
        </MemoryRouter>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

function setAuth() {
  useAuthStore.setState({
    isAuthenticated: true,
    isLoading: false,
    user: { id: '1', role: 'owner', email: 'test@test.com' },
    token: 'jwt-token',
  })
}

describe('AppRouter', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: null,
      token: null,
      isLoading: false,
      isAuthenticated: false,
    })
  })

  it('renders login at /login', () => {
    renderRouter('/login')
    expect(screen.getByText('Вход в личный кабинет')).toBeInTheDocument()
  })

  it('renders register at /register', () => {
    renderRouter('/register')
    expect(screen.getByText('Регистрация владельца')).toBeInTheDocument()
  })

  it('renders dashboard at / when authenticated', () => {
    setAuth()
    const { container } = renderRouter('/')
    // "Дашборд" appears in both sidebar menu and page content
    const mainContent = container.querySelector('.ant-layout-content')
    expect(mainContent).toBeTruthy()
    expect(mainContent!.textContent).toContain('Дашборд')
  })

  it('renders bookings page at /bookings when authenticated', () => {
    setAuth()
    const { container } = renderRouter('/bookings')
    const mainContent = container.querySelector('.ant-layout-content')
    expect(mainContent!.textContent).toContain('Бронирования')
  })

  it('renders reviews page at /reviews when authenticated', () => {
    setAuth()
    const { container } = renderRouter('/reviews')
    const mainContent = container.querySelector('.ant-layout-content')
    expect(mainContent!.textContent).toContain('Отзывы')
  })

  it('renders pricing page at /pricing when authenticated', () => {
    setAuth()
    const { container } = renderRouter('/pricing')
    const mainContent = container.querySelector('.ant-layout-content')
    expect(mainContent!.textContent).toContain('Цены')
  })

  it('renders settings page at /settings when authenticated', () => {
    setAuth()
    const { container } = renderRouter('/settings')
    const mainContent = container.querySelector('.ant-layout-content')
    expect(mainContent!.textContent).toContain('Настройки')
  })

  it('redirects unknown routes to / (then to login)', () => {
    renderRouter('/nonexistent-route')
    expect(screen.getByText('Вход в личный кабинет')).toBeInTheDocument()
  })

  it('redirects to login for protected routes when not authenticated', () => {
    renderRouter('/bookings')
    expect(screen.getByText('Вход в личный кабинет')).toBeInTheDocument()
  })
})
