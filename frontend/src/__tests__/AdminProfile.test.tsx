import { fireEvent, render, screen, waitFor } from '@testing-library/react'
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

vi.mock('@/api/generated/2fa/2fa', () => ({
  usePostAuth2faTotpEnable: vi.fn(),
  usePostAuth2faTotpVerify: vi.fn(),
  useDeleteAuth2faTotp: vi.fn(),
  usePostAuth2faSmsEnable: vi.fn(),
}))

import { useAuthStore } from '@/stores/auth'
import {
  usePutAuthMe,
  usePostAuthMeAvatar,
  useDeleteAuthMeAvatar,
} from '@/api/generated/auth/auth'
import {
  usePostAuth2faTotpEnable,
  usePostAuth2faTotpVerify,
  useDeleteAuth2faTotp,
  usePostAuth2faSmsEnable,
} from '@/api/generated/2fa/2fa'
import { ADMIN_2FA_START_EVENT } from '@/lib/admin2faNotice'

const mockUser = {
  id: 'admin-1',
  email: 'admin@relaxhub.ru',
  name: 'Иван Админов',
  phone: '+7 999 000-00-00',
  role: 'admin',
  two_fa_method: 'none',
  avatar_url: 'https://example.com/avatar.jpg',
}

function renderWithProviders(ui: React.ReactElement, route = '/admin/profile') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={[route]}>
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
const mockEnableTotp = { mutate: vi.fn(), isPending: false }
const mockVerifyTotp = { mutate: vi.fn(), isPending: false }
const mockDisableTotp = { mutate: vi.fn(), isPending: false }
const mockEnableSms = { mutate: vi.fn(), isPending: false }

describe('AdminProfile', () => {
  beforeEach(() => {
    Element.prototype.scrollIntoView = vi.fn()
    mockUpdateProfile.mutate.mockReset()
    mockUploadAvatar.mutate.mockReset()
    mockDeleteAvatar.mutate.mockReset()
    mockEnableTotp.mutate.mockReset()
    mockVerifyTotp.mutate.mockReset()
    mockDisableTotp.mutate.mockReset()
    mockEnableSms.mutate.mockReset()

    vi.mocked(useAuthStore).mockReturnValue({
      user: mockUser,
      loadProfile: vi.fn(),
      setUser: vi.fn(),
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
    vi.mocked(usePostAuth2faTotpEnable).mockReturnValue(
      mockEnableTotp as unknown as ReturnType<typeof usePostAuth2faTotpEnable>,
    )
    vi.mocked(usePostAuth2faTotpVerify).mockReturnValue(
      mockVerifyTotp as unknown as ReturnType<typeof usePostAuth2faTotpVerify>,
    )
    vi.mocked(useDeleteAuth2faTotp).mockReturnValue(
      mockDisableTotp as unknown as ReturnType<typeof useDeleteAuth2faTotp>,
    )
    vi.mocked(usePostAuth2faSmsEnable).mockReturnValue(
      mockEnableSms as unknown as ReturnType<typeof usePostAuth2faSmsEnable>,
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

  it('focuses the TOTP setup action from the 2FA deep link', async () => {
    renderWithProviders(<AdminProfile />, '/admin/profile?focus2fa=1#two-factor')

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Настроить TOTP/i })).toHaveFocus()
    })
    expect(Element.prototype.scrollIntoView).toHaveBeenCalled()
  })

  it('focuses the TOTP setup action from the admin 2FA banner event', async () => {
    renderWithProviders(<AdminProfile />)

    window.dispatchEvent(new CustomEvent(ADMIN_2FA_START_EVENT))

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Настроить TOTP/i })).toHaveFocus()
    })
    expect(mockEnableTotp.mutate).toHaveBeenCalledTimes(1)
    expect(Element.prototype.scrollIntoView).toHaveBeenCalled()
  })
})
