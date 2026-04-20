import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import CertificatePurchase from '@/pages/client/CertificatePurchase'

vi.mock('@/api/generated/certificates/certificates', () => ({
  usePostCertificatesOrders: vi.fn(),
  usePostCertificatesOrdersIdPay: vi.fn(),
  useGetCertificatesOrdersId: vi.fn(),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn((selector: (s: { isAuthenticated: boolean }) => boolean) =>
    selector({ isAuthenticated: true }),
  ),
}))

vi.mock('@/components/ApplePayButton', () => ({
  default: ({ onToken }: { onToken: (token: string) => void }) => (
    <button type="button" onClick={() => onToken('apple-token')}>
      Apple Pay Mock
    </button>
  ),
}))

vi.mock('@/components/GooglePayButton', () => ({
  default: ({ onToken }: { onToken: (token: string) => void }) => (
    <button type="button" onClick={() => onToken('google-token')}>
      Google Pay Mock
    </button>
  ),
}))

import {
  useGetCertificatesOrdersId,
  usePostCertificatesOrders,
  usePostCertificatesOrdersIdPay,
} from '@/api/generated/certificates/certificates'

function renderWithProviders(ui: React.ReactElement, route = '/certificates') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={[route]}>{ui}</MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockCreateOrderMutateAsync = vi.fn()
const mockPayOrderMutateAsync = vi.fn()

function setupMocks(overrides?: {
  orderData?: unknown
  createPending?: boolean
  payPending?: boolean
}) {
  vi.mocked(usePostCertificatesOrders).mockReturnValue({
    mutateAsync: mockCreateOrderMutateAsync,
    isPending: overrides?.createPending ?? false,
  } as unknown as ReturnType<typeof usePostCertificatesOrders>)

  vi.mocked(usePostCertificatesOrdersIdPay).mockReturnValue({
    mutateAsync: mockPayOrderMutateAsync,
    isPending: overrides?.payPending ?? false,
  } as unknown as ReturnType<typeof usePostCertificatesOrdersIdPay>)

  vi.mocked(useGetCertificatesOrdersId).mockReturnValue({
    data: overrides?.orderData,
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
  } as unknown as ReturnType<typeof useGetCertificatesOrdersId>)
}

describe('CertificatePurchase', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockCreateOrderMutateAsync.mockReset()
    mockPayOrderMutateAsync.mockReset()
  })

  it('renders premium hero, trust copy and checkout workspace', () => {
    setupMocks()
    renderWithProviders(<CertificatePurchase />)

    expect(screen.getByText('Подарочный сертификат BANI')).toBeInTheDocument()
    expect(screen.getByText(/Оплатите подарок один раз/i)).toBeInTheDocument()
    expect(screen.getByText('Срок действия')).toBeInTheDocument()
    expect(screen.getByText('Доставка')).toBeInTheDocument()
    expect(screen.getByText('Способ оплаты')).toBeInTheDocument()
    expect(screen.getByText('Превью сертификата')).toBeInTheDocument()
  })

  it('updates amount summary when preset is selected', () => {
    setupMocks()
    renderWithProviders(<CertificatePurchase />)

    fireEvent.click(screen.getByText('5000 ₽'))

    expect(screen.getByDisplayValue('5000')).toBeInTheDocument()
    expect(screen.getAllByText('5000 ₽').length).toBeGreaterThanOrEqual(1)
  })

  it('creates order and initiates card payment with order payload', async () => {
    mockCreateOrderMutateAsync.mockResolvedValue({
      data: {
        id: 'order-1',
        amount: 300000,
        status: 'draft',
      },
    })
    mockPayOrderMutateAsync.mockResolvedValue({
      data: {
        confirmation_url: 'https://pay.example/redirect',
      },
    })
    setupMocks()
    renderWithProviders(<CertificatePurchase />)

    fireEvent.change(screen.getByPlaceholderText('Введите сумму'), { target: { value: '3000' } })
    fireEvent.change(screen.getByPlaceholderText('your@email.com'), { target: { value: 'test@example.com' } })
    fireEvent.change(screen.getByPlaceholderText('Имя получателя'), { target: { value: 'Иван' } })
    fireEvent.click(screen.getByText('Карта'))
    fireEvent.click(screen.getByText('Перейти к оплате'))

    await waitFor(() => {
      expect(mockCreateOrderMutateAsync).toHaveBeenCalledWith({
        data: {
          amount: 300000,
          purchaser_email: 'test@example.com',
          recipient_name: 'Иван',
          recipient_email: undefined,
          message: undefined,
        },
      })
    })

    expect(mockPayOrderMutateAsync).toHaveBeenCalledWith({
      id: 'order-1',
      data: {
        payment_method: 'card',
      },
    })
  })

  it('initiates apple pay with token-based payload', async () => {
    mockCreateOrderMutateAsync.mockResolvedValue({
      data: {
        id: 'order-apple',
        amount: 500000,
        status: 'draft',
      },
    })
    mockPayOrderMutateAsync.mockResolvedValue({
      data: {
        confirmation_url: '',
      },
    })
    setupMocks()
    renderWithProviders(<CertificatePurchase />)

    fireEvent.change(screen.getByPlaceholderText('Введите сумму'), { target: { value: '5000' } })
    fireEvent.change(screen.getByPlaceholderText('your@email.com'), { target: { value: 'apple@example.com' } })
    fireEvent.click(screen.getByText('Apple Pay'))
    fireEvent.click(screen.getByText('Apple Pay Mock'))

    await waitFor(() => {
      expect(mockCreateOrderMutateAsync).toHaveBeenCalled()
    })

    expect(mockPayOrderMutateAsync).toHaveBeenCalledWith({
      id: 'order-apple',
      data: {
        payment_method: 'apple_pay',
        payment_token: 'apple-token',
      },
    })
  })

  it('shows branded success state when order is already paid', async () => {
    setupMocks({
      orderData: {
        data: {
          id: 'order-paid',
          amount: 300000,
          status: 'paid',
          recipient_name: 'Иван',
          purchaser_email: 'buyer@example.com',
          certificate: {
            id: 'cert-1',
            code: 'BANI-PAID-0001',
            amount: 300000,
            balance: 300000,
            status: 'active',
            valid_until: '2027-05-10T00:00:00Z',
          },
        },
      },
    })

    renderWithProviders(<CertificatePurchase />, '/certificates?order_id=order-paid')

    await waitFor(() => {
      expect(screen.getByText('Сертификат оплачен')).toBeInTheDocument()
    })
    expect(screen.getByText('BANI-PAID-0001')).toBeInTheDocument()
    expect(screen.getByText(/Мы отправили подтверждение/i)).toBeInTheDocument()
  })

  it('shows processing state when order is waiting for payment confirmation', async () => {
    setupMocks({
      orderData: {
        data: {
          id: 'order-pending',
          amount: 300000,
          status: 'pending_payment',
        },
      },
    })

    renderWithProviders(<CertificatePurchase />, '/certificates?order_id=order-pending')

    await waitFor(() => {
      expect(screen.getByText('Платёж обрабатывается')).toBeInTheDocument()
    })
    expect(screen.getByText(/Подтверждаем оплату и выпуск сертификата/i)).toBeInTheDocument()
  })
})
