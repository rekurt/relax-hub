import { fireEvent, render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
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
  images: ['/uploads/photo1.jpg'],
  working_hours: [
    { day_of_week: 0, open_time: '09:00', close_time: '23:00' },
    { day_of_week: 1, open_time: '09:00', close_time: '23:00' },
  ],
  booking_mode: 'instant',
  cancellation_policy: 'flexible',
  visiting_rules: 'Сменная обувь обязательна. Дети до 12 лет — бесплатно.',
  area_avg_price_per_hour: 250000,
  owner_profile: {
    name: 'Иван Петров',
    avatar_url: null,
    rating: 4.9,
    object_count: 3,
    member_since: '2024-06-01T00:00:00Z',
  },
}

const mockSlots = [
  { startTime: '2026-04-20T10:00:00', endTime: '2026-04-20T11:00:00', price: 300000, available: true },
  { startTime: '2026-04-20T11:00:00', endTime: '2026-04-20T12:00:00', price: 300000, available: true },
  { startTime: '2026-04-20T12:00:00', endTime: '2026-04-20T13:00:00', price: 350000, available: false },
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
    expect(screen.getAllByText('Лучшая баня в городе').length).toBeGreaterThan(0)
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

  it('renders public trust hero with booking facts and slot cta', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('Цена от')).toBeInTheDocument()
    expect(screen.getByText('Минимум')).toBeInTheDocument()
    expect(screen.getByText('Подтверждение')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Выбрать слот' })).toBeInTheDocument()
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

    expect(screen.getByText('Свободные слоты')).toBeInTheDocument()
    expect(document.body.textContent).toContain('10:00')
    expect(document.body.textContent).toContain('11:00')
  })

  it('selects a continuous range in one block without duration buttons', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.queryByText(/Длительность/)).not.toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /10:00/ }))
    fireEvent.click(screen.getByRole('button', { name: /11:00/ }))

    expect(screen.getByText(/10:00 - 12:00 · 2 ч/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Забронировать' })).toBeEnabled()
  })

  it('normalizes relative bathhouse photo urls', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    const photo = screen.getByAltText('Баня Премиум фото 1')
    expect(photo).toHaveAttribute('src', new URL('/uploads/photo1.jpg', window.location.origin).toString())
  })

  it('shows slot recovery state when availability request fails', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    vi.mocked(useGetBathhousesIdAvailableSlots).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetBathhousesIdAvailableSlots>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('Не удалось загрузить доступные слоты')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Обновить слоты' })).toBeInTheDocument()
  })

  it('renders reviews section', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByRole('heading', { name: /Отзывы/i })).toBeInTheDocument()
    expect(screen.getByText('Отличная баня!')).toBeInTheDocument()
    expect(screen.getByText('Спасибо за отзыв!')).toBeInTheDocument()
  })

  it('shows only three latest reviews until all reviews are opened', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    vi.mocked(useGetBathhousesIdReviews).mockReturnValue({
      data: {
        data: [
          { id: 'rev-4', rating: 5, text: 'Самый свежий отзыв', created_at: '2026-04-10T12:00:00Z', media: [] },
          { id: 'rev-3', rating: 5, text: 'Второй свежий отзыв', created_at: '2026-04-09T12:00:00Z', media: [] },
          { id: 'rev-2', rating: 4, text: 'Третий свежий отзыв', created_at: '2026-04-08T12:00:00Z', media: [] },
          { id: 'rev-1', rating: 4, text: 'Старый отзыв', created_at: '2026-04-07T12:00:00Z', media: [] },
        ],
        success: true,
        meta: { page: 1, page_size: 20, total_count: 4, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdReviews>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('Самый свежий отзыв')).toBeInTheDocument()
    expect(screen.getByText('Второй свежий отзыв')).toBeInTheDocument()
    expect(screen.getByText('Третий свежий отзыв')).toBeInTheDocument()
    expect(screen.queryByText('Старый отзыв')).not.toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Показать все отзывы' }))

    expect(screen.getByText('Старый отзыв')).toBeInTheDocument()
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

  it('does not render favorite button for public visitor', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.queryByText('В избранное')).not.toBeInTheDocument()
  })

  it('renders back button', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('К поиску')).toBeInTheDocument()
  })

  it('renders booking mode indicator', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('Мгновенное')).toBeInTheDocument()
  })

  it('renders request booking mode', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: { ...mockBathhouse, booking_mode: 'request' }, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('По запросу')).toBeInTheDocument()
  })

  it('renders cancellation policy with details', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('Гибкая')).toBeInTheDocument()
    expect(screen.getAllByText(/Бесплатная отмена за 24/).length).toBeGreaterThan(0)
  })

  it('renders visiting rules', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('Правила посещения:')).toBeInTheDocument()
    expect(screen.getByText(/Сменная обувь обязательна/)).toBeInTheDocument()
  })

  it('renders owner profile block', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('Иван Петров')).toBeInTheDocument()
    expect(screen.getByText('Объектов: 3')).toBeInTheDocument()
    expect(screen.getByText(/Рейтинг: 4.9/)).toBeInTheDocument()
  })

  it('renders price breakdown card', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    renderWithProviders(<BathhouseDetail />)

    expect(screen.getByText('Примерная стоимость')).toBeInTheDocument()
    expect(screen.getByText('Базовая стоимость')).toBeInTheDocument()
    expect(screen.getByText('Сервисный сбор')).toBeInTheDocument()
  })

  it('works with public /bathhouses/:slug route', () => {
    vi.mocked(useGetBathhousesBySlugSlug).mockReturnValue({
      data: { data: mockBathhouse, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesBySlugSlug>)

    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    render(
      <QueryClientProvider client={queryClient}>
        <ConfigProvider locale={ruRU}>
          <AntApp>
            <MemoryRouter initialEntries={['/bathhouses/banya-premium']}>
              <Routes>
                <Route path="/bathhouses/:slug" element={<BathhouseDetail />} />
              </Routes>
            </MemoryRouter>
          </AntApp>
        </ConfigProvider>
      </QueryClientProvider>,
    )

    expect(screen.getByText('Баня Премиум')).toBeInTheDocument()
    expect(useGetBathhousesBySlugSlug).toHaveBeenCalledWith('banya-premium', expect.anything())
  })
})
