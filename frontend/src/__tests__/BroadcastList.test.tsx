import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import BroadcastList from '@/pages/crm/BroadcastList'

vi.mock('@/api/generated/crm/crm', () => ({
  useGetMyCrmBroadcasts: vi.fn(),
  usePostMyCrmBroadcastsIdSend: vi.fn(),
}))

import {
  useGetMyCrmBroadcasts,
  usePostMyCrmBroadcastsIdSend,
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

const mockBroadcasts = [
  {
    id: 'bc-1',
    title: 'Акция на выходные',
    segment: 'regular',
    channels: ['push', 'email'],
    status: 'sent',
    delivered: 150,
    read: 85,
    sent_at: '2026-03-20T10:00:00Z',
    created_at: '2026-03-20T09:00:00Z',
  },
  {
    id: 'bc-2',
    title: 'Новая услуга',
    segment: 'vip',
    channels: ['push'],
    status: 'draft',
    delivered: 0,
    read: 0,
    sent_at: null,
    created_at: '2026-03-25T10:00:00Z',
  },
]

const mockSendMutation = { mutate: vi.fn(), isPending: false }

describe('BroadcastList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(usePostMyCrmBroadcastsIdSend).mockReturnValue(
      mockSendMutation as unknown as ReturnType<typeof usePostMyCrmBroadcastsIdSend>,
    )
  })

  it('renders broadcast list with title', () => {
    vi.mocked(useGetMyCrmBroadcasts).mockReturnValue({
      data: { data: mockBroadcasts, success: true, meta: { total_count: 2 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmBroadcasts>)

    renderWithProviders(<BroadcastList />)

    expect(screen.getByText('Рассылки')).toBeInTheDocument()
  })

  it('renders broadcast rows', () => {
    vi.mocked(useGetMyCrmBroadcasts).mockReturnValue({
      data: { data: mockBroadcasts, success: true, meta: { total_count: 2 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmBroadcasts>)

    renderWithProviders(<BroadcastList />)

    expect(screen.getByText('Акция на выходные')).toBeInTheDocument()
    expect(screen.getByText('Новая услуга')).toBeInTheDocument()
  })

  it('renders status tags', () => {
    vi.mocked(useGetMyCrmBroadcasts).mockReturnValue({
      data: { data: mockBroadcasts, success: true, meta: { total_count: 2 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmBroadcasts>)

    renderWithProviders(<BroadcastList />)

    expect(screen.getByText('Отправлено')).toBeInTheDocument()
    expect(screen.getByText('Черновик')).toBeInTheDocument()
  })

  it('shows delivery stats', () => {
    vi.mocked(useGetMyCrmBroadcasts).mockReturnValue({
      data: { data: mockBroadcasts, success: true, meta: { total_count: 2 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmBroadcasts>)

    renderWithProviders(<BroadcastList />)

    expect(screen.getByText('150 / 85 / 0')).toBeInTheDocument()
  })

  it('shows send button for draft broadcasts', () => {
    vi.mocked(useGetMyCrmBroadcasts).mockReturnValue({
      data: { data: mockBroadcasts, success: true, meta: { total_count: 2 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmBroadcasts>)

    renderWithProviders(<BroadcastList />)

    expect(screen.getByText('Отправить')).toBeInTheDocument()
  })

  it('shows create button', () => {
    vi.mocked(useGetMyCrmBroadcasts).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmBroadcasts>)

    renderWithProviders(<BroadcastList />)

    expect(screen.getByText('Создать рассылку')).toBeInTheDocument()
  })

  it('shows empty state', () => {
    vi.mocked(useGetMyCrmBroadcasts).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmBroadcasts>)

    renderWithProviders(<BroadcastList />)

    expect(screen.getByText(/Нет рассылок/)).toBeInTheDocument()
  })
})
