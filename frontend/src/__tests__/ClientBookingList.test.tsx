import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ClientBookingList from '@/pages/client/BookingList'

vi.mock('@/api/generated/bookings/bookings', () => ({
  useGetBookings: vi.fn(),
  usePatchBookingsIdCancel: vi.fn(),
}))

import { useGetBookings, usePatchBookingsIdCancel } from '@/api/generated/bookings/bookings'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/client/bookings']}>
            <Routes>
              <Route path="/client/bookings" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockBookings = [
  {
    id: 'book-1',
    bathhouse_id: 'bath-1',
    start_time: '2026-04-01T10:00:00Z',
    end_time: '2026-04-01T12:00:00Z',
    total_price: 600000,
    status: 'confirmed',
    payment_status: 'succeeded',
    guest_count: 4,
  },
  {
    id: 'book-2',
    bathhouse_id: 'bath-2',
    start_time: '2026-03-28T14:00:00Z',
    end_time: '2026-03-28T16:00:00Z',
    total_price: 400000,
    status: 'pending',
    payment_status: '',
    guest_count: 2,
  },
  {
    id: 'book-3',
    bathhouse_id: 'bath-1',
    start_time: '2026-03-15T09:00:00Z',
    end_time: '2026-03-15T11:00:00Z',
    total_price: 500000,
    status: 'completed',
    payment_status: 'succeeded',
    guest_count: 6,
  },
]

describe('ClientBookingList', () => {
  beforeEach(() => {
    vi.mocked(usePatchBookingsIdCancel).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
      variables: undefined,
    } as unknown as ReturnType<typeof usePatchBookingsIdCancel>)
  })

  it('renders bookings table with data', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: {
        data: mockBookings,
        success: true,
        meta: { page: 1, page_size: 10, total_count: 3, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    renderWithProviders(<ClientBookingList />)

    expect(screen.getByText('Мои бронирования')).toBeInTheDocument()
    expect(screen.getByText('6000 ₽')).toBeInTheDocument()
    expect(screen.getByText('4000 ₽')).toBeInTheDocument()
    expect(screen.getByText('5000 ₽')).toBeInTheDocument()
  })

  it('renders booking statuses', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: {
        data: mockBookings,
        success: true,
        meta: { page: 1, page_size: 10, total_count: 3, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    renderWithProviders(<ClientBookingList />)

    expect(screen.getByText('Подтверждено')).toBeInTheDocument()
    expect(screen.getByText('Ожидает')).toBeInTheDocument()
    expect(screen.getByText('Завершено')).toBeInTheDocument()
  })

  it('renders cancel button for pending and confirmed bookings', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: {
        data: mockBookings,
        success: true,
        meta: { page: 1, page_size: 10, total_count: 3, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    renderWithProviders(<ClientBookingList />)

    // confirmed + pending = 2 cancel buttons
    const cancelButtons = screen.getAllByText('Отменить')
    expect(cancelButtons.length).toBe(2)
  })

  it('renders details buttons for all bookings', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: {
        data: mockBookings,
        success: true,
        meta: { page: 1, page_size: 10, total_count: 3, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    renderWithProviders(<ClientBookingList />)

    const detailButtons = screen.getAllByText('Детали')
    expect(detailButtons.length).toBe(3)
  })

  it('renders empty state when no bookings', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: {
        data: [],
        success: true,
        meta: { page: 1, page_size: 10, total_count: 0, total_pages: 0 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    renderWithProviders(<ClientBookingList />)

    expect(screen.getByText('Нет бронирований')).toBeInTheDocument()
  })

  it('renders status filter dropdown', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: { data: [], success: true, meta: {} },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    renderWithProviders(<ClientBookingList />)

    expect(screen.getByText('Все статусы')).toBeInTheDocument()
  })

  it('renders total count in pagination', () => {
    vi.mocked(useGetBookings).mockReturnValue({
      data: {
        data: mockBookings,
        success: true,
        meta: { page: 1, page_size: 10, total_count: 3, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBookings>)

    renderWithProviders(<ClientBookingList />)

    expect(screen.getByText('Всего: 3')).toBeInTheDocument()
  })
})
