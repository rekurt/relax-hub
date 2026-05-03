import { Card, Typography, Space, Rate } from '@/components/design/system'
import { ClockCircleOutlined } from '@/components/design/icons'
import { formatPrice } from '@/lib/format'
import { useNavigate } from 'react-router-dom'
import { useGetMyRecentlyViewed } from '@/api/generated/saved-searches/saved-searches'
import PublicState from '@/components/PublicState'
import { resolveAssetUrl } from '@/lib/asset-url'
import { useAuthStore } from '@/stores/auth'

const { Text, Title } = Typography

export default function RecentlyViewed() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)

  const { data, isLoading, isError, refetch } = useGetMyRecentlyViewed(
    { limit: 20 },
    { query: { enabled: !!user } },
  )
  const items = (data?.data ?? []) as Array<Record<string, unknown> & { id: string; name: string; slug?: string; viewed_at?: string }>

  if (!user) return null

  if (isLoading) return null

  if (isError) {
    return (
      <div style={{ marginBottom: 24 }}>
        <PublicState
          kind="degraded"
          compact
          title="Недавно просмотренные временно недоступны"
          description="Попробуйте обновить этот блок чуть позже."
          actionText="Повторить"
          onAction={() => void refetch()}
        />
      </div>
    )
  }

  if (items.length === 0) return null

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
            cover={item.cover_photo ? (
              <img
                alt={item.name}
                src={resolveAssetUrl(item.cover_photo as string)}
                style={{ height: 100, objectFit: 'cover' }}
              />
            ) : undefined}
            onClick={() => navigate(`/bathhouses/${item.slug ?? item.id}`)}
          >
            <Space orientation="vertical" size={2}>
              <Text strong ellipsis style={{ maxWidth: 190 }}>
                {item.name}
              </Text>
              {item.rating != null && (
                <Rate disabled allowHalf value={item.rating as number} style={{ fontSize: 12 }} />
              )}
              {item.base_price != null && (
                <Text type="secondary" style={{ fontSize: 12 }}>
                  от {formatPrice(item.base_price as number)}
                </Text>
              )}
            </Space>
          </Card>
        ))}
      </div>
    </div>
  )
}
