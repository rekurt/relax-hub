import type { ReactNode } from 'react'

interface SectionHeaderProps {
  eyebrow?: ReactNode
  title: ReactNode
  subtitle?: ReactNode
  extra?: ReactNode
  className?: string
}

export default function SectionHeader({ eyebrow, title, subtitle, extra, className }: SectionHeaderProps) {
  return (
    <header className={`rh-section-head ${className ?? ''}`.trim()}>
      <div className="rh-section-head__copy">
        {eyebrow && <span className="rh-eyebrow">{eyebrow}</span>}
        <h1 className="rh-section-head__title">{title}</h1>
        {subtitle && <p className="rh-section-head__subtitle">{subtitle}</p>}
      </div>
      {extra && <div className="rh-section-head__extra">{extra}</div>}
    </header>
  )
}
