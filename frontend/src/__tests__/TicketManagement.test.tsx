import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import TicketManagement from '@/pages/admin/TicketManagement'

vi.mock('@/api/generated/support-admin/support-admin', () => ({
  useGetAdminTickets: vi.fn(),
  getGetAdminTicketsQueryKey: vi.fn(() => ['/admin/tickets']),
  useGetAdminTicketsStats: vi.fn(),
}))

import {
  useGetAdminTickets,
  useGetAdminTicketsStats,
} from '@/api/generated/support-admin/support-admin'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/tickets']}>
            <Routes>
              <Route path="/admin/tickets" element={ui} />
              <Route path="/admin/tickets/:id" element={<div>Admin Ticket Detail</div>} />
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
    user_id: 'u-100',
    category: 'question',
    status: 'open',
    priority: 'low',
    level: 'L1',
    subject: 'Вопрос по бронированию',
    created_at: '2026-03-20T10:00:00Z',
  },
  {
    id: 't-2',
    user_id: 'u-101',
    category: 'refund_request',
    status: 'escalated',
    priority: 'high',
    level: 'L2',
    subject: 'Возврат средств не поступил',
    assigned_to: 'admin-1',
    created_at: '2026-03-19T14:00:00Z',
  },
  {
    id: 't-3',
    user_id: 'u-102',
    category: 'complaint',
    status: 'resolved',
    priority: 'high',
    level: 'L1',
    subject: 'Жалоба на обслуживание',
    csat_score: 3,
    created_at: '2026-03-18T08:00:00Z',
    resolved_at: '2026-03-19T16:00:00Z',
  },
]

const mockStats = {
  open: 5,
  in_progress: 3,
  escalated: 2,
  resolved: 10,
  closed: 20,
}

beforeEach(() => {
  vi.mocked(useGetAdminTickets).mockReturnValue({
    data: {
      data: mockTickets,
      success: true,
      meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 },
    },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetAdminTickets>)

  vi.mocked(useGetAdminTicketsStats).mockReturnValue({
    data: { data: mockStats, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetAdminTicketsStats>)
})

describe('TicketManagement', () => {
  it('renders page title and ticket list', () => {
    renderWithProviders(<TicketManagement />)

    expect(screen.getByText('Управление обращениями')).toBeInTheDocument()
    expect(screen.getByText('Вопрос по бронированию')).toBeInTheDocument()
    expect(screen.getByText('Возврат средств не поступил')).toBeInTheDocument()
    expect(screen.getByText('Жалоба на обслуживание')).toBeInTheDocument()
  })

  it('renders status filter options', () => {
    renderWithProviders(<TicketManagement />)

    expect(screen.getByText('Все')).toBeInTheDocument()
    // Labels appear in both stats cards and segmented filter
    expect(screen.getAllByText('Открытые').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('В работе').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Эскалированные').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Решённые').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Закрытые').length).toBeGreaterThanOrEqual(1)
  })

  it('renders stats cards', () => {
    renderWithProviders(<TicketManagement />)

    // Stats values
    expect(screen.getByText('5')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.getByText('10')).toBeInTheDocument()
    expect(screen.getByText('20')).toBeInTheDocument()
  })

  it('renders status tags in table', () => {
    renderWithProviders(<TicketManagement />)

    expect(screen.getByText('Открыт')).toBeInTheDocument()
    expect(screen.getByText('Эскалирован')).toBeInTheDocument()
    expect(screen.getByText('Решён')).toBeInTheDocument()
  })

  it('renders priority tags', () => {
    renderWithProviders(<TicketManagement />)

    expect(screen.getByText('Низкий')).toBeInTheDocument()
    expect(screen.getAllByText('Высокий').length).toBeGreaterThanOrEqual(1)
  })

  it('renders level tags', () => {
    renderWithProviders(<TicketManagement />)

    expect(screen.getAllByText('L1').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('L2')).toBeInTheDocument()
  })

  it('renders empty state when no tickets', () => {
    vi.mocked(useGetAdminTickets).mockReturnValue({
      data: {
        data: [],
        success: true,
        meta: { page: 1, page_size: 20, total_count: 0, total_pages: 0 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminTickets>)

    renderWithProviders(<TicketManagement />)

    expect(screen.getByText('Нет обращений')).toBeInTheDocument()
  })

  it('renders loading state', () => {
    vi.mocked(useGetAdminTickets).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetAdminTickets>)

    renderWithProviders(<TicketManagement />)

    expect(screen.getByText('Управление обращениями')).toBeInTheDocument()
    expect(document.querySelector('.ant-spin')).toBeTruthy()
  })
})
