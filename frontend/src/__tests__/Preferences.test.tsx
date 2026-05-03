import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import Preferences from '@/pages/client/Preferences'

vi.mock('@/api/generated/recommendations/recommendations', () => ({
  useGetMyPreferences: vi.fn(),
  usePutMyPreferences: vi.fn(),
}))

vi.mock('@/api/generated/cities/cities', () => ({
  useGetCities: vi.fn(),
}))

import { useGetMyPreferences, usePutMyPreferences } from '@/api/generated/recommendations/recommendations'
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

const mockPreferences = {
  prefer_sauna: true,
  prefer_steam_room: false,
  prefer_pool: true,
  prefer_hot_tub: false,
  prefer_bbq: false,
  prefer_karaoke: false,
  preferred_city_id: 1,
  price_range_min: 100000,
  price_range_max: 500000,
}

const mockCities = [
  { id: 1, name: 'Москва', slug: 'moscow' },
  { id: 2, name: 'Санкт-Петербург', slug: 'spb' },
]

describe('Preferences', () => {
  const mockMutate = vi.fn()

  beforeEach(() => {
    vi.mocked(useGetCities).mockReturnValue({
      data: { data: mockCities, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetCities>)

    vi.mocked(usePutMyPreferences).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
    } as unknown as ReturnType<typeof usePutMyPreferences>)
  })

  it('renders preferences title', () => {
    vi.mocked(useGetMyPreferences).mockReturnValue({
      data: { data: mockPreferences, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyPreferences>)

    renderWithProviders(<Preferences />)
    expect(screen.getByText('Настройки рекомендаций')).toBeInTheDocument()
  })

  it('renders amenity toggles', () => {
    vi.mocked(useGetMyPreferences).mockReturnValue({
      data: { data: mockPreferences, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyPreferences>)

    renderWithProviders(<Preferences />)
    expect(screen.getAllByText('Сауна').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Бассейн').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Караоке').length).toBeGreaterThan(0)
  })

  it('renders city selector', () => {
    vi.mocked(useGetMyPreferences).mockReturnValue({
      data: { data: mockPreferences, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyPreferences>)

    renderWithProviders(<Preferences />)
    expect(screen.getByText('Предпочитаемый город')).toBeInTheDocument()
  })

  it('renders price range inputs', () => {
    vi.mocked(useGetMyPreferences).mockReturnValue({
      data: { data: mockPreferences, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyPreferences>)

    renderWithProviders(<Preferences />)
    expect(screen.getByText('От (руб/ч)')).toBeInTheDocument()
    expect(screen.getByText('До (руб/ч)')).toBeInTheDocument()
  })

  it('renders save button', () => {
    vi.mocked(useGetMyPreferences).mockReturnValue({
      data: { data: mockPreferences, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyPreferences>)

    renderWithProviders(<Preferences />)
    expect(screen.getByText('Сохранить предпочтения')).toBeInTheDocument()
  })

  it('shows loading state', () => {
    vi.mocked(useGetMyPreferences).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyPreferences>)

    renderWithProviders(<Preferences />)
    expect(document.querySelector('.ant-spin')).toBeInTheDocument()
  })

  it('renders section cards', () => {
    vi.mocked(useGetMyPreferences).mockReturnValue({
      data: { data: mockPreferences, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyPreferences>)

    renderWithProviders(<Preferences />)
    expect(screen.getByText('Удобства')).toBeInTheDocument()
    expect(screen.getByText('Город')).toBeInTheDocument()
    expect(screen.getByText('Ценовой диапазон')).toBeInTheDocument()
  })
})
