import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ClientNotifications from '@/pages/client/ClientNotifications'

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
    type: 'booking_confirmed',
    title: 'Бронирование подтверждено',
    body: 'Ваше бронирование на 15.03 подтверждено',
    is_read: false,
    created_at: '2026-03-14T18:00:00Z',
  },
  {
    id: 'notif-2',
    type: 'review_response',
    title: 'Ответ на отзыв',
    body: 'Владелец ответил на ваш отзыв',
    is_read: true,
    created_at: '2026-03-13T10:00:00Z',
  },
]

const mockMarkOneRead = { mutate: vi.fn(), isPending: false }
const mockMarkAllRead = { mutate: vi.fn(), isPending: false }

describe('ClientNotifications', () => {
  beforeEach(() => {
    vi.mocked(usePatchMyNotificationsIdRead).mockReturnValue(
      mockMarkOneRead as unknown as ReturnType<typeof usePatchMyNotificationsIdRead>,
    )
    vi.mocked(usePatchMyNotificationsReadAll).mockReturnValue(
      mockMarkAllRead as unknown as ReturnType<typeof usePatchMyNotificationsReadAll>,
    )
  })

  it('renders notifications page title', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<ClientNotifications />)
    expect(screen.getByText('Уведомления')).toBeInTheDocument()
  })

  it('shows empty state when no notifications', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<ClientNotifications />)
    expect(screen.getByText('Нет уведомлений')).toBeInTheDocument()
  })

  it('renders notification items', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: {
        data: mockNotifications,
        success: true,
        meta: { total_count: 2, page: 1, page_size: 20, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<ClientNotifications />)
    expect(screen.getByText('Бронирование подтверждено')).toBeInTheDocument()
    expect(screen.getByText('Ответ на отзыв')).toBeInTheDocument()
    expect(screen.getByText('Ваше бронирование на 15.03 подтверждено')).toBeInTheDocument()
  })

  it('shows mark all read button', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: {
        data: mockNotifications,
        success: true,
        meta: { total_count: 2 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<ClientNotifications />)
    expect(screen.getByText('Прочитать все')).toBeInTheDocument()
  })

  it('calls markAllRead when button is clicked', async () => {
    const mutateFn = vi.fn()
    vi.mocked(usePatchMyNotificationsReadAll).mockReturnValue({
      mutate: mutateFn,
      isPending: false,
    } as unknown as ReturnType<typeof usePatchMyNotificationsReadAll>)

    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: {
        data: mockNotifications,
        success: true,
        meta: { total_count: 2 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<ClientNotifications />)

    fireEvent.click(screen.getByText('Прочитать все'))

    await waitFor(() => {
      expect(mutateFn).toHaveBeenCalled()
    })
  })

  it('marks single notification as read when clicked', async () => {
    const mutateFn = vi.fn()
    vi.mocked(usePatchMyNotificationsIdRead).mockReturnValue({
      mutate: mutateFn,
      isPending: false,
    } as unknown as ReturnType<typeof usePatchMyNotificationsIdRead>)

    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: {
        data: mockNotifications,
        success: true,
        meta: { total_count: 2 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<ClientNotifications />)

    // Click on the unread notification's text
    fireEvent.click(screen.getByText('Ваше бронирование на 15.03 подтверждено'))

    await waitFor(() => {
      expect(mutateFn).toHaveBeenCalledWith({ id: 'notif-1' })
    })
  })

  it('highlights unread notifications', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: {
        data: mockNotifications,
        success: true,
        meta: { total_count: 2 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<ClientNotifications />)

    // The unread notification should have a "Прочитать" action button
    const readButtons = screen.getAllByText('Прочитать')
    // "Прочитать все" + "Прочитать" for unread item
    expect(readButtons.length).toBeGreaterThanOrEqual(1)
  })
})
