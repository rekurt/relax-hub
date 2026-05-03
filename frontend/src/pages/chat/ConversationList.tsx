import { Avatar, Typography, Input, Spin } from '@/components/design/system'
import { UserOutlined, SearchOutlined } from '@/components/design/icons'
import { useState, useMemo } from 'react'
import { useGetMyConversations } from '@/api/generated/chat/chat'
import { useAuthStore } from '@/stores/auth'
import type { InternalHandlerConversationResponse } from '@/api/generated/model'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/ru'
import EmptyState from '@/components/EmptyState'

dayjs.extend(relativeTime)
dayjs.locale('ru')

const { Text } = Typography

interface ConversationListProps {
  selectedId: string | null
  onSelect: (conversation: InternalHandlerConversationResponse) => void
}

export default function ConversationList({ selectedId, onSelect }: ConversationListProps) {
  const [search, setSearch] = useState('')
  const userRole = useAuthStore((s) => s.user?.role)

  const { data, isLoading } = useGetMyConversations({ page: 1, page_size: 50 })

  const conversations = useMemo(() => {
    const items = data?.data ?? []
    if (!search.trim()) return items
    const q = search.toLowerCase()
    return items.filter(
      (c) =>
        c.id?.toLowerCase().includes(q) ||
        c.client_id?.toLowerCase().includes(q) ||
        c.bathhouse_id?.toLowerCase().includes(q),
    )
  }, [data?.data, search])

  if (isLoading) {
    return (
      <div className="rh-chat-loading">
        <Spin />
      </div>
    )
  }

  return (
    <div className="rh-chat-conversation-list">
      <div className="rh-chat-search">
        <Input
          placeholder="Поиск бесед..."
          prefix={<SearchOutlined />}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          allowClear
        />
      </div>

      <div className="rh-chat-conversation-scroll">
        {conversations.length === 0 ? (
          <div className="rh-chat-empty">
            <EmptyState description="Нет бесед" />
          </div>
        ) : (
          conversations.map((conv) => (
            <div
              key={conv.id}
              onClick={() => onSelect(conv)}
              className={selectedId === conv.id ? 'rh-chat-conversation rh-chat-conversation--selected' : 'rh-chat-conversation'}
            >
              <Avatar icon={<UserOutlined />} />
              <div className="rh-chat-conversation__content">
                <Text ellipsis className="rh-chat-conversation__title">
                  {userRole === 'client'
                    ? `Баня ${conv.bathhouse_id?.slice(0, 8)}`
                    : `Клиент ${conv.client_id?.slice(0, 8)}`}
                </Text>
                <Text type="secondary" className="rh-chat-conversation__time">
                  {conv.last_message_at
                    ? dayjs(conv.last_message_at).fromNow()
                    : dayjs(conv.created_at).fromNow()}
                </Text>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
