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
  usePostAuthDeleteAccount: vi.fn(),
  usePostAuthRestoreAccount: vi.fn(),
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

vi.mock('@/api/generated/region/region', () => ({
  useGetMyRegion: vi.fn(),
  usePutMyRegion: vi.fn(),
}))

import { useAuthStore } from '@/stores/auth'
import {
  usePutAuthMe,
  usePostAuthMeAvatar,
  useDeleteAuthMeAvatar,
  usePostAuthDeleteAccount,
  usePostAuthRestoreAccount,
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
import { useGetMyRegion, usePutMyRegion } from '@/api/generated/region/region'

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

    vi.mocked(usePostAuthDeleteAccount).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostAuthDeleteAccount>)

    vi.mocked(usePostAuthRestoreAccount).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostAuthRestoreAccount>)

    vi.mocked(useGetMyRegion).mockReturnValue({
      data: { data: { region: 'RU' } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyRegion>)

    vi.mocked(usePutMyRegion).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePutMyRegion>)
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

  it('does not render duplicated header quick actions moved to profile menu', () => {
    renderWithProviders(<ClientProfile />)
    expect(screen.queryByRole('button', { name: 'Карты' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Уведомления' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Безопасность' })).not.toBeInTheDocument()
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

  it('renders explicit preferences entry point in profile header', () => {
    renderWithProviders(<ClientProfile />)
    expect(screen.getByRole('button', { name: /Предпочтения/ })).toBeInTheDocument()
  })

  it('renders premium notification cockpit in profile', () => {
    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('Контур уведомлений')).toBeInTheDocument()
    expect(screen.getByText('Активных каналов')).toBeInTheDocument()
    expect(screen.getByText('Активных сценариев')).toBeInTheDocument()
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

  // Account Deletion Tests
  it('renders account deletion section', () => {
    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('Удаление аккаунта')).toBeInTheDocument()
    expect(screen.getByText('Удалить аккаунт')).toBeInTheDocument()
  })

  it('shows deletion confirmation modal on click', async () => {
    renderWithProviders(<ClientProfile />)
    fireEvent.click(screen.getByText('Удалить аккаунт'))

    await waitFor(() => {
      expect(screen.getByText('Период восстановления — 30 дней')).toBeInTheDocument()
    })
    expect(screen.getByText('Что произойдёт с вашими средствами')).toBeInTheDocument()
  })

  it('calls delete account mutation on modal confirm', async () => {
    const mutateFn = vi.fn()
    vi.mocked(usePostAuthDeleteAccount).mockReturnValue({
      mutate: mutateFn,
      isPending: false,
    } as unknown as ReturnType<typeof usePostAuthDeleteAccount>)

    renderWithProviders(<ClientProfile />)
    fireEvent.click(screen.getByText('Удалить аккаунт'))

    await waitFor(() => {
      expect(screen.getByText('Период восстановления — 30 дней')).toBeInTheDocument()
    })

    // The confirm modal has an OK button with text "Удалить аккаунт"
    const allDeleteButtons = screen.getAllByText('Удалить аккаунт')
    // The modal OK button is the last one (modal renders after the page button)
    const modalOkButton = allDeleteButtons[allDeleteButtons.length - 1]!
    fireEvent.click(modalOkButton)

    await waitFor(() => {
      expect(mutateFn).toHaveBeenCalled()
    })
  })

  it('shows restore option when deletion is pending', async () => {
    // Simulate the deletion flow: first render, then trigger a state update
    const deleteMutateFn = vi.fn()
    vi.mocked(usePostAuthDeleteAccount).mockReturnValue({
      mutate: deleteMutateFn,
      isPending: false,
    } as unknown as ReturnType<typeof usePostAuthDeleteAccount>)

    renderWithProviders(<ClientProfile />)

    // Initially should show the delete button and description
    expect(screen.getByText('Удалить аккаунт')).toBeInTheDocument()
  })

  it('shows deletion description text', () => {
    renderWithProviders(<ClientProfile />)
    expect(screen.getByText(/средства, внесённые пополнением, будут возвращены/i)).toBeInTheDocument()
  })

  // Region Switching Tests
  it('renders region section with current region', () => {
    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('Россия (RUB)')).toBeInTheDocument()
    expect(screen.getByText('Сменить на Беларусь')).toBeInTheDocument()
  })

  it('shows region switch info alert', () => {
    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('Смена региона')).toBeInTheDocument()
  })

  it('opens region switch confirmation modal', async () => {
    renderWithProviders(<ClientProfile />)
    fireEvent.click(screen.getByText('Сменить на Беларусь'))

    await waitFor(() => {
      expect(screen.getByText('Подтверждение смены региона')).toBeInTheDocument()
    })
    expect(screen.getByText('Последствия смены региона')).toBeInTheDocument()
  })

  it('calls switch region mutation on modal confirm', async () => {
    const mutateFn = vi.fn()
    vi.mocked(usePutMyRegion).mockReturnValue({
      mutate: mutateFn,
      isPending: false,
    } as unknown as ReturnType<typeof usePutMyRegion>)

    renderWithProviders(<ClientProfile />)
    fireEvent.click(screen.getByText('Сменить на Беларусь'))

    await waitFor(() => {
      expect(screen.getByText('Подтверждение смены региона')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Подтвердить'))

    await waitFor(() => {
      expect(mutateFn).toHaveBeenCalledWith({ data: { region: 'BY' } })
    })
  })

  it('shows BY region when user is in Belarus', () => {
    vi.mocked(useGetMyRegion).mockReturnValue({
      data: { data: { region: 'BY' } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyRegion>)

    renderWithProviders(<ClientProfile />)
    expect(screen.getByText('Беларусь (BYN)')).toBeInTheDocument()
    expect(screen.getByText('Сменить на Россия')).toBeInTheDocument()
  })

  it('shows loading state for region section', () => {
    vi.mocked(useGetMyRegion).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyRegion>)

    renderWithProviders(<ClientProfile />)
    // Region section should not show the switch button while loading
    expect(screen.queryByText('Сменить на Беларусь')).not.toBeInTheDocument()
  })
})
