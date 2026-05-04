import type { ReactNode } from 'react'

export type DesignTagTone = 'default' | 'primary' | 'gold' | 'cyan' | 'red' | 'green' | 'ghost'

interface DesignTagProps {
  children: ReactNode
  tone?: DesignTagTone
  className?: string
}

function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(' ')
}

const toneClasses = {
  default: 'border-[rgba(15,23,42,0.12)] bg-white/80 text-rh-text',
  primary: 'rh-tag--primary border-[rgba(15,118,110,0.22)] bg-[rgba(15,118,110,0.10)] text-rh-primary-strong',
  gold: 'rh-tag--gold border-[rgba(217,119,6,0.28)] bg-[rgba(217,119,6,0.12)] text-[#92400e]',
  cyan: 'rh-tag--cyan border-[rgba(6,182,212,0.24)] bg-[rgba(6,182,212,0.10)] text-[#155e75]',
  red: 'rh-tag--red border-[rgba(180,35,24,0.22)] bg-[rgba(180,35,24,0.08)] text-[#991b1b]',
  green: 'rh-tag--green border-[rgba(21,128,61,0.24)] bg-[rgba(21,128,61,0.10)] text-rh-success',
  ghost: 'rh-tag--ghost border-[rgba(15,23,42,0.08)] bg-white/55 text-rh-text',
} as const

export default function DesignTag({ children, tone = 'default', className }: DesignTagProps) {
  return (
    <span className={cx('rh-tag inline-flex items-center gap-1.5 whitespace-nowrap rounded-rh-pill border px-3 py-1 font-sans text-xs font-semibold leading-none', toneClasses[tone], className)}>
      {children}
    </span>
  )
}
