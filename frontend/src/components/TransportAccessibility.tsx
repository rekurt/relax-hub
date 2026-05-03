import { Typography, Tag } from '@/components/design/system'
import {
  EnvironmentOutlined,
  CarOutlined,
  NodeIndexOutlined,
} from '@/components/design/icons'

const { Text } = Typography

export interface TransportItem {
  type: 'metro' | 'bus_stop' | 'parking'
  name: string
  distance_meters: number
  lat: number
  lng: number
}

const TRANSPORT_LABELS: Record<string, { label: string; color: string }> = {
  metro: { label: 'Метро', color: '#0f766e' },
  bus_stop: { label: 'Остановка', color: '#15803d' },
  parking: { label: 'Парковка', color: '#d97706' },
}

function formatDistance(meters: number): string {
  if (meters >= 1000) {
    return `${(meters / 1000).toFixed(1)} км`
  }
  return `${meters} м`
}

interface TransportAccessibilityProps {
  items: TransportItem[]
}

export default function TransportAccessibility({ items }: TransportAccessibilityProps) {
  if (items.length === 0) return null

  return (
    <div className="rh-transport">
      <Text strong>Транспорт рядом:</Text>
      <div className="rh-transport__list">
        {items.map((item, idx) => {
          const meta = TRANSPORT_LABELS[item.type] ?? { label: item.type, color: 'var(--rh-text-muted)' }
          return (
            <div key={`${item.type}-${idx}`} className="rh-transport__item">
              {item.type === 'metro' ? (
                <NodeIndexOutlined className="rh-transport__icon rh-transport__icon--metro" />
              ) : item.type === 'parking' ? (
                <CarOutlined className="rh-transport__icon rh-transport__icon--parking" />
              ) : (
                <EnvironmentOutlined className="rh-transport__icon rh-transport__icon--bus" />
              )}
              <Tag color={meta.color} className="rh-compact-tag">{meta.label}</Tag>
              <Text>{item.name}</Text>
              <Text type="secondary">— {formatDistance(item.distance_meters)}</Text>
            </div>
          )
        })}
      </div>
    </div>
  )
}
