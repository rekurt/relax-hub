import { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import { Card, Segmented, Spin, Typography, Alert, Slider, Row, Col, Statistic, Radio } from 'antd'
import { HeatMapOutlined, EnvironmentOutlined, ShopOutlined, SearchOutlined } from '@ant-design/icons'
import { useQuery } from '@tanstack/react-query'
import { axiosInstance } from '@/api/axios-instance'

const { Title } = Typography

const YMAPS_SRC = 'https://api-maps.yandex.ru/2.1/?apikey=&lang=ru_RU'
let ymapsLoadPromise: Promise<void> | null = null

function loadYmaps(): Promise<void> {
  if (window.ymaps) return Promise.resolve()
  if (ymapsLoadPromise) return ymapsLoadPromise
  ymapsLoadPromise = new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src = YMAPS_SRC
    script.async = true
    script.onload = () => {
      if (window.ymaps) {
        window.ymaps.ready(() => resolve())
      } else {
        reject(new Error('ymaps not available'))
      }
    }
    script.onerror = () => reject(new Error('Failed to load Yandex Maps'))
    document.head.appendChild(script)
  })
  return ymapsLoadPromise
}

const PERIOD_OPTIONS = [
  { label: 'День', value: '1d' },
  { label: 'Неделя', value: '7d' },
  { label: 'Месяц', value: '30d' },
  { label: '3 месяца', value: '90d' },
]

const CELL_SIZE_MARKS: Record<number, string> = {
  0.005: '~500м',
  0.01: '~1км',
  0.05: '~5км',
  0.1: '~10км',
}

interface HeatmapCell {
  latitude: number
  longitude: number
  listing_count: number
  booking_count: number
  search_count: number
}

interface HeatmapData {
  period: string
  cell_size: number
  cells: HeatmapCell[]
}

type HeatmapLayer = 'demand' | 'supply'

function fetchHeatmap(period: string, cellSize: number): Promise<{ data: HeatmapData }> {
  return axiosInstance.get('/admin/analytics/heatmap', {
    params: { period, cell_size: cellSize },
  }).then((res) => res.data)
}

function getDemandColor(value: number, max: number): string {
  if (max === 0) return 'rgba(59, 130, 246, 0.1)'
  const ratio = Math.min(value / max, 1)
  if (ratio < 0.2) return 'rgba(59, 130, 246, 0.15)'
  if (ratio < 0.4) return 'rgba(59, 130, 246, 0.35)'
  if (ratio < 0.6) return 'rgba(245, 158, 11, 0.5)'
  if (ratio < 0.8) return 'rgba(249, 115, 22, 0.65)'
  return 'rgba(239, 68, 68, 0.8)'
}

function getSupplyColor(value: number, max: number): string {
  if (max === 0) return 'rgba(34, 197, 94, 0.1)'
  const ratio = Math.min(value / max, 1)
  if (ratio < 0.2) return 'rgba(34, 197, 94, 0.15)'
  if (ratio < 0.4) return 'rgba(34, 197, 94, 0.35)'
  if (ratio < 0.6) return 'rgba(22, 163, 74, 0.5)'
  if (ratio < 0.8) return 'rgba(21, 128, 61, 0.65)'
  return 'rgba(20, 83, 45, 0.8)'
}

const DEFAULT_CENTER: [number, number] = [55.751244, 37.618423]
const DEFAULT_ZOOM = 10

