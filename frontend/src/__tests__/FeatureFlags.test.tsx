import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import FeatureFlags from '@/pages/admin/FeatureFlags'

vi.mock('@/api/generated/admin/admin', () => ({
  useGetAdminFeatureFlags: vi.fn(),
  getGetAdminFeatureFlagsQueryKey: vi.fn(() => ['/admin/feature-flags']),
  usePutAdminFeatureFlagsKey: vi.fn(),
}))

import {
  useGetAdminFeatureFlags,
  usePutAdminFeatureFlagsKey,
} from '@/api/generated/admin/admin'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/feature-flags']}>
            <Routes>
              <Route path="/admin/feature-flags" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockFlags = [
  {
    key: 'wallet_enabled',
    enabled: true,
    description: 'Включить кошелёк',
    region: '',
    updated_at: '2026-03-20T10:00:00Z',
    updated_by: 'admin-1',
  },
  {
    key: 'phone_auth_enabled',
    enabled: false,
    description: 'Аутентификация по телефону',
    region: 'RU',
    updated_at: '2026-03-19T12:00:00Z',
    updated_by: 'admin-1',
  },
  {
    key: 'escrow_enabled',
    enabled: true,
    description: 'Эскроу для платежей',
    region: 'BY',
    updated_at: '2026-03-18T08:00:00Z',
    updated_by: 'admin-1',
  },
]

const mutationDefault = { mutateAsync: vi.fn(), isPending: false, variables: undefined }

describe('FeatureFlags', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useGetAdminFeatureFlags).mockReturnValue({
      data: { data: mockFlags, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminFeatureFlags>)
    vi.mocked(usePutAdminFeatureFlagsKey).mockReturnValue(
      mutationDefault as unknown as ReturnType<typeof usePutAdminFeatureFlagsKey>,
    )
  })

  it('renders feature flags table', () => {
    renderWithProviders(<FeatureFlags />)

    expect(screen.getByText('Функциональные флаги')).toBeInTheDocument()
    expect(screen.getByText('wallet_enabled')).toBeInTheDocument()
    expect(screen.getByText('phone_auth_enabled')).toBeInTheDocument()
    expect(screen.getByText('escrow_enabled')).toBeInTheDocument()
  })

  it('shows toggle switches for enabled/disabled', () => {
    renderWithProviders(<FeatureFlags />)

    const switches = screen.getAllByRole('switch')
    expect(switches).toHaveLength(3)
  })

  it('shows region tags', () => {
    renderWithProviders(<FeatureFlags />)

    expect(screen.getByText('Глобальный')).toBeInTheDocument()
    expect(screen.getByText('RU')).toBeInTheDocument()
    expect(screen.getByText('BY')).toBeInTheDocument()
  })

  it('renders correct number of rows', () => {
    renderWithProviders(<FeatureFlags />)

    const switches = screen.getAllByRole('switch')
    expect(switches).toHaveLength(3)
  })

  it('shows loading state', () => {
    vi.mocked(useGetAdminFeatureFlags).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetAdminFeatureFlags>)

    renderWithProviders(<FeatureFlags />)

    expect(screen.getByText('Функциональные флаги')).toBeInTheDocument()
  })

  it('has region filter dropdown', () => {
    renderWithProviders(<FeatureFlags />)

    expect(screen.getByText('Все регионы')).toBeInTheDocument()
  })
})
