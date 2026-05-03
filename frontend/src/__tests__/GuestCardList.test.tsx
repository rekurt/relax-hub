import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import GuestCardList from '@/pages/crm/GuestCardList'

vi.mock('@/api/generated/crm/crm', () => ({
  useGetMyCrmGuests: vi.fn(),
  useGetMyCrmStats: vi.fn(),
}))

vi.mock('@/api/axios-instance', () => ({
  customInstance: vi.fn(),
}))

import {
  useGetMyCrmGuests,
  useGetMyCrmStats,
} from '@/api/generated/crm/crm'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU} theme={{ token: { motion: false } }}>
        <AntApp>
          <MemoryRouter>{ui}</MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockGuests = [
  {
    id: 'card-1',
    client_id: 'client-abc-123',
    owner_id: 'owner-1',
    bathhouse_id: 'bath-1',
    visit_count: 5,
    total_spent: 1500000,
    avg_check: 300000,
    notes: 'VIP гость',
    tags: ['vip', 'постоянный'],
    first_visit_at: '2025-06-01T10:00:00Z',
    last_visit_at: '2026-03-15T14:00:00Z',
    created_at: '2025-06-01T10:00:00Z',
    updated_at: '2026-03-15T14:00:00Z',
  },
  {
    id: 'card-2',
    client_id: 'client-def-456',
    owner_id: 'owner-1',
    bathhouse_id: 'bath-1',
    visit_count: 1,
    total_spent: 300000,
    avg_check: 300000,
    notes: '',
    tags: [],
    first_visit_at: '2026-03-20T10:00:00Z',
    last_visit_at: '2026-03-20T10:00:00Z',
    created_at: '2026-03-20T10:00:00Z',
    updated_at: '2026-03-20T10:00:00Z',
  },
]

const mockStats = {
  total_guests: 42,
  new_this_month: 8,
  avg_visit_count: 3,
  avg_spent: 350000,
}

describe('GuestCardList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders guest card list with title', () => {
    vi.mocked(useGetMyCrmGuests).mockReturnValue({
      data: { data: mockGuests, success: true, meta: { total_count: 2 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmGuests>)
    vi.mocked(useGetMyCrmStats).mockReturnValue({
      data: { data: mockStats },
    } as unknown as ReturnType<typeof useGetMyCrmStats>)

    renderWithProviders(<GuestCardList />)

    expect(screen.getByText('Гостевые карточки')).toBeInTheDocument()
  })

  it('renders stats cards', () => {
    vi.mocked(useGetMyCrmGuests).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmGuests>)
    vi.mocked(useGetMyCrmStats).mockReturnValue({
      data: { data: mockStats },
    } as unknown as ReturnType<typeof useGetMyCrmStats>)

    renderWithProviders(<GuestCardList />)

    expect(screen.getByText('Всего гостей')).toBeInTheDocument()
    expect(screen.getByText('42')).toBeInTheDocument()
    expect(screen.getByText('Новых за месяц')).toBeInTheDocument()
    expect(screen.getByText('8')).toBeInTheDocument()
  })

  it('renders guest rows with visit count and spent', () => {
    vi.mocked(useGetMyCrmGuests).mockReturnValue({
      data: { data: mockGuests, success: true, meta: { total_count: 2 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmGuests>)
    vi.mocked(useGetMyCrmStats).mockReturnValue({
      data: { data: mockStats },
    } as unknown as ReturnType<typeof useGetMyCrmStats>)

    renderWithProviders(<GuestCardList />)

    expect(screen.getByText('5')).toBeInTheDocument()
    expect(screen.getByText('15000 ₽')).toBeInTheDocument()
  })

  it('renders tags for guests', () => {
    vi.mocked(useGetMyCrmGuests).mockReturnValue({
      data: { data: mockGuests, success: true, meta: { total_count: 2 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmGuests>)
    vi.mocked(useGetMyCrmStats).mockReturnValue({
      data: { data: mockStats },
    } as unknown as ReturnType<typeof useGetMyCrmStats>)

    renderWithProviders(<GuestCardList />)

    expect(screen.getByText('vip')).toBeInTheDocument()
    expect(screen.getByText('постоянный')).toBeInTheDocument()
  })

  it('shows empty state when no guests', () => {
    vi.mocked(useGetMyCrmGuests).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmGuests>)
    vi.mocked(useGetMyCrmStats).mockReturnValue({
      data: { data: mockStats },
    } as unknown as ReturnType<typeof useGetMyCrmStats>)

    renderWithProviders(<GuestCardList />)

    expect(screen.getByText('Гостей пока нет. Они появятся после первого завершённого бронирования')).toBeInTheDocument()
  })

  it('shows export CSV button', () => {
    vi.mocked(useGetMyCrmGuests).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmGuests>)
    vi.mocked(useGetMyCrmStats).mockReturnValue({
      data: { data: mockStats },
    } as unknown as ReturnType<typeof useGetMyCrmStats>)

    renderWithProviders(<GuestCardList />)

    expect(screen.getByText('Экспорт CSV')).toBeInTheDocument()
  })
})
