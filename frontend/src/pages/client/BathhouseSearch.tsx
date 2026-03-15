import { useState } from 'react'
import {
  Typography,
  Input,
  Select,
  Slider,
  Checkbox,
  Row,
  Col,
  Pagination,
  Spin,
  Empty,
  Card,
  Space,
  Button,
  InputNumber,
  Collapse,
} from 'antd'
import {
  SearchOutlined,
  EnvironmentOutlined,
  FilterOutlined,
  SortAscendingOutlined,
} from '@ant-design/icons'
import { useGetBathhouses } from '@/api/generated/bathhouses/bathhouses'
import { useGetCities } from '@/api/generated/cities/cities'
import type { GetBathhousesParams } from '@/api/generated/model'
import BathhouseCard from '@/components/BathhouseCard'
import { formatPrice } from '@/lib/format'

const { Title } = Typography

const SORT_OPTIONS = [
  { value: 'rating_desc', label: 'По рейтингу' },
  { value: 'price_asc', label: 'Сначала дешёвые' },
  { value: 'price_desc', label: 'Сначала дорогие' },
  { value: 'created_at_desc', label: 'Новые' },
  { value: 'distance_asc', label: 'Ближайшие' },
]

function parseSortOption(value: string): { sort_by: string; sort_order: string } {
  const lastUnderscore = value.lastIndexOf('_')
  return {
    sort_by: value.slice(0, lastUnderscore),
    sort_order: value.slice(lastUnderscore + 1),
  }
}

