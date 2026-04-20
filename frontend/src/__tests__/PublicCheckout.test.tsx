import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import PublicCheckout from '@/pages/public/PublicCheckout'

const mockPublicPost = vi.fn()
const mockAxiosPost = vi.fn()
const mockNavigate = vi.fn()

vi.mock('axios', () => ({
  default: {
    create: vi.fn(() => ({
      post: (...args: unknown[]) => mockPublicPost(...args),
    })),
    isAxiosError: (error: unknown) => Boolean((error as { isAxiosError?: boolean })?.isAxiosError),
  },
}))

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    post: (...args: unknown[]) => mockAxiosPost(...args),
  },
}))

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetBathhousesId: vi.fn(),
  useGetBathhousesIdAvailableSlots: vi.fn(),
}))

vi.mock('@/api/generated/pricing/pricing', () => ({
  useGetBathhousesIdPriceCalculator: vi.fn(),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn(),
}))

import { useGetBathhousesId, useGetBathhousesIdAvailableSlots } from '@/api/generated/bathhouses/bathhouses'
import { useGetBathhousesIdPriceCalculator } from '@/api/generated/pricing/pricing'
import type { AuthState } from '@/stores/auth'
import { useAuthStore } from '@/stores/auth'

function renderWithProviders(route = '/checkout?bathhouse=bath-1&date=2026-04-20') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={[route]}>
            <Routes>
              <Route path="/checkout" element={<PublicCheckout />} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

function mockAuthStore(user: Record<string, unknown> | null = null) {
  const state = {
    user,
    token: null,
    isLoading: false,
    isAuthenticated: false,
    setAuth: vi.fn(),
    logout: vi.fn(),
    loadProfile: vi.fn(),
  } as unknown as AuthState
  vi.mocked(useAuthStore).mockImplementation((selector) => {
    return selector(state)
  })
  return state
}

