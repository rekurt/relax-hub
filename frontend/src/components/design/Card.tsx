import type { HTMLAttributes, ReactNode } from 'react'

interface DesignCardProps extends HTMLAttributes<HTMLElement> {
  children: ReactNode
  as?: 'article' | 'div' | 'section'
  tone?: 'default' | 'flat' | 'dark'
}

function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(' ')
}

const toneClasses = {
  default:
    'bg-[linear-gradient(180deg,rgba(255,255,255,0.92),rgba(255,252,246,0.84))] border-[rgba(15,23,42,0.12)] shadow-rh-soft backdrop-blur-[18px]',
  flat:
    'rh-card--flat border-[rgba(15,23,42,0.08)] bg-white/70 shadow-none backdrop-blur-[12px]',
  dark:
    'rh-card--dark border-white/10 bg-[radial-gradient(circle_at_top_left,rgba(15,118,110,0.32),transparent_28%),linear-gradient(180deg,rgba(17,27,33,0.98),rgba(12,18,24,0.98))] text-[#f7f4eb] shadow-rh',
} as const

export default function DesignCard({
  children,
  as: Component = 'div',
  tone = 'default',
  className,
  ...props
}: DesignCardProps) {
  return (
    <Component
      className={cx('rh-card min-w-0 rounded-rh-2xl border', toneClasses[tone], tone === 'dark' && 'rounded-rh-3xl', tone === 'flat' && 'rounded-[22px]', className)}
      {...props}
    >
      {children}
    </Component>
  )
}