export default function GeoHeatmap() {
  const [period, setPeriod] = useState('30d')
  const [cellSize, setCellSize] = useState(0.01)
  const [layer, setLayer] = useState<HeatmapLayer>('demand')
  const [mapReady, setMapReady] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<ReturnType<typeof Object> | null>(null)
  const rectanglesRef = useRef<unknown[]>([])

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin-heatmap', period, cellSize],
    queryFn: () => fetchHeatmap(period, cellSize),
  })

  const cells = useMemo(() => data?.data?.cells ?? [], [data?.data?.cells])
  const totalListings = cells.reduce((sum, c) => sum + c.listing_count, 0)
  const totalBookings = cells.reduce((sum, c) => sum + c.booking_count, 0)
  const totalSearches = cells.reduce((sum, c) => sum + c.search_count, 0)

  // Initialize Yandex Map
  useEffect(() => {
    let destroyed = false
    loadYmaps()
      .then(() => {
        if (destroyed || !containerRef.current || !window.ymaps) return
        const map = new window.ymaps.Map(containerRef.current, {
          center: DEFAULT_CENTER,
          zoom: DEFAULT_ZOOM,
          controls: ['zoomControl', 'typeSelector'],
        })
        mapRef.current = map
        setMapReady(true)
      })
      .catch(() => { /* map load failed */ })

    return () => {
      destroyed = true
      if (mapRef.current) {
        (mapRef.current as { destroy(): void }).destroy()
        mapRef.current = null
      }
    }
  }, [])

  // Draw heatmap rectangles
  const drawRectangles = useCallback(() => {
    const map = mapRef.current as { geoObjects: { add(r: unknown): void; remove(r: unknown): void } } | null
    if (!map || !window.ymaps || cells.length === 0) return

    // Remove old rectangles
    for (const rect of rectanglesRef.current) {
      map.geoObjects.remove(rect)
    }
    rectanglesRef.current = []

    const apiCellSize = data?.data?.cell_size ?? cellSize
    const half = apiCellSize / 2
    const maxSearch = Math.max(...cells.map((c) => c.search_count), 1)
    const maxListing = Math.max(...cells.map((c) => c.listing_count), 1)

    for (const cell of cells) {
      const value = layer === 'demand' ? cell.search_count : cell.listing_count
      const maxVal = layer === 'demand' ? maxSearch : maxListing
      const colorFn = layer === 'demand' ? getDemandColor : getSupplyColor
      const fillColor = colorFn(value, maxVal)

      const rect = new (window.ymaps as unknown as {
        Rectangle: new (
          bounds: number[][],
          props: Record<string, string>,
          opts: Record<string, unknown>,
        ) => unknown
      }).Rectangle(
        [
          [cell.latitude - half, cell.longitude - half],
          [cell.latitude + half, cell.longitude + half],
        ],
        {
          hintContent: `Объекты: ${cell.listing_count}, Запросы: ${cell.search_count}, Бронирования: ${cell.booking_count}`,
        },
        {
          fillColor,
          strokeColor: 'rgba(0,0,0,0.05)',
          strokeWidth: 1,
          fillOpacity: 1,
        },
      )
      map.geoObjects.add(rect)
      rectanglesRef.current.push(rect)
    }

    // Auto-fit bounds if cells exist
    if (cells.length > 0) {
      const lats = cells.map((c) => c.latitude)
      const lons = cells.map((c) => c.longitude)
      const bounds = [
        [Math.min(...lats) - half, Math.min(...lons) - half],
        [Math.max(...lats) + half, Math.max(...lons) + half],
      ]
      ;(map as unknown as { setBounds(b: number[][], opts: Record<string, unknown>): void })
        .setBounds(bounds, { checkZoomRange: true, duration: 300 })
    }
  }, [cells, cellSize, layer, data?.data?.cell_size])

  useEffect(() => {
    if (mapReady) drawRectangles()
  }, [mapReady, drawRectangles])

  return (
    <div style={{ padding: 24 }}>
      <Title level={3}>
        <HeatMapOutlined /> Тепловая карта спроса и предложения
      </Title>

      <Row gutter={16} style={{ marginBottom: 16 }} align="middle">
        <Col>
          <Segmented
            options={PERIOD_OPTIONS}
            value={period}
            onChange={(val) => setPeriod(val as string)}
          />
        </Col>
        <Col flex="auto">
          <span style={{ marginRight: 8 }}>Размер ячейки:</span>
          <Slider
            style={{ display: 'inline-block', width: 200 }}
            min={0.005}
            max={0.1}
            step={0.005}
            marks={CELL_SIZE_MARKS}
            value={cellSize}
            onChange={(val) => setCellSize(val)}
            tooltip={{ formatter: (val) => `${val} deg` }}
          />
        </Col>
        <Col>
          <Radio.Group
            value={layer}
            onChange={(e) => setLayer(e.target.value)}
            optionType="button"
            buttonStyle="solid"
          >
            <Radio.Button value="demand">Спрос</Radio.Button>
            <Radio.Button value="supply">Предложение</Radio.Button>
          </Radio.Group>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={8}>
          <Card size="small">
            <Statistic
              title="Объекты"
              value={totalListings}
              prefix={<ShopOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card size="small">
            <Statistic
              title="Поисковые запросы"
              value={totalSearches}
              prefix={<SearchOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card size="small">
            <Statistic
              title="Бронирования"
              value={totalBookings}
              prefix={<EnvironmentOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      {error && (
        <Alert
          type="error"
          message="Ошибка загрузки данных"
          description="Не удалось загрузить данные тепловой карты"
          style={{ marginBottom: 16 }}
        />
      )}

      <Card data-testid="heatmap-card">
        {isLoading ? (
          <div style={{ textAlign: 'center', padding: 60 }}>
            <Spin size="large" />
          </div>
        ) : cells.length === 0 ? (
          <Alert type="info" message="Нет данных для отображения за выбранный период" />
        ) : null}
        <div
          ref={containerRef}
          data-testid="heatmap-map"
          style={{ width: '100%', height: 500, minHeight: 400 }}
        />
        {cells.length > 0 && (
          <div style={{ padding: '12px 0 0', display: 'flex', alignItems: 'center', gap: 8 }}>
            <span style={{ fontSize: 12, color: '#999' }}>
              {layer === 'demand' ? 'Спрос (запросы):' : 'Предложение (объекты):'}
            </span>
            <div style={{ display: 'flex', gap: 2 }}>
              {['Мало', '', '', '', 'Много'].map((label, i) => (
                <div key={i} style={{ textAlign: 'center' }}>
                  <div
                    style={{
                      width: 40,
                      height: 12,
                      borderRadius: 2,
                      background: layer === 'demand'
                        ? getDemandColor((i + 1) * 20, 100)
                        : getSupplyColor((i + 1) * 20, 100),
                    }}
                  />
                  {label && <div style={{ fontSize: 10, color: '#999' }}>{label}</div>}
                </div>
              ))}
            </div>
            <span style={{ fontSize: 12, color: '#999', marginLeft: 'auto' }}>
              {cells.length} ячеек
            </span>
          </div>
        )}
      </Card>
    </div>
  )
}
