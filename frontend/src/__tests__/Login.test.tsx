import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider, App as AntApp } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import Login from '@/pages/Login'

vi.mock('@/api/generated/auth/auth', () => ({
  postAuthLogin: vi.fn(),
  getAuthMe: vi.fn(),
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
  it('renders login form', () => {
    renderLogin()
    expect(screen.getByText('Вход в личный кабинет')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Email')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Пароль')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /войти/i })).toBeInTheDocument()
  })

  it('renders registration link', () => {
    renderLogin()
    expect(screen.getByText('Зарегистрироваться')).toBeInTheDocument()
  })
})
