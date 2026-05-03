import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi } from 'vitest'
import PayoutPage from '@/pages/finance/PayoutPage'

vi.mock('@/api/generated/wallet/wallet', () => ({
  useGetMyWallet: vi.fn(),
  useGetMyWalletPayouts: vi.fn(),
  usePostMyWalletPayout: vi.fn(),
  usePutMyWalletAutoPayout: vi.fn(),
}))

import {
  useGetMyWallet,
  useGetMyWalletPayouts,
  usePostMyWalletPayout,
  usePutMyWalletAutoPayout,
} from '@/api/generated/wallet/wallet'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/finance/payouts']}>
            <Routes>
              <Route path="/finance/payouts" element={ui} />
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

const mockPayouts = [
  {
    id: 'payout-1',
    amount: 200000,
    payout_method: 'sbp',
    status: 'completed',
    requested_at: '2026-03-20T10:00:00Z',
    processed_at: '2026-03-20T10:05:00Z',
    failure_reason: null,
  },
  {
    id: 'payout-2',
    amount: 100000,
    payout_method: 'bank_transfer',
    status: 'pending',
    requested_at: '2026-03-28T14:00:00Z',
    processed_at: null,
    failure_reason: null,
  },
]

function setupMocks(overrides?: {
  wallet?: typeof mockWallet | null
  payouts?: typeof mockPayouts
}) {
  vi.mocked(useGetMyWallet).mockReturnValue({
    data: {
      data: overrides?.wallet !== undefined ? overrides.wallet : mockWallet,
      success: true,
    },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyWallet>)

  vi.mocked(useGetMyWalletPayouts).mockReturnValue({
    data: {
      data: overrides?.payouts ?? mockPayouts,
      success: true,
      meta: { page: 1, page_size: 10, total_count: (overrides?.payouts ?? mockPayouts).length, total_pages: 1 },
    },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyWalletPayouts>)

  vi.mocked(usePostMyWalletPayout).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof usePostMyWalletPayout>)

  vi.mocked(usePutMyWalletAutoPayout).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof usePutMyWalletAutoPayout>)
}

describe('PayoutPage', () => {
  it('renders page title', () => {
    setupMocks()
    renderWithProviders(<PayoutPage />)

    expect(screen.getByText('Выплаты')).toBeInTheDocument()
  })

  it('renders available balance card', () => {
    setupMocks()
    renderWithProviders(<PayoutPage />)

    expect(screen.getByText('Доступно к выводу')).toBeInTheDocument()
  })

  it('renders daily and monthly limits', () => {
    setupMocks()
    renderWithProviders(<PayoutPage />)

    expect(screen.getByText('Дневной лимит')).toBeInTheDocument()
    expect(screen.getByText('Месячный лимит')).toBeInTheDocument()
  })

  it('renders payout request form', () => {
    setupMocks()
    renderWithProviders(<PayoutPage />)

    expect(screen.getByText('Запросить выплату')).toBeInTheDocument()
    expect(screen.getByText('Вывести')).toBeInTheDocument()
  })

  it('renders auto-payout toggle', () => {
    setupMocks()
    renderWithProviders(<PayoutPage />)

    expect(screen.getByText('Автовыплата')).toBeInTheDocument()
  })

  it('renders payout history table', () => {
    setupMocks()
    renderWithProviders(<PayoutPage />)

    expect(screen.getByText('История выплат')).toBeInTheDocument()
  })

  it('renders payout method tags', () => {
    setupMocks()
    renderWithProviders(<PayoutPage />)

    expect(screen.getAllByText('СБП (мгновенный)').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Банковский перевод')).toBeInTheDocument()
  })

  it('renders payout status tags', () => {
    setupMocks()
    renderWithProviders(<PayoutPage />)

    expect(screen.getByText('Выполнена')).toBeInTheDocument()
    expect(screen.getByText('В обработке')).toBeInTheDocument()
  })

  it('renders empty state when no payouts', () => {
    setupMocks({ payouts: [] })
    renderWithProviders(<PayoutPage />)

    expect(screen.getByText('Нет выплат')).toBeInTheDocument()
  })

  it('validates payout form requires amount', async () => {
    setupMocks()
    renderWithProviders(<PayoutPage />)

    fireEvent.click(screen.getByText('Вывести'))

    await waitFor(() => {
      expect(screen.getByText('Введите сумму')).toBeInTheDocument()
    }, { timeout: 5000 })
  })
})
