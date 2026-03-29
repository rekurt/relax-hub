import { useState, useCallback, useEffect, useRef } from 'react'
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
  Badge,
  Segmented,
  App,
  Tag,
} from 'antd'
import {
  SearchOutlined,
  EnvironmentOutlined,
  FilterOutlined,
  SortAscendingOutlined,
  SwapOutlined,
  UnorderedListOutlined,
  EnvironmentFilled,
  SplitCellsOutlined,
  CarOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useGetBathhouses } from '@/api/generated/bathhouses/bathhouses'
import { useGetCities } from '@/api/generated/cities/cities'
import type { GetBathhousesParams } from '@/api/generated/model'
import BathhouseCard from '@/components/BathhouseCard'
import BathhouseMap from '@/components/BathhouseMap'
import { formatPrice } from '@/lib/format'
import { axiosInstance } from '@/api/axios-instance'

const { Title } = Typography

const SORT_OPTIONS = [
  { value: 'rating_desc', label: 'По рейтингу' },
  { value: 'price_asc', label: 'Сначала дешёвые' },
  { value: 'price_desc', label: 'Сначала дорогие' },
  { value: 'created_at_desc', label: 'Новые' },
  { value: 'distance_asc', label: 'Ближайшие' },
]

type ViewMode = 'list' | 'map' | 'split'
type GeoMode = 'radius' | 'travel_time'

const VIEW_MODE_OPTIONS = [
  { value: 'list', icon: <UnorderedListOutlined />, label: 'Список' },
  { value: 'split', icon: <SplitCellsOutlined />, label: 'Сплит' },
  { value: 'map', icon: <EnvironmentFilled />, label: 'Карта' },
]

const TRAVEL_TIME_OPTIONS = [
  { value: 15, label: '15 мин' },
  { value: 30, label: '30 мин' },
  { value: 45, label: '45 мин' },
  { value: 60, label: '60 мин' },
]

const TRAVEL_MODE_OPTIONS = [
  { value: 'car', label: 'На машине', icon: <CarOutlined /> },
  { value: 'transit', label: 'Пешком/транспорт', icon: <ClockCircleOutlined /> },
]

interface IsochroneData {
  coordinates: number[][]
  wkt: string
}

function parseSortOption(value: string): { sort_by: string; sort_order: string } {
  const lastUnderscore = value.lastIndexOf('_')
  return {
    sort_by: value.slice(0, lastUnderscore),
    sort_order: value.slice(lastUnderscore + 1),
  }
}

const MAX_COMPARE = 3

