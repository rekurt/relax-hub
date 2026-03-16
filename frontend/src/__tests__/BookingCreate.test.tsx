import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import BookingCreate from '@/pages/client/BookingCreate'

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetBathhousesId: vi.fn(),
  useGetBathhousesIdAvailableSlots: vi.fn(),
}))

vi.mock('@/api/generated/pricing/pricing', () => ({
  useGetBathhousesIdPriceCalculator: vi.fn(),
}))

vi.mock('@/api/generated/bookings/bookings', () => ({
  usePostBookings: vi.fn(),
}))

vi.mock('@/api/generated/promo-codes/promo-codes', () => ({
  usePostPromoCodesValidate: vi.fn(),
}))

vi.mock('@/api/generated/certificates/certificates', () => ({
  useGetCertificatesCodeBalance: vi.fn(),
}))

import { useGetBathhousesId, useGetBathhousesIdAvailableSlots } from '@/api/generated/bathhouses/bathhouses'
import { useGetBathhousesIdPriceCalculator } from '@/api/generated/pricing/pricing'
import { usePostBookings } from '@/api/generated/bookings/bookings'
import { usePostPromoCodesValidate } from '@/api/generated/promo-codes/promo-codes'
import { useGetCertificatesCodeBalance } from '@/api/generated/certificates/certificates'

function renderWithProviders(
  ui: React.ReactElement,
  { route = '/client/booking/new?bathhouse=bath-1&date=2026-04-01&from=10:00&to=11:00' } = {},
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
              <Route path="/client/booking/new" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockBathhouse = {
  id: 'bath-1',
  name: 'Баня Премиум',
  max_guests: 10,
  price_per_hour: 300000,
}

const mockSlots = [
  { startTime: '2026-03-16T10:00:00Z', endTime: '2026-03-16T11:00:00Z', price: 300000, available: true },
  { startTime: '2026-03-16T11:00:00Z', endTime: '2026-03-16T12:00:00Z', price: 300000, available: true },
  { startTime: '2026-03-16T12:00:00Z', endTime: '2026-03-16T13:00:00Z', price: 350000, available: false },
]

const mockPriceInfo = {
  base_price: 300000,
  final_price: 270000,
  hours: 1,
  price_saving: 30000,
}

describe('BookingCreate', () => {
  beforeEach(() => {
    vi.mocked(useGetBathhousesId).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesId>)

    vi.mocked(useGetBathhousesIdAvailableSlots).mockReturnValue({
      data: { data: mockSlots, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdAvailableSlots>)

    vi.mocked(useGetBathhousesIdPriceCalculator).mockReturnValue({
      data: { data: mockPriceInfo, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdPriceCalculator>)

    vi.mocked(usePostBookings).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostBookings>)

    vi.mocked(usePostPromoCodesValidate).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostPromoCodesValidate>)

    vi.mocked(useGetCertificatesCodeBalance).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetCertificatesCodeBalance>)
  })

  it('renders bathhouse name in title', () => {
    renderWithProviders(<BookingCreate />)
    expect(screen.getByText(/Бронирование: Баня Премиум/)).toBeInTheDocument()
  })

  it('renders available slots', () => {
    renderWithProviders(<BookingCreate />)
    expect(screen.getByText(/10:00 — 11:00/)).toBeInTheDocument()
    expect(screen.getByText(/11:00 — 12:00/)).toBeInTheDocument()
  })

  it('renders price summary from price calculator', () => {
    renderWithProviders(<BookingCreate />)
    expect(screen.getByText('3000 ₽')).toBeInTheDocument() // base price
    expect(screen.getByText('2700 ₽')).toBeInTheDocument() // final price
  })

  it('renders price saving when dynamic pricing applies', () => {
    renderWithProviders(<BookingCreate />)
    expect(screen.getByText(/-300 ₽/)).toBeInTheDocument()
  })

  it('renders promo code input', () => {
    renderWithProviders(<BookingCreate />)
    expect(screen.getByPlaceholderText('Введите промокод')).toBeInTheDocument()
    expect(screen.getByText('Применить')).toBeInTheDocument()
  })

  it('renders certificate code input', () => {
    renderWithProviders(<BookingCreate />)
    expect(screen.getByPlaceholderText('BANI-XXXX-XXXX')).toBeInTheDocument()
  })

  it('renders loyalty and referral toggles', () => {
    renderWithProviders(<BookingCreate />)
    expect(screen.getByText('Использовать баллы лояльности')).toBeInTheDocument()
    expect(screen.getByText('Использовать реферальный бонус')).toBeInTheDocument()
  })

  it('renders refund policy alert', () => {
    renderWithProviders(<BookingCreate />)
    expect(screen.getByText('Политика отмены')).toBeInTheDocument()
    expect(screen.getByText(/100% возврат/)).toBeInTheDocument()
  })

  it('renders booking button', () => {
    renderWithProviders(<BookingCreate />)
    expect(screen.getByText('Забронировать')).toBeInTheDocument()
  })

  it('shows empty state when bathhouse not found', () => {
    vi.mocked(useGetBathhousesId).mockReturnValue({
      data: { data: null, success: false },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesId>)

    renderWithProviders(<BookingCreate />)
    expect(screen.getByText('Баня не найдена')).toBeInTheDocument()
  })

  it('validates promo code when button clicked', () => {
    const mutateFn = vi.fn()
    vi.mocked(usePostPromoCodesValidate).mockReturnValue({
      mutate: mutateFn,
      isPending: false,
    } as unknown as ReturnType<typeof usePostPromoCodesValidate>)

    renderWithProviders(<BookingCreate />)

    const input = screen.getByPlaceholderText('Введите промокод')
    fireEvent.change(input, { target: { value: 'TESTCODE' } })
    fireEvent.click(screen.getByText('Применить'))

    expect(mutateFn).toHaveBeenCalledWith({
      data: {
        code: 'TESTCODE',
        bathhouse_id: 'bath-1',
        amount: 270000,
      },
    })
  })

  it('renders back button', () => {
    renderWithProviders(<BookingCreate />)
    expect(screen.getByText('Назад к бане')).toBeInTheDocument()
  })
})
