import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import BookingList from '@/pages/bookings/BookingList'

vi.mock('@/api/generated/bookings/bookings', () => ({
  useGetBathhousesIdBookings: vi.fn(),
  usePatchBookingsIdConfirm: vi.fn(),
  usePatchBookingsIdReject: vi.fn(),
  usePatchBookingsIdCancel: vi.fn(),
  usePatchBookingsIdComplete: vi.fn(),
  usePatchBookingsIdCheckIn: vi.fn(),
  usePatchBookingsIdCheckOut: vi.fn(),
}))

vi.mock('@/api/generated/payments/payments', () => ({
  useGetBookingsIdPayment: vi.fn().mockReturnValue({
    data: undefined,
    isLoading: false,
  }),
  usePostBookingsIdPay: vi.fn(),
}))

vi.mock('@/stores/bathhouse', () => ({
  useBathhouseStore: vi.fn(),
}))

import {
  useGetBathhousesIdBookings,
  usePatchBookingsIdConfirm,
  usePatchBookingsIdReject,
  usePatchBookingsIdCancel,
  usePatchBookingsIdComplete,
  usePatchBookingsIdCheckIn,
  usePatchBookingsIdCheckOut,
} from '@/api/generated/bookings/bookings'
import { usePostBookingsIdPay } from '@/api/generated/payments/payments'
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

const mockBookings = [
  {
    id: 'booking-1',
    user_id: 'user-abc-123-def',
    guest_count: 4,
    total_price: 300000,
    original_price: 350000,
    status: 'pending',
    start_time: '2026-03-20T14:00:00Z',
    end_time: '2026-03-20T16:00:00Z',
    comment: 'Праздник',
    created_at: '2026-03-15T10:00:00Z',
  },
  {
    id: 'booking-2',
    user_id: 'user-xyz-456-ghi',
    guest_count: 2,
    total_price: 150000,
    original_price: 150000,
    status: 'confirmed',
    start_time: '2026-03-21T10:00:00Z',
    end_time: '2026-03-21T12:00:00Z',
    comment: '',
    created_at: '2026-03-14T08:00:00Z',
  },
  {
    id: 'booking-3',
    user_id: 'user-qwe-789-rty',
    guest_count: 6,
    total_price: 500000,
    original_price: 500000,
    status: 'completed',
    start_time: '2026-03-10T18:00:00Z',
    end_time: '2026-03-10T21:00:00Z',
    comment: '',
    created_at: '2026-03-09T12:00:00Z',
  },
]

const mockMutation = { mutate: vi.fn(), isPending: false }

function mockBathhouseStore(id: string | null) {
  vi.mocked(useBathhouseStore).mockImplementation((selector) =>
    (selector as (state: { selectedBathhouseId: string | null }) => unknown)({
      selectedBathhouseId: id,
    }),
  )
}

