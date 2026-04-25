import { useEffect } from 'react'
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, useLocation } from 'react-router-dom'
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

function LocationProbe({ onChange }: { onChange: (location: string) => void }) {
  const location = useLocation()

  useEffect(() => {
    onChange(`${location.pathname}${location.search}`)
  }, [location.pathname, location.search, onChange])

  return null
}

function renderWithProviders(ui: React.ReactElement, route = '/catalog', onLocationChange?: (location: string) => void) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={[route]}>
            {ui}
            {onLocationChange ? <LocationProbe onChange={onLocationChange} /> : null}
          </MemoryRouter>
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
      isError: false,
      refetch: vi.fn(),
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

    expect(screen.getByText(/Попробуйте изменить параметры поиска или сбросить фильтры/)).toBeInTheDocument()
  })

  it('shows backend unavailable state when catalog and city endpoints fail together', () => {
    vi.mocked(useGetCities).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: { response: { status: 500 } },
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetCities>)

    vi.mocked(useGetBathhouses).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: { response: { status: 500 } },
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByText('Каталог временно недоступен')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Повторить' })).toBeInTheDocument()
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

  it('renders promoted shortcut scenarios including pool selection', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByText('Для двоих')).toBeInTheDocument()
    expect(screen.getAllByText('С бассейном').length).toBeGreaterThan(0)
  })

  it('hydrates preset catalog filters from URL without dropping them on first render', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: { data: [], success: true, meta: { page: 1, page_size: 12, total_count: 0, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />, '/catalog?guest_count=2')

    const calls = vi.mocked(useGetBathhouses).mock.calls
    const lastCall = calls[calls.length - 1]?.[0] as Record<string, unknown> | undefined
    expect(lastCall).toMatchObject({ guest_count: 2 })
  })

  it('does not restore previous shortcut params when switching catalog shortcuts', async () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: { data: [], success: true, meta: { page: 1, page_size: 12, total_count: 0, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    const locations: string[] = []
    let sawPoolLocation = false
    renderWithProviders(<BathhouseSearch />, '/catalog?guest_count=6', (location) => {
      locations.push(location)
      if (location === '/catalog?has_pool=true') sawPoolLocation = true
      if (sawPoolLocation && location === '/catalog?guest_count=6') {
        throw new Error('Previous shortcut params were restored after selecting a new shortcut')
      }
    })

    const shortcuts = screen.getByLabelText('Сценарии подбора')
    fireEvent.click(within(shortcuts).getByRole('button', { name: /С бассейном/ }))

    await waitFor(() => expect(locations).toContain('/catalog?has_pool=true'))
    await new Promise((resolve) => setTimeout(resolve, 20))

    const poolLocationIndex = locations.findIndex((location) => location === '/catalog?has_pool=true')
    expect(locations.slice(poolLocationIndex + 1)).not.toContain('/catalog?guest_count=6')
  })

  it('renders quick filters row', () => {
    vi.mocked(useGetBathhouses).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)

    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByText('Открыто сейчас')).toBeInTheDocument()
    expect(screen.getByText('Рейтинг 4+')).toBeInTheDocument()
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

})
