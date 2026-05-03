import { Typography } from '@/components/design/system'
import { ClockCircleOutlined } from '@/components/design/icons'
import { formatPrice } from '@/lib/format'
import { useNavigate } from 'react-router-dom'
import { useGetMyRecentlyViewed } from '@/api/generated/saved-searches/saved-searches'
import PublicState from '@/components/PublicState'
import { resolveAssetUrl } from '@/lib/asset-url'
import { useAuthStore } from '@/stores/auth'
import { DesignListingCard } from '@/components/design'

const { Title } = Typography

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
          gap: 16,
          overflowX: 'auto',
          paddingBottom: 8,
          scrollSnapType: 'x mandatory',
        }}
      >
        {items.map((item) => (
          <div
            key={item.id}
            style={{
              minWidth: 240,
              maxWidth: 280,
              flex: '0 0 auto',
              scrollSnapAlign: 'start',
            }}
          >
            <DesignListingCard
              onClick={() => navigate(`/bathhouses/${item.slug ?? item.id}`)}
              name={item.name}
              imageUrl={item.cover_photo ? resolveAssetUrl(item.cover_photo as string) : undefined}
              imageAlt={item.name}
              rating={item.rating != null ? (item.rating as number) : null}
              price={item.base_price != null ? `от ${formatPrice(item.base_price as number)}` : undefined}
            />
          </div>
        ))}
      </div>
    </div>
  )
}
