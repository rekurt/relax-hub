import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider, App as AntApp } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import Login from '@/pages/Login'
import { vi } from 'vitest'

vi.mock('@/api/generated/auth/auth', () => ({
  postAuthLogin: vi.fn(),
  postAuthLoginPhone: vi.fn(),
  postAuthVerifyPhone: vi.fn(),
  getAuthMe: vi.fn(),
}))

vi.mock('@/api/generated/2fa/2fa', () => ({
  postAuth2faVerify: vi.fn(),
}))

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    get: vi.fn().mockResolvedValue({
      data: {
        success: true,
        data: { providers: [] },
      },
    }),
  },
}))

function renderLogin(initialRoute = '/login') {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <MemoryRouter initialEntries={[initialRoute]}>
          <AntApp>
            <Login />
          </AntApp>
        </MemoryRouter>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

describe('Login page', () => {
  it('renders login form with email tab by default', () => {
    renderLogin()
    expect(screen.getByText('Вход в личный кабинет')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Email')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Пароль')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /^войти$/i })).toBeInTheDocument()
  })

  it('renders registration link', () => {
    renderLogin()
    expect(screen.getByText('Зарегистрироваться')).toBeInTheDocument()
  })

  it('renders forgot password link', () => {
    renderLogin()
    expect(screen.getByText('Забыли пароль?')).toBeInTheDocument()
  })

  it('renders email/phone tab switcher', () => {
    renderLogin()
    expect(screen.getByText('Email')).toBeInTheDocument()
    expect(screen.getByText('Телефон')).toBeInTheDocument()
  })

  it('switches to phone login mode', () => {
    renderLogin()
    fireEvent.click(screen.getByText('Телефон'))
    expect(screen.getByPlaceholderText('Телефон')).toBeInTheDocument()
    expect(screen.getByText('Получить код')).toBeInTheDocument()
    expect(screen.queryByPlaceholderText('Пароль')).not.toBeInTheDocument()
  })
})
