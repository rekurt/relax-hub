import { useEffect, useState, useMemo, useCallback } from 'react'
import { Typography, Input, Row, Col, Card, Select, Space, Tag, Carousel, Button } from '@/components/design/system'
import { useNavigate } from 'react-router-dom'
import {
  SearchOutlined,
  EnvironmentOutlined,
  FireOutlined,
  StarOutlined,
  GiftOutlined,
  TagOutlined,
  RocketOutlined,
} from '@/components/design/icons'
import { axiosInstance } from '@/api/axios-instance'
import { useGetRecommendations, useGetPopular } from '@/api/generated/recommendations/recommendations'
import { useGetCities } from '@/api/generated/cities/cities'
import type { InternalHandlerRecommendationResponse } from '@/api/generated/model'
import { useAuthStore } from '@/stores/auth'
import { formatPrice } from '@/lib/format'
import { resolveAssetUrl } from '@/lib/asset-url'
import { DesignListingCard } from '@/components/design'
import PublicState from '@/components/PublicState'
import RecentlyViewed from '@/components/RecentlyViewed'
import OnboardingTour from '@/components/OnboardingTour'
import { PUBLIC_SHORTCUT_CARDS } from '@/navigation/menu'

const { Title, Text } = Typography

const DISCOVERY_PROOF_POINTS = [
  {
    key: 'slots',
    eyebrow: 'Маршрут',
    title: 'Каталог с понятным входом в бронь',
    description: 'Поиск, сценарии отдыха и переход к слоту собраны в один public-first контур без кабинетообразной навигации.',
  },
  {
    key: 'filters',
    eyebrow: 'Фильтры',
    title: 'Сценарии вместо перегруза',
    description: 'Город, гости, дата и ключевые удобства вынесены на первый план, а вторичные настройки не мешают выбору.',
  },
  {
    key: 'trust',
    eyebrow: 'Доверие',
    title: 'Реальные объекты, рейтинги и ценовые ориентиры',
    description: 'Решение строится на живой выдаче, а не на рекламных обещаниях или vanity-метриках.',
  },
]

const DISCOVERY_STEPS = [
  'Выберите сценарий отдыха или сразу откройте каталог.',
  'Уточните город, гостей и дату без длинной формы.',
  'Перейдите к объекту и завершите бронь уже ближе к финалу.',
]

interface PromotionBanner {
  id: string
  title: string
  description: string
  type: 'promo' | 'welcome' | 'loyalty'
  promo_code?: string
  discount_text?: string
}

const HOME_BATHHOUSE_STATUS_TAGS = [
  { label: 'Фото проверены', matches: (item: InternalHandlerRecommendationResponse) => Boolean(item.is_photo_verified) },
  { label: 'Мгновенно', matches: (item: InternalHandlerRecommendationResponse) => item.booking_mode === 'instant' },
] as const

const HOME_BATHHOUSE_AMENITIES = [
  { label: 'Бассейн', matches: (item: InternalHandlerRecommendationResponse) => Boolean(item.has_pool) },
  { label: 'Чан', matches: (item: InternalHandlerRecommendationResponse) => Boolean(item.has_hot_tub) },
  { label: 'Сауна', matches: (item: InternalHandlerRecommendationResponse) => Boolean(item.has_sauna) },
  { label: 'Парная', matches: (item: InternalHandlerRecommendationResponse) => Boolean(item.has_steam_room) },
  { label: 'Мангал', matches: (item: InternalHandlerRecommendationResponse) => Boolean(item.has_bbq) },
  { label: 'Караоке', matches: (item: InternalHandlerRecommendationResponse) => Boolean(item.has_karaoke) },
] as const

function getBathhouseStatusTags(item: InternalHandlerRecommendationResponse) {
  return HOME_BATHHOUSE_STATUS_TAGS
    .filter((entry) => entry.matches(item))
    .map((entry) => entry.label)
}

function getBathhouseAmenities(item: InternalHandlerRecommendationResponse) {
  return HOME_BATHHOUSE_AMENITIES
    .filter((entry) => entry.matches(item))
    .map((entry) => entry.label)
}

