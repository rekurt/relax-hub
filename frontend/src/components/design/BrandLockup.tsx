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

function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(' ')
}

export default function DesignBrandLockup({
  size = 22,
  dark = false,
  subtitle,
  className,
}: DesignBrandLockupProps) {
  const resolvedSize = typeof size === 'number' ? size : SIZE_MAP[size]

  return (
    <span className={cx('rh-brand-lockup inline-flex flex-col font-sans leading-none', dark && 'rh-brand-lockup--dark', className)}>
      <span
        className={cx('rh-brand-lockup__title font-extrabold leading-[0.94] tracking-[-0.06em]', dark ? 'text-[#fffdf8]' : 'text-rh-text')}
        style={{ fontSize: resolvedSize }}
      >
        {PLATFORM_NAME}
      </span>
      {subtitle && (
        <span className={cx('rh-brand-lockup__subtitle mt-1.5 text-[10px] font-bold uppercase tracking-[0.14em]', dark ? 'text-[rgba(255,255,255,0.66)]' : 'text-rh-text-soft')}>
          {subtitle}
        </span>
      )}
    </span>
  )
}
