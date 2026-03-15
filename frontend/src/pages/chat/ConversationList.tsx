import { Avatar, Badge, Typography, Input, Empty, Spin } from 'antd'
import { UserOutlined, SearchOutlined } from '@ant-design/icons'
import { useState, useMemo } from 'react'
import { useGetMyConversations } from '@/api/generated/chat/chat'
import type { InternalHandlerConversationResponse } from '@/api/generated/model'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/ru'

dayjs.extend(relativeTime)
dayjs.locale('ru')

const { Text } = Typography

interface ConversationListProps {
  selectedId: string | null
  onSelect: (conversation: InternalHandlerConversationResponse) => void
}

export default function ConversationList({ selectedId, onSelect }: ConversationListProps) {
  const [search, setSearch] = useState('')

  const { data, isLoading } = useGetMyConversations({ page: 0, page_size: 50 })

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
      <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}>
        <Spin />
      </div>
    )
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ padding: '12px 16px', borderBottom: '1px solid #f0f0f0' }}>
        <Input
          placeholder="Поиск бесед..."
          prefix={<SearchOutlined />}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          allowClear
        />
      </div>

      <div style={{ flex: 1, overflow: 'auto' }}>
        {conversations.length === 0 ? (
          <Empty description="Нет бесед" style={{ marginTop: 40 }} />
        ) : (
          conversations.map((conv) => (
            <div
              key={conv.id}
              onClick={() => onSelect(conv)}
              style={{
                padding: '12px 16px',
                cursor: 'pointer',
                background: selectedId === conv.id ? '#e6f4ff' : 'transparent',
                borderBottom: '1px solid #f0f0f0',
                display: 'flex',
                alignItems: 'center',
                gap: 12,
              }}
            >
              <Badge dot={false}>
                <Avatar icon={<UserOutlined />} />
              </Badge>
              <div style={{ flex: 1, minWidth: 0 }}>
                <Text ellipsis style={{ maxWidth: 180, display: 'block' }}>
                  Клиент {conv.client_id?.slice(0, 8)}
                </Text>
                <Text type="secondary" style={{ fontSize: 12 }}>
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
