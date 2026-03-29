import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import LoyaltyDashboard from '@/pages/client/LoyaltyDashboard'

vi.mock('@/api/generated/loyalty/loyalty', () => ({
  useGetMyLoyalty: vi.fn(),
  useGetMyLoyaltyLevels: vi.fn(),
  useGetMyLoyaltyTransactions: vi.fn(),
}))

import {
  useGetMyLoyalty,
  useGetMyLoyaltyLevels,
  useGetMyLoyaltyTransactions,
} from '@/api/generated/loyalty/loyalty'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter>{ui}</MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockLoyalty = {
  user_id: 'user-1',
  level: 'silver',
  points: 350,
  visit_count: 8,
  total_earned: 500,
  total_spent: 150,
  privileges: {
    discount_percent: 5,
    point_multiplier: 1.5,
    next_level: 'gold',
    visits_to_next: 7,
  },
  created_at: '2025-01-01T00:00:00Z',
  updated_at: '2025-06-01T00:00:00Z',
}

const mockLevels = [
  { level: 'bronze', min_visits: 0, discount_percent: 0, point_multiplier: 1 },
  { level: 'silver', min_visits: 5, discount_percent: 5, point_multiplier: 1.5 },
  { level: 'gold', min_visits: 15, discount_percent: 10, point_multiplier: 2 },
  { level: 'platinum', min_visits: 30, discount_percent: 15, point_multiplier: 2.5 },
]

const mockTransactions = [
  {
    id: 'tx-1',
    user_id: 'user-1',
    type: 'earned',
    amount: 100,
    description: 'Бонус за бронирование',
    booking_id: 'booking-1',
    created_at: '2025-06-01T10:00:00Z',
  },
  {
    id: 'tx-2',
    user_id: 'user-1',
    type: 'spent',
    amount: -50,
    description: 'Использовано при оплате',
    booking_id: 'booking-2',
    created_at: '2025-06-10T14:00:00Z',
  },
]

function setupMocks(overrides?: {
  loyalty?: typeof mockLoyalty | null
  levels?: typeof mockLevels
  transactions?: typeof mockTransactions
  loadingLoyalty?: boolean
  loadingLevels?: boolean
  loadingTx?: boolean
}) {
  vi.mocked(useGetMyLoyalty).mockReturnValue({
    data: overrides?.loyalty !== undefined
      ? overrides.loyalty === null
        ? { data: undefined, success: true }
        : { data: overrides.loyalty, success: true }
      : { data: mockLoyalty, success: true },
    isLoading: overrides?.loadingLoyalty ?? false,
  } as unknown as ReturnType<typeof useGetMyLoyalty>)

  vi.mocked(useGetMyLoyaltyLevels).mockReturnValue({
    data: { data: overrides?.levels ?? mockLevels, success: true },
    isLoading: overrides?.loadingLevels ?? false,
  } as unknown as ReturnType<typeof useGetMyLoyaltyLevels>)

  vi.mocked(useGetMyLoyaltyTransactions).mockReturnValue({
    data: {
      data: overrides?.transactions ?? mockTransactions,
      success: true,
      meta: { page: 1, page_size: 10, total_count: 2, total_pages: 1 },
    },
    isLoading: overrides?.loadingTx ?? false,
  } as unknown as ReturnType<typeof useGetMyLoyaltyTransactions>)
}