export default function BathhouseSearch() {
  const [page, setPage] = useState(1)
  const [pageSize] = useState(12)
  const [search, setSearch] = useState('')
  const [sortValue, setSortValue] = useState('rating_desc')
  const [filters, setFilters] = useState<Partial<GetBathhousesParams>>({})
  const [geoEnabled, setGeoEnabled] = useState(false)
  const [geoCoords, setGeoCoords] = useState<{ lat: number; lng: number } | null>(null)
  const [viewMode, setViewMode] = useState<ViewMode>('list')
  const [compareIds, setCompareIds] = useState<string[]>([])
  const [highlightedId, setHighlightedId] = useState<string | null>(null)
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const debounceTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    debounceTimer.current = setTimeout(() => {
      setDebouncedSearch(search)
    }, 300)
    return () => {
      if (debounceTimer.current) clearTimeout(debounceTimer.current)
    }
  }, [search])

  // Isochrone state
  const [geoMode, setGeoMode] = useState<GeoMode>('radius')
  const [travelMode, setTravelMode] = useState<string>('car')
  const [travelMinutes, setTravelMinutes] = useState<number>(15)
  const [isochroneData, setIsochroneData] = useState<IsochroneData | null>(null)
  const [isochroneLoading, setIsochroneLoading] = useState(false)

  const navigate = useNavigate()
  const { message } = App.useApp()
  const { data: citiesData } = useGetCities()
  const cities = citiesData?.data ?? []

  const { sort_by, sort_order } = parseSortOption(sortValue)

  const queryParams: GetBathhousesParams = {
    page,
    page_size: pageSize,
    sort_by,
    sort_order,
    q: debouncedSearch || undefined,
    ...filters,
    ...(geoEnabled && geoCoords && geoMode === 'radius'
      ? { lat: geoCoords.lat, lng: geoCoords.lng, radius_km: filters.radius_km ?? 10 }
      : {}),
    ...(geoEnabled && geoCoords && geoMode === 'travel_time' && isochroneData
      ? { lat: geoCoords.lat, lng: geoCoords.lng, isochrone_wkt: isochroneData.wkt }
      : {}),
  }

  const { data, isLoading } = useGetBathhouses(queryParams as GetBathhousesParams)
  const bathhouses = data?.data ?? []
  const meta = data?.meta

  const fetchIsochrone = useCallback(async (lat: number, lng: number, mode: string, minutes: number) => {
    setIsochroneLoading(true)
    try {
      const resp = await axiosInstance.get('/isochrone', {
        params: { lat, lon: lng, mode, minutes },
      })
      const result = resp.data?.data ?? resp.data
      if (result?.coordinates && result?.wkt) {
        setIsochroneData({ coordinates: result.coordinates, wkt: result.wkt })
      }
    } catch {
      message.error('Не удалось построить зону доступности')
      setIsochroneData(null)
    } finally {
      setIsochroneLoading(false)
    }
  }, [message])

  const handleGeoSearch = () => {
    if (!navigator.geolocation) return
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        const coords = { lat: pos.coords.latitude, lng: pos.coords.longitude }
        setGeoCoords(coords)
        setGeoEnabled(true)
        setPage(1)
        if (geoMode === 'travel_time') {
          fetchIsochrone(coords.lat, coords.lng, travelMode, travelMinutes)
        }
      },
      () => {/* geolocation denied - ignore */},
    )
  }

  const handleGeoModeChange = (mode: GeoMode) => {
    setGeoMode(mode)
    if (mode === 'travel_time' && geoCoords) {
      fetchIsochrone(geoCoords.lat, geoCoords.lng, travelMode, travelMinutes)
    } else {
      setIsochroneData(null)
    }
    setPage(1)
  }

  const handleTravelModeChange = (mode: string) => {
    setTravelMode(mode)
    if (geoCoords && geoMode === 'travel_time') {
      fetchIsochrone(geoCoords.lat, geoCoords.lng, mode, travelMinutes)
    }
    setPage(1)
  }

  const handleTravelMinutesChange = (minutes: number) => {
    setTravelMinutes(minutes)
    if (geoCoords && geoMode === 'travel_time') {
      fetchIsochrone(geoCoords.lat, geoCoords.lng, travelMode, minutes)
    }
    setPage(1)
  }

  const updateFilter = (key: keyof GetBathhousesParams, value: unknown) => {
    setFilters((prev) => ({ ...prev, [key]: value || undefined }))
    setPage(1)
  }

  const handleCompareToggle = (id: string) => {
    setCompareIds((prev) => {
      if (prev.includes(id)) return prev.filter((x) => x !== id)
      if (prev.length >= MAX_COMPARE) {
        message.warning(`Можно сравнить не более ${MAX_COMPARE} бань`)
        return prev
      }
      return [...prev, id]
    })
  }

  const handleGoCompare = () => {
    if (compareIds.length < 2) {
      message.warning('Выберите минимум 2 бани для сравнения')
      return
    }
    navigate(`/client/comparison?ids=${compareIds.join(',')}`)
  }

  const handleBoundsChange = (bounds: { north: number; south: number; east: number; west: number }) => {
    if (geoMode === 'travel_time') return // Don't override isochrone with bounds
    setFilters((prev) => ({
      ...prev,
      lat: (bounds.north + bounds.south) / 2,
      lng: (bounds.east + bounds.west) / 2,
      radius_km: Math.max(
        1,
        Math.round(
          haversineDistance(
            bounds.south,
            bounds.west,
            bounds.north,
            bounds.east,
          ) / 2,
        ),
      ),
    }))
    setGeoEnabled(true)
    setPage(1)
  }

  const handleMarkerClick = (id: string) => {
    const found = bathhouses.find((b) => b.id === id)
    navigate(`/client/bathhouse/${found?.slug ?? id}`)
  }

  // Polygon for map display (convert [lng, lat] -> [lat, lng] for Yandex Maps)
  const isochronePolygon: number[][] | null = isochroneData?.coordinates
    ? isochroneData.coordinates.map((coord: number[]) => [coord[1] ?? 0, coord[0] ?? 0] as [number, number])
    : null

  const listContent = (
    <Spin spinning={isLoading || isochroneLoading}>
      {bathhouses.length === 0 && !isLoading ? (
        <Empty description="Бани не найдены" />
      ) : (
        <>
          <Row gutter={[16, 16]}>
            {bathhouses.map((b) => (
              <Col
                key={b.id}
                xs={24}
                sm={viewMode === 'split' ? 24 : 12}
                md={viewMode === 'split' ? 24 : 8}
                lg={viewMode === 'split' ? 12 : 6}
                onMouseEnter={() => setHighlightedId(b.id ?? null)}
                onMouseLeave={() => setHighlightedId(null)}
              >
                <BathhouseCard
                  bathhouse={b}
                  showCompare
                  isCompareSelected={compareIds.includes(b.id ?? '')}
                  onCompareToggle={handleCompareToggle}
                />
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
  )

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>Поиск бань</Title>
        <Segmented
          options={VIEW_MODE_OPTIONS}
          value={viewMode}
          onChange={(v) => setViewMode(v as ViewMode)}
        />
      </div>

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
                            price_min: min !== 0 ? min : undefined,
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
                        <Typography.Text type="secondary">Способ поиска</Typography.Text>
                        <Segmented
                          block
                          options={[
                            { value: 'radius', label: 'По радиусу' },
                            { value: 'travel_time', label: 'По времени пути' },
                          ]}
                          value={geoMode}
                          onChange={(v) => handleGeoModeChange(v as GeoMode)}
                          style={{ marginTop: 4 }}
                        />
                      </Col>
                    )}
                    {geoEnabled && geoMode === 'radius' && (
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
                    {geoEnabled && geoMode === 'travel_time' && (
                      <>
                        <Col xs={12} sm={6}>
                          <Typography.Text type="secondary">Способ</Typography.Text>
                          <Select
                            style={{ width: '100%' }}
                            value={travelMode}
                            onChange={handleTravelModeChange}
                            options={TRAVEL_MODE_OPTIONS}
                          />
                        </Col>
                        <Col xs={12} sm={6}>
                          <Typography.Text type="secondary">Время в пути</Typography.Text>
                          <Select
                            style={{ width: '100%' }}
                            value={travelMinutes}
                            onChange={handleTravelMinutesChange}
                            options={TRAVEL_TIME_OPTIONS}
                          />
                        </Col>
                      </>
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

      {/* Comparison bar */}
      {compareIds.length > 0 && (
        <div
          style={{
            position: 'sticky',
            top: 64,
            zIndex: 9,
            background: '#f0f5ff',
            border: '1px solid #adc6ff',
            borderRadius: 8,
            padding: '8px 16px',
            marginBottom: 16,
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}
        >
          <Space>
            <SwapOutlined />
            <span>
              Выбрано для сравнения: <Badge count={compareIds.length} style={{ backgroundColor: '#722ed1' }} />
            </span>
          </Space>
          <Space>
            <Button size="small" onClick={() => setCompareIds([])}>
              Сбросить
            </Button>
            <Button
              type="primary"
              size="small"
              disabled={compareIds.length < 2}
              onClick={handleGoCompare}
            >
              Сравнить
            </Button>
          </Space>
        </div>
      )}

      {/* Result counter */}
      {meta?.total_count != null && (
        <div style={{ marginBottom: 12 }} data-testid="result-counter">
          <Tag color="blue" style={{ fontSize: 14, padding: '2px 10px' }}>
            {isLoading ? '...' : `Найдено: ${pluralizeBathhouse(meta.total_count)}`}
          </Tag>
        </div>
      )}

      {/* Content area */}
      {viewMode === 'list' && listContent}

      {viewMode === 'map' && (
        <BathhouseMap
          bathhouses={bathhouses}
          highlightedId={highlightedId}
          onBoundsChange={handleBoundsChange}
          onMarkerClick={handleMarkerClick}
          onMarkerHover={setHighlightedId}
          center={geoCoords ?? undefined}
          isochronePolygon={isochronePolygon}
          style={{ height: 600, borderRadius: 8, overflow: 'hidden' }}
        />
      )}

      {viewMode === 'split' && (
        <Row gutter={16}>
          <Col xs={24} md={12}>
            <div style={{ maxHeight: 700, overflow: 'auto' }}>{listContent}</div>
          </Col>
          <Col xs={24} md={12}>
            <BathhouseMap
              bathhouses={bathhouses}
              highlightedId={highlightedId}
              onBoundsChange={handleBoundsChange}
              onMarkerClick={handleMarkerClick}
              onMarkerHover={setHighlightedId}
              center={geoCoords ?? undefined}
              isochronePolygon={isochronePolygon}
              style={{ height: 700, borderRadius: 8, overflow: 'hidden' }}
            />
          </Col>
        </Row>
      )}
    </div>
  )
}

function pluralizeBathhouse(count: number): string {
  const mod10 = count % 10
  const mod100 = count % 100
  if (mod100 >= 11 && mod100 <= 19) return `${count} бань`
  if (mod10 === 1) return `${count} баня`
  if (mod10 >= 2 && mod10 <= 4) return `${count} бани`
  return `${count} бань`
}

function haversineDistance(lat1: number, lon1: number, lat2: number, lon2: number): number {
  const R = 6371
  const dLat = ((lat2 - lat1) * Math.PI) / 180
  const dLon = ((lon2 - lon1) * Math.PI) / 180
  const a =
    Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos((lat1 * Math.PI) / 180) *
      Math.cos((lat2 * Math.PI) / 180) *
      Math.sin(dLon / 2) *
      Math.sin(dLon / 2)
  return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
}
