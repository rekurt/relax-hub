import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import CertificateList from '@/pages/client/CertificateList'

vi.mock('@/api/generated/certificates/certificates', () => ({
  useGetMyCertificates: vi.fn(),
  usePostCertificatesRedeem: vi.fn(),
  useGetCertificatesCodeBalance: vi.fn(),
}))

import {
  useGetMyCertificates,
  usePostCertificatesRedeem,
  useGetCertificatesCodeBalance,
} from '@/api/generated/certificates/certificates'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter>{ui}</MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockCertificates = [
  {
    id: 'cert-1',
    code: 'BANI-AAAA-1111',
    amount: 500000,
    balance: 300000,
    status: 'active',
    valid_until: '2027-01-15T00:00:00Z',
    created_at: '2026-01-15T00:00:00Z',
    purchaser_email: 'buyer@test.com',
    recipient_email: 'recipient@test.com',
    recipient_name: 'Иван',
    message: 'С днём рождения!',
  },
  {
    id: 'cert-2',
    code: 'BANI-BBBB-2222',
    amount: 200000,
    balance: 0,
    status: 'used',
    valid_until: '2026-12-01T00:00:00Z',
    created_at: '2025-12-01T00:00:00Z',
    purchaser_email: 'buyer@test.com',
  },
]

const mockRedeemMutate = vi.fn()

function setupMocks(overrides?: {
  certificates?: typeof mockCertificates
  loading?: boolean
  balanceData?: { code: string; amount: number; balance: number; status: string; valid_until: string } | null
}) {
  vi.mocked(useGetMyCertificates).mockReturnValue({
    data: {
      data: overrides?.certificates ?? mockCertificates,
      success: true,
      meta: { total_count: (overrides?.certificates ?? mockCertificates).length, page: 1, page_size: 10, total_pages: 1 },
    },
    isLoading: overrides?.loading ?? false,
  } as unknown as ReturnType<typeof useGetMyCertificates>)

  vi.mocked(usePostCertificatesRedeem).mockReturnValue({
    mutate: mockRedeemMutate,
    isPending: false,
  } as unknown as ReturnType<typeof usePostCertificatesRedeem>)

  vi.mocked(useGetCertificatesCodeBalance).mockReturnValue({
    data: overrides?.balanceData !== undefined
      ? overrides.balanceData === null
        ? undefined
        : { data: overrides.balanceData, success: true }
      : undefined,
    isLoading: false,
  } as unknown as ReturnType<typeof useGetCertificatesCodeBalance>)
}

describe('CertificateList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders redesigned hero and overview metrics', () => {
    setupMocks()
    renderWithProviders(<CertificateList />)

    expect(screen.getByText('Мои сертификаты')).toBeInTheDocument()
    expect(screen.getByText('Активный баланс')).toBeInTheDocument()
    expect(screen.getByText('Активных сертификатов')).toBeInTheDocument()
    expect(screen.getByText('Как использовать сертификат')).toBeInTheDocument()
  })

  it('renders certificate cards instead of a plain table-only view', () => {
    setupMocks()
    renderWithProviders(<CertificateList />)

    expect(screen.getByText('BANI-AAAA-1111')).toBeInTheDocument()
    expect(screen.getByText('BANI-BBBB-2222')).toBeInTheDocument()
    expect(screen.getByText('Активен')).toBeInTheDocument()
    expect(screen.getByText('Использован')).toBeInTheDocument()
  })

  it('redeems certificate code from activation block', async () => {
    setupMocks()
    renderWithProviders(<CertificateList />)

    fireEvent.change(screen.getByPlaceholderText('BANI-XXXX-XXXX'), { target: { value: 'BANI-TEST-CODE' } })
    fireEvent.click(screen.getByText('Активировать сертификат'))

    await waitFor(() => {
      expect(mockRedeemMutate).toHaveBeenCalledWith(
        { data: { code: 'BANI-TEST-CODE' } },
        expect.objectContaining({
          onSuccess: expect.any(Function),
          onError: expect.any(Function),
        }),
      )
    })
  })

  it('shows balance check panel after code verification', async () => {
    setupMocks({
      balanceData: {
        code: 'BANI-TEST-1234',
        amount: 500000,
        balance: 250000,
        status: 'active',
        valid_until: '2027-06-01T00:00:00Z',
      },
    })
    renderWithProviders(<CertificateList />)

    fireEvent.change(screen.getByPlaceholderText('BANI-XXXX-XXXX'), { target: { value: 'BANI-TEST-1234' } })
    fireEvent.click(screen.getByText('Проверить баланс'))

    await waitFor(() => {
      expect(screen.getByText('BANI-TEST-1234')).toBeInTheDocument()
    })
    expect(screen.getAllByText('2500 ₽').length).toBeGreaterThanOrEqual(1)
  })

  it('shows empty state with purchase CTA when user has no certificates', () => {
    setupMocks({ certificates: [] })
    renderWithProviders(<CertificateList />)

    expect(screen.getByText('У вас пока нет сертификатов')).toBeInTheDocument()
    expect(screen.getByText('Перейти к покупке')).toBeInTheDocument()
  })
})
