import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import ClientHome from '@/pages/client/ClientHome'

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    get: vi.fn().mockRejectedValue(new Error('not found')),
    post: vi.fn().mockResolvedValue({ data: { success: true } }),
  },
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

vi.mock('@/components/OnboardingTour', () => ({
  default: ({ open, onComplete }: { open: boolean; onComplete: () => void }) =>
    open ? (
      <div data-testid="onboarding-tour">
        <button onClick={onComplete}>Начать</button>
      </div>
    ) : null,
}))

import { axiosInstance } from '@/api/axios-instance'
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

const defaultUser = {
  id: 'u1',
  role: 'client',
  onboarding_completed: true,
  region: 'RU',
}

function mockAuthStore(userOverride?: Record<string, unknown> | null) {
  const u = userOverride === null ? null : { ...defaultUser, ...userOverride }
  const state = {
    user: u,
    token: u ? 'tok' : null,
    setAuth: vi.fn(),
    logout: vi.fn(),
    isLoading: false,
    isAuthenticated: !!u,
    loadProfile: vi.fn(),
  }
  const mockFn = vi.mocked(useAuthStore)
  mockFn.mockImplementation((selector) => {
    return (selector as (s: typeof state) => unknown)(state)
  })
  // Support imperative getState() used by handleTourComplete
  ;(mockFn as unknown as Record<string, unknown>).getState = () => state
}

