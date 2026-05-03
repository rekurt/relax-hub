import type { CSSProperties } from 'react'

type PhotoVariant = 'banya' | 'forest' | 'city' | 'spa' | 'sand'

interface PhotoPlaceholderProps {
  label?: string
  variant?: PhotoVariant
  className?: string
  style?: CSSProperties
}

export default function PhotoPlaceholder({
  label = 'Фото объекта',
  variant = 'banya',
  className,
  style,
}: PhotoPlaceholderProps) {
  return (
    <div className={`rh-photo rh-photo--${variant} ${className ?? ''}`.trim()} style={style}>
      <svg className="rh-photo__pattern" viewBox="0 0 320 220" aria-hidden="true">
        <circle cx="72" cy="62" r="60" />
        <circle cx="248" cy="168" r="80" />
      </svg>
      {label && <span className="rh-photo__label">{label}</span>}
    </div>
  )
}
