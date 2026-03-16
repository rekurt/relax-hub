import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi } from 'vitest'
import PaymentHistory from '@/pages/client/PaymentHistory'

vi.mock('@/api/generated/payments/payments', () => ({
  useGetMyPayments: vi.fn(),
}))

import { useGetMyPayments } from '@/api/generated/payments/payments'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/client/payments']}>
            <Routes>
              <Route path="/client/payments" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockPayments = [
  {
    id: 'pay-1',
    amount: 600000,
    status: 'succeeded',
    booking_id: 'book-1',
    created_at: '2026-03-10T14:00:00Z',
    currency: 'RUB',
    refund_amount: 0,
    refunded_at: null,
  },
  {
    id: 'pay-2',
    amount: 400000,
    status: 'pending',
    booking_id: 'book-2',
    created_at: '2026-03-12T10:00:00Z',
    currency: 'RUB',
    refund_amount: 0,
    refunded_at: null,
  },
  {
    id: 'pay-3',
    amount: 500000,
    status: 'refunded',
    booking_id: 'book-3',
    created_at: '2026-03-08T09:00:00Z',
    currency: 'RUB',
    refund_amount: 500000,
    refunded_at: '2026-03-09T12:00:00Z',
  },
  {
    id: 'pay-4',
    amount: 300000,
    status: 'canceled',
    booking_id: null,
    created_at: '2026-03-05T16:00:00Z',
    currency: 'RUB',
    refund_amount: 0,
    refunded_at: null,
  },
]

describe('PaymentHistory', () => {
  it('renders payment history table with data', () => {
    vi.mocked(useGetMyPayments).mockReturnValue({
      data: {
        data: mockPayments,
        success: true,
        meta: { page: 1, page_size: 10, total_count: 4, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyPayments>)

    renderWithProviders(<PaymentHistory />)

    expect(screen.getByText('История платежей')).toBeInTheDocument()
    expect(screen.getByText('6000 ₽')).toBeInTheDocument()
    expect(screen.getByText('4000 ₽')).toBeInTheDocument()
    expect(screen.getByText('3000 ₽')).toBeInTheDocument()
  })

  it('renders payment status badges', () => {
    vi.mocked(useGetMyPayments).mockReturnValue({
      data: {
        data: mockPayments,
        success: true,
        meta: { page: 1, page_size: 10, total_count: 4, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyPayments>)

    renderWithProviders(<PaymentHistory />)

    expect(screen.getByText('Оплачено')).toBeInTheDocument()
    expect(screen.getByText('Возвращён')).toBeInTheDocument()
    expect(screen.getByText('Отменён')).toBeInTheDocument()
    // "Ожидает оплаты" appears both in status column and in the filter options
    expect(screen.getAllByText('Ожидает оплаты').length).toBeGreaterThanOrEqual(1)
  })

  it('renders empty state when no payments', () => {
    vi.mocked(useGetMyPayments).mockReturnValue({
      data: {
        data: [],
        success: true,
        meta: { page: 1, page_size: 10, total_count: 0, total_pages: 0 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyPayments>)

    renderWithProviders(<PaymentHistory />)

    expect(screen.getByText('Нет платежей')).toBeInTheDocument()
  })

  it('renders status filter dropdown', () => {
    vi.mocked(useGetMyPayments).mockReturnValue({
      data: { data: [], success: true, meta: {} },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyPayments>)

    renderWithProviders(<PaymentHistory />)

    expect(screen.getByText('Все статусы')).toBeInTheDocument()
  })

  it('renders total count in pagination', () => {
    vi.mocked(useGetMyPayments).mockReturnValue({
      data: {
        data: mockPayments,
        success: true,
        meta: { page: 1, page_size: 10, total_count: 4, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyPayments>)

    renderWithProviders(<PaymentHistory />)

    expect(screen.getByText('Всего: 4')).toBeInTheDocument()
  })

  it('renders booking links for payments with booking_id', () => {
    vi.mocked(useGetMyPayments).mockReturnValue({
      data: {
        data: mockPayments,
        success: true,
        meta: { page: 1, page_size: 10, total_count: 4, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyPayments>)

    renderWithProviders(<PaymentHistory />)

    const openLinks = screen.getAllByText('Открыть')
    // 3 payments have booking_id, 1 doesn't
    expect(openLinks.length).toBe(3)
  })

  it('renders refund info for refunded payment', () => {
    vi.mocked(useGetMyPayments).mockReturnValue({
      data: {
        data: mockPayments,
        success: true,
        meta: { page: 1, page_size: 10, total_count: 4, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyPayments>)

    renderWithProviders(<PaymentHistory />)

    // refund amount for pay-3
    expect(screen.getByText('5000 ₽')).toBeInTheDocument()
  })
})
