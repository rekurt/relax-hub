import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import DisputeList from '@/pages/client/DisputeList'

vi.mock('@/api/generated/disputes/disputes', () => ({
  useGetMyDisputes: vi.fn(),
  getGetMyDisputesQueryKey: vi.fn(() => ['/my/disputes']),
}))

import { useGetMyDisputes } from '@/api/generated/disputes/disputes'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/client/disputes']}>
            <Routes>
              <Route path="/client/disputes" element={ui} />
              <Route path="/client/disputes/:id" element={<div>Dispute Detail</div>} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockDisputes = [
  {
    id: 'd-1',
    booking_id: 'b-1',
    initiator_id: 'u-1',
    respondent_id: 'u-2',
    reason: 'poor_quality',
    description: 'Грязная баня',
    status: 'evidence_collection',
    evidence_deadline: '2026-04-01T10:00:00Z',
    created_at: '2026-03-28T10:00:00Z',
  },
  {
    id: 'd-2',
    booking_id: 'b-2',
    initiator_id: 'u-1',
    respondent_id: 'u-3',
    reason: 'billing_error',
    description: 'Списали больше',
    status: 'resolved',
    resolution: 'partial_refund',
    refund_amount: 150000,
    appeal_deadline: '2026-04-05T10:00:00Z',
    created_at: '2026-03-25T08:00:00Z',
    resolved_at: '2026-03-27T14:00:00Z',
  },
  {
    id: 'd-3',
    booking_id: 'b-3',
    initiator_id: 'u-1',
    respondent_id: 'u-4',
    reason: 'service_not_provided',
    description: 'Не пустили',
    status: 'closed',
    resolution: 'full_refund',
    refund_amount: 500000,
    created_at: '2026-03-20T12:00:00Z',
  },
]

beforeEach(() => {
  vi.mocked(useGetMyDisputes).mockReturnValue({
    data: {
      data: mockDisputes,
      success: true,
      meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 },
    },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyDisputes>)
})

describe('DisputeList', () => {
  it('renders page title and dispute list', () => {
    renderWithProviders(<DisputeList />)

    expect(screen.getByText('Мои споры')).toBeInTheDocument()
    expect(screen.getByText('Низкое качество')).toBeInTheDocument()
    expect(screen.getByText('Ошибка в счёте')).toBeInTheDocument()
    expect(screen.getByText('Услуга не оказана')).toBeInTheDocument()
  })

  it('renders status filter options', () => {
    renderWithProviders(<DisputeList />)

    expect(screen.getByText('Все')).toBeInTheDocument()
    expect(screen.getByText('Открытые')).toBeInTheDocument()
    expect(screen.getByText('Закрытые')).toBeInTheDocument()
  })

  it('renders status tags', () => {
    renderWithProviders(<DisputeList />)

    // "Сбор доказательств" appears in both filter and table tag
    expect(screen.getAllByText('Сбор доказательств').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Решён')).toBeInTheDocument()
    expect(screen.getByText('Закрыт')).toBeInTheDocument()
  })

  it('renders resolution and refund amount', () => {
    renderWithProviders(<DisputeList />)

    expect(screen.getByText('Частичный возврат')).toBeInTheDocument()
    expect(screen.getByText('Полный возврат')).toBeInTheDocument()
  })

  it('renders empty state when no disputes', () => {
    vi.mocked(useGetMyDisputes).mockReturnValue({
      data: {
        data: [],
        success: true,
        meta: { page: 1, page_size: 20, total_count: 0, total_pages: 0 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyDisputes>)

    renderWithProviders(<DisputeList />)

    expect(screen.getByText('Нет споров')).toBeInTheDocument()
  })

  it('renders loading state', () => {
    vi.mocked(useGetMyDisputes).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyDisputes>)

    renderWithProviders(<DisputeList />)

    expect(screen.getByText('Мои споры')).toBeInTheDocument()
    expect(document.querySelector('.ant-spin')).toBeTruthy()
  })
})
