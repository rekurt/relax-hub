import { Card, Tag, Rate, Typography, Space, Button, Image } from 'antd'
import {
  EnvironmentOutlined,
  HeartOutlined,
  HeartFilled,
  CheckCircleOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { message } from 'antd'
import type { InternalHandlerBathhouseResponse } from '@/api/generated/model'
import { usePostBathhousesIdFavorite } from '@/api/generated/favorites/favorites'
import { formatPrice } from '@/lib/format'

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
}

export default function BathhouseCard({ bathhouse, showFavorite = true }: BathhouseCardProps) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

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

  const coverImage = bathhouse.images?.[0] ?? bathhouse.gallery_preview?.[0]?.url

  const handleFavoriteClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (bathhouse.id) {
      favoriteMutation.mutate({ id: bathhouse.id })
    }
  }

  return (
    <Card
      hoverable
      onClick={() => navigate(`/client/bathhouse/${bathhouse.id}`)}
      cover={
        coverImage ? (
          <Image
            alt={bathhouse.name}
            src={coverImage}
            height={200}
            style={{ objectFit: 'cover' }}
            preview={false}
          />
        ) : (
          <div
            style={{
              height: 200,
              background: '#f5f5f5',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Text type="secondary">Нет фото</Text>
          </div>
        )
      }
      actions={
        showFavorite
          ? [
              <Button
                key="favorite"
                type="text"
                icon={bathhouse.is_favorite ? <HeartFilled style={{ color: '#ff4d4f' }} /> : <HeartOutlined />}
                onClick={handleFavoriteClick}
                loading={favoriteMutation.isPending}
              >
                {bathhouse.is_favorite ? 'В избранном' : 'В избранное'}
              </Button>,
            ]
          : undefined
      }
    >
      <Space direction="vertical" size={4} style={{ width: '100%' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <Title level={5} style={{ margin: 0 }}>
            {bathhouse.name}
            {bathhouse.is_photo_verified && (
              <CheckCircleOutlined style={{ color: '#52c41a', marginLeft: 6, fontSize: 14 }} />
            )}
          </Title>
          <Text strong style={{ whiteSpace: 'nowrap' }}>
            {bathhouse.price_per_hour ? formatPrice(bathhouse.price_per_hour) + '/ч' : ''}
          </Text>
        </div>

        {bathhouse.address && (
          <Text type="secondary" style={{ fontSize: 13 }}>
            <EnvironmentOutlined style={{ marginRight: 4 }} />
            {bathhouse.address}
          </Text>
        )}

        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <Rate disabled allowHalf value={bathhouse.rating ?? 0} style={{ fontSize: 14 }} />
          <Text type="secondary" style={{ fontSize: 13 }}>
            {bathhouse.rating?.toFixed(1)} ({bathhouse.review_count ?? 0})
          </Text>
        </div>

        {amenities.length > 0 && (
          <div style={{ marginTop: 4 }}>
            {amenities.slice(0, 4).map((label) => (
              <Tag key={label} style={{ marginBottom: 4 }}>
                {label}
              </Tag>
            ))}
            {amenities.length > 4 && (
              <Tag>+{amenities.length - 4}</Tag>
            )}
          </div>
        )}
      </Space>
    </Card>
  )
}
