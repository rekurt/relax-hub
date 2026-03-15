import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ProfileSettings from '@/pages/settings/ProfileSettings'

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
  name: 'Иван Петров',
  email: 'ivan@example.com',
  phone: '+7 999 111-22-33',
  bio: 'Владелец бань',
  avatar_url: 'https://example.com/avatar.jpg',
  city_id: 1,
  role: 'owner',
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
    provider: 'vk',
    name: 'Иван ВК',
    email: 'ivan@vk.com',
    avatar_url: 'https://vk.com/avatar.jpg',
    linked_at: '2026-01-01T00:00:00Z',
  },
]

const mockCities = [
  { id: 1, name: 'Москва' },
  { id: 2, name: 'Санкт-Петербург' },
]

describe('ProfileSettings', () => {
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
    renderWithProviders(<ProfileSettings />)
    expect(screen.getByText('Настройки профиля')).toBeInTheDocument()
  })

  it('renders all settings sections', () => {
    renderWithProviders(<ProfileSettings />)
    expect(screen.getByText('Аватар')).toBeInTheDocument()
    expect(screen.getByText('Основная информация')).toBeInTheDocument()
    expect(screen.getByText('Настройки уведомлений')).toBeInTheDocument()
    expect(screen.getByText('Привязанные аккаунты')).toBeInTheDocument()
  })

  it('shows avatar with upload and delete buttons', () => {
    renderWithProviders(<ProfileSettings />)
    expect(screen.getByText('Загрузить')).toBeInTheDocument()
    expect(screen.getByText('Удалить')).toBeInTheDocument()
  })

  it('does not show delete avatar when no avatar', () => {
    vi.mocked(useAuthStore).mockReturnValue({
      user: { ...mockUser, avatar_url: undefined },
      loadProfile: vi.fn(),
    } as unknown as ReturnType<typeof useAuthStore>)

    renderWithProviders(<ProfileSettings />)
    expect(screen.getByText('Загрузить')).toBeInTheDocument()
    expect(screen.queryByText('Удалить')).not.toBeInTheDocument()
  })

  it('populates profile form with user data', () => {
    renderWithProviders(<ProfileSettings />)
    expect(screen.getByDisplayValue('Иван Петров')).toBeInTheDocument()
    expect(screen.getByDisplayValue('+7 999 111-22-33')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Владелец бань')).toBeInTheDocument()
  })

  it('submits profile form', async () => {
    const mutateFn = vi.fn()
    vi.mocked(usePutAuthMe).mockReturnValue({
      mutate: mutateFn,
      isPending: false,
    } as unknown as ReturnType<typeof usePutAuthMe>)

    renderWithProviders(<ProfileSettings />)

    const nameInput = screen.getByDisplayValue('Иван Петров')
    fireEvent.change(nameInput, { target: { value: 'Пётр Иванов' } })

    const saveButtons = screen.getAllByText('Сохранить')
    fireEvent.click(saveButtons[0]!)

    await waitFor(() => {
      expect(mutateFn).toHaveBeenCalledWith({
        data: expect.objectContaining({ name: 'Пётр Иванов' }),
      })
    })
  })

  it('renders notification preferences switches', () => {
    renderWithProviders(<ProfileSettings />)
    expect(screen.getByText('В приложении')).toBeInTheDocument()
    expect(screen.getByText('Email')).toBeInTheDocument()
    expect(screen.getByText('Push-уведомления')).toBeInTheDocument()
    expect(screen.getByText('Бронирования')).toBeInTheDocument()
    expect(screen.getByText('Отзывы')).toBeInTheDocument()
    expect(screen.getByText('Промокоды')).toBeInTheDocument()
    expect(screen.getByText('Напоминания')).toBeInTheDocument()
  })

  it('renders linked social accounts', () => {
    renderWithProviders(<ProfileSettings />)
    expect(screen.getByText('ВКонтакте')).toBeInTheDocument()
    expect(screen.getByText('ivan@vk.com')).toBeInTheDocument()
    expect(screen.getByText('Отвязать')).toBeInTheDocument()
  })

  it('shows link buttons for unlinked providers', () => {
    renderWithProviders(<ProfileSettings />)
    expect(screen.getByText('Привязать Яндекс')).toBeInTheDocument()
    expect(screen.getByText('Привязать Google')).toBeInTheDocument()
    expect(screen.queryByText('Привязать ВКонтакте')).not.toBeInTheDocument()
  })

  it('shows skeleton when preferences are loading', () => {
    vi.mocked(useGetMyNotificationPreferences).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyNotificationPreferences>)

    renderWithProviders(<ProfileSettings />)
    expect(screen.getByText('Настройки уведомлений')).toBeInTheDocument()
  })

  it('shows empty state for social accounts when none linked', () => {
    vi.mocked(useGetAuthMeSocialAccounts).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAuthMeSocialAccounts>)

    renderWithProviders(<ProfileSettings />)
    expect(screen.getByText('Нет привязанных аккаунтов')).toBeInTheDocument()
    expect(screen.getByText('Привязать ВКонтакте')).toBeInTheDocument()
    expect(screen.getByText('Привязать Яндекс')).toBeInTheDocument()
    expect(screen.getByText('Привязать Google')).toBeInTheDocument()
  })
})
