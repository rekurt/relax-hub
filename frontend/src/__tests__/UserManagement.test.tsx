import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi } from 'vitest'
import UserManagement from '@/pages/admin/UserManagement'

vi.mock('@/api/generated/admin-users/admin-users', () => ({
  useGetAdminUsers: vi.fn(),
  usePatchAdminUsersIdBlock: vi.fn(() => ({ mutateAsync: vi.fn() })),
  usePatchAdminUsersIdUnblock: vi.fn(() => ({ mutateAsync: vi.fn() })),
  getGetAdminUsersQueryKey: vi.fn(() => ['/admin/users']),
}))

import { useGetAdminUsers } from '@/api/generated/admin-users/admin-users'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/users']}>
            <Routes>
              <Route path="/admin/users" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockUsers = [
  {
    id: 'u-1',
    name: 'Иван Петров',
    email: 'ivan@example.com',
    phone: '+79001234567',
    role: 'client',
    is_active: true,
  },
  {
    id: 'u-2',
    name: 'Анна Сидорова',
    email: 'anna@example.com',
    phone: '+79007654321',
    role: 'owner',
    is_active: true,
  },
  {
    id: 'u-3',
    name: 'Blocked User',
    email: 'blocked@example.com',
    phone: null,
    role: 'client',
    is_active: false,
  },
  {
    id: 'u-4',
    name: 'Admin User',
    email: 'admin@example.com',
    phone: null,
    role: 'admin',
    is_active: true,
  },
]

describe('UserManagement', () => {
  it('renders user table with data', () => {
    vi.mocked(useGetAdminUsers).mockReturnValue({
      data: {
        data: mockUsers,
        success: true,
        meta: { page: 1, page_size: 20, total_count: 4, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminUsers>)

    renderWithProviders(<UserManagement />)

    expect(screen.getByText('Управление пользователями')).toBeInTheDocument()
    expect(screen.getByText('Иван Петров')).toBeInTheDocument()
    expect(screen.getByText('Анна Сидорова')).toBeInTheDocument()
    expect(screen.getByText('ivan@example.com')).toBeInTheDocument()
  })

  it('renders role badges', () => {
    vi.mocked(useGetAdminUsers).mockReturnValue({
      data: {
        data: mockUsers,
        success: true,
        meta: { page: 1, page_size: 20, total_count: 4, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminUsers>)

    renderWithProviders(<UserManagement />)

    expect(screen.getAllByText('Клиент').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Владелец')).toBeInTheDocument()
    expect(screen.getByText('Администратор')).toBeInTheDocument()
  })

  it('renders status badges for active and blocked users', () => {
    vi.mocked(useGetAdminUsers).mockReturnValue({
      data: {
        data: mockUsers,
        success: true,
        meta: { page: 1, page_size: 20, total_count: 4, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminUsers>)

    renderWithProviders(<UserManagement />)

    expect(screen.getAllByText('Активен').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Заблокирован')).toBeInTheDocument()
  })

  it('renders block/unblock action buttons', () => {
    vi.mocked(useGetAdminUsers).mockReturnValue({
      data: {
        data: mockUsers,
        success: true,
        meta: { page: 1, page_size: 20, total_count: 4, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminUsers>)

    renderWithProviders(<UserManagement />)

    // 2 active non-admin users should have block button
    const blockButtons = screen.getAllByText('Заблокировать')
    expect(blockButtons.length).toBe(2)

    // 1 blocked user should have unblock button
    expect(screen.getByText('Разблокировать')).toBeInTheDocument()
  })

  it('does not show actions for admin users', () => {
    vi.mocked(useGetAdminUsers).mockReturnValue({
      data: {
        data: [mockUsers[3]],
        success: true,
        meta: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminUsers>)

    renderWithProviders(<UserManagement />)

    expect(screen.queryByText('Заблокировать')).not.toBeInTheDocument()
    expect(screen.queryByText('Разблокировать')).not.toBeInTheDocument()
  })

  it('renders search input', () => {
    vi.mocked(useGetAdminUsers).mockReturnValue({
      data: { data: [], success: true, meta: {} },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminUsers>)

    renderWithProviders(<UserManagement />)

    expect(screen.getByPlaceholderText('Поиск по имени, email или телефону')).toBeInTheDocument()
  })

  it('renders empty state', () => {
    vi.mocked(useGetAdminUsers).mockReturnValue({
      data: { data: [], success: true, meta: { page: 1, page_size: 20, total_count: 0, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminUsers>)

    renderWithProviders(<UserManagement />)

    expect(screen.getByText('Нет пользователей')).toBeInTheDocument()
  })

  it('renders total count in pagination', () => {
    vi.mocked(useGetAdminUsers).mockReturnValue({
      data: {
        data: mockUsers,
        success: true,
        meta: { page: 1, page_size: 20, total_count: 4, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminUsers>)

    renderWithProviders(<UserManagement />)

    expect(screen.getByText('Всего: 4')).toBeInTheDocument()
  })
})
