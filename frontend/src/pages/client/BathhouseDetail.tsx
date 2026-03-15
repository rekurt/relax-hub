import { useState } from 'react'
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
  List,
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
} from '@ant-design/icons'
import dayjs from 'dayjs'
import { useQueryClient } from '@tanstack/react-query'
import { message } from 'antd'
import { useGetBathhousesId, useGetBathhousesIdAvailableSlots, useGetBathhousesIdSchema } from '@/api/generated/bathhouses/bathhouses'
import { useGetBathhousesIdPhotos } from '@/api/generated/photos/photos'
import { useGetBathhousesIdReviews } from '@/api/generated/reviews/reviews'
import { useGetBathhousesIdSimilar } from '@/api/generated/recommendations/recommendations'
import { useGetBathhousesIdGallery } from '@/api/generated/review-media/review-media'
import { usePostBathhousesIdFavorite } from '@/api/generated/favorites/favorites'
import { formatPrice, formatDayOfWeek } from '@/lib/format'
import BathhouseCard from '@/components/BathhouseCard'

const { Title, Text, Paragraph } = Typography

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

  const favoriteMutation = usePostBathhousesIdFavorite({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: [`/bathhouses/${id}`] })
        queryClient.invalidateQueries({ queryKey: ['/my/favorites'] })
      },
      onError: () => message.error('Не удалось обновить избранное'),
    },
  })

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

  // JSON-LD schema - the endpoint returns raw JSON-LD string
  const schemaJsonLd = typeof schemaData === 'string' ? schemaData : null

  return (
    <div>
      {schemaJsonLd && (
        <script type="application/ld+json">{schemaJsonLd}</script>
      )}

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
                          <Text>{slot.startTime} — {slot.endTime}</Text>
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
          <List
            dataSource={reviews}
            renderItem={(review) => (
              <List.Item>
                <div style={{ width: '100%' }}>
                  <Space>
                    <Rate disabled value={review.rating ?? 0} style={{ fontSize: 14 }} />
                    <Text type="secondary">{review.created_at ? dayjs(review.created_at).format('DD.MM.YYYY') : ''}</Text>
                  </Space>
                  <Paragraph style={{ marginTop: 8, marginBottom: 4 }}>{review.text}</Paragraph>
                  {review.media && review.media.length > 0 && (
                    <Image.PreviewGroup>
                      <Space>
                        {review.media.map((m) => (
                          <Image
                            key={m.id}
                            src={m.thumbnail_url ?? m.url}
                            alt="Фото отзыва"
                            width={80}
                            height={80}
                            style={{ borderRadius: 4, objectFit: 'cover' }}
                          />
                        ))}
                      </Space>
                    </Image.PreviewGroup>
                  )}
                  {review.owner_response && (
                    <Card size="small" style={{ marginTop: 8, background: '#f9f9f9' }}>
                      <Text strong>Ответ владельца:</Text>
                      <Paragraph style={{ margin: '4px 0 0' }}>{review.owner_response}</Paragraph>
                    </Card>
                  )}
                </div>
              </List.Item>
            )}
          />
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
