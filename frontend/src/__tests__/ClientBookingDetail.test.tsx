import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ClientBookingDetail from '@/pages/client/BookingDetail'

vi.mock('@/api/generated/bookings/bookings', () => ({
  useGetBookings: vi.fn(),
  usePatchBookingsIdCancel: vi.fn(),
}))

vi.mock('@/api/generated/payments/payments', () => ({
  useGetBookingsIdPayment: vi.fn(),
  usePostBookingsIdPay: vi.fn(),
}))

import { useGetBookings, usePatchBookingsIdCancel } from '@/api/generated/bookings/bookings'
import { useGetBookingsIdPayment, usePostBookingsIdPay } from '@/api/generated/payments/payments'

function renderWithProviders(
  ui: React.ReactElement,
  { route = '/client/bookings/book-1' } = {},
) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={[route]}>
            <Routes>
              <Route path="/client/bookings/:id" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockBooking = {
  id: 'book-1',
  bathhouse_id: 'bath-1',
  user_id: 'user-1',
  start_time: '2026-04-01T10:00:00Z',
  end_time: '2026-04-01T12:00:00Z',
  guest_count: 4,
  comment: 'Берём веники',
  original_price: 600000,
  promo_discount: 50000,
  loyalty_discount: 0,
  certificate_discount: 30000,
  points_spent: 0,
  referral_bonus_used: 10000,
  total_price: 510000,
  earned_points: 51,
  status: 'confirmed',
  created_at: '2026-03-20T08:00:00Z',
}

const mockPayment = {
  id: 'pay-1',
  amount: 510000,
  status: 'succeeded',
  provider: 'yookassa',
  refund_amount: 0,
  created_at: '2026-03-20T08:05:00Z',
}

describe('ClientBookingDetail', () => {
  beforeEach(() => {
    vi.mocked(usePatchBookingsIdCancel).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePatchBookingsIdCancel>)

    vi.mocked(usePostBookingsIdPay).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostBookingsIdPay>)
  })

  it('renders booking details with prices', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: { data: [mockBooking], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    vi.mocked(useGetBookingsIdPayment).mockReturnValue({
      data: { data: mockPayment, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

    renderWithProviders(<ClientBookingDetail />)

    expect(screen.getByText('Детали бронирования')).toBeInTheDocument()
    expect(screen.getByText('6000 ₽')).toBeInTheDocument() // original price
    // 5100 ₽ appears in both booking total and payment amount
    const totalPriceElements = screen.getAllByText('5100 ₽')
    expect(totalPriceElements.length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('-500 ₽')).toBeInTheDocument() // promo discount
    expect(screen.getByText('-300 ₽')).toBeInTheDocument() // certificate discount
    expect(screen.getByText('-100 ₽')).toBeInTheDocument() // referral bonus
  })

  it('renders booking status tag', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: { data: [mockBooking], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    vi.mocked(useGetBookingsIdPayment).mockReturnValue({
      data: { data: mockPayment, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

    renderWithProviders(<ClientBookingDetail />)

    expect(screen.getByText('Подтверждено')).toBeInTheDocument()
  })

  it('renders payment information', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: { data: [mockBooking], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    vi.mocked(useGetBookingsIdPayment).mockReturnValue({
      data: { data: mockPayment, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

    renderWithProviders(<ClientBookingDetail />)

    expect(screen.getByText('Оплачено')).toBeInTheDocument()
    expect(screen.getByText('yookassa')).toBeInTheDocument()
  })

  it('renders cancel button for confirmed booking', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: { data: [mockBooking], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    vi.mocked(useGetBookingsIdPayment).mockReturnValue({
      data: { data: mockPayment, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

    renderWithProviders(<ClientBookingDetail />)

    expect(screen.getByText('Отменить бронирование')).toBeInTheDocument()
  })

  it('renders refund policy alert for cancellable booking', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: { data: [mockBooking], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    vi.mocked(useGetBookingsIdPayment).mockReturnValue({
      data: { data: mockPayment, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

    renderWithProviders(<ClientBookingDetail />)

    expect(screen.getByText('Политика отмены')).toBeInTheDocument()
  })

  it('does not render cancel button for completed booking', () => {
    const completedBooking = { ...mockBooking, status: 'completed' }
    vi.mocked(useGetBookings).mockReturnValue({
      data: { data: [completedBooking], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    vi.mocked(useGetBookingsIdPayment).mockReturnValue({
      data: { data: mockPayment, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

    renderWithProviders(<ClientBookingDetail />)

    expect(screen.queryByText('Отменить бронирование')).not.toBeInTheDocument()
  })

  it('shows not found state when booking does not exist', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    vi.mocked(useGetBookingsIdPayment).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

    renderWithProviders(<ClientBookingDetail />)

    expect(screen.getByText('Бронирование не найдено')).toBeInTheDocument()
  })

  it('renders guest count and comment', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: { data: [mockBooking], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    vi.mocked(useGetBookingsIdPayment).mockReturnValue({
      data: { data: mockPayment, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

    renderWithProviders(<ClientBookingDetail />)

    expect(screen.getByText('4')).toBeInTheDocument()
    expect(screen.getByText('Берём веники')).toBeInTheDocument()
  })

  it('renders earned points', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: { data: [mockBooking], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    vi.mocked(useGetBookingsIdPayment).mockReturnValue({
      data: { data: mockPayment, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

    renderWithProviders(<ClientBookingDetail />)

    expect(screen.getByText('+51')).toBeInTheDocument()
  })

  it('renders back button', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: { data: [mockBooking], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    vi.mocked(useGetBookingsIdPayment).mockReturnValue({
      data: { data: mockPayment, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

    renderWithProviders(<ClientBookingDetail />)

    expect(screen.getByText('К бронированиям')).toBeInTheDocument()
  })
})
