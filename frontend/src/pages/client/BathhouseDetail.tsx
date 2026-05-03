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
} from '@/components/design/system'
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
} from '@/components/design/icons'
import dayjs from 'dayjs'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { App } from '@/components/design/system'
import { useGetBathhousesBySlugSlug, useGetBathhousesIdAvailableSlots, useGetBathhousesIdSchema } from '@/api/generated/bathhouses/bathhouses'
import { useGetBathhousesIdPhotos } from '@/api/generated/photos/photos'
import { useGetBathhousesIdReviews } from '@/api/generated/reviews/reviews'
import { useGetBathhousesIdSimilar } from '@/api/generated/recommendations/recommendations'
import { useGetBathhousesIdGallery } from '@/api/generated/review-media/review-media'
import { usePostBathhousesIdFavorite } from '@/api/generated/favorites/favorites'
import { useGetMyWallet } from '@/api/generated/wallet/wallet'
import { axiosInstance } from '@/api/axios-instance'
import {
  getBookingModeLabel,
  getBookingModeTrustCopy,
  getCancellationPolicyDetails,
  getCancellationPolicyLabel,
  getDepositSummary,
} from '@/lib/booking-flow'
import { formatPrice, formatDayOfWeek } from '@/lib/format'
import { useAuthStore } from '@/stores/auth'
import PublicState from '@/components/PublicState'
import ReviewCard from '@/components/ReviewCard'
import ShareButton from '@/components/ShareButton'
import TransportAccessibility, { type TransportItem } from '@/components/TransportAccessibility'
import SimilarBathhouses from '@/components/SimilarBathhouses'
import PriceBreakdown from '@/components/PriceBreakdown'
import ContiguousSlotSelector from '@/components/ContiguousSlotSelector'
import { resolveAssetUrl } from '@/lib/asset-url'
import { formatSlotTimeLabel, getRangeHours, resolveSlotRangeSelection, type SlotRangeSelection } from '@/lib/slot-selection'