function DiscoveryBathhouseCard({ item }: { item: InternalHandlerRecommendationResponse }) {
  const navigate = useNavigate()
  const statusTags = getBathhouseStatusTags(item)
  const amenities = getBathhouseAmenities(item)
  const tags: string[] = [...statusTags, ...amenities.slice(0, 4)]
  if (amenities.length > 4) tags.push(`+${amenities.length - 4}`)

  const lastMinute = (item as { last_minute_active?: boolean }).last_minute_active
  const lastMinutePercent = (item as { last_minute_discount_percent?: number }).last_minute_discount_percent
  const badge = lastMinute
    ? `Срочно ${lastMinutePercent ? `-${lastMinutePercent}%` : ''}`.trim()
    : undefined

  const cover = resolveAssetUrl(
    (item as { cover_photo?: string }).cover_photo
    ?? (item as { images?: string[] }).images?.[0],
  )

  return (
    <DesignListingCard
      onClick={() => navigate(`/bathhouses/${item.slug ?? item.id}`)}
      name={item.name}
      address={item.address}
      price={item.price_per_hour != null ? `${formatPrice(item.price_per_hour)}/ч` : undefined}
      rating={item.rating ?? 0}
      reviewCount={item.review_count ?? 0}
      verified={Boolean(item.is_photo_verified)}
      imageUrl={cover}
      imageAlt={item.name}
      badge={badge}
      badgeTone="red"
      tags={tags}
    />
  )
}

function PromoBanner() {
  const user = useAuthStore((s) => s.user)
  const [banners, setBanners] = useState<PromotionBanner[]>([])
  const [failedToLoad, setFailedToLoad] = useState(false)

  useEffect(() => {
    axiosInstance
      .get<{ success: boolean; data: PromotionBanner[] }>('/promotions/banners')
      .then((res) => {
        if (res.data.data?.length) {
          setBanners(res.data.data)
        }
        setFailedToLoad(false)
      })
      .catch(() => {
        setFailedToLoad(true)
      })
  }, [])

  const isNewUser = user && !user.onboarding_completed

  if (banners.length > 0) {
    const renderBanner = (banner: PromotionBanner) => {
      const gradients: Record<string, string> = {
        promo: 'linear-gradient(135deg, #b42318 0%, #d97706 100%)',
        welcome: 'linear-gradient(135deg, #15803d 0%, #0f766e 100%)',
        loyalty: 'linear-gradient(135deg, #0f766e 0%, #0a5f59 100%)',
      }
      return (
        <div key={banner.id}>
          <Card
            style={{
              background: gradients[banner.type] ?? gradients.loyalty,
              border: 'none',
              borderRadius: 12,
            }}
          >
            <div style={{ color: '#fff' }}>
              <Space>
                {banner.type === 'promo' ? (
                  <TagOutlined style={{ fontSize: 20 }} />
                ) : (
                  <GiftOutlined style={{ fontSize: 20 }} />
                )}
              <Title level={3} style={{ color: '#fff', margin: 0 }}>
                {banner.title}
              </Title>
              </Space>
              <Text style={{ color: 'rgba(255,255,255,0.85)', fontSize: 16, display: 'block', marginTop: 8 }}>
                {banner.description}
              </Text>
              {banner.promo_code && (
                <Tag
                  color="#fff"
                  style={{ color: 'var(--rh-text)', marginTop: 12, fontSize: 14, padding: '4px 12px', fontWeight: 600 }}
                >
                  {banner.promo_code}
                  {banner.discount_text && ` — ${banner.discount_text}`}
                </Tag>
              )}
            </div>
          </Card>
        </div>
      )
    }

    return (
      <div style={{ marginBottom: 24 }}>
        {banners.length === 1 ? (
          renderBanner(banners[0]!)
        ) : (
          <Carousel autoplay autoplaySpeed={5000} dots>
            {banners.map(renderBanner)}
          </Carousel>
        )}
      </div>
    )
  }

  return (
    <div style={{ marginBottom: 24 }}>
      <Card
        style={{
          marginBottom: failedToLoad ? 12 : 0,
          background: isNewUser
            ? 'linear-gradient(135deg, #15803d 0%, #0f766e 100%)'
            : 'linear-gradient(135deg, #1f4853 0%, #80502c 100%)',
          border: 'none',
          borderRadius: 12,
        }}
      >
        <div style={{ color: '#fff' }}>
          <Space>
            {isNewUser ? (
              <RocketOutlined style={{ fontSize: 20 }} />
            ) : (
              <GiftOutlined style={{ fontSize: 20 }} />
            )}
            <Title level={3} style={{ color: '#fff', margin: 0 }}>
              {isNewUser ? 'Вечер уже можно планировать' : 'Бронирование без лишних шагов'}
            </Title>
          </Space>
          <Text style={{ color: 'rgba(255,255,255,0.85)', fontSize: 16, display: 'block', marginTop: 8 }}>
            {isNewUser
              ? 'Приветственный бонус 500 ₽ уже в кошельке. Выберите сценарий, подтвердите телефон по SMS и переходите к брони.'
              : 'Каталог, понятные цены и SMS-подтверждение собраны в один короткий маршрут до бронирования.'}
          </Text>
        </div>
      </Card>
      {failedToLoad && (
        <PublicState
          kind="degraded"
          compact
          title="Акционные предложения временно недоступны"
          description="Показываем базовую подборку, пока промо-блок обновляется."
        />
      )}
    </div>
  )
}

