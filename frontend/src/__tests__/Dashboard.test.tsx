import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import Dashboard from '@/pages/Dashboard'
import { useBathhouseStore } from '@/stores/bathhouse'
import { useGetMyBathhousesIdAnalytics } from '@/api/generated/analytics/analytics'

vi.mock('@/api/generated/analytics/analytics', () => ({
  useGetMyBathhousesIdAnalytics: vi.fn(),
}))

function renderDashboard() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <Dashboard />
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockAnalyticsData = {
  data: {
    data: {
      bookings: 42,
      bookings_change: 15.5,
      revenue: 1500000,
      revenue_change: -3.2,
      views: 1200,
      views_change: 8.0,
      rating: 4.7,
      rating_change: 2.1,
      avg_check: 357000,
      conversion_rate: 0.035,
      period: '30d',
    },
  },
  isLoading: false,
  isError: false,
  error: null,
}

describe('Dashboard', () => {
  beforeEach(() => {
    useBathhouseStore.setState({ selectedBathhouseId: 'bath-1' })
    vi.mocked(useGetMyBathhousesIdAnalytics).mockReturnValue(
      mockAnalyticsData as ReturnType<typeof useGetMyBathhousesIdAnalytics>,
    )
  })

  it('shows prompt to select bathhouse when none selected', () => {
    useBathhouseStore.setState({ selectedBathhouseId: null })
    renderDashboard()
    expect(screen.getByText('Выберите баню')).toBeInTheDocument()
    expect(
      screen.getByText('Для просмотра аналитики выберите баню в верхнем меню.'),
    ).toBeInTheDocument()
  })

  it('renders KPI cards with data', () => {
    const { container } = renderDashboard()
    const text = container.textContent ?? ''
    expect(screen.getByText('Бронирования')).toBeInTheDocument()
    expect(text).toContain('42')
    expect(screen.getByText('Выручка')).toBeInTheDocument()
    expect(text).toContain('15000 ₽')
    expect(screen.getByText('Просмотры')).toBeInTheDocument()
    expect(text).toContain('1,200')
    expect(screen.getByText('Рейтинг')).toBeInTheDocument()
    expect(screen.getByText('Средний чек')).toBeInTheDocument()
    expect(text).toContain('3570 ₽')
    expect(screen.getByText('Конверсия')).toBeInTheDocument()
    expect(text).toContain('3.5%')
  })

  it('renders change indicators with correct colors', () => {
    renderDashboard()
    const positiveChanges = screen.getAllByText(/\+.*% к пред\. периоду/)
    expect(positiveChanges.length).toBeGreaterThan(0)
    positiveChanges.forEach((el) => {
      expect(el).toHaveClass('rh-stat-change--positive')
    })

    const negativeChanges = screen.getAllByText(/-.*% к пред\. периоду/)
    expect(negativeChanges.length).toBeGreaterThan(0)
    negativeChanges.forEach((el) => {
      expect(el).toHaveClass('rh-stat-change--negative')
    })
  })

  it('renders period selector with 30d default', () => {
    renderDashboard()
    expect(screen.getByText('1 день')).toBeInTheDocument()
    expect(screen.getByText('7 дней')).toBeInTheDocument()
    expect(screen.getByText('30 дней')).toBeInTheDocument()
    expect(screen.getByText('90 дней')).toBeInTheDocument()
  })

  it('calls API hook with selected period', () => {
    renderDashboard()
    expect(useGetMyBathhousesIdAnalytics).toHaveBeenCalledWith(
      'bath-1',
      { period: '30d' },
      { query: { enabled: true } },
    )
  })

  it('switches period on click', () => {
    renderDashboard()
    fireEvent.click(screen.getByText('7 дней'))
    expect(useGetMyBathhousesIdAnalytics).toHaveBeenCalledWith(
      'bath-1',
      { period: '7d' },
      { query: { enabled: true } },
    )
  })

  it('shows loading skeletons', () => {
    vi.mocked(useGetMyBathhousesIdAnalytics).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    } as ReturnType<typeof useGetMyBathhousesIdAnalytics>)

    const { container } = renderDashboard()
    const skeletons = container.querySelectorAll('.ant-skeleton')
    expect(skeletons.length).toBe(6)
  })

  it('shows error alert on API failure', () => {
    vi.mocked(useGetMyBathhousesIdAnalytics).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: { error: { message: 'Сервер недоступен' } },
    } as unknown as ReturnType<typeof useGetMyBathhousesIdAnalytics>)

    renderDashboard()
    expect(screen.getByText('Ошибка загрузки')).toBeInTheDocument()
    expect(screen.getByText('Сервер недоступен')).toBeInTheDocument()
  })

  it('shows default error message when error has no message', () => {
    vi.mocked(useGetMyBathhousesIdAnalytics).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: {},
    } as unknown as ReturnType<typeof useGetMyBathhousesIdAnalytics>)

    renderDashboard()
    expect(screen.getByText('Не удалось загрузить аналитику')).toBeInTheDocument()
  })

  it('disables query when no bathhouse selected', () => {
    useBathhouseStore.setState({ selectedBathhouseId: null })
    renderDashboard()
    expect(useGetMyBathhousesIdAnalytics).toHaveBeenCalledWith(
      '',
      { period: '30d' },
      { query: { enabled: false } },
    )
  })

  it('renders title "Дашборд"', () => {
    renderDashboard()
    expect(screen.getByText('Дашборд')).toBeInTheDocument()
  })
})
