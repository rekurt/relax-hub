import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ReviewModeration from '@/pages/admin/ReviewModeration'

vi.mock('@/api/generated/admin-reviews/admin-reviews', () => ({
  useGetAdminReviews: vi.fn(),
  getGetAdminReviewsQueryKey: vi.fn(() => ['/admin/reviews']),
  useGetAdminReviewsPendingCount: vi.fn(),
  getGetAdminReviewsPendingCountQueryKey: vi.fn(() => ['/admin/reviews/pending-count']),
  usePatchAdminReviewsIdApprove: vi.fn(),
  usePatchAdminReviewsIdReject: vi.fn(),
  usePostAdminReviewsBatchApprove: vi.fn(),
  usePostAdminReviewsBatchReject: vi.fn(),
}))

import {
  useGetAdminReviews,
  useGetAdminReviewsPendingCount,
  usePatchAdminReviewsIdApprove,
  usePatchAdminReviewsIdReject,
  usePostAdminReviewsBatchApprove,
  usePostAdminReviewsBatchReject,
} from '@/api/generated/admin-reviews/admin-reviews'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/reviews']}>
            <Routes>
              <Route path="/admin/reviews" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockReviews = [
  {
    id: 'r-1',
    rating: 5,
    text: 'Отличная баня, очень понравилось!',
    status: 'pending',
    user_id: 'u-1',
    bathhouse_id: 'b-1',
    images: ['https://example.com/photo1.jpg', 'https://example.com/photo2.jpg'],
    created_at: '2026-03-15T10:00:00Z',
  },
  {
    id: 'r-2',
    rating: 2,
    text: 'Не рекомендую',
    status: 'approved',
    user_id: 'u-2',
    bathhouse_id: 'b-2',
    images: [],
    created_at: '2026-03-14T14:00:00Z',
  },
  {
    id: 'r-3',
    rating: 4,
    text: 'Хорошее место',
    status: 'rejected',
    user_id: 'u-3',
    bathhouse_id: 'b-1',
    images: ['https://example.com/photo3.jpg'],
    rejection_reasons: ['spam'],
    created_at: '2026-03-13T09:00:00Z',
  },
]

const mutationDefault = { mutateAsync: vi.fn(), isPending: false }

beforeEach(() => {
  vi.mocked(useGetAdminReviews).mockReturnValue({
    data: {
      data: mockReviews,
      success: true,
      meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 },
    },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetAdminReviews>)

  vi.mocked(useGetAdminReviewsPendingCount).mockReturnValue({
    data: { data: { pending_count: 5 }, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetAdminReviewsPendingCount>)

  vi.mocked(usePatchAdminReviewsIdApprove).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePatchAdminReviewsIdApprove>,
  )
  vi.mocked(usePatchAdminReviewsIdReject).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePatchAdminReviewsIdReject>,
  )
  vi.mocked(usePostAdminReviewsBatchApprove).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePostAdminReviewsBatchApprove>,
  )
  vi.mocked(usePostAdminReviewsBatchReject).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePostAdminReviewsBatchReject>,
  )
})

describe('ReviewModeration', () => {
  it('renders page title and review list', () => {
    renderWithProviders(<ReviewModeration />)

    expect(screen.getByText('Модерация отзывов')).toBeInTheDocument()
    expect(screen.getByText('Отличная баня, очень понравилось!')).toBeInTheDocument()
    expect(screen.getByText('Не рекомендую')).toBeInTheDocument()
    expect(screen.getByText('Хорошее место')).toBeInTheDocument()
  })

  it('renders pending count badge', () => {
    renderWithProviders(<ReviewModeration />)

    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('renders status filter segmented control', () => {
    renderWithProviders(<ReviewModeration />)

    expect(screen.getByText('Все')).toBeInTheDocument()
    expect(screen.getByText('Одобренные')).toBeInTheDocument()
    expect(screen.getByText('Отклонённые')).toBeInTheDocument()
    expect(screen.getByText('Скрытые')).toBeInTheDocument()
  })

  it('renders status tags for reviews', () => {
    renderWithProviders(<ReviewModeration />)

    // "На рассмотрении" appears in segmented control and table tag
    expect(screen.getAllByText('На рассмотрении').length).toBeGreaterThanOrEqual(2)
    expect(screen.getByText('Одобрен')).toBeInTheDocument()
    expect(screen.getByText('Отклонён')).toBeInTheDocument()
  })

  it('renders media count for reviews with images', () => {
    renderWithProviders(<ReviewModeration />)

    // Verify media info through the detail drawer - click review with images
    fireEvent.click(screen.getByText('Отличная баня, очень понравилось!'))
    expect(screen.getByText(/Медиа \(2\)/)).toBeInTheDocument()
  })

  it('shows approve action for non-approved reviews', () => {
    renderWithProviders(<ReviewModeration />)

    const approveLinks = screen.getAllByText('Одобрить')
    // r-1 (pending) and r-3 (rejected) have approve action; status filter label also has it
    expect(approveLinks.length).toBeGreaterThanOrEqual(2)
  })

  it('shows reject action for non-rejected reviews', () => {
    renderWithProviders(<ReviewModeration />)

    const rejectLinks = screen.getAllByText('Отклонить')
    // r-1 (pending) and r-2 (approved) have reject action
    expect(rejectLinks.length).toBeGreaterThanOrEqual(2)
  })

  it('opens detail drawer when clicking review text', async () => {
    renderWithProviders(<ReviewModeration />)

    fireEvent.click(screen.getByText('Отличная баня, очень понравилось!'))

    await waitFor(() => {
      expect(screen.getByText('Детали отзыва')).toBeInTheDocument()
      expect(screen.getByText('Пользователь')).toBeInTheDocument()
      expect(screen.getByText('Баня')).toBeInTheDocument()
    })
  })

  it('shows batch action buttons when items selected', () => {
    renderWithProviders(<ReviewModeration />)

    // Select all checkbox (header)
    const checkboxes = screen.getAllByRole('checkbox')
    fireEvent.click(checkboxes[0]!) // header "select all"

    expect(screen.getByText(/Выбрано: 3/)).toBeInTheDocument()
    expect(screen.getByText('Одобрить выбранные')).toBeInTheDocument()
    expect(screen.getByText('Отклонить выбранные')).toBeInTheDocument()
  })

  it('renders total count in pagination', () => {
    renderWithProviders(<ReviewModeration />)

    expect(screen.getByText('Всего: 3')).toBeInTheDocument()
  })

  it('renders empty state when no reviews', () => {
    vi.mocked(useGetAdminReviews).mockReturnValue({
      data: {
        data: [],
        success: true,
        meta: { page: 1, page_size: 20, total_count: 0, total_pages: 0 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminReviews>)

    renderWithProviders(<ReviewModeration />)

    expect(screen.getByText('Нет отзывов')).toBeInTheDocument()
  })
})
