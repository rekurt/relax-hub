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

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    get: vi.fn(),
  },
}))

import {
  useGetAdminTickets,
  useGetAdminTicketsStats,
} from '@/api/generated/support-admin/support-admin'
import { axiosInstance } from '@/api/axios-instance'

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

const mockMetrics = {
  fcr_percent: 72.0,
  aht_seconds: 2400,
  avg_csat: 4.1,
  sla_compliance_percent: 88.5,
  total_resolved: 150,
  total_tickets: 200,
}

const mockAgentThroughput = [
  { agent_id: 'a-1', agent_name: 'Иван Петров', resolved_count: 45, avg_resolution_seconds: 1800, csat_avg: 4.3 },
  { agent_id: 'a-2', agent_name: 'Мария Сидорова', resolved_count: 38, avg_resolution_seconds: 2100, csat_avg: 4.5 },
]

beforeEach(() => {
  vi.mocked(axiosInstance.get).mockImplementation((url: string) => {
    if (url === '/admin/tickets/metrics') {
      return Promise.resolve({ data: { data: mockMetrics } })
    }
    if (url === '/admin/tickets/agent-throughput') {
      return Promise.resolve({ data: { data: mockAgentThroughput } })
    }
    return Promise.reject(new Error('not found'))
  })

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

  it('renders operation metrics when available', async () => {
    renderWithProviders(<TicketManagement />)

    expect(await screen.findByText('Операционные метрики')).toBeInTheDocument()
    expect(screen.getByText('FCR')).toBeInTheDocument()
    expect(screen.getByText('AHT')).toBeInTheDocument()
    expect(screen.getByText('SLA (24ч)')).toBeInTheDocument()
  })

  it('renders escalation indicator on escalated tickets', () => {
    renderWithProviders(<TicketManagement />)

    const escalationIcons = screen.getAllByTestId('escalation-icon')
    expect(escalationIcons.length).toBe(1)
  })

  it('renders SLA timer for recent open tickets', () => {
    const recentDate = new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString()
    vi.mocked(useGetAdminTickets).mockReturnValue({
      data: {
        data: [
          {
            id: 't-recent',
            user_id: 'u-200',
            category: 'question',
            status: 'open',
            priority: 'medium',
            level: 'L1',
            subject: 'Свежий тикет',
            created_at: recentDate,
          },
        ],
        success: true,
        meta: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminTickets>)

    renderWithProviders(<TicketManagement />)

    const slaTimers = screen.getAllByTestId('sla-timer')
    expect(slaTimers.length).toBe(1)
  })

  it('renders SLA breached tag for old tickets', () => {
    vi.mocked(useGetAdminTickets).mockReturnValue({
      data: {
        data: [
          {
            id: 't-old',
            user_id: 'u-999',
            category: 'problem',
            status: 'open',
            priority: 'critical',
            level: 'L1',
            subject: 'Просроченный тикет',
            created_at: '2026-03-01T10:00:00Z',
          },
        ],
        success: true,
        meta: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminTickets>)

    renderWithProviders(<TicketManagement />)

    expect(screen.getByTestId('sla-breached')).toBeInTheDocument()
    expect(screen.getByText('Просрочен')).toBeInTheDocument()
  })

  it('renders agent throughput card', async () => {
    renderWithProviders(<TicketManagement />)

    expect(await screen.findByTestId('agent-throughput-card')).toBeInTheDocument()
    expect(screen.getByText('Производительность агентов')).toBeInTheDocument()
  })
})
