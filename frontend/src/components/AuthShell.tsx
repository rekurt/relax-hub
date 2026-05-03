import { useEffect, type ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { Typography } from '@/components/design/system'
import BrandLockup from '@/components/BrandLockup'
import { PLATFORM_NAME } from '@/content/support'

const { Text, Title } = Typography

interface AuthShellProps {
  eyebrow?: ReactNode
  title: ReactNode
  description?: ReactNode
  asideTitle: ReactNode
  asideDescription: ReactNode
  highlights: ReactNode[]
  children: ReactNode
  footer?: ReactNode
}

export default function AuthShell({
  eyebrow,
  title,
  description,
  asideTitle,
  asideDescription,
  highlights,
  children,
  footer,
}: AuthShellProps) {
  useEffect(() => {
    const titleText = typeof title === 'string' ? title : 'Авторизация'
    document.title = `${PLATFORM_NAME} — ${titleText}`
  }, [title])

  return (
    <div className="rh-auth-layout">
      <div className="rh-auth-shell">
        <aside className="rh-auth-aside">
          <Link to="/" className="rh-auth-brand">
            <BrandLockup
              tone="inverse"
              size="auth"
              layout="stacked"
              subtitle="Маркетплейс бань и бронирований"
              className="rh-auth-brand__lockup"
            />
          </Link>

          <div className="rh-auth-aside__copy">
            <Text className="rh-tag rh-tag--gold rh-auth-badge">Быстрый вход</Text>
            <Title level={2} className="rh-auth-aside__title">
              {asideTitle}
            </Title>
            <Text className="rh-auth-aside__description">
              {asideDescription}
            </Text>
          </div>

          <div className="rh-auth-highlights">
            {highlights.map((highlight, index) => (
              <div key={index} className="rh-card rh-card--flat rh-auth-highlight">
                {highlight}
              </div>
            ))}
          </div>
        </aside>

        <section className="rh-card rh-auth-panel">
          <div className="rh-auth-panel__intro">
            {eyebrow && <Text className="rh-auth-panel__eyebrow">{eyebrow}</Text>}
            <Title level={3} className="rh-auth-panel__title">
              {title}
            </Title>
            {description && (
              <Text type="secondary" className="rh-auth-panel__description">
                {description}
              </Text>
            )}
          </div>

          <div className="rh-auth-panel__content">{children}</div>

          {footer && <div className="rh-auth-panel__footer">{footer}</div>}
        </section>
      </div>
    </div>
  )
}
