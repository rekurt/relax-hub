import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ClientChat from '@/pages/client/ClientChat'

vi.mock('@/api/generated/chat/chat', () => ({
  useGetMyConversations: vi.fn(),
  useGetConversationsIdMessages: vi.fn(),
  usePostConversationsIdMessages: vi.fn(),
  usePatchConversationsIdRead: vi.fn(),
  useGetMyUnreadMessagesCount: vi.fn(),
  getGetConversationsIdMessagesQueryKey: vi.fn(() => ['messages']),
  getGetMyConversationsQueryKey: vi.fn(() => ['conversations']),
  getGetMyUnreadMessagesCountQueryKey: vi.fn(() => ['unread']),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn((selector) =>
    selector({
      user: { id: 'client-1', name: 'Client', email: 'client@test.com', role: 'client' },
      token: 'test-token',
      isAuthenticated: true,
      isLoading: false,
    }),
  ),
}))

vi.mock('@/lib/useWebSocketNotifications', () => ({
  useWebSocketNotifications: vi.fn(),
}))

import {
  useGetMyConversations,
  useGetConversationsIdMessages,
  usePostConversationsIdMessages,
  usePatchConversationsIdRead,
} from '@/api/generated/chat/chat'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter>{ui}</MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockConversations = [
  {
    id: 'conv-1',
    bathhouse_id: 'bath-1',
    client_id: 'client-aaa',
    created_at: '2026-03-10T14:00:00Z',
    last_message_at: '2026-03-14T18:00:00Z',
  },
  {
    id: 'conv-2',
    bathhouse_id: 'bath-2',
    client_id: 'client-bbb',
    created_at: '2026-03-09T12:00:00Z',
    last_message_at: '2026-03-13T10:00:00Z',
  },
]

const mockMessages = [
  {
    id: 'msg-1',
    conversation_id: 'conv-1',
    sender_id: 'owner-1',
    text: 'Добро пожаловать!',
    is_read: true,
    created_at: '2026-03-14T17:00:00Z',
  },
  {
    id: 'msg-2',
    conversation_id: 'conv-1',
    sender_id: 'client-1',
    text: 'Спасибо! Хочу забронировать.',
    is_read: true,
    created_at: '2026-03-14T17:05:00Z',
  },
]

const mockSendMutation = { mutate: vi.fn(), isPending: false }
const mockMarkReadMutation = { mutate: vi.fn(), isPending: false }

describe('ClientChat', () => {
  beforeEach(() => {
    vi.mocked(usePostConversationsIdMessages).mockReturnValue(
      mockSendMutation as unknown as ReturnType<typeof usePostConversationsIdMessages>,
    )
    vi.mocked(usePatchConversationsIdRead).mockReturnValue(
      mockMarkReadMutation as unknown as ReturnType<typeof usePatchConversationsIdRead>,
    )
    vi.mocked(useGetConversationsIdMessages).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetConversationsIdMessages>)
  })

  it('renders chat page title', () => {
    vi.mocked(useGetMyConversations).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyConversations>)

    renderWithProviders(<ClientChat />)
    expect(screen.getByText('Чат')).toBeInTheDocument()
  })

  it('shows empty state when no conversations', () => {
    vi.mocked(useGetMyConversations).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyConversations>)

    renderWithProviders(<ClientChat />)
    expect(screen.getByText('Нет бесед')).toBeInTheDocument()
  })

  it('renders conversation list', () => {
    vi.mocked(useGetMyConversations).mockReturnValue({
      data: {
        data: mockConversations,
        success: true,
        meta: { total_count: 2, page: 0, page_size: 50, total_pages: 1 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyConversations>)

    renderWithProviders(<ClientChat />)
    expect(screen.getByText('Баня bath-1')).toBeInTheDocument()
    expect(screen.getByText('Баня bath-2')).toBeInTheDocument()
  })

  it('shows messages when conversation is selected', async () => {
    vi.mocked(useGetMyConversations).mockReturnValue({
      data: {
        data: mockConversations,
        success: true,
        meta: { total_count: 2 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyConversations>)

    vi.mocked(useGetConversationsIdMessages).mockReturnValue({
      data: {
        data: mockMessages,
        success: true,
        meta: { total_count: 2 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetConversationsIdMessages>)

    renderWithProviders(<ClientChat />)

    fireEvent.click(screen.getByText('Баня bath-1'))

    await waitFor(() => {
      expect(screen.getByText('← Назад к беседам')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Введите сообщение...')).toBeInTheDocument()
    })
  })

  it('calls send mutation when message is submitted', async () => {
    const mutateFn = vi.fn()
    vi.mocked(usePostConversationsIdMessages).mockReturnValue({
      mutate: mutateFn,
      isPending: false,
    } as unknown as ReturnType<typeof usePostConversationsIdMessages>)

    vi.mocked(useGetMyConversations).mockReturnValue({
      data: { data: mockConversations, success: true, meta: { total_count: 2 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyConversations>)

    vi.mocked(useGetConversationsIdMessages).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetConversationsIdMessages>)

    renderWithProviders(<ClientChat />)

    fireEvent.click(screen.getByText('Баня bath-1'))

    await waitFor(() => {
      const input = screen.getByPlaceholderText('Введите сообщение...')
      fireEvent.change(input, { target: { value: 'Привет!' } })
    })

    const sendButton = screen.getByRole('button', { name: /send/i })
    fireEvent.click(sendButton)

    expect(mutateFn).toHaveBeenCalledWith({
      id: 'conv-1',
      data: { text: 'Привет!' },
    })
  })

  it('has search input for filtering conversations', () => {
    vi.mocked(useGetMyConversations).mockReturnValue({
      data: {
        data: mockConversations,
        success: true,
        meta: { total_count: 2 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyConversations>)

    renderWithProviders(<ClientChat />)
    expect(screen.getByPlaceholderText('Поиск бесед...')).toBeInTheDocument()
  })
})
