import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ReviewList from '@/pages/reviews/ReviewList'

vi.mock('@/api/generated/reviews/reviews', () => ({
  useGetBathhousesIdReviews: vi.fn(),
  usePostReviewsIdResponse: vi.fn(),
}))

vi.mock('@/api/generated/complaints/complaints', () => ({
  usePostReviewsIdReport: vi.fn(),
  usePostBathhousesIdReport: vi.fn(),
  usePostUsersIdReport: vi.fn(),
}))

vi.mock('@/stores/bathhouse', () => ({
  useBathhouseStore: vi.fn(),
}))

import {
  useGetBathhousesIdReviews,
  usePostReviewsIdResponse,
} from '@/api/generated/reviews/reviews'
import {
  usePostReviewsIdReport,
  usePostBathhousesIdReport,
  usePostUsersIdReport,
} from '@/api/generated/complaints/complaints'
import { useBathhouseStore } from '@/stores/bathhouse'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU} theme={{ token: { motion: false } }}>
        <AntApp>
          <MemoryRouter>{ui}</MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockReviews = [
  {
    id: 'review-1',
    user_id: 'user-111',
    rating: 5,
    text: 'Отличная баня, рекомендую!',
    status: 'approved',
    owner_response: null,
    owner_response_at: null,
    media: [
      { id: 'media-1', type: 'image', url: 'https://example.com/photo1.jpg', thumbnail_url: 'https://example.com/thumb1.jpg' },
    ],
    images: [],
    created_at: '2026-03-10T14:00:00Z',
  },
  {
    id: 'review-2',
    user_id: 'user-222',
    rating: 3,
    text: 'Средне, можно лучше',
    status: 'approved',
    owner_response: 'Спасибо за обратную связь, исправим!',
    owner_response_at: '2026-03-11T10:00:00Z',
    media: [],
    images: [],
    created_at: '2026-03-09T12:00:00Z',
  },
  {
    id: 'review-3',
    user_id: 'user-333',
    rating: 1,
    text: 'Ужасно',
    status: 'pending',
    owner_response: null,
    owner_response_at: null,
    media: [],
    images: ['https://example.com/legacy-photo.jpg'],
    created_at: '2026-03-08T08:00:00Z',
  },
]

const mockResponseMutation = { mutate: vi.fn(), isPending: false }
const mockReportMutation = { mutate: vi.fn(), isPending: false }

function mockBathhouseStore(id: string | null) {
  vi.mocked(useBathhouseStore).mockImplementation((selector) =>
    (selector as (state: { selectedBathhouseId: string | null }) => unknown)({
      selectedBathhouseId: id,
    }),
  )
}

