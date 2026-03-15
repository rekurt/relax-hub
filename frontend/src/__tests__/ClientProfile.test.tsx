import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ClientProfile from '@/pages/client/ClientProfile'

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn(),
}))

vi.mock('@/api/generated/auth/auth', () => ({
  usePutAuthMe: vi.fn(),
  usePostAuthMeAvatar: vi.fn(),
  useDeleteAuthMeAvatar: vi.fn(),
}))

vi.mock('@/api/generated/notifications/notifications', () => ({
  useGetMyNotificationPreferences: vi.fn(),
  usePutMyNotificationPreferences: vi.fn(),
}))

vi.mock('@/api/generated/oauth/oauth', () => ({
  useGetAuthMeSocialAccounts: vi.fn(),
  useDeleteAuthLinkProvider: vi.fn(),
}))

vi.mock('@/api/generated/cities/cities', () => ({
  useGetCities: vi.fn(),
}))

vi.mock('@/api/generated/users/users', () => ({
  useGetMyStats: vi.fn(),
}))

import { useAuthStore } from '@/stores/auth'
import {
  usePutAuthMe,
  usePostAuthMeAvatar,
  useDeleteAuthMeAvatar,
} from '@/api/generated/auth/auth'
import {
  useGetMyNotificationPreferences,
  usePutMyNotificationPreferences,
} from '@/api/generated/notifications/notifications'
import {
  useGetAuthMeSocialAccounts,
  useDeleteAuthLinkProvider,
} from '@/api/generated/oauth/oauth'
import { useGetCities } from '@/api/generated/cities/cities'
import { useGetMyStats } from '@/api/generated/users/users'

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

const mockUser = {
  id: 'user-1',
  name: 'Мария Иванова',
  email: 'maria@example.com',
  phone: '+7 999 222-33-44',
  bio: 'Люблю бани',
  avatar_url: 'https://example.com/avatar.jpg',
  city_id: 1,
  role: 'client',
}

const mockStats = {
  total_visits: 15,
  review_count: 8,
  total_spent: 4500000,
  avg_rating: 4.7,
  avg_check: 300000,
}

const mockPrefs = {
  in_app: true,
  email: true,
  push: false,
  booking_events: true,
  review_events: true,
  promo_events: false,
  reminders: true,
}

const mockSocialAccounts = [
  {
    id: 'social-1',
    provider: 'google',
    name: 'Maria Google',
    email: 'maria@gmail.com',
    avatar_url: 'https://google.com/avatar.jpg',
    linked_at: '2026-01-15T00:00:00Z',
  },
]

const mockCities = [
  { id: 1, name: 'Москва' },
  { id: 2, name: 'Казань' },
]

