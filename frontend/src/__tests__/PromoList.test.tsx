import { render, screen, fireEvent, act, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import PromoList from '@/pages/promo/PromoList'

vi.mock('@/api/generated/promo-codes/promo-codes', () => ({
  useGetMyBathhousesIdPromoCodes: vi.fn(),
  usePostMyBathhousesIdPromoCodes: vi.fn(),
  useDeletePromoCodesId: vi.fn(),
}))

vi.mock('@/stores/bathhouse', () => ({
  useBathhouseStore: vi.fn(),
}))

import {
  useGetMyBathhousesIdPromoCodes,
  usePostMyBathhousesIdPromoCodes,
  useDeletePromoCodesId,
} from '@/api/generated/promo-codes/promo-codes'
import { useBathhouseStore } from '@/stores/bathhouse'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU} theme={{ token: { motion: false } }}>
        <AntApp>
          <MemoryRouter>{ui}</MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockPromos = [
  {
    id: 'promo-1',
    code: 'SUMMER20',
    type: 'percentage',
    value: 20,
    bathhouse_id: 'bath-1',
    creator_id: 'user-1',
    is_active: true,
    max_uses: 100,
    current_uses: 35,
    min_amount: 500000,
    valid_from: '2026-01-01T00:00:00Z',
    valid_until: '2026-12-31T23:59:59Z',
    created_at: '2026-01-01T10:00:00Z',
  },
  {
    id: 'promo-2',
    code: 'FIXED500',
    type: 'fixed_amount',
    value: 50000,
    bathhouse_id: 'bath-1',
    creator_id: 'user-1',
    is_active: true,
    max_uses: 50,
    current_uses: 50,
    min_amount: 0,
    valid_from: null,
    valid_until: null,
    created_at: '2026-02-01T10:00:00Z',
  },
  {
    id: 'promo-3',
    code: 'FREEHOUR',
    type: 'free_hour',
    value: 1,
    bathhouse_id: 'bath-1',
    creator_id: 'user-1',
    is_active: false,
    max_uses: 0,
    current_uses: 10,
    min_amount: 0,
    valid_from: '2025-01-01T00:00:00Z',
    valid_until: '2025-06-30T23:59:59Z',
    created_at: '2025-01-01T10:00:00Z',
  },
]

const mockCreateMutation = { mutate: vi.fn(), isPending: false }
const mockDeleteMutation = { mutate: vi.fn(), isPending: false }

function mockBathhouseStore(id: string | null) {
  vi.mocked(useBathhouseStore).mockImplementation((selector) =>
    (selector as (state: { selectedBathhouseId: string | null }) => unknown)({
      selectedBathhouseId: id,
    }),
  )
}

function setupDefaultMocks() {
  vi.mocked(usePostMyBathhousesIdPromoCodes).mockReturnValue(
    mockCreateMutation as unknown as ReturnType<typeof usePostMyBathhousesIdPromoCodes>,
  )
  vi.mocked(useDeletePromoCodesId).mockReturnValue(
    mockDeleteMutation as unknown as ReturnType<typeof useDeletePromoCodesId>,
  )
}

