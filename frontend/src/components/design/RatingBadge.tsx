import DesignIcon from '@/components/design/Icon'

interface RatingBadgeProps {
  value?: number | null
  count?: number | null
  className?: string
}

function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(' ')
}

export default function RatingBadge({ value, count, className }: RatingBadgeProps) {
  const normalized = typeof value === 'number' && Number.isFinite(value) ? value : 0
  const ratingText = count != null ? `${normalized.toFixed(1)} (${count})` : normalized.toFixed(1)

  return (
    <span className={cx('rh-stars inline-flex items-center gap-1 font-sans text-xs font-semibold tabular-nums text-[#faad14]', className)}>
      <DesignIcon name="star" size={14} className="rh-stars__icon fill-[#faad14] stroke-[#faad14]" />
      <span className="rh-stars__value">{ratingText}</span>
    </span>
  )
}
