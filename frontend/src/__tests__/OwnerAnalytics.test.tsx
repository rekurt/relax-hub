import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import OwnerAnalytics from '@/pages/analytics/OwnerAnalytics'
import { useBathhouseStore } from '@/stores/bathhouse'
import {
  useGetMyBathhousesIdAnalytics,
  useGetMyBathhousesIdAnalyticsDaily,
  useGetMyBathhousesIdAnalyticsPerformance,
} from '@/api/generated/analytics/analytics'

vi.mock('@/api/generated/analytics/analytics', () => ({
  useGetMyBathhousesIdAnalytics: vi.fn(),
  useGetMyBathhousesIdAnalyticsDaily: vi.fn(),
  useGetMyBathhousesIdAnalyticsPerformance: vi.fn(),
}))

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <OwnerAnalytics />
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockDashboard = {
  data: {
    data: {
      bookings: 25,
      bookings_change: 10.0,
      revenue: 750000,
      revenue_change: -5.2,
      views: 800,
      views_change: 3.0,
      unique_views: 450,
      rating: 4.5,
      rating_change: 1.0,
      avg_check: 300000,
      conversion_rate: 0.031,
      period: '30d',
    },
  },
  isLoading: false,
  isError: false,
  error: null,
}

const mockDaily = {
  data: {
    data: [
      { date: '2026-03-29', bookings: 3, revenue: 90000, views: 50, unique_views: 30, avg_rating: 4.5 },
      { date: '2026-03-28', bookings: 5, revenue: 150000, views: 80, unique_views: 45, avg_rating: 4.8 },
    ],
  },
  isLoading: false,
  isError: false,
  error: null,
}

const mockPerformance = {
  data: {
    data: {
      bathhouse_id: 'bath-1',
      bathhouse_name: 'Тест баня',
      occupancy_rate: 0.65,
      avg_city_occupancy_rate: 0.55,
      conversion_rate: 0.04,
      avg_city_conversion_rate: 0.03,
      avg_rating: 4.5,
      avg_city_rating: 4.2,
      revenue: 750000,
    },
  },
  isLoading: false,
  isError: false,
  error: null,
}

describe('OwnerAnalytics', () => {
  beforeEach(() => {
    useBathhouseStore.setState({ selectedBathhouseId: 'bath-1' })
    vi.mocked(useGetMyBathhousesIdAnalytics).mockReturnValue(
      mockDashboard as ReturnType<typeof useGetMyBathhousesIdAnalytics>,
    )
    vi.mocked(useGetMyBathhousesIdAnalyticsDaily).mockReturnValue(
      mockDaily as ReturnType<typeof useGetMyBathhousesIdAnalyticsDaily>,
    )
    vi.mocked(useGetMyBathhousesIdAnalyticsPerformance).mockReturnValue(
      mockPerformance as ReturnType<typeof useGetMyBathhousesIdAnalyticsPerformance>,
    )
  })

  it('shows prompt when no bathhouse selected', () => {
    useBathhouseStore.setState({ selectedBathhouseId: null })
    renderPage()
    expect(screen.getByText('Выберите баню')).toBeInTheDocument()
  })

  it('renders page title', () => {
    renderPage()
    expect(screen.getByText('Аналитика')).toBeInTheDocument()
  })

  it('renders KPI cards with data', () => {
    const { container } = renderPage()
    const text = container.textContent ?? ''
    expect(screen.getAllByText('Бронирования').length).toBeGreaterThan(0)
    expect(text).toContain('25')
    expect(screen.getAllByText('Выручка').length).toBeGreaterThan(0)
    expect(text).toContain('7500 ₽')
    expect(screen.getAllByText('Просмотры').length).toBeGreaterThan(0)
    expect(text).toContain('800')
    expect(screen.getAllByText('Конверсия').length).toBeGreaterThan(0)
    expect(text).toContain('3.1%')
    expect(screen.getAllByText('Рейтинг').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Средний чек').length).toBeGreaterThan(0)
    expect(text).toContain('3000 ₽')
  })

  it('renders period selector with options', () => {
    renderPage()
    expect(screen.getByText('7 дней')).toBeInTheDocument()
    expect(screen.getByText('30 дней')).toBeInTheDocument()
    expect(screen.getByText('90 дней')).toBeInTheDocument()
  })

  it('switches period on click', () => {
    renderPage()
    fireEvent.click(screen.getByText('7 дней'))
    expect(useGetMyBathhousesIdAnalytics).toHaveBeenCalledWith(
      'bath-1',
      { period: '7d' },
      { query: { enabled: true } },
    )
  })

  it('renders competitor comparison section', () => {
    const { container } = renderPage()
    const text = container.textContent ?? ''
    expect(screen.getByText('Сравнение с конкурентами')).toBeInTheDocument()
    expect(screen.getByText('Ваша загрузка')).toBeInTheDocument()
    expect(text).toContain('65.0%')
    expect(text).toContain('55.0%')
    expect(screen.getByText('Ваша конверсия')).toBeInTheDocument()
    expect(text).toContain('4.0%')
    expect(text).toContain('3.0%')
    expect(screen.getByText('Ваш рейтинг')).toBeInTheDocument()
  })

  it('renders daily breakdown table', () => {
    renderPage()
    expect(screen.getByText('Динамика по дням')).toBeInTheDocument()
    expect(screen.getByText('29.03.2026')).toBeInTheDocument()
    expect(screen.getByText('28.03.2026')).toBeInTheDocument()
  })

  it('shows loading skeletons', () => {
    vi.mocked(useGetMyBathhousesIdAnalytics).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    } as ReturnType<typeof useGetMyBathhousesIdAnalytics>)

    const { container } = renderPage()
    const skeletons = container.querySelectorAll('.ant-skeleton')
    expect(skeletons.length).toBeGreaterThan(0)
  })

  it('calls all three API hooks with selected bathhouse', () => {
    renderPage()
    expect(useGetMyBathhousesIdAnalytics).toHaveBeenCalledWith(
      'bath-1',
      { period: '30d' },
      { query: { enabled: true } },
    )
    expect(useGetMyBathhousesIdAnalyticsDaily).toHaveBeenCalledWith(
      'bath-1',
      expect.objectContaining({ from: expect.any(String), to: expect.any(String) }),
      { query: { enabled: true } },
    )
    expect(useGetMyBathhousesIdAnalyticsPerformance).toHaveBeenCalledWith(
      'bath-1',
      { period: '30d' },
      { query: { enabled: true } },
    )
  })
})
