import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider, App as AntApp } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import Register from '@/pages/Register'

vi.mock('@/api/generated/auth/auth', () => ({
  postAuthRegister: vi.fn(),
  getAuthMe: vi.fn(),
}))

function renderRegister() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <MemoryRouter initialEntries={['/register']}>
          <AntApp>
            <Register />
          </AntApp>
        </MemoryRouter>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

describe('Register page', () => {
  it('renders registration form', () => {
    renderRegister()
    expect(screen.getByText('Регистрация клиента')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Имя')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Email')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Телефон')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Пароль')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Подтвердите пароль')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /зарегистрироваться/i })).toBeInTheDocument()
  })

  it('renders login link', () => {
    renderRegister()
    expect(screen.getByText('Войти')).toBeInTheDocument()
  })
})
