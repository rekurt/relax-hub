import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import WalletManagement from '@/pages/admin/WalletManagement'

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

import { axiosInstance } from '@/api/axios-instance'

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

const mockWallet = {
  id: 'wallet-uuid-1234',
  user_id: 'user-uuid-5678',
  balance: 1500000,
  held_amount: 200000,
  currency: 'RUB',
  status: 'active',
}

describe('WalletManagement', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders page title and search input', () => {
    renderWithProviders(<WalletManagement />)

    expect(screen.getByText('Управление кошельками')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('ID кошелька (UUID)')).toBeInTheDocument()
    expect(screen.getByText('Найти')).toBeInTheDocument()
  })

  it('shows empty state before search', () => {
    renderWithProviders(<WalletManagement />)

    expect(
      screen.getByText(
        'Введите ID кошелька для поиска. Вы сможете зачислить, списать средства или заморозить кошелёк.',
      ),
    ).toBeInTheDocument()
  })

  it('displays wallet info after successful search', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { data: mockWallet, success: true },
    })

    renderWithProviders(<WalletManagement />)

    const input = screen.getByPlaceholderText('ID кошелька (UUID)')
    fireEvent.change(input, { target: { value: 'wallet-uuid-1234' } })
    fireEvent.click(screen.getByText('Найти'))

    await waitFor(() => {
      expect(screen.getByText('Информация о кошельке')).toBeInTheDocument()
    })

    expect(screen.getByText('15000 ₽')).toBeInTheDocument()
    expect(screen.getByText('2000 ₽')).toBeInTheDocument()
    expect(screen.getByText('13000 ₽')).toBeInTheDocument()
    expect(screen.getByText('Активен')).toBeInTheDocument()
  })

  it('shows action buttons for active wallet', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { data: mockWallet, success: true },
    })

    renderWithProviders(<WalletManagement />)

    const input = screen.getByPlaceholderText('ID кошелька (UUID)')
    fireEvent.change(input, { target: { value: 'wallet-uuid-1234' } })
    fireEvent.click(screen.getByText('Найти'))

    await waitFor(() => {
      expect(screen.getByText('Информация о кошельке')).toBeInTheDocument()
    })

    expect(screen.getByText('Зачислить')).toBeInTheDocument()
    expect(screen.getByText('Списать')).toBeInTheDocument()
    expect(screen.getByText('Заморозить')).toBeInTheDocument()
  })

  it('shows unfreeze button for frozen wallet', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { data: { ...mockWallet, status: 'frozen' }, success: true },
    })

    renderWithProviders(<WalletManagement />)

    const input = screen.getByPlaceholderText('ID кошелька (UUID)')
    fireEvent.change(input, { target: { value: 'wallet-uuid-1234' } })
    fireEvent.click(screen.getByText('Найти'))

    await waitFor(() => {
      expect(screen.getByText('Заморожен')).toBeInTheDocument()
    })

    expect(screen.getByText('Разморозить')).toBeInTheDocument()
    expect(screen.queryByText('Заморозить')).not.toBeInTheDocument()
  })

  it('opens credit modal with amount and reason fields', async () => {
    vi.mocked(axiosInstance.get).mockResolvedValue({
      data: { data: mockWallet, success: true },
    })

    renderWithProviders(<WalletManagement />)

    const input = screen.getByPlaceholderText('ID кошелька (UUID)')
    fireEvent.change(input, { target: { value: 'wallet-uuid-1234' } })
    fireEvent.click(screen.getByText('Найти'))

    await waitFor(() => {
      expect(screen.getByText('Зачислить')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Зачислить'))

    await waitFor(() => {
      expect(screen.getByText('Зачисление средств')).toBeInTheDocument()
      expect(screen.getByText('Сумма (в рублях)')).toBeInTheDocument()
      expect(screen.getByText('Причина')).toBeInTheDocument()
    })
  })
})
