import { render, screen, fireEvent, within, act, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import PricingRules from '@/pages/pricing/PricingRules'

vi.mock('@/api/generated/pricing/pricing', () => ({
  useGetMyBathhousesIdPricingRules: vi.fn(),
  usePostMyBathhousesIdPricingRules: vi.fn(),
  usePutPricingRulesId: vi.fn(),
  useDeletePricingRulesId: vi.fn(),
}))

vi.mock('@/stores/bathhouse', () => ({
  useBathhouseStore: vi.fn(),
}))

import {
  useGetMyBathhousesIdPricingRules,
  usePostMyBathhousesIdPricingRules,
  usePutPricingRulesId,
  useDeletePricingRulesId,
} from '@/api/generated/pricing/pricing'
import { useBathhouseStore } from '@/stores/bathhouse'

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

const mockRules = [
  {
    id: 'rule-1',
    bathhouse_id: 'bath-1',
    name: 'Выходные повышение',
    type: 'weekend',
    multiplier: 1.5,
    days_of_week: [5, 6],
    time_from: null,
    time_to: null,
    date_from: null,
    date_to: null,
    priority: 10,
    is_active: true,
    created_at: '2026-03-01T10:00:00Z',
  },
  {
    id: 'rule-2',
    bathhouse_id: 'bath-1',
    name: 'Ночное время',
    type: 'time_range',
    multiplier: 0.8,
    days_of_week: null,
    time_from: '22:00',
    time_to: '06:00',
    date_from: null,
    date_to: null,
    priority: 5,
    is_active: false,
    created_at: '2026-03-02T10:00:00Z',
  },
  {
    id: 'rule-3',
    bathhouse_id: 'bath-1',
    name: 'Новогодние праздники',
    type: 'holiday',
    multiplier: 2.0,
    days_of_week: null,
    time_from: null,
    time_to: null,
    date_from: '2026-12-31',
    date_to: '2027-01-08',
    priority: 20,
    is_active: true,
    created_at: '2026-03-03T10:00:00Z',
  },
]

const mockCreateMutation = { mutate: vi.fn(), isPending: false }
const mockUpdateMutation = { mutate: vi.fn(), isPending: false }
const mockDeleteMutation = { mutate: vi.fn(), isPending: false }

function mockBathhouseStore(id: string | null) {
  vi.mocked(useBathhouseStore).mockImplementation((selector) =>
    (selector as (state: { selectedBathhouseId: string | null }) => unknown)({
      selectedBathhouseId: id,
    }),
  )
}

function setupDefaultMocks() {
  vi.mocked(usePostMyBathhousesIdPricingRules).mockReturnValue(
    mockCreateMutation as unknown as ReturnType<typeof usePostMyBathhousesIdPricingRules>,
  )
  vi.mocked(usePutPricingRulesId).mockReturnValue(
    mockUpdateMutation as unknown as ReturnType<typeof usePutPricingRulesId>,
  )
  vi.mocked(useDeletePricingRulesId).mockReturnValue(
    mockDeleteMutation as unknown as ReturnType<typeof useDeletePricingRulesId>,
  )
}

describe('PricingRules', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setupDefaultMocks()
  })

  it('shows prompt when no bathhouse selected', () => {
    mockBathhouseStore(null)
    vi.mocked(useGetMyBathhousesIdPricingRules).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPricingRules>)

    renderWithProviders(<PricingRules />)

    expect(screen.getByText('Правила ценообразования')).toBeInTheDocument()
    expect(screen.getByText('Выберите баню для управления ценами')).toBeInTheDocument()
  })

  it('renders rules table with data', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPricingRules).mockReturnValue({
      data: { data: mockRules, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPricingRules>)

    renderWithProviders(<PricingRules />)

    expect(screen.getByText('Правила ценообразования')).toBeInTheDocument()
    expect(screen.getByText('Выходные повышение')).toBeInTheDocument()
    expect(screen.getByText('Ночное время')).toBeInTheDocument()
    expect(screen.getByText('Новогодние праздники')).toBeInTheDocument()
  })

  it('shows rule type tags with correct labels', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPricingRules).mockReturnValue({
      data: { data: mockRules, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPricingRules>)

    renderWithProviders(<PricingRules />)

    expect(screen.getByText('Выходные')).toBeInTheDocument()
    expect(screen.getByText('Диапазон времени')).toBeInTheDocument()
    expect(screen.getByText('Праздники')).toBeInTheDocument()
  })

  it('shows multiplier with percentage', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPricingRules).mockReturnValue({
      data: { data: [mockRules[0]], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPricingRules>)

    renderWithProviders(<PricingRules />)

    expect(screen.getByText(/x1\.5.*\+50%/)).toBeInTheDocument()
  })

  it('shows discount multiplier in green', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPricingRules).mockReturnValue({
      data: { data: [mockRules[1]], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPricingRules>)

    renderWithProviders(<PricingRules />)

    expect(screen.getByText(/x0\.8.*-20%/)).toBeInTheDocument()
  })

  it('shows conditions for time_range rules', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPricingRules).mockReturnValue({
      data: { data: [mockRules[1]], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPricingRules>)

    renderWithProviders(<PricingRules />)

    expect(screen.getByText(/22:00–06:00/)).toBeInTheDocument()
  })

  it('shows conditions for weekend rules with day names', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPricingRules).mockReturnValue({
      data: { data: [mockRules[0]], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPricingRules>)

    renderWithProviders(<PricingRules />)

    expect(screen.getByText(/Сб, Вс/)).toBeInTheDocument()
  })

  it('shows empty state when no rules', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPricingRules).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPricingRules>)

    renderWithProviders(<PricingRules />)

    expect(screen.getByText('Нет правил ценообразования')).toBeInTheDocument()
  })

  it('opens create modal when button clicked', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPricingRules).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPricingRules>)

    renderWithProviders(<PricingRules />)

    fireEvent.click(screen.getByText('Добавить правило'))

    expect(screen.getByText('Новое правило')).toBeInTheDocument()
    expect(screen.getByLabelText('Название')).toBeInTheDocument()
    expect(screen.getByText('Множитель цены')).toBeInTheDocument()
  })

  it('opens edit modal with existing data', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPricingRules).mockReturnValue({
      data: { data: [mockRules[0]], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPricingRules>)

    renderWithProviders(<PricingRules />)

    const editButtons = screen.getAllByRole('button').filter(
      (btn) => btn.querySelector('.anticon-edit'),
    )
    expect(editButtons.length).toBeGreaterThan(0)
    fireEvent.click(editButtons[0]!)

    expect(screen.getByText('Редактировать правило')).toBeInTheDocument()
  })

  it('calls delete mutation when confirmed', async () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPricingRules).mockReturnValue({
      data: { data: [mockRules[0]], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPricingRules>)

    renderWithProviders(<PricingRules />)

    const deleteButtons = screen.getAllByRole('button').filter(
      (btn) => btn.querySelector('.anticon-delete'),
    )
    expect(deleteButtons.length).toBeGreaterThan(0)
    await act(async () => {
      fireEvent.click(deleteButtons[0]!)
    })

    await waitFor(() => {
      expect(screen.getByText('Удалить правило?')).toBeInTheDocument()
    }, { timeout: 10000 })

    const confirmBtn = screen.getByRole('button', { name: 'Удалить' })
    await act(async () => {
      fireEvent.click(confirmBtn)
    })

    expect(mockDeleteMutation.mutate).toHaveBeenCalledWith({ id: 'rule-1' })
  }, 15000)

  it('renders switch for is_active toggle', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPricingRules).mockReturnValue({
      data: { data: mockRules, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPricingRules>)

    renderWithProviders(<PricingRules />)

    const switches = screen.getAllByRole('switch')
    expect(switches.length).toBe(3)
  })

  it('shows add button', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPricingRules).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPricingRules>)

    renderWithProviders(<PricingRules />)

    expect(screen.getByText('Добавить правило')).toBeInTheDocument()
  })

  it('shows priority column with values', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPricingRules).mockReturnValue({
      data: { data: mockRules, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPricingRules>)

    renderWithProviders(<PricingRules />)

    const table = screen.getByRole('table')
    const rows = within(table).getAllByRole('row')
    // header + 3 data rows
    expect(rows.length).toBeGreaterThanOrEqual(4)
  })
})
