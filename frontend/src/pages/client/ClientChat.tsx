import { useState } from 'react'
import { Typography, Grid } from 'antd'
import type { InternalHandlerConversationResponse } from '@/api/generated/model'
import { useWebSocketNotifications } from '@/lib/useWebSocketNotifications'
import ConversationList from '@/pages/chat/ConversationList'
import MessageArea from '@/pages/chat/MessageArea'

const { Title } = Typography
const { useBreakpoint } = Grid

export default function ClientChat() {
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
            background: '#fff',
            borderRadius: 8,
            border: '1px solid #f0f0f0',
            height: chatHeight,
            overflow: 'hidden',
          }}
        >
          {selectedConversation ? (
            <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
              <div
                style={{
                  padding: '8px 16px',
                  borderBottom: '1px solid #f0f0f0',
                  cursor: 'pointer',
                  color: '#1677ff',
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
          background: '#fff',
          borderRadius: 8,
          border: '1px solid #f0f0f0',
          height: chatHeight,
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            width: 320,
            borderRight: '1px solid #f0f0f0',
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
