import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import SupportTickets from '@/pages/client/SupportTickets'

vi.mock('@/api/generated/support/support', () => ({
  useGetMyTickets: vi.fn(),
  getGetMyTicketsQueryKey: vi.fn(() => ['/my/tickets']),
  usePostMyTickets: vi.fn(),
}))

import {
  useGetMyTickets,
  usePostMyTickets,
} from '@/api/generated/support/support'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/client/tickets']}>
            <Routes>
              <Route path="/client/tickets" element={ui} />
              <Route path="/client/tickets/:id" element={<div>Ticket Detail</div>} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockTickets = [
  {
    id: 't-1',
    user_id: 'u-1',
    category: 'question',
    status: 'open',
    priority: 'low',
    level: 'L1',
    subject: 'Как отменить бронирование?',
    created_at: '2026-03-20T10:00:00Z',
  },
  {
    id: 't-2',
    user_id: 'u-1',
    category: 'refund_request',
    status: 'resolved',
    priority: 'high',
    level: 'L2',
    subject: 'Возврат за некачественную услугу',
    csat_score: 4,
    created_at: '2026-03-18T08:00:00Z',
    resolved_at: '2026-03-19T14:00:00Z',
  },
  {
    id: 't-3',
    user_id: 'u-1',
    category: 'problem',
    status: 'in_progress',
    priority: 'medium',
    level: 'L1',
    subject: 'Не проходит оплата',
    created_at: '2026-03-21T12:00:00Z',
  },
]

const mutationDefault = { mutateAsync: vi.fn(), isPending: false, isError: false }

beforeEach(() => {
  vi.mocked(useGetMyTickets).mockReturnValue({
    data: {
      data: mockTickets,
      success: true,
      meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 },
    },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyTickets>)

  vi.mocked(usePostMyTickets).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePostMyTickets>,
  )
})

describe('SupportTickets', () => {
  it('renders page title and ticket list', () => {
    renderWithProviders(<SupportTickets />)

    expect(screen.getByText('Мои обращения')).toBeInTheDocument()
    expect(screen.getByText('Как отменить бронирование?')).toBeInTheDocument()
    expect(screen.getByText('Возврат за некачественную услугу')).toBeInTheDocument()
    expect(screen.getByText('Не проходит оплата')).toBeInTheDocument()
  })

  it('renders status filter options', () => {
    renderWithProviders(<SupportTickets />)

    expect(screen.getByText('Все')).toBeInTheDocument()
    expect(screen.getByText('Открытые')).toBeInTheDocument()
    // "В работе" appears in both segmented filter and status tag
    expect(screen.getAllByText('В работе').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Решённые')).toBeInTheDocument()
    expect(screen.getByText('Закрытые')).toBeInTheDocument()
  })

  it('renders status tags', () => {
    renderWithProviders(<SupportTickets />)

    expect(screen.getByText('Открыт')).toBeInTheDocument()
    expect(screen.getByText('Решён')).toBeInTheDocument()
    // "В работе" appears in both segmented filter and status tag
    expect(screen.getAllByText('В работе').length).toBeGreaterThanOrEqual(1)
  })

  it('renders priority tags', () => {
    renderWithProviders(<SupportTickets />)

    expect(screen.getByText('Низкий')).toBeInTheDocument()
    expect(screen.getByText('Высокий')).toBeInTheDocument()
    expect(screen.getByText('Средний')).toBeInTheDocument()
  })

  it('renders category labels', () => {
    renderWithProviders(<SupportTickets />)

    expect(screen.getByText('Вопрос')).toBeInTheDocument()
    expect(screen.getByText('Запрос возврата')).toBeInTheDocument()
    expect(screen.getByText('Проблема')).toBeInTheDocument()
  })

  it('shows "Новое обращение" button', () => {
    renderWithProviders(<SupportTickets />)

    expect(screen.getByText('Новое обращение')).toBeInTheDocument()
  })

  it('opens create modal when clicking button', async () => {
    renderWithProviders(<SupportTickets />)

    fireEvent.click(screen.getByText('Новое обращение'))

    await waitFor(() => {
      // Modal renders with its content in the DOM
      expect(document.querySelector('.ant-modal')).toBeTruthy()
    })
  })

  it('renders empty state when no tickets', () => {
    vi.mocked(useGetMyTickets).mockReturnValue({
      data: {
        data: [],
        success: true,
        meta: { page: 1, page_size: 20, total_count: 0, total_pages: 0 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyTickets>)

    renderWithProviders(<SupportTickets />)

    expect(screen.getByText('Нет обращений')).toBeInTheDocument()
    expect(screen.getByText('Создать обращение')).toBeInTheDocument()
  })

  it('renders loading state', () => {
    vi.mocked(useGetMyTickets).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyTickets>)

    renderWithProviders(<SupportTickets />)

    expect(screen.getByText('Мои обращения')).toBeInTheDocument()
    expect(document.querySelector('.ant-spin')).toBeTruthy()
  })
})
