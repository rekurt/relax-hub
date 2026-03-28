import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import GuestCardDetail from '@/pages/crm/GuestCardDetail'

vi.mock('@/api/generated/crm/crm', () => ({
  useGetMyCrmGuests: vi.fn(),
  usePutMyCrmGuestsId: vi.fn(),
}))

import {
  useGetMyCrmGuests,
  usePutMyCrmGuestsId,
} from '@/api/generated/crm/crm'

function renderWithProviders(ui: React.ReactElement, initialRoute = '/crm/guests/card-1') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU} theme={{ token: { motion: false } }}>
        <AntApp>
          <MemoryRouter initialEntries={[initialRoute]}>
            <Routes>
              <Route path="/crm/guests/:id" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockGuest = {
  id: 'card-1',
  client_id: 'client-abc-12345678',
  owner_id: 'owner-1',
  bathhouse_id: 'bath-1',
  visit_count: 7,
  total_spent: 2100000,
  avg_check: 300000,
  notes: 'Предпочитает веники из дуба',
  tags: ['vip', 'постоянный'],
  first_visit_at: '2025-06-01T10:00:00Z',
  last_visit_at: '2026-03-20T14:00:00Z',
  created_at: '2025-06-01T10:00:00Z',
  updated_at: '2026-03-20T14:00:00Z',
}

const mockUpdateMutation = { mutate: vi.fn(), isPending: false }

describe('GuestCardDetail', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(usePutMyCrmGuestsId).mockReturnValue(
      mockUpdateMutation as unknown as ReturnType<typeof usePutMyCrmGuestsId>,
    )
  })

  it('renders guest detail with stats', () => {
    vi.mocked(useGetMyCrmGuests).mockReturnValue({
      data: { data: [mockGuest], success: true, meta: { total_count: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmGuests>)

    renderWithProviders(<GuestCardDetail />)

    expect(screen.getByText('Карточка гостя')).toBeInTheDocument()
    expect(screen.getByText('Визиты')).toBeInTheDocument()
    expect(screen.getByText('7')).toBeInTheDocument()
  })

  it('renders guest tags', () => {
    vi.mocked(useGetMyCrmGuests).mockReturnValue({
      data: { data: [mockGuest], success: true, meta: { total_count: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmGuests>)

    renderWithProviders(<GuestCardDetail />)

    expect(screen.getByText('vip')).toBeInTheDocument()
    expect(screen.getByText('постоянный')).toBeInTheDocument()
  })

  it('renders notes and tags editor', () => {
    vi.mocked(useGetMyCrmGuests).mockReturnValue({
      data: { data: [mockGuest], success: true, meta: { total_count: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmGuests>)

    renderWithProviders(<GuestCardDetail />)

    expect(screen.getByText('Заметки и теги')).toBeInTheDocument()
    expect(screen.getByText('Сохранить')).toBeInTheDocument()
  })

  it('shows not found when guest does not exist', () => {
    vi.mocked(useGetMyCrmGuests).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmGuests>)

    renderWithProviders(<GuestCardDetail />, '/crm/guests/nonexistent')

    expect(screen.getByText('Карточка не найдена')).toBeInTheDocument()
  })

  it('shows back button', () => {
    vi.mocked(useGetMyCrmGuests).mockReturnValue({
      data: { data: [mockGuest], success: true, meta: { total_count: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmGuests>)

    renderWithProviders(<GuestCardDetail />)

    expect(screen.getByText('Назад')).toBeInTheDocument()
  })
})
