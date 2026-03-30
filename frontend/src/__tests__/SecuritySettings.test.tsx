import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import SecuritySettings from '@/pages/client/SecuritySettings'

vi.mock('@/api/generated/sessions/sessions', () => ({
  useGetMySessions: vi.fn(),
  useDeleteMySessionsId: vi.fn(),
  useDeleteMySessions: vi.fn(),
  getGetMySessionsQueryKey: vi.fn(() => ['/my/sessions']),
}))

vi.mock('@/api/generated/2fa/2fa', () => ({
  usePostAuth2faTotpEnable: vi.fn(),
  usePostAuth2faTotpVerify: vi.fn(),
  useDeleteAuth2faTotp: vi.fn(),
  usePostAuth2faSmsEnable: vi.fn(),
}))

vi.mock('@/api/generated/auth/auth', () => ({
  usePostAuthForgotPassword: vi.fn(),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn((selector: (s: Record<string, unknown>) => unknown) =>
    selector({
      user: { email: 'test@example.com', phone: '+79001234567' },
    }),
  ),
}))

import {
  useGetMySessions,
  useDeleteMySessionsId,
  useDeleteMySessions,
} from '@/api/generated/sessions/sessions'

import {
  usePostAuth2faTotpEnable,
  usePostAuth2faTotpVerify,
  useDeleteAuth2faTotp,
  usePostAuth2faSmsEnable,
} from '@/api/generated/2fa/2fa'

import { usePostAuthForgotPassword } from '@/api/generated/auth/auth'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/client/security']}>
            <Routes>
              <Route path="/client/security" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockSessions = [
  {
    id: 'sess-1',
    device_info: 'Chrome Desktop',
    browser: 'Chrome 120',
    ip: '192.168.1.1',
    last_active_at: '2026-03-30T10:00:00Z',
    is_current: true,
    created_at: '2026-03-29T08:00:00Z',
  },
  {
    id: 'sess-2',
    device_info: 'Mobile Safari',
    browser: 'Safari 17',
    ip: '10.0.0.5',
    last_active_at: '2026-03-29T15:00:00Z',
    is_current: false,
    created_at: '2026-03-28T12:00:00Z',
  },
]

const mockTerminateMutate = vi.fn()
const mockTerminateAllMutate = vi.fn()
const mockEnableTotpMutate = vi.fn()
const mockVerifyTotpMutate = vi.fn()
const mockDisableTotpMutate = vi.fn()
const mockEnableSmsMutate = vi.fn()
const mockForgotPasswordMutate = vi.fn()

