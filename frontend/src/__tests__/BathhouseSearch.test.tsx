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

import { useGetBathhouses } from '@/api/generated/bathhouses/bathhouses'
import { useGetCities } from '@/api/generated/cities/cities'
import { usePostBathhousesIdFavorite } from '@/api/generated/favorites/favorites'

function renderWithProviders(ui: React.ReactElement, route = '/catalog') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={[route]}>{ui}</MemoryRouter>
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
    has_pool: false,
    images: ['https://example.com/p1.jpg'],
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
  },
]

const mockCities = [
  { id: 1, name: 'Москва', slug: 'moscow' },
  { id: 2, name: 'Санкт-Петербург', slug: 'spb' },
]

describe('BathhouseSearch', () => {
  beforeEach(() => {
    vi.mocked(useGetCities).mockReturnValue({
      data: { data: mockCities, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetCities>)

    vi.mocked(usePostBathhousesIdFavorite).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostBathhousesIdFavorite>)
  })

  it('renders search title', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: { data: mockBathhouses, success: true, meta: { page: 1, page_size: 12, total_count: 2, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByText('Поиск бань')).toBeInTheDocument()
  })

  it('renders bathhouse cards', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: { data: mockBathhouses, success: true, meta: { page: 1, page_size: 12, total_count: 2, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByText('Баня Классик')).toBeInTheDocument()
    expect(screen.getByText('Баня Люкс')).toBeInTheDocument()
  })

  it('shows empty state when no results', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: { data: [], success: true, meta: { page: 1, page_size: 12, total_count: 0, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByText(/По вашему запросу ничего не найдено/)).toBeInTheDocument()
  })

  it('renders search input', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByPlaceholderText('Поиск по названию...')).toBeInTheDocument()
  })

  it('renders city filter', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByText('Город')).toBeInTheDocument()
  })

  it('syncs city selector with city_slug from URL', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />, '/catalog?city_slug=spb')

    expect(screen.getByText('Санкт-Петербург')).toBeInTheDocument()
  })

  it('renders geo search button', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByText('Найти рядом')).toBeInTheDocument()
  })

  it('renders filter section toggle', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByText('Фильтры')).toBeInTheDocument()
  })

  it('shows amenity filters when expanded', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    fireEvent.click(screen.getByText('Фильтры'))

    expect(screen.getByText('Сауна')).toBeInTheDocument()
    expect(screen.getByText('Бассейн')).toBeInTheDocument()
    expect(screen.getByText('Караоке')).toBeInTheDocument()
  })

  it('shows pagination when multiple pages', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: {
        data: mockBathhouses,
        success: true,
        meta: { page: 1, page_size: 12, total_count: 30, total_pages: 3 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    const pagination = document.querySelector('.ant-pagination')
    expect(pagination).toBeInTheDocument()
  })

  it('displays live result counter with total count', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: {
        data: mockBathhouses,
        success: true,
        meta: { page: 1, page_size: 12, total_count: 42, total_pages: 4 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    const counter = screen.getByTestId('result-counter')
    expect(counter).toBeInTheDocument()
    expect(counter).toHaveTextContent('Найдено: 42 бани')
  })

  it('shows correct pluralization for result counter', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: {
        data: [mockBathhouses[0]],
        success: true,
        meta: { page: 1, page_size: 12, total_count: 1, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByTestId('result-counter')).toHaveTextContent('Найдено: 1 баня')
  })

  it('shows counter with 0 results', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: {
        data: [],
        success: true,
        meta: { page: 1, page_size: 12, total_count: 0, total_pages: 0 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByTestId('result-counter')).toHaveTextContent('Найдено: 0 бань')
  })

  it('debounces search input before querying', async () => {
    vi.useFakeTimers()

    vi.mocked(useGetBathhouses).mockReturnValue({
      data: { data: [], success: true, meta: { page: 1, page_size: 12, total_count: 0, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    const input = screen.getByPlaceholderText('Поиск по названию...')
    fireEvent.change(input, { target: { value: 'люкс' } })

    // Before debounce, query should still use empty string (debouncedSearch hasn't updated)
    const callsBefore = vi.mocked(useGetBathhouses).mock.calls
    const lastCallBefore = callsBefore[callsBefore.length - 1]!
    expect(lastCallBefore[0]).toHaveProperty('q', undefined)

    // After debounce timer fires
    vi.advanceTimersByTime(300)

    vi.useRealTimers()
  })
})