interface CityOption {
  value: number
  label: string
}

function CitySelector({
  cities,
  selectedCityId,
  onSelect,
}: {
  cities: CityOption[]
  selectedCityId: number | undefined
  onSelect: (cityId: number) => void
}) {
  if (cities.length <= 1) return null

  return (
    <Select
      value={selectedCityId}
      onChange={onSelect}
      options={cities}
      placeholder="Выберите город"
      style={{ minWidth: 180 }}
      suffixIcon={<EnvironmentOutlined />}
      size="middle"
    />
  )
}

function PopularNearby({ detectedCityId }: { detectedCityId?: number }) {
  const { data: citiesData } = useGetCities()
  const citiesRaw = citiesData?.data
  const cities = useMemo(() => citiesRaw ?? [], [citiesRaw])

  const cityOptions: CityOption[] = useMemo(
    () => cities.map((c) => ({ value: c.id!, label: c.name! })),
    [cities],
  )

  const preferredMoscow = useMemo(
    () => cities.find((city) => city.slug === 'moscow' || city.slug === 'moskva' || city.name === 'Москва'),
    [cities],
  )

  const defaultCityId = useMemo(() => {
    if (cities.length === 0) return undefined
    if (detectedCityId && cities.some((c) => c.id === detectedCityId)) return detectedCityId
    return preferredMoscow?.id ?? cities[0]?.id
  }, [cities, detectedCityId, preferredMoscow])

  const [userSelectedCityId, setUserSelectedCityId] = useState<number | undefined>(undefined)
  const selectedCityId = userSelectedCityId ?? defaultCityId
  const selectedCity = cities.find((city) => city.id === selectedCityId)

  const { data: popularData, isLoading, isError, refetch } = useGetPopular(
    { city_id: selectedCityId ?? 0, limit: 6 },
    { query: { enabled: !!selectedCityId } },
  )
  const popular = popularData?.data ?? []

  if (cities.length === 0) return null

  return (
    <div style={{ marginBottom: 24 }}>
      <div
        style={{
          display: 'flex',
          alignItems: 'flex-start',
          justifyContent: 'space-between',
          gap: 16,
          marginBottom: 16,
          flexWrap: 'wrap',
        }}
      >
        <div>
          <Title level={4} style={{ margin: 0 }}>
            <FireOutlined style={{ marginRight: 8, color: '#b42318' }} />
            Популярные рядом
          </Title>
          <Text data-testid="popular-city-caption" type="secondary">
            {detectedCityId && selectedCity?.name
              ? `Показываем подборку рядом с вами: ${selectedCity.name}`
              : `${selectedCity?.name ?? 'Москва'} по умолчанию, если геолокация недоступна`}
          </Text>
        </div>
        <CitySelector cities={cityOptions} selectedCityId={selectedCityId} onSelect={setUserSelectedCityId} />
      </div>
      {isLoading ? null : isError ? (
        <PublicState
          kind="degraded"
          compact
          title="Популярные подборки временно недоступны"
          description="Попробуйте обновить блок позже."
          actionText="Повторить"
          onAction={() => void refetch()}
        />
      ) : popular.length === 0 ? (
        <Text type="secondary">В выбранном городе пока нет популярных бань</Text>
      ) : (
        <Row gutter={[16, 16]}>
          {popular.slice(0, 6).map((item) => (
            <Col key={item.id} xs={24} sm={12} md={8}>
              <DiscoveryBathhouseCard item={item} />
            </Col>
          ))}
        </Row>
      )}
    </div>
  )
}

