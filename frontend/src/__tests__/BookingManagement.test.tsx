import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import BookingManagement from '@/pages/admin/BookingManagement'

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

import { axiosInstance } from '@/api/axios-instance'

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
    id: 'b1-full-uuid-aaaa',
    user_id: 'u1-full-uuid-bbbb',
    bathhouse_id: 'bh1-full-uuid-cc',
    start_time: '2026-04-20T10:00:00Z',
    end_time: '2026-04-20T13:00:00Z',
    guest_count: 3,
    total_price: 450000,
    status: 'confirmed',
    comment: '',
    created_at: '2026-04-18T08:00:00Z',
  },
  {
    id: 'b2-full-uuid-dddd',
    user_id: 'u2-full-uuid-eeee',
    bathhouse_id: 'bh2-full-uuid-ff',
    start_time: '2026-04-21T14:00:00Z',
    end_time: '2026-04-21T17:00:00Z',
    guest_count: 5,
    total_price: 600000,
    status: 'cancelled',
    comment: '',
    created_at: '2026-04-17T12:00:00Z',
  },
]

describe('BookingManagement', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders page title and filters', () => {
    renderWithProviders(<BookingManagement />)

    expect(screen.getByText('Управление бронированиями')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('User ID')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Bathhouse ID')).toBeInTheDocument()
    expect(screen.getByText('Поиск')).toBeInTheDocument()
  })

  it('shows empty state before search', () => {
    renderWithProviders(<BookingManagement />)

    expect(
      screen.getByText('Нет бронирований. Используйте фильтры и нажмите «Поиск» для загрузки данных'),
    ).toBeInTheDocument()
  })

  it('loads and displays bookings on search click', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: {
        data: mockBookings,
        meta: { page: 1, page_size: 20, total_count: 2, total_pages: 1 },
      },
    })

    renderWithProviders(<BookingManagement />)
    fireEvent.click(screen.getByText('Поиск'))

    await waitFor(() => {
      expect(screen.getByText('4500 ₽')).toBeInTheDocument()
      expect(screen.getByText('6000 ₽')).toBeInTheDocument()
    })

    expect(screen.getByText('Подтверждено')).toBeInTheDocument()
    expect(screen.getByText('Отменено')).toBeInTheDocument()
  })

  it('disables cancel button for already cancelled bookings', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: {
        data: mockBookings,
        meta: { page: 1, page_size: 20, total_count: 2, total_pages: 1 },
      },
    })

    renderWithProviders(<BookingManagement />)
    fireEvent.click(screen.getByText('Поиск'))

    await waitFor(() => {
      expect(screen.getByText('Подтверждено')).toBeInTheDocument()
    })

    const cancelButtons = screen.getAllByRole('button', { name: /Отменить/i })
      .filter(btn => btn.closest('td'))
    expect(cancelButtons.length).toBe(2)
    const disabledButton = cancelButtons.find(btn => btn.hasAttribute('disabled'))
    expect(disabledButton).toBeTruthy()
  })

  it('opens cancel modal when cancel button clicked', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: {
        data: [mockBookings[0]],
        meta: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
      },
    })

    renderWithProviders(<BookingManagement />)
    fireEvent.click(screen.getByText('Поиск'))

    await waitFor(() => {
      expect(screen.getByText('Подтверждено')).toBeInTheDocument()
    })

    const cancelBtns = screen.getAllByRole('button', { name: /Отменить/i })
    const tableCancelBtn = cancelBtns.find(btn => btn.closest('td'))
    fireEvent.click(tableCancelBtn!)

    await waitFor(() => {
      expect(screen.getByText('Отмена бронирования')).toBeInTheDocument()
      expect(screen.getByText('Причина')).toBeInTheDocument()
    })
  })

  it('opens change status modal when status button clicked', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: {
        data: [mockBookings[0]],
        meta: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
      },
    })

    renderWithProviders(<BookingManagement />)
    fireEvent.click(screen.getByText('Поиск'))

    await waitFor(() => {
      expect(screen.getByText('Подтверждено')).toBeInTheDocument()
    })

    // Click the "Статус" button in the actions column (not the column header)
    const statusBtns = screen.getAllByRole('button', { name: /Статус/i })
    const tableStatusBtn = statusBtns.find(btn => btn.closest('td'))
    fireEvent.click(tableStatusBtn!)

    await waitFor(() => {
      expect(screen.getByText('Изменение статуса')).toBeInTheDocument()
      expect(screen.getByText('Новый статус')).toBeInTheDocument()
    })
  })
})
