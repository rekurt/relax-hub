import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import NotificationList from '@/pages/notifications/NotificationList'

vi.mock('@/api/generated/notifications/notifications', () => ({
  useGetMyNotifications: vi.fn(),
  usePatchMyNotificationsIdRead: vi.fn(),
  usePatchMyNotificationsReadAll: vi.fn(),
}))

import {
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
    body: 'Клиент забронировал баню',
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
  {
    id: 'notif-3',
    type: 'chat_message',
    title: 'Новое сообщение',
    body: 'У вас новое сообщение в чате',
    is_read: false,
    created_at: '2026-03-13T12:00:00Z',
  },
]

describe('NotificationList', () => {
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

  it('renders page title and read-all button', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: mockNotifications, meta: { total_count: 3, page: 0, page_size: 20, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<NotificationList />)

    expect(screen.getByText('Уведомления')).toBeInTheDocument()
    expect(screen.getByText('Прочитать все')).toBeInTheDocument()
  })

  it('renders notification items', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: mockNotifications, meta: { total_count: 3, page: 0, page_size: 20, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<NotificationList />)

    expect(screen.getByText('Новое бронирование')).toBeInTheDocument()
    expect(screen.getByText('Новый отзыв')).toBeInTheDocument()
    expect(screen.getByText('Новое сообщение')).toBeInTheDocument()
    expect(screen.getByText('Клиент забронировал баню')).toBeInTheDocument()
  })

  it('shows empty state when no notifications', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: [], meta: { total_count: 0, page: 0, page_size: 20, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<NotificationList />)

    expect(screen.getByText('Нет уведомлений')).toBeInTheDocument()
  })

  it('calls markAllRead when button clicked', () => {
    const markAllMutate = vi.fn()
    vi.mocked(usePatchMyNotificationsReadAll).mockReturnValue({
      mutate: markAllMutate,
      isPending: false,
    } as unknown as ReturnType<typeof usePatchMyNotificationsReadAll>)
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: mockNotifications, meta: { total_count: 3, page: 0, page_size: 20, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<NotificationList />)
    fireEvent.click(screen.getByText('Прочитать все'))

    expect(markAllMutate).toHaveBeenCalled()
  })

  it('shows "read" action for unread notifications', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: {
        data: [mockNotifications[0]],
        meta: { total_count: 1, page: 0, page_size: 20, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<NotificationList />)

    expect(screen.getByText('Прочитать')).toBeInTheDocument()
  })

  it('marks individual notification as read', () => {
    const markOneMutate = vi.fn()
    vi.mocked(usePatchMyNotificationsIdRead).mockReturnValue({
      mutate: markOneMutate,
      isPending: false,
    } as unknown as ReturnType<typeof usePatchMyNotificationsIdRead>)
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: {
        data: [mockNotifications[0]],
        meta: { total_count: 1, page: 0, page_size: 20, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<NotificationList />)
    fireEvent.click(screen.getByText('Прочитать'))

    expect(markOneMutate).toHaveBeenCalledWith({ id: 'notif-1' })
  })

  it('shows total count in pagination', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: mockNotifications, meta: { total_count: 3, page: 0, page_size: 20, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<NotificationList />)

    expect(screen.getByText('Всего: 3')).toBeInTheDocument()
  })
})