describe('ClientHome', () => {
  let originalGeolocation: Geolocation

  beforeEach(() => {
    mockAuthStore()

    vi.mocked(useGetMyRecentlyViewed).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyRecentlyViewed>)

    vi.mocked(useGetRecommendations).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetRecommendations>)

    vi.mocked(useGetCities).mockReturnValue({
      data: { data: [{ id: 1, name: 'Москва', slug: 'moscow' }, { id: 2, name: 'Минск', slug: 'minsk' }], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetCities>)

    vi.mocked(useGetPopular).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetPopular>)

    vi.mocked(axiosInstance.get).mockRejectedValue(new Error('not found'))

    // Save and mock geolocation
    originalGeolocation = navigator.geolocation
    Object.defineProperty(navigator, 'geolocation', {
      value: {
        getCurrentPosition: vi.fn((_success: unknown, error: (err: { code: number; message: string }) => void) => {
          // Default: GPS unavailable
          error({ code: 1, message: 'User denied Geolocation' })
        }),
      },
      writable: true,
      configurable: true,
    })
  })

  afterEach(() => {
    Object.defineProperty(navigator, 'geolocation', {
      value: originalGeolocation,
      writable: true,
      configurable: true,
    })
  })

  it('renders search bar', () => {
    renderWithProviders(<ClientHome />)
    expect(screen.getByPlaceholderText('Поиск бань...')).toBeInTheDocument()
  })

  it('renders promo banner with fallback text for returning user', () => {
    renderWithProviders(<ClientHome />)
    expect(screen.getByText('Бани для вечера вдвоем, компании и выходных за городом')).toBeInTheDocument()
  })

  it('renders welcome message for new user', () => {
    mockAuthStore({ onboarding_completed: false })
    renderWithProviders(<ClientHome />)
    expect(screen.getByText('Вечер уже можно планировать')).toBeInTheDocument()
  })

  it('shows onboarding tour modal for new users', () => {
    mockAuthStore({ onboarding_completed: false })
    renderWithProviders(<ClientHome />)
    expect(screen.getByTestId('onboarding-tour')).toBeInTheDocument()
  })

  it('does not show onboarding tour for completed users', () => {
    mockAuthStore({ onboarding_completed: true })
    renderWithProviders(<ClientHome />)
    expect(screen.queryByTestId('onboarding-tour')).not.toBeInTheDocument()
  })

  it('closes onboarding tour when completed', () => {
    mockAuthStore({ onboarding_completed: false })
    renderWithProviders(<ClientHome />)
    expect(screen.getByTestId('onboarding-tour')).toBeInTheDocument()

    fireEvent.click(screen.getByText('Начать'))
    expect(screen.queryByTestId('onboarding-tour')).not.toBeInTheDocument()
  })

  it('does not show noisy geolocation warning when GPS is unavailable', () => {
    renderWithProviders(<ClientHome />)
    expect(screen.queryByText(/Не удалось определить местоположение/)).not.toBeInTheDocument()
  })

  it('does not show city picker when GPS succeeds', () => {
    Object.defineProperty(navigator, 'geolocation', {
      value: {
        getCurrentPosition: vi.fn((success) => {
          success({ coords: { latitude: 55.75, longitude: 37.62 } })
        }),
      },
      writable: true,
      configurable: true,
    })

    // Cities with coordinates so geo detection works
    vi.mocked(useGetCities).mockReturnValue({
      data: {
        data: [
          { id: 1, name: 'Москва', slug: 'moscow', latitude: 55.75, longitude: 37.62 },
          { id: 2, name: 'Минск', slug: 'minsk', latitude: 53.9, longitude: 27.56 },
        ],
        success: true,
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetCities>)

    renderWithProviders(<ClientHome />)
    expect(screen.queryByText(/Не удалось определить местоположение/)).not.toBeInTheDocument()
  })

  it('prefers Moscow as default public city when GPS is unavailable', () => {
    vi.mocked(useGetCities).mockReturnValue({
      data: {
        data: [
          { id: 2, name: 'Красногорск', slug: 'krasnogorsk', latitude: 55.83, longitude: 37.33 },
          { id: 1, name: 'Москва', slug: 'moscow', latitude: 55.75, longitude: 37.62 },
        ],
        success: true,
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetCities>)

    renderWithProviders(<ClientHome />)

    expect(screen.getByTestId('popular-city-caption')).toHaveTextContent('Москва по умолчанию')
    const popularCalls = vi.mocked(useGetPopular).mock.calls
    const lastPopularCall = popularCalls[popularCalls.length - 1]
    expect(lastPopularCall?.[0]).toMatchObject({ city_id: 1, limit: 6 })
  })

  it('renders promo banner from API when available', async () => {
    vi.mocked(axiosInstance.get).mockImplementation((url: string) => {
      if (url === '/promotions/banners') {
        return Promise.resolve({
          data: {
            success: true,
            data: [
              {
                id: 'p1',
                title: 'Скидка 20%',
                description: 'На первое бронирование',
                type: 'promo',
                promo_code: 'FIRST20',
                discount_text: '20% скидка',
              },
            ],
          },
        })
      }
      return Promise.reject(new Error('not found'))
    })

    renderWithProviders(<ClientHome />)

    expect(await screen.findByText('Скидка 20%')).toBeInTheDocument()
    expect(screen.getByText('На первое бронирование')).toBeInTheDocument()
    expect(screen.getByText(/FIRST20/)).toBeInTheDocument()
  })

  it('renders multiple banners in carousel', async () => {
    vi.mocked(axiosInstance.get).mockImplementation((url: string) => {
      if (url === '/promotions/banners') {
        return Promise.resolve({
          data: {
            success: true,
            data: [
              { id: 'p1', title: 'Акция 1', description: 'Описание 1', type: 'promo' },
              { id: 'p2', title: 'Акция 2', description: 'Описание 2', type: 'loyalty' },
            ],
          },
        })
      }
      return Promise.reject(new Error('not found'))
    })

    renderWithProviders(<ClientHome />)

    // Carousel clones slides so text may appear multiple times
    const items = await screen.findAllByText('Акция 1')
    expect(items.length).toBeGreaterThan(0)
    expect(screen.getAllByText('Акция 2').length).toBeGreaterThan(0)
  })

  it('renders city selector in popular section when multiple cities', () => {
    renderWithProviders(<ClientHome />)
    expect(screen.getByText('Популярные рядом')).toBeInTheDocument()
  })

  it('shows empty state when no popular bathhouses in selected city', () => {
    vi.mocked(useGetPopular).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetPopular>)

    renderWithProviders(<ClientHome />)
    expect(screen.getByText('В выбранном городе пока нет популярных бань')).toBeInTheDocument()
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
})
