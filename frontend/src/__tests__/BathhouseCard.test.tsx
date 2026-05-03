import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import BathhouseCard from '@/components/BathhouseCard'

vi.mock('@/api/generated/favorites/favorites', () => ({
  usePostBathhousesIdFavorite: vi.fn(),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn(),
}))

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

const mockBathhouse = {
  id: 'bath-1',
  name: 'Баня на Пушкина',
  address: 'ул. Пушкина, д. 10',
  price_per_hour: 150000,
  rating: 4.5,
  review_count: 12,
  has_sauna: true,
  has_pool: true,
  has_bbq: false,
  has_steam_room: false,
  has_hot_tub: false,
  has_karaoke: false,
  is_favorite: false,
  is_photo_verified: true,
  images: ['https://example.com/photo1.jpg'],
}

describe('BathhouseCard', () => {
  beforeEach(() => {
    vi.mocked(usePostBathhousesIdFavorite).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostBathhousesIdFavorite>)

    vi.mocked(useAuthStore).mockImplementation((selector) => {
      const state = { user: { id: 'client-1', role: 'client' } }
      return (selector as (value: typeof state) => unknown)(state)
    })
  })

  it('renders bathhouse name and address', () => {
    renderWithProviders(<BathhouseCard bathhouse={mockBathhouse} />)

    expect(screen.getByText('Баня на Пушкина')).toBeInTheDocument()
    expect(screen.getByText(/ул\. Пушкина/)).toBeInTheDocument()
  })

  it('renders price per hour', () => {
    renderWithProviders(<BathhouseCard bathhouse={mockBathhouse} />)

    expect(screen.getByText('1500 ₽/ч')).toBeInTheDocument()
  })

  it('renders rating and review count', () => {
    renderWithProviders(<BathhouseCard bathhouse={mockBathhouse} />)

    expect(screen.getByText('4.5 (12)')).toBeInTheDocument()
  })

  it('renders amenity tags', () => {
    renderWithProviders(<BathhouseCard bathhouse={mockBathhouse} />)

    expect(screen.getByText('Сауна')).toBeInTheDocument()
    expect(screen.getByText('Бассейн')).toBeInTheDocument()
    expect(screen.queryByText('Мангал')).not.toBeInTheDocument()
  })

  it('shows favorite button with correct state', () => {
    renderWithProviders(<BathhouseCard bathhouse={mockBathhouse} />)

    expect(screen.getByText('В избранное')).toBeInTheDocument()
  })

  it('shows "В избранном" when bathhouse is favorite', () => {
    renderWithProviders(
      <BathhouseCard bathhouse={{ ...mockBathhouse, is_favorite: true }} />,
    )

    expect(screen.getByText('В избранном')).toBeInTheDocument()
  })

  it('hides favorite button when showFavorite is false', () => {
    renderWithProviders(
      <BathhouseCard bathhouse={mockBathhouse} showFavorite={false} />,
    )

    expect(screen.queryByText('В избранное')).not.toBeInTheDocument()
  })

  it('shows "Нет фото" when no images', () => {
    renderWithProviders(
      <BathhouseCard bathhouse={{ ...mockBathhouse, images: undefined, gallery_preview: undefined }} />,
    )

    expect(screen.getByText('Фото объекта')).toBeInTheDocument()
  })

  it('normalizes relative cover image urls', () => {
    renderWithProviders(
      <BathhouseCard bathhouse={{ ...mockBathhouse, images: ['uploads/photo1.jpg'] }} />,
    )

    expect(screen.getByAltText('Баня на Пушкина')).toHaveAttribute(
      'src',
      new URL('/uploads/photo1.jpg', window.location.origin).toString(),
    )
  })

  it('shows last minute badge when last_minute_active is true', () => {
    renderWithProviders(
      <BathhouseCard bathhouse={{ ...mockBathhouse, last_minute_active: true, last_minute_discount_percent: 20 }} />,
    )

    expect(screen.getByText('Срочная скидка -20%')).toBeInTheDocument()
  })

  it('does not show last minute badge when last_minute_active is false', () => {
    renderWithProviders(
      <BathhouseCard bathhouse={{ ...mockBathhouse, last_minute_active: false }} />,
    )

    expect(screen.queryByText(/Срочная скидка/)).not.toBeInTheDocument()
  })

  it('does not show last minute badge when last_minute_active is undefined', () => {
    renderWithProviders(
      <BathhouseCard bathhouse={mockBathhouse} />,
    )

    expect(screen.queryByText(/Срочная скидка/)).not.toBeInTheDocument()
  })
})
