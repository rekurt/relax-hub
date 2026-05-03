import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ReportModal from '@/components/ReportModal'

vi.mock('@/api/generated/complaints/complaints', () => ({
  usePostReviewsIdReport: vi.fn(),
  usePostBathhousesIdReport: vi.fn(),
  usePostUsersIdReport: vi.fn(),
}))

import {
  usePostReviewsIdReport,
  usePostBathhousesIdReport,
  usePostUsersIdReport,
} from '@/api/generated/complaints/complaints'

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

const mockMutation = { mutate: vi.fn(), isPending: false }

describe('ReportModal', () => {
  beforeEach(() => {
    vi.mocked(usePostReviewsIdReport).mockReturnValue(mockMutation as unknown as ReturnType<typeof usePostReviewsIdReport>)
    vi.mocked(usePostBathhousesIdReport).mockReturnValue(mockMutation as unknown as ReturnType<typeof usePostBathhousesIdReport>)
    vi.mocked(usePostUsersIdReport).mockReturnValue(mockMutation as unknown as ReturnType<typeof usePostUsersIdReport>)
    mockMutation.mutate.mockClear()
  })

  it('renders with correct title for review target', () => {
    renderWithProviders(
      <ReportModal open targetType="review" targetId="review-1" onClose={vi.fn()} />,
    )
    expect(screen.getByText('Пожаловаться на отзыв')).toBeInTheDocument()
  })

  it('renders with correct title for bathhouse target', () => {
    renderWithProviders(
      <ReportModal open targetType="bathhouse" targetId="bh-1" onClose={vi.fn()} />,
    )
    expect(screen.getByText('Пожаловаться на баню')).toBeInTheDocument()
  })

  it('renders with correct title for user target', () => {
    renderWithProviders(
      <ReportModal open targetType="user" targetId="user-1" onClose={vi.fn()} />,
    )
    expect(screen.getByText('Пожаловаться на пользователя')).toBeInTheDocument()
  })

  it('renders all reason options', () => {
    renderWithProviders(
      <ReportModal open targetType="review" targetId="review-1" onClose={vi.fn()} />,
    )
    expect(screen.getByText('Спам')).toBeInTheDocument()
    expect(screen.getByText('Оскорбительное содержание')).toBeInTheDocument()
    expect(screen.getByText('Фейковый контент')).toBeInTheDocument()
    expect(screen.getByText('Мошенничество')).toBeInTheDocument()
    expect(screen.getByText('Другое')).toBeInTheDocument()
  })

  it('disables submit when no reason selected', () => {
    renderWithProviders(
      <ReportModal open targetType="review" targetId="review-1" onClose={vi.fn()} />,
    )
    const submitBtn = screen.getByText('Отправить').closest('button')
    expect(submitBtn).toBeDisabled()
  })

  it('enables submit when reason is selected', () => {
    renderWithProviders(
      <ReportModal open targetType="review" targetId="review-1" onClose={vi.fn()} />,
    )
    fireEvent.click(screen.getByText('Спам'))
    const submitBtn = screen.getByText('Отправить').closest('button')
    expect(submitBtn).not.toBeDisabled()
  })

  it('calls review report mutation with correct data', () => {
    renderWithProviders(
      <ReportModal open targetType="review" targetId="review-1" onClose={vi.fn()} />,
    )
    fireEvent.click(screen.getByText('Мошенничество'))
    fireEvent.click(screen.getByText('Отправить'))

    expect(mockMutation.mutate).toHaveBeenCalledWith({
      id: 'review-1',
      data: { reason: 'fraud', description: undefined },
    })
  })

  it('calls bathhouse report mutation for bathhouse target', () => {
    renderWithProviders(
      <ReportModal open targetType="bathhouse" targetId="bh-1" onClose={vi.fn()} />,
    )
    fireEvent.click(screen.getByText('Спам'))
    fireEvent.click(screen.getByText('Отправить'))

    expect(mockMutation.mutate).toHaveBeenCalledWith({
      id: 'bh-1',
      data: { reason: 'spam', description: undefined },
    })
  })

  it('calls user report mutation for user target', () => {
    renderWithProviders(
      <ReportModal open targetType="user" targetId="user-1" onClose={vi.fn()} />,
    )
    fireEvent.click(screen.getByText('Другое'))
    fireEvent.click(screen.getByText('Отправить'))

    expect(mockMutation.mutate).toHaveBeenCalledWith({
      id: 'user-1',
      data: { reason: 'other', description: undefined },
    })
  })

  it('includes description when provided', () => {
    renderWithProviders(
      <ReportModal open targetType="review" targetId="review-1" onClose={vi.fn()} />,
    )
    fireEvent.click(screen.getByText('Спам'))
    fireEvent.change(screen.getByPlaceholderText('Опишите проблему подробнее...'), {
      target: { value: 'Это спам-отзыв' },
    })
    fireEvent.click(screen.getByText('Отправить'))

    expect(mockMutation.mutate).toHaveBeenCalledWith({
      id: 'review-1',
      data: { reason: 'spam', description: 'Это спам-отзыв' },
    })
  })

  it('calls onClose when cancel is clicked', () => {
    const onClose = vi.fn()
    renderWithProviders(
      <ReportModal open targetType="review" targetId="review-1" onClose={onClose} />,
    )
    fireEvent.click(screen.getByText('Отмена'))
    expect(onClose).toHaveBeenCalled()
  })

  it('does not render content when closed', () => {
    renderWithProviders(
      <ReportModal open={false} targetType="review" targetId="review-1" onClose={vi.fn()} />,
    )
    expect(screen.queryByText('Пожаловаться на отзыв')).not.toBeInTheDocument()
  })
})
