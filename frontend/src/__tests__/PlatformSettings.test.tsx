import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import PlatformSettings from '@/pages/admin/PlatformSettings'

vi.mock('@/api/generated/admin/admin', () => ({
  useGetAdminSettings: vi.fn(),
  getGetAdminSettingsQueryKey: vi.fn(() => ['/admin/settings']),
  usePutAdminSettingsKey: vi.fn(),
}))

import {
  useGetAdminSettings,
  usePutAdminSettingsKey,
} from '@/api/generated/admin/admin'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/settings']}>
            <Routes>
              <Route path="/admin/settings" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockSettings = [
  {
    key: 'service_fee_percent',
    value: '10',
    type: 'int',
    description: 'Процент комиссии платформы',
    updated_at: '2026-03-20T10:00:00Z',
    updated_by: 'admin-1',
  },
  {
    key: 'welcome_bonus_amount',
    value: '50000',
    type: 'int',
    description: 'Приветственный бонус в копейках',
    updated_at: '2026-03-19T12:00:00Z',
    updated_by: 'admin-1',
  },
  {
    key: 'escrow_enabled',
    value: 'true',
    type: 'bool',
    description: 'Включить эскроу',
    updated_at: '2026-03-18T08:00:00Z',
    updated_by: 'admin-1',
  },
]

const mutationDefault = { mutateAsync: vi.fn(), isPending: false, variables: undefined }

describe('PlatformSettings', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useGetAdminSettings).mockReturnValue({
      data: { data: mockSettings, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminSettings>)
    vi.mocked(usePutAdminSettingsKey).mockReturnValue(
      mutationDefault as unknown as ReturnType<typeof usePutAdminSettingsKey>,
    )
  })

  it('renders settings table with data', () => {
    renderWithProviders(<PlatformSettings />)

    expect(screen.getByText('Настройки платформы')).toBeInTheDocument()
    expect(screen.getByText('service_fee_percent')).toBeInTheDocument()
    expect(screen.getByText('welcome_bonus_amount')).toBeInTheDocument()
    expect(screen.getByText('escrow_enabled')).toBeInTheDocument()
  })

  it('shows type tags correctly', () => {
    renderWithProviders(<PlatformSettings />)

    const intTags = screen.getAllByText('Целое число')
    expect(intTags).toHaveLength(2)
    expect(screen.getByText('Логическое')).toBeInTheDocument()
  })

  it('shows bool value as colored tag', () => {
    renderWithProviders(<PlatformSettings />)

    expect(screen.getByText('Да')).toBeInTheDocument()
  })

  it('opens edit modal when edit button clicked', () => {
    renderWithProviders(<PlatformSettings />)

    const editButtons = screen.getAllByText('Изменить')
    fireEvent.click(editButtons[0])

    expect(screen.getByText('Изменить: service_fee_percent')).toBeInTheDocument()
    expect(screen.getByText('Процент комиссии платформы')).toBeInTheDocument()
  })

  it('shows loading state', () => {
    vi.mocked(useGetAdminSettings).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetAdminSettings>)

    renderWithProviders(<PlatformSettings />)

    expect(screen.getByText('Настройки платформы')).toBeInTheDocument()
  })

  it('calls update mutation when saving', async () => {
    const mutateAsync = vi.fn().mockResolvedValue({})
    vi.mocked(usePutAdminSettingsKey).mockReturnValue({
      mutateAsync,
      isPending: false,
      variables: undefined,
    } as unknown as ReturnType<typeof usePutAdminSettingsKey>)

    renderWithProviders(<PlatformSettings />)

    const editButtons = screen.getAllByText('Изменить')
    fireEvent.click(editButtons[0])

    await waitFor(() => {
      expect(screen.getByText('Сохранить')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Сохранить'))

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalledWith({
        key: 'service_fee_percent',
        data: { value: '10' },
      })
    })
  })
})
