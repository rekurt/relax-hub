import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi } from 'vitest'
import BathhouseModeration from '@/pages/admin/BathhouseModeration'

vi.mock('@/api/generated/admin-bathhouses/admin-bathhouses', () => ({
  useGetAdminBathhouses: vi.fn(),
  usePatchAdminBathhousesIdApprove: vi.fn(() => ({ mutateAsync: vi.fn() })),
  usePatchAdminBathhousesIdReject: vi.fn(() => ({ mutateAsync: vi.fn() })),
  getGetAdminBathhousesQueryKey: vi.fn(() => ['/admin/bathhouses']),
}))

import { useGetAdminBathhouses } from '@/api/generated/admin-bathhouses/admin-bathhouses'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/bathhouses']}>
            <Routes>
              <Route path="/admin/bathhouses" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockBathhouses = [
  {
    id: 'b-1',
    name: 'Баня на берёзовой',
    address: 'ул. Берёзовая, 10',
    status: 'pending',
    price_per_hour: 300000,
    rating: 4.5,
    review_count: 12,
    max_guests: 8,
    has_sauna: true,
    has_pool: true,
    has_steam_room: false,
    is_photo_verified: true,
    created_at: '2025-01-15T10:00:00Z',
  },
  {
    id: 'b-2',
    name: 'Парная люкс',
    address: 'пр. Ленина, 55',
    status: 'active',
    price_per_hour: 500000,
    rating: 4.9,
    review_count: 30,
    max_guests: 12,
    has_sauna: true,
    has_pool: true,
    has_steam_room: true,
    has_karaoke: true,
    is_photo_verified: false,
    created_at: '2024-12-01T08:00:00Z',
  },
  {
    id: 'b-3',
    name: 'Отклонённая баня',
    address: 'ул. Тестовая, 1',
    status: 'rejected',
    price_per_hour: 150000,
    rating: 0,
    review_count: 0,
    max_guests: 4,
    is_photo_verified: false,
    created_at: '2025-02-01T12:00:00Z',
  },
]

describe('BathhouseModeration', () => {
  it('renders bathhouse table with data', () => {
    vi.mocked(useGetAdminBathhouses).mockReturnValue({
      data: {
        data: mockBathhouses,
        success: true,
        meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminBathhouses>)

    renderWithProviders(<BathhouseModeration />)

    expect(screen.getByText('Модерация бань')).toBeInTheDocument()
    expect(screen.getByText('Баня на берёзовой')).toBeInTheDocument()
    expect(screen.getByText('Парная люкс')).toBeInTheDocument()
    expect(screen.getByText('Отклонённая баня')).toBeInTheDocument()
  })

  it('renders status badges', () => {
    vi.mocked(useGetAdminBathhouses).mockReturnValue({
      data: {
        data: mockBathhouses,
        success: true,
        meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminBathhouses>)

    renderWithProviders(<BathhouseModeration />)

    // "На рассмотрении" appears in both the Segmented filter and the table Tag
    expect(screen.getAllByText('На рассмотрении').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Активна')).toBeInTheDocument()
    expect(screen.getByText('Отклонена')).toBeInTheDocument()
  })

  it('renders approve/reject action buttons based on status', () => {
    vi.mocked(useGetAdminBathhouses).mockReturnValue({
      data: {
        data: mockBathhouses,
        success: true,
        meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminBathhouses>)

    renderWithProviders(<BathhouseModeration />)

    const approveButtons = screen.getAllByText('Одобрить')
    expect(approveButtons.length).toBe(1)

    const rejectButtons = screen.getAllByText('Отклонить')
    expect(rejectButtons.length).toBe(1)

    expect(screen.getByText('Объект уже опубликован')).toBeInTheDocument()
    expect(screen.getByText('Заявка уже отклонена')).toBeInTheDocument()
  })

  it('renders status filter', () => {
    vi.mocked(useGetAdminBathhouses).mockReturnValue({
      data: { data: [], success: true, meta: { page: 1, page_size: 20, total_count: 0, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminBathhouses>)

    renderWithProviders(<BathhouseModeration />)

    expect(screen.getByText('Все')).toBeInTheDocument()
    expect(screen.getByText('Активные')).toBeInTheDocument()
    expect(screen.getByText('Отклонённые')).toBeInTheDocument()
  })

  it('renders search input', () => {
    vi.mocked(useGetAdminBathhouses).mockReturnValue({
      data: { data: [], success: true, meta: {} },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminBathhouses>)

    renderWithProviders(<BathhouseModeration />)

    expect(screen.getByPlaceholderText('Поиск по названию или адресу')).toBeInTheDocument()
  })

  it('renders empty state', () => {
    vi.mocked(useGetAdminBathhouses).mockReturnValue({
      data: { data: [], success: true, meta: { page: 1, page_size: 20, total_count: 0, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminBathhouses>)

    renderWithProviders(<BathhouseModeration />)

    expect(screen.getByText('Нет бань')).toBeInTheDocument()
  })

  it('renders total count in pagination', () => {
    vi.mocked(useGetAdminBathhouses).mockReturnValue({
      data: {
        data: mockBathhouses,
        success: true,
        meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminBathhouses>)

    renderWithProviders(<BathhouseModeration />)

    expect(screen.getByText('Всего: 3')).toBeInTheDocument()
  })
})
