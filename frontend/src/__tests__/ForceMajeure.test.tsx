import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi } from 'vitest'
import ForceMajeure from '@/pages/admin/ForceMajeure'

vi.mock('@/api/generated/admin/admin', () => ({
  useGetAdminForceMajeure: vi.fn(),
  usePostAdminForceMajeure: vi.fn(() => ({ mutateAsync: vi.fn(), isPending: false })),
  getGetAdminForceMajeureQueryKey: vi.fn(() => ['/admin/force-majeure']),
}))

import { useGetAdminForceMajeure } from '@/api/generated/admin/admin'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/force-majeure']}>
            <Routes>
              <Route path="/admin/force-majeure" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockEvents = [
  {
    id: 'fm-1',
    region: 'RU',
    date_from: '2026-03-01',
    date_to: '2026-03-05',
    reason: 'Чрезвычайная ситуация в регионе — наводнение',
    affected_count: 42,
    total_refund: 1500000,
    admin_id: 'admin-1',
    created_at: '2026-03-01T10:00:00Z',
  },
  {
    id: 'fm-2',
    region: 'BY',
    date_from: '2026-02-15',
    date_to: '2026-02-20',
    reason: 'Экстремальные морозы',
    affected_count: 15,
    total_refund: 450000,
    admin_id: 'admin-2',
    created_at: '2026-02-15T08:00:00Z',
  },
]

describe('ForceMajeure', () => {
  it('renders page title and activation form', () => {
    vi.mocked(useGetAdminForceMajeure).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminForceMajeure>)

    renderWithProviders(<ForceMajeure />)

    expect(screen.getByText('Форс-мажор')).toBeInTheDocument()
    expect(screen.getByText('Активировать форс-мажор', { selector: '.ant-card-head-title' })).toBeInTheDocument()
    expect(screen.getByText('Активировать форс-мажор', { selector: 'button span' })).toBeInTheDocument()
  })

  it('renders event history table with data', () => {
    vi.mocked(useGetAdminForceMajeure).mockReturnValue({
      data: { data: mockEvents, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminForceMajeure>)

    renderWithProviders(<ForceMajeure />)

    expect(screen.getByText('История событий')).toBeInTheDocument()
    expect(screen.getByText('Чрезвычайная ситуация в регионе — наводнение')).toBeInTheDocument()
    expect(screen.getByText('Экстремальные морозы')).toBeInTheDocument()
    expect(screen.getByText('RU')).toBeInTheDocument()
    expect(screen.getByText('BY')).toBeInTheDocument()
  })

  it('renders affected count and refund amounts', () => {
    vi.mocked(useGetAdminForceMajeure).mockReturnValue({
      data: { data: mockEvents, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminForceMajeure>)

    renderWithProviders(<ForceMajeure />)

    expect(screen.getByText('42')).toBeInTheDocument()
    expect(screen.getByText('15')).toBeInTheDocument()
  })

  it('renders empty state when no events', () => {
    vi.mocked(useGetAdminForceMajeure).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminForceMajeure>)

    renderWithProviders(<ForceMajeure />)

    expect(screen.getByText('Нет событий форс-мажора')).toBeInTheDocument()
  })

  it('renders form fields with labels', () => {
    vi.mocked(useGetAdminForceMajeure).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminForceMajeure>)

    renderWithProviders(<ForceMajeure />)

    expect(screen.getByLabelText('Регион')).toBeInTheDocument()
    expect(screen.getByLabelText('Период')).toBeInTheDocument()
    expect(screen.getByLabelText('Причина')).toBeInTheDocument()
  })

  it('renders detail drawer when clicking event', () => {
    vi.mocked(useGetAdminForceMajeure).mockReturnValue({
      data: { data: mockEvents, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminForceMajeure>)

    renderWithProviders(<ForceMajeure />)

    const detailLinks = screen.getAllByText('Подробнее')
    fireEvent.click(detailLinks[0])

    expect(screen.getByText('Детали форс-мажора')).toBeInTheDocument()
    expect(screen.getByText('Затронутые бронирования')).toBeInTheDocument()
    expect(screen.getByText('Дата активации')).toBeInTheDocument()
  })

  it('renders description text about force majeure', () => {
    vi.mocked(useGetAdminForceMajeure).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminForceMajeure>)

    renderWithProviders(<ForceMajeure />)

    expect(screen.getByText(/Массовая отмена бронирований по региону/)).toBeInTheDocument()
  })
})
