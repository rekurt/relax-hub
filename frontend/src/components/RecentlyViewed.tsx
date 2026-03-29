import { Card, Typography, Space } from 'antd'
import { ClockCircleOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useGetMyRecentlyViewed } from '@/api/generated/saved-searches/saved-searches'
import { useAuthStore } from '@/stores/auth'

const { Text, Title } = Typography

export default function RecentlyViewed() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)

  const { data, isLoading } = useGetMyRecentlyViewed(
    { limit: 20 },
    { query: { enabled: !!user } },
  )
  const items = data?.data ?? []

  if (!user || isLoading || items.length === 0) return null

  return (
    <div style={{ marginBottom: 24 }}>
      <Title level={4} style={{ marginBottom: 12 }}>
        <ClockCircleOutlined style={{ marginRight: 8 }} />
        Недавно просмотренные
      </Title>
      <div
        style={{
          display: 'flex',
          gap: 12,
          overflowX: 'auto',
          paddingBottom: 8,
          scrollSnapType: 'x mandatory',
        }}
      >
        {items.map((item) => (
          <Card
            key={item.id}
            hoverable
            size="small"
            style={{
              minWidth: 180,
              maxWidth: 220,
              flex: '0 0 auto',
              scrollSnapAlign: 'start',
            }}
            onClick={() => navigate(`/client/bathhouse/${item.slug ?? item.id}`)}
          >
            <Space direction="vertical" size={2}>
              <Text strong ellipsis style={{ maxWidth: 190 }}>
                {item.name}
              </Text>
              {item.viewed_at && (
                <Text type="secondary" style={{ fontSize: 12 }}>
                  {new Date(item.viewed_at).toLocaleDateString('ru-RU')}
                </Text>
              )}
            </Space>
          </Card>
        ))}
      </div>
    </div>
  )
}
