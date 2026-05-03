import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi } from 'vitest'
import BookingDetails from '@/pages/bookings/BookingDetails'
import type { InternalHandlerBookingResponse } from '@/api/generated/model'

vi.mock('@/api/generated/payments/payments', () => ({
  useGetBookingsIdPayment: vi.fn(),
}))

import { useGetBookingsIdPayment } from '@/api/generated/payments/payments'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>{ui}</AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockBooking: InternalHandlerBookingResponse = {
  id: 'b1a2c3d4-5678-9abc-def0-123456789abc',
  user_id: 'user-abc-123',
  guest_count: 4,
  total_price: 300000,
  original_price: 350000,
  promo_discount: 50000,
  loyalty_discount: 0,
  certificate_discount: 0,
  points_spent: 0,
  referral_bonus_used: 0,
  earned_points: 100,
  status: 'confirmed',
  start_time: '2026-03-20T14:00:00Z',
  end_time: '2026-03-20T16:00:00Z',
  comment: 'Праздник',
  created_at: '2026-03-15T10:00:00Z',
}

const mockPayment = {
  id: 'p9z8y7x6-5432-1abc-def0-123456789abc',
  amount: 295000,
  status: 'succeeded',
  provider: 'yookassa',
  refund_amount: 0,
  created_at: '2026-03-15T10:05:00Z',
}

describe('BookingDetails', () => {
  it('does not render when booking is null', () => {
    vi.mocked(useGetBookingsIdPayment).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

    renderWithProviders(
      <BookingDetails booking={null} onClose={vi.fn()} />,
    )

    expect(screen.queryByText('Детали бронирования')).not.toBeInTheDocument()
  })

  it('renders booking details with all fields', () => {
    vi.mocked(useGetBookingsIdPayment).mockReturnValue({
      data: { data: mockPayment, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

    renderWithProviders(
      <BookingDetails booking={mockBooking} onClose={vi.fn()} />,
    )

    expect(screen.getByText('Детали бронирования')).toBeInTheDocument()
    expect(screen.getByText('b1a2c3d4...')).toBeInTheDocument()
    expect(screen.getByText('Подтверждено')).toBeInTheDocument()
    expect(screen.getByText('4')).toBeInTheDocument()
    expect(screen.getByText('Праздник')).toBeInTheDocument()
    expect(screen.getByText('3500 ₽')).toBeInTheDocument()
    expect(screen.getByText('-500 ₽')).toBeInTheDocument()
    expect(screen.getByText('3000 ₽')).toBeInTheDocument()
    expect(screen.getByText('+100')).toBeInTheDocument()
  })

  it('renders payment information', () => {
    vi.mocked(useGetBookingsIdPayment).mockReturnValue({
      data: { data: mockPayment, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

    renderWithProviders(
      <BookingDetails booking={mockBooking} onClose={vi.fn()} />,
    )

    expect(screen.getByText('Платёж')).toBeInTheDocument()
    expect(screen.getByText('p9z8y7x6...')).toBeInTheDocument()
    expect(screen.getByText('Оплачено')).toBeInTheDocument()
    expect(screen.getByText('2950 ₽')).toBeInTheDocument()
    expect(screen.getByText('yookassa')).toBeInTheDocument()
  })

  it('shows no payment message when payment not found', () => {
    vi.mocked(useGetBookingsIdPayment).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

    renderWithProviders(
      <BookingDetails booking={mockBooking} onClose={vi.fn()} />,
    )

    expect(screen.getByText('Платёж не найден')).toBeInTheDocument()
  })

  it('shows loading spinner for payment', () => {
    vi.mocked(useGetBookingsIdPayment).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

    renderWithProviders(
      <BookingDetails booking={mockBooking} onClose={vi.fn()} />,
    )

    expect(screen.getByText('Платёж')).toBeInTheDocument()
  })
})
