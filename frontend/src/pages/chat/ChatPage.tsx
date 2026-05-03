import { useState } from 'react'
import { Typography, Grid } from 'antd'
import type { InternalHandlerConversationResponse } from '@/api/generated/model'
import { useWebSocketNotifications } from '@/lib/useWebSocketNotifications'
import ConversationList from './ConversationList'
import MessageArea from './MessageArea'

const { Title } = Typography
const { useBreakpoint } = Grid

export default function ChatPage() {
  const [selectedConversation, setSelectedConversation] =
    useState<InternalHandlerConversationResponse | null>(null)
  const screens = useBreakpoint()
  const isMobile = !screens.md

  useWebSocketNotifications({
    activeConversationId: selectedConversation?.id ?? null,
  })

  const handleSelect = (conv: InternalHandlerConversationResponse) => {
    setSelectedConversation(conv)
  }

  const handleBack = () => {
    setSelectedConversation(null)
  }

  const chatHeight = 'calc(100vh - 180px)'

  if (isMobile) {
    return (
      <div>
        <Title level={4} style={{ marginBottom: 16 }}>
          Чат
        </Title>
        <div
          style={{
            background: 'var(--rh-card-bg)',
            borderRadius: 28,
            border: '1px solid var(--rh-border)',
            boxShadow: 'var(--rh-shadow-soft)',
            backdropFilter: 'blur(18px)',
            height: chatHeight,
            overflow: 'hidden',
          }}
        >
          {selectedConversation ? (
            <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
              <div
                style={{
                  padding: '8px 16px',
                  borderBottom: '1px solid var(--rh-border)',
                  cursor: 'pointer',
                  color: 'var(--rh-primary)',
                  fontWeight: 700,
                }}
                onClick={handleBack}
              >
                ← Назад к беседам
              </div>
              <div style={{ flex: 1, overflow: 'hidden' }}>
                <MessageArea conversationId={selectedConversation.id ?? null} />
              </div>
            </div>
          ) : (
            <ConversationList selectedId={null} onSelect={handleSelect} />
          )}
        </div>
      </div>
    )
  }

  return (
    <div>
      <Title level={4} style={{ marginBottom: 16 }}>
        Чат
      </Title>
      <div
        style={{
          display: 'flex',
          background: 'var(--rh-card-bg)',
          borderRadius: 28,
          border: '1px solid var(--rh-border)',
          boxShadow: 'var(--rh-shadow-soft)',
          backdropFilter: 'blur(18px)',
          height: chatHeight,
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            width: 320,
            borderRight: '1px solid var(--rh-border)',
            overflow: 'hidden',
          }}
        >
          <ConversationList selectedId={selectedConversation?.id ?? null} onSelect={handleSelect} />
        </div>
        <div style={{ flex: 1, overflow: 'hidden' }}>
          <MessageArea conversationId={selectedConversation?.id ?? null} />
        </div>
      </div>
    </div>
  )
}
