import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import AntiFraudDashboard from '@/pages/admin/AntiFraudDashboard'

vi.mock('@/api/generated/admin-antifraud/admin-antifraud', () => ({
  useGetAdminAntifraudFlags: vi.fn(),
  usePatchAdminAntifraudFlagsId: vi.fn(),
  useGetAdminChatFiltered: vi.fn(),
}))

import {
  useGetAdminAntifraudFlags,
  usePatchAdminAntifraudFlagsId,
} from '@/api/generated/admin-antifraud/admin-antifraud'

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

const mockFlags = [
  {
    id: 'f1',
    user_id: 'user-abc-123',
    rule: 'multi_card_topup',
    severity: 'high',
    status: 'pending',
    action: 'freeze_wallet',
    details: { card_count: 5, period_hours: 24 },
    created_at: '2026-03-20T12:00:00Z',
    reviewed_at: null,
    reviewed_by: null,
  },
  {
    id: 'f2',
    user_id: 'user-def-456',
    rule: 'self_booking',
    severity: 'medium',
    status: 'reviewed',
    action: 'flag',
    details: { booking_id: 'b-123' },
    created_at: '2026-03-19T10:00:00Z',
    reviewed_at: '2026-03-19T15:00:00Z',
    reviewed_by: 'admin-1',
  },
  {
    id: 'f3',
    user_id: 'user-ghi-789',
    rule: 'rapid_bookings',
    severity: 'low',
    status: 'dismissed',
    action: 'notify_admin',
    details: { booking_count: 3 },
    created_at: '2026-03-18T08:00:00Z',
    reviewed_at: '2026-03-18T09:00:00Z',
    reviewed_by: 'admin-2',
  },
]

describe('AntiFraudDashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks()

    vi.mocked(usePatchAdminAntifraudFlagsId).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePatchAdminAntifraudFlagsId>)
  })

  it('renders page title', () => {
    vi.mocked(useGetAdminAntifraudFlags).mockReturnValue({
      data: { data: mockFlags, success: true, meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAntifraudFlags>)

    renderWithProviders(<AntiFraudDashboard />)

    expect(screen.getByText('Антифрод')).toBeInTheDocument()
  })

  it('renders fraud flags in table', () => {
    vi.mocked(useGetAdminAntifraudFlags).mockReturnValue({
      data: { data: mockFlags, success: true, meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAntifraudFlags>)

    renderWithProviders(<AntiFraudDashboard />)

    expect(screen.getByText('Мульти-карта пополнение')).toBeInTheDocument()
    expect(screen.getByText('Самобронирование')).toBeInTheDocument()
    expect(screen.getByText('Быстрые бронирования')).toBeInTheDocument()
  })

  it('renders severity tags', () => {
    vi.mocked(useGetAdminAntifraudFlags).mockReturnValue({
      data: { data: mockFlags, success: true, meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAntifraudFlags>)

    renderWithProviders(<AntiFraudDashboard />)

    expect(screen.getByText('Высокий')).toBeInTheDocument()
    expect(screen.getByText('Средний')).toBeInTheDocument()
    expect(screen.getByText('Низкий')).toBeInTheDocument()
  })

  it('renders status filter segmented control', () => {
    vi.mocked(useGetAdminAntifraudFlags).mockReturnValue({
      data: { data: mockFlags, success: true, meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAntifraudFlags>)

    renderWithProviders(<AntiFraudDashboard />)

    expect(screen.getByText('Все')).toBeInTheDocument()
    const pendingElements = screen.getAllByText('Ожидают')
    expect(pendingElements.length).toBeGreaterThanOrEqual(1)
  })

  it('renders stats cards', () => {
    vi.mocked(useGetAdminAntifraudFlags).mockReturnValue({
      data: { data: mockFlags, success: true, meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAntifraudFlags>)

    renderWithProviders(<AntiFraudDashboard />)

    const pendingStatElements = screen.getAllByText('На рассмотрении')
    expect(pendingStatElements.length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Всего флагов')).toBeInTheDocument()
  })

  it('shows approve and dismiss buttons for pending flags', () => {
    vi.mocked(useGetAdminAntifraudFlags).mockReturnValue({
      data: { data: mockFlags, success: true, meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAntifraudFlags>)

    renderWithProviders(<AntiFraudDashboard />)

    const approveButtons = screen.getAllByText('Подтвердить')
    expect(approveButtons.length).toBeGreaterThanOrEqual(1)
    const dismissButtons = screen.getAllByText('Отклонить')
    expect(dismissButtons.length).toBeGreaterThanOrEqual(1)
  })

  it('opens detail drawer on click', () => {
    vi.mocked(useGetAdminAntifraudFlags).mockReturnValue({
      data: { data: mockFlags, success: true, meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAntifraudFlags>)

    renderWithProviders(<AntiFraudDashboard />)

    const detailButtons = screen.getAllByText('Детали')
    fireEvent.click(detailButtons[0])

    expect(screen.getByText('Детали подозрительной операции')).toBeInTheDocument()
  })

  it('shows detail data in drawer', () => {
    vi.mocked(useGetAdminAntifraudFlags).mockReturnValue({
      data: { data: mockFlags, success: true, meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAntifraudFlags>)

    renderWithProviders(<AntiFraudDashboard />)

    const detailButtons = screen.getAllByText('Детали')
    fireEvent.click(detailButtons[0])

    const walletFreezeTexts = screen.getAllByText('Заморозка кошелька')
    expect(walletFreezeTexts.length).toBeGreaterThanOrEqual(1)
  })

  it('renders rule filter dropdown', () => {
    vi.mocked(useGetAdminAntifraudFlags).mockReturnValue({
      data: { data: mockFlags, success: true, meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAntifraudFlags>)

    renderWithProviders(<AntiFraudDashboard />)

    expect(screen.getByText('Фильтр по правилу')).toBeInTheDocument()
  })
})
