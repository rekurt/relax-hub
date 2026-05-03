import type { ReactNode } from 'react'
import DesignBrandLockup from '@/components/design/BrandLockup'

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
    <DesignBrandLockup
      className={classes}
      dark={tone === 'inverse'}
      size={size}
      subtitle={subtitle}
    />
  )
}
