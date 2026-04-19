import { useEffect, useState, useMemo, useCallback } from 'react'
import { Alert, Progress, Typography, Input, Row, Col, Card, Rate, Divider, Select, Space, Tag, Carousel, Button } from 'antd'
import { useNavigate } from 'react-router-dom'
import {
  SearchOutlined,
  EnvironmentOutlined,
  FireOutlined,
  StarOutlined,
  GiftOutlined,
  TagOutlined,
  RocketOutlined,
} from '@ant-design/icons'
import { axiosInstance } from '@/api/axios-instance'
import { useGetRecommendations, useGetPopular } from '@/api/generated/recommendations/recommendations'
import { useGetCities } from '@/api/generated/cities/cities'
import { useAuthStore } from '@/stores/auth'
import { formatPrice } from '@/lib/format'
import PublicState from '@/components/PublicState'
import RecentlyViewed from '@/components/RecentlyViewed'
import OnboardingTour from '@/components/OnboardingTour'
import { PUBLIC_SHORTCUT_CARDS } from '@/navigation/menu'

const { Title, Text } = Typography

interface CompletenessData {
  percentage: number
  items: Array<{ field: string; label: string; complete: boolean }>
}

interface PromotionBanner {
  id: string
  title: string
  description: string
  type: 'promo' | 'welcome' | 'loyalty'
  promo_code?: string
  discount_text?: string
}

function ProfileNudge() {
  const [data, setData] = useState<CompletenessData | null>(null)
  const [failedToLoad, setFailedToLoad] = useState(false)
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)

  useEffect(() => {
    if (!user) return
    axiosInstance
      .get<{ success: boolean; data: CompletenessData }>('/my/profile-completeness')
      .then((res) => {
        setData(res.data.data)
        setFailedToLoad(false)
      })
      .catch(() => {
        setFailedToLoad(true)
      })
  }, [user])

  if (!user) return null

  if (failedToLoad) {
    return (
      <PublicState
        kind="degraded"
        compact
        title="Прогресс профиля временно недоступен"
        description="Попробуйте открыть профиль позже."
      />
    )
  }

  if (!data || data.percentage === 100) return null

  const missing = data.items.filter((i) => !i.complete).map((i) => i.label)
  const strokeColor = data.percentage >= 80 ? '#52c41a' : data.percentage >= 50 ? '#faad14' : '#ff4d4f'

  return (
    <Alert
      type="info"
      showIcon
      closable
      style={{ marginBottom: 16 }}
      title={
        <span style={{ cursor: 'pointer' }} onClick={() => navigate('/client/profile')}>
          Заполните профиль ({data.percentage}%) — не хватает: {missing.join(', ')}
        </span>
      }
      description={
        <Progress percent={data.percentage} strokeColor={strokeColor} size="small" showInfo={false} />
      }
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
        promo: 'linear-gradient(135deg, #f5222d 0%, #fa541c 100%)',
        welcome: 'linear-gradient(135deg, #52c41a 0%, #13c2c2 100%)',
        loyalty: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
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
                  style={{ color: '#333', marginTop: 12, fontSize: 14, padding: '4px 12px', fontWeight: 600 }}
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
            ? 'linear-gradient(135deg, #52c41a 0%, #13c2c2 100%)'
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
  const navigate = useNavigate()
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
            <FireOutlined style={{ marginRight: 8, color: '#ff4d4f' }} />
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
              <Card
                hoverable
                onClick={() => navigate(`/bathhouses/${item.slug ?? item.id}`)}
              >
                <Title level={5} style={{ margin: 0, marginBottom: 8 }}>{item.name}</Title>
                {item.address && (
                  <Text type="secondary" style={{ fontSize: 13, display: 'block', marginBottom: 4 }}>
                    <EnvironmentOutlined style={{ marginRight: 4 }} />
                    {item.address}
                  </Text>
                )}
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                  <Rate disabled allowHalf value={item.rating ?? 0} style={{ fontSize: 14 }} />
                  <Text type="secondary" style={{ fontSize: 13 }}>
                    {item.rating?.toFixed(1)} ({item.review_count ?? 0})
                  </Text>
                </div>
                {item.price_per_hour != null && (
                  <Text strong>{formatPrice(item.price_per_hour)}/ч</Text>
                )}
              </Card>
            </Col>
          ))}
        </Row>
      )}
    </div>
  )
}

