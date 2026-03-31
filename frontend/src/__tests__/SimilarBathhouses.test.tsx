import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { App as AntApp } from 'antd'
import { describe, it, expect, vi } from 'vitest'
import SimilarBathhouses from '@/components/SimilarBathhouses'

vi.mock('@/api/generated/favorites/favorites', () => ({
  usePostBathhousesIdFavorite: vi.fn().mockReturnValue({ mutate: vi.fn(), isPending: false }),
}))

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

const mockSimilar = [
  { id: '1', slug: 'banya-1', name: 'Баня Люкс', address: 'ул. Мира 1', price_per_hour: 200000, rating: 4.5, review_count: 10 },
  { id: '2', slug: 'banya-2', name: 'Баня Премиум', address: 'ул. Мира 2', price_per_hour: 300000, rating: 4.8, review_count: 20 },
  { id: '3', slug: 'banya-3', name: 'Баня Стандарт', address: 'ул. Мира 3', price_per_hour: 150000, rating: 4.0, review_count: 5 },
]

describe('SimilarBathhouses', () => {
  it('renders nothing when items is empty', () => {
    render(
      <QueryClientProvider client={queryClient}>
        <AntApp>
          <MemoryRouter>
            <SimilarBathhouses items={[]} />
          </MemoryRouter>
        </AntApp>
      </QueryClientProvider>,
    )
    expect(screen.queryByText('Похожие бани')).not.toBeInTheDocument()
  })

  it('renders title and bathhouse cards', () => {
    render(
      <QueryClientProvider client={queryClient}>
        <AntApp>
          <MemoryRouter>
            <SimilarBathhouses items={mockSimilar} />
          </MemoryRouter>
        </AntApp>
      </QueryClientProvider>,
    )

    expect(screen.getByText('Похожие бани')).toBeInTheDocument()
    expect(screen.getByText('Баня Люкс')).toBeInTheDocument()
    expect(screen.getByText('Баня Премиум')).toBeInTheDocument()
    expect(screen.getByText('Баня Стандарт')).toBeInTheDocument()
  })

  it('limits displayed items to maxCount', () => {
    render(
      <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
        <AntApp>
          <MemoryRouter>
            <SimilarBathhouses items={mockSimilar} maxCount={2} />
          </MemoryRouter>
        </AntApp>
      </QueryClientProvider>,
    )

    expect(screen.getByText('Баня Люкс')).toBeInTheDocument()
    expect(screen.getByText('Баня Премиум')).toBeInTheDocument()
    expect(screen.queryByText('Баня Стандарт')).not.toBeInTheDocument()
  })
})
