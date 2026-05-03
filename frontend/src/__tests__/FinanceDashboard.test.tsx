import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi } from 'vitest'
import FinanceDashboard from '@/pages/finance/FinanceDashboard'

vi.mock('@/api/generated/wallet/wallet', () => ({
  useGetMyWallet: vi.fn(),
  useGetMyWalletTransactions: vi.fn(),
}))

import {
  useGetMyWallet,
  useGetMyWalletTransactions,
} from '@/api/generated/wallet/wallet'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/finance']}>
            <Routes>
              <Route path="/finance" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockWallet = {
  balance: 5000000,
  available: 4500000,
  held_amount: 500000,
  currency: 'RUB',
}

const mockTransactions = [
  {
    id: 'tx-1',
    amount: 300000,
    type: 'income',
    status: 'completed',
    description: 'Оплата бронирования #123',
    balance_after: 5000000,
    created_at: '2026-03-25T10:00:00Z',
    is_bonus: false,
  },
  {
    id: 'tx-2',
    amount: 50000,
    type: 'service_fee',
    status: 'completed',
    description: 'Комиссия платформы',
    balance_after: 4950000,
    created_at: '2026-03-25T10:00:00Z',
    is_bonus: false,
  },
]

function setupMocks(overrides?: {
  wallet?: typeof mockWallet | null
  transactions?: typeof mockTransactions
  walletLoading?: boolean
}) {
  vi.mocked(useGetMyWallet).mockReturnValue({
    data: {
      data: overrides?.wallet !== undefined ? overrides.wallet : mockWallet,
      success: true,
    },
    isLoading: overrides?.walletLoading ?? false,
  } as unknown as ReturnType<typeof useGetMyWallet>)

  vi.mocked(useGetMyWalletTransactions).mockReturnValue({
    data: {
      data: overrides?.transactions ?? mockTransactions,
      success: true,
      meta: { page: 1, page_size: 10, total_count: (overrides?.transactions ?? mockTransactions).length, total_pages: 1 },
    },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyWalletTransactions>)
}

describe('FinanceDashboard', () => {
  it('renders page title', () => {
    setupMocks()
    renderWithProviders(<FinanceDashboard />)

    expect(screen.getByText('Финансы')).toBeInTheDocument()
  })

  it('renders balance cards', () => {
    setupMocks()
    renderWithProviders(<FinanceDashboard />)

    expect(screen.getByText('Баланс')).toBeInTheDocument()
    expect(screen.getByText('Доступно')).toBeInTheDocument()
    expect(screen.getByText('Заморожено')).toBeInTheDocument()
  })

  it('renders income period selector', () => {
    setupMocks()
    renderWithProviders(<FinanceDashboard />)

    expect(screen.getByText('Доходы за период')).toBeInTheDocument()
  })

  it('renders transaction table', () => {
    setupMocks()
    renderWithProviders(<FinanceDashboard />)

    expect(screen.getByText('Последние операции')).toBeInTheDocument()
    expect(screen.getByText('Оплата бронирования #123')).toBeInTheDocument()
    expect(screen.getByText('Комиссия платформы')).toBeInTheDocument()
  })

  it('renders transaction type tags', () => {
    setupMocks()
    renderWithProviders(<FinanceDashboard />)

    expect(screen.getByText('Доход')).toBeInTheDocument()
    expect(screen.getByText('Комиссия')).toBeInTheDocument()
  })

  it('renders type filter dropdown', () => {
    setupMocks()
    renderWithProviders(<FinanceDashboard />)

    expect(screen.getByText('Все типы')).toBeInTheDocument()
  })

  it('renders empty state when no transactions', () => {
    setupMocks({ transactions: [] })
    renderWithProviders(<FinanceDashboard />)

    expect(screen.getByText('Нет операций')).toBeInTheDocument()
  })

  it('renders positive income amount in green', () => {
    setupMocks()
    renderWithProviders(<FinanceDashboard />)

    expect(screen.getByText('+3000 ₽')).toBeInTheDocument()
  })

  it('renders fee amount in red', () => {
    setupMocks()
    renderWithProviders(<FinanceDashboard />)

    expect(screen.getByText('-500 ₽')).toBeInTheDocument()
  })
})