describe('BookingList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(usePatchBookingsIdConfirm).mockReturnValue(mockMutation as unknown as ReturnType<typeof usePatchBookingsIdConfirm>)
    vi.mocked(usePatchBookingsIdReject).mockReturnValue(mockMutation as unknown as ReturnType<typeof usePatchBookingsIdReject>)
    vi.mocked(usePatchBookingsIdCancel).mockReturnValue(mockMutation as unknown as ReturnType<typeof usePatchBookingsIdCancel>)
    vi.mocked(usePatchBookingsIdComplete).mockReturnValue(mockMutation as unknown as ReturnType<typeof usePatchBookingsIdComplete>)
    vi.mocked(usePatchBookingsIdCheckIn).mockReturnValue(mockMutation as unknown as ReturnType<typeof usePatchBookingsIdCheckIn>)
    vi.mocked(usePatchBookingsIdCheckOut).mockReturnValue(mockMutation as unknown as ReturnType<typeof usePatchBookingsIdCheckOut>)
    vi.mocked(usePostBookingsIdPay).mockReturnValue(mockMutation as unknown as ReturnType<typeof usePostBookingsIdPay>)
  })

  it('shows prompt when no bathhouse selected', () => {
    mockBathhouseStore(null)
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)

    expect(screen.getByText('Бронирования')).toBeInTheDocument()
    expect(screen.getByText('Выберите баню для просмотра бронирований')).toBeInTheDocument()
  })

  it('renders bookings table with data', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: mockBookings,
        success: true,
        meta: { total_count: 3, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)

    expect(screen.getByText('Бронирования')).toBeInTheDocument()
    expect(screen.getByText('Ожидает')).toBeInTheDocument()
    expect(screen.getByText('Подтверждено')).toBeInTheDocument()
    expect(screen.getByText('Завершено')).toBeInTheDocument()
    expect(screen.getByText('Всего: 3')).toBeInTheDocument()
  })

  it('renders status tags with correct colors', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: mockBookings,
        success: true,
        meta: { total_count: 3, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)

    const pendingTag = screen.getByText('Ожидает')
    expect(pendingTag.closest('.ant-tag')).toHaveClass('ant-tag-orange')

    const confirmedTag = screen.getByText('Подтверждено')
    expect(confirmedTag.closest('.ant-tag')).toHaveClass('ant-tag-blue')

    const completedTag = screen.getByText('Завершено')
    expect(completedTag.closest('.ant-tag')).toHaveClass('ant-tag-green')
  })

  it('shows confirm/reject actions for pending bookings', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [mockBookings[0]],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)

    expect(screen.getByText('Подтвердить')).toBeInTheDocument()
    expect(screen.getByText('Отклонить')).toBeInTheDocument()
  })

  it('shows complete/cancel actions for confirmed bookings with check-in and check-out', () => {
    mockBathhouseStore('bathhouse-1')
    const checkedOutBooking = {
      ...mockBookings[1],
      checked_in_at: '2026-03-21T10:00:00Z',
      checked_out_at: '2026-03-21T12:00:00Z',
    }
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [checkedOutBooking],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)

    expect(screen.getByText('Завершить')).toBeInTheDocument()
    expect(screen.getByText('Отменить')).toBeInTheDocument()
  })

  it('calls confirm mutation when confirm button clicked', () => {
    const confirmMutate = vi.fn()
    vi.mocked(usePatchBookingsIdConfirm).mockReturnValue({
      mutate: confirmMutate,
      isPending: false,
    } as unknown as ReturnType<typeof usePatchBookingsIdConfirm>)
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [mockBookings[0]],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)
    fireEvent.click(screen.getByText('Подтвердить'))

    expect(confirmMutate).toHaveBeenCalledWith({ id: 'booking-1' })
  })

  it('shows details button and opens modal', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [mockBookings[0]],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)

    fireEvent.click(screen.getByText('Детали'))
    expect(screen.getByText('Детали бронирования')).toBeInTheDocument()
  })

  it('shows empty state when no bookings', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [],
        success: true,
        meta: { total_count: 0, page: 0, page_size: 10, total_pages: 0 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)

    expect(screen.getByText('Пока нет бронирований. Убедитесь, что ваш объект активен и заполнен')).toBeInTheDocument()
  })

  it('shows loading state', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)

    expect(screen.getByText('Бронирования')).toBeInTheDocument()
  })

  it('has status filter dropdown', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: mockBookings,
        success: true,
        meta: { total_count: 3, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)

    expect(screen.getByText('Все статусы')).toBeInTheDocument()
  })

  it('formats prices correctly', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [mockBookings[0]],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)

    expect(screen.getByText('3000 ₽')).toBeInTheDocument()
  })

  it('shows payment status column with "Не оплачено" for no payment', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [{ ...mockBookings[0], payment_status: undefined }],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)
    expect(screen.getByText('Не оплачено')).toBeInTheDocument()
  })

  it('shows "Оплачено" tag for succeeded payment', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [{ ...mockBookings[1], payment_status: 'succeeded' }],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)
    expect(screen.getByText('Оплачено')).toBeInTheDocument()
  })

  it('shows pay button for confirmed booking without payment', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [{ ...mockBookings[1], payment_status: undefined }],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)
    expect(screen.getByText('Оплатить')).toBeInTheDocument()
  })

  it('hides pay button for confirmed booking with succeeded payment', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [{ ...mockBookings[1], payment_status: 'succeeded' }],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)
    expect(screen.queryByText('Оплатить')).not.toBeInTheDocument()
  })

  it('shows check-in button for confirmed booking without check-in', () => {
    mockBathhouseStore('bathhouse-1')
    const now = new Date()
    const startTime = new Date(now.getTime() + 5 * 60 * 1000).toISOString()
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [{ ...mockBookings[1], start_time: startTime, checked_in_at: undefined }],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)
    expect(screen.getByText('Гость прибыл')).toBeInTheDocument()
  })

  it('shows check-out button for confirmed booking with check-in but no check-out', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [{
          ...mockBookings[1],
          checked_in_at: '2026-03-21T10:00:00Z',
          checked_out_at: undefined,
        }],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)
    expect(screen.getByText('Гость ушёл')).toBeInTheDocument()
  })

  it('hides check-in and check-out when both are done', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [{
          ...mockBookings[1],
          checked_in_at: '2026-03-21T10:00:00Z',
          checked_out_at: '2026-03-21T12:00:00Z',
        }],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)
    expect(screen.queryByText('Гость прибыл')).not.toBeInTheDocument()
    expect(screen.queryByText('Гость ушёл')).not.toBeInTheDocument()
  })

  it('does not show check-in button when guest already checked in', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [{
          ...mockBookings[1],
          checked_in_at: '2026-03-21T10:00:00Z',
          checked_out_at: undefined,
        }],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)
    expect(screen.queryByText('Гость прибыл')).not.toBeInTheDocument()
    expect(screen.getByText('Гость ушёл')).toBeInTheDocument()
  })

  it('calls check-in mutation when button clicked', () => {
    const checkInMutate = vi.fn()
    vi.mocked(usePatchBookingsIdCheckIn).mockReturnValue({
      mutate: checkInMutate,
      isPending: false,
    } as unknown as ReturnType<typeof usePatchBookingsIdCheckIn>)
    mockBathhouseStore('bathhouse-1')
    const now = new Date()
    const startTime = new Date(now.getTime() + 5 * 60 * 1000).toISOString()
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [{ ...mockBookings[1], start_time: startTime, checked_in_at: undefined }],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)
    fireEvent.click(screen.getByText('Гость прибыл'))

    expect(checkInMutate).toHaveBeenCalledWith({ id: 'booking-2' })
  })

  it('shows extension request button for confirmed bookings', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [mockBookings[1]],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)

    expect(screen.getByText('Продление')).toBeInTheDocument()
  })

  it('does not show extension request button for pending bookings', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [mockBookings[0]],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)

    expect(screen.queryByText('Продление')).not.toBeInTheDocument()
  })

  it('calls pay mutation when pay button clicked', () => {
    const payMutate = vi.fn()
    vi.mocked(usePostBookingsIdPay).mockReturnValue({
      mutate: payMutate,
      isPending: false,
    } as unknown as ReturnType<typeof usePostBookingsIdPay>)
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: {
        data: [{ ...mockBookings[1], payment_status: 'pending' }],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)

    renderWithProviders(<BookingList />)
    fireEvent.click(screen.getByText('Оплатить'))

    expect(payMutate).toHaveBeenCalledWith({ id: 'booking-2', data: {} })
  })
})