describe('ReviewList', () => {
  beforeEach(() => {
    vi.mocked(usePostReviewsIdResponse).mockReturnValue(
      mockResponseMutation as unknown as ReturnType<typeof usePostReviewsIdResponse>,
    )
    vi.mocked(usePostReviewsIdReport).mockReturnValue(
      mockReportMutation as unknown as ReturnType<typeof usePostReviewsIdReport>,
    )
    vi.mocked(usePostBathhousesIdReport).mockReturnValue(
      mockReportMutation as unknown as ReturnType<typeof usePostBathhousesIdReport>,
    )
    vi.mocked(usePostUsersIdReport).mockReturnValue(
      mockReportMutation as unknown as ReturnType<typeof usePostUsersIdReport>,
    )
  })

  it('shows prompt when no bathhouse selected', () => {
    mockBathhouseStore(null)
    vi.mocked(useGetBathhousesIdReviews).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdReviews>)

    renderWithProviders(<ReviewList />)

    expect(screen.getByText('Отзывы')).toBeInTheDocument()
    expect(screen.getByText('Выберите баню для просмотра отзывов')).toBeInTheDocument()
  })

  it('renders reviews list with data', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdReviews).mockReturnValue({
      data: {
        data: mockReviews,
        success: true,
        meta: { total_count: 3, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdReviews>)

    renderWithProviders(<ReviewList />)

    expect(screen.getByText('Отзывы')).toBeInTheDocument()
    expect(screen.getByText('Отличная баня, рекомендую!')).toBeInTheDocument()
    expect(screen.getByText('Средне, можно лучше')).toBeInTheDocument()
    expect(screen.getByText('Ужасно')).toBeInTheDocument()
  })

  it('displays owner response when present', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdReviews).mockReturnValue({
      data: {
        data: [mockReviews[1]],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdReviews>)

    renderWithProviders(<ReviewList />)

    expect(screen.getByText('Ваш ответ:')).toBeInTheDocument()
    expect(screen.getByText('Спасибо за обратную связь, исправим!')).toBeInTheDocument()
  })

  it('shows response form for reviews without response', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdReviews).mockReturnValue({
      data: {
        data: [mockReviews[0]],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdReviews>)

    renderWithProviders(<ReviewList />)

    expect(screen.getByPlaceholderText('Напишите ответ на отзыв...')).toBeInTheDocument()
    expect(screen.getByText('Отправить ответ')).toBeInTheDocument()
  })

  it('calls response mutation when form submitted', () => {
    const mutateFn = vi.fn()
    vi.mocked(usePostReviewsIdResponse).mockReturnValue({
      mutate: mutateFn,
      isPending: false,
    } as unknown as ReturnType<typeof usePostReviewsIdResponse>)
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdReviews).mockReturnValue({
      data: {
        data: [mockReviews[0]],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdReviews>)

    renderWithProviders(<ReviewList />)

    const textarea = screen.getByPlaceholderText('Напишите ответ на отзыв...')
    fireEvent.change(textarea, { target: { value: 'Благодарим за отзыв!' } })
    fireEvent.click(screen.getByText('Отправить ответ'))

    expect(mutateFn).toHaveBeenCalledWith({
      id: 'review-1',
      data: { response: 'Благодарим за отзыв!' },
    })
  })

  it('shows pending status tag', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdReviews).mockReturnValue({
      data: {
        data: [mockReviews[2]],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdReviews>)

    renderWithProviders(<ReviewList />)

    expect(screen.getByText('На модерации')).toBeInTheDocument()
  })

  it('shows rating distribution stats', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdReviews).mockReturnValue({
      data: {
        data: mockReviews,
        success: true,
        meta: { total_count: 3, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdReviews>)

    renderWithProviders(<ReviewList />)

    expect(screen.getByText(/Статистика отзывов/)).toBeInTheDocument()
    expect(screen.getByText('Средний рейтинг')).toBeInTheDocument()
    expect(screen.getByText('3 отзывов')).toBeInTheDocument()
  })

  it('shows empty state when no reviews', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdReviews).mockReturnValue({
      data: {
        data: [],
        success: true,
        meta: { total_count: 0, page: 0, page_size: 10, total_pages: 0 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdReviews>)

    renderWithProviders(<ReviewList />)

    expect(screen.getByText('Нет отзывов')).toBeInTheDocument()
  })

  it('has filter dropdown with options', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdReviews).mockReturnValue({
      data: {
        data: mockReviews,
        success: true,
        meta: { total_count: 3, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdReviews>)

    renderWithProviders(<ReviewList />)

    expect(screen.getByText('Все отзывы')).toBeInTheDocument()
  })

  it('shows media images for reviews with media', { timeout: 15000 }, async () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdReviews).mockReturnValue({
      data: {
        data: [mockReviews[0]],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdReviews>)

    renderWithProviders(<ReviewList />)

    await waitFor(() => {
      const imgs = screen.getAllByRole('img')
      const mediaImg = imgs.find((img) => img.getAttribute('src')?.includes('thumb1.jpg'))
      expect(mediaImg).toBeTruthy()
    }, { timeout: 10000 })
  })

  it('disables submit button when response is empty', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdReviews).mockReturnValue({
      data: {
        data: [mockReviews[0]],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdReviews>)

    renderWithProviders(<ReviewList />)

    const submitBtn = screen.getByText('Отправить ответ').closest('button')
    expect(submitBtn).toBeDisabled()
  })

  it('shows report button on reviews', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdReviews).mockReturnValue({
      data: {
        data: [mockReviews[0]],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdReviews>)

    renderWithProviders(<ReviewList />)
    expect(screen.getByText('Пожаловаться')).toBeInTheDocument()
  })

  it('opens report modal when report button clicked', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdReviews).mockReturnValue({
      data: {
        data: [mockReviews[0]],
        success: true,
        meta: { total_count: 1, page: 0, page_size: 10, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdReviews>)

    renderWithProviders(<ReviewList />)
    fireEvent.click(screen.getByText('Пожаловаться'))
    expect(screen.getByText('Пожаловаться на отзыв')).toBeInTheDocument()
  })
})
