import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ClientBookingDetail from '@/pages/client/BookingDetail'

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    get: vi.fn(),
    put: vi.fn(),
    post: vi.fn(),
  },
}))

vi.mock('@/api/generated/bookings/bookings', () => ({
  usePatchBookingsIdCancel: vi.fn(),
}))

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetBathhousesId: vi.fn(),
}))

vi.mock('@/api/generated/payments/payments', () => ({
  useGetBookingsIdPayment: vi.fn(),
  usePostBookingsIdPay: vi.fn(),
}))

vi.mock('@/api/generated/share/share', () => ({
  usePostApiV1BookingsShare: vi.fn(),
}))

vi.mock('@/components/ApplePayButton', () => ({
  default: () => <div data-testid="apple-pay-btn">Apple Pay</div>,
}))

vi.mock('@/components/GooglePayButton', () => ({
  default: () => <div data-testid="google-pay-btn">Google Pay</div>,
}))

vi.mock('@/components/ShareButton', () => ({
  default: ({ url }: { url: string }) => <button data-testid="share-btn">{url}</button>,
}))

vi.mock('@/lib/useDeviceToken', () => ({
  useDeviceToken: () => ({ requestPushPermission: vi.fn() }),
}))

import { axiosInstance } from '@/api/axios-instance'
import { usePatchBookingsIdCancel } from '@/api/generated/bookings/bookings'
import { useGetBathhousesId } from '@/api/generated/bathhouses/bathhouses'
import { useGetBookingsIdPayment, usePostBookingsIdPay } from '@/api/generated/payments/payments'
import { usePostApiV1BookingsShare } from '@/api/generated/share/share'

function renderWithProviders(ui: React.ReactElement, bookingId = 'booking-123') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={[`/client/bookings/${bookingId}`]}>
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
  id: 'booking-123-full-uuid',
  bathhouse_id: 'bath-1',
  status: 'confirmed',
  start_time: '2026-05-01T10:00:00Z',
  end_time: '2026-05-01T13:00:00Z',
  guest_count: 4,
  comment: 'Тестовый комментарий',
  original_price: 600000,
  total_price: 540000,
  promo_discount: 60000,
  loyalty_discount: 0,
  certificate_discount: 0,
  points_spent: 0,
  referral_bonus_used: 0,
  earned_points: 50,
  modification_count: 1,
  created_at: '2026-04-15T08:00:00Z',
}

const mockPayment = {
  id: 'pay-123-full-uuid',
  status: 'succeeded',
  amount: 540000,
  provider: 'yookassa',
  refund_amount: 0,
  created_at: '2026-04-15T08:05:00Z',
}

function setupMocks(overrides?: {
  booking?: typeof mockBooking | null
  payment?: typeof mockPayment | null
  paymentLoading?: boolean
}) {
  const booking = overrides?.booking !== undefined ? overrides.booking : mockBooking

  vi.mocked(axiosInstance.get).mockImplementation((url: string) => {
    if (url.includes('/bookings/')) {
      return Promise.resolve({ data: { data: booking, success: true } })
    }
    return Promise.reject(new Error('not found'))
  })

  vi.mocked(useGetBathhousesId).mockReturnValue({
    data: { data: { cancellation_policy: 'flexible' } },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetBathhousesId>)

  vi.mocked(useGetBookingsIdPayment).mockReturnValue({
    data: overrides?.payment !== undefined
      ? { data: overrides.payment, success: true }
      : { data: mockPayment, success: true },
    isLoading: overrides?.paymentLoading ?? false,
  } as unknown as ReturnType<typeof useGetBookingsIdPayment>)

  vi.mocked(usePatchBookingsIdCancel).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof usePatchBookingsIdCancel>)

  vi.mocked(usePostBookingsIdPay).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof usePostBookingsIdPay>)

  vi.mocked(usePostApiV1BookingsShare).mockReturnValue({
    mutateAsync: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof usePostApiV1BookingsShare>)
}

describe('ClientBookingDetail', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('renders booking details when data is loaded', async () => {
    setupMocks()
    renderWithProviders(<ClientBookingDetail />)

    expect(await screen.findByText('Детали бронирования')).toBeInTheDocument()
    expect(screen.getByText('4')).toBeInTheDocument()
    expect(screen.getByText('Тестовый комментарий')).toBeInTheDocument()
    // total_price appears in both booking and payment sections
    expect(screen.getAllByText('5400 ₽').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('6000 ₽')).toBeInTheDocument()
  })

  it('shows empty state when booking not found', async () => {
    setupMocks({ booking: null })
    renderWithProviders(<ClientBookingDetail />)

    expect(await screen.findByText('Бронирование не найдено')).toBeInTheDocument()
  })

  it('shows promo discount row when promo was applied', async () => {
    setupMocks()
    renderWithProviders(<ClientBookingDetail />)

    expect(await screen.findByText('Скидка по промокоду')).toBeInTheDocument()
    expect(screen.getByText('-600 ₽')).toBeInTheDocument()
  })

  it('shows cancel button for confirmed booking', async () => {
    setupMocks()
    renderWithProviders(<ClientBookingDetail />)

    expect(await screen.findByText('Отменить бронирование')).toBeInTheDocument()
  })

  it('shows modify button for confirmed booking with modifications left', async () => {
    setupMocks()
    renderWithProviders(<ClientBookingDetail />)

    expect(await screen.findByText('Изменить бронирование')).toBeInTheDocument()
    expect(screen.getByText('1 / 3')).toBeInTheDocument()
  })

  it('hides cancel and modify for completed booking, shows review and dispute', async () => {
    setupMocks({
      booking: { ...mockBooking, status: 'completed' },
    })
    renderWithProviders(<ClientBookingDetail />)

    expect(await screen.findByText('Детали бронирования')).toBeInTheDocument()
    expect(screen.queryByText('Отменить бронирование')).not.toBeInTheDocument()
    expect(screen.queryByText('Изменить бронирование')).not.toBeInTheDocument()
    expect(screen.getByText('Оставить отзыв')).toBeInTheDocument()
    expect(screen.getByText('Открыть спор')).toBeInTheDocument()
  })

  it('shows payment info when payment exists', async () => {
    setupMocks()
    renderWithProviders(<ClientBookingDetail />)

    expect(await screen.findByText('Платёж')).toBeInTheDocument()
    expect(screen.getByText('yookassa')).toBeInTheDocument()
  })

  it('shows "Платёж не найден" when no payment', async () => {
    setupMocks({ payment: null })
    renderWithProviders(<ClientBookingDetail />)

    expect(await screen.findByText('Платёж не найден')).toBeInTheDocument()
  })

  it('shows cancellation policy alert for cancellable booking', async () => {
    setupMocks()
    renderWithProviders(<ClientBookingDetail />)

    expect(await screen.findByText('Политика отмены')).toBeInTheDocument()
  })
})
