import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
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
  getMyBathhouses: vi.fn().mockResolvedValue({ success: true, data: [] }),
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

  it('redirects to login when not authenticated', () => {
    renderWithProviders(<App />, { route: '/' })
    expect(screen.getByText('Вход в личный кабинет')).toBeInTheDocument()
  })

  it('renders login page on /login route', () => {
    renderWithProviders(<App />, { route: '/login' })
    expect(screen.getByText('Вход в личный кабинет')).toBeInTheDocument()
  })

  it('renders register page on /register route', () => {
    renderWithProviders(<App />, { route: '/register' })
    expect(screen.getByText('Регистрация владельца')).toBeInTheDocument()
  })

  it('renders dashboard when authenticated as owner', () => {
    useAuthStore.setState({
      isAuthenticated: true,
      isLoading: false,
      user: { id: '1', role: 'owner', email: 'test@test.com' },
      token: 'jwt-token',
    })
    const { container } = renderWithProviders(<App />, { route: '/' })
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
    const { container } = renderWithProviders(<App />, { route: '/' })
    const mainContent = container.querySelector('.ant-layout-content')
    expect(mainContent!.textContent).toContain('Дашборд')
  })

  it('redirects client role to login', () => {
    useAuthStore.setState({
      isAuthenticated: true,
      isLoading: false,
      user: { id: '3', role: 'client', email: 'client@test.com' },
      token: 'jwt-token',
    })
    renderWithProviders(<App />, { route: '/' })
    expect(screen.getByText('Вход в личный кабинет')).toBeInTheDocument()
  })

  it('shows spinner while loading auth', () => {
    useAuthStore.setState({ isLoading: true, token: 'jwt-token' })
    const { container } = renderWithProviders(<App />, { route: '/' })
    expect(container.querySelector('.ant-spin')).toBeTruthy()
  })
})
