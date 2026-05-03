import DesignIcon from '@/components/design/Icon'

interface RatingBadgeProps {
  value?: number | null
  count?: number | null
  className?: string
}

export default function RatingBadge({ value, count, className }: RatingBadgeProps) {
  const normalized = typeof value === 'number' && Number.isFinite(value) ? value : 0

  return (
    <span className={`rh-stars ${className ?? ''}`.trim()}>
      <DesignIcon name="star" size={14} className="rh-stars__icon" />
      <span className="rh-stars__value">
        {normalized.toFixed(1)}
        {count != null && <span className="rh-stars__count"> ({count})</span>}
      </span>
    </span>
  )
}
