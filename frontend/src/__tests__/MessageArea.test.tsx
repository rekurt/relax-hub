import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import MessageArea from '@/pages/chat/MessageArea'

vi.mock('@/api/generated/chat/chat', () => ({
  useGetConversationsIdMessages: vi.fn(),
  usePostConversationsIdMessages: vi.fn(),
  usePatchConversationsIdRead: vi.fn(),
  getGetConversationsIdMessagesQueryKey: vi.fn(() => ['messages']),
  getGetMyConversationsQueryKey: vi.fn(() => ['conversations']),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn((selector) =>
    selector({
      user: { id: 'user-1', name: 'Test User' },
    }),
  ),
}))

import {
  useGetConversationsIdMessages,
  usePostConversationsIdMessages,
  usePatchConversationsIdRead,
} from '@/api/generated/chat/chat'

const mockSendMutate = vi.fn()
const mockMarkRead = vi.fn()

function renderWithProviders(conversationId: string | null) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter>
            <div style={{ height: 500 }}>
              <MessageArea conversationId={conversationId} />
            </div>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

function setupMocks(overrides?: {
  messages?: Array<{ id: string; text: string; sender_id: string; created_at: string; is_read?: boolean }>
  sendOnSuccess?: unknown
}) {
  const messages = overrides?.messages ?? []

  vi.mocked(useGetConversationsIdMessages).mockReturnValue({
    data: { data: messages, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetConversationsIdMessages>)

  let capturedOnSuccess: ((data: unknown) => void) | undefined

  vi.mocked(usePostConversationsIdMessages).mockImplementation((options) => {
    capturedOnSuccess = options?.mutation?.onSuccess as ((data: unknown) => void) | undefined
    return {
      mutate: (...args: unknown[]) => {
        mockSendMutate(...args)
        if (overrides?.sendOnSuccess && capturedOnSuccess) {
          capturedOnSuccess(overrides.sendOnSuccess)
        }
      },
      isPending: false,
    } as unknown as ReturnType<typeof usePostConversationsIdMessages>
  })

  vi.mocked(usePatchConversationsIdRead).mockReturnValue({
    mutate: mockMarkRead,
  } as unknown as ReturnType<typeof usePatchConversationsIdRead>)

  return { getCapturedOnSuccess: () => capturedOnSuccess }
}

describe('MessageArea', () => {
  beforeEach(() => {
    mockSendMutate.mockReset()
    mockMarkRead.mockReset()
  })

  it('renders empty state when no conversation selected', () => {
    setupMocks()
    renderWithProviders(null)

    expect(screen.getByText('Выберите беседу для просмотра сообщений')).toBeInTheDocument()
  })

  it('renders messages', () => {
    setupMocks({
      messages: [
        { id: 'msg-1', text: 'Привет!', sender_id: 'user-1', created_at: '2026-03-28T12:00:00Z' },
        { id: 'msg-2', text: 'Здравствуйте!', sender_id: 'other-1', created_at: '2026-03-28T12:01:00Z' },
      ],
    })
    renderWithProviders('conv-1')

    expect(screen.getByText('Привет!')).toBeInTheDocument()
    expect(screen.getByText('Здравствуйте!')).toBeInTheDocument()
  })

  it('sends message on button click', () => {
    setupMocks()
    renderWithProviders('conv-1')

    const input = screen.getByPlaceholderText('Введите сообщение...')
    fireEvent.change(input, { target: { value: 'Новое сообщение' } })

    const sendButton = screen.getByRole('button', { name: /send/i })
    fireEvent.click(sendButton)

    expect(mockSendMutate).toHaveBeenCalledWith({
      id: 'conv-1',
      data: { text: 'Новое сообщение' },
    })
  })

  it('shows filter warning when sent message differs from returned', async () => {
    // Simulate the server filtering contact info
    setupMocks({
      sendOnSuccess: { data: { id: 'msg-new', text: 'Позвоните мне ***', sender_id: 'user-1' } },
    })
    renderWithProviders('conv-1')

    const input = screen.getByPlaceholderText('Введите сообщение...')
    fireEvent.change(input, { target: { value: 'Позвоните мне +7 999 123 4567' } })

    const sendButton = screen.getByRole('button', { name: /send/i })
    fireEvent.click(sendButton)

    await waitFor(() => {
      expect(screen.getByText('Контактные данные скрыты')).toBeInTheDocument()
    })
  })

  it('does not show filter warning when message is unchanged', async () => {
    setupMocks({
      sendOnSuccess: { data: { id: 'msg-new', text: 'Обычное сообщение', sender_id: 'user-1' } },
    })
    renderWithProviders('conv-1')

    const input = screen.getByPlaceholderText('Введите сообщение...')
    fireEvent.change(input, { target: { value: 'Обычное сообщение' } })

    const sendButton = screen.getByRole('button', { name: /send/i })
    fireEvent.click(sendButton)

    await waitFor(() => {
      expect(screen.queryByText('Контактные данные скрыты')).not.toBeInTheDocument()
    })
  })

  it('filter warning is dismissible', async () => {
    setupMocks({
      sendOnSuccess: { data: { id: 'msg-new', text: '***', sender_id: 'user-1' } },
    })
    renderWithProviders('conv-1')

    const input = screen.getByPlaceholderText('Введите сообщение...')
    fireEvent.change(input, { target: { value: 'test@email.com' } })

    const sendButton = screen.getByRole('button', { name: /send/i })
    fireEvent.click(sendButton)

    await waitFor(() => {
      expect(screen.getByText('Контактные данные скрыты')).toBeInTheDocument()
    })

    const closeButton = screen.getByRole('button', { name: /close/i })
    fireEvent.click(closeButton)

    await waitFor(() => {
      expect(screen.queryByText('Контактные данные скрыты')).not.toBeInTheDocument()
    })
  })
})
