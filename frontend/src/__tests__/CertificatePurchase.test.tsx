import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import CertificatePurchase from '@/pages/client/CertificatePurchase'

vi.mock('@/api/generated/certificates/certificates', () => ({
  usePostCertificatesPurchase: vi.fn(),
}))

import { usePostCertificatesPurchase } from '@/api/generated/certificates/certificates'

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

const mockMutate = vi.fn()

function setupMocks(overrides?: { isPending?: boolean }) {
  vi.mocked(usePostCertificatesPurchase).mockReturnValue({
    mutate: mockMutate,
    isPending: overrides?.isPending ?? false,
  } as unknown as ReturnType<typeof usePostCertificatesPurchase>)
}

describe('CertificatePurchase', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders page title', () => {
    setupMocks()
    renderWithProviders(<CertificatePurchase />)
    expect(screen.getByText('Купить подарочный сертификат')).toBeInTheDocument()
  })

  it('renders form fields', () => {
    setupMocks()
    renderWithProviders(<CertificatePurchase />)
    expect(screen.getByText('Сумма (в рублях)')).toBeInTheDocument()
    expect(screen.getByText('Ваш email')).toBeInTheDocument()
    expect(screen.getByText('Имя получателя')).toBeInTheDocument()
    expect(screen.getByText('Email получателя')).toBeInTheDocument()
    expect(screen.getByText('Сообщение')).toBeInTheDocument()
  })

  it('renders preset amount buttons', () => {
    setupMocks()
    renderWithProviders(<CertificatePurchase />)
    // Preset amounts: 1000, 2000, 3000, 5000 rubles
    expect(screen.getByText('1000 ₽')).toBeInTheDocument()
    expect(screen.getByText('2000 ₽')).toBeInTheDocument()
    expect(screen.getByText('3000 ₽')).toBeInTheDocument()
    expect(screen.getByText('5000 ₽')).toBeInTheDocument()
  })

  it('renders submit button', () => {
    setupMocks()
    renderWithProviders(<CertificatePurchase />)
    expect(screen.getByText('Купить сертификат')).toBeInTheDocument()
  })

  it('renders info alert about certificates', () => {
    setupMocks()
    renderWithProviders(<CertificatePurchase />)
    expect(screen.getByText('Подарочный сертификат')).toBeInTheDocument()
    expect(screen.getByText(/Срок действия — 365 дней/)).toBeInTheDocument()
  })

  it('renders back button', () => {
    setupMocks()
    renderWithProviders(<CertificatePurchase />)
    expect(screen.getByText('Назад к сертификатам')).toBeInTheDocument()
  })

  it('shows validation error when amount is missing', async () => {
    setupMocks()
    renderWithProviders(<CertificatePurchase />)

    fireEvent.click(screen.getByText('Купить сертификат'))

    await waitFor(() => {
      expect(screen.getByText('Укажите сумму')).toBeInTheDocument()
    })
  })

  it('shows validation error when email is missing', async () => {
    setupMocks()
    renderWithProviders(<CertificatePurchase />)

    fireEvent.click(screen.getByText('Купить сертификат'))

    await waitFor(() => {
      expect(screen.getByText('Укажите email')).toBeInTheDocument()
    })
  })

  it('submits form with correct data', async () => {
    let capturedOnSuccess: (() => void) | undefined
    mockMutate.mockImplementation((_data: unknown, opts: { onSuccess: () => void }) => {
      capturedOnSuccess = opts.onSuccess
    })
    setupMocks()
    renderWithProviders(<CertificatePurchase />)

    // Fill in amount
    const amountInput = screen.getByPlaceholderText('Введите сумму')
    fireEvent.change(amountInput, { target: { value: '3000' } })

    // Fill in email
    const emailInput = screen.getByPlaceholderText('your@email.com')
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } })

    // Fill in recipient name
    const recipientNameInput = screen.getByPlaceholderText('Имя получателя')
    fireEvent.change(recipientNameInput, { target: { value: 'Иван' } })

    // Submit the form
    fireEvent.click(screen.getByText('Купить сертификат'))

    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        {
          data: {
            amount: 300000, // 3000 rubles -> 300000 kopecks
            purchaser_email: 'test@example.com',
            recipient_name: 'Иван',
            recipient_email: undefined,
            message: undefined,
          },
        },
        expect.objectContaining({
          onSuccess: expect.any(Function),
          onError: expect.any(Function),
        }),
      )
    })

    // Simulate success to show the result page
    if (capturedOnSuccess) {
      // Can't easily trigger success state from here since we need the response.
      // The mutation was called correctly - that's the important assertion.
    }
  })

  it('shows success result after purchase', async () => {
    mockMutate.mockImplementation((_data: unknown, opts: { onSuccess: (r: unknown) => void }) => {
      opts.onSuccess({
        data: {
          id: 'cert-new',
          code: 'BANI-NEW1-CODE',
          amount: 300000,
          balance: 300000,
          status: 'active',
          valid_until: '2027-03-16T00:00:00Z',
          recipient_name: 'Иван',
          message: 'С праздником!',
        },
      })
    })
    setupMocks()
    renderWithProviders(<CertificatePurchase />)

    // Fill in required fields
    fireEvent.change(screen.getByPlaceholderText('Введите сумму'), { target: { value: '3000' } })
    fireEvent.change(screen.getByPlaceholderText('your@email.com'), { target: { value: 'test@example.com' } })
    fireEvent.click(screen.getByText('Купить сертификат'))

    await waitFor(() => {
      expect(screen.getByText('Сертификат создан!')).toBeInTheDocument()
    })
    expect(screen.getByText('BANI-NEW1-CODE')).toBeInTheDocument()
    expect(screen.getByText('3000 ₽')).toBeInTheDocument()
    expect(screen.getByText('Иван')).toBeInTheDocument()
    expect(screen.getByText('С праздником!')).toBeInTheDocument()
    expect(screen.getByText('Мои сертификаты')).toBeInTheDocument()
    expect(screen.getByText('Купить ещё')).toBeInTheDocument()
  })

  it('shows loading state while submitting', () => {
    setupMocks({ isPending: true })
    renderWithProviders(<CertificatePurchase />)
    const submitButton = screen.getByText('Купить сертификат').closest('button')
    expect(submitButton).toHaveClass('ant-btn-loading')
  })

  it('sets amount when preset button is clicked', () => {
    setupMocks()
    renderWithProviders(<CertificatePurchase />)

    fireEvent.click(screen.getByText('3000 ₽'))

    const amountInput = screen.getByPlaceholderText('Введите сумму') as HTMLInputElement
    expect(amountInput.value).toBe('3000')
  })
})
