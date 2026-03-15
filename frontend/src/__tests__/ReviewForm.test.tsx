import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ReviewForm from '@/pages/client/ReviewForm'

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetBathhousesId: vi.fn(),
}))

vi.mock('@/api/generated/reviews/reviews', () => ({
  usePostBathhousesIdReviews: vi.fn(),
  usePutReviewsId: vi.fn(),
}))

vi.mock('@/api/generated/review-media/review-media', () => ({
  usePostReviewsIdMedia: vi.fn(),
  useDeleteMediaId: vi.fn(),
}))

import { useGetBathhousesId } from '@/api/generated/bathhouses/bathhouses'
import { usePostBathhousesIdReviews, usePutReviewsId } from '@/api/generated/reviews/reviews'
import { usePostReviewsIdMedia, useDeleteMediaId } from '@/api/generated/review-media/review-media'

function renderWithProviders(
  ui: React.ReactElement,
  { route = '/client/review?bathhouse=bath-1&booking=booking-1' } = {},
) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={[route]}>
            <Routes>
              <Route path="/client/review" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

describe('ReviewForm', () => {
  beforeEach(() => {
    vi.mocked(useGetBathhousesId).mockReturnValue({
      data: { data: { id: 'bath-1', name: 'Баня Премиум' }, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesId>)

    vi.mocked(usePostBathhousesIdReviews).mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostBathhousesIdReviews>)

    vi.mocked(usePutReviewsId).mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePutReviewsId>)

    vi.mocked(usePostReviewsIdMedia).mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostReviewsIdMedia>)

    vi.mocked(useDeleteMediaId).mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof useDeleteMediaId>)
  })

  it('renders form with bathhouse name', () => {
    renderWithProviders(<ReviewForm />)
    expect(screen.getByText(/Оставить отзыв/)).toBeInTheDocument()
    expect(screen.getByText(/Баня Премиум/)).toBeInTheDocument()
  })

  it('renders rating input', () => {
    renderWithProviders(<ReviewForm />)
    expect(screen.getByText('Оценка:')).toBeInTheDocument()
  })

  it('renders text area for review', () => {
    renderWithProviders(<ReviewForm />)
    expect(screen.getByPlaceholderText('Расскажите о вашем опыте...')).toBeInTheDocument()
  })

  it('renders media uploader section', () => {
    renderWithProviders(<ReviewForm />)
    expect(screen.getByText('Фото и видео:')).toBeInTheDocument()
    expect(screen.getByText('Загрузить')).toBeInTheDocument()
  })

  it('renders submit button', () => {
    renderWithProviders(<ReviewForm />)
    expect(screen.getByText('Отправить отзыв')).toBeInTheDocument()
  })

  it('renders back button', () => {
    renderWithProviders(<ReviewForm />)
    expect(screen.getByText('Назад к бане')).toBeInTheDocument()
  })

  it('submit button is disabled when no rating', () => {
    renderWithProviders(<ReviewForm />)
    const submitBtn = screen.getByText('Отправить отзыв').closest('button')
    expect(submitBtn).toBeDisabled()
  })

  it('shows warning when no booking id', () => {
    renderWithProviders(<ReviewForm />, { route: '/client/review?bathhouse=bath-1' })
    expect(screen.getByText(/Для написания отзыва нужно завершённое бронирование/)).toBeInTheDocument()
  })

  it('shows edit mode title when review param present', () => {
    renderWithProviders(<ReviewForm />, {
      route: '/client/review?bathhouse=bath-1&review=review-1',
    })
    expect(screen.getByText(/Редактировать отзыв/)).toBeInTheDocument()
    expect(screen.getByText('Сохранить')).toBeInTheDocument()
  })

  it('renders media counter text', () => {
    renderWithProviders(<ReviewForm />)
    expect(screen.getByText(/Фото: 0\/10/)).toBeInTheDocument()
    expect(screen.getByText(/Видео: 0\/1/)).toBeInTheDocument()
  })

  it('shows empty state when no bathhouse id', () => {
    renderWithProviders(<ReviewForm />, { route: '/client/review' })
    expect(screen.getByText('Не указана баня')).toBeInTheDocument()
  })

  it('renders rating labels when rating selected', () => {
    renderWithProviders(<ReviewForm />)
    // Find the star role=radio elements and click the 5th one
    const starRadios = screen.getAllByRole('radio')
    expect(starRadios.length).toBe(5)
    fireEvent.click(starRadios[4]!)
    expect(screen.getByText('Отлично')).toBeInTheDocument()
  })
})
