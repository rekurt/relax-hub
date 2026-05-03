import { Card, Tag, Rate, Typography, Space, Button, Image, App, Checkbox } from 'antd'
import {
  EnvironmentOutlined,
  HeartOutlined,
  HeartFilled,
  CheckCircleOutlined,
  SwapOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import type { InternalHandlerBathhouseResponse } from '@/api/generated/model'
import { usePostBathhousesIdFavorite } from '@/api/generated/favorites/favorites'
import { resolveAssetUrl } from '@/lib/asset-url'
import { formatPrice } from '@/lib/format'
import { useAuthStore } from '@/stores/auth'

const { Text, Title } = Typography

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

  const handleFavoriteClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (bathhouse.id) {
      favoriteMutation.mutate({ id: bathhouse.id })
    }
  }

  const cardActions: React.ReactNode[] = []
  if (showFavorite && currentUser?.role === 'client') {
    cardActions.push(
      <Button
        key="favorite"
        type="text"
        icon={bathhouse.is_favorite ? <HeartFilled style={{ color: '#ff4d4f' }} /> : <HeartOutlined />}
        onClick={handleFavoriteClick}
        loading={favoriteMutation.isPending}
      >
        {bathhouse.is_favorite ? 'В избранном' : 'В избранное'}
      </Button>,
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
        <SwapOutlined /> Сравнить
      </Checkbox>,
    )
  }

  return (
    <Card
      hoverable
      className="bani-listing-card"
      onClick={() => navigate(`/bathhouses/${bathhouse.slug ?? bathhouse.id}`)}
      cover={
        coverImage ? (
          <Image
            className="bani-listing-card__image"
            alt={bathhouse.name}
            src={coverImage}
            height={200}
            style={{ objectFit: 'cover' }}
            preview={false}
          />
        ) : (
          <div className="bani-listing-card__image-placeholder">
            <Text type="secondary">Фото объекта</Text>
          </div>
        )
      }
      actions={cardActions.length > 0 ? cardActions : undefined}
    >
      <Space orientation="vertical" size={8} className="bani-listing-card__body">
        <div className="bani-listing-card__head">
          <Title level={5} className="bani-listing-card__title">
            {bathhouse.name}
            {bathhouse.is_photo_verified && (
              <CheckCircleOutlined className="bani-listing-card__verified" />
            )}
          </Title>
          <Text strong className="bani-listing-card__price">
            {bathhouse.price_per_hour ? formatPrice(bathhouse.price_per_hour) + '/ч' : ''}
          </Text>
        </div>

        {bathhouse.last_minute_active && (
          <Tag className="bani-listing-card__deal-tag" icon={<ThunderboltOutlined />}>
            Срочная скидка {bathhouse.last_minute_discount_percent ? `-${bathhouse.last_minute_discount_percent}%` : ''}
          </Tag>
        )}

        {bathhouse.address && (
          <Text type="secondary" className="bani-listing-card__address">
            <EnvironmentOutlined />
            {bathhouse.address}
          </Text>
        )}

        <div className="bani-listing-card__rating">
          <Rate disabled allowHalf value={bathhouse.rating ?? 0} className="bani-listing-card__stars" />
          <Text type="secondary" className="bani-listing-card__rating-text">
            {bathhouse.rating?.toFixed(1)} ({bathhouse.review_count ?? 0})
          </Text>
        </div>

        {amenities.length > 0 && (
          <div className="bani-listing-card__tags">
            {amenities.slice(0, 4).map((label) => (
              <Tag key={label} className="bani-listing-card__tag">
                {label}
              </Tag>
            ))}
            {amenities.length > 4 && (
              <Tag className="bani-listing-card__tag">+{amenities.length - 4}</Tag>
            )}
          </div>
        )}
      </Space>
    </Card>
  )
}
