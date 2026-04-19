import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import DisputeDetail from '@/pages/client/DisputeDetail'

vi.mock('@/api/generated/disputes/disputes', () => ({
  useGetMyDisputesId: vi.fn(),
  getGetMyDisputesIdQueryKey: vi.fn(() => ['/my/disputes/d-1']),
  useGetMyDisputesIdEvidence: vi.fn(),
  getGetMyDisputesIdEvidenceQueryKey: vi.fn(() => ['/my/disputes/d-1/evidence']),
  usePostMyDisputesIdEvidence: vi.fn(),
  usePostMyDisputesIdAppeal: vi.fn(),
  getGetMyDisputesQueryKey: vi.fn(() => ['/my/disputes']),
}))

import {
  useGetMyDisputesId,
  useGetMyDisputesIdEvidence,
  usePostMyDisputesIdEvidence,
  usePostMyDisputesIdAppeal,
} from '@/api/generated/disputes/disputes'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/client/disputes/d-1']}>
            <Routes>
              <Route path="/client/disputes/:id" element={ui} />
              <Route path="/client/disputes" element={<div>Dispute List</div>} />
              <Route path="/client/bookings/:id" element={<div>Booking Detail</div>} />
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
  description: 'Баня была грязной, не соответствует описанию',
  status: 'evidence_collection',
  evidence_deadline: '2027-04-01T10:00:00Z',
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
  {
    id: 'e-2',
    dispute_id: 'd-1',
    user_id: 'u-1',
    type: 'receipt',
    url: 'https://example.com/receipt.pdf',
    description: 'Чек об оплате',
    created_at: '2026-03-28T11:30:00Z',
  },
]

beforeEach(() => {
  vi.mocked(useGetMyDisputesId).mockReturnValue({
    data: { data: mockDispute, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyDisputesId>)

  vi.mocked(useGetMyDisputesIdEvidence).mockReturnValue({
    data: { data: mockEvidence, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyDisputesIdEvidence>)

  vi.mocked(usePostMyDisputesIdEvidence).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePostMyDisputesIdEvidence>,
  )

  vi.mocked(usePostMyDisputesIdAppeal).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePostMyDisputesIdAppeal>,
  )
})

describe('DisputeDetail', () => {
  it('renders dispute details', () => {
    renderWithProviders(<DisputeDetail />)

    expect(screen.getByText(/Спор: Низкое качество/)).toBeInTheDocument()
    expect(screen.getByText('Баня была грязной, не соответствует описанию')).toBeInTheDocument()
  })

  it('renders status and reason', () => {
    renderWithProviders(<DisputeDetail />)

    // Status tag and reason label in descriptions
    expect(screen.getAllByText('Сбор доказательств').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Низкое качество').length).toBeGreaterThanOrEqual(1)
  })

  it('renders evidence list', () => {
    renderWithProviders(<DisputeDetail />)

    expect(screen.getByText('Доказательства')).toBeInTheDocument()
    expect(screen.getByText('Фото грязной комнаты')).toBeInTheDocument()
    expect(screen.getByText('Чек об оплате')).toBeInTheDocument()
  })

  it('renders evidence type tags', () => {
    renderWithProviders(<DisputeDetail />)

    expect(screen.getByText('Фото')).toBeInTheDocument()
    expect(screen.getByText('Чек')).toBeInTheDocument()
  })

  it('renders add evidence button when evidence window open', () => {
    renderWithProviders(<DisputeDetail />)

    expect(screen.getByText('Добавить')).toBeInTheDocument()
  })

  it('does not render appeal button for non-resolved dispute', () => {
    renderWithProviders(<DisputeDetail />)

    expect(screen.queryByText('Подать апелляцию')).not.toBeInTheDocument()
  })

  it('renders appeal button for resolved dispute within appeal window', () => {
    vi.mocked(useGetMyDisputesId).mockReturnValue({
      data: {
        data: {
          ...mockDispute,
          status: 'resolved',
          resolution: 'no_refund',
          refund_amount: 0,
          appeal_deadline: '2027-04-10T10:00:00Z',
          resolved_at: '2026-03-28T14:00:00Z',
        },
        success: true,
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyDisputesId>)

    renderWithProviders(<DisputeDetail />)

    expect(screen.getByText('Подать апелляцию')).toBeInTheDocument()
  })

  it('renders empty state when dispute not found', () => {
    vi.mocked(useGetMyDisputesId).mockReturnValue({
      data: { data: null, success: false },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyDisputesId>)

    renderWithProviders(<DisputeDetail />)

    expect(screen.getByText('Спор не найден')).toBeInTheDocument()
  })

  it('renders loading state', () => {
    vi.mocked(useGetMyDisputesId).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyDisputesId>)

    vi.mocked(useGetMyDisputesIdEvidence).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyDisputesIdEvidence>)

    renderWithProviders(<DisputeDetail />)

    expect(document.querySelector('.ant-spin')).toBeTruthy()
  })
})
