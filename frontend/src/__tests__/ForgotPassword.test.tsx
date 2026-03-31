import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider, App as AntApp } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import ForgotPassword from '@/pages/ForgotPassword'

const mockPostAuthForgotPassword = vi.fn()
vi.mock('@/api/generated/auth/auth', () => ({
  postAuthForgotPassword: (...args: unknown[]) => mockPostAuthForgotPassword(...args),
}))

function renderForgotPassword() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <MemoryRouter initialEntries={['/forgot-password']}>
          <AntApp>
            <ForgotPassword />
          </AntApp>
        </MemoryRouter>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

describe('ForgotPassword page', () => {
  beforeEach(() => {
    mockPostAuthForgotPassword.mockReset()
  })

  it('renders forgot password form', () => {
    renderForgotPassword()
    expect(screen.getByText('Восстановление пароля')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Email')).toBeInTheDocument()
    expect(screen.getByText('Отправить ссылку')).toBeInTheDocument()
  })

  it('renders back to login link', () => {
    renderForgotPassword()
    expect(screen.getByText('Вернуться к входу')).toBeInTheDocument()
  })

  it('shows success message after sending', async () => {
    mockPostAuthForgotPassword.mockResolvedValue({ success: true })
    renderForgotPassword()

    fireEvent.change(screen.getByPlaceholderText('Email'), { target: { value: 'test@example.com' } })
    fireEvent.click(screen.getByText('Отправить ссылку'))

    await waitFor(() => {
      expect(screen.getByText('Письмо отправлено')).toBeInTheDocument()
    })
    expect(mockPostAuthForgotPassword).toHaveBeenCalledWith({ email: 'test@example.com' })
  })

  it('shows return to login button after success', async () => {
    mockPostAuthForgotPassword.mockResolvedValue({ success: true })
    renderForgotPassword()

    fireEvent.change(screen.getByPlaceholderText('Email'), { target: { value: 'test@example.com' } })
    fireEvent.click(screen.getByText('Отправить ссылку'))

    await waitFor(() => {
      expect(screen.getByText('Вернуться к входу')).toBeInTheDocument()
    })
  })
})
