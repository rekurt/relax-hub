import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ServiceFeeConfig from '@/pages/admin/ServiceFeeConfig'

vi.mock('@/api/generated/admin/admin', () => ({
  useGetAdminServiceFee: vi.fn(),
  getGetAdminServiceFeeQueryKey: vi.fn(() => ['/admin/service-fee']),
  usePutAdminServiceFee: vi.fn(),
}))

import {
  useGetAdminServiceFee,
  usePutAdminServiceFee,
} from '@/api/generated/admin/admin'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/service-fees']}>
            <Routes>
              <Route path="/admin/service-fees" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockConfigs = [
  {
    id: 'fee-1',
    region: '*',
    category: '',
    fee_percent: 10,
    created_at: '2026-03-15T10:00:00Z',
    updated_at: '2026-03-20T10:00:00Z',
  },
  {
    id: 'fee-2',
    region: 'RU',
    category: 'premium',
    fee_percent: 8,
    created_at: '2026-03-16T10:00:00Z',
    updated_at: '2026-03-19T12:00:00Z',
  },
  {
    id: 'fee-3',
    region: 'BY',
    category: '',
    fee_percent: 12,
    created_at: '2026-03-17T10:00:00Z',
    updated_at: '2026-03-18T08:00:00Z',
  },
]

const mutationDefault = { mutateAsync: vi.fn(), isPending: false }

describe('ServiceFeeConfig', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useGetAdminServiceFee).mockReturnValue({
      data: { data: mockConfigs, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminServiceFee>)
    vi.mocked(usePutAdminServiceFee).mockReturnValue(
      mutationDefault as unknown as ReturnType<typeof usePutAdminServiceFee>,
    )
  })

  it('renders service fee table with data', () => {
    renderWithProviders(<ServiceFeeConfig />)

    expect(screen.getByText('Комиссия платформы')).toBeInTheDocument()
    expect(screen.getByText('10%')).toBeInTheDocument()
    expect(screen.getByText('8%')).toBeInTheDocument()
    expect(screen.getByText('12%')).toBeInTheDocument()
  })

  it('shows region tags', () => {
    renderWithProviders(<ServiceFeeConfig />)

    expect(screen.getByText('Глобальный')).toBeInTheDocument()
    expect(screen.getByText('RU')).toBeInTheDocument()
    expect(screen.getByText('BY')).toBeInTheDocument()
  })

  it('shows category or "Все категории" tag', () => {
    renderWithProviders(<ServiceFeeConfig />)

    expect(screen.getByText('premium')).toBeInTheDocument()
    const allCategoryTags = screen.getAllByText('Все категории')
    expect(allCategoryTags.length).toBeGreaterThanOrEqual(2)
  })

  it('opens create modal when add button clicked', () => {
    renderWithProviders(<ServiceFeeConfig />)

    fireEvent.click(screen.getByText('Добавить'))

    expect(screen.getByText('Добавить комиссию')).toBeInTheDocument()
  })

  it('opens edit modal when edit button clicked', () => {
    renderWithProviders(<ServiceFeeConfig />)

    const editButtons = screen.getAllByText('Изменить')
    fireEvent.click(editButtons[0])

    expect(screen.getByText('Редактировать комиссию')).toBeInTheDocument()
  })

  it('shows empty state when no configs', () => {
    vi.mocked(useGetAdminServiceFee).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminServiceFee>)

    renderWithProviders(<ServiceFeeConfig />)

    expect(screen.getByText('Нет настроек комиссии')).toBeInTheDocument()
  })

  it('shows loading state', () => {
    vi.mocked(useGetAdminServiceFee).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetAdminServiceFee>)

    renderWithProviders(<ServiceFeeConfig />)

    expect(screen.getByText('Комиссия платформы')).toBeInTheDocument()
  })

  it('calls upsert mutation when saving new config', async () => {
    const mutateAsync = vi.fn().mockResolvedValue({})
    vi.mocked(usePutAdminServiceFee).mockReturnValue({
      mutateAsync,
      isPending: false,
    } as unknown as ReturnType<typeof usePutAdminServiceFee>)

    renderWithProviders(<ServiceFeeConfig />)

    fireEvent.click(screen.getByText('Добавить'))

    await waitFor(() => {
      expect(screen.getByText('Добавить комиссию')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Создать'))

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalled()
    })
  })
})
