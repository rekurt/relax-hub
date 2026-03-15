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
  {
    id: 'cert-3',
    code: 'BANI-CCCC-3333',
    amount: 100000,
    balance: 100000,
    status: 'expired',
    valid_until: '2025-06-01T00:00:00Z',
    created_at: '2024-06-01T00:00:00Z',
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

  it('renders page title', () => {
    setupMocks()
    renderWithProviders(<CertificateList />)
    expect(screen.getByText('Подарочные сертификаты')).toBeInTheDocument()
  })

  it('displays active certificates count', () => {
    setupMocks()
    renderWithProviders(<CertificateList />)
    expect(screen.getByText('Активных сертификатов')).toBeInTheDocument()
  })

  it('displays total balance of active certificates', () => {
    setupMocks()
    renderWithProviders(<CertificateList />)
    expect(screen.getByText('Общий баланс')).toBeInTheDocument()
    // 300000 kopecks = 3000 rubles - appears in both stat card and table
    const balanceElements = screen.getAllByText('3000 ₽')
    expect(balanceElements.length).toBeGreaterThanOrEqual(1)
  })

  it('displays total certificates count', () => {
    setupMocks()
    renderWithProviders(<CertificateList />)
    expect(screen.getByText('Всего сертификатов')).toBeInTheDocument()
  })

  it('renders certificate codes in table', () => {
    setupMocks()
    renderWithProviders(<CertificateList />)
    expect(screen.getByText('BANI-AAAA-1111')).toBeInTheDocument()
    expect(screen.getByText('BANI-BBBB-2222')).toBeInTheDocument()
    expect(screen.getByText('BANI-CCCC-3333')).toBeInTheDocument()
  })

  it('renders certificate status tags', () => {
    setupMocks()
    renderWithProviders(<CertificateList />)
    expect(screen.getByText('Активен')).toBeInTheDocument()
    expect(screen.getByText('Использован')).toBeInTheDocument()
    expect(screen.getByText('Истёк')).toBeInTheDocument()
  })

  it('shows balance for each certificate', () => {
    setupMocks()
    renderWithProviders(<CertificateList />)
    // cert-2 balance: 0 kopecks = 0 ₽
    expect(screen.getByText('0 ₽')).toBeInTheDocument()
    // cert-3 balance: 100000 kopecks = 1000 ₽ (may appear in both amount and balance columns)
    const balanceElements = screen.getAllByText('1000 ₽')
    expect(balanceElements.length).toBeGreaterThanOrEqual(1)
  })

  it('renders redeem code input section', () => {
    setupMocks()
    renderWithProviders(<CertificateList />)
    expect(screen.getByText('Активировать сертификат')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('BANI-XXXX-XXXX')).toBeInTheDocument()
    expect(screen.getByText('Проверить')).toBeInTheDocument()
    expect(screen.getByText('Активировать')).toBeInTheDocument()
  })

  it('calls redeem mutation when activate button clicked', async () => {
    setupMocks()
    renderWithProviders(<CertificateList />)

    const input = screen.getByPlaceholderText('BANI-XXXX-XXXX')
    fireEvent.change(input, { target: { value: 'BANI-TEST-CODE' } })
    fireEvent.click(screen.getByText('Активировать'))

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

  it('displays balance check result after checking code', async () => {
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

    // Type code and click check
    const input = screen.getByPlaceholderText('BANI-XXXX-XXXX')
    fireEvent.change(input, { target: { value: 'BANI-TEST-1234' } })
    fireEvent.click(screen.getByText('Проверить'))

    await waitFor(() => {
      const balanceElements = screen.getAllByText('2500 ₽')
      expect(balanceElements.length).toBeGreaterThanOrEqual(1)
    })
  })

  it('shows purchase button', () => {
    setupMocks()
    renderWithProviders(<CertificateList />)
    const buyButtons = screen.getAllByText('Купить сертификат')
    expect(buyButtons.length).toBeGreaterThanOrEqual(1)
  })

  it('shows empty state when no certificates', () => {
    setupMocks({ certificates: [] })
    renderWithProviders(<CertificateList />)
    expect(screen.getByText('У вас пока нет сертификатов')).toBeInTheDocument()
  })

  it('shows loading spinner', () => {
    setupMocks({ loading: true })
    renderWithProviders(<CertificateList />)
    expect(document.querySelector('.ant-spin-spinning')).toBeInTheDocument()
  })

  it('shows usage info alert', () => {
    setupMocks()
    renderWithProviders(<CertificateList />)
    expect(screen.getByText('Как использовать сертификат')).toBeInTheDocument()
  })

  it('renders amount column in table', () => {
    setupMocks()
    renderWithProviders(<CertificateList />)
    // cert-1: 500000 kopecks = 5000 rubles
    expect(screen.getByText('5000 ₽')).toBeInTheDocument()
    // cert-2: 200000 = 2000 rubles
    expect(screen.getByText('2000 ₽')).toBeInTheDocument()
  })
})
