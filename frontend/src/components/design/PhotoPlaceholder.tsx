import type { CSSProperties } from 'react'

type PhotoVariant = 'banya' | 'forest' | 'city' | 'spa' | 'sand'

interface PhotoPlaceholderProps {
  label?: string
  variant?: PhotoVariant
  className?: string
  style?: CSSProperties
}

const variantClasses = {
  banya: 'rh-photo--banya bg-[linear-gradient(135deg,#10313a,#38606a_50%,#9a5c30)]',
  forest: 'rh-photo--forest bg-[linear-gradient(135deg,#0f3a3a,#2a6a5e_60%,#5b8a6d)]',
  city: 'rh-photo--city bg-[linear-gradient(135deg,#1f2a3a,#3a4a5a_60%,#6a7a8a)]',
  spa: 'rh-photo--spa bg-[linear-gradient(135deg,#2a1f3a,#5e3a6a_50%,#b06a8a)]',
  sand: 'rh-photo--sand bg-[linear-gradient(135deg,#b08a4d,#d9b577_60%,#e7d8b9)]',
} as const

function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(' ')
}

export default function PhotoPlaceholder({
  label = 'Фото объекта',
  variant = 'banya',
  className,
  style,
}: PhotoPlaceholderProps) {
  return (
    <div
      className={cx('rh-photo relative flex min-h-[160px] items-center justify-center overflow-hidden rounded-rh-lg text-[11px] font-bold uppercase tracking-[0.18em] text-white/60', variantClasses[variant], className)}
      style={style}
    >
      <svg className="rh-photo__pattern absolute inset-0 h-full w-full opacity-15" viewBox="0 0 320 220" aria-hidden="true">
        <circle cx="72" cy="62" r="60" />
        <circle cx="248" cy="168" r="80" />
      </svg>
      {label && <span className="rh-photo__label relative z-[1]">{label}</span>}
    </div>
  )
}
