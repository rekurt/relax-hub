import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import DisputeManagement from '@/pages/admin/DisputeManagement'

vi.mock('@/api/generated/disputes-admin/disputes-admin', () => ({
  useGetAdminDisputes: vi.fn(),
  getGetAdminDisputesQueryKey: vi.fn(() => ['/admin/disputes']),
}))

import { useGetAdminDisputes } from '@/api/generated/disputes-admin/disputes-admin'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/disputes']}>
            <Routes>
              <Route path="/admin/disputes" element={ui} />
              <Route path="/admin/disputes/:id" element={<div>Admin Dispute Detail</div>} />
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
    status: 'evidence_collection',
    created_at: '2026-03-28T10:00:00Z',
  },
  {
    id: 'd-2',
    booking_id: 'b-2',
    initiator_id: 'u-3',
    respondent_id: 'u-4',
    reason: 'billing_error',
    status: 'under_review',
    mediator_id: 'admin-1',
    created_at: '2026-03-27T08:00:00Z',
  },
  {
    id: 'd-3',
    booking_id: 'b-3',
    initiator_id: 'u-5',
    respondent_id: 'u-6',
    reason: 'service_not_provided',
    status: 'resolved',
    resolution: 'full_refund',
    refund_amount: 500000,
    created_at: '2026-03-25T12:00:00Z',
  },
]

beforeEach(() => {
  vi.mocked(useGetAdminDisputes).mockReturnValue({
    data: {
      data: mockDisputes,
      success: true,
      meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 },
    },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetAdminDisputes>)
})

describe('DisputeManagement', () => {
  it('renders page title and dispute list', () => {
    renderWithProviders(<DisputeManagement />)

    expect(screen.getByText('Управление спорами')).toBeInTheDocument()
    expect(screen.getByText('Низкое качество')).toBeInTheDocument()
    expect(screen.getByText('Ошибка в счёте')).toBeInTheDocument()
    expect(screen.getByText('Услуга не оказана')).toBeInTheDocument()
  })

  it('renders status filter options', () => {
    renderWithProviders(<DisputeManagement />)

    expect(screen.getByText('Все')).toBeInTheDocument()
    expect(screen.getByText('Открытые')).toBeInTheDocument()
    // "На рассмотрении" appears in both filter and table status tag
    expect(screen.getAllByText('На рассмотрении').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Закрытые')).toBeInTheDocument()
  })

  it('renders mediator column', () => {
    renderWithProviders(<DisputeManagement />)

    expect(screen.getByText('admin-1...')).toBeInTheDocument()
    expect(screen.getAllByText('Не назначен').length).toBeGreaterThanOrEqual(1)
  })

  it('renders empty state', () => {
    vi.mocked(useGetAdminDisputes).mockReturnValue({
      data: {
        data: [],
        success: true,
        meta: { page: 1, page_size: 20, total_count: 0, total_pages: 0 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminDisputes>)

    renderWithProviders(<DisputeManagement />)

    expect(screen.getByText('Нет споров')).toBeInTheDocument()
  })

  it('renders loading state', () => {
    vi.mocked(useGetAdminDisputes).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetAdminDisputes>)

    renderWithProviders(<DisputeManagement />)

    expect(screen.getByText('Управление спорами')).toBeInTheDocument()
    expect(document.querySelector('.ant-spin')).toBeTruthy()
  })
})
