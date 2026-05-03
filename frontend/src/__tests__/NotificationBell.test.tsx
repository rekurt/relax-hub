import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import NotificationBell from '@/components/NotificationBell'

vi.mock('@/api/generated/notifications/notifications', () => ({
  useGetMyNotificationsUnreadCount: vi.fn(),
  useGetMyNotifications: vi.fn(),
  usePatchMyNotificationsIdRead: vi.fn(),
  usePatchMyNotificationsReadAll: vi.fn(),
}))

import {
  useGetMyNotificationsUnreadCount,
  useGetMyNotifications,
  usePatchMyNotificationsIdRead,
  usePatchMyNotificationsReadAll,
} from '@/api/generated/notifications/notifications'

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

const mockNotifications = [
  {
    id: 'notif-1',
    type: 'booking_new',
    title: 'Новое бронирование',
    body: 'Клиент забронировал баню на 15 марта',
    is_read: false,
    created_at: '2026-03-15T10:00:00Z',
  },
  {
    id: 'notif-2',
    type: 'review_new',
    title: 'Новый отзыв',
    body: 'Оставлен отзыв с оценкой 5',
    is_read: true,
    created_at: '2026-03-14T08:00:00Z',
  },
]

describe('NotificationBell', () => {
  beforeEach(() => {
    vi.mocked(usePatchMyNotificationsIdRead).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePatchMyNotificationsIdRead>)
    vi.mocked(usePatchMyNotificationsReadAll).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePatchMyNotificationsReadAll>)
  })

  it('renders bell icon', () => {
    vi.mocked(useGetMyNotificationsUnreadCount).mockReturnValue({
      data: { data: { unread_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotificationsUnreadCount>)
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<NotificationBell />)

    expect(screen.getByRole('button')).toBeInTheDocument()
  })

  it('displays unread count badge', () => {
    vi.mocked(useGetMyNotificationsUnreadCount).mockReturnValue({
      data: { data: { unread_count: 3 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotificationsUnreadCount>)
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<NotificationBell />)

    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('shows popover with notifications on click', async () => {
    vi.mocked(useGetMyNotificationsUnreadCount).mockReturnValue({
      data: { data: { unread_count: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotificationsUnreadCount>)
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: mockNotifications },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<NotificationBell />)

    fireEvent.click(screen.getByRole('button'))

    await waitFor(() => {
      expect(screen.getByText('Уведомления')).toBeInTheDocument()
    })
    expect(screen.getByText('Новое бронирование')).toBeInTheDocument()
    expect(screen.getByText('Клиент забронировал баню на 15 марта')).toBeInTheDocument()
  })

  it('shows "read all" button when there are unread notifications', async () => {
    vi.mocked(useGetMyNotificationsUnreadCount).mockReturnValue({
      data: { data: { unread_count: 2 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotificationsUnreadCount>)
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: mockNotifications },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<NotificationBell />)
    fireEvent.click(screen.getByRole('button'))

    await waitFor(() => {
      expect(screen.getByText('Прочитать все')).toBeInTheDocument()
    })
  })

  it('calls markAllRead mutation when "read all" clicked', async () => {
    const markAllMutate = vi.fn()
    vi.mocked(usePatchMyNotificationsReadAll).mockReturnValue({
      mutate: markAllMutate,
      isPending: false,
    } as unknown as ReturnType<typeof usePatchMyNotificationsReadAll>)
    vi.mocked(useGetMyNotificationsUnreadCount).mockReturnValue({
      data: { data: { unread_count: 2 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotificationsUnreadCount>)
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: mockNotifications },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<NotificationBell />)
    fireEvent.click(screen.getByRole('button'))

    await waitFor(() => {
      expect(screen.getByText('Прочитать все')).toBeInTheDocument()
    })
    fireEvent.click(screen.getByText('Прочитать все'))

    expect(markAllMutate).toHaveBeenCalled()
  })

  it('shows "all notifications" link', async () => {
    vi.mocked(useGetMyNotificationsUnreadCount).mockReturnValue({
      data: { data: { unread_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotificationsUnreadCount>)
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<NotificationBell />)
    fireEvent.click(screen.getByRole('button'))

    await waitFor(() => {
      expect(screen.getByText('Все уведомления')).toBeInTheDocument()
    })
  })

  it('does not show badge when count is zero', () => {
    vi.mocked(useGetMyNotificationsUnreadCount).mockReturnValue({
      data: { data: { unread_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotificationsUnreadCount>)
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<NotificationBell />)

    expect(screen.queryByText('0')).not.toBeInTheDocument()
  })
})
