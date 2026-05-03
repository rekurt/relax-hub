import { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import { Card, Segmented, Spin, Alert, Slider, Row, Col, Statistic, Radio } from '@/components/design/system'
import { EnvironmentOutlined, ShopOutlined, SearchOutlined } from '@/components/design/icons'
import { useQuery } from '@tanstack/react-query'
import { axiosInstance } from '@/api/axios-instance'
import PageHeader from '@/components/PageHeader'

const YMAPS_API_KEY = import.meta.env.VITE_YMAPS_API_KEY || ''
const YMAPS_SRC = `https://api-maps.yandex.ru/2.1/?apikey=${YMAPS_API_KEY}&lang=ru_RU`
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

const LEGEND_STEPS = ['Мало', '', '', '', 'Много']

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
    <div className="rh-admin-heatmap-page">
      <PageHeader
        eyebrow="Геоаналитика"
        title="Тепловая карта спроса и предложения"
        description="Сопоставление поискового спроса, объектов и бронирований по географическим ячейкам."
        extra={(
          <Segmented
            className="rh-admin-heatmap-periods"
            options={PERIOD_OPTIONS}
            value={period}
            onChange={(val) => setPeriod(val as string)}
          />
        )}
      />

      <Card className="rh-admin-filter-card rh-admin-heatmap-controls">
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} md={14}>
            <div className="rh-admin-heatmap-slider">
              <span className="rh-admin-heatmap-slider__label">Размер ячейки</span>
              <Slider
                className="rh-admin-heatmap-slider__control"
                min={0.005}
                max={0.1}
                step={0.005}
                marks={CELL_SIZE_MARKS}
                value={cellSize}
                onChange={(val) => setCellSize(val)}
                tooltip={{ formatter: (val) => `${val} deg` }}
              />
            </div>
          </Col>
          <Col xs={24} md={10}>
            <Radio.Group
              className="rh-admin-heatmap-layer"
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
      </Card>

      <Row gutter={[16, 16]} className="rh-admin-heatmap-stats">
        <Col xs={24} sm={8}>
          <Card size="small" className="rh-admin-heatmap-stat rh-admin-heatmap-stat--supply">
            <Statistic
              title="Объекты"
              value={totalListings}
              prefix={<ShopOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card size="small" className="rh-admin-heatmap-stat rh-admin-heatmap-stat--demand">
            <Statistic
              title="Поисковые запросы"
              value={totalSearches}
              prefix={<SearchOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card size="small" className="rh-admin-heatmap-stat rh-admin-heatmap-stat--booking">
            <Statistic
              title="Бронирования"
              value={totalBookings}
              prefix={<EnvironmentOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {error && (
        <Alert
          type="error"
          title="Ошибка загрузки данных"
          description="Не удалось загрузить данные тепловой карты"
          className="rh-admin-alert"
        />
      )}

      <Card data-testid="heatmap-card" className="rh-admin-heatmap-card">
        {isLoading ? (
          <div className="rh-admin-state-card">
            <Spin size="large" />
          </div>
        ) : cells.length === 0 ? (
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Нет данных</div>
            <p className="rh-admin-empty-state__text">
              Нет данных для отображения за выбранный период. Попробуйте увеличить период или изменить размер ячейки.
            </p>
          </div>
        ) : null}
        <div
          ref={containerRef}
          data-testid="heatmap-map"
          className="rh-admin-heatmap-map"
        />
        {cells.length > 0 && (
          <div className="rh-admin-heatmap-legend">
            <span className="rh-admin-heatmap-legend__label">
              {layer === 'demand' ? 'Спрос (запросы):' : 'Предложение (объекты):'}
            </span>
            <div className="rh-admin-heatmap-legend__scale">
              {LEGEND_STEPS.map((label, i) => (
                <div key={`${layer}-${i}`} className="rh-admin-heatmap-legend__item">
                  <div
                    className={`rh-admin-heatmap-legend__swatch rh-admin-heatmap-legend__swatch--${layer}-${i}`}
                  />
                  {label && <div className="rh-admin-heatmap-legend__caption">{label}</div>}
                </div>
              ))}
            </div>
            <span className="rh-admin-heatmap-legend__total">
              {cells.length} ячеек
            </span>
          </div>
        )}
      </Card>
    </div>
  )
}
