import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import AdminTicketDetail from '@/pages/admin/AdminTicketDetail'

vi.mock('@/api/generated/support-admin/support-admin', () => ({
  useGetAdminTicketsId: vi.fn(),
  getGetAdminTicketsIdQueryKey: vi.fn(() => ['/admin/tickets/t-1']),
  getGetAdminTicketsQueryKey: vi.fn(() => ['/admin/tickets']),
  usePatchAdminTicketsIdAssign: vi.fn(),
  usePatchAdminTicketsIdEscalate: vi.fn(),
  usePatchAdminTicketsIdResolve: vi.fn(),
  usePostAdminTicketsIdMessages: vi.fn(),
}))

vi.mock('@/api/generated/support/support', () => ({
  useGetMyTicketsIdMessages: vi.fn(),
  getGetMyTicketsIdMessagesQueryKey: vi.fn(() => ['/my/tickets/t-1/messages']),
}))

import {
  useGetAdminTicketsId,
  usePatchAdminTicketsIdAssign,
  usePatchAdminTicketsIdEscalate,
  usePatchAdminTicketsIdResolve,
  usePostAdminTicketsIdMessages,
} from '@/api/generated/support-admin/support-admin'

import {
  useGetMyTicketsIdMessages,
} from '@/api/generated/support/support'

function renderWithProviders(ui: React.ReactElement, route = '/admin/tickets/t-1') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={[route]}>
            <Routes>
              <Route path="/admin/tickets/:id" element={ui} />
              <Route path="/admin/tickets" element={<div>Ticket List</div>} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockTicket = {
  id: 't-1',
  user_id: 'u-100',
  category: 'refund_request',
  status: 'open',
  priority: 'high',
  level: 'L1',
  subject: 'Возврат не поступил',
  assigned_to: 'admin-1',
  booking_id: 'b-200',
  created_at: '2026-03-20T10:00:00Z',
}

const mockMessages = [
  {
    id: 'm-1',
    ticket_id: 't-1',
    sender_id: 'u-100',
    sender_type: 'user',
    body: 'Прошло 5 дней, возврат не поступил на карту',
    attachments: [],
    created_at: '2026-03-20T10:00:00Z',
  },
  {
    id: 'm-2',
    ticket_id: 't-1',
    sender_id: 'admin-1',
    sender_type: 'admin',
    body: 'Проверяем статус возврата',
    attachments: ['https://example.com/screenshot.png'],
    created_at: '2026-03-20T12:00:00Z',
  },
]

const mutationDefault = { mutateAsync: vi.fn(), isPending: false }

beforeEach(() => {
  vi.mocked(useGetAdminTicketsId).mockReturnValue({
    data: { data: mockTicket, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetAdminTicketsId>)

  vi.mocked(useGetMyTicketsIdMessages).mockReturnValue({
    data: { data: mockMessages, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyTicketsIdMessages>)

  vi.mocked(usePatchAdminTicketsIdAssign).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePatchAdminTicketsIdAssign>,
  )
  vi.mocked(usePatchAdminTicketsIdEscalate).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePatchAdminTicketsIdEscalate>,
  )
  vi.mocked(usePatchAdminTicketsIdResolve).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePatchAdminTicketsIdResolve>,
  )
  vi.mocked(usePostAdminTicketsIdMessages).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePostAdminTicketsIdMessages>,
  )
})

describe('AdminTicketDetail', () => {
  it('renders ticket subject as title', () => {
    renderWithProviders(<AdminTicketDetail />)

    expect(screen.getByText('Возврат не поступил')).toBeInTheDocument()
  })

  it('renders ticket info with status, category, priority', () => {
    renderWithProviders(<AdminTicketDetail />)

    expect(screen.getByText('Открыт')).toBeInTheDocument()
    expect(screen.getByText('Запрос возврата')).toBeInTheDocument()
    expect(screen.getByText('Высокий')).toBeInTheDocument()
  })

  it('renders message thread', () => {
    renderWithProviders(<AdminTicketDetail />)

    expect(screen.getByText('Прошло 5 дней, возврат не поступил на карту')).toBeInTheDocument()
    expect(screen.getByText('Проверяем статус возврата')).toBeInTheDocument()
  })

  it('renders sender labels (reversed for admin view)', () => {
    renderWithProviders(<AdminTicketDetail />)

    // "Пользователь" appears in both Descriptions label and message sender
    expect(screen.getAllByText('Пользователь').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Поддержка')).toBeInTheDocument()
  })

  it('renders action buttons for open ticket', () => {
    renderWithProviders(<AdminTicketDetail />)

    expect(screen.getByText('Назначить')).toBeInTheDocument()
    expect(screen.getByText('Эскалировать')).toBeInTheDocument()
    expect(screen.getByText('Решить')).toBeInTheDocument()
  })

  it('renders message input for open ticket', () => {
    renderWithProviders(<AdminTicketDetail />)

    expect(screen.getByPlaceholderText('Ответ пользователю...')).toBeInTheDocument()
    expect(screen.getByText('Отправить')).toBeInTheDocument()
  })

  it('shows attachment links', () => {
    renderWithProviders(<AdminTicketDetail />)

    expect(screen.getByText('Вложение 1')).toBeInTheDocument()
  })

  it('hides action buttons for resolved ticket', () => {
    vi.mocked(useGetAdminTicketsId).mockReturnValue({
      data: {
        data: { ...mockTicket, status: 'resolved', resolved_at: '2026-03-21T12:00:00Z' },
        success: true,
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminTicketsId>)

    renderWithProviders(<AdminTicketDetail />)

    expect(screen.queryByText('Назначить')).not.toBeInTheDocument()
    expect(screen.queryByText('Эскалировать')).not.toBeInTheDocument()
    expect(screen.queryByText('Решить')).not.toBeInTheDocument()
  })

  it('hides message input for closed ticket', () => {
    vi.mocked(useGetAdminTicketsId).mockReturnValue({
      data: {
        data: { ...mockTicket, status: 'closed' },
        success: true,
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminTicketsId>)

    renderWithProviders(<AdminTicketDetail />)

    expect(screen.queryByPlaceholderText('Ответ пользователю...')).not.toBeInTheDocument()
  })

  it('renders loading state', () => {
    vi.mocked(useGetAdminTicketsId).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetAdminTicketsId>)

    vi.mocked(useGetMyTicketsIdMessages).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyTicketsIdMessages>)

    renderWithProviders(<AdminTicketDetail />)

    expect(document.querySelector('.ant-spin')).toBeTruthy()
  })

  it('renders back button', () => {
    renderWithProviders(<AdminTicketDetail />)

    expect(screen.getByText('Назад')).toBeInTheDocument()
  })

  it('shows assigned info when ticket is assigned', () => {
    renderWithProviders(<AdminTicketDetail />)

    // assigned_to shows truncated UUID (slice(0,8) + "...")
    expect(screen.queryByText('Не назначен')).not.toBeInTheDocument()
  })
})