const { Title, Text, Paragraph } = Typography
const REVIEW_PREVIEW_LIMIT = 3

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
  const [reviewView, setReviewView] = useState({ bathhouseId: '', page: 1, expanded: false })
  const [selectedSlotRange, setSelectedSlotRange] = useState<SlotRangeSelection | null>(null)

  const { data: bathhouseData, isLoading, isError: bathhouseIsError, refetch: refetchBathhouse } = useGetBathhousesBySlugSlug(slug ?? '', {
    query: { enabled: !!slug },
  })
  const bathhouse = bathhouseData?.data
  const id = bathhouse?.id
  const reviewBathhouseId = id ?? ''
  const activeReviewView = reviewView.bathhouseId === reviewBathhouseId
    ? reviewView
    : { bathhouseId: reviewBathhouseId, page: 1, expanded: false }
  const reviewPage = activeReviewView.page
  const reviewsExpanded = activeReviewView.expanded

  const { data: photosData } = useGetBathhousesIdPhotos(id ?? '', {
    query: { enabled: !!id },
  })
  const photos = photosData?.data ?? []

  const {
    data: slotsData,
    isLoading: slotsLoading,
    isError: slotsIsError,
    refetch: refetchSlots,
  } = useGetBathhousesIdAvailableSlots(
    id ?? '',
    { date: selectedDate },
    { query: { enabled: !!id && !!selectedDate } },
  )
  const slots = slotsData?.data ?? []

  const reviewPageSize = reviewsExpanded ? 20 : REVIEW_PREVIEW_LIMIT
  const { data: reviewsData } = useGetBathhousesIdReviews(id ?? '', { page: reviewPage, page_size: reviewPageSize }, {
    query: { enabled: !!id },
  })
  const reviews = reviewsData?.data ?? []
  const reviewMeta = reviewsData?.meta

  const { data: similarData, isError: similarIsError, refetch: refetchSimilar } = useGetBathhousesIdSimilar(id ?? '', { limit: 6 }, {
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

  const { data: transportData, isError: transportIsError, refetch: refetchTransport } = useQuery({
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

  const minDurationHours = Math.max(1, bathhouse?.min_duration ?? 1)
  const resolvedSlotRange = resolveSlotRangeSelection(slots, selectedSlotRange)
  const selectedRangeHours = resolvedSlotRange ? getRangeHours(resolvedSlotRange.from, resolvedSlotRange.to) : 0
  const selectedRange = resolvedSlotRange && selectedRangeHours >= minDurationHours
    ? {
      ...resolvedSlotRange,
      hours: selectedRangeHours,
    }
    : null
  const selectedRangeNeedsHours = resolvedSlotRange ? Math.max(0, minDurationHours - selectedRangeHours) : 0
  const displayedReviewCount = reviewMeta?.total_count ?? bathhouse?.review_count ?? reviews.length
  const visibleReviews = reviewsExpanded ? reviews : reviews.slice(0, REVIEW_PREVIEW_LIMIT)
  const hasHiddenReviews = !reviewsExpanded && displayedReviewCount > visibleReviews.length

  if (isLoading) {
    return (
      <PublicState
        kind="loading"
        title="Загружаем баню"
        description="Собираем описание, фото и ближайшие доступные слоты."
      />
    )
  }

  if (bathhouseIsError) {
    return (
      <PublicState
        kind="error"
        title="Не удалось загрузить карточку бани"
        description="Повторите попытку или вернитесь в каталог."
        actionText="Повторить"
        onAction={() => void refetchBathhouse()}
        secondaryActionText="В каталог"
        secondaryActionLink="/catalog"
      />
    )
  }

  if (!bathhouse) {
    return (
      <PublicState
        kind="empty"
        title="Баня не найдена"
        description="Возможно, объект уже недоступен или ссылка устарела."
        actionText="Вернуться в каталог"
        actionLink="/catalog"
      />
    )
  }

  const allPhotos = [
    ...photos.map((p) => resolveAssetUrl(p.url)).filter(Boolean),
    ...(bathhouse.images ?? []).map((url) => resolveAssetUrl(url)),
    ...gallery.map((g) => resolveAssetUrl(g.url)).filter(Boolean),
  ] as string[]
  const uniquePhotos = [...new Set(allPhotos)]

  const amenities = AMENITY_LIST.filter(
    (a) => bathhouse[a.key as keyof typeof bathhouse],
  )
  const bookingMode = (bathhouse as Record<string, unknown>).booking_mode as string | undefined
  const cancellationPolicy = (bathhouse as Record<string, unknown>).cancellation_policy as string | undefined
  const depositPercent = (bathhouse as Record<string, unknown>).security_deposit_percent as number | undefined
  const bookingModeLabel = getBookingModeLabel(bookingMode)
  const bookingModeTrustCopy = getBookingModeTrustCopy(bookingMode)
  const cancellationDetails = getCancellationPolicyDetails(cancellationPolicy)
  const minimumDurationLabel = bathhouse.min_duration ? `от ${bathhouse.min_duration} ч` : 'Без ограничения'
  const areaAveragePrice = bathhouse.area_avg_price_per_hour
  const showAreaAveragePrice = areaAveragePrice != null
    && areaAveragePrice > 0
    && bathhouse.price_per_hour != null
    && areaAveragePrice >= Math.round(bathhouse.price_per_hour * 0.2)
    && areaAveragePrice <= Math.round(bathhouse.price_per_hour * 5)

  const scrollToSection = (sectionId: string) => {
    const section = document.getElementById(sectionId)
    section?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  return (
    <div className="rh-stack">
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/catalog')}
      >
        К поиску
      </Button>

      <section className="rh-hero-panel rh-hero-panel--dark rh-detail-hero">
        <div className="rh-detail-hero__header">
          <div className="rh-stack">
            <div className="rh-hero-panel__eyebrow">Публичное бронирование</div>
            <h1 className="rh-hero-panel__title">
              {bathhouse.name}
              {bathhouse.is_photo_verified && <CheckCircleOutlined style={{ marginLeft: 10, fontSize: 22 }} />}
            </h1>
            <div className="rh-detail-hero__lead">
              {bathhouse.address && (
                <span className="rh-detail-hero__lead-item">
                  <EnvironmentOutlined />
                  {bathhouse.address}
                </span>
              )}
              <span className="rh-detail-hero__lead-item">
                <Rate disabled allowHalf value={bathhouse.rating ?? 0} />
                <span>{bathhouse.rating?.toFixed(1)} · {displayedReviewCount} отзывов</span>
              </span>
            </div>
            <div className="rh-hero-panel__description">
              {bathhouse.description || 'Свободные слоты, правила и стоимость собраны прямо на этой странице.'}
            </div>
          </div>

          <div className="rh-detail-hero__actions">
            <Button type="primary" size="large" onClick={() => scrollToSection('bathhouse-slots')}>
              Выбрать слот
            </Button>
            <Button size="large" onClick={() => scrollToSection('bathhouse-rules')}>
              Правила и условия
            </Button>
          </div>
        </div>

        <div className="rh-hero-panel__meta">
          <div className="rh-hero-panel__meta-item">
            <span className="rh-hero-panel__meta-label">Цена от</span>
            <div className="rh-hero-panel__meta-value">
              {bathhouse.price_per_hour ? formatPrice(bathhouse.price_per_hour) : 'По запросу'}
            </div>
          </div>
          <div className="rh-hero-panel__meta-item">
            <span className="rh-hero-panel__meta-label">Минимум</span>
            <div className="rh-hero-panel__meta-value">{minimumDurationLabel}</div>
          </div>
          <div className="rh-hero-panel__meta-item">
            <span className="rh-hero-panel__meta-label">Подтверждение</span>
            <div className="rh-hero-panel__meta-value">{bookingModeTrustCopy}</div>
          </div>
          <div className="rh-hero-panel__meta-item">
            <span className="rh-hero-panel__meta-label">Отмена</span>
            <div className="rh-hero-panel__meta-value">{getCancellationPolicyLabel(cancellationPolicy)} отмена</div>
          </div>
        </div>
      </section>

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
                        style={{ borderRadius: 20, objectFit: 'cover', width: '100%', height: i === 0 ? 300 : 100 }}
                      />
                    </Col>
                  ))}
                </Row>
                {uniquePhotos.slice(5).map((url) => (
                  <Image key={url} src={url} style={{ display: 'none' }} />
                ))}
              </Image.PreviewGroup>
            ) : (
              <div style={{ height: 200, background: 'rgba(248, 244, 236, 0.78)', borderRadius: 20, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Text type="secondary">Нет фото</Text>
              </div>
            )}

            <div style={{ marginTop: 16 }}>
              <div className="rh-toolbar">
                <div>
                  <Title level={3} style={{ margin: 0 }}>Описание и условия</Title>
                  <Text type="secondary">Сначала то, что влияет на бронирование, затем вторичные детали объекта.</Text>
                </div>
                <Space>
                  <ShareButton
                    url={`${window.location.origin}/bathhouses/${slug}`}
                    title={bathhouse.name ?? 'Баня на Bani'}
                    text={bathhouse.description?.slice(0, 100) ?? ''}
                  />
                  {currentUser?.role === 'client' && (
                    <Button
                      type={bathhouse.is_favorite ? 'primary' : 'default'}
                      icon={bathhouse.is_favorite ? <HeartFilled /> : <HeartOutlined />}
                      onClick={() => id && favoriteMutation.mutate({ id })}
                      loading={favoriteMutation.isPending}
                    >
                      {bathhouse.is_favorite ? 'В избранном' : 'В избранное'}
                    </Button>
                  )}
                </Space>
              </div>

              <Divider />

              <Paragraph>{bathhouse.description}</Paragraph>

              <Descriptions column={{ xs: 1, sm: 2 }} bordered size="small" style={{ marginTop: 16 }}>
                <Descriptions.Item label="Цена за час">
                  {bathhouse.price_per_hour ? formatPrice(bathhouse.price_per_hour) : '—'}
                </Descriptions.Item>
                {showAreaAveragePrice ? (
                  <Descriptions.Item label="Средняя цена в районе">
                    {formatPrice(areaAveragePrice)}
                  </Descriptions.Item>
                ) : null}
                <Descriptions.Item label="Мин. длительность">
                  {bathhouse.min_duration ? `${bathhouse.min_duration} ч` : '—'}
                </Descriptions.Item>
                <Descriptions.Item label="Макс. гостей">
                  {bathhouse.max_guests ?? '—'}
                </Descriptions.Item>
                <Descriptions.Item label="Режим бронирования">
                  {bookingMode === 'request' ? (
                    <Tag color="orange">{bookingModeLabel}</Tag>
                  ) : (
                    <Tag icon={<ThunderboltOutlined />} color="green">{bookingModeLabel}</Tag>
                  )}
                </Descriptions.Item>
                <Descriptions.Item label="Политика отмены">
                  <div>
                    <Text strong>{cancellationDetails.label}</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: 12 }}>{cancellationDetails.description}</Text>
                  </div>
                </Descriptions.Item>
                {depositPercent && depositPercent > 0 ? (
                  <Descriptions.Item label="Залог">
                    {depositPercent}% от базовой цены (возврат через 48ч после визита)
                  </Descriptions.Item>
                ) : null}
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

              {transportIsError ? (
                <div style={{ marginTop: 16 }}>
                  <PublicState
                    kind="degraded"
                    compact
                    title="Транспортная информация временно недоступна"
                    description="Попробуйте обновить этот блок позже."
                    actionText="Повторить"
                    onAction={() => void refetchTransport()}
                  />
                </div>
              ) : (
                <TransportAccessibility items={transportItems} />
              )}

              {!!(bathhouse as Record<string, unknown>).visiting_rules && (
                <div id="bathhouse-rules" style={{ marginTop: 16 }}>
                  <Text strong>Правила посещения:</Text>
                  <Alert
                    type="info"
                    showIcon={false}
                    style={{ marginTop: 8 }}
                    title={
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
                      src={resolveAssetUrl(bathhouse.owner_profile.avatar_url)}
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
          <div className="rh-detail-rail">
            <Card id="bathhouse-slots" title="Свободные слоты">
              <div className="rh-section-card">
                <Text type="secondary">
                  Сначала дата и время, потом checkout. Никаких скрытых условий перед переходом к брони.
                </Text>
                <DatePicker
                  value={dayjs(selectedDate)}
                  onChange={(d) => {
                    if (!d) return
                    setSelectedDate(d.format('YYYY-MM-DD'))
                    setSelectedSlotRange(null)
                  }}
                  disabledDate={(d) => d.isBefore(dayjs(), 'day')}
                  style={{ width: '100%' }}
                />
                <Spin spinning={slotsLoading && !slotsIsError}>
                  {slotsIsError ? (
                    <PublicState
                      kind="error"
                      compact
                      title="Не удалось загрузить доступные слоты"
                      description="Обновите список слотов или выберите другую дату."
                      actionText="Обновить слоты"
                      onAction={() => void refetchSlots()}
                    />
                ) : slots.length === 0 ? (
                  <Empty description="Нет доступных слотов" image={Empty.PRESENTED_IMAGE_SIMPLE} />
                ) : (
                  <Space orientation="vertical" style={{ width: '100%' }} size={12}>
                    <ContiguousSlotSelector
                      slots={slots}
                      value={selectedSlotRange}
                      onChange={setSelectedSlotRange}
                      minDurationHours={minDurationHours}
                      label={`1. Выберите непрерывный интервал${minDurationHours > 1 ? ` (от ${minDurationHours} ч)` : ''}`}
                      description="Выбирайте соседние слоты подряд в одном блоке. Повторный клик по краю диапазона сокращает интервал."
                    />

                    {resolvedSlotRange && (
                      <Alert
                        type={selectedRange ? 'success' : 'info'}
                        showIcon={false}
                        title={`${formatSlotTimeLabel(resolvedSlotRange.from)} - ${formatSlotTimeLabel(resolvedSlotRange.to)} · ${selectedRangeHours} ч`}
                        description={selectedRange
                          ? 'В checkout уже попадет выбранный непрерывный интервал.'
                          : `Добавьте ещё ${selectedRangeNeedsHours} ч, чтобы перейти к бронированию.`}
                        action={(
                          <Button
                            type="primary"
                            disabled={!selectedRange}
                            onClick={() => {
                              if (!selectedRange) return
                              navigate(`/checkout?bathhouse=${id}&date=${selectedDate}&from=${selectedRange.from}&to=${selectedRange.to}`)
                            }}
                          >
                            Забронировать
                          </Button>
                        )}
                      />
                    )}
                  </Space>
                )}
              </Spin>
            </div>
            </Card>

            <Card title="Что важно до бронирования" size="small">
              <div className="rh-feature-list">
                <div className="rh-feature-item">
                  <div className="rh-feature-item__icon"><ClockCircleOutlined /></div>
                  <div className="rh-feature-item__copy">
                    <div className="rh-feature-item__title">Минимум {bathhouse.min_duration ?? 1} ч</div>
                    <div className="rh-feature-item__description">Минимальная длительность совпадает с реальными правилами объекта.</div>
                  </div>
                </div>
                <div className="rh-feature-item">
                  <div className="rh-feature-item__icon"><ThunderboltOutlined /></div>
                  <div className="rh-feature-item__copy">
                    <div className="rh-feature-item__title">{bookingModeTrustCopy}</div>
                    <div className="rh-feature-item__description">Модель подтверждения известна заранее, до ввода контактов.</div>
                  </div>
                </div>
                <div className="rh-feature-item">
                  <div className="rh-feature-item__icon"><CheckCircleOutlined /></div>
                  <div className="rh-feature-item__copy">
                    <div className="rh-feature-item__title">{getCancellationPolicyLabel(cancellationPolicy)} отмена</div>
                    <div className="rh-feature-item__description">{cancellationDetails.description}</div>
                  </div>
                </div>
                <div className="rh-feature-item">
                  <div className="rh-feature-item__icon"><WalletOutlined /></div>
                  <div className="rh-feature-item__copy">
                    <div className="rh-feature-item__title">{getDepositSummary(depositPercent)}</div>
                    <div className="rh-feature-item__description">Размер залога, если он нужен, виден ещё до checkout.</div>
                  </div>
                </div>
              </div>
            </Card>

            {bathhouse.price_per_hour && (
              <Card title="Примерная стоимость" size="small">
                <PriceBreakdown
                  basePrice={bathhouse.price_per_hour}
                  serviceFee={bathhouse.price_per_hour ? Math.round(bathhouse.price_per_hour * 0.1) : undefined}
                  areaAveragePrice={showAreaAveragePrice ? areaAveragePrice : undefined}
                />
              </Card>
            )}

            {currentUser?.role === 'client' && walletBalance?.balance != null && walletBalance.balance > 0 && bathhouse.price_per_hour && walletBalance.balance >= bathhouse.price_per_hour && (
              <Card size="small">
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                  <WalletOutlined style={{ fontSize: 18, color: '#15803d' }} />
                  <Text>Баланс кошелька: <Text strong>{formatPrice(walletBalance.balance)}</Text></Text>
                </div>
                <Button
                  type="primary"
                  ghost
                  block
                  icon={<WalletOutlined />}
                  onClick={() => navigate(`/checkout?bathhouse=${id}`)}
                >
                  Перейти к бронированию
                </Button>
              </Card>
            )}
          </div>
        </Col>
      </Row>

      <Divider />

      <Title level={4}>Отзывы ({displayedReviewCount})</Title>
      {reviews.length === 0 ? (
        <Empty description="Нет отзывов" image={Empty.PRESENTED_IMAGE_SIMPLE} />
      ) : (
        <>
          {visibleReviews.map((review) => (
            <ReviewCard
              key={review.id}
              review={review}
              showActions={!!currentUser}
              isAuthor={!!currentUser?.id && currentUser.id === review.user_id}
              onEdit={(reviewId) => navigate(`/client/review?bathhouse=${id}&edit=${reviewId}`)}
              onDeleted={() => queryClient.invalidateQueries({ queryKey: [`/bathhouses/${id}/reviews`] })}
            />
          ))}
          {hasHiddenReviews && (
            <Button
              type="default"
              onClick={() => {
                setReviewView({ bathhouseId: reviewBathhouseId, page: 1, expanded: true })
              }}
            >
              Показать все отзывы
            </Button>
          )}
          {reviewsExpanded && reviewMeta && reviewMeta.total_pages && reviewMeta.total_pages > 1 && (
            <Pagination
              current={reviewPage}
              pageSize={reviewPageSize}
              total={reviewMeta.total_count}
              onChange={(page) => setReviewView({ bathhouseId: reviewBathhouseId, page, expanded: true })}
              style={{ marginTop: 16 }}
            />
          )}
        </>
      )}

      {similarIsError ? (
        <>
          <Divider />
          <PublicState
            kind="degraded"
            compact
            title="Похожие бани временно недоступны"
            description="Попробуйте обновить подборку позже."
            actionText="Повторить"
            onAction={() => void refetchSimilar()}
          />
        </>
      ) : similar.length > 0 && (
        <>
          <Divider />
          <SimilarBathhouses items={similar} />
        </>
      )}
    </div>
  )
}
