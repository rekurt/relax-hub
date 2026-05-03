import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ModificationRequests from '@/pages/bookings/ModificationRequests'

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    get: vi.fn(),
    patch: vi.fn(),
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

const mockRequests = [
  {
    id: 'req-1',
    booking_id: 'booking-1',
    user_id: 'user-1',
    bathhouse_id: 'bath-1',
    status: 'pending',
    old_start_time: '2026-05-01T10:00:00Z',
    old_end_time: '2026-05-01T13:00:00Z',
    old_guest_count: 3,
    old_total_price: 450000,
    proposed_start_time: '2026-05-01T11:00:00Z',
    proposed_end_time: '2026-05-01T15:00:00Z',
    proposed_guest_count: 5,
    proposed_total_price: 600000,
    created_at: '2026-04-18T08:00:00Z',
    expires_at: '2026-04-19T08:00:00Z',
  },
  {
    id: 'req-2',
    booking_id: 'booking-1',
    user_id: 'user-1',
    bathhouse_id: 'bath-1',
    status: 'rejected',
    old_start_time: '2026-05-01T10:00:00Z',
    old_end_time: '2026-05-01T13:00:00Z',
    old_guest_count: 3,
    old_total_price: 450000,
    proposed_start_time: '2026-05-02T10:00:00Z',
    proposed_end_time: '2026-05-02T12:00:00Z',
    proposed_guest_count: 2,
    proposed_total_price: 300000,
    rejection_reason: 'Время занято',
    created_at: '2026-04-17T08:00:00Z',
    expires_at: '2026-04-18T08:00:00Z',
    resolved_at: '2026-04-17T12:00:00Z',
  },
]

describe('ModificationRequests', () => {
  const onClose = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders modal title when open', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { data: mockRequests, success: true },
    })

    renderWithProviders(
      <ModificationRequests bookingId="booking-1" open={true} onClose={onClose} />,
    )

    expect(screen.getByText('Запросы на изменение')).toBeInTheDocument()
  })

  it('shows empty state when no requests', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { data: [], success: true },
    })

    renderWithProviders(
      <ModificationRequests bookingId="booking-1" open={true} onClose={onClose} />,
    )

    await waitFor(() => {
      expect(screen.getByText('Нет запросов на изменение')).toBeInTheDocument()
    })
  })

  it('displays pending request with approve/reject buttons', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { data: mockRequests, success: true },
    })

    renderWithProviders(
      <ModificationRequests bookingId="booking-1" open={true} onClose={onClose} />,
    )

    await waitFor(() => {
      expect(screen.getByText('Ожидает')).toBeInTheDocument()
    })

    expect(screen.getByText('Одобрить')).toBeInTheDocument()
    expect(screen.getByText('Отклонить')).toBeInTheDocument()
  })

  it('shows rejected request with rejection reason', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { data: mockRequests, success: true },
    })

    renderWithProviders(
      <ModificationRequests bookingId="booking-1" open={true} onClose={onClose} />,
    )

    await waitFor(() => {
      expect(screen.getByText('Отклонено')).toBeInTheDocument()
    })

    expect(screen.getByText('Причина отклонения: Время занято')).toBeInTheDocument()
  })

  it('shows pending count warning', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { data: mockRequests, success: true },
    })

    renderWithProviders(
      <ModificationRequests bookingId="booking-1" open={true} onClose={onClose} />,
    )

    await waitFor(() => {
      expect(
        screen.getByText(/1 запрос\(ов\) ожидают вашего решения/),
      ).toBeInTheDocument()
    })
  })

  it('opens reject reason modal when reject clicked', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { data: [mockRequests[0]], success: true },
    })

    renderWithProviders(
      <ModificationRequests bookingId="booking-1" open={true} onClose={onClose} />,
    )

    await waitFor(() => {
      expect(screen.getByText('Одобрить')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Отклонить'))

    await waitFor(() => {
      expect(screen.getByText('Причина отклонения')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Укажите причину отклонения (необязательно)')).toBeInTheDocument()
    })
  })

  it('displays price comparison between old and proposed', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { data: [mockRequests[0]], success: true },
    })

    renderWithProviders(
      <ModificationRequests bookingId="booking-1" open={true} onClose={onClose} />,
    )

    await waitFor(() => {
      expect(screen.getByText('4500 ₽')).toBeInTheDocument()
      expect(screen.getByText('6000 ₽')).toBeInTheDocument()
    })

    expect(screen.getByText('Текущие гости')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
    expect(screen.getByText('Предложенные гости')).toBeInTheDocument()
    expect(screen.getByText('5')).toBeInTheDocument()
  })
})
