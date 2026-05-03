import { fireEvent, render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import AdminProfile from '@/pages/admin/AdminProfile'

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn(),
}))

vi.mock('@/api/generated/auth/auth', () => ({
  usePutAuthMe: vi.fn(),
  usePostAuthMeAvatar: vi.fn(),
  useDeleteAuthMeAvatar: vi.fn(),
}))

import { useAuthStore } from '@/stores/auth'
import {
  usePutAuthMe,
  usePostAuthMeAvatar,
  useDeleteAuthMeAvatar,
} from '@/api/generated/auth/auth'

const mockUser = {
  id: 'admin-1',
  email: 'admin@relaxhub.ru',
  name: 'Иван Админов',
  phone: '+7 999 000-00-00',
  role: 'admin',
  avatar_url: 'https://example.com/avatar.jpg',
}

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/profile']}>
            <Routes>
              <Route path="/admin/profile" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockUpdateProfile = { mutate: vi.fn(), isPending: false }
const mockUploadAvatar = { mutate: vi.fn(), isPending: false }
const mockDeleteAvatar = { mutate: vi.fn(), isPending: false }

describe('AdminProfile', () => {
  beforeEach(() => {
    vi.mocked(useAuthStore).mockReturnValue({
      user: mockUser,
      loadProfile: vi.fn(),
    } as unknown as ReturnType<typeof useAuthStore>)
    vi.mocked(usePutAuthMe).mockReturnValue(
      mockUpdateProfile as unknown as ReturnType<typeof usePutAuthMe>,
    )
    vi.mocked(usePostAuthMeAvatar).mockReturnValue(
      mockUploadAvatar as unknown as ReturnType<typeof usePostAuthMeAvatar>,
    )
    vi.mocked(useDeleteAuthMeAvatar).mockReturnValue(
      mockDeleteAvatar as unknown as ReturnType<typeof useDeleteAuthMeAvatar>,
    )
  })

  it('renders page title', () => {
    renderWithProviders(<AdminProfile />)
    expect(screen.getByText('Профиль администратора')).toBeInTheDocument()
  })

  it('renders avatar section', () => {
    renderWithProviders(<AdminProfile />)
    expect(screen.getByText('Аватар')).toBeInTheDocument()
    expect(screen.getByText('Загрузить')).toBeInTheDocument()
    expect(screen.getByText('Удалить')).toBeInTheDocument()
  })

  it('renders profile form with user data', () => {
    renderWithProviders(<AdminProfile />)
    expect(screen.getByText('Основная информация')).toBeInTheDocument()
    expect(screen.getByDisplayValue('admin@relaxhub.ru')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Иван Админов')).toBeInTheDocument()
    expect(screen.getByDisplayValue('+7 999 000-00-00')).toBeInTheDocument()
  })

  it('renders email as disabled field', () => {
    renderWithProviders(<AdminProfile />)
    const emailInput = screen.getByDisplayValue('admin@relaxhub.ru')
    expect(emailInput).toBeDisabled()
  })

  it('renders save button', () => {
    renderWithProviders(<AdminProfile />)
    expect(screen.getByText('Сохранить')).toBeInTheDocument()
  })

  it('does not show delete avatar when no avatar', () => {
    vi.mocked(useAuthStore).mockReturnValue({
      user: { ...mockUser, avatar_url: undefined },
      loadProfile: vi.fn(),
    } as unknown as ReturnType<typeof useAuthStore>)

    renderWithProviders(<AdminProfile />)
    expect(screen.queryByText('Удалить')).not.toBeInTheDocument()
  })

  it('opens avatar delete confirmation as an overlay', () => {
    renderWithProviders(<AdminProfile />)

    fireEvent.click(screen.getByRole('button', { name: /Удалить/i }))

    const confirmation = screen.getByText('Удалить аватар?')
    const overlay = confirmation.closest('.ant-popover-inner')

    expect(overlay).toHaveStyle({ position: 'absolute' })
  })
})
