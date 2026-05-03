import { useEffect, useRef, useCallback, useState } from 'react'
import { Button, Spin } from '@/components/design/system'
import { AimOutlined } from '@/components/design/icons'
import type { InternalHandlerBathhouseResponse } from '@/api/generated/model'
import { formatPrice } from '@/lib/format'
import { resolveAssetUrl } from '@/lib/asset-url'

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
}

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
  showMiniCard?: boolean
  center?: { lat: number; lng: number }
  zoom?: number
  isochronePolygon?: number[][] | null // [lat, lng] pairs for Yandex Maps
  style?: React.CSSProperties
}

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

const DEFAULT_CENTER = { lat: 55.751244, lng: 37.618423 } // Moscow
const DEFAULT_ZOOM = 11

export default function BathhouseMap({
  bathhouses,
  highlightedId,
  onBoundsChange,
  onMarkerClick,
  onMarkerHover,
  showMiniCard = false,
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
          preset: 'islands#invertedDarkGreenClusterIcons',
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

        const coverImage = resolveAssetUrl(b.images?.[0] ?? b.gallery_preview?.[0]?.url)
        const ratingStr = b.rating ? b.rating.toFixed(1) : '—'
        const reviewCountStr = b.review_count ?? 0
        const slug = b.slug ?? b.id ?? ''

        const safeName = escapeHtml(b.name ?? '')
        const safeAddress = escapeHtml(b.address ?? '')
        const safeSlug = encodeURIComponent(slug)

        const balloonBody = showMiniCard
          ? `<div style="min-width:208px;max-width:268px;font-family:Manrope,Segoe UI,sans-serif;color:#16212b;">
              ${coverImage ? `<img src="${escapeHtml(coverImage)}" alt="" style="width:100%;height:124px;object-fit:cover;border-radius:18px;margin-bottom:10px;" />` : ''}
              <div style="font-weight:800;font-size:15px;line-height:1.25;margin-bottom:5px;">${safeName}</div>
              <div style="color:#5f6877;font-size:12px;line-height:1.45;margin-bottom:8px;">${safeAddress}</div>
              <div style="display:flex;align-items:center;gap:8px;flex-wrap:wrap;margin-bottom:10px;">
                <span style="color:#d97706;font-weight:700;">★ ${ratingStr}</span>
                <span style="color:rgba(22,33,43,0.48);font-size:12px;">(${reviewCountStr})</span>
                <strong>${priceLabel}/ч</strong>
              </div>
              <a href="/bathhouses/${safeSlug}" style="display:inline-block;background:#0f766e;color:#fff;padding:8px 14px;border-radius:999px;text-decoration:none;font-size:13px;font-weight:800;box-shadow:0 12px 24px rgba(15,118,110,0.18);">Подробнее</a>
            </div>`
          : `<div style="font-family:Manrope,Segoe UI,sans-serif;color:#16212b;line-height:1.5;">${safeAddress}<br/><strong>${priceLabel}/ч</strong></div>`

        const placemark = new window.ymaps!.Placemark(
          [b.latitude!, b.longitude!],
          {
            balloonContentHeader: showMiniCard ? '' : safeName,
            balloonContentBody: balloonBody,
            hintContent: `${safeName} — ${priceLabel}/ч`,
          },
          {
            preset: isHighlighted
              ? 'islands#redDotIcon'
              : 'islands#darkGreenDotIcon',
            iconLayout: 'default#imageWithContent',
            iconImageHref: '',
            iconImageSize: [0, 0],
            iconContentLayout: window.ymaps!.templateLayoutFactory.createClass(
              `<div style="
                background: ${isHighlighted ? '#b42318' : '#fffdf8'};
                color: ${isHighlighted ? '#fff' : '#16212b'};
                border: 2px solid ${isHighlighted ? '#b42318' : '#0f766e'};
                border-radius: 999px;
                padding: 6px 12px;
                font-size: 12px;
                font-weight: 800;
                white-space: nowrap;
                box-shadow: 0 14px 28px rgba(15,23,42,0.14);
                transform: translate(-50%, -100%);
                cursor: pointer;
              ">${priceLabel}/ч</div>`,
            ),
          },
        )

        placemark.events.add('click', () => {
          if (showMiniCard) {
            // Let the balloon open naturally (default behavior)
          } else if (b.id && onMarkerClick) {
            onMarkerClick(b.id)
          }
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
  }, [bathhouses, highlightedId, onMarkerClick, onMarkerHover, showMiniCard])

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
          fillColor: '#0f766e22',
          strokeColor: '#0f766e',
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
            boxShadow: '0 14px 28px rgba(15,118,110,0.20)',
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
            borderRadius: 999,
            borderColor: 'rgba(15, 23, 42, 0.10)',
            boxShadow: '0 12px 24px rgba(15,23,42,0.12)',
            background: 'rgba(255, 253, 248, 0.92)',
            fontWeight: 700,
          }}
          onClick={handleToggleMapType}
        >
          {mapType === 'scheme' ? 'Спутник' : 'Схема'}
        </Button>
      )}
    </div>
  )
}
