import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import SavedSearches from '@/pages/client/SavedSearches'

vi.mock('@/api/generated/saved-searches/saved-searches', () => ({
  useGetMySavedSearches: vi.fn(),
  useDeleteMySavedSearchesId: vi.fn(),
}))

import {
  useGetMySavedSearches,
  useDeleteMySavedSearchesId,
} from '@/api/generated/saved-searches/saved-searches'

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

const mockSearches = [
  {
    id: 's1',
    name: 'Бани рядом с домом',
    filters: { q: 'сауна', has_sauna: true, price_min: 100000 },
    notify_on_new: true,
    created_at: '2026-03-01T10:00:00Z',
  },
  {
    id: 's2',
    name: 'Дешёвые бани',
    filters: { price_max: 200000 },
    notify_on_new: false,
    created_at: '2026-03-15T14:30:00Z',
  },
]

describe('SavedSearches', () => {
  beforeEach(() => {
    vi.clearAllMocks()

    vi.mocked(useDeleteMySavedSearchesId).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof useDeleteMySavedSearchesId>)
  })

  it('renders page title', () => {
    vi.mocked(useGetMySavedSearches).mockReturnValue({
      data: { data: mockSearches, success: true, meta: { page: 1, page_size: 20, total_count: 2, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySavedSearches>)

    renderWithProviders(<SavedSearches />)

    expect(screen.getByText('Сохранённые поиски')).toBeInTheDocument()
  })

  it('renders saved search names', () => {
    vi.mocked(useGetMySavedSearches).mockReturnValue({
      data: { data: mockSearches, success: true, meta: { page: 1, page_size: 20, total_count: 2, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySavedSearches>)

    renderWithProviders(<SavedSearches />)

    expect(screen.getByText('Бани рядом с домом')).toBeInTheDocument()
    expect(screen.getByText('Дешёвые бани')).toBeInTheDocument()
  })

  it('shows notification status tags', () => {
    vi.mocked(useGetMySavedSearches).mockReturnValue({
      data: { data: mockSearches, success: true, meta: { page: 1, page_size: 20, total_count: 2, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySavedSearches>)

    renderWithProviders(<SavedSearches />)

    expect(screen.getByText('Включены')).toBeInTheDocument()
    expect(screen.getByText('Выключены')).toBeInTheDocument()
  })

  it('shows empty state when no searches', () => {
    vi.mocked(useGetMySavedSearches).mockReturnValue({
      data: { data: [], success: true, meta: { page: 1, page_size: 20, total_count: 0, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySavedSearches>)

    renderWithProviders(<SavedSearches />)

    expect(screen.getByText('У вас нет сохранённых поисков')).toBeInTheDocument()
  })

  it('shows delete confirmation on button click', () => {
    vi.mocked(useGetMySavedSearches).mockReturnValue({
      data: { data: mockSearches, success: true, meta: { page: 1, page_size: 20, total_count: 2, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySavedSearches>)

    renderWithProviders(<SavedSearches />)

    const deleteButtons = document.querySelectorAll('.ant-btn-dangerous')
    expect(deleteButtons.length).toBeGreaterThanOrEqual(1)
    fireEvent.click(deleteButtons[0])

    expect(screen.getByText('Удалить сохранённый поиск?')).toBeInTheDocument()
  })

  it('renders table columns', () => {
    vi.mocked(useGetMySavedSearches).mockReturnValue({
      data: { data: mockSearches, success: true, meta: { page: 1, page_size: 20, total_count: 2, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMySavedSearches>)

    renderWithProviders(<SavedSearches />)

    expect(screen.getAllByText('Название').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Уведомления').length).toBeGreaterThanOrEqual(1)
  })
})
