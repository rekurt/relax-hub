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
    <div style={{ marginTop: 16 }}>
      <Text strong>Транспорт рядом:</Text>
      <div style={{ marginTop: 8 }}>
        {items.map((item, idx) => {
          const meta = TRANSPORT_LABELS[item.type] ?? { label: item.type, color: 'var(--rh-text-muted)' }
          return (
            <div key={`${item.type}-${idx}`} style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
              {item.type === 'metro' ? (
                <NodeIndexOutlined style={{ color: meta.color }} />
              ) : item.type === 'parking' ? (
                <CarOutlined style={{ color: meta.color }} />
              ) : (
                <EnvironmentOutlined style={{ color: meta.color }} />
              )}
              <Tag color={meta.color} style={{ margin: 0 }}>{meta.label}</Tag>
              <Text>{item.name}</Text>
              <Text type="secondary">— {formatDistance(item.distance_meters)}</Text>
            </div>
          )
        })}
      </div>
    </div>
  )
}
