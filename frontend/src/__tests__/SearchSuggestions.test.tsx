import { render, screen, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import SearchSuggestions from '@/components/SearchSuggestions'

vi.mock('@/api/generated/search/search', () => ({
  useGetSearchSuggestions: vi.fn(),
}))

import { useGetSearchSuggestions } from '@/api/generated/search/search'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      {ui}
    </QueryClientProvider>,
  )
}

const mockSuggestions = [
  { text: 'Баня Классик', type: 'bathhouse' },
  { text: 'Баня Люкс', type: 'bathhouse' },
  { text: 'Москва', type: 'city' },
  { text: 'баня с бассейном', type: 'popular' },
]

describe('SearchSuggestions', () => {
  const onSelect = vi.fn()
  const onClose = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('does not render when query is too short', () => {
    vi.mocked(useGetSearchSuggestions).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetSearchSuggestions>)

    renderWithProviders(
      <SearchSuggestions query="б" visible={true} onSelect={onSelect} onClose={onClose} />,
    )

    expect(screen.queryByTestId('search-suggestions')).not.toBeInTheDocument()
  })

  it('does not render when not visible', () => {
    vi.mocked(useGetSearchSuggestions).mockReturnValue({
      data: { data: mockSuggestions },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetSearchSuggestions>)

    renderWithProviders(
      <SearchSuggestions query="баня" visible={false} onSelect={onSelect} onClose={onClose} />,
    )

    expect(screen.queryByTestId('search-suggestions')).not.toBeInTheDocument()
  })

  it('shows grouped suggestions after debounce', () => {
    vi.mocked(useGetSearchSuggestions).mockReturnValue({
      data: { data: mockSuggestions },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetSearchSuggestions>)

    renderWithProviders(
      <SearchSuggestions query="баня" visible={true} onSelect={onSelect} onClose={onClose} />,
    )

    act(() => { vi.advanceTimersByTime(300) })

    expect(screen.getByTestId('search-suggestions')).toBeInTheDocument()
    expect(screen.getByText('Бани')).toBeInTheDocument()
    expect(screen.getByText('Города')).toBeInTheDocument()
    expect(screen.getByText('Популярное')).toBeInTheDocument()
  })

  it('shows suggestion texts', () => {
    vi.mocked(useGetSearchSuggestions).mockReturnValue({
      data: { data: mockSuggestions },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetSearchSuggestions>)

    renderWithProviders(
      <SearchSuggestions query="баня" visible={true} onSelect={onSelect} onClose={onClose} />,
    )

    act(() => { vi.advanceTimersByTime(300) })

    expect(screen.getByText('Баня Классик')).toBeInTheDocument()
    expect(screen.getByText('Баня Люкс')).toBeInTheDocument()
    expect(screen.getByText('Москва')).toBeInTheDocument()
    expect(screen.getByText('баня с бассейном')).toBeInTheDocument()
  })

  it('shows loading state', () => {
    vi.mocked(useGetSearchSuggestions).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetSearchSuggestions>)

    renderWithProviders(
      <SearchSuggestions query="баня" visible={true} onSelect={onSelect} onClose={onClose} />,
    )

    act(() => { vi.advanceTimersByTime(300) })

    expect(screen.getByTestId('search-suggestions')).toBeInTheDocument()
    expect(document.querySelector('.ant-spin')).toBeInTheDocument()
  })

  it('shows empty state when no results', () => {
    vi.mocked(useGetSearchSuggestions).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetSearchSuggestions>)

    renderWithProviders(
      <SearchSuggestions query="несуществующее" visible={true} onSelect={onSelect} onClose={onClose} />,
    )

    act(() => { vi.advanceTimersByTime(300) })

    expect(screen.getByText('Ничего не найдено')).toBeInTheDocument()
  })

  it('calls onSelect when suggestion clicked', () => {
    vi.mocked(useGetSearchSuggestions).mockReturnValue({
      data: { data: mockSuggestions },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetSearchSuggestions>)

    renderWithProviders(
      <SearchSuggestions query="баня" visible={true} onSelect={onSelect} onClose={onClose} />,
    )

    act(() => { vi.advanceTimersByTime(300) })

    const option = screen.getByText('Баня Классик')
    option.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }))

    expect(onSelect).toHaveBeenCalledWith('Баня Классик')
  })
})
