import { useEffect, useMemo, useRef, useState } from 'react'
import { Input, Button, Empty, Spin, Typography, Space } from 'antd'
import { SendOutlined } from '@ant-design/icons'
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
      style={{
        display: 'flex',
        justifyContent: isOwn ? 'flex-end' : 'flex-start',
        marginBottom: 8,
      }}
    >
      <div
        style={{
          maxWidth: '70%',
          padding: '8px 12px',
          borderRadius: 12,
          background: isOwn ? '#1677ff' : '#f0f0f0',
          color: isOwn ? '#fff' : 'inherit',
        }}
      >
        <div style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>{message.text}</div>
        <Text
          type="secondary"
          style={{
            fontSize: 11,
            color: isOwn ? 'rgba(255,255,255,0.7)' : undefined,
            display: 'block',
            textAlign: 'right',
            marginTop: 4,
          }}
        >
          {dayjs(message.created_at).format('HH:mm')}
        </Text>
      </div>
    </div>
  )
}

export default function MessageArea({ conversationId }: MessageAreaProps) {
  const [text, setText] = useState('')
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
      onSuccess: () => {
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
    if (
      conversationId &&
      conversationId !== markedReadForRef.current &&
      messages.some((m) => !m.is_read && m.sender_id !== currentUser?.id)
    ) {
      markReadRef.current({ id: conversationId })
      markedReadForRef.current = conversationId
    }
  }, [conversationId, messages, currentUser?.id])

  const handleSend = () => {
    if (!text.trim() || !conversationId) return
    sendMutation.mutate({
      id: conversationId,
      data: { text: text.trim() },
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
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100%',
        }}
      >
        <Empty description="Выберите беседу для просмотра сообщений" />
      </div>
    )
  }

  if (isLoading) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100%',
        }}
      >
        <Spin />
      </div>
    )
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div
        style={{
          flex: 1,
          overflow: 'auto',
          padding: 16,
        }}
      >
        {messages.length === 0 ? (
          <Empty description="Нет сообщений" style={{ marginTop: 40 }} />
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

      <div
        style={{
          padding: '12px 16px',
          borderTop: '1px solid #f0f0f0',
          background: '#fafafa',
        }}
      >
        <Space.Compact style={{ width: '100%' }}>
          <Input.TextArea
            value={text}
            onChange={(e) => setText(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Введите сообщение..."
            autoSize={{ minRows: 1, maxRows: 4 }}
            style={{ resize: 'none' }}
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