describe('PublicCheckout', () => {
  beforeEach(() => {
    vi.clearAllMocks()

    mockAuthStore(null)

    vi.mocked(useGetBathhousesId).mockReturnValue({
      data: {
        data: {
          id: 'bath-1',
          name: 'Баня Премиум',
          address: 'ул. Мира, 10',
          max_guests: 10,
          min_duration: 2,
          booking_mode: 'instant',
          cancellation_policy: 'flexible',
          security_deposit_percent: 30,
        },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesId>)

    vi.mocked(useGetBathhousesIdAvailableSlots).mockReturnValue({
      data: {
        data: [
          {
            startTime: '2026-04-20T10:00:00',
            endTime: '2026-04-20T11:00:00',
            available: true,
          },
          {
            startTime: '2026-04-20T11:00:00',
            endTime: '2026-04-20T12:00:00',
            available: true,
          },
          {
            startTime: '2026-04-20T12:00:00',
            endTime: '2026-04-20T13:00:00',
            available: true,
          },
        ],
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdAvailableSlots>)

    vi.mocked(useGetBathhousesIdPriceCalculator).mockReturnValue({
      data: {
        data: {
          final_price: 800000,
        },
      },
    } as unknown as ReturnType<typeof useGetBathhousesIdPriceCalculator>)
  })

  it('shows inline OTP recovery message when SMS request fails', async () => {
    mockPublicPost.mockRejectedValueOnce(new Error('sms failed'))

    renderWithProviders()

    fireEvent.click(screen.getByRole('button', { name: /10:00/ }))
    fireEvent.click(screen.getByRole('button', { name: /11:00/ }))
    fireEvent.change(screen.getByPlaceholderText('Имя'), { target: { value: 'Иван' } })
    fireEvent.change(screen.getByPlaceholderText('Телефон'), { target: { value: '+79990000000' } })
    fireEvent.click(screen.getByRole('checkbox'))
    fireEvent.click(screen.getByRole('button', { name: 'Получить SMS-код' }))

    expect((await screen.findAllByText('Не удалось отправить код')).length).toBeGreaterThan(0)
    expect(screen.getByDisplayValue('Иван')).toBeInTheDocument()
    expect(screen.getByDisplayValue('+79990000000')).toBeInTheDocument()
  })

  it('shows trust summary and booking rules next to checkout', () => {
    renderWithProviders('/checkout?bathhouse=bath-1&date=2026-04-20&from=2026-04-20T10:00:00&to=2026-04-20T12:00:00')

    expect(screen.getByText('Что важно до подтверждения')).toBeInTheDocument()
    expect(screen.getByText('Минимум 2 ч')).toBeInTheDocument()
    expect(screen.getByText('Гибкая отмена')).toBeInTheDocument()
    expect(screen.getAllByText('Мгновенное подтверждение').length).toBeGreaterThan(0)
  })

  it('shows inline OTP verification error without resetting the form', async () => {
    mockPublicPost
      .mockResolvedValueOnce({ data: { success: true } })
      .mockRejectedValueOnce({
        isAxiosError: true,
        response: {
          status: 400,
          data: { error: { message: 'Неверный код' } },
        },
      })

    renderWithProviders()

    fireEvent.click(screen.getByRole('button', { name: /10:00/ }))
    fireEvent.click(screen.getByRole('button', { name: /11:00/ }))
    fireEvent.change(screen.getByPlaceholderText('Имя'), { target: { value: 'Иван' } })
    fireEvent.change(screen.getByPlaceholderText('Телефон'), { target: { value: '+79990000000' } })
    fireEvent.click(screen.getByRole('checkbox'))
    fireEvent.click(screen.getByRole('button', { name: 'Получить SMS-код' }))

    expect(await screen.findByPlaceholderText('Код из SMS')).toBeInTheDocument()
    fireEvent.change(screen.getByPlaceholderText('Код из SMS'), { target: { value: '123456' } })
    fireEvent.click(screen.getByRole('button', { name: 'Подтвердить и создать бронь' }))

    expect(await screen.findByText('Неверный код')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Иван')).toBeInTheDocument()
    expect(screen.getByDisplayValue('+79990000000')).toBeInTheDocument()
  })

  it('shows slot conflict inline for logged-in client booking failure', async () => {
    mockAuthStore({
      id: 'client-1',
      role: 'client',
      name: 'Иван',
      phone: '+79990000000',
    })

    mockAxiosPost.mockRejectedValueOnce({
      isAxiosError: true,
      response: {
        status: 409,
        data: { error: { message: 'Слот уже занят' } },
      },
    })

    renderWithProviders('/checkout?bathhouse=bath-1&date=2026-04-20')

    fireEvent.click(screen.getByRole('button', { name: /10:00/ }))
    fireEvent.click(screen.getByRole('button', { name: /11:00/ }))
    const submitButton = screen.getByRole('button', { name: 'Создать бронь' })
    await waitFor(() => {
      expect(submitButton).toBeEnabled()
    })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(mockAxiosPost).toHaveBeenCalled()
    })
    expect(await screen.findByText('Выбранный слот уже недоступен')).toBeInTheDocument()
    expect(screen.getAllByText(/10:00 - 12:00/).length).toBeGreaterThan(0)
  })

  it('completes guest booking and redirects to client booking page', async () => {
    const authState = mockAuthStore(null)
    mockPublicPost
      .mockResolvedValueOnce({ data: { success: true } })
      .mockResolvedValueOnce({
        data: {
          data: {
            token: 'jwt-token',
            user: {
              id: 'client-2',
              role: 'client',
              name: 'Иван',
              phone: '+79990000000',
            },
          },
        },
      })
    mockAxiosPost.mockResolvedValueOnce({
      data: {
        data: {
          id: 'booking-77',
        },
      },
    })

    renderWithProviders()

    fireEvent.click(screen.getByRole('button', { name: /10:00/ }))
    fireEvent.click(screen.getByRole('button', { name: /11:00/ }))
    fireEvent.change(screen.getByPlaceholderText('Имя'), { target: { value: 'Иван' } })
    fireEvent.change(screen.getByPlaceholderText('Телефон'), { target: { value: '+79990000000' } })
    fireEvent.click(screen.getByRole('checkbox'))
    fireEvent.click(screen.getByRole('button', { name: 'Получить SMS-код' }))

    expect(await screen.findByPlaceholderText('Код из SMS')).toBeInTheDocument()
    fireEvent.change(screen.getByPlaceholderText('Код из SMS'), { target: { value: '123456' } })
    fireEvent.click(screen.getByRole('button', { name: 'Подтвердить и создать бронь' }))

    await waitFor(() => {
      expect(authState.setAuth).toHaveBeenCalled()
      expect(mockNavigate).toHaveBeenCalledWith('/client/bookings/booking-77', { replace: true })
    })
  })

  it('does not render separate duration controls', () => {
    renderWithProviders()

    expect(screen.queryByText(/Длительность/)).not.toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /10:00/ }))
    expect(screen.getByText(/Выберите конечный слот/)).toBeInTheDocument()
  })
})
