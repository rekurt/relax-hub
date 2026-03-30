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

vi.mock('@/api/generated/add-ons/add-ons', () => ({
  useGetBathhousesIdAddons: vi.fn(),
}))

vi.mock('@/api/generated/wallet/wallet', () => ({
  useGetMyWallet: vi.fn(),
}))

import { useGetBathhousesId, useGetBathhousesIdAvailableSlots } from '@/api/generated/bathhouses/bathhouses'
import { useGetBathhousesIdPriceCalculator } from '@/api/generated/pricing/pricing'
import { usePostBookings } from '@/api/generated/bookings/bookings'
import { usePostPromoCodesValidate } from '@/api/generated/promo-codes/promo-codes'
import { useGetCertificatesCodeBalance } from '@/api/generated/certificates/certificates'
import { useGetBathhousesIdAddons } from '@/api/generated/add-ons/add-ons'
import { useGetMyWallet } from '@/api/generated/wallet/wallet'

function renderWithProviders(
  ui: React.ReactElement,
  { route = '/client/booking/new?bathhouse=bath-1&date=2026-04-01&from=2026-04-01T10:00:00Z&to=2026-04-01T11:00:00Z' } = {},
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
  booking_mode: 'instant',
}

const mockRequestModeBathhouse = {
  ...mockBathhouse,
  booking_mode: 'request',
}

const mockSlots = [
  { startTime: '2026-04-01T10:00:00Z', endTime: '2026-04-01T11:00:00Z', price: 300000, available: true },
  { startTime: '2026-04-01T11:00:00Z', endTime: '2026-04-01T12:00:00Z', price: 300000, available: true },
  { startTime: '2026-04-01T12:00:00Z', endTime: '2026-04-01T13:00:00Z', price: 350000, available: false },
]

const mockPriceInfo = {
  base_price: 300000,
  final_price: 270000,
  hours: 1,
  price_saving: 30000,
}

const mockAddons = [
  { id: 'addon-1', name: 'Веники', description: 'Берёзовые веники', price: 50000, unit: 'per_item', is_active: true, sort_order: 1 },
  { id: 'addon-2', name: 'Полотенце', description: '', price: 20000, unit: 'per_person', is_active: true, sort_order: 2 },
]

