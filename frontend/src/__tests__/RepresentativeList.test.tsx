import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import RepresentativeList from '@/pages/representatives/RepresentativeList'

vi.mock('@/api/generated/representatives/representatives', () => ({
  useGetBathhousesIdRepresentatives: vi.fn(),
  usePostBathhousesIdRepresentatives: vi.fn(),
  useDeleteRepresentativesId: vi.fn(),
}))

vi.mock('@/stores/bathhouse', () => ({
  useBathhouseStore: vi.fn(),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn(),
}))

import {
  useGetBathhousesIdRepresentatives,
  usePostBathhousesIdRepresentatives,
  useDeleteRepresentativesId,
} from '@/api/generated/representatives/representatives'
import { useBathhouseStore } from '@/stores/bathhouse'
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

const mockRepresentatives = [
  {
    id: 'rep-1',
    user_id: 'user-abc12345-6789',
    bathhouse_id: 'bath-1',
    owner_id: 'owner-1',
    created_at: '2026-02-15T10:30:00Z',
  },
  {
    id: 'rep-2',
    user_id: 'user-def98765-4321',
    bathhouse_id: 'bath-1',
    owner_id: 'owner-1',
    created_at: '2026-03-01T14:00:00Z',
  },
]

const mockInviteMutation = { mutate: vi.fn(), isPending: false }
const mockDeleteMutation = { mutate: vi.fn(), isPending: false }

function mockBathhouseStore(id: string | null) {
  vi.mocked(useBathhouseStore).mockImplementation((selector) =>
    (selector as (state: { selectedBathhouseId: string | null }) => unknown)({
      selectedBathhouseId: id,
    }),
  )
}

function mockAuthStoreWithRole(role: string) {
  vi.mocked(useAuthStore).mockImplementation((selector) =>
    (selector as (state: { user: { role: string } | null }) => unknown)({
      user: { role },
    }),
  )
}

function setupDefaultMocks() {
  vi.mocked(usePostBathhousesIdRepresentatives).mockReturnValue(
    mockInviteMutation as unknown as ReturnType<typeof usePostBathhousesIdRepresentatives>,
  )
  vi.mocked(useDeleteRepresentativesId).mockReturnValue(
    mockDeleteMutation as unknown as ReturnType<typeof useDeleteRepresentativesId>,
  )
}

describe('RepresentativeList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setupDefaultMocks()
  })

  it('shows prompt when no bathhouse selected', () => {
    mockBathhouseStore(null)
    mockAuthStoreWithRole('owner')
    vi.mocked(useGetBathhousesIdRepresentatives).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdRepresentatives>)

    renderWithProviders(<RepresentativeList />)

    expect(screen.getByText('Представители')).toBeInTheDocument()
    expect(screen.getByText('Выберите баню для управления представителями')).toBeInTheDocument()
  })

  it('renders representatives table with data', () => {
    mockBathhouseStore('bath-1')
    mockAuthStoreWithRole('owner')
    vi.mocked(useGetBathhousesIdRepresentatives).mockReturnValue({
      data: { data: mockRepresentatives, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdRepresentatives>)

    renderWithProviders(<RepresentativeList />)

    expect(screen.getByText(/user-abc/)).toBeInTheDocument()
    expect(screen.getByText(/user-def/)).toBeInTheDocument()
  })

  it('shows formatted dates', () => {
    mockBathhouseStore('bath-1')
    mockAuthStoreWithRole('owner')
    vi.mocked(useGetBathhousesIdRepresentatives).mockReturnValue({
      data: { data: mockRepresentatives, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdRepresentatives>)

    renderWithProviders(<RepresentativeList />)

    expect(screen.getByText(/15\.02\.2026/)).toBeInTheDocument()
    expect(screen.getByText(/01\.03\.2026/)).toBeInTheDocument()
  })

  it('shows empty state when no representatives', () => {
    mockBathhouseStore('bath-1')
    mockAuthStoreWithRole('owner')
    vi.mocked(useGetBathhousesIdRepresentatives).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdRepresentatives>)

    renderWithProviders(<RepresentativeList />)

    expect(screen.getByText(/Нет представителей/)).toBeInTheDocument()
  })

  it('shows invite button for owner role', () => {
    mockBathhouseStore('bath-1')
    mockAuthStoreWithRole('owner')
    vi.mocked(useGetBathhousesIdRepresentatives).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdRepresentatives>)

    renderWithProviders(<RepresentativeList />)

    expect(screen.getByText('Пригласить')).toBeInTheDocument()
  })

  it('hides invite button for representative role', () => {
    mockBathhouseStore('bath-1')
    mockAuthStoreWithRole('representative')
    vi.mocked(useGetBathhousesIdRepresentatives).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdRepresentatives>)

    renderWithProviders(<RepresentativeList />)

    expect(screen.queryByText('Пригласить')).not.toBeInTheDocument()
  })

  it('hides delete buttons for representative role', () => {
    mockBathhouseStore('bath-1')
    mockAuthStoreWithRole('representative')
    vi.mocked(useGetBathhousesIdRepresentatives).mockReturnValue({
      data: { data: mockRepresentatives, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdRepresentatives>)

    renderWithProviders(<RepresentativeList />)

    const deleteButtons = screen.queryAllByRole('button').filter(
      (btn) => btn.querySelector('.anticon-delete'),
    )
    expect(deleteButtons.length).toBe(0)
  })

  it('opens invite modal when button clicked', () => {
    mockBathhouseStore('bath-1')
    mockAuthStoreWithRole('owner')
    vi.mocked(useGetBathhousesIdRepresentatives).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdRepresentatives>)

    renderWithProviders(<RepresentativeList />)

    fireEvent.click(screen.getByText('Пригласить'))

    expect(screen.getByText('Пригласить представителя')).toBeInTheDocument()
    expect(screen.getByLabelText('Email пользователя')).toBeInTheDocument()
  })

  it('calls delete mutation when confirmed', () => {
    mockBathhouseStore('bath-1')
    mockAuthStoreWithRole('owner')
    vi.mocked(useGetBathhousesIdRepresentatives).mockReturnValue({
      data: { data: [mockRepresentatives[0]], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdRepresentatives>)

    renderWithProviders(<RepresentativeList />)

    const deleteButtons = screen.getAllByRole('button').filter(
      (btn) => btn.querySelector('.anticon-delete'),
    )
    expect(deleteButtons.length).toBeGreaterThan(0)
    fireEvent.click(deleteButtons[0]!)

    expect(screen.getByText('Удалить представителя?')).toBeInTheDocument()

    const confirmBtn = screen.getByRole('button', { name: 'Удалить' })
    fireEvent.click(confirmBtn)

    expect(mockDeleteMutation.mutate).toHaveBeenCalledWith({ id: 'rep-1' })
  })

  it('shows delete buttons for owner role', () => {
    mockBathhouseStore('bath-1')
    mockAuthStoreWithRole('owner')
    vi.mocked(useGetBathhousesIdRepresentatives).mockReturnValue({
      data: { data: mockRepresentatives, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdRepresentatives>)

    renderWithProviders(<RepresentativeList />)

    const deleteButtons = screen.getAllByRole('button').filter(
      (btn) => btn.querySelector('.anticon-delete'),
    )
    expect(deleteButtons.length).toBe(2)
  })
})
