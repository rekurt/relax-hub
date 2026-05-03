import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import BroadcastCreate from '@/pages/crm/BroadcastCreate'

vi.mock('@/api/generated/crm/crm', () => ({
  useGetMyCrmSegments: vi.fn(),
  usePostMyCrmBroadcasts: vi.fn(),
}))

import {
  useGetMyCrmSegments,
  usePostMyCrmBroadcasts,
} from '@/api/generated/crm/crm'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU} theme={{ token: { motion: false } }}>
        <AntApp>
          <MemoryRouter>{ui}</MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockSegments = [
  { slug: 'new', name: 'Новые', count: 12 },
  { slug: 'regular', name: 'Постоянные', count: 45 },
  { slug: 'vip', name: 'VIP', count: 3 },
]

const mockCreateMutation = { mutate: vi.fn(), isPending: false }

describe('BroadcastCreate', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(usePostMyCrmBroadcasts).mockReturnValue(
      mockCreateMutation as unknown as ReturnType<typeof usePostMyCrmBroadcasts>,
    )
  })

  it('renders page title', () => {
    vi.mocked(useGetMyCrmSegments).mockReturnValue({
      data: { data: mockSegments },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmSegments>)

    renderWithProviders(<BroadcastCreate />)

    expect(screen.getByText('Новая рассылка')).toBeInTheDocument()
  })

  it('renders form fields', () => {
    vi.mocked(useGetMyCrmSegments).mockReturnValue({
      data: { data: mockSegments },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmSegments>)

    renderWithProviders(<BroadcastCreate />)

    expect(screen.getByText('Сегмент получателей')).toBeInTheDocument()
    expect(screen.getByText('Заголовок')).toBeInTheDocument()
    expect(screen.getByText('Текст сообщения')).toBeInTheDocument()
    expect(screen.getByText('Каналы отправки')).toBeInTheDocument()
  })

  it('renders channel checkboxes', () => {
    vi.mocked(useGetMyCrmSegments).mockReturnValue({
      data: { data: mockSegments },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmSegments>)

    renderWithProviders(<BroadcastCreate />)

    expect(screen.getByText('Push-уведомление')).toBeInTheDocument()
    expect(screen.getByText('Email')).toBeInTheDocument()
    expect(screen.getByText('Telegram')).toBeInTheDocument()
  })

  it('shows submit and cancel buttons', () => {
    vi.mocked(useGetMyCrmSegments).mockReturnValue({
      data: { data: mockSegments },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmSegments>)

    renderWithProviders(<BroadcastCreate />)

    expect(screen.getByText('Создать черновик')).toBeInTheDocument()
    expect(screen.getByText('Отмена')).toBeInTheDocument()
  })

  it('shows back button', () => {
    vi.mocked(useGetMyCrmSegments).mockReturnValue({
      data: { data: mockSegments },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmSegments>)

    renderWithProviders(<BroadcastCreate />)

    expect(screen.getByText('Назад')).toBeInTheDocument()
  })
})
