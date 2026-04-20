import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import AppLayout from '@/components/AppLayout'
import { useAuthStore } from '@/stores/auth'
import { useBathhouseStore } from '@/stores/bathhouse'
import { useGetMyBathhouses } from '@/api/generated/bathhouses/bathhouses'

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetMyBathhouses: vi.fn().mockReturnValue({
    data: {
      data: [
        { id: 'bath-1', name: 'Баня на Речной' },
        { id: 'bath-2', name: 'Парная Люкс' },
      ],
    },
    isLoading: false,
  }),
  getMyBathhouses: vi.fn(),
}))

const defaultBathhousesMock = {
  data: {
    data: [
      { id: 'bath-1', name: 'Баня на Речной' },
      { id: 'bath-2', name: 'Парная Люкс' },
    ],
  },
  isLoading: false,
} as ReturnType<typeof useGetMyBathhouses>

function mockDesktopViewport() {
  window.matchMedia = vi.fn().mockImplementation((query: string) => ({
    matches: query.includes('min-width'),
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  }))
}

function renderLayout(route = '/') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <MemoryRouter initialEntries={[route]}>
          <AppLayout />
        </MemoryRouter>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

describe('AppLayout', () => {
  beforeEach(() => {
    mockDesktopViewport()
    vi.mocked(useGetMyBathhouses).mockReturnValue(defaultBathhousesMock)
    useAuthStore.setState({
      user: { id: '1', role: 'owner', name: 'Иван', email: 'ivan@test.com' },
      token: 'jwt-token',
      isAuthenticated: true,
      isLoading: false,
    })
    useBathhouseStore.setState({ selectedBathhouseId: null })
  })

  it('renders compact section trigger for owner area', () => {
    const { container } = renderLayout()
    const header = container.querySelector('.ant-layout-header')
    expect(header).toBeTruthy()

    expect(screen.getByText('RelaxHUB')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Раздел: Обзор' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Объекты' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Брони' })).not.toBeInTheDocument()
  })

  it('renders user name in header dropdown button', () => {
    const { container } = renderLayout()
    const header = container.querySelector('.ant-layout-header')
    expect(header).toBeTruthy()
    expect(header!.textContent).toContain('Иван')
  })

  it('does not render legacy breadcrumbs in the new shell', () => {
    renderLayout()
    expect(document.querySelector('.ant-breadcrumb')).toBeNull()
  })

  it('keeps the header navigation on nested route', () => {
    renderLayout('/bookings')
    expect(document.querySelector('.ant-breadcrumb')).toBeNull()
    expect(screen.getByRole('button', { name: 'Раздел: Брони' })).toBeInTheDocument()
  })

  it('renders BathhouseSelector with bathhouses', () => {
    renderLayout()
    expect(screen.getByText('Баня:')).toBeInTheDocument()
  })

  it('auto-selects first bathhouse when none selected', () => {
    renderLayout()
    const { selectedBathhouseId } = useBathhouseStore.getState()
    expect(selectedBathhouseId).toBe('bath-1')
  })

  it('shows "Нет бань" when no bathhouses available', () => {
    vi.mocked(useGetMyBathhouses).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as ReturnType<typeof useGetMyBathhouses>)

    renderLayout()
    expect(screen.getByText('Нет бань')).toBeInTheDocument()
  })

  it('falls back to email when user has no name', () => {
    useAuthStore.setState({
      user: { id: '1', role: 'owner', email: 'test@test.com' },
      token: 'jwt-token',
      isAuthenticated: true,
      isLoading: false,
    })
    const { container } = renderLayout()
    const header = container.querySelector('.ant-layout-header')
    expect(header!.textContent).toContain('test@test.com')
  })
})
