interface DesignAvatarProps {
  name?: string
  size?: number
  tone?: 'teal' | 'amber' | 'blue' | 'plum' | 'sand'
  className?: string
}

const PALETTES = {
  teal: ['#0f766e', '#0a5f59'],
  amber: ['#d97706', '#b45309'],
  blue: ['#2563eb', '#1d4ed8'],
  plum: ['#7c3aed', '#5b21b6'],
  sand: ['#e7d8b9', '#cdb585'],
} as const

function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(' ')
}

export default function DesignAvatar({ name = 'RelaxHUB', size = 36, tone = 'teal', className }: DesignAvatarProps) {
  const initials = name
    .split(/\s+/)
    .filter(Boolean)
    .map((part) => part[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()
  const [from, to] = PALETTES[tone]

  return (
    <span
      className={cx('rh-avatar inline-flex shrink-0 items-center justify-center rounded-rh-pill font-sans font-bold tracking-[0.02em] text-[#fffdf8]', className)}
      style={{
        width: size,
        height: size,
        background: `linear-gradient(135deg, ${from}, ${to})`,
        fontSize: Math.round(size * 0.34),
      }}
      aria-hidden="true"
    >
      {initials}
    </span>
  )
}
