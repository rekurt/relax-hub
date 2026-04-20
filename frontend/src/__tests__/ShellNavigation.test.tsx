import { fireEvent, render, screen, within } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, beforeEach, vi } from 'vitest'
import PublicLayout from '@/components/PublicLayout'
import ClientLayout from '@/components/ClientLayout'
import { useAuthStore } from '@/stores/auth'

vi.mock('@/api/generated/cities/cities', () => ({
  useGetCities: vi.fn(),
}))

vi.mock('@/components/NotificationBell', () => ({
  default: () => <div data-testid="notification-bell" />,
}))

vi.mock('@/components/OnboardingTour', () => ({
  default: () => null,
}))

import { useGetCities } from '@/api/generated/cities/cities'

function mockDesktopViewport() {
  window.matchMedia = vi.fn().mockImplementation((query: string) => ({
    matches: query.includes('min-width'),
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  }))
}

function mockMobileViewport() {
  window.matchMedia = vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  }))
}

function renderPublicShell(route = '/catalog') {
  return render(
    <ConfigProvider locale={ruRU}>
      <MemoryRouter initialEntries={[route]}>
        <Routes>
          <Route path="/" element={<PublicLayout />}>
            <Route index element={<div>Главная</div>} />
            <Route path="catalog" element={<div>Каталог</div>} />
            <Route path="faq" element={<div>Помощь</div>} />
            <Route path="contacts" element={<div>Контакты</div>} />
            <Route path="certificates" element={<div>Сертификаты</div>} />
            <Route path="terms" element={<div>Terms of Use</div>} />
          </Route>
        </Routes>
      </MemoryRouter>
    </ConfigProvider>,
  )
}

function renderClientShell(route = '/client/bookings') {
  return render(
    <ConfigProvider locale={ruRU}>
      <MemoryRouter initialEntries={[route]}>
        <Routes>
          <Route path="/client" element={<ClientLayout />}>
            <Route index element={<div>Клиент</div>} />
            <Route path="bookings" element={<div>Мои брони</div>} />
            <Route path="profile" element={<div>Профиль</div>} />
            <Route path="wallet" element={<div>Кошелёк</div>} />
            <Route path="payments" element={<div>Платежи</div>} />
            <Route path="cards" element={<div>Карты</div>} />
            <Route path="notifications" element={<div>Уведомления</div>} />
            <Route path="security" element={<div>Безопасность</div>} />
            <Route path="favorites" element={<div>Избранное</div>} />
            <Route path="chat" element={<div>Чат</div>} />
            <Route path="tickets" element={<div>Поддержка</div>} />
            <Route path="disputes" element={<div>Споры</div>} />
          </Route>
        </Routes>
      </MemoryRouter>
    </ConfigProvider>,
  )
}

describe('Shell navigation', () => {
  beforeEach(() => {
    vi.mocked(useGetCities).mockReturnValue({
      data: { data: [{ id: 1, name: 'Москва', slug: 'moscow' }] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetCities>)

    useAuthStore.setState({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
    })
  })

  it('renders public level-0 navigation and premium footer', () => {
    mockDesktopViewport()
    renderPublicShell('/catalog')

    expect(screen.getAllByText('RelaxHUB').length).toBeGreaterThan(0)
    expect(screen.getByRole('button', { name: 'Каталог' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Сертификаты' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Помощь' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Контакты' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Разделы' })).not.toBeInTheDocument()

    expect(screen.getByText('Terms of Use')).toBeInTheDocument()
    expect(screen.getByText(/© relaxhub/i)).toBeInTheDocument()
    expect(screen.getByText('support@relaxhub.ru')).toBeInTheDocument()
    expect(screen.getByText('+7 (495) 555-21-21')).toBeInTheDocument()
    expect(screen.getByText('@relaxhub_reserve')).toBeInTheDocument()
  })

  it('renders client level-0 navigation and grouped profile menu', async () => {
    mockDesktopViewport()
    useAuthStore.setState({
      user: { id: 'client-1', role: 'client', name: 'Мария Иванова', email: 'maria@example.com', onboarding_completed: true },
      token: 'jwt-token',
      isAuthenticated: true,
      isLoading: false,
    })

    renderClientShell()

    expect(screen.getByRole('button', { name: 'Каталог' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Мои брони' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Сертификаты' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Помощь' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Контакты' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Разделы' })).not.toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /Мария Иванова/i }))

    const profileMenu = await screen.findByRole('menu')

    expect(within(profileMenu).getByText('Аккаунт')).toBeInTheDocument()
    expect(within(profileMenu).getByText('Финансы')).toBeInTheDocument()
    expect(within(profileMenu).getByText('Профиль')).toBeInTheDocument()
    expect(within(profileMenu).getByText('Кошелёк')).toBeInTheDocument()
    expect(within(profileMenu).getByText('Платежи')).toBeInTheDocument()
    expect(within(profileMenu).getByText('Карты')).toBeInTheDocument()
    expect(within(profileMenu).getByText('Уведомления')).toBeInTheDocument()
    expect(within(profileMenu).getByText('Безопасность')).toBeInTheDocument()
    expect(within(profileMenu).getByText('Выйти')).toBeInTheDocument()
  })

  it('separates navigation, profile and service blocks in mobile drawer', async () => {
    mockMobileViewport()
    useAuthStore.setState({
      user: { id: 'client-1', role: 'client', name: 'Мария Иванова', email: 'maria@example.com', onboarding_completed: true },
      token: 'jwt-token',
      isAuthenticated: true,
      isLoading: false,
    })

    renderClientShell()

    fireEvent.click(screen.getByRole('button', { name: 'Открыть меню' }))

    const drawer = await screen.findByRole('dialog')

    expect(within(drawer).getByText('Навигация')).toBeInTheDocument()
    expect(within(drawer).getByText('Аккаунт и финансы')).toBeInTheDocument()
    expect(within(drawer).getByText('Сервисы')).toBeInTheDocument()
    expect(within(drawer).getByText('Избранное')).toBeInTheDocument()
    expect(within(drawer).getByText('Чат')).toBeInTheDocument()
    expect(within(drawer).getByText('Поддержка')).toBeInTheDocument()
    expect(within(drawer).getByText('Споры')).toBeInTheDocument()
  })
})
