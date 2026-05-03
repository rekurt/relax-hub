import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi } from 'vitest'
import SubscriptionManagement from '@/pages/admin/SubscriptionManagement'

vi.mock('@/api/generated/subscriptions/subscriptions', () => ({
  useGetMySubscriptions: vi.fn(),
  getGetMySubscriptionsQueryKey: vi.fn(() => ['subscriptions']),
  useDeleteMyBathhousesIdSubscription: vi.fn(),
  usePostMyBathhousesIdSubscription: vi.fn(),
}))

import {
  useGetMySubscriptions,
  useDeleteMyBathhousesIdSubscription,
  usePostMyBathhousesIdSubscription,
} from '@/api/generated/subscriptions/subscriptions'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/subscriptions']}>
            <Routes>
              <Route path="/admin/subscriptions" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockSubscriptions = [
  {
    id: 'sub-001',
    bathhouse_id: 'bath-001',
    owner_id: 'owner-001',
    plan: 'premium',
    status: 'active',
    price_kopecks: 99900,
    start_date: '2026-01-01T00:00:00Z',
    end_date: '2026-12-31T23:59:59Z',
    auto_renew: true,
  },
  {
    id: 'sub-002',
    bathhouse_id: 'bath-002',
    owner_id: 'owner-002',
    plan: 'promoted',
    status: 'expired',
    price_kopecks: 199900,
    start_date: '2025-01-01T00:00:00Z',
    end_date: '2025-12-31T23:59:59Z',
    auto_renew: false,
  },
  {
    id: 'sub-003',
    bathhouse_id: 'bath-003',
    owner_id: 'owner-003',
    plan: 'free',
    status: 'active',
    price_kopecks: 0,
    start_date: '2026-03-01T00:00:00Z',
    end_date: '2026-06-01T00:00:00Z',
    auto_renew: false,
  },
]

function setupMocks() {
  vi.mocked(useGetMySubscriptions).mockReturnValue({
    data: { data: mockSubscriptions, meta: { total_count: 3 }, success: true },
    isLoading: false,
  } as ReturnType<typeof useGetMySubscriptions>)
  vi.mocked(useDeleteMyBathhousesIdSubscription).mockReturnValue({
    mutateAsync: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof useDeleteMyBathhousesIdSubscription>)
  vi.mocked(usePostMyBathhousesIdSubscription).mockReturnValue({
    mutateAsync: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof usePostMyBathhousesIdSubscription>)
}

describe('SubscriptionManagement', () => {
  it('renders title and summary cards', () => {
    setupMocks()
    renderWithProviders(<SubscriptionManagement />)

    expect(screen.getByText('Управление подписками')).toBeInTheDocument()
    expect(screen.getByText('Активных')).toBeInTheDocument()
    // "Премиум" appears in both card and table
    expect(screen.getAllByText('Премиум').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Продвинутых')).toBeInTheDocument()
  })

  it('renders subscription plans in table', () => {
    setupMocks()
    renderWithProviders(<SubscriptionManagement />)

    expect(screen.getAllByText('Премиум').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Продвинутый')).toBeInTheDocument()
    expect(screen.getByText('Бесплатный')).toBeInTheDocument()
  })

  it('shows active count correctly', () => {
    setupMocks()
    renderWithProviders(<SubscriptionManagement />)

    // 2 active subscriptions (sub-001 and sub-003)
    expect(screen.getByText('2')).toBeInTheDocument()
  })

  it('renders status tags', () => {
    setupMocks()
    renderWithProviders(<SubscriptionManagement />)

    expect(screen.getAllByText('Активна').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Истекла')).toBeInTheDocument()
  })

  it('shows deactivate button for active subscriptions', () => {
    setupMocks()
    renderWithProviders(<SubscriptionManagement />)

    const deactivateButtons = screen.getAllByText('Деактивировать')
    expect(deactivateButtons.length).toBe(2)
  })

  it('shows activate button for non-active subscriptions', () => {
    setupMocks()
    renderWithProviders(<SubscriptionManagement />)

    expect(screen.getByText('Активировать')).toBeInTheDocument()
  })

  it('renders refresh button', () => {
    setupMocks()
    renderWithProviders(<SubscriptionManagement />)

    expect(screen.getByText('Обновить')).toBeInTheDocument()
  })

  it('shows loading state', () => {
    vi.mocked(useGetMySubscriptions).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as ReturnType<typeof useGetMySubscriptions>)
    vi.mocked(useDeleteMyBathhousesIdSubscription).mockReturnValue({
      mutateAsync: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof useDeleteMyBathhousesIdSubscription>)
    vi.mocked(usePostMyBathhousesIdSubscription).mockReturnValue({
      mutateAsync: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostMyBathhousesIdSubscription>)

    renderWithProviders(<SubscriptionManagement />)
    expect(screen.getByText('Управление подписками')).toBeInTheDocument()
  })

  it('renders search input', () => {
    setupMocks()
    renderWithProviders(<SubscriptionManagement />)

    expect(screen.getByPlaceholderText('Поиск по ID бани или владельца')).toBeInTheDocument()
  })

  it('shows deactivate confirmation on click', () => {
    setupMocks()
    renderWithProviders(<SubscriptionManagement />)

    const deactivateButtons = screen.getAllByText('Деактивировать')
    fireEvent.click(deactivateButtons[0]!)

    expect(screen.getAllByText('Деактивировать подписку?').length).toBeGreaterThanOrEqual(1)
  })
})