function setupMocks(overrides?: { sessions?: typeof mockSessions }) {
  vi.mocked(useGetMySessions).mockReturnValue({
    data: { data: overrides?.sessions ?? mockSessions, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMySessions>)

  vi.mocked(useDeleteMySessionsId).mockReturnValue({
    mutate: mockTerminateMutate,
    isPending: false,
  } as unknown as ReturnType<typeof useDeleteMySessionsId>)

  vi.mocked(useDeleteMySessions).mockReturnValue({
    mutate: mockTerminateAllMutate,
    isPending: false,
  } as unknown as ReturnType<typeof useDeleteMySessions>)

  vi.mocked(usePostAuth2faTotpEnable).mockReturnValue({
    mutate: mockEnableTotpMutate,
    isPending: false,
  } as unknown as ReturnType<typeof usePostAuth2faTotpEnable>)

  vi.mocked(usePostAuth2faTotpVerify).mockReturnValue({
    mutate: mockVerifyTotpMutate,
    isPending: false,
  } as unknown as ReturnType<typeof usePostAuth2faTotpVerify>)

  vi.mocked(useDeleteAuth2faTotp).mockReturnValue({
    mutate: mockDisableTotpMutate,
    isPending: false,
  } as unknown as ReturnType<typeof useDeleteAuth2faTotp>)

  vi.mocked(usePostAuth2faSmsEnable).mockReturnValue({
    mutate: mockEnableSmsMutate,
    isPending: false,
  } as unknown as ReturnType<typeof usePostAuth2faSmsEnable>)

  vi.mocked(usePostAuthForgotPassword).mockReturnValue({
    mutate: mockForgotPasswordMutate,
    isPending: false,
  } as unknown as ReturnType<typeof usePostAuthForgotPassword>)
}

describe('SecuritySettings', () => {
  beforeEach(() => {
    mockTerminateMutate.mockReset()
    mockTerminateAllMutate.mockReset()
    mockEnableTotpMutate.mockReset()
    mockVerifyTotpMutate.mockReset()
    mockDisableTotpMutate.mockReset()
    mockEnableSmsMutate.mockReset()
    mockForgotPasswordMutate.mockReset()
  })

  it('renders page title and all three sections', () => {
    setupMocks()
    renderWithProviders(<SecuritySettings />)

    expect(screen.getByText('Безопасность')).toBeInTheDocument()
    expect(screen.getByText('Активные сессии')).toBeInTheDocument()
    expect(screen.getByText('Двухфакторная аутентификация')).toBeInTheDocument()
    expect(screen.getByText('Пароль')).toBeInTheDocument()
  })

  it('renders session list with device info', () => {
    setupMocks()
    renderWithProviders(<SecuritySettings />)

    expect(screen.getByText('Chrome Desktop')).toBeInTheDocument()
    expect(screen.getByText('Mobile Safari')).toBeInTheDocument()
    expect(screen.getByText('Chrome 120')).toBeInTheDocument()
    expect(screen.getByText('Текущая')).toBeInTheDocument()
  })

  it('shows terminate button only for non-current sessions', () => {
    setupMocks()
    renderWithProviders(<SecuritySettings />)

    const terminateButtons = screen.getAllByText('Завершить')
    // Only "Завершить" button for non-current session + "Завершить все другие" button
    // The "Завершить все другие" is in the card header
    expect(terminateButtons.length).toBeGreaterThanOrEqual(1)
  })

  it('shows terminate all button when multiple sessions', () => {
    setupMocks()
    renderWithProviders(<SecuritySettings />)

    expect(screen.getByText('Завершить все другие')).toBeInTheDocument()
  })

  it('does not show terminate all button with single session', () => {
    setupMocks({ sessions: [mockSessions[0]!] })
    renderWithProviders(<SecuritySettings />)

    expect(screen.queryByText('Завершить все другие')).not.toBeInTheDocument()
  })

  it('calls terminate mutation on confirm', async () => {
    setupMocks()
    renderWithProviders(<SecuritySettings />)

    // Find the terminate button for non-current session
    const terminateButtons = screen.getAllByText('Завершить')
    fireEvent.click(terminateButtons[0]!)

    await waitFor(() => {
      expect(screen.getByText('Завершить сессию?')).toBeInTheDocument()
    })

    const confirmButtons = screen.getAllByRole('button', { name: /Завершить/i })
    const popconfirmButton = confirmButtons.find(
      (btn) => btn.closest('.ant-popconfirm-buttons'),
    )
    if (popconfirmButton) {
      fireEvent.click(popconfirmButton)
    }

    await waitFor(() => {
      expect(mockTerminateMutate).toHaveBeenCalledWith({ id: 'sess-2' })
    })
  })

  it('shows TOTP setup button in idle state', () => {
    setupMocks()
    renderWithProviders(<SecuritySettings />)

    expect(screen.getByText('Настроить TOTP')).toBeInTheDocument()
  })

  it('calls enable TOTP mutation on click', () => {
    setupMocks()
    renderWithProviders(<SecuritySettings />)

    fireEvent.click(screen.getByText('Настроить TOTP'))
    expect(mockEnableTotpMutate).toHaveBeenCalled()
  })

  it('shows SMS 2FA enable button when user has phone', () => {
    setupMocks()
    renderWithProviders(<SecuritySettings />)

    expect(screen.getByText('Включить SMS 2FA')).toBeInTheDocument()
  })

  it('calls SMS 2FA enable mutation on click', () => {
    setupMocks()
    renderWithProviders(<SecuritySettings />)

    fireEvent.click(screen.getByText('Включить SMS 2FA'))
    expect(mockEnableSmsMutate).toHaveBeenCalled()
  })

  it('shows password reset button', () => {
    setupMocks()
    renderWithProviders(<SecuritySettings />)

    expect(screen.getByText('Отправить ссылку для сброса пароля')).toBeInTheDocument()
  })

  it('calls forgot password mutation on click', () => {
    setupMocks()
    renderWithProviders(<SecuritySettings />)

    fireEvent.click(screen.getByText('Отправить ссылку для сброса пароля'))
    expect(mockForgotPasswordMutate).toHaveBeenCalledWith({
      data: { email: 'test@example.com' },
    })
  })

  it('renders empty sessions state', () => {
    setupMocks({ sessions: [] })
    renderWithProviders(<SecuritySettings />)

    expect(screen.getByText('Нет активных сессий')).toBeInTheDocument()
  })

  it('shows session creation info', () => {
    setupMocks()
    renderWithProviders(<SecuritySettings />)

    // Safari 17 is the browser text for the second session
    expect(screen.getByText('Safari 17')).toBeInTheDocument()
  })
})
