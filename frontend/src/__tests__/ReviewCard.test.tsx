import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ReviewCard from '@/components/ReviewCard'

vi.mock('@/api/generated/complaints/complaints', () => ({
  usePostReviewsIdReport: vi.fn(),
}))

vi.mock('@/api/generated/reviews/reviews', () => ({
  useDeleteReviewsId: vi.fn(),
}))

import { usePostReviewsIdReport } from '@/api/generated/complaints/complaints'
import { useDeleteReviewsId } from '@/api/generated/reviews/reviews'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>{ui}</AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockReview = {
  id: 'review-1',
  rating: 4,
  text: 'Отличная баня, рекомендую!',
  created_at: '2026-03-10T12:00:00Z',
  status: 'approved',
  user_id: 'user-1',
  media: [
    { id: 'media-1', type: 'image' as const, url: 'http://example.com/photo.jpg', thumbnail_url: 'http://example.com/thumb.jpg' },
  ],
}

describe('ReviewCard', () => {
  beforeEach(() => {
    vi.mocked(usePostReviewsIdReport).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostReviewsIdReport>)

    vi.mocked(useDeleteReviewsId).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof useDeleteReviewsId>)
  })

  it('renders review text', () => {
    renderWithProviders(<ReviewCard review={mockReview} />)
    expect(screen.getByText('Отличная баня, рекомендую!')).toBeInTheDocument()
  })

  it('renders review date', () => {
    renderWithProviders(<ReviewCard review={mockReview} />)
    expect(screen.getByText('10.03.2026')).toBeInTheDocument()
  })

  it('renders rating stars', () => {
    renderWithProviders(<ReviewCard review={mockReview} />)
    const stars = document.querySelectorAll('.ant-rate-star-full')
    expect(stars.length).toBe(4)
  })

  it('renders media thumbnail', () => {
    renderWithProviders(<ReviewCard review={mockReview} />)
    const img = document.querySelector('img[src="http://example.com/thumb.jpg"]')
    expect(img).toBeInTheDocument()
  })

  it('renders report button for non-author', () => {
    renderWithProviders(<ReviewCard review={mockReview} isAuthor={false} />)
    expect(screen.getByText('Жалоба')).toBeInTheDocument()
  })

  it('renders edit/delete buttons for author', () => {
    const onEdit = vi.fn()
    renderWithProviders(<ReviewCard review={mockReview} isAuthor={true} onEdit={onEdit} />)
    // Edit and delete icons should be present
    const deleteBtn = document.querySelector('[aria-label="delete"]')
    expect(deleteBtn).toBeInTheDocument()
  })

  it('does not show report button for author', () => {
    renderWithProviders(<ReviewCard review={mockReview} isAuthor={true} />)
    expect(screen.queryByText('Жалоба')).not.toBeInTheDocument()
  })

  it('opens report modal on report click', () => {
    renderWithProviders(<ReviewCard review={mockReview} isAuthor={false} />)
    fireEvent.click(screen.getByText('Жалоба'))
    expect(screen.getByText('Пожаловаться на отзыв')).toBeInTheDocument()
    expect(screen.getByText('Причина:')).toBeInTheDocument()
  })

  it('renders owner response when present', () => {
    const reviewWithResponse = {
      ...mockReview,
      owner_response: 'Спасибо за отзыв!',
      owner_response_at: '2026-03-11T10:00:00Z',
    }
    renderWithProviders(<ReviewCard review={reviewWithResponse} />)
    expect(screen.getByText('Ответ владельца:')).toBeInTheDocument()
    expect(screen.getByText('Спасибо за отзыв!')).toBeInTheDocument()
  })

  it('shows pending status tag', () => {
    const pendingReview = { ...mockReview, status: 'pending' }
    renderWithProviders(<ReviewCard review={pendingReview} />)
    expect(screen.getByText('На модерации')).toBeInTheDocument()
  })

  it('shows rejected status tag', () => {
    const rejectedReview = { ...mockReview, status: 'rejected' }
    renderWithProviders(<ReviewCard review={rejectedReview} />)
    expect(screen.getByText('Отклонён')).toBeInTheDocument()
  })

  it('shows "Без текста" for empty review text', () => {
    const noTextReview = { ...mockReview, text: '' }
    renderWithProviders(<ReviewCard review={noTextReview} />)
    expect(screen.getByText('Без текста')).toBeInTheDocument()
  })

  it('hides actions when showActions is false', () => {
    renderWithProviders(<ReviewCard review={mockReview} showActions={false} />)
    expect(screen.queryByText('Жалоба')).not.toBeInTheDocument()
  })
})
