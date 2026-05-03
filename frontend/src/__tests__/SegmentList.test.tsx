import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import SegmentList from '@/pages/crm/SegmentList'

vi.mock('@/api/generated/crm/crm', () => ({
  useGetMyCrmSegments: vi.fn(),
  useGetMyCrmSegmentsSlugGuests: vi.fn(),
}))

import {
  useGetMyCrmSegments,
  useGetMyCrmSegmentsSlugGuests,
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

const mockSegments = [
  { slug: 'new', name: 'Новые', description: '1 визит', count: 12 },
  { slug: 'regular', name: 'Постоянные', description: '3+ визитов', count: 45 },
  { slug: 'lost', name: 'Потерянные', description: 'Более 90 дней без визита', count: 8 },
  { slug: 'vip', name: 'VIP', description: 'Потратили более 50 000 ₽', count: 3 },
  { slug: 'birthday_soon', name: 'День рождения скоро', description: 'В ближайшие 7 дней', count: 2 },
]

const mockSegmentGuests = [
  {
    id: 'card-1',
    client_id: 'client-abc-123',
    visit_count: 1,
    total_spent: 300000,
    last_visit_at: '2026-03-25T10:00:00Z',
    tags: ['новичок'],
  },
]

describe('SegmentList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useGetMyCrmSegmentsSlugGuests).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmSegmentsSlugGuests>)
  })

  it('renders page title', () => {
    vi.mocked(useGetMyCrmSegments).mockReturnValue({
      data: { data: mockSegments },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmSegments>)

    renderWithProviders(<SegmentList />)

    expect(screen.getByText('Сегменты гостей')).toBeInTheDocument()
  })

  it('renders all segment cards with names and counts', () => {
    vi.mocked(useGetMyCrmSegments).mockReturnValue({
      data: { data: mockSegments },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmSegments>)

    renderWithProviders(<SegmentList />)

    expect(screen.getByText('Новые')).toBeInTheDocument()
    expect(screen.getByText('Постоянные')).toBeInTheDocument()
    expect(screen.getByText('Потерянные')).toBeInTheDocument()
    expect(screen.getByText('VIP')).toBeInTheDocument()
    expect(screen.getByText('12')).toBeInTheDocument()
    expect(screen.getByText('45')).toBeInTheDocument()
  })

  it('shows placeholder when no segment selected', () => {
    vi.mocked(useGetMyCrmSegments).mockReturnValue({
      data: { data: mockSegments },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmSegments>)

    renderWithProviders(<SegmentList />)

    expect(screen.getByText('Выберите сегмент для просмотра гостей')).toBeInTheDocument()
  })

  it('loads guests when segment is clicked', () => {
    vi.mocked(useGetMyCrmSegments).mockReturnValue({
      data: { data: mockSegments },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmSegments>)
    vi.mocked(useGetMyCrmSegmentsSlugGuests).mockReturnValue({
      data: { data: mockSegmentGuests, success: true, meta: { total_count: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmSegmentsSlugGuests>)

    renderWithProviders(<SegmentList />)

    fireEvent.click(screen.getByText('Новые'))

    expect(screen.getByText('новичок')).toBeInTheDocument()
  })

  it('shows empty segments state', () => {
    vi.mocked(useGetMyCrmSegments).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmSegments>)

    renderWithProviders(<SegmentList />)

    expect(screen.getByText('Нет сегментов')).toBeInTheDocument()
  })
})
