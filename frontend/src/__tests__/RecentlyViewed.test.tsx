import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import RecentlyViewed from '@/components/RecentlyViewed'

vi.mock('@/api/generated/saved-searches/saved-searches', () => ({
  useGetMyRecentlyViewed: vi.fn(),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn(),
}))

import { useGetMyRecentlyViewed } from '@/api/generated/saved-searches/saved-searches'
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

const mockItems = [
  { id: '1', name: 'Баня на Неве', slug: 'banya-na-neve', viewed_at: '2026-03-29T12:00:00Z', base_price: 350000, rating: 4.5, cover_photo: '/photos/1.jpg' },
  { id: '2', name: 'Парная Люкс', slug: 'parnaya-lyuks', viewed_at: '2026-03-28T10:00:00Z', base_price: 500000, rating: 4.0 },
  { id: '3', name: 'Сауна Релакс', slug: 'sauna-relax', viewed_at: '2026-03-27T08:00:00Z', base_price: 200000, rating: 3.5 },
]

describe('RecentlyViewed', () => {
  beforeEach(() => {
    vi.mocked(useAuthStore).mockImplementation((selector) => {
      const state = { user: { id: 'u1', role: 'client' }, token: 'tok', setAuth: vi.fn(), logout: vi.fn(), isLoading: false, isAuthenticated: true, loadProfile: vi.fn() }
      return (selector as (s: typeof state) => unknown)(state)
    })
  })

  it('renders recently viewed items', () => {
    vi.mocked(useGetMyRecentlyViewed).mockReturnValue({
      data: { data: mockItems, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyRecentlyViewed>)

    renderWithProviders(<RecentlyViewed />)
    expect(screen.getByText('Недавно просмотренные')).toBeInTheDocument()
    expect(screen.getByText('Баня на Неве')).toBeInTheDocument()
    expect(screen.getByText('Парная Люкс')).toBeInTheDocument()
    expect(screen.getByText('Сауна Релакс')).toBeInTheDocument()
  })

  it('renders nothing when no items', () => {
    vi.mocked(useGetMyRecentlyViewed).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyRecentlyViewed>)

    renderWithProviders(<RecentlyViewed />)
    expect(screen.queryByText('Недавно просмотренные')).not.toBeInTheDocument()
  })

  it('renders nothing when loading', () => {
    vi.mocked(useGetMyRecentlyViewed).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyRecentlyViewed>)

    renderWithProviders(<RecentlyViewed />)
    expect(screen.queryByText('Недавно просмотренные')).not.toBeInTheDocument()
  })

  it('renders nothing when user is not authenticated', () => {
    vi.mocked(useAuthStore).mockImplementation((selector) => {
      const state = { user: null, token: null, setAuth: vi.fn(), logout: vi.fn(), isLoading: false, isAuthenticated: false, loadProfile: vi.fn() }
      return (selector as (s: typeof state) => unknown)(state)
    })

    vi.mocked(useGetMyRecentlyViewed).mockReturnValue({
      data: { data: mockItems, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyRecentlyViewed>)

    renderWithProviders(<RecentlyViewed />)
    expect(screen.queryByText('Недавно просмотренные')).not.toBeInTheDocument()
  })

  it('displays price for items', () => {
    vi.mocked(useGetMyRecentlyViewed).mockReturnValue({
      data: { data: mockItems, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyRecentlyViewed>)

    renderWithProviders(<RecentlyViewed />)
    expect(screen.getByText('от 3500 ₽')).toBeInTheDocument()
  })

  it('normalizes relative cover urls in cards', () => {
    vi.mocked(useGetMyRecentlyViewed).mockReturnValue({
      data: { data: mockItems, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyRecentlyViewed>)

    renderWithProviders(<RecentlyViewed />)

    expect(screen.getByAltText('Баня на Неве')).toHaveAttribute(
      'src',
      new URL('/photos/1.jpg', window.location.origin).toString(),
    )
  })
})
