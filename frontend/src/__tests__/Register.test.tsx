import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider, App as AntApp } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import Register from '@/pages/Register'

vi.mock('@/api/generated/auth/auth', () => ({
  postAuthRegister: vi.fn(),
  postAuthRegisterPhone: vi.fn(),
  postAuthVerifyPhone: vi.fn(),
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
  it('renders registration form with email tab', () => {
    renderRegister()
    expect(screen.getByText('Регистрация клиента')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Имя')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Email')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Телефон')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Пароль')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Подтвердите пароль')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /зарегистрироваться/i })).toBeInTheDocument()
  })

  it('renders age confirmation checkbox', () => {
    renderRegister()
    expect(screen.getByText('Мне исполнилось 18 лет')).toBeInTheDocument()
  })

  it('renders login link', () => {
    renderRegister()
    expect(screen.getByText('Войти')).toBeInTheDocument()
  })

  it('renders email/phone tab switcher', () => {
    renderRegister()
    expect(screen.getByText('Email')).toBeInTheDocument()
    expect(screen.getByText('Телефон')).toBeInTheDocument()
  })

  it('switches to phone registration mode', () => {
    renderRegister()
    fireEvent.click(screen.getByText('Телефон'))
    expect(screen.getByPlaceholderText('Имя')).toBeInTheDocument()
    expect(screen.getByText('Мне исполнилось 18 лет')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Телефон')).toBeInTheDocument()
    expect(screen.getByText('Получить код')).toBeInTheDocument()
    expect(screen.queryByPlaceholderText('Пароль')).not.toBeInTheDocument()
    expect(screen.queryByPlaceholderText('Email')).not.toBeInTheDocument()
  })

  it('renders role switcher', () => {
    renderRegister()
    expect(screen.getByText('Клиент')).toBeInTheDocument()
    expect(screen.getByText('Владелец бани')).toBeInTheDocument()
  })

  it('switches role description', () => {
    renderRegister()
    fireEvent.click(screen.getByText('Владелец бани'))
    expect(screen.getByText('Регистрация владельца')).toBeInTheDocument()
  })
})
