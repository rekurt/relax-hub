import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import RoleManagement from '@/pages/admin/RoleManagement'

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    get: vi.fn(),
    put: vi.fn(),
  },
}))

import { axiosInstance } from '@/api/axios-instance'

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

const mockAdmins = [
  {
    id: 'admin-1',
    email: 'admin@relaxhub.ru',
    name: 'Алексей Админов',
    admin_sub_role: 'super_admin',
    two_fa_method: 'totp',
    is_active: true,
  },
  {
    id: 'admin-2',
    email: 'mod@relaxhub.ru',
    name: 'Мария Модератор',
    admin_sub_role: 'moderator',
    two_fa_method: 'none',
    is_active: true,
  },
  {
    id: 'admin-3',
    email: 'support@relaxhub.ru',
    name: 'Иван Поддержкин',
    admin_sub_role: 'support_l1',
    two_fa_method: 'sms',
    is_active: false,
  },
]

const mockMatrix = {
  roles: ['super_admin', 'moderator', 'support_l1'],
  permissions: ['users.manage', 'bathhouses.moderate', 'tickets.manage'],
  matrix: {
    super_admin: ['users.manage', 'bathhouses.moderate', 'tickets.manage'],
    moderator: ['bathhouses.moderate'],
    support_l1: ['tickets.manage'],
  },
}

function setupMocks(overrides?: { admins?: typeof mockAdmins; matrix?: typeof mockMatrix | null; reject?: boolean }) {
  vi.mocked(axiosInstance.get).mockImplementation((url: string) => {
    if (overrides?.reject) return Promise.reject(new Error('fail'))
    if (url.includes('/admin/roles/permissions')) {
      return Promise.resolve({ data: { data: overrides?.matrix ?? mockMatrix, success: true } })
    }
    if (url.includes('/admin/roles')) {
      return Promise.resolve({ data: { data: overrides?.admins ?? mockAdmins, success: true } })
    }
    return Promise.reject(new Error('not found'))
  })
}

describe('RoleManagement', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders page title after loading', async () => {
    setupMocks()
    renderWithProviders(<RoleManagement />)

    await waitFor(() => {
      expect(screen.getByText('Управление ролями администраторов')).toBeInTheDocument()
    })
  })

  it('shows loading spinner initially', () => {
    vi.mocked(axiosInstance.get).mockReturnValue(new Promise(() => {}))
    renderWithProviders(<RoleManagement />)

    expect(document.querySelector('.ant-spin')).toBeInTheDocument()
  })

  it('displays admin users in table', async () => {
    setupMocks()
    renderWithProviders(<RoleManagement />)

    await waitFor(() => {
      expect(screen.getByText('Алексей Админов')).toBeInTheDocument()
    })

    expect(screen.getByText('admin@relaxhub.ru')).toBeInTheDocument()
    expect(screen.getByText('Мария Модератор')).toBeInTheDocument()
    expect(screen.getByText('mod@relaxhub.ru')).toBeInTheDocument()
    expect(screen.getByText('Иван Поддержкин')).toBeInTheDocument()
  })

  it('shows 2FA status badges correctly', async () => {
    setupMocks()
    renderWithProviders(<RoleManagement />)

    await waitFor(() => {
      expect(screen.getByText('Алексей Админов')).toBeInTheDocument()
    })

    expect(screen.getByText('TOTP')).toBeInTheDocument()
    expect(screen.getByText('SMS')).toBeInTheDocument()
    expect(screen.getByText('Не настроена')).toBeInTheDocument()
  })

  it('shows active/blocked status tags', async () => {
    setupMocks()
    renderWithProviders(<RoleManagement />)

    await waitFor(() => {
      expect(screen.getByText('Алексей Админов')).toBeInTheDocument()
    })

    const activeTags = screen.getAllByText('Активен')
    expect(activeTags.length).toBe(2)
    expect(screen.getByText('Заблокирован')).toBeInTheDocument()
  })

  it('renders permissions matrix card', async () => {
    setupMocks()
    renderWithProviders(<RoleManagement />)

    await waitFor(() => {
      expect(screen.getByText('Матрица разрешений')).toBeInTheDocument()
    })

    expect(screen.getByText('Пользователи')).toBeInTheDocument()
    expect(screen.getByText('Модерация бань')).toBeInTheDocument()
    expect(screen.getByText('Тикеты')).toBeInTheDocument()
  })
})
