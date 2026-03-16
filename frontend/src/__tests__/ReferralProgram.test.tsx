import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import ReferralProgram from '@/pages/client/ReferralProgram'

vi.mock('@/api/generated/referral/referral', () => ({
  useGetMyReferral: vi.fn(),
  useGetMyReferralStats: vi.fn(),
  useGetMyReferralBalance: vi.fn(),
}))

import {
  useGetMyReferral,
  useGetMyReferralStats,
  useGetMyReferralBalance,
} from '@/api/generated/referral/referral'

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

const mockReferral = {
  referral_code: 'ABC123',
  referral_link: 'https://bani.app/ref/ABC123',
}

const mockStats = {
  total_invited: 5,
  total_completed: 3,
  total_earned: 300000, // 3000 rubles in kopecks
}

const mockBalance = {
  balance: 150000, // 1500 rubles in kopecks
  total_earned: 300000,
}

function setupMocks(overrides?: {
  referral?: typeof mockReferral | null
  stats?: typeof mockStats | null
  balance?: typeof mockBalance | null
  loading?: boolean
}) {
  vi.mocked(useGetMyReferral).mockReturnValue({
    data: overrides?.referral !== undefined
      ? overrides.referral === null
        ? { data: undefined, success: true }
        : { data: overrides.referral, success: true }
      : { data: mockReferral, success: true },
    isLoading: overrides?.loading ?? false,
  } as unknown as ReturnType<typeof useGetMyReferral>)

  vi.mocked(useGetMyReferralStats).mockReturnValue({
    data: overrides?.stats !== undefined
      ? overrides.stats === null
        ? { data: undefined, success: true }
        : { data: overrides.stats, success: true }
      : { data: mockStats, success: true },
    isLoading: overrides?.loading ?? false,
  } as unknown as ReturnType<typeof useGetMyReferralStats>)

  vi.mocked(useGetMyReferralBalance).mockReturnValue({
    data: overrides?.balance !== undefined
      ? overrides.balance === null
        ? { data: undefined, success: true }
        : { data: overrides.balance, success: true }
      : { data: mockBalance, success: true },
    isLoading: overrides?.loading ?? false,
  } as unknown as ReturnType<typeof useGetMyReferralBalance>)
}

describe('ReferralProgram', () => {
  const originalClipboard = navigator.clipboard

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    Object.assign(navigator, { clipboard: originalClipboard })
  })

  it('renders page title', () => {
    setupMocks()
    renderWithProviders(<ReferralProgram />)
    expect(screen.getByText('Реферальная программа')).toBeInTheDocument()
  })

  it('displays referral code', () => {
    setupMocks()
    renderWithProviders(<ReferralProgram />)
    expect(screen.getByText('ABC123')).toBeInTheDocument()
  })

  it('displays referral code section title', () => {
    setupMocks()
    renderWithProviders(<ReferralProgram />)
    expect(screen.getByText('Ваш реферальный код')).toBeInTheDocument()
  })

  it('renders copy and share buttons', () => {
    setupMocks()
    renderWithProviders(<ReferralProgram />)
    expect(screen.getByText('Копировать код')).toBeInTheDocument()
    expect(screen.getByText('Поделиться')).toBeInTheDocument()
    expect(screen.getByText('Скопировать ссылку')).toBeInTheDocument()
  })

  it('copies referral code to clipboard on button click', async () => {
    setupMocks()
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.assign(navigator, { clipboard: { writeText } })

    renderWithProviders(<ReferralProgram />)
    fireEvent.click(screen.getByText('Копировать код'))

    await waitFor(() => {
      expect(writeText).toHaveBeenCalledWith('ABC123')
    })
  })

  it('copies referral link to clipboard', async () => {
    setupMocks()
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.assign(navigator, { clipboard: { writeText } })

    renderWithProviders(<ReferralProgram />)
    fireEvent.click(screen.getByText('Скопировать ссылку'))

    await waitFor(() => {
      expect(writeText).toHaveBeenCalledWith('https://bani.app/ref/ABC123')
    })
  })

  it('displays invited count', () => {
    setupMocks()
    renderWithProviders(<ReferralProgram />)
    expect(screen.getByText('Приглашено')).toBeInTheDocument()
    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('displays completed bookings count', () => {
    setupMocks()
    renderWithProviders(<ReferralProgram />)
    expect(screen.getByText('Завершённых бронирований')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('displays total earned in stats', () => {
    setupMocks()
    renderWithProviders(<ReferralProgram />)
    const earnedElements = screen.getAllByText('Всего заработано')
    expect(earnedElements.length).toBeGreaterThanOrEqual(1)
    // 300000 kopecks = 3000 rubles
    const priceElements = screen.getAllByText('3000 ₽')
    expect(priceElements.length).toBeGreaterThanOrEqual(1)
  })

  it('displays referral balance', () => {
    setupMocks()
    renderWithProviders(<ReferralProgram />)
    expect(screen.getByText('Реферальный баланс')).toBeInTheDocument()
    expect(screen.getByText('Текущий баланс')).toBeInTheDocument()
    expect(screen.getByText('1500 ₽')).toBeInTheDocument()
  })

  it('displays bonus info text', () => {
    setupMocks()
    renderWithProviders(<ReferralProgram />)
    expect(
      screen.getByText(/вы оба получите бонус 500 ₽/)
    ).toBeInTheDocument()
  })

  it('displays balance usage text', () => {
    setupMocks()
    renderWithProviders(<ReferralProgram />)
    expect(
      screen.getByText(/Реферальный баланс можно использовать при оплате бронирований/)
    ).toBeInTheDocument()
  })

  it('shows loading spinner when data is loading', () => {
    setupMocks({ loading: true })
    renderWithProviders(<ReferralProgram />)
    expect(document.querySelector('.ant-spin-spinning')).toBeInTheDocument()
  })

  it('shows empty state when no referral code', () => {
    setupMocks({ referral: null })
    renderWithProviders(<ReferralProgram />)
    expect(screen.getByText('Реферальная программа пока недоступна')).toBeInTheDocument()
  })

  it('renders zero stats gracefully', () => {
    setupMocks({
      stats: { total_invited: 0, total_completed: 0, total_earned: 0 },
      balance: { balance: 0, total_earned: 0 },
    })
    renderWithProviders(<ReferralProgram />)
    // Should show 0 values without breaking
    const zeroElements = screen.getAllByText('0')
    expect(zeroElements.length).toBeGreaterThanOrEqual(1)
    const zeroPriceElements = screen.getAllByText('0 ₽')
    expect(zeroPriceElements.length).toBeGreaterThanOrEqual(1)
  })
})