function PersonalRecommendations() {
  const user = useAuthStore((s) => s.user)

  const { data: recsData, isLoading, isError, refetch } = useGetRecommendations(
    { page: 1, page_size: 6 },
    { query: { enabled: !!user } },
  )
  const recs = recsData?.data ?? []

  if (!user || isLoading) return null

  if (isError) {
    return (
      <div style={{ marginBottom: 24 }}>
        <PublicState
          kind="degraded"
          compact
          title="Персональные рекомендации временно недоступны"
          description="Попробуйте обновить подборку позже."
          actionText="Повторить"
          onAction={() => void refetch()}
        />
      </div>
    )
  }

  if (recs.length === 0) return null

  return (
    <div style={{ marginBottom: 24 }}>
      <Title level={4}>
        <StarOutlined style={{ marginRight: 8, color: '#d97706' }} />
        Рекомендации для вас
      </Title>
      <Row gutter={[16, 16]}>
        {recs.slice(0, 6).map((item) => (
          <Col key={item.id} xs={24} sm={12} md={8}>
            <DiscoveryBathhouseCard item={item} />
          </Col>
        ))}
      </Row>
    </div>
  )
}

function useGeoCity(cities: Array<{ id?: number; name?: string; latitude?: number; longitude?: number }>) {
  const [detectedCityId, setDetectedCityId] = useState<number | undefined>(undefined)
  const hasGeoApi = typeof navigator !== 'undefined' && !!navigator.geolocation

  useEffect(() => {
    if (!hasGeoApi || cities.length === 0) return
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        const { latitude, longitude } = pos.coords
        let closest: { id?: number; dist: number } = { dist: Infinity }
        for (const city of cities) {
          if (city.latitude == null || city.longitude == null) continue
          const dist = Math.hypot(city.latitude - latitude, city.longitude - longitude)
          if (dist < closest.dist) {
            closest = { id: city.id, dist }
          }
        }
        if (closest.id != null) {
          setDetectedCityId(closest.id)
        }
      },
      () => {},
      { timeout: 5000, maximumAge: 300000 },
    )
  }, [cities, hasGeoApi])

  return { detectedCityId }
}