function setupDefaultMocks() {
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

  vi.mocked(useGetBathhousesIdAddons).mockReturnValue({
    data: { data: mockAddons, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetBathhousesIdAddons>)

  vi.mocked(useGetMyWallet).mockReturnValue({
    data: { data: { balance: 500000 }, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyWallet>)
}

describe('BookingCreate', () => {
  beforeEach(() => {
    setupDefaultMocks()
  })

  describe('Step 1: Date and Time', () => {
    it('renders bathhouse name in title', () => {
      renderWithProviders(<BookingCreate />)
      expect(screen.getByText(/Бронирование: Баня Премиум/)).toBeInTheDocument()
    })

    it('renders stepper with 4 steps', () => {
      renderWithProviders(<BookingCreate />)
      // "Дата и время" appears in both stepper and card title
      expect(screen.getAllByText('Дата и время').length).toBeGreaterThanOrEqual(1)
      expect(screen.getByText('Доп. услуги')).toBeInTheDocument()
      expect(screen.getByText('Скидки')).toBeInTheDocument()
      expect(screen.getByText('Оплата')).toBeInTheDocument()
    })

    it('renders available slots on step 1', () => {
      renderWithProviders(<BookingCreate />)
      expect(screen.getByText(/10:00 — 11:00/)).toBeInTheDocument()
      expect(screen.getByText(/11:00 — 12:00/)).toBeInTheDocument()
    })

    it('renders "Далее" button on step 1', () => {
      renderWithProviders(<BookingCreate />)
      expect(screen.getByText('Далее')).toBeInTheDocument()
    })

    it('shows empty state when bathhouse not found', () => {
      vi.mocked(useGetBathhousesId).mockReturnValue({
        data: { data: null, success: false },
        isLoading: false,
      } as unknown as ReturnType<typeof useGetBathhousesId>)

      renderWithProviders(<BookingCreate />)
      expect(screen.getByText('Баня не найдена')).toBeInTheDocument()
    })

    it('shows back to bathhouse button', () => {
      renderWithProviders(<BookingCreate />)
      expect(screen.getByText('Назад к бане')).toBeInTheDocument()
    })
  })

  describe('Step navigation', () => {
    it('navigates to step 2 when "Далее" is clicked with slot selected', () => {
      renderWithProviders(<BookingCreate />)
      // Slot is pre-selected via URL params
      fireEvent.click(screen.getByText('Далее'))
      expect(screen.getByText('Дополнительные услуги')).toBeInTheDocument()
      // Should show add-ons
      expect(screen.getByText('Веники')).toBeInTheDocument()
    })

    it('navigates through all 4 steps', () => {
      renderWithProviders(<BookingCreate />)

      // Step 1 -> 2
      fireEvent.click(screen.getByText('Далее'))
      expect(screen.getByText('Веники')).toBeInTheDocument()

      // Step 2 -> 3
      fireEvent.click(screen.getByText('Далее'))
      expect(screen.getByPlaceholderText('Введите промокод')).toBeInTheDocument()

      // Step 3 -> 4
      fireEvent.click(screen.getByText('Далее'))
      expect(screen.getByText('Способ оплаты')).toBeInTheDocument()
      expect(screen.getByText('Забронировать')).toBeInTheDocument()
    })

    it('shows "Назад" button on step 2+', () => {
      renderWithProviders(<BookingCreate />)
      // No back on step 1
      expect(screen.queryByText('Назад')).not.toBeInTheDocument()

      fireEvent.click(screen.getByText('Далее'))
      expect(screen.getByText('Назад')).toBeInTheDocument()
    })

    it('navigates back to previous step', () => {
      renderWithProviders(<BookingCreate />)
      fireEvent.click(screen.getByText('Далее'))
      expect(screen.getByText('Веники')).toBeInTheDocument()

      fireEvent.click(screen.getByText('Назад'))
      expect(screen.getByText(/10:00 — 11:00/)).toBeInTheDocument()
    })
  })

  describe('Step 2: Add-ons', () => {
    it('renders add-on items', () => {
      renderWithProviders(<BookingCreate />)
      fireEvent.click(screen.getByText('Далее'))

      expect(screen.getByText('Веники')).toBeInTheDocument()
      expect(screen.getByText('Полотенце')).toBeInTheDocument()
    })

    it('shows empty state when no add-ons', () => {
      vi.mocked(useGetBathhousesIdAddons).mockReturnValue({
        data: { data: [], success: true },
        isLoading: false,
      } as unknown as ReturnType<typeof useGetBathhousesIdAddons>)

      renderWithProviders(<BookingCreate />)
      fireEvent.click(screen.getByText('Далее'))

      expect(screen.getByText('Нет доступных дополнительных услуг')).toBeInTheDocument()
    })
  })

  describe('Step 3: Discounts', () => {
    function goToStep3() {
      renderWithProviders(<BookingCreate />)
      fireEvent.click(screen.getByText('Далее')) // to step 2
      fireEvent.click(screen.getByText('Далее')) // to step 3
    }

    it('renders promo code input', () => {
      goToStep3()
      expect(screen.getByPlaceholderText('Введите промокод')).toBeInTheDocument()
      expect(screen.getByText('Применить')).toBeInTheDocument()
    })

    it('renders certificate code input', () => {
      goToStep3()
      expect(screen.getByPlaceholderText('BANI-XXXX-XXXX')).toBeInTheDocument()
    })

    it('renders loyalty and referral toggles', () => {
      goToStep3()
      expect(screen.getByText('Использовать баллы лояльности')).toBeInTheDocument()
      expect(screen.getByText('Использовать реферальный бонус')).toBeInTheDocument()
    })

    it('validates promo code when button clicked', () => {
      const mutateFn = vi.fn()
      vi.mocked(usePostPromoCodesValidate).mockReturnValue({
        mutate: mutateFn,
        isPending: false,
      } as unknown as ReturnType<typeof usePostPromoCodesValidate>)

      goToStep3()

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
  })

  describe('Step 4: Payment', () => {
    function goToStep4() {
      renderWithProviders(<BookingCreate />)
      fireEvent.click(screen.getByText('Далее'))
      fireEvent.click(screen.getByText('Далее'))
      fireEvent.click(screen.getByText('Далее'))
    }

    it('renders payment method options', () => {
      goToStep4()
      expect(screen.getByText('Способ оплаты')).toBeInTheDocument()
      expect(screen.getByText('Банковская карта')).toBeInTheDocument()
      expect(screen.getByText('СБП')).toBeInTheDocument()
    })

    it('renders price summary on step 4', () => {
      goToStep4()
      expect(screen.getByText('3000 ₽')).toBeInTheDocument() // base price
      expect(screen.getByText('2700 ₽')).toBeInTheDocument() // total price
    })

    it('renders price saving', () => {
      goToStep4()
      expect(screen.getByText(/-300 ₽/)).toBeInTheDocument()
    })

    it('renders booking button on step 4', () => {
      goToStep4()
      expect(screen.getByText('Забронировать')).toBeInTheDocument()
    })

    it('renders refund policy on step 4', () => {
      goToStep4()
      expect(screen.getByText('Политика отмены')).toBeInTheDocument()
    })

    it('shows combo payment option when wallet balance < total price', () => {
      vi.mocked(useGetMyWallet).mockReturnValue({
        data: { data: { balance: 100000 }, success: true },
        isLoading: false,
      } as unknown as ReturnType<typeof useGetMyWallet>)

      goToStep4()
      expect(screen.getByText(/Кошелёк \+ Карта/)).toBeInTheDocument()
    })

    it('shows wallet-only option when balance covers total', () => {
      vi.mocked(useGetMyWallet).mockReturnValue({
        data: { data: { balance: 500000 }, success: true },
        isLoading: false,
      } as unknown as ReturnType<typeof useGetMyWallet>)

      goToStep4()
      expect(screen.getByText(/Кошелёк \(5000 ₽\)/)).toBeInTheDocument()
    })
  })

  describe('Request mode', () => {
    it('shows request mode alert for request-mode bathhouse', () => {
      vi.mocked(useGetBathhousesId).mockReturnValue({
        data: { data: mockRequestModeBathhouse, success: true },
        isLoading: false,
      } as unknown as ReturnType<typeof useGetBathhousesId>)

      renderWithProviders(<BookingCreate />)
      expect(screen.getByText('Бронирование по заявке')).toBeInTheDocument()
    })

    it('shows "Отправить заявку" button on step 4 for request mode', () => {
      vi.mocked(useGetBathhousesId).mockReturnValue({
        data: { data: mockRequestModeBathhouse, success: true },
        isLoading: false,
      } as unknown as ReturnType<typeof useGetBathhousesId>)

      renderWithProviders(<BookingCreate />)
      fireEvent.click(screen.getByText('Далее'))
      fireEvent.click(screen.getByText('Далее'))
      fireEvent.click(screen.getByText('Далее'))

      expect(screen.getByText('Отправить заявку')).toBeInTheDocument()
    })
  })

  describe('Slot conflict', () => {
    it('shows conflict message when booking fails with slot_unavailable', () => {
      let onErrorCallback: ((error: unknown) => void) | undefined
      vi.mocked(usePostBookings).mockImplementation((options) => {
        onErrorCallback = options?.mutation?.onError as (error: unknown) => void
        return {
          mutate: vi.fn().mockImplementation(() => {
            onErrorCallback?.({ error: { code: 'slot_unavailable', message: 'Slot taken' } })
          }),
          isPending: false,
        } as unknown as ReturnType<typeof usePostBookings>
      })

      renderWithProviders(<BookingCreate />)
      // Go to step 4
      fireEvent.click(screen.getByText('Далее'))
      fireEvent.click(screen.getByText('Далее'))
      fireEvent.click(screen.getByText('Далее'))

      fireEvent.click(screen.getByText('Забронировать'))

      expect(screen.getByText('Слот только что занят')).toBeInTheDocument()
      expect(screen.getByText('Выбрать другое время')).toBeInTheDocument()
    })
  })
})
