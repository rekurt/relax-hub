import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ComparisonPage from '@/pages/client/ComparisonPage'

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useSearchParams: vi.fn(),
  }
})

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  usePostApiV1BathhousesCompare: vi.fn(),
  useGetBathhouses: vi.fn().mockReturnValue({ data: null, isLoading: false }),
}))

vi.mock('@/api/generated/favorites/favorites', () => ({
  usePostBathhousesIdFavorite: vi.fn().mockReturnValue({ mutate: vi.fn(), isPending: false }),
}))

import { useSearchParams } from 'react-router-dom'
import { usePostApiV1BathhousesCompare } from '@/api/generated/bathhouses/bathhouses'

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

const mockItems = [
  {
    id: 'b1',
    name: 'Баня Классик',
    address: 'ул. Ленина, 1',
    price_per_hour: 200000,
    rating: 4.2,
    review_count: 10,
    max_guests: 8,
    min_duration: 2,
    has_sauna: true,
    has_pool: false,
    has_steam_room: true,
    has_hot_tub: false,
    has_bbq: false,
    has_karaoke: false,
    images: ['https://example.com/p1.jpg'],
    distance: 2.5,
  },
  {
    id: 'b2',
    name: 'Баня Люкс',
    address: 'ул. Мира, 5',
    price_per_hour: 500000,
    rating: 4.8,
    review_count: 25,
    max_guests: 12,
    min_duration: 1,
    has_sauna: true,
    has_pool: true,
    has_steam_room: true,
    has_hot_tub: true,
    has_bbq: true,
    has_karaoke: true,
    images: ['https://example.com/p2.jpg'],
    distance: 5.1,
  },
]

describe('ComparisonPage', () => {
  const mockSetSearchParams = vi.fn()
  let mockMutate: ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.clearAllMocks()
    mockMutate = vi.fn()

    vi.mocked(usePostApiV1BathhousesCompare).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
    } as unknown as ReturnType<typeof usePostApiV1BathhousesCompare>)
  })

  it('shows empty state when less than 2 IDs', () => {
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams('ids=b1'),
      mockSetSearchParams,
    ] as unknown as ReturnType<typeof useSearchParams>)

    renderWithProviders(<ComparisonPage />)

    expect(screen.getByText('Выберите минимум 2 бани для сравнения')).toBeInTheDocument()
  })

  it('shows go to search button on empty state', () => {
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams(''),
      mockSetSearchParams,
    ] as unknown as ReturnType<typeof useSearchParams>)

    renderWithProviders(<ComparisonPage />)

    expect(screen.getByText('Перейти к поиску')).toBeInTheDocument()
  })

  it('calls compare API when 2 IDs provided', () => {
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams('ids=b1,b2'),
      mockSetSearchParams,
    ] as unknown as ReturnType<typeof useSearchParams>)

    renderWithProviders(<ComparisonPage />)

    expect(mockMutate).toHaveBeenCalledWith({
      data: { ids: ['b1', 'b2'] },
    })
  })

  it('renders page title', () => {
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams('ids=b1,b2'),
      mockSetSearchParams,
    ] as unknown as ReturnType<typeof useSearchParams>)

    renderWithProviders(<ComparisonPage />)

    expect(screen.getByText('Сравнение бань')).toBeInTheDocument()
  })

  it('renders back to search button', () => {
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams('ids=b1,b2'),
      mockSetSearchParams,
    ] as unknown as ReturnType<typeof useSearchParams>)

    renderWithProviders(<ComparisonPage />)

    const backBtn = screen.getByText('Назад к поиску')
    expect(backBtn).toBeInTheDocument()
    fireEvent.click(backBtn)
    expect(mockNavigate).toHaveBeenCalledWith('/client/search')
  })

  it('renders comparison table with data', () => {
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams('ids=b1,b2'),
      mockSetSearchParams,
    ] as unknown as ReturnType<typeof useSearchParams>)

    // Simulate onSuccess callback
    vi.mocked(usePostApiV1BathhousesCompare).mockImplementation(((opts: { mutation?: { onSuccess?: (data: unknown) => void } }) => {
      const mutate = vi.fn().mockImplementation(() => {
        opts?.mutation?.onSuccess?.({
          data: { items: mockItems },
        })
      })
      return { mutate, isPending: false }
    }) as unknown as typeof usePostApiV1BathhousesCompare)

    renderWithProviders(<ComparisonPage />)

    expect(screen.getAllByText('Параметр').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Цена за час').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Рейтинг').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Адрес').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Макс. гостей').length).toBeGreaterThanOrEqual(1)
  })

  it('renders bathhouse names as links in table headers', () => {
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams('ids=b1,b2'),
      mockSetSearchParams,
    ] as unknown as ReturnType<typeof useSearchParams>)

    vi.mocked(usePostApiV1BathhousesCompare).mockImplementation(((opts: { mutation?: { onSuccess?: (data: unknown) => void } }) => {
      const mutate = vi.fn().mockImplementation(() => {
        opts?.mutation?.onSuccess?.({
          data: { items: mockItems },
        })
      })
      return { mutate, isPending: false }
    }) as unknown as typeof usePostApiV1BathhousesCompare)

    renderWithProviders(<ComparisonPage />)

    const klassikLinks = screen.getAllByText('Баня Классик')
    expect(klassikLinks.length).toBeGreaterThanOrEqual(1)
    const luxLinks = screen.getAllByText('Баня Люкс')
    expect(luxLinks.length).toBeGreaterThanOrEqual(1)
  })

  it('renders remove buttons for each bathhouse', () => {
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams('ids=b1,b2'),
      mockSetSearchParams,
    ] as unknown as ReturnType<typeof useSearchParams>)

    vi.mocked(usePostApiV1BathhousesCompare).mockImplementation(((opts: { mutation?: { onSuccess?: (data: unknown) => void } }) => {
      const mutate = vi.fn().mockImplementation(() => {
        opts?.mutation?.onSuccess?.({
          data: { items: mockItems },
        })
      })
      return { mutate, isPending: false }
    }) as unknown as typeof usePostApiV1BathhousesCompare)

    renderWithProviders(<ComparisonPage />)

    const removeButtons = screen.getAllByText('Убрать')
    expect(removeButtons.length).toBeGreaterThanOrEqual(2)
  })
})
