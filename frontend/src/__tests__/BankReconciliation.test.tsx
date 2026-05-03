import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi } from 'vitest'
import BankReconciliation from '@/pages/admin/BankReconciliation'

vi.mock('@/api/generated/admin/admin', () => ({
  useGetApiV1AdminFinanceReconciliation: vi.fn(),
  usePostApiV1AdminFinanceBankStatement: vi.fn(),
  usePutApiV1AdminFinanceReconciliationIdMatch: vi.fn(),
}))

import {
  useGetApiV1AdminFinanceReconciliation,
  usePostApiV1AdminFinanceBankStatement,
  usePutApiV1AdminFinanceReconciliationIdMatch,
} from '@/api/generated/admin/admin'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/finance/reconciliation']}>
            <Routes>
              <Route path="/admin/finance/reconciliation" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockEntries = [
  {
    id: 'e-1',
    date: '2026-03-28T00:00:00Z',
    amount: 150000,
    reference_num: 'REF-001',
    counterparty: 'ООО Баня',
    description: 'Оплата за услуги',
    status: 'pending',
  },
  {
    id: 'e-2',
    date: '2026-03-27T00:00:00Z',
    amount: 250000,
    reference_num: 'REF-002',
    counterparty: 'ИП Иванов',
    description: 'Возврат средств',
    status: 'matched',
    matched_tx_id: 'tx-12345678-abcd-efgh',
    matched_tx_type: 'payment',
  },
]

function setupMocks() {
  vi.mocked(useGetApiV1AdminFinanceReconciliation).mockReturnValue({
    data: { data: mockEntries, meta: { total_count: 2 }, success: true },
    isLoading: false,
    refetch: vi.fn(),
  } as unknown as ReturnType<typeof useGetApiV1AdminFinanceReconciliation>)
  vi.mocked(usePostApiV1AdminFinanceBankStatement).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof usePostApiV1AdminFinanceBankStatement>)
  vi.mocked(usePutApiV1AdminFinanceReconciliationIdMatch).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof usePutApiV1AdminFinanceReconciliationIdMatch>)
}

describe('BankReconciliation', () => {
  it('renders title and upload button', () => {
    setupMocks()
    renderWithProviders(<BankReconciliation />)

    expect(screen.getByText('Банковская сверка')).toBeInTheDocument()
    expect(screen.getByText('Загрузить выписку')).toBeInTheDocument()
  })

  it('renders bank entries table with data', () => {
    setupMocks()
    renderWithProviders(<BankReconciliation />)

    expect(screen.getByText('REF-001')).toBeInTheDocument()
    expect(screen.getByText('ООО Баня')).toBeInTheDocument()
    expect(screen.getByText('Ожидает')).toBeInTheDocument()
    expect(screen.getByText('Сопоставлено')).toBeInTheDocument()
  })

  it('renders match button for pending entries', () => {
    setupMocks()
    renderWithProviders(<BankReconciliation />)

    expect(screen.getByText('Связать')).toBeInTheDocument()
  })

  it('opens match modal on click', () => {
    setupMocks()
    renderWithProviders(<BankReconciliation />)

    fireEvent.click(screen.getByText('Связать'))
    expect(screen.getByText('Ручное сопоставление')).toBeInTheDocument()
    expect(screen.getByText('ID транзакции')).toBeInTheDocument()
    expect(screen.getByText('Тип транзакции')).toBeInTheDocument()
  })

  it('renders status filter', () => {
    setupMocks()
    renderWithProviders(<BankReconciliation />)

    expect(screen.getAllByText('Статус').length).toBeGreaterThanOrEqual(1)
  })

  it('renders empty state when no entries', () => {
    vi.mocked(useGetApiV1AdminFinanceReconciliation).mockReturnValue({
      data: { data: [], meta: { total_count: 0 }, success: true },
      isLoading: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetApiV1AdminFinanceReconciliation>)
    vi.mocked(usePostApiV1AdminFinanceBankStatement).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostApiV1AdminFinanceBankStatement>)
    vi.mocked(usePutApiV1AdminFinanceReconciliationIdMatch).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePutApiV1AdminFinanceReconciliationIdMatch>)

    renderWithProviders(<BankReconciliation />)
    expect(screen.getByText('Нет записей')).toBeInTheDocument()
  })
})
