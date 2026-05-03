import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import OAuthCallback from '@/pages/OAuthCallback'

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

const mockSetAuth = vi.fn()
vi.mock('@/stores/auth', () => ({
  useAuthStore: (selector: (s: Record<string, unknown>) => unknown) =>
    selector({ setAuth: mockSetAuth }),
  getRoleHomePath: (role?: string) => {
    if (role === 'client') return '/client'
    if (role === 'admin') return '/admin'
    return '/'
  },
}))

const mockGetCallback = vi.fn()
vi.mock('@/api/generated/oauth/oauth', () => ({
  getAuthOauthProviderCallback: (...args: unknown[]) => mockGetCallback(...args),
}))

function renderCallback(route: string) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <MemoryRouter initialEntries={[route]}>
          <Routes>
            <Route path="/auth/oauth/callback/:provider" element={<OAuthCallback />} />
          </Routes>
        </MemoryRouter>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

describe('OAuthCallback', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading spinner during OAuth processing', () => {
    mockGetCallback.mockReturnValue(new Promise(() => {}))
    renderCallback('/auth/oauth/callback/vk?code=abc&state=xyz')
    expect(screen.getByText(/авторизация через vk/i)).toBeInTheDocument()
  })

  it('calls backend callback API with code and state', async () => {
    mockGetCallback.mockResolvedValue({
      success: true,
      data: { token: 'jwt-token', user: { id: '1', role: 'client', name: 'Test' } },
    })

    renderCallback('/auth/oauth/callback/vk?code=auth-code&state=state-123')

    await waitFor(() => {
      expect(mockGetCallback).toHaveBeenCalledWith('vk', { code: 'auth-code', state: 'state-123' })
    })
  })

  it('sets auth and navigates on successful callback', async () => {
    mockGetCallback.mockResolvedValue({
      success: true,
      data: { token: 'jwt-token', user: { id: '1', role: 'client', name: 'Test' } },
    })

    renderCallback('/auth/oauth/callback/google?code=abc&state=xyz')

    await waitFor(() => {
      expect(mockSetAuth).toHaveBeenCalledWith('jwt-token', { id: '1', role: 'client', name: 'Test' })
      expect(mockNavigate).toHaveBeenCalledWith('/client', { replace: true })
    })
  })

  it('shows error when OAuth returns error parameter', () => {
    renderCallback('/auth/oauth/callback/vk?error=access_denied&error_description=User+denied')
    expect(screen.getByText('Ошибка авторизации')).toBeInTheDocument()
    expect(screen.getByText('User denied')).toBeInTheDocument()
  })

  it('shows error when code or state is missing', () => {
    renderCallback('/auth/oauth/callback/vk?code=abc')
    expect(screen.getByText('Отсутствуют необходимые параметры авторизации')).toBeInTheDocument()
  })

  it('shows error on API failure', async () => {
    mockGetCallback.mockRejectedValue(new Error('Network error'))

    renderCallback('/auth/oauth/callback/yandex?code=abc&state=xyz')

    await waitFor(() => {
      expect(screen.getByText('Не удалось завершить авторизацию. Попробуйте снова.')).toBeInTheDocument()
    })
  })

  it('shows link back to login on error', () => {
    renderCallback('/auth/oauth/callback/vk?error=access_denied')
    expect(screen.getByText('Вернуться на страницу входа')).toBeInTheDocument()
  })
})
