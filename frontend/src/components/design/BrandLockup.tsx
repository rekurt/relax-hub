import type { ReactNode } from 'react'
import { PLATFORM_NAME } from '@/content/support'

interface DesignBrandLockupProps {
  size?: number | 'header' | 'footer' | 'auth'
  dark?: boolean
  subtitle?: ReactNode
  className?: string
}

const SIZE_MAP = {
  header: 22,
  footer: 28,
  auth: 28,
} as const

export default function DesignBrandLockup({
  size = 22,
  dark = false,
  subtitle,
  className,
}: DesignBrandLockupProps) {
  const resolvedSize = typeof size === 'number' ? size : SIZE_MAP[size]

  return (
    <span className={`rh-brand-lockup ${dark ? 'rh-brand-lockup--dark' : ''} ${className ?? ''}`.trim()}>
      <span className="rh-brand-lockup__title" style={{ fontSize: resolvedSize }}>
        {PLATFORM_NAME}
      </span>
      {subtitle && <span className="rh-brand-lockup__subtitle">{subtitle}</span>}
    </span>
  )
}
