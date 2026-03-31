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
  Avatar,
  Alert,
} from 'antd'
import {
  ArrowLeftOutlined,
  EnvironmentOutlined,
  HeartOutlined,
  HeartFilled,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ThunderboltOutlined,
  UserOutlined,
  WalletOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { App } from 'antd'
import { useGetBathhousesBySlugSlug, useGetBathhousesIdAvailableSlots, useGetBathhousesIdSchema } from '@/api/generated/bathhouses/bathhouses'
import { useGetBathhousesIdPhotos } from '@/api/generated/photos/photos'
import { useGetBathhousesIdReviews } from '@/api/generated/reviews/reviews'
import { useGetBathhousesIdSimilar } from '@/api/generated/recommendations/recommendations'
import { useGetBathhousesIdGallery } from '@/api/generated/review-media/review-media'
import { usePostBathhousesIdFavorite } from '@/api/generated/favorites/favorites'
import { useGetMyWallet } from '@/api/generated/wallet/wallet'
import { axiosInstance } from '@/api/axios-instance'
import { formatPrice, formatDayOfWeek } from '@/lib/format'
import { useAuthStore } from '@/stores/auth'
import ReviewCard from '@/components/ReviewCard'
import ShareButton from '@/components/ShareButton'
import TransportAccessibility, { type TransportItem } from '@/components/TransportAccessibility'
import SimilarBathhouses from '@/components/SimilarBathhouses'
import PriceBreakdown from '@/components/PriceBreakdown'

const { Title, Text, Paragraph } = Typography

const CANCELLATION_POLICY_DETAILS: Record<string, { label: string; description: string }> = {
  flexible: {
    label: 'Гибкая',
    description: 'Бесплатная отмена за 24+ ч. Возврат 50% менее чем за 24 ч.',
  },
  moderate: {
    label: 'Умеренная',
    description: 'Бесплатная отмена за 72+ ч. Возврат 50% за 24-72 ч. Без возврата менее 24 ч.',
  },
  strict: {
    label: 'Строгая',
    description: 'Бесплатная отмена за 7+ дней. Возврат 50% за 3-7 дней. Без возврата менее 3 дней.',
  },
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
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { message } = App.useApp()
  const currentUser = useAuthStore((s) => s.user)
  const [selectedDate, setSelectedDate] = useState<string>(dayjs().format('YYYY-MM-DD'))
  const [reviewPage, setReviewPage] = useState(1)

  const { data: bathhouseData, isLoading } = useGetBathhousesBySlugSlug(slug ?? '', {
    query: { enabled: !!slug },
  })
  const bathhouse = bathhouseData?.data
  const id = bathhouse?.id

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

  const { data: similarData } = useGetBathhousesIdSimilar(id ?? '', { limit: 6 }, {
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

  const { data: walletData } = useGetMyWallet({
    query: { enabled: !!currentUser },
  })
  const walletBalance = (walletData as Record<string, unknown>)?.data as { balance?: number } | undefined

  const favoriteMutation = usePostBathhousesIdFavorite({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: [`/bathhouses/by-slug/${slug}`] })
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
    if (!id || !currentUser) return
    axiosInstance.post('/my/recently-viewed', { bathhouse_id: id }).catch(() => {})
  }, [id, currentUser])

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
                <Space>
                  <ShareButton
                    url={`${window.location.origin}/bathhouses/${slug}`}
                    title={bathhouse.name ?? 'Баня на Bani'}
                    text={bathhouse.description?.slice(0, 100) ?? ''}
                  />
                  <Button
                    type={bathhouse.is_favorite ? 'primary' : 'default'}
                    icon={bathhouse.is_favorite ? <HeartFilled /> : <HeartOutlined />}
                    onClick={() => id && favoriteMutation.mutate({ id })}
                    loading={favoriteMutation.isPending}
                  >
                    {bathhouse.is_favorite ? 'В избранном' : 'В избранное'}
                  </Button>
                </Space>
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
                {bathhouse.area_avg_price_per_hour ? (
                  <Descriptions.Item label="Средняя цена в районе">
                    {formatPrice(bathhouse.area_avg_price_per_hour)}
                  </Descriptions.Item>
                ) : null}
                <Descriptions.Item label="Мин. длительность">
                  {bathhouse.min_duration ? `${bathhouse.min_duration} ч` : '—'}
                </Descriptions.Item>
                <Descriptions.Item label="Макс. гостей">
                  {bathhouse.max_guests ?? '—'}
                </Descriptions.Item>
                <Descriptions.Item label="Режим бронирования">
                  {(() => {
                    const mode = (bathhouse as Record<string, unknown>).booking_mode as string
                    return mode === 'request' ? (
                      <Tag color="orange">По запросу</Tag>
                    ) : (
                      <Tag icon={<ThunderboltOutlined />} color="green">Мгновенное</Tag>
                    )
                  })()}
                </Descriptions.Item>
                <Descriptions.Item label="Политика отмены">
                  {(() => {
                    const policy = (bathhouse as Record<string, unknown>).cancellation_policy as string
                    const details = CANCELLATION_POLICY_DETAILS[policy] ?? CANCELLATION_POLICY_DETAILS.flexible
                    return (
                      <div>
                        <Text strong>{details.label}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: 12 }}>{details.description}</Text>
                      </div>
                    )
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

              <TransportAccessibility items={transportItems} />

              {(bathhouse as Record<string, unknown>).visiting_rules && (
                <div style={{ marginTop: 16 }}>
                  <Text strong>Правила посещения:</Text>
                  <Alert
                    type="info"
                    showIcon={false}
                    style={{ marginTop: 8 }}
                    message={
                      <Paragraph style={{ margin: 0, whiteSpace: 'pre-line' }}>
                        {(bathhouse as Record<string, unknown>).visiting_rules as string}
                      </Paragraph>
                    }
                  />
                </div>
              )}

              {bathhouse.owner_profile && (
                <div style={{ marginTop: 16 }}>
                  <Text strong>Владелец:</Text>
                  <div style={{ marginTop: 8, display: 'flex', alignItems: 'center', gap: 12 }}>
                    <Avatar
                      size={48}
                      src={bathhouse.owner_profile.avatar_url}
                      icon={!bathhouse.owner_profile.avatar_url && <UserOutlined />}
                    />
                    <div>
                      <Text strong>{bathhouse.owner_profile.name || 'Владелец'}</Text>
                      <div>
                        {(bathhouse.owner_profile.rating ?? 0) > 0 && (
                          <Text type="secondary" style={{ marginRight: 12 }}>
                            Рейтинг: {bathhouse.owner_profile.rating!.toFixed(1)}
                          </Text>
                        )}
                        <Text type="secondary" style={{ marginRight: 12 }}>
                          Объектов: {bathhouse.owner_profile.object_count}
                        </Text>
                        {bathhouse.owner_profile.member_since && (
                          <Text type="secondary">
                            На платформе с {dayjs(bathhouse.owner_profile.member_since).format('MMM YYYY')}
                          </Text>
                        )}
                      </div>
                    </div>
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

          {bathhouse.price_per_hour && (
            <Card title="Примерная стоимость" size="small" style={{ marginTop: 16 }}>
              <PriceBreakdown
                basePrice={bathhouse.price_per_hour}
                serviceFee={bathhouse.price_per_hour ? Math.round(bathhouse.price_per_hour * 0.1) : undefined}
                areaAveragePrice={bathhouse.area_avg_price_per_hour}
              />
            </Card>
          )}

          {currentUser && walletBalance?.balance != null && walletBalance.balance > 0 && bathhouse.price_per_hour && walletBalance.balance >= bathhouse.price_per_hour && (
            <Card size="small" style={{ marginTop: 16 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                <WalletOutlined style={{ fontSize: 18, color: '#52c41a' }} />
                <Text>Баланс кошелька: <Text strong>{formatPrice(walletBalance.balance)}</Text></Text>
              </div>
              <Button
                type="primary"
                ghost
                block
                icon={<WalletOutlined />}
                onClick={() => navigate(`/client/booking/new?bathhouse=${id}&payment_method=wallet`)}
              >
                Оплатить из кошелька
              </Button>
            </Card>
          )}
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
          <SimilarBathhouses items={similar} />
        </>
      )}
    </div>
  )
}