export default function BathhouseSearch() {
  const [page, setPage] = useState(1)
  const [pageSize] = useState(12)
  const [search, setSearch] = useState('')
  const [sortValue, setSortValue] = useState('rating_desc')
  const [filters, setFilters] = useState<Partial<GetBathhousesParams>>({})
  const [geoEnabled, setGeoEnabled] = useState(false)
  const [geoCoords, setGeoCoords] = useState<{ lat: number; lng: number } | null>(null)

  const { data: citiesData } = useGetCities()
  const cities = citiesData?.data ?? []

  const { sort_by, sort_order } = parseSortOption(sortValue)

  const queryParams: GetBathhousesParams = {
    page,
    page_size: pageSize,
    sort_by,
    sort_order,
    q: search || undefined,
    ...filters,
    ...(geoEnabled && geoCoords ? { lat: geoCoords.lat, lng: geoCoords.lng, radius_km: filters.radius_km ?? 10 } : {}),
  }

  const { data, isLoading } = useGetBathhouses(queryParams)
  const bathhouses = data?.data ?? []
  const meta = data?.meta

  const handleGeoSearch = () => {
    if (!navigator.geolocation) return
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setGeoCoords({ lat: pos.coords.latitude, lng: pos.coords.longitude })
        setGeoEnabled(true)
        setPage(1)
      },
      () => {/* geolocation denied - ignore */},
    )
  }

  const updateFilter = (key: keyof GetBathhousesParams, value: unknown) => {
    setFilters((prev) => ({ ...prev, [key]: value || undefined }))
    setPage(1)
  }

  return (
    <div>
      <Title level={3}>Поиск бань</Title>

      <Space direction="vertical" size="middle" style={{ width: '100%', marginBottom: 24 }}>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={12} md={8}>
            <Input
              placeholder="Поиск по названию..."
              prefix={<SearchOutlined />}
              value={search}
              onChange={(e) => { setSearch(e.target.value); setPage(1) }}
              allowClear
            />
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Select
              placeholder="Город"
              style={{ width: '100%' }}
              allowClear
              value={filters.city_id}
              onChange={(v) => updateFilter('city_id', v)}
              options={cities.map((c) => ({ value: c.id, label: c.name }))}
            />
          </Col>
          <Col xs={12} md={5}>
            <Select
              style={{ width: '100%' }}
              value={sortValue}
              onChange={setSortValue}
              options={SORT_OPTIONS}
              suffixIcon={<SortAscendingOutlined />}
            />
          </Col>
          <Col xs={12} md={5}>
            <Button
              icon={<EnvironmentOutlined />}
              onClick={handleGeoSearch}
              type={geoEnabled ? 'primary' : 'default'}
            >
              {geoEnabled ? 'Рядом со мной' : 'Найти рядом'}
            </Button>
          </Col>
        </Row>

        <Collapse
          ghost
          items={[
            {
              key: 'filters',
              label: (
                <Space>
                  <FilterOutlined />
                  Фильтры
                </Space>
              ),
              children: (
                <Card size="small">
                  <Row gutter={[16, 16]}>
                    <Col xs={24} sm={12}>
                      <Typography.Text type="secondary">Цена за час</Typography.Text>
                      <Slider
                        range
                        min={0}
                        max={1000000}
                        step={10000}
                        value={[filters.price_min ?? 0, filters.price_max ?? 1000000]}
                        onChange={(values: number[]) => {
                          const [min, max] = values
                          setFilters((prev) => ({
                            ...prev,
                            price_min: min || undefined,
                            price_max: max !== undefined && max < 1000000 ? max : undefined,
                          }))
                          setPage(1)
                        }}
                        tooltip={{
                          formatter: (v) => (v != null ? formatPrice(v) : ''),
                        }}
                      />
                    </Col>
                    <Col xs={24} sm={12}>
                      <Typography.Text type="secondary">Мин. рейтинг</Typography.Text>
                      <Slider
                        min={0}
                        max={5}
                        step={0.5}
                        value={filters.min_rating ?? 0}
                        onChange={(v) => updateFilter('min_rating', v || undefined)}
                        tooltip={{ formatter: (v) => v?.toString() ?? '' }}
                      />
                    </Col>
                    <Col xs={24} sm={12}>
                      <Typography.Text type="secondary">Гости</Typography.Text>
                      <InputNumber
                        min={1}
                        max={50}
                        placeholder="Кол-во гостей"
                        style={{ width: '100%' }}
                        value={filters.guest_count}
                        onChange={(v) => updateFilter('guest_count', v)}
                      />
                    </Col>
                    {geoEnabled && (
                      <Col xs={24} sm={12}>
                        <Typography.Text type="secondary">Радиус (км)</Typography.Text>
                        <InputNumber
                          min={1}
                          max={100}
                          value={filters.radius_km ?? 10}
                          onChange={(v) => updateFilter('radius_km', v)}
                          style={{ width: '100%' }}
                        />
                      </Col>
                    )}
                    <Col xs={24}>
                      <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>Удобства</Typography.Text>
                      <Space wrap>
                        <Checkbox checked={filters.has_sauna} onChange={(e) => updateFilter('has_sauna', e.target.checked || undefined)}>Сауна</Checkbox>
                        <Checkbox checked={filters.has_steam_room} onChange={(e) => updateFilter('has_steam_room', e.target.checked || undefined)}>Парная</Checkbox>
                        <Checkbox checked={filters.has_pool} onChange={(e) => updateFilter('has_pool', e.target.checked || undefined)}>Бассейн</Checkbox>
                        <Checkbox checked={filters.has_hot_tub} onChange={(e) => updateFilter('has_hot_tub', e.target.checked || undefined)}>Джакузи</Checkbox>
                        <Checkbox checked={filters.has_bbq} onChange={(e) => updateFilter('has_bbq', e.target.checked || undefined)}>Мангал</Checkbox>
                        <Checkbox checked={filters.has_karaoke} onChange={(e) => updateFilter('has_karaoke', e.target.checked || undefined)}>Караоке</Checkbox>
                      </Space>
                    </Col>
                  </Row>
                </Card>
              ),
            },
          ]}
        />
      </Space>

      <Spin spinning={isLoading}>
        {bathhouses.length === 0 && !isLoading ? (
          <Empty description="Бани не найдены" />
        ) : (
          <>
            <Row gutter={[16, 16]}>
              {bathhouses.map((b) => (
                <Col key={b.id} xs={24} sm={12} md={8} lg={6}>
                  <BathhouseCard bathhouse={b} />
                </Col>
              ))}
            </Row>
            {meta && meta.total_pages && meta.total_pages > 1 && (
              <div style={{ textAlign: 'center', marginTop: 24 }}>
                <Pagination
                  current={page}
                  pageSize={pageSize}
                  total={meta.total_count}
                  onChange={(p) => setPage(p)}
                  showSizeChanger={false}
                />
              </div>
            )}
          </>
        )}
      </Spin>
    </div>
  )
}
