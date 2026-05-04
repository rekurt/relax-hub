import { useEffect, useMemo, useRef, useState } from 'react'
import { Input, Button, Spin, Typography, Space, Alert } from '@/components/design/system'
import { SendOutlined, WarningOutlined } from '@/components/design/icons'
import {
  useGetConversationsIdMessages,
  usePostConversationsIdMessages,
  usePatchConversationsIdRead,
  getGetConversationsIdMessagesQueryKey,
  getGetMyConversationsQueryKey,
} from '@/api/generated/chat/chat'
import { useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '@/stores/auth'
import type { InternalHandlerMessageResponse } from '@/api/generated/model'
import dayjs from 'dayjs'
import EmptyState from '@/components/EmptyState'

const { Text } = Typography

interface MessageAreaProps {
  conversationId: string | null
}

function MessageBubble({
  message,
  isOwn,
}: {
  message: InternalHandlerMessageResponse
  isOwn: boolean
}) {
  return (
    <div
      className={isOwn ? 'rh-message-row rh-message-row--own' : 'rh-message-row'}
    >
      <div
        className={isOwn ? 'rh-message-bubble rh-message-bubble--own' : 'rh-message-bubble'}
      >
        <div className="rh-message-bubble__text">{message.text}</div>
        <Text
          type="secondary"
          className={isOwn ? 'rh-message-bubble__time rh-message-bubble__time--own' : 'rh-message-bubble__time'}
        >
          {dayjs(message.created_at).format('HH:mm')}
        </Text>
      </div>
    </div>
  )
}

export default function MessageArea({ conversationId }: MessageAreaProps) {
  const [text, setText] = useState('')
  const [filterWarning, setFilterWarning] = useState(false)
  const lastSentTextRef = useRef<string | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const queryClient = useQueryClient()
  const currentUser = useAuthStore((s) => s.user)

  const { data, isLoading } = useGetConversationsIdMessages(
    conversationId ?? '',
    { page: 1, page_size: 100 },
    { query: { enabled: !!conversationId, refetchInterval: 5000 } },
  )

  const sendMutation = usePostConversationsIdMessages({
    mutation: {
      onSuccess: (response) => {
        const sentText = lastSentTextRef.current
        const returnedText = response?.data?.text
        if (sentText && returnedText && sentText !== returnedText) {
          setFilterWarning(true)
        }
        lastSentTextRef.current = null
        setText('')
        queryClient.invalidateQueries({
          queryKey: getGetConversationsIdMessagesQueryKey(conversationId ?? ''),
        })
        queryClient.invalidateQueries({
          queryKey: getGetMyConversationsQueryKey(),
        })
      },
    },
  })

  const { mutate: markRead } = usePatchConversationsIdRead()
  const markReadRef = useRef(markRead)
  const markedReadForRef = useRef<string | null>(null)

  useEffect(() => {
    markReadRef.current = markRead
  }, [markRead])

  const messages = useMemo(() => data?.data ?? [], [data?.data])

  useEffect(() => {
    if (typeof messagesEndRef.current?.scrollIntoView === 'function') {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [messages.length])

  useEffect(() => {
    const unreadCount = messages.filter((m) => !m.is_read && m.sender_id !== currentUser?.id).length
    const unreadKey = `${conversationId}:${unreadCount}`
    if (
      conversationId &&
      unreadCount > 0 &&
      unreadKey !== markedReadForRef.current
    ) {
      markReadRef.current({ id: conversationId })
      markedReadForRef.current = unreadKey
    }
  }, [conversationId, messages, currentUser?.id])

  const handleSend = () => {
    const trimmed = text.trim()
    if (!trimmed || !conversationId) return
    lastSentTextRef.current = trimmed
    setFilterWarning(false)
    sendMutation.mutate({
      id: conversationId,
      data: { text: trimmed },
    })
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  if (!conversationId) {
    return (
      <div className="rh-chat-center-state">
        <EmptyState description="Выберите беседу для просмотра сообщений" />
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="rh-chat-center-state">
        <Spin />
      </div>
    )
  }

  return (
    <div className="rh-message-area">
      <div className="rh-message-scroll">
        {messages.length === 0 ? (
          <div className="rh-chat-empty">
            <EmptyState description="Нет сообщений" />
          </div>
        ) : (
          <>
            {[...messages].reverse().map((msg) => (
              <MessageBubble
                key={msg.id}
                message={msg}
                isOwn={msg.sender_id === currentUser?.id}
              />
            ))}
            <div ref={messagesEndRef} />
          </>
        )}
      </div>

      {filterWarning && (
        <Alert
          title="Контактные данные скрыты"
          description="Телефоны, email и ссылки автоматически скрываются в чате для вашей безопасности. Обмен контактами возможен после бронирования."
          type="warning"
          showIcon
          icon={<WarningOutlined />}
          closable
          onClose={() => setFilterWarning(false)}
          className="rh-chat-alert"
        />
      )}

      <div className="rh-message-composer">
        <Space.Compact className="rh-full-width">
          <Input.TextArea
            value={text}
            onChange={(e) => setText(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Введите сообщение..."
            rows={1}
            className="rh-message-input"
          />
          <Button
            type="primary"
            icon={<SendOutlined />}
            onClick={handleSend}
            loading={sendMutation.isPending}
            disabled={!text.trim()}
          />
        </Space.Compact>
      </div>
    </div>
  )
}
