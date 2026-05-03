import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import AppRouter from '@/router'
import { useAuthStore } from '@/stores/auth'

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    get: vi.fn().mockResolvedValue({ data: { data: [], meta: null } }),
    post: vi.fn().mockResolvedValue({ data: { success: true } }),
    put: vi.fn().mockResolvedValue({ data: { success: true } }),
    delete: vi.fn().mockResolvedValue({ data: { success: true } }),
  },
}))

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
  useGetBathhouses: vi.fn().mockReturnValue({
    data: { data: [], success: true },
    isLoading: false,
  }),
  getMyBathhouses: vi.fn(),
}))

vi.mock('@/api/generated/chat/chat', () => ({
  useGetMyUnreadMessagesCount: vi.fn().mockReturnValue({
    data: { data: { unread_count: 0 } },
    isLoading: false,
  }),
}))

vi.mock('@/api/generated/favorites/favorites', () => ({
  usePostBathhousesIdFavorite: vi.fn().mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  }),
}))

function renderRouter(route: string) {
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

describe('PMSIntegration route', () => {
  beforeEach(() => {
    useAuthStore.setState({
      isAuthenticated: true,
      isLoading: false,
      user: { id: '1', role: 'owner', email: 'test@test.com' },
      token: 'jwt-token',
    })
  })

  it('renders PMSIntegration page at /settings/pms for owner', () => {
    const { container } = renderRouter('/settings/pms')
    const mainContent = container.querySelector('.ant-layout-content')
    expect(mainContent).toBeTruthy()
    expect(mainContent!.textContent).toContain('Интеграция с PMS')
  })

  it('renders PMS connection button', () => {
    renderRouter('/settings/pms')
    expect(screen.getByText('Подключить PMS')).toBeInTheDocument()
  })

  it('shows empty state when no connections exist', () => {
    renderRouter('/settings/pms')
    expect(screen.getByText(/Нет подключений к PMS/)).toBeInTheDocument()
  })
})