function PersonalRecommendations() {
  const navigate = useNavigate()
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
        <StarOutlined style={{ marginRight: 8, color: '#faad14' }} />
        Рекомендации для вас
      </Title>
      <Row gutter={[16, 16]}>
        {recs.slice(0, 6).map((item) => (
          <Col key={item.id} xs={24} sm={12} md={8}>
            <Card
              hoverable
              onClick={() => navigate(`/bathhouses/${item.slug ?? item.id}`)}
            >
              <Title level={5} style={{ margin: 0, marginBottom: 8 }}>{item.name}</Title>
              {item.address && (
                <Text type="secondary" style={{ fontSize: 13, display: 'block', marginBottom: 4 }}>
                  <EnvironmentOutlined style={{ marginRight: 4 }} />
                  {item.address}
                </Text>
              )}
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                <Rate disabled allowHalf value={item.rating ?? 0} style={{ fontSize: 14 }} />
                <Text type="secondary" style={{ fontSize: 13 }}>
                  {item.rating?.toFixed(1)} ({item.review_count ?? 0})
                </Text>
              </div>
              {item.price_per_hour != null && (
                <Text strong>{formatPrice(item.price_per_hour)}/ч</Text>
              )}
            </Card>
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
    <>
      {showTour && (
        <OnboardingTour
          open={showTour}
          onComplete={handleTourComplete}
          region={user?.region}
        />
      )}

      <ProfileNudge />

      <PromoBanner />

      <Card
        style={{
          borderRadius: 28,
          border: 'none',
          background: 'linear-gradient(135deg, #14323b 0%, #7a4a2c 100%)',
          marginBottom: 24,
        }}
      >
        <Row gutter={[24, 24]} align="middle">
          <Col xs={24} lg={14}>
            <Tag color="gold">Быстрое бронирование</Tag>
            <Title level={1} style={{ color: '#fff', marginTop: 16, marginBottom: 12 }}>
              Бани для вечера вдвоем, компании и выходных за городом
            </Title>
            <Text style={{ color: 'rgba(255,255,255,0.78)', fontSize: 16 }}>
              Выбирайте по сценарию отдыха, смотрите доступные слоты и подтверждайте телефон только в финальном шаге. Без длинной регистрации и без лишних экранов.
            </Text>
            <Space wrap style={{ display: 'flex', marginTop: 24 }}>
              <Button size="large" type="primary" onClick={() => navigate('/catalog')}>
                Подобрать баню
              </Button>
              <Button size="large" ghost onClick={() => navigate('/certificates')}>
                Подарочный сертификат
              </Button>
            </Space>
          </Col>
          <Col xs={24} lg={10}>
            <Row gutter={[12, 12]}>
              {PUBLIC_SHORTCUT_CARDS.map((card) => (
                <Col key={card.key} xs={24} sm={12}>
                  <Card
                    hoverable
                    onClick={() => navigate(card.to)}
                    style={{ borderRadius: 20, background: 'rgba(255,255,255,0.08)', border: '1px solid rgba(255,255,255,0.1)' }}
                  >
                    <Tag color="cyan">{card.eyebrow}</Tag>
                    <Title level={4} style={{ color: '#fff', marginTop: 12, marginBottom: 8 }}>
                      {card.title}
                    </Title>
                    <Text style={{ color: 'rgba(255,255,255,0.76)' }}>{card.description}</Text>
                  </Card>
                </Col>
              ))}
            </Row>
          </Col>
        </Row>
      </Card>

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
        style={{ marginBottom: 24, cursor: 'pointer' }}
      />

      <RecentlyViewed />

      <PersonalRecommendations />

      <Divider />

      <PopularNearby detectedCityId={detectedCityId} />
    </>
  )
}
