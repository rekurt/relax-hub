import type { MouseEvent } from 'react'
import { App, Checkbox } from '@/components/design/system'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import type { InternalHandlerBathhouseResponse } from '@/api/generated/model'
import { usePostBathhousesIdFavorite } from '@/api/generated/favorites/favorites'
import { resolveAssetUrl } from '@/lib/asset-url'
import { formatPrice } from '@/lib/format'
import { useAuthStore } from '@/stores/auth'
import { DesignButton, DesignIcon, DesignListingCard } from '@/components/design'

const AMENITY_LABELS: Record<string, string> = {
  has_sauna: 'Сауна',
  has_steam_room: 'Парная',
  has_pool: 'Бассейн',
  has_hot_tub: 'Джакузи',
  has_bbq: 'Мангал',
  has_karaoke: 'Караоке',
}

interface BathhouseCardProps {
  bathhouse: InternalHandlerBathhouseResponse
  showFavorite?: boolean
  showCompare?: boolean
  isCompareSelected?: boolean
  onCompareToggle?: (id: string) => void
}

export default function BathhouseCard({
  bathhouse,
  showFavorite = true,
  showCompare = false,
  isCompareSelected = false,
  onCompareToggle,
}: BathhouseCardProps) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { message } = App.useApp()
  const currentUser = useAuthStore((s) => s.user)

  const favoriteMutation = usePostBathhousesIdFavorite({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['/bathhouses'] })
        queryClient.invalidateQueries({ queryKey: ['/my/favorites'] })
      },
      onError: () => message.error('Не удалось обновить избранное'),
    },
  })

  const amenities = Object.entries(AMENITY_LABELS)
    .filter(([key]) => bathhouse[key as keyof InternalHandlerBathhouseResponse])
    .map(([, label]) => label)

  const coverImage = resolveAssetUrl(bathhouse.images?.[0] ?? bathhouse.gallery_preview?.[0]?.url)

  const handleFavoriteClick = (e: MouseEvent) => {
    e.stopPropagation()
    if (bathhouse.id) {
      favoriteMutation.mutate({ id: bathhouse.id })
    }
  }

  const cardActions: React.ReactNode[] = []
  if (showFavorite && currentUser?.role === 'client') {
    cardActions.push(
      <DesignButton
        key="favorite"
        variant="ghost"
        size="sm"
        icon={(
          <DesignIcon
            name="heart"
            size={16}
            fill={bathhouse.is_favorite ? 'currentColor' : 'none'}
            style={{ color: bathhouse.is_favorite ? 'var(--rh-error)' : undefined }}
          />
        )}
        onClick={handleFavoriteClick}
        disabled={favoriteMutation.isPending}
      >
        {bathhouse.is_favorite ? 'В избранном' : 'В избранное'}
      </DesignButton>,
    )
  }
  if (showCompare) {
    cardActions.push(
      <Checkbox
        key="compare"
        checked={isCompareSelected}
        onChange={(e) => {
          e.stopPropagation()
          if (bathhouse.id && onCompareToggle) onCompareToggle(bathhouse.id)
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <DesignIcon name="grid" size={15} /> Сравнить
      </Checkbox>,
    )
  }

  const badge = bathhouse.last_minute_active
    ? `Срочная скидка ${bathhouse.last_minute_discount_percent ? `-${bathhouse.last_minute_discount_percent}%` : ''}`.trim()
    : undefined

  return (
    <DesignListingCard
      className="rh-listing-card"
      onClick={() => navigate(`/bathhouses/${bathhouse.slug ?? bathhouse.id}`)}
      name={bathhouse.name}
      address={bathhouse.address}
      price={bathhouse.price_per_hour ? `${formatPrice(bathhouse.price_per_hour)}/ч` : undefined}
      rating={bathhouse.rating ?? 0}
      reviewCount={bathhouse.review_count ?? 0}
      verified={bathhouse.is_photo_verified}
      imageUrl={coverImage}
      imageAlt={bathhouse.name}
      badge={badge}
      badgeTone="red"
      tags={[
        ...amenities.slice(0, 4),
        ...(amenities.length > 4 ? [`+${amenities.length - 4}`] : []),
      ]}
      actions={cardActions.length > 0 ? cardActions : undefined}
    />
  )
}
