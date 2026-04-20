import type { ReactNode } from 'react'
import { PLATFORM_NAME } from '@/content/support'

type BrandLockupTone = 'default' | 'inverse'
type BrandLockupSize = 'header' | 'footer' | 'auth'
type BrandLockupLayout = 'horizontal' | 'stacked'

interface BrandLockupProps {
  tone?: BrandLockupTone
  size?: BrandLockupSize
  layout?: BrandLockupLayout
  subtitle?: ReactNode
  className?: string
}

export default function BrandLockup({
  tone = 'default',
  size = 'header',
  layout = 'horizontal',
  subtitle,
  className,
}: BrandLockupProps) {
  const classes = [
    'bani-brand-lockup',
    `bani-brand-lockup--${tone}`,
    `bani-brand-lockup--${size}`,
    `bani-brand-lockup--${layout}`,
    className,
  ].filter(Boolean).join(' ')

  return (
    <div className={classes}>
      <div className="bani-brand-lockup__copy">
        <span className="bani-brand-lockup__title">{PLATFORM_NAME}</span>
        {subtitle && <span className="bani-brand-lockup__subtitle">{subtitle}</span>}
      </div>
    </div>
  )
}
