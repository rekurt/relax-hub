import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import BathhouseSearch from '@/pages/client/BathhouseSearch'

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetBathhouses: vi.fn(),
  usePostApiV1BathhousesCompare: vi.fn().mockReturnValue({ mutate: vi.fn(), isPending: false }),
}))

vi.mock('@/api/generated/cities/cities', () => ({
  useGetCities: vi.fn(),
}))

vi.mock('@/api/generated/favorites/favorites', () => ({
  usePostBathhousesIdFavorite: vi.fn(),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn(),
}))

vi.mock('@/components/BathhouseMap', () => ({
  default: ({
    bathhouses,
    onBoundsChange,
    onMarkerHover,
    showMiniCard,
    highlightedId,
  }: {
    bathhouses: { id?: string; name?: string }[]
    onBoundsChange?: (b: unknown) => void
    onMarkerHover?: (id: string | null) => void
    showMiniCard?: boolean
    highlightedId?: string | null
  }) => (
    <div data-testid="bathhouse-map">
      Map with {bathhouses.length} markers
      {showMiniCard && <span data-testid="mini-card-enabled" />}
      {highlightedId && <span data-testid="map-highlighted">{highlightedId}</span>}
      {onBoundsChange && (
        <button
          data-testid="search-area-btn"
          onClick={() =>
            onBoundsChange({ north: 56, south: 55, east: 38, west: 37 })
          }
        >
          Search area
        </button>
      )}
      {onMarkerHover && bathhouses.map((b) => (
        <div
          key={b.id}
          data-testid={`map-marker-${b.id}`}
          onMouseEnter={() => onMarkerHover(b.id ?? null)}
          onMouseLeave={() => onMarkerHover(null)}
        >
          {b.name}
        </div>
      ))}
    </div>
  ),
}))

import { useGetBathhouses } from '@/api/generated/bathhouses/bathhouses'
import { useGetCities } from '@/api/generated/cities/cities'
import { usePostBathhousesIdFavorite } from '@/api/generated/favorites/favorites'
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
    latitude: 55.75,
    longitude: 37.62,
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
    latitude: 55.76,
    longitude: 37.63,
  },
  {
    id: '3',
    name: 'Баня Эконом',
    address: 'ул. Гагарина, д. 10',
    price_per_hour: 100000,
    rating: 3.5,
    review_count: 5,
    has_sauna: false,
    has_pool: false,
    images: [],
    latitude: 55.77,
    longitude: 37.64,
  },
]

const mockCities = [
  { id: 1, name: 'Москва', slug: 'moscow' },
]

