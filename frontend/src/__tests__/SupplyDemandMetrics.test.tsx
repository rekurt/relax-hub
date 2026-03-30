import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi } from 'vitest'
import SupplyDemandMetrics from '@/pages/admin/SupplyDemandMetrics'

vi.mock('@/api/generated/admin-analytics/admin-analytics', () => ({
  useGetAdminAnalyticsGeo: vi.fn(),
  useGetAdminAnalyticsWallet: vi.fn(),
  useGetAdminAnalyticsBusinessMetrics: vi.fn(),
  useGetAdminAnalyticsPnl: vi.fn(),
}))

import {
  useGetAdminAnalyticsGeo,
  useGetAdminAnalyticsWallet,
  useGetAdminAnalyticsBusinessMetrics,
  useGetAdminAnalyticsPnl,
} from '@/api/generated/admin-analytics/admin-analytics'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/analytics/supply-demand']}>
            <Routes>
              <Route path="/admin/analytics/supply-demand" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockGeo = {
  period: '30d',
  cities: [
    { city_id: 1, city_name: 'Москва', listing_count: 150, search_count: 5000, booking_count: 800 },
    { city_id: 2, city_name: 'Санкт-Петербург', listing_count: 80, search_count: 2500, booking_count: 400 },
  ],
}

const mockWallet = {
  active_wallets: 1200,
  total_client_balance: 15000000,
  total_owner_balance: 8000000,
  total_escrow: 3000000,
  wallet_payment_share: 35.5,
  expired_bonus_volume: 500000,
}

const mockBiz = {
  dau: 450,
  mau: 12000,
  adr: 350000,
  arpu: 120000,
  churn_rate: 8.5,
  ltv: 950000,
}

const mockPnl = {
  gmv: 50000000,
  take_rate: 10.5,
  platform_revenue: 5250000,
  revenue_per_booking: 52500,
  service_fees_total: 4000000,
  subscriptions_total: 750000,
  promotions_total: 500000,
  total_bookings: 100,
  gmv_per_booking: 500000,
}

function setupMocks() {
  vi.mocked(useGetAdminAnalyticsGeo).mockReturnValue({
    data: { data: mockGeo, success: true },
    isLoading: false,
  } as ReturnType<typeof useGetAdminAnalyticsGeo>)
  vi.mocked(useGetAdminAnalyticsWallet).mockReturnValue({
    data: { data: mockWallet, success: true },
    isLoading: false,
  } as ReturnType<typeof useGetAdminAnalyticsWallet>)
  vi.mocked(useGetAdminAnalyticsBusinessMetrics).mockReturnValue({
    data: { data: mockBiz, success: true },
    isLoading: false,
  } as ReturnType<typeof useGetAdminAnalyticsBusinessMetrics>)
  vi.mocked(useGetAdminAnalyticsPnl).mockReturnValue({
    data: { data: mockPnl, success: true },
    isLoading: false,
  } as ReturnType<typeof useGetAdminAnalyticsPnl>)
}

describe('SupplyDemandMetrics', () => {
  it('renders title and period selector', () => {
    setupMocks()
    renderWithProviders(<SupplyDemandMetrics />)

    expect(screen.getByText('Метрики предложения и спроса')).toBeInTheDocument()
    expect(screen.getByText('Месяц')).toBeInTheDocument()
  })

  it('renders supply/demand KPI cards', () => {
    setupMocks()
    renderWithProviders(<SupplyDemandMetrics />)

    expect(screen.getByText('Активные листинги')).toBeInTheDocument()
    expect(screen.getByText('Поисковые запросы')).toBeInTheDocument()
    expect(screen.getByText('Конверсия поиск→бронь')).toBeInTheDocument()
  })

  it('renders business metrics cards', () => {
    setupMocks()
    renderWithProviders(<SupplyDemandMetrics />)

    expect(screen.getByText('DAU')).toBeInTheDocument()
    expect(screen.getByText('ADR')).toBeInTheDocument()
    expect(screen.getByText('ARPU')).toBeInTheDocument()
    expect(screen.getByText('Отток')).toBeInTheDocument()
  })

  it('renders P&L metrics cards', () => {
    setupMocks()
    renderWithProviders(<SupplyDemandMetrics />)

    expect(screen.getByText('GMV')).toBeInTheDocument()
    expect(screen.getByText('Take Rate')).toBeInTheDocument()
    expect(screen.getByText('Выручка платформы')).toBeInTheDocument()
  })

  it('renders wallet metrics cards', () => {
    setupMocks()
    renderWithProviders(<SupplyDemandMetrics />)

    expect(screen.getByText('Активные кошельки')).toBeInTheDocument()
    expect(screen.getByText('Баланс клиентов')).toBeInTheDocument()
    expect(screen.getByText('Баланс владельцев')).toBeInTheDocument()
    expect(screen.getByText('Оплата кошельком')).toBeInTheDocument()
  })

  it('renders city table', () => {
    setupMocks()
    renderWithProviders(<SupplyDemandMetrics />)

    expect(screen.getByText('Спрос и предложение по городам')).toBeInTheDocument()
    expect(screen.getByText('Москва')).toBeInTheDocument()
    expect(screen.getByText('Санкт-Петербург')).toBeInTheDocument()
  })

  it('renders conversion rate in city table', () => {
    setupMocks()
    renderWithProviders(<SupplyDemandMetrics />)

    // Moscow: 800/5000 = 16.0%, SPb: 400/2500 = 16.0%
    expect(screen.getAllByText('16.0%').length).toBeGreaterThanOrEqual(1)
  })
})
