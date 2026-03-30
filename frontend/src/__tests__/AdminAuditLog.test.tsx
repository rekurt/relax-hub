import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi } from 'vitest'
import AdminAuditLog from '@/pages/admin/AdminAuditLog'

vi.mock('@/api/generated/admin-audit/admin-audit', () => ({
  useGetAdminAuditLog: vi.fn(),
  useGetAdminAuditLogActions: vi.fn(),
}))

import {
  useGetAdminAuditLog,
  useGetAdminAuditLogActions,
} from '@/api/generated/admin-audit/admin-audit'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/audit-log']}>
            <Routes>
              <Route path="/admin/audit-log" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockEntityLogs = [
  {
    id: 'al-1',
    entity_type: 'bathhouse',
    entity_id: 'bh-12345678-abcd-efgh',
    user_id: 'u-12345678-abcd-efgh',
    action: 'update',
    changed_fields: [1, 2, 3],
    created_at: '2026-03-29T10:00:00Z',
  },
  {
    id: 'al-2',
    entity_type: 'booking',
    entity_id: 'bk-12345678-abcd-efgh',
    user_id: 'u-87654321-dcba-hgfe',
    action: 'create',
    changed_fields: [],
    created_at: '2026-03-29T09:00:00Z',
  },
]

const mockAdminLogs = [
  {
    id: 'al-3',
    admin_id: 'a-12345678-abcd-efgh',
    action: 'update',
    created_at: '2026-03-29T11:00:00Z',
  },
  {
    id: 'al-4',
    admin_id: 'a-87654321-dcba-hgfe',
    action: 'delete',
    created_at: '2026-03-29T10:30:00Z',
  },
]

function setupMocks() {
  vi.mocked(useGetAdminAuditLog).mockReturnValue({
    data: { data: mockEntityLogs, meta: { total_count: 2 }, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetAdminAuditLog>)
  vi.mocked(useGetAdminAuditLogActions).mockReturnValue({
    data: { data: mockAdminLogs, meta: { total_count: 2 }, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetAdminAuditLogActions>)
}

describe('AdminAuditLog', () => {
  it('renders title and tab selector', () => {
    setupMocks()
    renderWithProviders(<AdminAuditLog />)

    expect(screen.getByText('Журнал аудита')).toBeInTheDocument()
    expect(screen.getByText('Действия админов')).toBeInTheDocument()
    expect(screen.getByText('Изменения сущностей')).toBeInTheDocument()
  })

  it('renders admin actions tab by default', () => {
    setupMocks()
    renderWithProviders(<AdminAuditLog />)

    expect(screen.getByText('Обновление')).toBeInTheDocument()
    expect(screen.getByText('Удаление')).toBeInTheDocument()
  })

  it('switches to entity changes tab', () => {
    setupMocks()
    renderWithProviders(<AdminAuditLog />)

    fireEvent.click(screen.getByText('Изменения сущностей'))
    expect(screen.getByText('Тип сущности')).toBeInTheDocument()
    expect(screen.getByText('ID сущности')).toBeInTheDocument()
  })

  it('renders filter controls', () => {
    setupMocks()
    renderWithProviders(<AdminAuditLog />)

    expect(screen.getAllByText('Действие').length).toBeGreaterThanOrEqual(1)
  })

  it('renders empty state when no data', () => {
    vi.mocked(useGetAdminAuditLog).mockReturnValue({
      data: { data: [], meta: { total_count: 0 }, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAuditLog>)
    vi.mocked(useGetAdminAuditLogActions).mockReturnValue({
      data: { data: [], meta: { total_count: 0 }, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminAuditLogActions>)

    renderWithProviders(<AdminAuditLog />)
    expect(screen.getByText('Нет записей')).toBeInTheDocument()
  })
})
