import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import TicketDetail from '@/pages/client/TicketDetail'

vi.mock('@/api/generated/support/support', () => ({
  useGetMyTicketsId: vi.fn(),
  useGetMyTicketsIdMessages: vi.fn(),
  getGetMyTicketsIdMessagesQueryKey: vi.fn(() => ['/my/tickets/t-1/messages']),
  getGetMyTicketsIdQueryKey: vi.fn(() => ['/my/tickets/t-1']),
  usePostMyTicketsIdMessages: vi.fn(),
  usePostMyTicketsIdCsat: vi.fn(),
}))

import {
  useGetMyTicketsId,
  useGetMyTicketsIdMessages,
  usePostMyTicketsIdMessages,
  usePostMyTicketsIdCsat,
} from '@/api/generated/support/support'

function renderWithProviders(ui: React.ReactElement, route = '/client/tickets/t-1') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={[route]}>
            <Routes>
              <Route path="/client/tickets/:id" element={ui} />
              <Route path="/client/tickets" element={<div>Ticket List</div>} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockTicket = {
  id: 't-1',
  user_id: 'u-1',
  category: 'problem',
  status: 'open',
  priority: 'medium',
  level: 'L1',
  subject: 'Не проходит оплата',
  created_at: '2026-03-20T10:00:00Z',
}

const mockMessages = [
  {
    id: 'm-1',
    ticket_id: 't-1',
    sender_id: 'u-1',
    sender_type: 'user',
    body: 'Оплата не проходит, карта отклоняется',
    attachments: [],
    created_at: '2026-03-20T10:00:00Z',
  },
  {
    id: 'm-2',
    ticket_id: 't-1',
    sender_id: 'admin-1',
    sender_type: 'admin',
    body: 'Попробуйте другую карту или способ оплаты',
    attachments: [],
    created_at: '2026-03-20T11:00:00Z',
  },
]

const mutationDefault = { mutateAsync: vi.fn(), isPending: false, isError: false }

beforeEach(() => {
  vi.mocked(useGetMyTicketsId).mockReturnValue({
    data: { data: mockTicket, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyTicketsId>)

  vi.mocked(useGetMyTicketsIdMessages).mockReturnValue({
    data: { data: mockMessages, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyTicketsIdMessages>)

  vi.mocked(usePostMyTicketsIdMessages).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePostMyTicketsIdMessages>,
  )
  vi.mocked(usePostMyTicketsIdCsat).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePostMyTicketsIdCsat>,
  )
})

describe('TicketDetail', () => {
  it('renders ticket subject as title', () => {
    renderWithProviders(<TicketDetail />)

    expect(screen.getByText('Не проходит оплата')).toBeInTheDocument()
  })

  it('renders ticket info card with status and category', () => {
    renderWithProviders(<TicketDetail />)

    expect(screen.getByText('Открыт')).toBeInTheDocument()
    expect(screen.getByText('Проблема')).toBeInTheDocument()
    expect(screen.getByText('Средний')).toBeInTheDocument()
  })

  it('renders message thread', () => {
    renderWithProviders(<TicketDetail />)

    expect(screen.getByText('Оплата не проходит, карта отклоняется')).toBeInTheDocument()
    expect(screen.getByText('Попробуйте другую карту или способ оплаты')).toBeInTheDocument()
  })

  it('renders sender labels', () => {
    renderWithProviders(<TicketDetail />)

    expect(screen.getByText('Вы')).toBeInTheDocument()
    expect(screen.getByText('Поддержка')).toBeInTheDocument()
  })

  it('renders message input for open ticket', () => {
    renderWithProviders(<TicketDetail />)

    expect(screen.getByPlaceholderText('Введите сообщение...')).toBeInTheDocument()
    expect(screen.getByText('Отправить')).toBeInTheDocument()
  })

  it('shows CSAT form for resolved ticket', () => {
    vi.mocked(useGetMyTicketsId).mockReturnValue({
      data: {
        data: { ...mockTicket, status: 'resolved', resolved_at: '2026-03-21T12:00:00Z' },
        success: true,
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyTicketsId>)

    renderWithProviders(<TicketDetail />)

    expect(screen.getByText('Оцените качество поддержки')).toBeInTheDocument()
    expect(screen.getByText('Отправить оценку')).toBeInTheDocument()
  })

  it('does not show CSAT form if already submitted', () => {
    vi.mocked(useGetMyTicketsId).mockReturnValue({
      data: {
        data: { ...mockTicket, status: 'resolved', csat_score: 5 },
        success: true,
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyTicketsId>)

    renderWithProviders(<TicketDetail />)

    expect(screen.queryByText('Оцените качество поддержки')).not.toBeInTheDocument()
  })

  it('does not show message input for closed ticket', () => {
    vi.mocked(useGetMyTicketsId).mockReturnValue({
      data: {
        data: { ...mockTicket, status: 'closed' },
        success: true,
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyTicketsId>)

    renderWithProviders(<TicketDetail />)

    expect(screen.queryByPlaceholderText('Введите сообщение...')).not.toBeInTheDocument()
  })

  it('renders loading state', () => {
    vi.mocked(useGetMyTicketsId).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyTicketsId>)

    vi.mocked(useGetMyTicketsIdMessages).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyTicketsIdMessages>)

    renderWithProviders(<TicketDetail />)

    expect(document.querySelector('.ant-spin')).toBeTruthy()
  })

  it('renders back button', () => {
    renderWithProviders(<TicketDetail />)

    expect(screen.getByText('Назад')).toBeInTheDocument()
  })
})
