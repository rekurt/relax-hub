import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi } from 'vitest'
import AdminDashboard from '@/pages/admin/AdminDashboard'

vi.mock('@/api/generated/admin-analytics/admin-analytics', () => ({
  useGetAdminAnalytics: vi.fn(),
  useGetAdminAnalyticsTop: vi.fn(),
}))

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    get: vi.fn(),
  },
}))

import { useGetAdminAnalytics, useGetAdminAnalyticsTop } from '@/api/generated/admin-analytics/admin-analytics'
import { axiosInstance } from '@/api/axios-instance'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin']}>
            <Routes>
              <Route path="/admin" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockDashboard = {
  total_bookings: 1250,
  total_revenue: 5000000,
  total_users: 340,
  total_bathhouses: 45,
  total_views: 12300,
  avg_rating: 4.6,
  new_users: 28,
  dau: 85,
  wau: 210,
  mau: 340,
}

const mockTopBathhouses = [
  {
    bathhouse_id: 'b-1',
    name: 'Баня Люкс',
    views: 1200,
    bookings: 85,
    revenue: 850000,
    rating: 4.8,
  },
  {
    bathhouse_id: 'b-2',
    name: 'Парная на Неве',
    views: 980,
    bookings: 62,
    revenue: 620000,
    rating: 4.5,
  },
]

const mockSupportMetrics = {
  fcr_percent: 75.2,
  aht_seconds: 1800,
  sla_compliance_percent: 92.5,
  queue_size: 8,
}

function setupMocks(opts?: { withSupportMetrics?: boolean }) {
  vi.mocked(axiosInstance.get).mockImplementation((url: string) => {
    if (url === '/admin/tickets/metrics' && opts?.withSupportMetrics !== false) {
      return Promise.resolve({ data: { data: mockSupportMetrics } })
    }
    return Promise.reject(new Error('not found'))
  })
}

describe('AdminDashboard', () => {
  it('renders KPI cards with analytics data', () => {
    vi.mocked(useGetAdminAnalytics).mockReturnValue({
      data: { data: mockDashboard, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalytics>)
    vi.mocked(useGetAdminAnalyticsTop).mockReturnValue({
      data: { data: { bathhouses: mockTopBathhouses, metric: 'bookings', limit: 10 }, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalyticsTop>)

    renderWithProviders(<AdminDashboard />)

    expect(screen.getByText('Панель администратора')).toBeInTheDocument()
    expect(screen.getByText('1,250')).toBeInTheDocument()
    expect(screen.getByText('340')).toBeInTheDocument()
    expect(screen.getByText('45')).toBeInTheDocument()
  })

  it('renders top bathhouses table', () => {
    vi.mocked(useGetAdminAnalytics).mockReturnValue({
      data: { data: mockDashboard, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalytics>)
    vi.mocked(useGetAdminAnalyticsTop).mockReturnValue({
      data: { data: { bathhouses: mockTopBathhouses, metric: 'bookings', limit: 10 }, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalyticsTop>)

    renderWithProviders(<AdminDashboard />)

    expect(screen.getByText('Топ бань')).toBeInTheDocument()
    expect(screen.getByText('Баня Люкс')).toBeInTheDocument()
    expect(screen.getByText('Парная на Неве')).toBeInTheDocument()
  })

  it('renders period selector', () => {
    vi.mocked(useGetAdminAnalytics).mockReturnValue({
      data: { data: mockDashboard, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalytics>)
    vi.mocked(useGetAdminAnalyticsTop).mockReturnValue({
      data: { data: { bathhouses: [], metric: 'bookings', limit: 10 }, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalyticsTop>)

    renderWithProviders(<AdminDashboard />)

    expect(screen.getByText('День')).toBeInTheDocument()
    expect(screen.getByText('Неделя')).toBeInTheDocument()
    expect(screen.getByText('Месяц')).toBeInTheDocument()
    expect(screen.getByText('3 месяца')).toBeInTheDocument()
  })

  it('renders metric selector for top bathhouses', () => {
    vi.mocked(useGetAdminAnalytics).mockReturnValue({
      data: { data: mockDashboard, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalytics>)
    vi.mocked(useGetAdminAnalyticsTop).mockReturnValue({
      data: { data: { bathhouses: [], metric: 'bookings', limit: 10 }, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalyticsTop>)

    renderWithProviders(<AdminDashboard />)

    expect(screen.getAllByText('Просмотры').length).toBeGreaterThanOrEqual(2)
    expect(screen.getAllByText('Выручка').length).toBeGreaterThanOrEqual(2)
    expect(screen.getAllByText('Рейтинг').length).toBeGreaterThanOrEqual(1)
  })

  it('renders empty state for top bathhouses', () => {
    vi.mocked(useGetAdminAnalytics).mockReturnValue({
      data: { data: mockDashboard, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalytics>)
    vi.mocked(useGetAdminAnalyticsTop).mockReturnValue({
      data: { data: { bathhouses: [], metric: 'bookings', limit: 10 }, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalyticsTop>)

    renderWithProviders(<AdminDashboard />)

    expect(screen.getByText('Нет данных')).toBeInTheDocument()
  })

  it('renders revenue formatted from kopecks', () => {
    vi.mocked(useGetAdminAnalytics).mockReturnValue({
      data: { data: mockDashboard, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalytics>)
    vi.mocked(useGetAdminAnalyticsTop).mockReturnValue({
      data: { data: { bathhouses: mockTopBathhouses, metric: 'bookings', limit: 10 }, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalyticsTop>)

    renderWithProviders(<AdminDashboard />)

    // 5000000 kopecks = 50000 rubles displayed via Statistic
    expect(screen.getByText('50,000')).toBeInTheDocument()
  })

  it('renders active users metrics (DAU/WAU/MAU)', () => {
    vi.mocked(useGetAdminAnalytics).mockReturnValue({
      data: { data: mockDashboard, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalytics>)
    vi.mocked(useGetAdminAnalyticsTop).mockReturnValue({
      data: { data: { bathhouses: [], metric: 'bookings', limit: 10 }, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalyticsTop>)

    renderWithProviders(<AdminDashboard />)

    expect(screen.getByText('DAU')).toBeInTheDocument()
    expect(screen.getByText(/WAU: 210/)).toBeInTheDocument()
    expect(screen.getByText(/MAU: 340/)).toBeInTheDocument()
  })

  it('renders support metrics widget when data is available', async () => {
    setupMocks({ withSupportMetrics: true })
    vi.mocked(useGetAdminAnalytics).mockReturnValue({
      data: { data: mockDashboard, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalytics>)
    vi.mocked(useGetAdminAnalyticsTop).mockReturnValue({
      data: { data: { bathhouses: [], metric: 'bookings', limit: 10 }, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAnalyticsTop>)

    renderWithProviders(<AdminDashboard />)

    // The widget renders async via useQuery, so check for the title text
    expect(await screen.findByText('Поддержка')).toBeInTheDocument()
    expect(screen.getByText('FCR')).toBeInTheDocument()
    expect(screen.getByText('AHT')).toBeInTheDocument()
    expect(screen.getByText('SLA (24ч)')).toBeInTheDocument()
    expect(screen.getByText('В очереди')).toBeInTheDocument()
  })
})
