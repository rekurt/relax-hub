import { useEffect, type ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { Typography } from 'antd'
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
    <div className="bani-auth-layout">
      <div className="bani-auth-shell">
        <aside className="bani-auth-aside">
          <Link to="/" className="bani-auth-brand">
            <BrandLockup
              tone="inverse"
              size="auth"
              layout="stacked"
              subtitle="Маркетплейс бань и бронирований"
              className="bani-auth-brand__lockup"
            />
          </Link>

          <div className="bani-auth-aside__copy">
            <Text className="rh-tag rh-tag--gold bani-auth-badge">Быстрый вход</Text>
            <Title level={2} className="bani-auth-aside__title">
              {asideTitle}
            </Title>
            <Text className="bani-auth-aside__description">
              {asideDescription}
            </Text>
          </div>

          <div className="bani-auth-highlights">
            {highlights.map((highlight, index) => (
              <div key={index} className="rh-card rh-card--flat bani-auth-highlight">
                {highlight}
              </div>
            ))}
          </div>
        </aside>

        <section className="rh-card bani-auth-panel">
          <div className="bani-auth-panel__intro">
            {eyebrow && <Text className="bani-auth-panel__eyebrow">{eyebrow}</Text>}
            <Title level={3} className="bani-auth-panel__title">
              {title}
            </Title>
            {description && (
              <Text type="secondary" className="bani-auth-panel__description">
                {description}
              </Text>
            )}
          </div>

          <div className="bani-auth-panel__content">{children}</div>

          {footer && <div className="bani-auth-panel__footer">{footer}</div>}
        </section>
      </div>
    </div>
  )
}
