import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import Favorites from '@/pages/client/Favorites'

vi.mock('@/api/generated/favorites/favorites', () => ({
  useGetMyFavorites: vi.fn(),
  usePostBathhousesIdFavorite: vi.fn(),
}))

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetBathhousesId: vi.fn(),
  getGetBathhousesIdQueryOptions: vi.fn(),
}))

import { useGetMyFavorites, usePostBathhousesIdFavorite } from '@/api/generated/favorites/favorites'
import { getGetBathhousesIdQueryOptions } from '@/api/generated/bathhouses/bathhouses'

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

const mockFavorites = [
  { id: 'fav-1', user_id: 'user-1', bathhouse_id: 'bath-1', created_at: '2025-01-01' },
  { id: 'fav-2', user_id: 'user-1', bathhouse_id: 'bath-2', created_at: '2025-01-02' },
]

describe('Favorites', () => {
  beforeEach(() => {
    vi.mocked(usePostBathhousesIdFavorite).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostBathhousesIdFavorite>)

    vi.mocked(getGetBathhousesIdQueryOptions).mockImplementation(((id: string) => ({
      queryKey: [`/bathhouses/${id}`],
      queryFn: vi.fn(),
      enabled: true,
    })) as unknown as typeof getGetBathhousesIdQueryOptions)
  })

  it('renders favorites title', () => {
    vi.mocked(useGetMyFavorites).mockReturnValue({
      data: { data: [], success: true, meta: { page: 1, page_size: 12, total_count: 0, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyFavorites>)

    renderWithProviders(<Favorites />)
    expect(screen.getByText('Избранное')).toBeInTheDocument()
  })

  it('shows empty state when no favorites', () => {
    vi.mocked(useGetMyFavorites).mockReturnValue({
      data: { data: [], success: true, meta: { page: 1, page_size: 12, total_count: 0, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyFavorites>)

    renderWithProviders(<Favorites />)
    expect(screen.getByText('У вас пока нет избранных бань')).toBeInTheDocument()
  })

  it('renders loading state', () => {
    vi.mocked(useGetMyFavorites).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyFavorites>)

    renderWithProviders(<Favorites />)
    expect(document.querySelector('.ant-spin')).toBeInTheDocument()
  })

  it('fetches bathhouse details for each favorite', () => {
    vi.mocked(useGetMyFavorites).mockReturnValue({
      data: {
        data: mockFavorites,
        success: true,
        meta: { page: 1, page_size: 12, total_count: 2, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyFavorites>)

    renderWithProviders(<Favorites />)

    expect(getGetBathhousesIdQueryOptions).toHaveBeenCalledWith('bath-1')
    expect(getGetBathhousesIdQueryOptions).toHaveBeenCalledWith('bath-2')
  })
})
