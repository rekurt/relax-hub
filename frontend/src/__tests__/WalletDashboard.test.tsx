import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi } from 'vitest'
import WalletDashboard from '@/pages/client/WalletDashboard'

vi.mock('@/api/generated/wallet/wallet', () => ({
  useGetMyWallet: vi.fn(),
  useGetMyWalletTransactions: vi.fn(),
  useGetMyWalletHolds: vi.fn(),
  usePostMyWalletTopup: vi.fn(),
}))

import {
  useGetMyWallet,
  useGetMyWalletTransactions,
  useGetMyWalletHolds,
  usePostMyWalletTopup,
} from '@/api/generated/wallet/wallet'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/client/wallet']}>
            <Routes>
              <Route path="/client/wallet" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockWallet = {
  balance: 1500000,
  available: 1200000,
  held_amount: 300000,
  currency: 'RUB',
  expiring_soon: 50000,
  earliest_expiry: '2026-04-15T00:00:00Z',
}

const mockTransactions = [
  {
    id: 'tx-1',
    amount: 100000,
    type: 'topup',
    status: 'completed',
    description: 'Пополнение картой',
    balance_after: 1500000,
    created_at: '2026-03-25T10:00:00Z',
    is_bonus: false,
    reference_id: null,
    reference_type: null,
    expires_at: null,
  },
  {
    id: 'tx-2',
    amount: 50000,
    type: 'spend',
    status: 'completed',
    description: 'Оплата бронирования',
    balance_after: 1400000,
    created_at: '2026-03-26T14:00:00Z',
    is_bonus: false,
    reference_id: 'book-1',
    reference_type: 'booking',
    expires_at: null,
  },
  {
    id: 'tx-3',
    amount: 50000,
    type: 'cashback',
    status: 'completed',
    description: 'Кэшбэк за бронирование',
    balance_after: 1450000,
    created_at: '2026-03-27T09:00:00Z',
    is_bonus: true,
    reference_id: null,
    reference_type: null,
    expires_at: '2026-04-10T00:00:00Z',
  },
]

const mockHolds = [
  {
    id: 'hold-1',
    amount: 300000,
    description: 'Холд за бронирование',
    expires_at: '2026-04-01T12:00:00Z',
    status: 'active',
    reference_id: 'book-2',
    reference_type: 'booking',
    created_at: '2026-03-28T10:00:00Z',
  },
]

function setupMocks(overrides?: {
  wallet?: typeof mockWallet | null
  transactions?: typeof mockTransactions
  holds?: typeof mockHolds
}) {
  vi.mocked(useGetMyWallet).mockReturnValue({
    data: {
      data: overrides?.wallet !== undefined ? overrides.wallet : mockWallet,
      success: true,
    },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyWallet>)

  vi.mocked(useGetMyWalletTransactions).mockReturnValue({
    data: {
      data: overrides?.transactions ?? mockTransactions,
      success: true,
      meta: { page: 1, page_size: 10, total_count: (overrides?.transactions ?? mockTransactions).length, total_pages: 1 },
    },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyWalletTransactions>)

  vi.mocked(useGetMyWalletHolds).mockReturnValue({
    data: {
      data: overrides?.holds ?? mockHolds,
      success: true,
    },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyWalletHolds>)

  vi.mocked(usePostMyWalletTopup).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof usePostMyWalletTopup>)
}

describe('WalletDashboard', () => {
  it('renders wallet balance cards', () => {
    setupMocks()
    renderWithProviders(<WalletDashboard />)

    expect(screen.getByText('Кошелёк')).toBeInTheDocument()
    expect(screen.getByText('Баланс')).toBeInTheDocument()
    expect(screen.getByText('Доступно')).toBeInTheDocument()
    expect(screen.getByText('Заморожено')).toBeInTheDocument()
  })

  it('renders expiring bonus warning', () => {
    setupMocks()
    renderWithProviders(<WalletDashboard />)

    expect(screen.getByText(/Бонусы на сумму 500 ₽ скоро сгорят/)).toBeInTheDocument()
  })

  it('renders holds section when holds exist', () => {
    setupMocks()
    renderWithProviders(<WalletDashboard />)

    expect(screen.getByText('Замороженные средства')).toBeInTheDocument()
    expect(screen.getByText('Холд за бронирование')).toBeInTheDocument()
  })

  it('does not render holds section when no holds', () => {
    setupMocks({ holds: [] })
    renderWithProviders(<WalletDashboard />)

    expect(screen.queryByText('Замороженные средства')).not.toBeInTheDocument()
  })

  it('renders transaction history table', () => {
    setupMocks()
    renderWithProviders(<WalletDashboard />)

    expect(screen.getByText('История операций')).toBeInTheDocument()
    expect(screen.getByText('Пополнение картой')).toBeInTheDocument()
    expect(screen.getByText('Оплата бронирования')).toBeInTheDocument()
  })

  it('renders transaction type tags', () => {
    setupMocks()
    renderWithProviders(<WalletDashboard />)

    expect(screen.getByText('Пополнение')).toBeInTheDocument()
    expect(screen.getByText('Списание')).toBeInTheDocument()
    expect(screen.getByText('Кэшбэк')).toBeInTheDocument()
  })

  it('renders type filter dropdown', () => {
    setupMocks()
    renderWithProviders(<WalletDashboard />)

    expect(screen.getByText('Все типы')).toBeInTheDocument()
  })

  it('renders top-up form', () => {
    setupMocks()
    renderWithProviders(<WalletDashboard />)

    expect(screen.getByText('Пополнить кошелёк')).toBeInTheDocument()
    expect(screen.getByText('Пополнить')).toBeInTheDocument()
  })

  it('validates top-up requires amount', async () => {
    setupMocks()
    renderWithProviders(<WalletDashboard />)

    fireEvent.click(screen.getByText('Пополнить'))

    await waitFor(() => {
      expect(screen.getByText('Введите сумму')).toBeInTheDocument()
    }, { timeout: 5000 })
  })

  it('renders empty state when no transactions', () => {
    setupMocks({ transactions: [] })
    renderWithProviders(<WalletDashboard />)

    expect(screen.getByText('Нет операций по кошельку')).toBeInTheDocument()
  })

  it('renders transaction amounts with correct formatting', () => {
    setupMocks()
    renderWithProviders(<WalletDashboard />)

    // topup +1000 ₽
    expect(screen.getByText('+1000 ₽')).toBeInTheDocument()
    // spend shows without + prefix
    expect(screen.getAllByText('500 ₽').length).toBeGreaterThanOrEqual(1)
    // cashback +500 ₽
    expect(screen.getByText('+500 ₽')).toBeInTheDocument()
  })

  it('renders export button', () => {
    setupMocks()
    renderWithProviders(<WalletDashboard />)

    expect(screen.getByText('Экспорт')).toBeInTheDocument()
  })
})