export default function ClientHome() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const isNewUser = !!user && !user.onboarding_completed
  const [tourDismissed, setTourDismissed] = useState(false)
  const showTour = isNewUser && !tourDismissed

  const { data: citiesData } = useGetCities()
  const citiesRawHome = citiesData?.data
  const cities = useMemo(() => citiesRawHome ?? [], [citiesRawHome])
  const { detectedCityId } = useGeoCity(cities)

  const handleTourComplete = useCallback(() => {
    setTourDismissed(true)
    useAuthStore.getState().loadProfile()
  }, [])

  return (
    <div className="rh-home">
      {showTour && (
        <OnboardingTour
          open={showTour}
          onComplete={handleTourComplete}
          region={user?.region}
        />
      )}
      <section className="rh-home__hero">
        <div className="rh-home__hero-copy">
          <Tag color="gold">Быстрое бронирование</Tag>
          <Title level={1} className="rh-home__hero-title">
            Бани для вечера вдвоем, компании и выходных за городом
          </Title>
          <Text className="rh-home__hero-description">
            Public-маршрут начинается с выбора сценария: находите подходящий формат отдыха, уточняете параметры и переходите к слоту без длинной регистрации и без лишних экранов.
          </Text>
          <Input
            size="large"
            placeholder="Поиск бань..."
            prefix={<SearchOutlined />}
            onPressEnter={(e) => {
              const val = (e.target as HTMLInputElement).value
              navigate(`/catalog${val ? `?q=${encodeURIComponent(val)}` : ''}`)
            }}
            onClick={() => navigate('/catalog')}
            readOnly
            className="rh-home__search"
          />
          <Space wrap className="rh-home__hero-actions">
            <Button size="large" type="primary" onClick={() => navigate('/catalog')}>
              Подобрать баню
            </Button>
            <Button size="large" className="rh-home__hero-secondary" onClick={() => navigate('/certificates')}>
              Подарочный сертификат
            </Button>
          </Space>
          <div className="rh-home__steps">
            {DISCOVERY_STEPS.map((step, index) => (
              <div key={step} className="rh-home__step">
                <span className="rh-home__step-index">{index + 1}</span>
                <Text className="rh-home__step-text">{step}</Text>
              </div>
            ))}
          </div>
        </div>

        <div className="rh-home__hero-side">
          <div className="rh-home__section-copy">
            <Text className="rh-home__section-eyebrow">Сценарии</Text>
            <Title level={3} className="rh-home__section-title">
              Начните с готовой подборки
            </Title>
            <Text className="rh-home__section-description">
              Сценарии привязаны к реальным фильтрам каталога, поэтому переход сразу открывает рабочую выдачу, а не декоративный экран.
            </Text>
          </div>
          <div className="rh-home__shortcut-grid">
            {PUBLIC_SHORTCUT_CARDS.map((card) => (
              <button
                key={card.key}
                type="button"
                className="rh-home__shortcut-card"
                onClick={() => navigate(card.to)}
              >
                <Tag color="cyan">{card.eyebrow}</Tag>
                <Title level={4} className="rh-home__shortcut-title">
                  {card.title}
                </Title>
                <Text className="rh-home__shortcut-description">{card.description}</Text>
              </button>
            ))}
          </div>
        </div>
      </section>

      <section className="rh-home__proof">
        <div className="rh-home__section-copy">
          <Text className="rh-home__section-eyebrow">Как устроен выбор</Text>
          <Title level={3} className="rh-home__section-title">
            Discovery без дешёвого маркетингового шума
          </Title>
          <Text className="rh-home__section-description">
            Новый клиент должен сразу понимать, что здесь можно выбрать, по каким параметрам сравнивать варианты и как быстро дойти до бронирования.
          </Text>
        </div>
        <div className="rh-home__proof-grid">
          {DISCOVERY_PROOF_POINTS.map((item) => (
            <Card key={item.key} variant="borderless" className="rh-home__proof-card">
              <Text className="rh-home__proof-eyebrow">{item.eyebrow}</Text>
              <Title level={4} className="rh-home__proof-title">
                {item.title}
              </Title>
              <Text className="rh-home__proof-description">{item.description}</Text>
            </Card>
          ))}
        </div>
      </section>

      <section className="rh-home__inventory">
        <div className="rh-home__section-copy rh-home__section-copy--compact">
          <Text className="rh-home__section-eyebrow">Подборка</Text>
          <Title level={3} className="rh-home__section-title">
            С чего обычно начинают выбор
          </Title>
        </div>
        <PopularNearby detectedCityId={detectedCityId} />
      </section>

      <PersonalRecommendations />

      <section className="rh-home__secondary">
        <PromoBanner />
        <RecentlyViewed />
      </section>
    </div>
  )
}
