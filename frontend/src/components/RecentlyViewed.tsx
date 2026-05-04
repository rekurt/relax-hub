import { ClockCircleOutlined } from '@/components/design/icons'
import { formatPrice } from '@/lib/format'
import { useNavigate } from 'react-router-dom'
import { useGetMyRecentlyViewed } from '@/api/generated/saved-searches/saved-searches'
import PublicState from '@/components/PublicState'
import { resolveAssetUrl } from '@/lib/asset-url'
import { useAuthStore } from '@/stores/auth'
import { DesignListingCard } from '@/components/design'

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
      <div className="rh-recently-viewed">
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
    <div className="rh-recently-viewed">
      <h2 className="rh-component-section-title">
        <ClockCircleOutlined className="rh-component-section-title__icon" />
        Недавно просмотренные
      </h2>
      <div className="rh-horizontal-card-scroll">
        {items.map((item) => (
          <div
            key={item.id}
            className="rh-horizontal-card-scroll__item"
          >
            <DesignListingCard
              onClick={() => navigate(`/bathhouses/${item.slug ?? item.id}`)}
              name={item.name}
              imageUrl={item.cover_photo ? resolveAssetUrl(item.cover_photo as string) : undefined}
              imageAlt={item.name}
              rating={typeof item.rating === 'number' ? item.rating : undefined}
              price={item.base_price != null ? `от ${formatPrice(item.base_price as number)}` : undefined}
            />
          </div>
        ))}
      </div>
    </div>
  )
}
