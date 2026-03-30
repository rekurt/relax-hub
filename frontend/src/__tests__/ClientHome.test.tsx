import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ClientHome from '@/pages/client/ClientHome'

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: { get: vi.fn().mockRejectedValue(new Error('not found')) },
}))

vi.mock('@/api/generated/saved-searches/saved-searches', () => ({
  useGetMyRecentlyViewed: vi.fn(),
}))

vi.mock('@/api/generated/recommendations/recommendations', () => ({
  useGetRecommendations: vi.fn(),
  useGetPopular: vi.fn(),
}))

vi.mock('@/api/generated/cities/cities', () => ({
  useGetCities: vi.fn(),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn(),
}))

import { useGetMyRecentlyViewed } from '@/api/generated/saved-searches/saved-searches'
import { useGetRecommendations, useGetPopular } from '@/api/generated/recommendations/recommendations'
import { useGetCities } from '@/api/generated/cities/cities'
import { useAuthStore } from '@/stores/auth'

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

describe('ClientHome', () => {
  beforeEach(() => {
    vi.mocked(useAuthStore).mockImplementation((selector) => {
      const state = { user: { id: 'u1', role: 'client' }, token: 'tok', setAuth: vi.fn(), logout: vi.fn(), isLoading: false, isAuthenticated: true, loadProfile: vi.fn() }
      return (selector as (s: typeof state) => unknown)(state)
    })

    vi.mocked(useGetMyRecentlyViewed).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyRecentlyViewed>)

    vi.mocked(useGetRecommendations).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetRecommendations>)

    vi.mocked(useGetCities).mockReturnValue({
      data: { data: [{ id: 1, name: 'Москва', slug: 'moscow' }], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetCities>)

    vi.mocked(useGetPopular).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetPopular>)
  })

  it('renders promo banner', () => {
    renderWithProviders(<ClientHome />)
    expect(screen.getByText('Добро пожаловать в Bani!')).toBeInTheDocument()
  })

  it('renders search bar', () => {
    renderWithProviders(<ClientHome />)
    expect(screen.getByPlaceholderText('Поиск бань...')).toBeInTheDocument()
  })

  it('renders recently viewed when items exist', () => {
    vi.mocked(useGetMyRecentlyViewed).mockReturnValue({
      data: {
        data: [
          { id: '1', name: 'Тест Баня', slug: 'test', viewed_at: '2026-03-29T12:00:00Z' },
        ],
        success: true,
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyRecentlyViewed>)

    renderWithProviders(<ClientHome />)
    expect(screen.getByText('Недавно просмотренные')).toBeInTheDocument()
    expect(screen.getByText('Тест Баня')).toBeInTheDocument()
  })

  it('renders recommendations when available', () => {
    vi.mocked(useGetRecommendations).mockReturnValue({
      data: {
        data: [
          { id: '1', name: 'Рекомендованная Баня', address: 'ул. Тестовая', price_per_hour: 300000, rating: 4.5, review_count: 10 },
        ],
        success: true,
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetRecommendations>)

    renderWithProviders(<ClientHome />)
    expect(screen.getByText('Рекомендации для вас')).toBeInTheDocument()
    expect(screen.getByText('Рекомендованная Баня')).toBeInTheDocument()
  })

  it('renders popular nearby section', () => {
    vi.mocked(useGetPopular).mockReturnValue({
      data: {
        data: [
          { id: '1', name: 'Популярная Баня', address: 'пр. Ленина', price_per_hour: 200000, rating: 4.8, review_count: 50 },
        ],
        success: true,
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetPopular>)

    renderWithProviders(<ClientHome />)
    expect(screen.getByText('Популярные рядом')).toBeInTheDocument()
    expect(screen.getByText('Популярная Баня')).toBeInTheDocument()
  })

  it('shows welcome bonus text in promo banner', () => {
    renderWithProviders(<ClientHome />)
    expect(screen.getByText(/Бонус 500 ₽ новым пользователям/)).toBeInTheDocument()
  })
})
