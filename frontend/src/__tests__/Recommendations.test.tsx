import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import Recommendations from '@/pages/client/Recommendations'

vi.mock('@/api/generated/recommendations/recommendations', () => ({
  useGetRecommendations: vi.fn(),
  useGetPopular: vi.fn(),
}))

vi.mock('@/api/generated/cities/cities', () => ({
  useGetCities: vi.fn(),
}))

import { useGetRecommendations, useGetPopular } from '@/api/generated/recommendations/recommendations'
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

const mockRecommendations = [
  { id: '1', name: 'Баня Рекомендованная', address: 'ул. Тестовая, 1', price_per_hour: 300000, rating: 4.5, review_count: 12 },
  { id: '2', name: 'Баня Популярная', address: 'ул. Мира, 5', price_per_hour: 200000, rating: 4.0, review_count: 8 },
]

const mockPopular = [
  { id: '3', name: 'Топ Баня', address: 'пр. Ленина, 10', price_per_hour: 400000, rating: 4.9, review_count: 50 },
]

const mockCities = [
  { id: 1, name: 'Москва', slug: 'moscow' },
  { id: 2, name: 'Санкт-Петербург', slug: 'spb' },
]

describe('Recommendations', () => {
  beforeEach(() => {
    vi.mocked(useGetCities).mockReturnValue({
      data: { data: mockCities, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetCities>)

    vi.mocked(useGetPopular).mockReturnValue({
      data: { data: mockPopular, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetPopular>)
  })

  it('renders recommendations title', () => {
    vi.mocked(useGetRecommendations).mockReturnValue({
      data: { data: mockRecommendations, success: true, meta: { page: 1, page_size: 12, total_count: 2, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetRecommendations>)

    renderWithProviders(<Recommendations />)
    expect(screen.getByText('Рекомендации для вас')).toBeInTheDocument()
  })

  it('renders recommendation cards', () => {
    vi.mocked(useGetRecommendations).mockReturnValue({
      data: { data: mockRecommendations, success: true, meta: { page: 1, page_size: 12, total_count: 2, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetRecommendations>)

    renderWithProviders(<Recommendations />)
    expect(screen.getByText('Баня Рекомендованная')).toBeInTheDocument()
    expect(screen.getByText('Баня Популярная')).toBeInTheDocument()
  })

  it('renders popular section', () => {
    vi.mocked(useGetRecommendations).mockReturnValue({
      data: { data: [], success: true, meta: { page: 1, page_size: 12, total_count: 0, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetRecommendations>)

    renderWithProviders(<Recommendations />)
    expect(screen.getByText('Популярные бани')).toBeInTheDocument()
    expect(screen.getByText('Топ Баня')).toBeInTheDocument()
  })

  it('shows empty state when no recommendations', () => {
    vi.mocked(useGetRecommendations).mockReturnValue({
      data: { data: [], success: true, meta: { page: 1, page_size: 12, total_count: 0, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetRecommendations>)

    renderWithProviders(<Recommendations />)
    expect(screen.getByText(/Пока нет персональных рекомендаций/)).toBeInTheDocument()
  })

  it('renders city selector for popular section', () => {
    vi.mocked(useGetRecommendations).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetRecommendations>)

    renderWithProviders(<Recommendations />)
    expect(screen.getByText('Москва')).toBeInTheDocument()
  })

  it('shows price on recommendation cards', () => {
    vi.mocked(useGetRecommendations).mockReturnValue({
      data: { data: mockRecommendations, success: true, meta: { page: 1, page_size: 12, total_count: 2, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetRecommendations>)

    renderWithProviders(<Recommendations />)
    expect(screen.getByText('3000 ₽/ч')).toBeInTheDocument()
  })
})