describe('LoyaltyDashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders page title', () => {
    setupMocks()
    renderWithProviders(<LoyaltyDashboard />)
    expect(screen.getByText('Программа лояльности')).toBeInTheDocument()
  })

  it('displays current level name', () => {
    setupMocks()
    renderWithProviders(<LoyaltyDashboard />)
    const silverElements = screen.getAllByText('Серебро')
    expect(silverElements.length).toBeGreaterThanOrEqual(1)
  })

  it('displays points balance', () => {
    setupMocks()
    renderWithProviders(<LoyaltyDashboard />)
    expect(screen.getByText('350')).toBeInTheDocument()
  })

  it('displays visit count', () => {
    setupMocks()
    renderWithProviders(<LoyaltyDashboard />)
    expect(screen.getByText('8')).toBeInTheDocument()
  })

  it('displays total earned and spent', () => {
    setupMocks()
    renderWithProviders(<LoyaltyDashboard />)
    expect(screen.getByText('500 баллов')).toBeInTheDocument()
    expect(screen.getByText('150 баллов')).toBeInTheDocument()
  })

  it('shows progress to next level', () => {
    setupMocks()
    renderWithProviders(<LoyaltyDashboard />)
    expect(screen.getByText(/До уровня Золото/)).toBeInTheDocument()
    expect(screen.getByText('ещё 7 визитов')).toBeInTheDocument()
  })

  it('displays privileges section', () => {
    setupMocks()
    renderWithProviders(<LoyaltyDashboard />)
    expect(screen.getByText('Ваши привилегии')).toBeInTheDocument()
    expect(screen.getByText('5%')).toBeInTheDocument()
    expect(screen.getByText('x1.5')).toBeInTheDocument()
  })

  it('renders all level info cards', () => {
    setupMocks()
    renderWithProviders(<LoyaltyDashboard />)
    expect(screen.getByText('Уровни программы')).toBeInTheDocument()
    expect(screen.getByText('Бронза')).toBeInTheDocument()
    // Серебро is also shown in the main card, so it should appear multiple times
    expect(screen.getByText('Золото')).toBeInTheDocument()
    expect(screen.getByText('Платина')).toBeInTheDocument()
  })

  it('marks current level card', () => {
    setupMocks()
    renderWithProviders(<LoyaltyDashboard />)
    expect(screen.getByText('Текущий')).toBeInTheDocument()
  })

  it('displays transaction history', () => {
    setupMocks()
    renderWithProviders(<LoyaltyDashboard />)
    expect(screen.getByText('История операций')).toBeInTheDocument()
    expect(screen.getByText('Бонус за бронирование')).toBeInTheDocument()
    expect(screen.getByText('Использовано при оплате')).toBeInTheDocument()
  })

  it('shows transaction type tags', () => {
    setupMocks()
    renderWithProviders(<LoyaltyDashboard />)
    expect(screen.getByText('Начислено')).toBeInTheDocument()
    expect(screen.getByText('Списано')).toBeInTheDocument()
  })

  it('formats transaction amounts with signs', () => {
    setupMocks()
    renderWithProviders(<LoyaltyDashboard />)
    expect(screen.getByText('+100')).toBeInTheDocument()
    expect(screen.getByText('-50')).toBeInTheDocument()
  })

  it('shows loading spinner when data is loading', () => {
    setupMocks({ loadingLoyalty: true })
    renderWithProviders(<LoyaltyDashboard />)
    expect(document.querySelector('.ant-spin-spinning')).toBeInTheDocument()
  })

  it('shows empty state when no loyalty data', () => {
    setupMocks({ loyalty: null })
    renderWithProviders(<LoyaltyDashboard />)
    expect(screen.getByText(/Начните пользоваться сервисом/)).toBeInTheDocument()
  })

  it('shows empty state when no transactions', () => {
    setupMocks({ transactions: [] })
    renderWithProviders(<LoyaltyDashboard />)
    expect(screen.getByText('Пока нет операций с баллами')).toBeInTheDocument()
  })

  it('displays level requirements in cards', () => {
    setupMocks()
    renderWithProviders(<LoyaltyDashboard />)
    expect(screen.getByText('От 0 визитов')).toBeInTheDocument()
    expect(screen.getByText('От 5 визитов')).toBeInTheDocument()
    expect(screen.getByText('От 15 визитов')).toBeInTheDocument()
    expect(screen.getByText('От 30 визитов')).toBeInTheDocument()
  })
})
