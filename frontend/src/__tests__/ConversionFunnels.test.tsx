import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi } from 'vitest'
import ConversionFunnels from '@/pages/admin/ConversionFunnels'

vi.mock('@/api/generated/admin-analytics/admin-analytics', () => ({
  useGetAdminAnalyticsFunnel: vi.fn(),
}))

import { useGetAdminAnalyticsFunnel } from '@/api/generated/admin-analytics/admin-analytics'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/analytics/funnels']}>
            <Routes>
              <Route path="/admin/analytics/funnels" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockFunnel = {
  period: '30d',
  steps: [
    { name: 'Визит', count: 10000, percentage: 100 },
    { name: 'Поиск', count: 7500, percentage: 75 },
    { name: 'Просмотр', count: 3000, percentage: 30 },
    { name: 'Бронирование', count: 1000, percentage: 10 },
    { name: 'Оплата', count: 800, percentage: 8 },
    { name: 'Завершение', count: 750, percentage: 7.5 },
  ],
}

function setupMocks(data = mockFunnel, loading = false) {
  vi.mocked(useGetAdminAnalyticsFunnel).mockReturnValue({
    data: { data, success: true },
    isLoading: loading,
  } as ReturnType<typeof useGetAdminAnalyticsFunnel>)
}

describe('ConversionFunnels', () => {
  it('renders title and period selector', () => {
    setupMocks()
    renderWithProviders(<ConversionFunnels />)

    expect(screen.getByText('Воронка конверсии')).toBeInTheDocument()
    expect(screen.getByText('Месяц')).toBeInTheDocument()
  })

  it('renders funnel steps', () => {
    setupMocks()
    renderWithProviders(<ConversionFunnels />)

    expect(screen.getByText('Визит')).toBeInTheDocument()
    expect(screen.getByText('Поиск')).toBeInTheDocument()
    expect(screen.getByText('Бронирование')).toBeInTheDocument()
    expect(screen.getByText('Завершение')).toBeInTheDocument()
  })

  it('renders summary KPI cards', () => {
    setupMocks()
    renderWithProviders(<ConversionFunnels />)

    expect(screen.getByText('Начало воронки')).toBeInTheDocument()
    expect(screen.getByText('Конец воронки')).toBeInTheDocument()
    expect(screen.getByText('Общая конверсия')).toBeInTheDocument()
  })

  it('renders step percentages', () => {
    setupMocks()
    renderWithProviders(<ConversionFunnels />)

    expect(screen.getByText('100.0%')).toBeInTheDocument()
    expect(screen.getByText('75.0%')).toBeInTheDocument()
  })

  it('clamps visual bar width while preserving raw percentages above 100', () => {
    setupMocks({
      period: '30d',
      steps: [
        { name: 'visit', count: 2, percentage: 100 },
        { name: 'start_booking', count: 8, percentage: 400 },
      ],
    })

    renderWithProviders(<ConversionFunnels />)

    const oversizedBar = screen.getByTestId('funnel-bar-1')
    expect(screen.getByText('400.0%')).toBeInTheDocument()
    expect(oversizedBar).toHaveAttribute('aria-valuenow', '100')
    expect(oversizedBar.style.getPropertyValue('--rh-funnel-bar-width')).toBe('100%')
  })

  it('shows growth between funnel stages instead of a negative drop-off', () => {
    setupMocks({
      period: '30d',
      steps: [
        { name: 'search', count: 0, percentage: 0 },
        { name: 'view_card', count: 3, percentage: 150 },
      ],
    })

    renderWithProviders(<ConversionFunnels />)

    expect(screen.getByText('Прирост: +3 (100.0%)')).toBeInTheDocument()
    expect(screen.queryByText(/Отсев: -/)).not.toBeInTheDocument()
  })

  it('renders empty state when no data', () => {
    vi.mocked(useGetAdminAnalyticsFunnel).mockReturnValue({
      data: { data: { period: '30d', steps: [] }, success: true },
      isLoading: false,
    } as ReturnType<typeof useGetAdminAnalyticsFunnel>)

    renderWithProviders(<ConversionFunnels />)
    expect(screen.getByText('Нет данных')).toBeInTheDocument()
  })
})
