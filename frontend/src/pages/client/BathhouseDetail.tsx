import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Typography,
  Spin,
  Row,
  Col,
  Card,
  Image,
  Rate,
  Tag,
  Space,
  Button,
  Descriptions,
  Divider,
  DatePicker,
  Empty,
  Pagination,
} from 'antd'
import {
  ArrowLeftOutlined,
  EnvironmentOutlined,
  HeartOutlined,
  HeartFilled,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CarOutlined,
  NodeIndexOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { App } from 'antd'
import { useGetBathhousesId, useGetBathhousesIdAvailableSlots, useGetBathhousesIdSchema } from '@/api/generated/bathhouses/bathhouses'
import { useGetBathhousesIdPhotos } from '@/api/generated/photos/photos'
import { useGetBathhousesIdReviews } from '@/api/generated/reviews/reviews'
import { useGetBathhousesIdSimilar } from '@/api/generated/recommendations/recommendations'
import { useGetBathhousesIdGallery } from '@/api/generated/review-media/review-media'
import { usePostBathhousesIdFavorite } from '@/api/generated/favorites/favorites'
import { axiosInstance } from '@/api/axios-instance'
import { formatPrice, formatDayOfWeek } from '@/lib/format'
import { useAuthStore } from '@/stores/auth'
import BathhouseCard from '@/components/BathhouseCard'
import ReviewCard from '@/components/ReviewCard'

const { Title, Text, Paragraph } = Typography

interface TransportItem {
  type: 'metro' | 'bus_stop' | 'parking'
  name: string
  distance_meters: number
  lat: number
  lng: number
}

const TRANSPORT_LABELS: Record<string, { label: string; color: string }> = {
  metro: { label: 'Метро', color: '#1677ff' },
  bus_stop: { label: 'Остановка', color: '#52c41a' },
  parking: { label: 'Парковка', color: '#faad14' },
}

function formatDistance(meters: number): string {
  if (meters >= 1000) {
    return `${(meters / 1000).toFixed(1)} км`
  }
  return `${meters} м`
}

const AMENITY_LIST = [
  { key: 'has_sauna', label: 'Сауна' },
  { key: 'has_steam_room', label: 'Парная' },
  { key: 'has_pool', label: 'Бассейн' },
  { key: 'has_hot_tub', label: 'Джакузи' },
  { key: 'has_bbq', label: 'Мангал' },
  { key: 'has_karaoke', label: 'Караоке' },
] as const

export default function BathhouseDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { message } = App.useApp()
  const currentUser = useAuthStore((s) => s.user)
  const [selectedDate, setSelectedDate] = useState<string>(dayjs().format('YYYY-MM-DD'))
  const [reviewPage, setReviewPage] = useState(1)

  const { data: bathhouseData, isLoading } = useGetBathhousesId(id ?? '', {
    query: { enabled: !!id },
  })
  const bathhouse = bathhouseData?.data

  const { data: photosData } = useGetBathhousesIdPhotos(id ?? '', {
    query: { enabled: !!id },
  })
  const photos = photosData?.data ?? []

  const { data: slotsData, isLoading: slotsLoading } = useGetBathhousesIdAvailableSlots(
    id ?? '',
    { date: selectedDate },
    { query: { enabled: !!id && !!selectedDate } },
  )
  const slots = slotsData?.data ?? []

  const { data: reviewsData } = useGetBathhousesIdReviews(id ?? '', { page: reviewPage, page_size: 5 }, {
    query: { enabled: !!id },
  })
  const reviews = reviewsData?.data ?? []
  const reviewMeta = reviewsData?.meta

  const { data: similarData } = useGetBathhousesIdSimilar(id ?? '', { limit: 4 }, {
    query: { enabled: !!id },
  })
  const similar = similarData?.data ?? []

  const { data: galleryData } = useGetBathhousesIdGallery(id ?? '', { page: 1, page_size: 20 }, {
    query: { enabled: !!id },
  })
  const gallery = galleryData?.data ?? []

  const { data: schemaData } = useGetBathhousesIdSchema(id ?? '', {
    query: { enabled: !!id },
  })

  const { data: transportData } = useQuery({
    queryKey: [`/bathhouses/${id}/transport`],
    queryFn: () => axiosInstance.get<{ success: boolean; data: { items: TransportItem[] } }>(`/bathhouses/${id}/transport`).then((r) => r.data),
    enabled: !!id,
    staleTime: 7 * 24 * 60 * 60 * 1000, // 7 days - transport infrastructure rarely changes
  })
  const transportItems = transportData?.data?.items ?? []

  const favoriteMutation = usePostBathhousesIdFavorite({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: [`/bathhouses/${id}`] })
        queryClient.invalidateQueries({ queryKey: ['/my/favorites'] })
      },
      onError: () => message.error('Не удалось обновить избранное'),
    },
  })

  // JSON-LD schema - inject via textContent (inherently safe, no XSS risk)
  const schemaJsonLd = (() => {
    try {
      if (schemaData && typeof schemaData === 'object') {
        return JSON.stringify(schemaData)
      }
      if (typeof schemaData === 'string') {
        return JSON.stringify(JSON.parse(schemaData))
      }
    } catch {
      // serialization failed, skip
    }
    return null
  })()

  useEffect(() => {
    if (!schemaJsonLd) return
    const script = document.createElement('script')
    script.type = 'application/ld+json'
    script.textContent = schemaJsonLd
    document.head.appendChild(script)
    return () => {
      document.head.removeChild(script)
    }
  }, [schemaJsonLd])

  if (isLoading) {
    return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  }

  if (!bathhouse) {
    return <Empty description="Баня не найдена" />
  }

  const allPhotos = [
    ...photos.map((p) => p.url).filter(Boolean),
    ...(bathhouse.images ?? []),
    ...gallery.map((g) => g.url).filter(Boolean),
  ] as string[]
  const uniquePhotos = [...new Set(allPhotos)]

  const amenities = AMENITY_LIST.filter(
    (a) => bathhouse[a.key as keyof typeof bathhouse],
  )

  return (
    <div>
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/client')}
        style={{ marginBottom: 16 }}
      >
        К поиску
      </Button>

      <Row gutter={[24, 24]}>
        <Col xs={24} lg={16}>
          <Card>
            {uniquePhotos.length > 0 ? (
              <Image.PreviewGroup>
                <Row gutter={[8, 8]}>
                  {uniquePhotos.slice(0, 5).map((url, i) => (
                    <Col key={url} span={i === 0 ? 24 : 6}>
                      <Image
                        src={url}
                        alt={`${bathhouse.name} фото ${i + 1}`}
                        style={{ borderRadius: 8, objectFit: 'cover', width: '100%', height: i === 0 ? 300 : 100 }}
                      />
                    </Col>
                  ))}
                </Row>
                {uniquePhotos.slice(5).map((url) => (
                  <Image key={url} src={url} style={{ display: 'none' }} />
                ))}
              </Image.PreviewGroup>
            ) : (
              <div style={{ height: 200, background: '#f5f5f5', borderRadius: 8, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Text type="secondary">Нет фото</Text>
              </div>
            )}

            <div style={{ marginTop: 16 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div>
                  <Title level={3} style={{ margin: 0 }}>
                    {bathhouse.name}
                    {bathhouse.is_photo_verified && (
                      <CheckCircleOutlined style={{ color: '#52c41a', marginLeft: 8, fontSize: 18 }} />
                    )}
                  </Title>
                  {bathhouse.address && (
                    <Text type="secondary">
                      <EnvironmentOutlined style={{ marginRight: 4 }} />
                      {bathhouse.address}
                    </Text>
                  )}
                </div>
                <Button
                  type={bathhouse.is_favorite ? 'primary' : 'default'}
                  icon={bathhouse.is_favorite ? <HeartFilled /> : <HeartOutlined />}
                  onClick={() => id && favoriteMutation.mutate({ id })}
                  loading={favoriteMutation.isPending}
                >
                  {bathhouse.is_favorite ? 'В избранном' : 'В избранное'}
                </Button>
              </div>

              <Space style={{ marginTop: 8 }}>
                <Rate disabled allowHalf value={bathhouse.rating ?? 0} />
                <Text>{bathhouse.rating?.toFixed(1)} ({bathhouse.review_count ?? 0} отзывов)</Text>
              </Space>

              <Divider />

              <Paragraph>{bathhouse.description}</Paragraph>

              <Descriptions column={{ xs: 1, sm: 2 }} bordered size="small" style={{ marginTop: 16 }}>
                <Descriptions.Item label="Цена за час">
                  {bathhouse.price_per_hour ? formatPrice(bathhouse.price_per_hour) : '—'}
                </Descriptions.Item>
                <Descriptions.Item label="Мин. длительность">
                  {bathhouse.min_duration ? `${bathhouse.min_duration} ч` : '—'}
                </Descriptions.Item>
                <Descriptions.Item label="Макс. гостей">
                  {bathhouse.max_guests ?? '—'}
                </Descriptions.Item>
                <Descriptions.Item label="Политика отмены">
                  {(() => {
                    const policy = (bathhouse as Record<string, unknown>).cancellation_policy as string
                    switch (policy) {
                      case 'flexible': return 'Гибкая'
                      case 'moderate': return 'Умеренная'
                      case 'strict': return 'Строгая'
                      default: return 'Гибкая'
                    }
                  })()}
                </Descriptions.Item>
                {(() => {
                  const depositPercent = (bathhouse as Record<string, unknown>).security_deposit_percent as number
                  return depositPercent > 0 ? (
                    <Descriptions.Item label="Залог">
                      {depositPercent}% от базовой цены (возврат через 48ч после визита)
                    </Descriptions.Item>
                  ) : null
                })()}
              </Descriptions>

              {amenities.length > 0 && (
                <div style={{ marginTop: 16 }}>
                  <Text strong>Удобства:</Text>
                  <div style={{ marginTop: 8 }}>
                    {amenities.map((a) => (
                      <Tag key={a.key} color="blue" style={{ marginBottom: 4 }}>
                        {a.label}
                      </Tag>
                    ))}
                  </div>
                </div>
              )}

              {bathhouse.working_hours && bathhouse.working_hours.length > 0 && (
                <div style={{ marginTop: 16 }}>
                  <Text strong>Режим работы:</Text>
                  <div style={{ marginTop: 8 }}>
                    {bathhouse.working_hours.map((wh) => (
                      <div key={wh.day_of_week}>
                        <Text>
                          {formatDayOfWeek(wh.day_of_week ?? 0)}: {wh.open_time} — {wh.close_time}
                        </Text>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {transportItems.length > 0 && (
                <div style={{ marginTop: 16 }}>
                  <Text strong>Транспорт рядом:</Text>
                  <div style={{ marginTop: 8 }}>
                    {transportItems.map((item, idx) => {
                      const meta = TRANSPORT_LABELS[item.type] ?? { label: item.type, color: '#999' }
                      return (
                        <div key={`${item.type}-${idx}`} style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                          {item.type === 'metro' ? (
                            <NodeIndexOutlined style={{ color: meta.color }} />
                          ) : item.type === 'parking' ? (
                            <CarOutlined style={{ color: meta.color }} />
                          ) : (
                            <EnvironmentOutlined style={{ color: meta.color }} />
                          )}
                          <Tag color={meta.color} style={{ margin: 0 }}>{meta.label}</Tag>
                          <Text>{item.name}</Text>
                          <Text type="secondary">— {formatDistance(item.distance_meters)}</Text>
                        </div>
                      )
                    })}
                  </div>
                </div>
              )}
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title="Доступные слоты">
            <DatePicker
              value={dayjs(selectedDate)}
              onChange={(d) => d && setSelectedDate(d.format('YYYY-MM-DD'))}
              disabledDate={(d) => d.isBefore(dayjs(), 'day')}
              style={{ width: '100%', marginBottom: 16 }}
            />
            <Spin spinning={slotsLoading}>
              {slots.length === 0 ? (
                <Empty description="Нет доступных слотов" image={Empty.PRESENTED_IMAGE_SIMPLE} />
              ) : (
                <Space direction="vertical" style={{ width: '100%' }} size={8}>
                  {slots.map((slot) => (
                    <Card
                      key={`${slot.startTime}-${slot.endTime}`}
                      size="small"
                      style={{
                        opacity: slot.available ? 1 : 0.5,
                        cursor: slot.available ? 'pointer' : 'not-allowed',
                      }}
                      onClick={() => {
                        if (slot.available) {
                          navigate(`/client/booking/new?bathhouse=${id}&date=${selectedDate}&from=${slot.startTime}&to=${slot.endTime}`)
                        }
                      }}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <Space>
                          <ClockCircleOutlined />
                          <Text>{typeof slot.startTime === 'string' && slot.startTime.length > 5 ? slot.startTime.slice(11, 16) : slot.startTime} — {typeof slot.endTime === 'string' && slot.endTime.length > 5 ? slot.endTime.slice(11, 16) : slot.endTime}</Text>
                        </Space>
                        <div>
                          {slot.price != null && <Text strong>{formatPrice(slot.price)}</Text>}
                          {!slot.available && <Tag color="red" style={{ marginLeft: 8 }}>Занято</Tag>}
                        </div>
                      </div>
                    </Card>
                  ))}
                </Space>
              )}
            </Spin>
          </Card>
        </Col>
      </Row>

      <Divider />

      <Title level={4}>Отзывы ({bathhouse.review_count ?? 0})</Title>
      {reviews.length === 0 ? (
        <Empty description="Нет отзывов" image={Empty.PRESENTED_IMAGE_SIMPLE} />
      ) : (
        <>
          {reviews.map((review) => (
            <ReviewCard
              key={review.id}
              review={review}
              showActions={true}
              isAuthor={!!currentUser?.id && currentUser.id === review.user_id}
              onEdit={(reviewId) => navigate(`/client/review?bathhouse=${id}&edit=${reviewId}`)}
              onDeleted={() => queryClient.invalidateQueries({ queryKey: [`/bathhouses/${id}/reviews`] })}
            />
          ))}
          {reviewMeta && reviewMeta.total_pages && reviewMeta.total_pages > 1 && (
            <Pagination
              current={reviewPage}
              pageSize={5}
              total={reviewMeta.total_count}
              onChange={setReviewPage}
              style={{ marginTop: 16 }}
            />
          )}
        </>
      )}

      {similar.length > 0 && (
        <>
          <Divider />
          <Title level={4}>Похожие бани</Title>
          <Row gutter={[16, 16]}>
            {similar.map((s) => (
              <Col key={s.id} xs={24} sm={12} md={6}>
                <BathhouseCard
                  bathhouse={{
                    id: s.id,
                    name: s.name,
                    address: s.address,
                    price_per_hour: s.price_per_hour,
                    rating: s.rating,
                    review_count: s.review_count,
                    city_id: s.city_id,
                    latitude: s.latitude,
                    longitude: s.longitude,
                    description: s.description,
                  }}
                  showFavorite={false}
                />
              </Col>
            ))}
          </Row>
        </>
      )}
    </div>
  )
}
