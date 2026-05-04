import type { ReactNode } from 'react'

interface SectionHeaderProps {
  eyebrow?: ReactNode
  title: ReactNode
  subtitle?: ReactNode
  extra?: ReactNode
  className?: string
}

function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(' ')
}

export default function SectionHeader({ eyebrow, title, subtitle, extra, className }: SectionHeaderProps) {
  return (
    <header className={cx('rh-section-head flex items-end justify-between gap-6', className)}>
      <div className="rh-section-head__copy min-w-0 max-w-[720px]">
        {eyebrow && <span className="rh-eyebrow font-sans text-xs font-extrabold uppercase tracking-normal text-rh-primary">{eyebrow}</span>}
        <h1 className="rh-section-head__title m-0 mt-2 font-sans text-[clamp(28px,3vw,38px)] font-extrabold leading-[1.08] tracking-normal text-rh-text">{title}</h1>
        {subtitle && <p className="rh-section-head__subtitle m-0 mt-2 text-sm font-medium leading-[1.65] text-rh-text-soft">{subtitle}</p>}
      </div>
      {extra && <div className="rh-section-head__extra flex shrink-0 items-center gap-2">{extra}</div>}
    </header>
  )
}
