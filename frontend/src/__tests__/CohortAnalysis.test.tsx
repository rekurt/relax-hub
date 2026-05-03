import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi } from 'vitest'
import CohortAnalysis from '@/pages/admin/CohortAnalysis'

vi.mock('@/api/generated/admin-analytics/admin-analytics', () => ({
  useGetAdminAnalyticsCohorts: vi.fn(),
}))

import { useGetAdminAnalyticsCohorts } from '@/api/generated/admin-analytics/admin-analytics'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/analytics/cohorts']}>
            <Routes>
              <Route path="/admin/analytics/cohorts" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockCohorts = {
  cohorts: [
    {
      cohort_month: '2026-01',
      users_count: 120,
      total_spending: 5000000,
      retention_weeks: [100, 80, 60, 45, 30],
    },
    {
      cohort_month: '2026-02',
      users_count: 150,
      total_spending: 6500000,
      retention_weeks: [100, 85, 70, 50],
    },
  ],
}

function setupMocks(data = mockCohorts, loading = false) {
  vi.mocked(useGetAdminAnalyticsCohorts).mockReturnValue({
    data: { data, success: true },
    isLoading: loading,
  } as ReturnType<typeof useGetAdminAnalyticsCohorts>)
}

describe('CohortAnalysis', () => {
  it('renders title and months selector', () => {
    setupMocks()
    renderWithProviders(<CohortAnalysis />)

    expect(screen.getByText('Когортный анализ')).toBeInTheDocument()
    expect(screen.getByText('6 мес')).toBeInTheDocument()
  })

  it('renders cohort table with month names', () => {
    setupMocks()
    renderWithProviders(<CohortAnalysis />)

    expect(screen.getByText('2026-01')).toBeInTheDocument()
    expect(screen.getByText('2026-02')).toBeInTheDocument()
  })

  it('renders column headers', () => {
    setupMocks()
    renderWithProviders(<CohortAnalysis />)

    expect(screen.getAllByText('Когорта').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Пользователей').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Расход').length).toBeGreaterThanOrEqual(1)
  })

  it('renders retention percentages', () => {
    setupMocks()
    renderWithProviders(<CohortAnalysis />)

    expect(screen.getAllByText('100%').length).toBe(2)
    expect(screen.getByText('80%')).toBeInTheDocument()
  })

  it('renders empty state when no cohorts', () => {
    vi.mocked(useGetAdminAnalyticsCohorts).mockReturnValue({
      data: { data: { cohorts: [] }, success: true },
      isLoading: false,
    } as ReturnType<typeof useGetAdminAnalyticsCohorts>)

    renderWithProviders(<CohortAnalysis />)
    expect(screen.getByText('Нет данных')).toBeInTheDocument()
  })
})
