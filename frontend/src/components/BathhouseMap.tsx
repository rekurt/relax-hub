import { useEffect, useRef, useCallback, useState } from 'react'
import { Button, Spin } from 'antd'
import { AimOutlined } from '@ant-design/icons'
import type { InternalHandlerBathhouseResponse } from '@/api/generated/model'
import { formatPrice } from '@/lib/format'

declare global {
  interface Window {
    ymaps?: YMaps
  }
}

interface YMaps {
  ready(cb: () => void): void
  Map: new (
    el: HTMLElement,
    state: { center: number[]; zoom: number; controls: string[] },
  ) => YMap
  Placemark: new (
    coords: number[],
    props: { balloonContentHeader?: string; balloonContentBody?: string; hintContent?: string },
    options?: Record<string, unknown>,
  ) => YPlacemark
  Clusterer: new (options?: Record<string, unknown>) => YClusterer
  Polygon: new (
    coordinates: number[][][],
    props?: Record<string, unknown>,
    options?: Record<string, unknown>,
  ) => YGeoObject
  templateLayoutFactory: {
    createClass(template: string): unknown
  }
}

interface YMap {
  geoObjects: { add(obj: unknown): void; remove(obj: unknown): void; removeAll(): void }
  events: { add(event: string, handler: () => void): void }
  getBounds(): number[][]
  getCenter(): number[]
  setCenter(center: number[], zoom?: number): void
  setType(type: string): void
  getType(): string
  destroy(): void
}

interface YPlacemark {
  events: { add(event: string, handler: () => void): void }
}

interface YClusterer {
  add(placemarks: YPlacemark[]): void
  removeAll(): void
}

interface YGeoObject {
  events: { add(event: string, handler: () => void): void }
}

interface BathhouseMapProps {
  bathhouses: InternalHandlerBathhouseResponse[]
  highlightedId?: string | null
  onBoundsChange?: (bounds: { north: number; south: number; east: number; west: number }) => void
  onMarkerClick?: (id: string) => void
  onMarkerHover?: (id: string | null) => void
  center?: { lat: number; lng: number }
  zoom?: number
  isochronePolygon?: number[][] | null // [lat, lng] pairs for Yandex Maps
  style?: React.CSSProperties
}

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

const DEFAULT_CENTER = { lat: 55.751244, lng: 37.618423 } // Moscow
const DEFAULT_ZOOM = 11

