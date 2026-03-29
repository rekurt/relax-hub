import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import BathhouseDetail from '@/pages/client/BathhouseDetail'

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetBathhousesId: vi.fn(),
  useGetBathhousesBySlugSlug: vi.fn(),
  useGetBathhousesIdAvailableSlots: vi.fn(),
  useGetBathhousesIdSchema: vi.fn(),
}))

vi.mock('@/api/generated/photos/photos', () => ({
  useGetBathhousesIdPhotos: vi.fn(),
}))

vi.mock('@/api/generated/reviews/reviews', () => ({
  useGetBathhousesIdReviews: vi.fn(),
  useDeleteReviewsId: vi.fn(),
}))

vi.mock('@/api/generated/recommendations/recommendations', () => ({
  useGetBathhousesIdSimilar: vi.fn(),
}))

vi.mock('@/api/generated/review-media/review-media', () => ({
  useGetBathhousesIdGallery: vi.fn(),
}))

vi.mock('@/api/generated/favorites/favorites', () => ({
  usePostBathhousesIdFavorite: vi.fn(),
}))

vi.mock('@/api/generated/complaints/complaints', () => ({
  usePostReviewsIdReport: vi.fn(),
}))

import { useGetBathhousesBySlugSlug, useGetBathhousesIdAvailableSlots, useGetBathhousesIdSchema } from '@/api/generated/bathhouses/bathhouses'
import { useGetBathhousesIdPhotos } from '@/api/generated/photos/photos'
import { useGetBathhousesIdReviews } from '@/api/generated/reviews/reviews'
import { useDeleteReviewsId } from '@/api/generated/reviews/reviews'
import { useGetBathhousesIdSimilar } from '@/api/generated/recommendations/recommendations'
import { useGetBathhousesIdGallery } from '@/api/generated/review-media/review-media'
import { usePostBathhousesIdFavorite } from '@/api/generated/favorites/favorites'
import { usePostReviewsIdReport } from '@/api/generated/complaints/complaints'

function renderWithProviders(ui: React.ReactElement, { route = '/client/bathhouse/banya-premium' } = {}) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={[route]}>
            <Routes>
              <Route path="/client/bathhouse/:slug" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockBathhouse = {
  id: 'bath-1',
  slug: 'banya-premium',
  name: 'Баня Премиум',
  description: 'Лучшая баня в городе',
  address: 'ул. Мира, д. 10',
  price_per_hour: 300000,
  rating: 4.7,
  review_count: 15,
  min_duration: 2,
  max_guests: 10,
  has_sauna: true,
  has_pool: true,
  has_bbq: true,
  has_steam_room: false,
  has_hot_tub: false,
  has_karaoke: false,
  is_favorite: false,
  is_photo_verified: true,
  images: ['https://example.com/photo1.jpg'],
  working_hours: [
    { day_of_week: 0, open_time: '09:00', close_time: '23:00' },
    { day_of_week: 1, open_time: '09:00', close_time: '23:00' },
  ],
}

const mockSlots = [
  { startTime: '10:00', endTime: '11:00', price: 300000, available: true },
  { startTime: '11:00', endTime: '12:00', price: 300000, available: false },
  { startTime: '12:00', endTime: '13:00', price: 350000, available: true },
]

const mockReviews = [
  {
    id: 'rev-1',
    rating: 5,
    text: 'Отличная баня!',
    created_at: '2026-01-15T12:00:00Z',
    media: [],
    owner_response: 'Спасибо за отзыв!',
  },
]

describe('BathhouseDetail', () => {
  beforeEach(() => {
    vi.mocked(useGetBathhousesIdPhotos).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdPhotos>)

    vi.mocked(useGetBathhousesIdAvailableSlots).mockReturnValue({
      data: { data: mockSlots, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdAvailableSlots>)

    vi.mocked(useGetBathhousesIdReviews).mockReturnValue({
      data: { data: mockReviews, success: true, meta: { page: 1, page_size: 5, total_count: 1, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdReviews>)

    vi.mocked(useGetBathhousesIdSimilar).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdSimilar>)

    vi.mocked(useGetBathhousesIdGallery).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdGallery>)

    vi.mocked(useGetBathhousesIdSchema).mockReturnValue({
      data: { data: null, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdSchema>)

    vi.mocked(usePostBathhousesIdFavorite).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostBathhousesIdFavorite>)

    vi.mocked(useDeleteReviewsId).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof useDeleteReviewsId>)

    vi.mocked(usePostReviewsIdReport).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostReviewsIdReport>)
  })

  it('renders bathhouse name and description', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('Баня Премиум')).toBeInTheDocument()
    expect(screen.getByText('Лучшая баня в городе')).toBeInTheDocument()
  })

  it('renders price and capacity info', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    const priceElements = screen.getAllByText('3000 ₽')
    expect(priceElements.length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('2 ч')).toBeInTheDocument()
    expect(screen.getByText('10')).toBeInTheDocument()
  })

  it('renders amenity tags', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('Сауна')).toBeInTheDocument()
    expect(screen.getByText('Бассейн')).toBeInTheDocument()
    expect(screen.getByText('Мангал')).toBeInTheDocument()
  })

  it('renders available slots', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('Доступные слоты')).toBeInTheDocument()
    expect(screen.getByText('10:00 — 11:00')).toBeInTheDocument()
    expect(screen.getByText('11:00 — 12:00')).toBeInTheDocument()
    expect(screen.getByText('Занято')).toBeInTheDocument()
  })

  it('renders reviews section', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('Отзывы (15)')).toBeInTheDocument()
    expect(screen.getByText('Отличная баня!')).toBeInTheDocument()
    expect(screen.getByText('Спасибо за отзыв!')).toBeInTheDocument()
  })

  it('renders working hours', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText(/Понедельник.*09:00.*23:00/)).toBeInTheDocument()
    expect(screen.getByText(/Вторник.*09:00.*23:00/)).toBeInTheDocument()
  })

  it('shows empty state when bathhouse not found', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: null, success: false },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('Баня не найдена')).toBeInTheDocument()
  })

  it('renders favorite button', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('В избранное')).toBeInTheDocument()
  })

  it('renders back button', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('К поиску')).toBeInTheDocument()
  })
})
