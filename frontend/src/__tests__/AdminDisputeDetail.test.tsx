import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import AdminDisputeDetail from '@/pages/admin/AdminDisputeDetail'

vi.mock('@/api/generated/disputes-admin/disputes-admin', () => ({
  useGetAdminDisputesId: vi.fn(),
  getGetAdminDisputesIdQueryKey: vi.fn(() => ['/admin/disputes/d-1']),
  getGetAdminDisputesQueryKey: vi.fn(() => ['/admin/disputes']),
  usePatchAdminDisputesIdAssign: vi.fn(),
  usePatchAdminDisputesIdResolve: vi.fn(),
  usePatchAdminDisputesIdClose: vi.fn(),
}))

vi.mock('@/api/generated/disputes/disputes', () => ({
  useGetMyDisputesIdEvidence: vi.fn(),
  getGetMyDisputesIdEvidenceQueryKey: vi.fn(() => ['/my/disputes/d-1/evidence']),
}))

import {
  useGetAdminDisputesId,
  usePatchAdminDisputesIdAssign,
  usePatchAdminDisputesIdResolve,
  usePatchAdminDisputesIdClose,
} from '@/api/generated/disputes-admin/disputes-admin'

import {
  useGetMyDisputesIdEvidence,
} from '@/api/generated/disputes/disputes'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/disputes/d-1']}>
            <Routes>
              <Route path="/admin/disputes/:id" element={ui} />
              <Route path="/admin/disputes" element={<div>Admin Disputes</div>} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mutationDefault = { mutateAsync: vi.fn(), isPending: false, isError: false }

const mockDispute = {
  id: 'd-1',
  booking_id: 'b-1',
  initiator_id: 'u-1',
  respondent_id: 'u-2',
  reason: 'poor_quality',
  description: 'Баня была грязной',
  status: 'under_review',
  mediator_id: 'admin-1',
  evidence_deadline: '2026-03-30T10:00:00Z',
  created_at: '2026-03-28T10:00:00Z',
}

const mockEvidence = [
  {
    id: 'e-1',
    dispute_id: 'd-1',
    user_id: 'u-1',
    type: 'photo',
    url: 'https://example.com/photo1.jpg',
    description: 'Фото грязной комнаты',
    created_at: '2026-03-28T11:00:00Z',
  },
]

beforeEach(() => {
  vi.mocked(useGetAdminDisputesId).mockReturnValue({
    data: { data: mockDispute, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetAdminDisputesId>)

  vi.mocked(useGetMyDisputesIdEvidence).mockReturnValue({
    data: { data: mockEvidence, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyDisputesIdEvidence>)

  vi.mocked(usePatchAdminDisputesIdAssign).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePatchAdminDisputesIdAssign>,
  )
  vi.mocked(usePatchAdminDisputesIdResolve).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePatchAdminDisputesIdResolve>,
  )
  vi.mocked(usePatchAdminDisputesIdClose).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePatchAdminDisputesIdClose>,
  )
})

describe('AdminDisputeDetail', () => {
  it('renders dispute details', () => {
    renderWithProviders(<AdminDisputeDetail />)

    expect(screen.getByText(/Спор: Низкое качество/)).toBeInTheDocument()
    expect(screen.getByText('Баня была грязной')).toBeInTheDocument()
  })

  it('renders admin action buttons for under_review dispute', () => {
    renderWithProviders(<AdminDisputeDetail />)

    expect(screen.getByText('Назначить медиатора')).toBeInTheDocument()
    expect(screen.getByText('Решить спор')).toBeInTheDocument()
  })

  it('renders parties information', () => {
    renderWithProviders(<AdminDisputeDetail />)

    // Initiator and respondent IDs are shown truncated
    expect(screen.getByText('u-1...')).toBeInTheDocument()
    expect(screen.getByText('u-2...')).toBeInTheDocument()
    expect(screen.getByText('admin-1...')).toBeInTheDocument()
  })

  it('renders evidence list', () => {
    renderWithProviders(<AdminDisputeDetail />)

    expect(screen.getByText('Доказательства')).toBeInTheDocument()
    expect(screen.getByText('Фото грязной комнаты')).toBeInTheDocument()
    expect(screen.getByText('Фото')).toBeInTheDocument()
  })

  it('renders close button for resolved dispute', () => {
    vi.mocked(useGetAdminDisputesId).mockReturnValue({
      data: {
        data: {
          ...mockDispute,
          status: 'resolved',
          resolution: 'full_refund',
          refund_amount: 500000,
          resolved_at: '2026-03-29T10:00:00Z',
        },
        success: true,
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminDisputesId>)

    renderWithProviders(<AdminDisputeDetail />)

    expect(screen.getByText('Закрыть')).toBeInTheDocument()
  })

  it('renders empty state when dispute not found', () => {
    vi.mocked(useGetAdminDisputesId).mockReturnValue({
      data: { data: null, success: false },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminDisputesId>)

    renderWithProviders(<AdminDisputeDetail />)

    expect(screen.getByText('Спор не найден')).toBeInTheDocument()
  })

  it('renders loading state', () => {
    vi.mocked(useGetAdminDisputesId).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetAdminDisputesId>)

    vi.mocked(useGetMyDisputesIdEvidence).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyDisputesIdEvidence>)

    renderWithProviders(<AdminDisputeDetail />)

    expect(document.querySelector('.ant-spin')).toBeTruthy()
  })
})
