import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi } from 'vitest'
import AdminFinanceDashboard from '@/pages/admin/AdminFinanceDashboard'

vi.mock('@/api/generated/admin/admin', () => ({
  useGetApiV1AdminReconciliationSummary: vi.fn(),
  useGetApiV1AdminReconciliationReports: vi.fn(),
  useGetApiV1AdminReconciliationSnapshots: vi.fn(),
  usePostApiV1AdminReconciliationSnapshot: vi.fn(),
  usePostApiV1AdminReconciliationReconcile: vi.fn(),
}))

import {
  useGetApiV1AdminReconciliationSummary,
  useGetApiV1AdminReconciliationReports,
  useGetApiV1AdminReconciliationSnapshots,
  usePostApiV1AdminReconciliationSnapshot,
  usePostApiV1AdminReconciliationReconcile,
} from '@/api/generated/admin/admin'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/finance']}>
            <Routes>
              <Route path="/admin/finance" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockSummary = {
  client_wallets_count: 120,
  client_wallets_total: 5000000,
  owner_wallets_count: 30,
  owner_wallets_total: 2500000,
  escrow_count: 15,
  escrow_held_total: 1000000,
  wallet_holds_total: 300000,
  platform_total: 8800000,
  last_report: {
    id: 'r-1',
    status: 'ok',
    internal_payments_count: 100,
    provider_payments_count: 100,
    payment_discrepancy: 0,
    created_at: '2026-03-29T12:00:00Z',
  },
}

const mockReports = [
  {
    id: 'r-1',
    created_at: '2026-03-29T12:00:00Z',
    period_start: '2026-03-01T00:00:00Z',
    period_end: '2026-03-29T00:00:00Z',
    status: 'ok',
    internal_payments_count: 100,
    provider_payments_count: 100,
    payment_discrepancy: 0,
    refund_discrepancy: 0,
  },
]

const mockSnapshots = [
  {
    id: 's-1',
    snapshot_date: '2026-03-29T00:00:00Z',
    expected_total: 8800000,
    actual_total: 8800000,
    discrepancy: 0,
    status: 'ok',
  },
]

function setupMocks() {
  vi.mocked(useGetApiV1AdminReconciliationSummary).mockReturnValue({
    data: { data: mockSummary, success: true },
    isLoading: false,
  } as ReturnType<typeof useGetApiV1AdminReconciliationSummary>)
  vi.mocked(useGetApiV1AdminReconciliationReports).mockReturnValue({
    data: { data: mockReports, meta: { total_count: 1 }, success: true },
    isLoading: false,
  } as ReturnType<typeof useGetApiV1AdminReconciliationReports>)
  vi.mocked(useGetApiV1AdminReconciliationSnapshots).mockReturnValue({
    data: { data: mockSnapshots, meta: { total_count: 1 }, success: true },
    isLoading: false,
  } as ReturnType<typeof useGetApiV1AdminReconciliationSnapshots>)
  vi.mocked(usePostApiV1AdminReconciliationSnapshot).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof usePostApiV1AdminReconciliationSnapshot>)
  vi.mocked(usePostApiV1AdminReconciliationReconcile).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof usePostApiV1AdminReconciliationReconcile>)
}

describe('AdminFinanceDashboard', () => {
  it('renders title and float summary cards', () => {
    setupMocks()
    renderWithProviders(<AdminFinanceDashboard />)

    expect(screen.getByText('Финансовый дашборд')).toBeInTheDocument()
    expect(screen.getByText('Кошельки клиентов')).toBeInTheDocument()
    expect(screen.getByText('Кошельки владельцев')).toBeInTheDocument()
    expect(screen.getByText('Эскроу')).toBeInTheDocument()
    expect(screen.getByText('Холды')).toBeInTheDocument()
  })

  it('renders wallet counts in card descriptions', () => {
    setupMocks()
    renderWithProviders(<AdminFinanceDashboard />)

    expect(screen.getByText('120 кошельков')).toBeInTheDocument()
    expect(screen.getByText('30 кошельков')).toBeInTheDocument()
    expect(screen.getByText('15 записей')).toBeInTheDocument()
  })

  it('renders reconciliation reports table', () => {
    setupMocks()
    renderWithProviders(<AdminFinanceDashboard />)

    expect(screen.getByText('Отчёты сверки')).toBeInTheDocument()
    expect(screen.getAllByText('OK').length).toBeGreaterThanOrEqual(1)
  })

  it('renders float snapshots table', () => {
    setupMocks()
    renderWithProviders(<AdminFinanceDashboard />)

    expect(screen.getByText('Снимки флоата')).toBeInTheDocument()
  })

  it('renders action buttons', () => {
    setupMocks()
    renderWithProviders(<AdminFinanceDashboard />)

    expect(screen.getByText('Снять снимок')).toBeInTheDocument()
    expect(screen.getByText('Сверка с провайдером')).toBeInTheDocument()
  })

  it('calls snapshot mutation on button click', () => {
    setupMocks()
    const mutateFn = vi.fn()
    vi.mocked(usePostApiV1AdminReconciliationSnapshot).mockReturnValue({
      mutate: mutateFn,
      isPending: false,
    } as unknown as ReturnType<typeof usePostApiV1AdminReconciliationSnapshot>)

    renderWithProviders(<AdminFinanceDashboard />)
    fireEvent.click(screen.getByText('Снять снимок'))
    expect(mutateFn).toHaveBeenCalled()
  })

  it('renders last report info when available', () => {
    setupMocks()
    renderWithProviders(<AdminFinanceDashboard />)

    expect(screen.getByText('Последняя сверка')).toBeInTheDocument()
  })
})
