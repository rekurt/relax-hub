import { useEffect, type ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { Typography } from '@/components/design/system'
import BrandLockup from '@/components/BrandLockup'
import { PLATFORM_NAME } from '@/content/support'

const { Text } = Typography

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
    <div className="rh-auth-layout min-h-screen bg-[radial-gradient(circle_at_top_left,rgba(15,118,110,0.12),transparent_30%),radial-gradient(circle_at_top_right,rgba(217,119,6,0.10),transparent_24%),linear-gradient(180deg,#f8f2e8,#fffdf8)] px-4 py-6 font-sans text-rh-text">
      <div className="rh-auth-shell mx-auto grid min-h-[calc(100vh-48px)] w-full max-w-[1180px] grid-cols-1 overflow-hidden rounded-rh-3xl border border-[rgba(15,23,42,0.10)] bg-white/55 shadow-rh backdrop-blur-[18px] lg:grid-cols-[0.95fr_1.05fr]">
        <aside className="rh-auth-aside flex min-h-[360px] flex-col justify-between bg-[linear-gradient(135deg,#10313a_0%,#38606a_44%,#9a5c30_100%)] p-7 text-[#f7f4eb] sm:p-10">
          <Link to="/" className="rh-auth-brand inline-flex">
            <BrandLockup
              tone="inverse"
              size="auth"
              layout="stacked"
              subtitle="Маркетплейс бань и бронирований"
              className="rh-auth-brand__lockup"
            />
          </Link>

          <div className="rh-auth-aside__copy max-w-[460px]">
            <Text className="rh-tag rh-tag--gold rh-auth-badge">Быстрый вход</Text>
            <h1 className="rh-auth-aside__title">
              {asideTitle}
            </h1>
            <Text className="rh-auth-aside__description">
              {asideDescription}
            </Text>
          </div>

          <div className="rh-auth-highlights grid gap-3">
            {highlights.map((highlight, index) => (
              <div key={index} className="rh-card rh-card--flat rh-auth-highlight rounded-rh-xl border border-white/10 bg-white/10 p-4 text-sm font-medium leading-relaxed text-[#f7f4eb] shadow-none backdrop-blur-[12px]">
                {highlight}
              </div>
            ))}
          </div>
        </aside>

        <section className="rh-card rh-auth-panel rounded-none border-0 bg-[linear-gradient(180deg,rgba(255,255,255,0.96),rgba(255,252,246,0.90))] p-6 shadow-none sm:p-10 lg:rounded-l-none">
          <div className="rh-auth-panel__intro mb-6">
            {eyebrow && <Text className="rh-auth-panel__eyebrow">{eyebrow}</Text>}
            <h2 className="rh-auth-panel__title">
              {title}
            </h2>
            {description && (
              <Text type="secondary" className="rh-auth-panel__description">
                {description}
              </Text>
            )}
          </div>

          <div className="rh-auth-panel__content">{children}</div>

          {footer && <div className="rh-auth-panel__footer mt-6 border-t border-[rgba(15,23,42,0.08)] pt-5 text-sm text-rh-text-soft">{footer}</div>}
        </section>
      </div>
    </div>
  )
}
