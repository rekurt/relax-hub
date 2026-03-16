import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import AdminNotifications from '@/pages/admin/AdminNotifications'

vi.mock('@/api/generated/notifications/notifications', () => ({
  useGetMyNotifications: vi.fn(),
  usePatchMyNotificationsIdRead: vi.fn(),
  usePatchMyNotificationsReadAll: vi.fn(),
  useGetMyNotificationPreferences: vi.fn(),
  usePutMyNotificationPreferences: vi.fn(),
}))

import {
  useGetMyNotifications,
  usePatchMyNotificationsIdRead,
  usePatchMyNotificationsReadAll,
  useGetMyNotificationPreferences,
  usePutMyNotificationPreferences,
} from '@/api/generated/notifications/notifications'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/notifications']}>
            <Routes>
              <Route path="/admin/notifications" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockNotifications = [
  {
    id: 'notif-1',
    type: 'review_new',
    title: 'Новый отзыв на модерации',
    body: 'Пользователь оставил отзыв для проверки',
    is_read: false,
    created_at: '2026-03-14T18:00:00Z',
  },
  {
    id: 'notif-2',
    type: 'booking_new',
    title: 'Новое бронирование',
    body: 'Создано новое бронирование',
    is_read: true,
    created_at: '2026-03-13T10:00:00Z',
  },
]

const mockMarkOneRead = { mutate: vi.fn(), isPending: false }
const mockMarkAllRead = { mutate: vi.fn(), isPending: false }
const mockUpdatePrefs = { mutate: vi.fn(), isPending: false }

describe('AdminNotifications', () => {
  beforeEach(() => {
    vi.mocked(usePatchMyNotificationsIdRead).mockReturnValue(
      mockMarkOneRead as unknown as ReturnType<typeof usePatchMyNotificationsIdRead>,
    )
    vi.mocked(usePatchMyNotificationsReadAll).mockReturnValue(
      mockMarkAllRead as unknown as ReturnType<typeof usePatchMyNotificationsReadAll>,
    )
    vi.mocked(usePutMyNotificationPreferences).mockReturnValue(
      mockUpdatePrefs as unknown as ReturnType<typeof usePutMyNotificationPreferences>,
    )
    vi.mocked(useGetMyNotificationPreferences).mockReturnValue({
      data: {
        data: {
          in_app: true,
          email: false,
          push: false,
          booking_events: true,
          review_events: true,
          promo_events: true,
          reminders: true,
        },
        success: true,
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotificationPreferences>)
  })

  it('renders page title', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<AdminNotifications />)
    expect(screen.getByText('Уведомления')).toBeInTheDocument()
  })

  it('shows empty state when no notifications', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<AdminNotifications />)
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

    renderWithProviders(<AdminNotifications />)
    expect(screen.getByText('Новый отзыв на модерации')).toBeInTheDocument()
    expect(screen.getByText('Новое бронирование')).toBeInTheDocument()
    expect(screen.getByText('Пользователь оставил отзыв для проверки')).toBeInTheDocument()
  })

  it('renders mark all read button', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: {
        data: mockNotifications,
        success: true,
        meta: { total_count: 2 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<AdminNotifications />)
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

    renderWithProviders(<AdminNotifications />)
    fireEvent.click(screen.getByText('Прочитать все'))

    await waitFor(() => {
      expect(mutateFn).toHaveBeenCalled()
    })
  })

  it('renders notification preferences section', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<AdminNotifications />)
    expect(screen.getByText('Настройки уведомлений')).toBeInTheDocument()
    expect(screen.getByText('Каналы доставки')).toBeInTheDocument()
    expect(screen.getByText('События')).toBeInTheDocument()
    expect(screen.getByText('Сохранить настройки')).toBeInTheDocument()
  })

  it('renders preference channel switches', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<AdminNotifications />)
    expect(screen.getByText('В приложении')).toBeInTheDocument()
    expect(screen.getByText('Email')).toBeInTheDocument()
    expect(screen.getByText('Push-уведомления')).toBeInTheDocument()
  })

  it('renders preference event switches', () => {
    vi.mocked(useGetMyNotifications).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotifications>)

    renderWithProviders(<AdminNotifications />)
    expect(screen.getByText('Бронирования')).toBeInTheDocument()
    expect(screen.getByText('Отзывы')).toBeInTheDocument()
    expect(screen.getByText('Промокоды')).toBeInTheDocument()
    expect(screen.getByText('Напоминания')).toBeInTheDocument()
  })
})
