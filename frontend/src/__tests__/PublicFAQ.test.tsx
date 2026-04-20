import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, beforeEach, vi } from 'vitest'
import PublicFAQ from '@/pages/public/PublicFAQ'

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    get: vi.fn(),
  },
}))

import { axiosInstance } from '@/api/axios-instance'

function renderWithProviders() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter>
            <PublicFAQ />
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

describe('PublicFAQ', () => {
  beforeEach(() => {
    vi.mocked(axiosInstance.get).mockRejectedValue(new Error('faq unavailable'))
  })

  it('renders a full help center with real fallback scenarios and support channels', async () => {
    renderWithProviders()

    expect(await screen.findByText('Чем можем помочь прямо сейчас')).toBeInTheDocument()
    expect(screen.getByText('Как проходит бронирование')).toBeInTheDocument()
    expect(screen.getByText('Как забронировать баню?')).toBeInTheDocument()
    expect(screen.getByText('Какие способы оплаты поддерживаются?')).toBeInTheDocument()
    expect(screen.getByText('Как оставить отзыв?')).toBeInTheDocument()
    expect(screen.getByText('support@relaxhub.ru')).toBeInTheDocument()
    expect(screen.getByText('+7 (495) 555-21-21')).toBeInTheDocument()
  })
})
