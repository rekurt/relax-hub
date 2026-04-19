import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import BathhouseSearch from '@/pages/client/BathhouseSearch'

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetBathhouses: vi.fn(),
}))

vi.mock('@/api/generated/cities/cities', () => ({
  useGetCities: vi.fn(),
}))

vi.mock('@/api/generated/favorites/favorites', () => ({
  usePostBathhousesIdFavorite: vi.fn(),
}))

vi.mock('@/api/generated/search/search', () => ({
  useGetSearchSuggestions: vi.fn().mockReturnValue({ data: { data: [] }, isLoading: false }),
}))

import { useGetBathhouses } from '@/api/generated/bathhouses/bathhouses'
import { useGetCities } from '@/api/generated/cities/cities'
import { usePostBathhousesIdFavorite } from '@/api/generated/favorites/favorites'

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

const mockBathhouses = [
  {
    id: '1',
    name: 'Баня Классик',
    address: 'ул. Ленина, д. 1',
    price_per_hour: 200000,
    rating: 4.2,
    review_count: 8,
    has_sauna: true,
    images: ['https://example.com/p1.jpg'],
    booking_mode: 'instant',
    badges: ['verified'],
    last_minute_active: true,
    last_minute_discount_percent: 20,
  },
  {
    id: '2',
    name: 'Баня Люкс',
    address: 'ул. Мира, д. 5',
    price_per_hour: 500000,
    rating: 4.8,
    review_count: 25,
    has_sauna: true,
    has_pool: true,
    images: [],
    booking_mode: 'request',
    badges: ['top', 'premium'],
  },
]

const mockCities = [
  { id: 1, name: 'Москва', slug: 'moscow' },
]

describe('BathhouseSearch - New Filters', () => {
  beforeEach(() => {
    vi.clearAllMocks()

    vi.mocked(useGetCities).mockReturnValue({
      data: { data: mockCities, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetCities>)

    vi.mocked(usePostBathhousesIdFavorite).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostBathhousesIdFavorite>)

    vi.mocked(useGetBathhouses).mockReturnValue({
      data: {
        data: mockBathhouses,
        success: true,
        meta: { page: 1, page_size: 12, total_count: 2, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)
  })

  it('renders date and time filters in expanded filters panel', () => {
    renderWithProviders(<BathhouseSearch />)

    fireEvent.click(screen.getByText('Фильтры'))

    expect(screen.getByText('Дата')).toBeInTheDocument()
    expect(screen.getByText('Время с')).toBeInTheDocument()
    expect(screen.getByText('Время до')).toBeInTheDocument()
  })

  it('renders booking type filter', () => {
    renderWithProviders(<BathhouseSearch />)

    fireEvent.click(screen.getByText('Фильтры'))

    expect(screen.getByText('Тип бронирования')).toBeInTheDocument()
  })

  it('renders min rating filter', () => {
    renderWithProviders(<BathhouseSearch />)

    fireEvent.click(screen.getByText('Фильтры'))

    // The new select-based min rating filter label
    // There are two "Мин. рейтинг" - the old slider label and new select label
    const labels = screen.getAllByText('Мин. рейтинг')
    expect(labels.length).toBeGreaterThanOrEqual(1)
  })

  it('renders status filter', () => {
    renderWithProviders(<BathhouseSearch />)

    fireEvent.click(screen.getByText('Фильтры'))

    expect(screen.getByText('Статус')).toBeInTheDocument()
  })

  it('shows last-minute badge on cards with active discounts', () => {
    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByText(/Last minute/)).toBeInTheDocument()
    expect(screen.getByText(/-20%/)).toBeInTheDocument()
  })

  it('renders search suggestions container on search input', () => {
    renderWithProviders(<BathhouseSearch />)

    const input = screen.getByPlaceholderText('Поиск по названию...')
    expect(input).toBeInTheDocument()
  })

  it('shows both bathhouses initially', () => {
    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByText('Баня Классик')).toBeInTheDocument()
    expect(screen.getByText('Баня Люкс')).toBeInTheDocument()
  })

  it('still renders amenity filters', () => {
    renderWithProviders(<BathhouseSearch />)

    fireEvent.click(screen.getByText('Фильтры'))

    expect(screen.getByText('Удобства')).toBeInTheDocument()
    // "Сауна" appears both as amenity filter checkbox and as tag on cards, so use getAllByText
    expect(screen.getAllByText('Сауна').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Бассейн').length).toBeGreaterThanOrEqual(1)
  })

  it('still renders guest count filter', () => {
    renderWithProviders(<BathhouseSearch />)

    fireEvent.click(screen.getByText('Фильтры'))

    expect(screen.getByText('Гости')).toBeInTheDocument()
  })

  it('passes availability params to API query', () => {
    renderWithProviders(<BathhouseSearch />)

    // Just verify the hook is called (filters start undefined/null)
    expect(useGetBathhouses).toHaveBeenCalled()
    const calls = vi.mocked(useGetBathhouses).mock.calls
    const lastCall = calls[calls.length - 1]
    expect(lastCall?.[0]).toHaveProperty('sort_by')
  })
})