describe('ClientProfile', () => {
  beforeEach(() => {
    vi.mocked(useAuthStore).mockReturnValue({
      user: mockUser,
      loadProfile: vi.fn(),
    } as unknown as ReturnType<typeof useAuthStore>)

    vi.mocked(usePutAuthMe).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePutAuthMe>)

    vi.mocked(usePostAuthMeAvatar).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostAuthMeAvatar>)

    vi.mocked(useDeleteAuthMeAvatar).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof useDeleteAuthMeAvatar>)

    vi.mocked(useGetMyStats).mockReturnValue({
      data: { data: mockStats },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyStats>)

    vi.mocked(useGetMyNotificationPreferences).mockReturnValue({
      data: { data: mockPrefs },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyNotificationPreferences>)

    vi.mocked(usePutMyNotificationPreferences).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePutMyNotificationPreferences>)

    vi.mocked(useGetAuthMeSocialAccounts).mockReturnValue({
      data: { data: mockSocialAccounts },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAuthMeSocialAccounts>)

    vi.mocked(useDeleteAuthLinkProvider).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof useDeleteAuthLinkProvider>)

    vi.mocked(useGetCities).mockReturnValue({
      data: { data: mockCities },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetCities>)
  })

  it('renders page title', () => {
    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('Мой профиль')).toBeInTheDocument()
  })

  it('renders all sections', () => {
    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('Статистика')).toBeInTheDocument()
    expect(screen.getByText('Аватар')).toBeInTheDocument()
    expect(screen.getByText('Основная информация')).toBeInTheDocument()
    expect(screen.getByText('Настройки уведомлений')).toBeInTheDocument()
    expect(screen.getByText('Привязанные аккаунты')).toBeInTheDocument()
  })

  it('displays user statistics', () => {
    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('Визиты')).toBeInTheDocument()
    expect(screen.getByText('15')).toBeInTheDocument()
    expect(screen.getAllByText('Отзывы').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('8')).toBeInTheDocument()
    expect(screen.getByText('Потрачено')).toBeInTheDocument()
    expect(screen.getByText('Средний рейтинг')).toBeInTheDocument()
  })

  it('shows stats loading skeleton', () => {
    vi.mocked(useGetMyStats).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyStats>)

    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('Статистика')).toBeInTheDocument()
  })

  it('shows zero stats when no data', () => {
    vi.mocked(useGetMyStats).mockReturnValue({
      data: { data: { total_visits: 0, review_count: 0, total_spent: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyStats>)

    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('0 ₽')).toBeInTheDocument()
    expect(screen.getByText('—')).toBeInTheDocument()
  })

  it('shows avatar with upload and delete buttons', () => {
    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('Загрузить')).toBeInTheDocument()
    expect(screen.getByText('Удалить')).toBeInTheDocument()
  })

  it('hides delete button when no avatar', () => {
    vi.mocked(useAuthStore).mockReturnValue({
      user: { ...mockUser, avatar_url: undefined },
      loadProfile: vi.fn(),
    } as unknown as ReturnType<typeof useAuthStore>)

    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('Загрузить')).toBeInTheDocument()
    expect(screen.queryByText('Удалить')).not.toBeInTheDocument()
  })

  it('populates profile form with user data', () => {
    renderWithProviders(<ClientProfile />)
    expect(screen.getByDisplayValue('Мария Иванова')).toBeInTheDocument()
    expect(screen.getByDisplayValue('+7 999 222-33-44')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Люблю бани')).toBeInTheDocument()
  })

  it('submits profile form', async () => {
    const mutateFn = vi.fn()
    vi.mocked(usePutAuthMe).mockReturnValue({
      mutate: mutateFn,
      isPending: false,
    } as unknown as ReturnType<typeof usePutAuthMe>)

    renderWithProviders(<ClientProfile />)

    const nameInput = screen.getByDisplayValue('Мария Иванова')
    fireEvent.change(nameInput, { target: { value: 'Анна Сидорова' } })

    const saveButtons = screen.getAllByText('Сохранить')
    fireEvent.click(saveButtons[0]!)

    await waitFor(() => {
      expect(mutateFn).toHaveBeenCalledWith({
        data: expect.objectContaining({ name: 'Анна Сидорова' }),
      })
    })
  })

  it('renders notification preferences switches', () => {
    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('В приложении')).toBeInTheDocument()
    expect(screen.getByText('Push-уведомления')).toBeInTheDocument()
    expect(screen.getByText('Бронирования')).toBeInTheDocument()
    expect(screen.getByText('Напоминания')).toBeInTheDocument()
  })

  it('renders linked social account (Google)', () => {
    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('Google')).toBeInTheDocument()
    expect(screen.getByText('maria@gmail.com')).toBeInTheDocument()
    expect(screen.getByText('Отвязать')).toBeInTheDocument()
  })

  it('shows link buttons for unlinked providers', () => {
    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('Привязать ВКонтакте')).toBeInTheDocument()
    expect(screen.getByText('Привязать Яндекс')).toBeInTheDocument()
    expect(screen.queryByText('Привязать Google')).not.toBeInTheDocument()
  })

  it('shows empty state for social accounts when none linked', () => {
    vi.mocked(useGetAuthMeSocialAccounts).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAuthMeSocialAccounts>)

    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('Нет привязанных аккаунтов')).toBeInTheDocument()
    expect(screen.getByText('Привязать ВКонтакте')).toBeInTheDocument()
    expect(screen.getByText('Привязать Яндекс')).toBeInTheDocument()
    expect(screen.getByText('Привязать Google')).toBeInTheDocument()
  })
})
