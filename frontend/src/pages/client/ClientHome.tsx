import { useEffect, useState } from 'react'
import { Alert, Progress, Typography, Input, Row, Col, Card, Rate, Divider } from 'antd'
import { useNavigate } from 'react-router-dom'
import { SearchOutlined, EnvironmentOutlined, FireOutlined, StarOutlined } from '@ant-design/icons'
import { axiosInstance } from '@/api/axios-instance'
import { useGetRecommendations, useGetPopular } from '@/api/generated/recommendations/recommendations'
import { useGetCities } from '@/api/generated/cities/cities'
import { useAuthStore } from '@/stores/auth'
import { formatPrice } from '@/lib/format'
import RecentlyViewed from '@/components/RecentlyViewed'

const { Title, Text } = Typography

interface CompletenessData {
  percentage: number
  items: Array<{ field: string; label: string; complete: boolean }>
}

function ProfileNudge() {
  const [data, setData] = useState<CompletenessData | null>(null)
  const navigate = useNavigate()

  useEffect(() => {
    axiosInstance
      .get<{ success: boolean; data: CompletenessData }>('/my/profile-completeness')
      .then((res) => setData(res.data.data))
      .catch(() => {})
  }, [])

  if (!data || data.percentage === 100) return null

  const missing = data.items.filter((i) => !i.complete).map((i) => i.label)
  const strokeColor = data.percentage >= 80 ? '#52c41a' : data.percentage >= 50 ? '#faad14' : '#ff4d4f'

  return (
    <Alert
      type="info"
      showIcon
      closable
      style={{ marginBottom: 16 }}
      message={
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
  return (
    <Card
      style={{
        marginBottom: 24,
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        border: 'none',
        borderRadius: 12,
      }}
    >
      <div style={{ color: '#fff' }}>
        <Title level={3} style={{ color: '#fff', margin: 0 }}>
          Добро пожаловать в Bani!
        </Title>
        <Text style={{ color: 'rgba(255,255,255,0.85)', fontSize: 16 }}>
          Найдите идеальную баню рядом с вами. Бонус 500 ₽ новым пользователям!
        </Text>
      </div>
    </Card>
  )
}

function PopularNearby() {
  const navigate = useNavigate()
  const { data: citiesData } = useGetCities()
  const cities = citiesData?.data ?? []
  const firstCityId = cities.length > 0 ? cities[0].id : undefined

  const { data: popularData, isLoading } = useGetPopular(
    { city_id: firstCityId ?? 0, limit: 6 },
    { query: { enabled: !!firstCityId } },
  )
  const popular = popularData?.data ?? []

  if (isLoading || popular.length === 0) return null

  return (
    <div style={{ marginBottom: 24 }}>
      <Title level={4}>
        <FireOutlined style={{ marginRight: 8, color: '#ff4d4f' }} />
        Популярные рядом
      </Title>
      <Row gutter={[16, 16]}>
        {popular.slice(0, 6).map((item) => (
          <Col key={item.id} xs={24} sm={12} md={8}>
            <Card
              hoverable
              onClick={() => navigate(`/client/bathhouse/${item.slug ?? item.id}`)}
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

function PersonalRecommendations() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)

  const { data: recsData, isLoading } = useGetRecommendations(
    { page: 1, page_size: 6 },
    { query: { enabled: !!user } },
  )
  const recs = recsData?.data ?? []

  if (!user || isLoading || recs.length === 0) return null

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
              onClick={() => navigate(`/client/bathhouse/${item.slug ?? item.id}`)}
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

export default function ClientHome() {
  const navigate = useNavigate()

  return (
    <>
      <ProfileNudge />

      <PromoBanner />

      <Input
        size="large"
        placeholder="Поиск бань..."
        prefix={<SearchOutlined />}
        onPressEnter={(e) => {
          const val = (e.target as HTMLInputElement).value
          navigate(`/client/search${val ? `?q=${encodeURIComponent(val)}` : ''}`)
        }}
        onClick={() => navigate('/client/search')}
        readOnly
        style={{ marginBottom: 24, cursor: 'pointer' }}
      />

      <RecentlyViewed />

      <PersonalRecommendations />

      <Divider />

      <PopularNearby />
    </>
  )
}
