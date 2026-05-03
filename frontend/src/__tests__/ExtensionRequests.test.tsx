import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ExtensionRequests from '@/pages/bookings/ExtensionRequests'

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

const mockExtensionRequests = [
  {
    id: 'ext-1',
    booking_id: 'booking-1',
    user_id: 'user-1',
    bathhouse_id: 'bath-1',
    status: 'pending',
    extra_hours: 2,
    extension_price: 200000,
    new_end_time: '2026-03-20T18:00:00Z',
    created_at: '2026-03-20T15:30:00Z',
    expires_at: new Date(Date.now() + 20 * 60000).toISOString(),
  },
  {
    id: 'ext-2',
    booking_id: 'booking-1',
    user_id: 'user-1',
    bathhouse_id: 'bath-1',
    status: 'approved',
    extra_hours: 1,
    extension_price: 100000,
    new_end_time: '2026-03-19T17:00:00Z',
    created_at: '2026-03-19T15:00:00Z',
    expires_at: '2026-03-19T15:30:00Z',
    resolved_at: '2026-03-19T15:10:00Z',
  },
]

describe('ExtensionRequests', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders empty state when no requests', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { success: true, data: [] },
    })

    renderWithProviders(
      <ExtensionRequests bookingId="booking-1" open={true} onClose={vi.fn()} />,
    )

    await waitFor(() => {
      expect(screen.getByText('Нет запросов на продление')).toBeInTheDocument()
    })
  })

  it('renders extension requests with correct data', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { success: true, data: mockExtensionRequests },
    })

    renderWithProviders(
      <ExtensionRequests bookingId="booking-1" open={true} onClose={vi.fn()} />,
    )

    await waitFor(() => {
      expect(screen.getByText('+2 ч')).toBeInTheDocument()
    })

    expect(screen.getByText('Запросы на продление')).toBeInTheDocument()
    expect(screen.getByText('Ожидает')).toBeInTheDocument()
    expect(screen.getByText('Одобрено')).toBeInTheDocument()
    expect(screen.getByText('+1 ч')).toBeInTheDocument()
    expect(screen.getByText('2000 ₽')).toBeInTheDocument()
    expect(screen.getByText('1000 ₽')).toBeInTheDocument()
  })

  it('shows approve and reject buttons for pending requests', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { success: true, data: [mockExtensionRequests[0]] },
    })

    renderWithProviders(
      <ExtensionRequests bookingId="booking-1" open={true} onClose={vi.fn()} />,
    )

    await waitFor(() => {
      expect(screen.getByText('Одобрить')).toBeInTheDocument()
      expect(screen.getByText('Отклонить')).toBeInTheDocument()
    })
  })

  it('does not show approve/reject buttons for resolved requests', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { success: true, data: [mockExtensionRequests[1]] },
    })

    renderWithProviders(
      <ExtensionRequests bookingId="booking-1" open={true} onClose={vi.fn()} />,
    )

    await waitFor(() => {
      expect(screen.getByText('Одобрено')).toBeInTheDocument()
    })

    expect(screen.queryByText('Одобрить')).not.toBeInTheDocument()
    expect(screen.queryByText('Отклонить')).not.toBeInTheDocument()
  })

  it('calls approve API when approve button clicked', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { success: true, data: [mockExtensionRequests[0]] },
    })
    vi.mocked(axiosInstance.patch).mockResolvedValue({ data: { success: true } })

    renderWithProviders(
      <ExtensionRequests bookingId="booking-1" open={true} onClose={vi.fn()} />,
    )

    await waitFor(() => {
      expect(screen.getByText('Одобрить')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Одобрить'))

    await waitFor(() => {
      expect(axiosInstance.patch).toHaveBeenCalledWith('/bookings/extension-requests/ext-1/approve')
    })
  })

  it('opens reject modal when reject button clicked', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { success: true, data: [mockExtensionRequests[0]] },
    })

    renderWithProviders(
      <ExtensionRequests bookingId="booking-1" open={true} onClose={vi.fn()} />,
    )

    await waitFor(() => {
      expect(screen.getByText('Отклонить')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Отклонить'))

    await waitFor(() => {
      expect(screen.getByText('Причина отклонения')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Укажите причину отклонения (необязательно)')).toBeInTheDocument()
    })
  })

  it('shows pending count warning with 30-minute timeout info', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { success: true, data: [mockExtensionRequests[0]] },
    })

    renderWithProviders(
      <ExtensionRequests bookingId="booking-1" open={true} onClose={vi.fn()} />,
    )

    await waitFor(() => {
      expect(screen.getByText(/1 запрос\(ов\) ожидают вашего решения/)).toBeInTheDocument()
      expect(screen.getByText(/через 30 минут/)).toBeInTheDocument()
    })
  })

  it('shows time remaining tag for pending requests', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { success: true, data: [mockExtensionRequests[0]] },
    })

    renderWithProviders(
      <ExtensionRequests bookingId="booking-1" open={true} onClose={vi.fn()} />,
    )

    await waitFor(() => {
      expect(screen.getByText(/Осталось:/)).toBeInTheDocument()
    })
  })

  it('does not fetch when modal is closed', () => {
    renderWithProviders(
      <ExtensionRequests bookingId="booking-1" open={false} onClose={vi.fn()} />,
    )

    expect(axiosInstance.get).not.toHaveBeenCalled()
  })

  it('shows rejection reason for rejected requests', async () => {
    const rejectedRequest = {
      ...mockExtensionRequests[1],
      status: 'rejected',
      rejection_reason: 'Нет свободных слотов',
    }
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { success: true, data: [rejectedRequest] },
    })

    renderWithProviders(
      <ExtensionRequests bookingId="booking-1" open={true} onClose={vi.fn()} />,
    )

    await waitFor(() => {
      expect(screen.getByText('Причина отклонения: Нет свободных слотов')).toBeInTheDocument()
    })
  })
})
