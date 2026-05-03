import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import SubscriptionPage from '@/pages/subscriptions/SubscriptionPage'

vi.mock('@/api/generated/subscriptions/subscriptions', () => ({
  useGetMyBathhousesIdSubscription: vi.fn(),
  usePostMyBathhousesIdSubscription: vi.fn(),
  useDeleteMyBathhousesIdSubscription: vi.fn(),
  useGetMySubscriptions: vi.fn(),
  useGetMyBathhousesIdPromotion: vi.fn(),
  usePostMyBathhousesIdPromotion: vi.fn(),
}))

vi.mock('@/stores/bathhouse', () => ({
  useBathhouseStore: vi.fn(),
}))

import {
  useGetMyBathhousesIdSubscription,
  usePostMyBathhousesIdSubscription,
  useDeleteMyBathhousesIdSubscription,
  useGetMySubscriptions,
  useGetMyBathhousesIdPromotion,
  usePostMyBathhousesIdPromotion,
} from '@/api/generated/subscriptions/subscriptions'
import { useBathhouseStore } from '@/stores/bathhouse'

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

function mockBathhouseStore(id: string | null) {
  vi.mocked(useBathhouseStore).mockImplementation((selector) =>
    (selector as (state: { selectedBathhouseId: string | null }) => unknown)({
      selectedBathhouseId: id,
    }),
  )
}

const mockSubscribeMutation = { mutate: vi.fn(), isPending: false }
const mockCancelMutation = { mutate: vi.fn(), isPending: false }
const mockPromotionMutation = { mutate: vi.fn(), isPending: false }

function setupDefaultMocks() {
  vi.mocked(usePostMyBathhousesIdSubscription).mockReturnValue(
    mockSubscribeMutation as unknown as ReturnType<typeof usePostMyBathhousesIdSubscription>,
  )
  vi.mocked(useDeleteMyBathhousesIdSubscription).mockReturnValue(
    mockCancelMutation as unknown as ReturnType<typeof useDeleteMyBathhousesIdSubscription>,
  )
  vi.mocked(usePostMyBathhousesIdPromotion).mockReturnValue(
    mockPromotionMutation as unknown as ReturnType<typeof usePostMyBathhousesIdPromotion>,
  )
}

const mockSubscription = {
  id: 'sub-1',
  bathhouse_id: 'bath-1',
  owner_id: 'user-1',
  plan: 'premium',
  status: 'active',
  price_kopecks: 99900,
  start_date: '2026-01-01T00:00:00Z',
  end_date: '2026-04-01T00:00:00Z',
  auto_renew: true,
  created_at: '2026-01-01T00:00:00Z',
}

const mockPromotion = {
  id: 'promo-1',
  bathhouse_id: 'bath-1',
  budget_kopecks: 500000,
  spent_kopecks: 125000,
  impression_count: 4500,
  click_count: 120,
  start_date: '2026-03-01T00:00:00Z',
  end_date: '2026-03-31T23:59:59Z',
  status: 'active',
}