describe('PromoList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setupDefaultMocks()
  })

  it('shows prompt when no bathhouse selected', () => {
    mockBathhouseStore(null)
    vi.mocked(useGetMyBathhousesIdPromoCodes).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromoCodes>)

    renderWithProviders(<PromoList />)

    expect(screen.getByText('Промокоды')).toBeInTheDocument()
    expect(screen.getByText('Выберите баню для управления промокодами')).toBeInTheDocument()
  })

  it('renders promo codes table with data', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPromoCodes).mockReturnValue({
      data: { data: mockPromos, success: true, meta: { total_count: 3 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromoCodes>)

    renderWithProviders(<PromoList />)

    expect(screen.getByText('SUMMER20')).toBeInTheDocument()
    expect(screen.getByText('FIXED500')).toBeInTheDocument()
    expect(screen.getByText('FREEHOUR')).toBeInTheDocument()
  })

  it('shows type tags with correct labels', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPromoCodes).mockReturnValue({
      data: { data: mockPromos, success: true, meta: { total_count: 3 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromoCodes>)

    renderWithProviders(<PromoList />)

    expect(screen.getByText('Процент')).toBeInTheDocument()
    expect(screen.getByText('Фиксированная сумма')).toBeInTheDocument()
    expect(screen.getByText('Бесплатный час')).toBeInTheDocument()
  })

  it('shows formatted promo values', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPromoCodes).mockReturnValue({
      data: { data: mockPromos, success: true, meta: { total_count: 3 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromoCodes>)

    renderWithProviders(<PromoList />)

    expect(screen.getByText('20%')).toBeInTheDocument()
    expect(screen.getByText('500 ₽')).toBeInTheDocument()
    expect(screen.getByText('1 ч.')).toBeInTheDocument()
  })

  it('shows usage counts', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPromoCodes).mockReturnValue({
      data: { data: mockPromos, success: true, meta: { total_count: 3 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromoCodes>)

    renderWithProviders(<PromoList />)

    expect(screen.getByText('35 / 100')).toBeInTheDocument()
    expect(screen.getByText('50 / 50')).toBeInTheDocument()
    expect(screen.getByText('10 / ∞')).toBeInTheDocument()
  })

  it('shows correct statuses', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPromoCodes).mockReturnValue({
      data: { data: mockPromos, success: true, meta: { total_count: 3 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromoCodes>)

    renderWithProviders(<PromoList />)

    expect(screen.getByText('Активен')).toBeInTheDocument()
    expect(screen.getByText('Исчерпан')).toBeInTheDocument()
    expect(screen.getByText('Неактивен')).toBeInTheDocument()
  })

  it('shows period info', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPromoCodes).mockReturnValue({
      data: { data: mockPromos, success: true, meta: { total_count: 3 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromoCodes>)

    renderWithProviders(<PromoList />)

    expect(screen.getByText(/01\.01\.2026/)).toBeInTheDocument()
  })

  it('shows "Бессрочно" when no validity dates', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPromoCodes).mockReturnValue({
      data: { data: [mockPromos[1]], success: true, meta: { total_count: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromoCodes>)

    renderWithProviders(<PromoList />)

    expect(screen.getByText('Бессрочно')).toBeInTheDocument()
  })

  it('shows min amount', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPromoCodes).mockReturnValue({
      data: { data: [mockPromos[0]], success: true, meta: { total_count: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromoCodes>)

    renderWithProviders(<PromoList />)

    expect(screen.getByText('5000 ₽')).toBeInTheDocument()
  })

  it('shows empty state when no promo codes', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPromoCodes).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromoCodes>)

    renderWithProviders(<PromoList />)

    expect(screen.getByText(/Нет промокодов/)).toBeInTheDocument()
  })

  it('opens create modal when button clicked', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPromoCodes).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromoCodes>)

    renderWithProviders(<PromoList />)

    fireEvent.click(screen.getByText('Создать промокод'))

    expect(screen.getByText('Новый промокод')).toBeInTheDocument()
    expect(screen.getByLabelText('Код промокода')).toBeInTheDocument()
    expect(screen.getAllByText('Тип скидки').length).toBeGreaterThanOrEqual(2)
  })

  it('calls delete mutation when confirmed', async () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPromoCodes).mockReturnValue({
      data: { data: [mockPromos[0]], success: true, meta: { total_count: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromoCodes>)

    renderWithProviders(<PromoList />)

    const deleteButtons = screen.getAllByRole('button').filter(
      (btn) => btn.querySelector('.anticon-delete'),
    )
    expect(deleteButtons.length).toBeGreaterThan(0)
    await act(async () => {
      fireEvent.click(deleteButtons[0]!)
    })

    await waitFor(() => {
      expect(screen.getByText('Деактивировать промокод?')).toBeInTheDocument()
    }, { timeout: 10000 })

    const confirmBtn = screen.getByRole('button', { name: 'Деактивировать' })
    await act(async () => {
      fireEvent.click(confirmBtn)
    })

    expect(mockDeleteMutation.mutate).toHaveBeenCalledWith({ id: 'promo-1' })
  }, 15000)

  it('shows copy buttons for each promo code', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPromoCodes).mockReturnValue({
      data: { data: mockPromos, success: true, meta: { total_count: 3 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromoCodes>)

    renderWithProviders(<PromoList />)

    const copyButtons = screen.getAllByRole('button').filter(
      (btn) => btn.querySelector('.anticon-copy'),
    )
    expect(copyButtons.length).toBe(3)
  })

  it('shows create button', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPromoCodes).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromoCodes>)

    renderWithProviders(<PromoList />)

    expect(screen.getByText('Создать промокод')).toBeInTheDocument()
  })
})
