import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider, App as AntApp } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import Login from '@/pages/Login'
import Register from '@/pages/Register'

vi.mock('@/api/generated/auth/auth', () => ({
  postAuthLogin: vi.fn(),
  postAuthRegister: vi.fn(),
  getAuthMe: vi.fn(),
}))

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    get: vi.fn().mockResolvedValue({
      data: {
        success: true,
        data: { providers: ['vk', 'yandex', 'google'] },
      },
    }),
  },
}))

function renderInProviders(ui: React.ReactElement, route = '/login') {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <MemoryRouter initialEntries={[route]}>
          <AntApp>
            {ui}
          </AntApp>
        </MemoryRouter>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

describe('Login page OAuth buttons', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders VK, Yandex, Google OAuth buttons', async () => {
    renderInProviders(<Login />)
    expect(await screen.findByRole('button', { name: /войти через vk/i })).toBeInTheDocument()
    expect(await screen.findByRole('button', { name: /войти через яндекс/i })).toBeInTheDocument()
    expect(await screen.findByRole('button', { name: /войти через google/i })).toBeInTheDocument()
  })

  it('renders divider text', async () => {
    renderInProviders(<Login />)
    expect(await screen.findByText('или войдите через')).toBeInTheDocument()
  })

  it('OAuth buttons are clickable', async () => {
    renderInProviders(<Login />)
    const vkButton = await screen.findByRole('button', { name: /войти через vk/i })
    expect(vkButton).toBeEnabled()
    expect(vkButton).toHaveAttribute('type', 'button')
  })
})

describe('Register page OAuth buttons', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders VK, Yandex, Google OAuth buttons on register page', async () => {
    renderInProviders(<Register />, '/register')
    expect(await screen.findByRole('button', { name: /войти через vk/i })).toBeInTheDocument()
    expect(await screen.findByRole('button', { name: /войти через яндекс/i })).toBeInTheDocument()
    expect(await screen.findByRole('button', { name: /войти через google/i })).toBeInTheDocument()
  })

  it('still renders the regular registration form', () => {
    renderInProviders(<Register />, '/register')
    expect(screen.getByPlaceholderText('Имя')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Email')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /зарегистрироваться/i })).toBeInTheDocument()
  })
})