describe('BathhouseSearch - Map & Compare features', () => {
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

    vi.mocked(useAuthStore).mockImplementation((selector) =>
      selector({
        user: { id: 'client-1', role: 'client', name: 'Иван' } as never,
        token: 'jwt-token',
        isLoading: false,
        isAuthenticated: true,
        setAuth: vi.fn(),
        setUser: vi.fn(),
        logout: vi.fn(),
        loadProfile: vi.fn(),
        notice2faRequiredAt: 0,
        bumpNotice2FARequired: vi.fn(),
      }),
    )

    vi.mocked(useGetBathhouses).mockReturnValue({
      data: {
        data: mockBathhouses,
        success: true,
        meta: { page: 1, page_size: 12, total_count: 3, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhouses>)
  })

  it('renders view mode switcher', () => {
    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByText('Список')).toBeInTheDocument()
    expect(screen.getByText('Сплит')).toBeInTheDocument()
    expect(screen.getByText('Карта')).toBeInTheDocument()
  })

  it('shows list view by default', () => {
    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByText('Баня Классик')).toBeInTheDocument()
    expect(screen.getByText('Баня Люкс')).toBeInTheDocument()
    expect(screen.queryByTestId('bathhouse-map')).not.toBeInTheDocument()
  })

  it('switches to map view', () => {
    renderWithProviders(<BathhouseSearch />)

    fireEvent.click(screen.getByText('Карта'))

    expect(screen.getByTestId('bathhouse-map')).toBeInTheDocument()
    expect(screen.getByText('Map with 3 markers')).toBeInTheDocument()
  })

  it('switches to split view showing both list and map', () => {
    renderWithProviders(<BathhouseSearch />)

    fireEvent.click(screen.getByText('Сплит'))

    expect(screen.getByTestId('bathhouse-map')).toBeInTheDocument()
    // Both list and map markers show bathhouse names, so use getAllByText
    expect(screen.getAllByText('Баня Классик').length).toBeGreaterThanOrEqual(1)
  })

  it('shows compare checkboxes on cards', () => {
    renderWithProviders(<BathhouseSearch />)

    const compareLabels = screen.getAllByText('Сравнить')
    expect(compareLabels.length).toBeGreaterThanOrEqual(3)
  })

  it('shows comparison bar when items selected', () => {
    renderWithProviders(<BathhouseSearch />)

    // Click first compare checkbox
    const compareLabels = screen.getAllByText('Сравнить')
    fireEvent.click(compareLabels[0]!)

    expect(screen.getByText(/Выбрано для сравнения/)).toBeInTheDocument()
    expect(screen.getByText('Сбросить')).toBeInTheDocument()
  })

  it('enables compare button when 2+ items selected', () => {
    renderWithProviders(<BathhouseSearch />)

    const compareCheckboxes = screen.getAllByText('Сравнить')
    fireEvent.click(compareCheckboxes[0]!)
    fireEvent.click(compareCheckboxes[1]!)

    const compareBtn = screen.getByRole('button', { name: 'Сравнить' })
    expect(compareBtn).not.toBeDisabled()
  })

  it('navigates to comparison page on compare click', () => {
    renderWithProviders(<BathhouseSearch />)

    const compareCheckboxes = screen.getAllByText('Сравнить')
    fireEvent.click(compareCheckboxes[0]!)
    fireEvent.click(compareCheckboxes[1]!)

    // Find the "Сравнить" button in the comparison bar (not the checkboxes)
    const buttons = screen.getAllByRole('button', { name: 'Сравнить' })
    const compareBarBtn = buttons.find((b) => b.closest('[style*="sticky"]'))
    if (compareBarBtn) fireEvent.click(compareBarBtn)

    expect(mockNavigate).toHaveBeenCalled()
  })

  it('resets comparison on reset button click', () => {
    renderWithProviders(<BathhouseSearch />)

    const compareCheckboxes = screen.getAllByText('Сравнить')
    fireEvent.click(compareCheckboxes[0]!)

    expect(screen.getByText(/Выбрано для сравнения/)).toBeInTheDocument()

    fireEvent.click(screen.getByText('Сбросить'))

    expect(screen.queryByText(/Выбрано для сравнения/)).not.toBeInTheDocument()
  })

  it('renders search area button on map', () => {
    renderWithProviders(<BathhouseSearch />)

    fireEvent.click(screen.getByText('Карта'))

    expect(screen.getByTestId('search-area-btn')).toBeInTheDocument()
  })

  it('still renders search input and filters', () => {
    renderWithProviders(<BathhouseSearch />)

    expect(screen.getByPlaceholderText('Поиск по названию...')).toBeInTheDocument()
    expect(screen.getByText('Быстрые фильтры')).toBeInTheDocument()
    expect(screen.getByText('Найти рядом')).toBeInTheDocument()
  })

  it('enables mini-card popups on map view', () => {
    renderWithProviders(<BathhouseSearch />)

    fireEvent.click(screen.getByText('Карта'))

    expect(screen.getByTestId('mini-card-enabled')).toBeInTheDocument()
  })

  it('enables mini-card popups on split view', () => {
    renderWithProviders(<BathhouseSearch />)

    fireEvent.click(screen.getByText('Сплит'))

    expect(screen.getByTestId('mini-card-enabled')).toBeInTheDocument()
  })

  it('highlights list card when map marker is hovered', () => {
    renderWithProviders(<BathhouseSearch />)

    fireEvent.click(screen.getByText('Сплит'))

    // Hover over map marker
    const marker = screen.getByTestId('map-marker-1')
    fireEvent.mouseEnter(marker)

    const cardWrapper = screen.getByTestId('card-wrapper-1')
    expect(cardWrapper.className).toContain('rh-catalog__card-shell--highlighted')
  })

  it('removes highlight when map marker hover ends', () => {
    renderWithProviders(<BathhouseSearch />)

    fireEvent.click(screen.getByText('Сплит'))

    const marker = screen.getByTestId('map-marker-1')
    fireEvent.mouseEnter(marker)
    fireEvent.mouseLeave(marker)

    const cardWrapper = screen.getByTestId('card-wrapper-1')
    expect(cardWrapper.className).not.toContain('rh-catalog__card-shell--highlighted')
  })

  it('highlights map marker when list card is hovered', () => {
    renderWithProviders(<BathhouseSearch />)

    fireEvent.click(screen.getByText('Сплит'))

    // Hover over list card
    const cardWrapper = screen.getByTestId('card-wrapper-1')
    const col = cardWrapper.closest('.ant-col')
    if (col) fireEvent.mouseEnter(col)

    expect(screen.getByTestId('map-highlighted')).toHaveTextContent('1')
  })
})
