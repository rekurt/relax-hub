import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import AuditLog from '@/pages/bathhouses/AuditLog'

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetMyBathhousesIdHistory: vi.fn(),
}))

import { useGetMyBathhousesIdHistory } from '@/api/generated/bathhouses/bathhouses'

function renderWithProviders(bathhouseId = 'bh-1') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={[`/bathhouses/${bathhouseId}/audit`]}>
            <Routes>
              <Route path="/bathhouses/:id/audit" element={<AuditLog />} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockEntries = [
  {
    id: 'audit-1',
    action: 'create',
    user_id: 'user-abc-123-def',
    entity_type: 'bathhouse',
    entity_id: 'bh-1',
    changed_fields: { name: { old: null, new: 'Баня Люкс' } },
    created_at: '2026-03-20T14:00:00Z',
  },
  {
    id: 'audit-2',
    action: 'update',
    user_id: 'user-xyz-456-ghi',
    entity_type: 'bathhouse',
    entity_id: 'bh-1',
    changed_fields: { base_price: { old: 200000, new: 300000 }, name: { old: 'Баня Люкс', new: 'Баня Премиум' } },
    created_at: '2026-03-21T10:00:00Z',
  },
]

describe('AuditLog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders audit log page with title', () => {
    vi.mocked(useGetMyBathhousesIdHistory).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0, page: 0, page_size: 20, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdHistory>)

    renderWithProviders()

    expect(screen.getByText('История изменений')).toBeInTheDocument()
  })

  it('renders empty state when no entries', () => {
    vi.mocked(useGetMyBathhousesIdHistory).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0, page: 0, page_size: 20, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdHistory>)

    renderWithProviders()

    expect(screen.getByText('История изменений пуста')).toBeInTheDocument()
  })

  it('renders audit entries with action tags', () => {
    vi.mocked(useGetMyBathhousesIdHistory).mockReturnValue({
      data: { data: mockEntries, success: true, meta: { total_count: 2, page: 0, page_size: 20, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdHistory>)

    renderWithProviders()

    expect(screen.getByText('Создание')).toBeInTheDocument()
    expect(screen.getByText('Изменение')).toBeInTheDocument()
  })

  it('renders changed fields with diff visualization', () => {
    vi.mocked(useGetMyBathhousesIdHistory).mockReturnValue({
      data: { data: mockEntries, success: true, meta: { total_count: 2, page: 0, page_size: 20, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdHistory>)

    renderWithProviders()

    expect(screen.getAllByText('Баня Люкс').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Баня Премиум')).toBeInTheDocument()
  })

  it('renders date filter', () => {
    vi.mocked(useGetMyBathhousesIdHistory).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0, page: 0, page_size: 20, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdHistory>)

    renderWithProviders()

    expect(screen.getByText('Фильтр по дате:')).toBeInTheDocument()
  })

  it('shows total count in pagination', () => {
    vi.mocked(useGetMyBathhousesIdHistory).mockReturnValue({
      data: { data: mockEntries, success: true, meta: { total_count: 2, page: 0, page_size: 20, total_pages: 1 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdHistory>)

    renderWithProviders()

    expect(screen.getByText('Всего: 2')).toBeInTheDocument()
  })
})
