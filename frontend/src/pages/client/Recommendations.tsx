import { useState } from 'react'
import { Typography, Row, Col, Pagination, Spin, Empty, Card, Rate, Select, Divider } from 'antd'
import { EnvironmentOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useGetRecommendations, useGetPopular } from '@/api/generated/recommendations/recommendations'
import { useGetCities } from '@/api/generated/cities/cities'
import { formatPrice } from '@/lib/format'
import type { InternalHandlerRecommendationResponse } from '@/api/generated/model'

const { Title, Text } = Typography

const PAGE_SIZE = 12

function RecommendationCard({ item }: { item: InternalHandlerRecommendationResponse }) {
  const navigate = useNavigate()

  return (
    <Card
      hoverable
      onClick={() => navigate(`/client/bathhouse/${item.id}`)}
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
  )
}

export default function Recommendations() {
  const [page, setPage] = useState(1)
  const [popularCityId, setPopularCityId] = useState<number | undefined>()

  const { data: recommendationsData, isLoading: loadingRecs } = useGetRecommendations(
    { page, page_size: PAGE_SIZE },
  )

  const { data: citiesData } = useGetCities()
  const cities = citiesData?.data ?? []

  const effectiveCityId = popularCityId ?? cities[0]?.id
  const { data: popularData, isLoading: loadingPopular } = useGetPopular(
    { city_id: effectiveCityId!, limit: 8 },
    { query: { enabled: !!effectiveCityId } },
  )

  const recommendations = recommendationsData?.data ?? []
  const recsMeta = recommendationsData?.meta
  const popular = popularData?.data ?? []

  return (
    <div>
      <Title level={3}>Рекомендации для вас</Title>

      <Spin spinning={loadingRecs}>
        {recommendations.length > 0 ? (
          <>
            <Row gutter={[16, 16]}>
              {recommendations.map((item) => (
                <Col key={item.id} xs={24} sm={12} md={8} lg={6}>
                  <RecommendationCard item={item} />
                </Col>
              ))}
            </Row>

            {recsMeta && recsMeta.total_pages && recsMeta.total_pages > 1 && (
              <div style={{ textAlign: 'center', marginTop: 24 }}>
                <Pagination
                  current={page}
                  total={recsMeta.total_count}
                  pageSize={PAGE_SIZE}
                  onChange={setPage}
                  showSizeChanger={false}
                />
              </div>
            )}
          </>
        ) : (
          !loadingRecs && (
            <Empty
              description="Пока нет персональных рекомендаций. Настройте предпочтения для лучших результатов."
              style={{ marginTop: 24 }}
            />
          )
        )}
      </Spin>

      <Divider />

      <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>Популярные бани</Title>
        {cities.length > 0 && (
          <Select
            value={effectiveCityId}
            onChange={setPopularCityId}
            style={{ width: 200 }}
            placeholder="Выберите город"
          >
            {cities.map((city) => (
              <Select.Option key={city.id} value={city.id}>
                {city.name}
              </Select.Option>
            ))}
          </Select>
        )}
      </div>

      <Spin spinning={loadingPopular}>
        {popular.length > 0 ? (
          <Row gutter={[16, 16]}>
            {popular.map((item) => (
              <Col key={item.id} xs={24} sm={12} md={8} lg={6}>
                <RecommendationCard item={item} />
              </Col>
            ))}
          </Row>
        ) : (
          !loadingPopular && effectiveCityId && (
            <Empty description="Нет популярных бань в этом городе" />
          )
        )}
      </Spin>
    </div>
  )
}
