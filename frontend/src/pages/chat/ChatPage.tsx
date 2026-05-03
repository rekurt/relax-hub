import { useState } from 'react'
import { Grid } from '@/components/design/system'
import type { InternalHandlerConversationResponse } from '@/api/generated/model'
import { useWebSocketNotifications } from '@/lib/useWebSocketNotifications'
import ConversationList from './ConversationList'
import MessageArea from './MessageArea'
import PageHeader from '@/components/PageHeader'

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

  if (isMobile) {
    return (
      <div className="rh-stack rh-chat-page">
        <PageHeader
          eyebrow="Коммуникации"
          title="Чат"
          description="Переписка с гостями по активным обращениям и бронированиям."
          size="compact"
        />
        <div className="rh-chat-shell">
          {selectedConversation ? (
            <div className="rh-chat-mobile-view">
              <button
                type="button"
                className="rh-chat-back"
                onClick={handleBack}
              >
                ← Назад к беседам
              </button>
              <div className="rh-chat-main">
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
    <div className="rh-stack rh-chat-page">
      <PageHeader
        eyebrow="Коммуникации"
        title="Чат"
        description="Переписка с гостями по активным обращениям и бронированиям."
        size="compact"
      />
      <div className="rh-chat-shell rh-chat-shell--desktop">
        <div className="rh-chat-sidebar">
          <ConversationList selectedId={selectedConversation?.id ?? null} onSelect={handleSelect} />
        </div>
        <div className="rh-chat-main">
          <MessageArea conversationId={selectedConversation?.id ?? null} />
        </div>
      </div>
    </div>
  )
}