export default function BathhouseMap({
  bathhouses,
  highlightedId,
  onBoundsChange,
  onMarkerClick,
  onMarkerHover,
  center,
  zoom,
  isochronePolygon,
  style,
}: BathhouseMapProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<YMap | null>(null)
  const clustererRef = useRef<YClusterer | null>(null)
  const polygonRef = useRef<YGeoObject | null>(null)
  const [loading, setLoading] = useState(true)
  const [showSearchArea, setShowSearchArea] = useState(false)
  const [mapType, setMapType] = useState<'scheme' | 'satellite'>('scheme')

  const handleToggleMapType = useCallback(() => {
    setMapType((prev) => {
      const next = prev === 'scheme' ? 'satellite' : 'scheme'
      if (mapRef.current) {
        mapRef.current.setType(next === 'scheme' ? 'yandex#map' : 'yandex#satellite')
      }
      return next
    })
  }, [])

  const handleBoundsChange = useCallback(() => {
    if (!mapRef.current || !onBoundsChange) return
    setShowSearchArea(true)
  }, [onBoundsChange])

  const handleSearchArea = useCallback(() => {
    if (!mapRef.current || !onBoundsChange) return
    const bounds = mapRef.current.getBounds()
    if (bounds && bounds.length === 2) {
      const [sw, ne] = bounds as [number[], number[]]
      onBoundsChange({
        south: sw[0]!,
        west: sw[1]!,
        north: ne[0]!,
        east: ne[1]!,
      })
    }
    setShowSearchArea(false)
  }, [onBoundsChange])

  // Initialize map
  useEffect(() => {
    let destroyed = false

    loadYmaps()
      .then(() => {
        if (destroyed || !containerRef.current || !window.ymaps) return

        const mapCenter = center
          ? [center.lat, center.lng]
          : [DEFAULT_CENTER.lat, DEFAULT_CENTER.lng]

        const map = new window.ymaps.Map(containerRef.current, {
          center: mapCenter,
          zoom: zoom ?? DEFAULT_ZOOM,
          controls: ['zoomControl'],
        })

        mapRef.current = map

        const clusterer = new window.ymaps.Clusterer({
          preset: 'islands#invertedVioletClusterIcons',
          groupByCoordinates: false,
          clusterDisableClickZoom: false,
        })
        clustererRef.current = clusterer
        map.geoObjects.add(clusterer as unknown as YPlacemark)

        map.events.add('boundschange', handleBoundsChange)

        setLoading(false)
      })
      .catch(() => {
        setLoading(false)
      })

    return () => {
      destroyed = true
      if (mapRef.current) {
        mapRef.current.destroy()
        mapRef.current = null
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Update markers
  useEffect(() => {
    if (!mapRef.current || !window.ymaps || !clustererRef.current) return

    clustererRef.current.removeAll()

    const placemarks = bathhouses
      .filter((b) => b.latitude && b.longitude)
      .map((b) => {
        const isHighlighted = highlightedId === b.id
        const priceLabel = b.price_per_hour ? formatPrice(b.price_per_hour) : ''

        const placemark = new window.ymaps!.Placemark(
          [b.latitude!, b.longitude!],
          {
            balloonContentHeader: b.name ?? '',
            balloonContentBody: `<div>${b.address ?? ''}<br/><strong>${priceLabel}/ч</strong></div>`,
            hintContent: `${b.name} — ${priceLabel}/ч`,
          },
          {
            preset: isHighlighted
              ? 'islands#redDotIcon'
              : 'islands#violetDotIcon',
            iconLayout: 'default#imageWithContent',
            iconImageHref: '',
            iconImageSize: [0, 0],
            iconContentLayout: window.ymaps!.templateLayoutFactory.createClass(
              `<div style="
                background: ${isHighlighted ? '#ff4d4f' : '#fff'};
                color: ${isHighlighted ? '#fff' : '#333'};
                border: 2px solid ${isHighlighted ? '#ff4d4f' : '#722ed1'};
                border-radius: 16px;
                padding: 4px 10px;
                font-size: 12px;
                font-weight: 600;
                white-space: nowrap;
                box-shadow: 0 2px 6px rgba(0,0,0,0.15);
                transform: translate(-50%, -100%);
                cursor: pointer;
              ">${priceLabel}/ч</div>`,
            ),
          },
        )

        placemark.events.add('click', () => {
          if (b.id && onMarkerClick) onMarkerClick(b.id)
        })
        placemark.events.add('mouseenter', () => {
          if (b.id && onMarkerHover) onMarkerHover(b.id)
        })
        placemark.events.add('mouseleave', () => {
          if (onMarkerHover) onMarkerHover(null)
        })

        return placemark
      })

    clustererRef.current.add(placemarks)
  }, [bathhouses, highlightedId, onMarkerClick, onMarkerHover])

  // Update isochrone polygon overlay
  useEffect(() => {
    if (!mapRef.current || !window.ymaps) return

    // Remove previous polygon
    if (polygonRef.current) {
      mapRef.current.geoObjects.remove(polygonRef.current)
      polygonRef.current = null
    }

    // Add new polygon if present
    if (isochronePolygon && isochronePolygon.length > 2) {
      const polygon = new window.ymaps.Polygon(
        [isochronePolygon],
        { hintContent: 'Зона доступности' },
        {
          fillColor: '#722ed120',
          strokeColor: '#722ed1',
          strokeWidth: 2,
          strokeStyle: 'shortdash',
        },
      )
      polygonRef.current = polygon
      mapRef.current.geoObjects.add(polygon)
    }
  }, [isochronePolygon])

  return (
    <div style={{ position: 'relative', ...style }}>
      <Spin spinning={loading} style={{ width: '100%', height: '100%' }}>
        <div
          ref={containerRef}
          data-testid="bathhouse-map"
          style={{ width: '100%', height: '100%', minHeight: 400 }}
        />
      </Spin>

      {showSearchArea && onBoundsChange && (
        <Button
          type="primary"
          icon={<AimOutlined />}
          style={{
            position: 'absolute',
            top: 12,
            left: '50%',
            transform: 'translateX(-50%)',
            zIndex: 10,
            boxShadow: '0 2px 8px rgba(0,0,0,0.2)',
          }}
          onClick={handleSearchArea}
        >
          Искать в этой области
        </Button>
      )}

      {!loading && (
        <Button
          size="small"
          data-testid="map-layer-toggle"
          style={{
            position: 'absolute',
            top: 12,
            right: 12,
            zIndex: 10,
            boxShadow: '0 2px 8px rgba(0,0,0,0.2)',
            background: '#fff',
            fontWeight: 500,
          }}
          onClick={handleToggleMapType}
        >
          {mapType === 'scheme' ? 'Спутник' : 'Схема'}
        </Button>
      )}
    </div>
  )
}
