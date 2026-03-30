import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ActivePromoCodes from '@/pages/client/ActivePromoCodes'

vi.mock('@/api/generated/promo-codes/promo-codes', () => ({
  usePostPromoCodesValidate: vi.fn(),
}))

import { usePostPromoCodesValidate } from '@/api/generated/promo-codes/promo-codes'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/client/promos']}>
            <Routes>
              <Route path="/client/promos" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockValidateMutate = vi.fn()

function setupMocks(overrides?: { isPending?: boolean; isError?: boolean }) {
  vi.mocked(usePostPromoCodesValidate).mockReturnValue({
    mutate: mockValidateMutate,
    isPending: overrides?.isPending ?? false,
    isError: overrides?.isError ?? false,
  } as unknown as ReturnType<typeof usePostPromoCodesValidate>)
}

describe('ActivePromoCodes', () => {
  beforeEach(() => {
    mockValidateMutate.mockReset()
    localStorage.removeItem('bani_validated_promos')
  })

  it('renders page title and input', () => {
    setupMocks()
    renderWithProviders(<ActivePromoCodes />)

    expect(screen.getByText(/\u041f\u0440\u043e\u043c\u043e\u043a\u043e\u0434\u044b/)).toBeInTheDocument()
    expect(screen.getByPlaceholderText(/\u0412\u0432\u0435\u0434\u0438\u0442\u0435 \u043f\u0440\u043e\u043c\u043e\u043a\u043e\u0434/)).toBeInTheDocument()
  })

  it('calls validate mutation on button click', () => {
    setupMocks()
    renderWithProviders(<ActivePromoCodes />)

    const input = screen.getByPlaceholderText(/\u0412\u0432\u0435\u0434\u0438\u0442\u0435 \u043f\u0440\u043e\u043c\u043e\u043a\u043e\u0434/)
    fireEvent.change(input, { target: { value: 'SUMMER2026' } })
    fireEvent.click(screen.getByRole('button', { name: /\u041f\u0440\u043e\u0432\u0435\u0440\u0438\u0442\u044c/ }))

    expect(mockValidateMutate).toHaveBeenCalledWith({
      data: { code: 'SUMMER2026' },
    })
  })

  it('does not call validate with empty code', () => {
    setupMocks()
    renderWithProviders(<ActivePromoCodes />)

    fireEvent.click(screen.getByRole('button', { name: /\u041f\u0440\u043e\u0432\u0435\u0440\u0438\u0442\u044c/ }))

    expect(mockValidateMutate).not.toHaveBeenCalled()
  })

  it('converts input to uppercase', () => {
    setupMocks()
    renderWithProviders(<ActivePromoCodes />)

    const input = screen.getByPlaceholderText(/\u0412\u0432\u0435\u0434\u0438\u0442\u0435 \u043f\u0440\u043e\u043c\u043e\u043a\u043e\u0434/)
    fireEvent.change(input, { target: { value: 'summer' } })

    expect(input).toHaveValue('SUMMER')
  })

  it('shows error alert when validation fails', () => {
    setupMocks({ isError: true })
    renderWithProviders(<ActivePromoCodes />)

    expect(screen.getByText(/\u043d\u0435\u0434\u0435\u0439\u0441\u0442\u0432\u0438\u0442\u0435\u043b\u0435\u043d/)).toBeInTheDocument()
  })

  it('shows empty state when no validated promos', () => {
    setupMocks()
    renderWithProviders(<ActivePromoCodes />)

    expect(screen.getByText(/\u041d\u0435\u0442 \u043f\u0440\u043e\u0432\u0435\u0440\u0435\u043d\u043d\u044b\u0445 \u043f\u0440\u043e\u043c\u043e\u043a\u043e\u0434\u043e\u0432/)).toBeInTheDocument()
  })

  it('renders previously validated promos from localStorage', () => {
    const storedPromos = [
      {
        code: 'SAVE10',
        type: 'percentage',
        value: 10,
        discount: 50000,
        validatedAt: '2026-03-28T12:00:00Z',
      },
      {
        code: 'FLAT500',
        type: 'fixed_amount',
        value: 50000,
        discount: 50000,
        validatedAt: '2026-03-27T10:00:00Z',
      },
    ]
    localStorage.setItem('bani_validated_promos', JSON.stringify(storedPromos))

    setupMocks()
    renderWithProviders(<ActivePromoCodes />)

    expect(screen.getByText('SAVE10')).toBeInTheDocument()
    expect(screen.getByText('FLAT500')).toBeInTheDocument()
  })

  it('shows discount display for percentage type', () => {
    localStorage.setItem(
      'bani_validated_promos',
      JSON.stringify([
        { code: 'PCT20', type: 'percentage', value: 20, validatedAt: '2026-03-28T12:00:00Z' },
      ]),
    )

    setupMocks()
    renderWithProviders(<ActivePromoCodes />)

    expect(screen.getByText('20%')).toBeInTheDocument()
  })

  it('shows loading state when validating', () => {
    setupMocks({ isPending: true })
    renderWithProviders(<ActivePromoCodes />)

    const button = screen.getByRole('button', { name: /\u041f\u0440\u043e\u0432\u0435\u0440\u0438\u0442\u044c/ })
    expect(button.classList.contains('ant-btn-loading')).toBe(true)
  })

  it('shows copy and remove buttons for validated promos', () => {
    localStorage.setItem(
      'bani_validated_promos',
      JSON.stringify([
        { code: 'TEST1', type: 'percentage', value: 5, validatedAt: '2026-03-28T12:00:00Z' },
      ]),
    )

    setupMocks()
    renderWithProviders(<ActivePromoCodes />)

    expect(screen.getByText(/\u041a\u043e\u043f\u0438\u0440\u043e\u0432\u0430\u0442\u044c/)).toBeInTheDocument()
    expect(screen.getByText(/\u0423\u0431\u0440\u0430\u0442\u044c/)).toBeInTheDocument()
  })

  it('removes promo from list on remove click', async () => {
    localStorage.setItem(
      'bani_validated_promos',
      JSON.stringify([
        { code: 'REMOVE_ME', type: 'percentage', value: 5, validatedAt: '2026-03-28T12:00:00Z' },
      ]),
    )

    setupMocks()
    renderWithProviders(<ActivePromoCodes />)

    expect(screen.getByText('REMOVE_ME')).toBeInTheDocument()

    fireEvent.click(screen.getByText(/\u0423\u0431\u0440\u0430\u0442\u044c/))

    await waitFor(() => {
      expect(screen.queryByText('REMOVE_ME')).not.toBeInTheDocument()
    })
  })
})