describe('SubscriptionPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setupDefaultMocks()
  })

  it('shows prompt when no bathhouse selected', () => {
    mockBathhouseStore(null)
    vi.mocked(useGetMyBathhousesIdSubscription).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdSubscription>)
    vi.mocked(useGetMySubscriptions).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySubscriptions>)
    vi.mocked(useGetMyBathhousesIdPromotion).mockReturnValue({
      data: undefined,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromotion>)

    renderWithProviders(<SubscriptionPage />)

    expect(screen.getByText('Подписки')).toBeInTheDocument()
    expect(screen.getByText('Выберите баню для управления подписками')).toBeInTheDocument()
  })

  it('renders plan cards', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdSubscription).mockReturnValue({
      data: { data: null, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdSubscription>)
    vi.mocked(useGetMySubscriptions).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySubscriptions>)
    vi.mocked(useGetMyBathhousesIdPromotion).mockReturnValue({
      data: undefined,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromotion>)

    renderWithProviders(<SubscriptionPage />)

    expect(screen.getByText('Бесплатный')).toBeInTheDocument()
    expect(screen.getByText('Премиум')).toBeInTheDocument()
    expect(screen.getByText('Продвижение')).toBeInTheDocument()
  })

  it('shows plan features', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdSubscription).mockReturnValue({
      data: { data: null, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdSubscription>)
    vi.mocked(useGetMySubscriptions).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySubscriptions>)
    vi.mocked(useGetMyBathhousesIdPromotion).mockReturnValue({
      data: undefined,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromotion>)

    renderWithProviders(<SubscriptionPage />)

    expect(screen.getByText('Базовый листинг')).toBeInTheDocument()
    expect(screen.getByText('Приоритет в поиске (+10)')).toBeInTheDocument()
    expect(screen.getByText('Промо-кампании')).toBeInTheDocument()
  })

  it('marks current plan as disabled', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdSubscription).mockReturnValue({
      data: { data: mockSubscription, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdSubscription>)
    vi.mocked(useGetMySubscriptions).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySubscriptions>)
    vi.mocked(useGetMyBathhousesIdPromotion).mockReturnValue({
      data: undefined,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromotion>)

    renderWithProviders(<SubscriptionPage />)

    expect(screen.getByText('Текущий план')).toBeInTheDocument()
  })

  it('shows current subscription details', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdSubscription).mockReturnValue({
      data: { data: mockSubscription, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdSubscription>)
    vi.mocked(useGetMySubscriptions).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySubscriptions>)
    vi.mocked(useGetMyBathhousesIdPromotion).mockReturnValue({
      data: undefined,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromotion>)

    renderWithProviders(<SubscriptionPage />)

    expect(screen.getByText('Текущая подписка')).toBeInTheDocument()
    expect(screen.getByText('999 ₽')).toBeInTheDocument()
  })

  it('calls subscribe mutation when plan selected', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdSubscription).mockReturnValue({
      data: { data: null, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdSubscription>)
    vi.mocked(useGetMySubscriptions).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySubscriptions>)
    vi.mocked(useGetMyBathhousesIdPromotion).mockReturnValue({
      data: undefined,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromotion>)

    renderWithProviders(<SubscriptionPage />)

    const selectButtons = screen.getAllByText('Выбрать')
    fireEvent.click(selectButtons[0]!)

    expect(mockSubscribeMutation.mutate).toHaveBeenCalledWith({
      id: 'bath-1',
      data: { plan: 'free' },
    })
  })

  it('shows cancel button when auto_renew is true', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdSubscription).mockReturnValue({
      data: { data: mockSubscription, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdSubscription>)
    vi.mocked(useGetMySubscriptions).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySubscriptions>)
    vi.mocked(useGetMyBathhousesIdPromotion).mockReturnValue({
      data: undefined,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromotion>)

    renderWithProviders(<SubscriptionPage />)

    expect(screen.getByText('Отменить подписку')).toBeInTheDocument()
  })

  it('shows promotion stats for promoted plan', () => {
    mockBathhouseStore('bath-1')
    const promotedSub = { ...mockSubscription, plan: 'promoted', price_kopecks: 299900 }
    vi.mocked(useGetMyBathhousesIdSubscription).mockReturnValue({
      data: { data: promotedSub, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdSubscription>)
    vi.mocked(useGetMySubscriptions).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySubscriptions>)
    vi.mocked(useGetMyBathhousesIdPromotion).mockReturnValue({
      data: { data: mockPromotion, success: true },
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromotion>)

    renderWithProviders(<SubscriptionPage />)

    expect(screen.getAllByText('Промо-кампании').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Создать кампанию')).toBeInTheDocument()
    expect(screen.getByText('Показы')).toBeInTheDocument()
    expect(screen.getByText('Клики')).toBeInTheDocument()
  })

  it('shows all subscriptions table', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdSubscription).mockReturnValue({
      data: { data: null, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdSubscription>)
    vi.mocked(useGetMySubscriptions).mockReturnValue({
      data: {
        data: [mockSubscription],
        success: true,
        meta: { total_count: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySubscriptions>)
    vi.mocked(useGetMyBathhousesIdPromotion).mockReturnValue({
      data: undefined,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromotion>)

    renderWithProviders(<SubscriptionPage />)

    expect(screen.getByText('Все подписки')).toBeInTheDocument()
    expect(screen.getByText('Активна')).toBeInTheDocument()
  })

  it('shows empty state for subscriptions table', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdSubscription).mockReturnValue({
      data: { data: null, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdSubscription>)
    vi.mocked(useGetMySubscriptions).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySubscriptions>)
    vi.mocked(useGetMyBathhousesIdPromotion).mockReturnValue({
      data: undefined,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromotion>)

    renderWithProviders(<SubscriptionPage />)

    expect(screen.getByText('Нет подписок')).toBeInTheDocument()
  })

  it('opens promotion creation modal', () => {
    mockBathhouseStore('bath-1')
    const promotedSub = { ...mockSubscription, plan: 'promoted' }
    vi.mocked(useGetMyBathhousesIdSubscription).mockReturnValue({
      data: { data: promotedSub, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdSubscription>)
    vi.mocked(useGetMySubscriptions).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySubscriptions>)
    vi.mocked(useGetMyBathhousesIdPromotion).mockReturnValue({
      data: undefined,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPromotion>)

    renderWithProviders(<SubscriptionPage />)

    fireEvent.click(screen.getByText('Создать кампанию'))

    expect(screen.getByText('Новая промо-кампания')).toBeInTheDocument()
    expect(screen.getByLabelText('Бюджет (₽)')).toBeInTheDocument()
  })
})
