import type { ReactNode } from 'react'

interface PageHeaderProps {
  eyebrow?: ReactNode
  title: ReactNode
  description?: ReactNode
  extra?: ReactNode
}

export default function PageHeader({ eyebrow, title, description, extra }: PageHeaderProps) {
  return (
    <header className="bani-page-header">
      <div className="bani-page-header__copy">
        {eyebrow && (
          <div className="bani-page-header__eyebrow">
            {eyebrow}
          </div>
        )}
        <h1 className="bani-page-header__title">
          {title}
        </h1>
        {description && (
          <p className="bani-page-header__description">
            {description}
          </p>
        )}
      </div>

      {extra && <div className="bani-page-header__extra">{extra}</div>}
    </header>
  )
}
