import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import BathhouseList from '@/pages/bathhouses/BathhouseList'
import type { AuthState } from '@/stores/auth'

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetMyBathhouses: vi.fn(),
  useDeleteBathhousesId: vi.fn(),
  usePostMyBathhousesIdDuplicate: vi.fn(),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn(),
}))

import { useGetMyBathhouses, useDeleteBathhousesId, usePostMyBathhousesIdDuplicate } from '@/api/generated/bathhouses/bathhouses'
import { useAuthStore } from '@/stores/auth'

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

const mockBathhouses = [
  {
    id: '1',
    name: 'Баня на Пушкина',
    address: 'ул. Пушкина, д. 10',
    status: 'active',
    rating: 4.5,
    review_count: 12,
    price_per_hour: 150000,
  },
  {
    id: '2',
    name: 'Баня Люкс',
    address: 'ул. Ленина, д. 5',
    status: 'pending',
    rating: 0,
    review_count: 0,
    price_per_hour: 300000,
  },
]

function mockAuth(role: string) {
  vi.mocked(useAuthStore).mockImplementation((selector) =>
    (selector as (state: AuthState) => unknown)({
      user: { role },
      token: 'test',
      isLoading: false,
      isAuthenticated: true,
      setAuth: vi.fn(),
      logout: vi.fn(),
      loadProfile: vi.fn(),
    } as unknown as AuthState),
  )
}

describe('BathhouseList', () => {
  beforeEach(() => {
    vi.mocked(useDeleteBathhousesId).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof useDeleteBathhousesId>)
    vi.mocked(usePostMyBathhousesIdDuplicate).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostMyBathhousesIdDuplicate>)
  })

  it('renders bathhouse table with data', () => {
    vi.mocked(useGetMyBathhouses).mockReturnValue({
      data: { data: mockBathhouses, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhouses>)
    mockAuth('owner')

    renderWithProviders(<BathhouseList />)

    expect(screen.getByText('Мои бани')).toBeInTheDocument()
    expect(screen.getByText('Баня на Пушкина')).toBeInTheDocument()
    expect(screen.getByText('Баня Люкс')).toBeInTheDocument()
    expect(screen.getByText('Активна')).toBeInTheDocument()
    expect(screen.getByText('На модерации')).toBeInTheDocument()
  })

  it('shows create button for owner role', () => {
    vi.mocked(useGetMyBathhouses).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhouses>)
    mockAuth('owner')

    renderWithProviders(<BathhouseList />)

    expect(screen.getAllByText('Добавить баню').length).toBeGreaterThanOrEqual(1)
  })

  it('hides create button for representative role', () => {
    vi.mocked(useGetMyBathhouses).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhouses>)
    mockAuth('representative')

    renderWithProviders(<BathhouseList />)

    expect(screen.queryByText('Добавить баню')).not.toBeInTheDocument()
  })

  it('shows loading state', () => {
    vi.mocked(useGetMyBathhouses).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyBathhouses>)
    mockAuth('owner')

    renderWithProviders(<BathhouseList />)

    expect(screen.getByText('Мои бани')).toBeInTheDocument()
  })

  it('shows empty state when no bathhouses', () => {
    vi.mocked(useGetMyBathhouses).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhouses>)
    mockAuth('owner')

    renderWithProviders(<BathhouseList />)

    expect(screen.getByText(/У вас пока нет объектов/)).toBeInTheDocument()
  })

  it('shows delete button only for owner', () => {
    vi.mocked(useGetMyBathhouses).mockReturnValue({
      data: { data: mockBathhouses, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhouses>)
    mockAuth('representative')

    renderWithProviders(<BathhouseList />)

    expect(screen.queryAllByText('Удалить')).toHaveLength(0)
    expect(screen.getAllByText('Изменить')).toHaveLength(2)
  })

  it('shows duplicate button for owner', () => {
    vi.mocked(useGetMyBathhouses).mockReturnValue({
      data: { data: mockBathhouses, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhouses>)
    mockAuth('owner')

    renderWithProviders(<BathhouseList />)

    expect(screen.getAllByText('Дублировать')).toHaveLength(2)
  })

  it('hides duplicate button for representative', () => {
    vi.mocked(useGetMyBathhouses).mockReturnValue({
      data: { data: mockBathhouses, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhouses>)
    mockAuth('representative')

    renderWithProviders(<BathhouseList />)

    expect(screen.queryAllByText('Дублировать')).toHaveLength(0)
  })

  it('shows import button for owner', () => {
    vi.mocked(useGetMyBathhouses).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhouses>)
    mockAuth('owner')

    renderWithProviders(<BathhouseList />)

    expect(screen.getByText('Импорт')).toBeInTheDocument()
  })
})
