import { Link } from 'react-router-dom'
import BrandLockup from '@/components/BrandLockup'
import { CLIENT_FOOTER_ACCOUNT_ITEMS, PUBLIC_FOOTER_NAV_ITEMS, type NavigationItem } from '@/navigation/menu'
import { PLATFORM_BRAND_STATEMENT, PLATFORM_CONTACTS, PLATFORM_NAME } from '@/content/support'

interface ShellFooterProps {
  showClientSection?: boolean
}

function FooterLinks({ items }: { items: NavigationItem[] }) {
  return (
    <div className="rh-footer__links rh-shell-footer__links grid gap-2">
      {items.map((item) => (
        <Link key={item.key} to={item.to} className="rh-footer__link rh-shell-footer__link text-sm font-medium leading-relaxed text-[rgba(247,244,235,0.72)] transition hover:text-[#f7f4eb]">
          {item.label}
        </Link>
      ))}
    </div>
  )
}

export default function ShellFooter({ showClientSection = false }: ShellFooterProps) {
  const supportMessengerLink = `https://t.me/${PLATFORM_CONTACTS.supportMessenger.replace(/^@/, '')}`

  return (
    <footer className="rh-footer rh-shell-footer rounded-rh-3xl border border-white/10 bg-[radial-gradient(circle_at_top_left,rgba(15,118,110,0.32),transparent_28%),linear-gradient(180deg,rgba(17,27,33,0.98),rgba(12,18,24,0.98))] p-6 text-[#f7f4eb] shadow-rh sm:p-8">
      <div className="rh-footer__grid rh-shell-footer__grid grid gap-8 lg:grid-cols-[1.4fr_0.8fr_0.8fr_0.9fr]">
        <section className="rh-shell-footer__column rh-shell-footer__column--brand">
          <BrandLockup
            tone="inverse"
            size="footer"
            layout="stacked"
            subtitle="Премиальный маркетплейс бронирования"
            className="rh-shell-footer__lockup"
          />
          <h2 className="rh-shell-footer__title mt-5 max-w-[420px] text-2xl font-extrabold leading-tight tracking-normal">Сервис бронирования с премиальной подачей и понятной навигацией.</h2>
          <p className="rh-footer__description rh-shell-footer__description mt-3 max-w-[460px] text-sm leading-relaxed text-[rgba(247,244,235,0.66)]">{PLATFORM_BRAND_STATEMENT}</p>
          <div className="rh-shell-footer__support-hours mt-4 inline-flex rounded-rh-pill border border-white/10 bg-white/10 px-3 py-1.5 text-xs font-bold uppercase tracking-normal text-[rgba(247,244,235,0.78)]">{PLATFORM_CONTACTS.supportHours}</div>
        </section>

        <section className="rh-shell-footer__column">
          <div className="rh-footer__section-title rh-shell-footer__section-title mb-3 text-xs font-extrabold uppercase tracking-normal text-[rgba(247,244,235,0.54)]">Навигация</div>
          <FooterLinks items={PUBLIC_FOOTER_NAV_ITEMS} />
        </section>

        {showClientSection && (
          <section className="rh-shell-footer__column">
            <div className="rh-footer__section-title rh-shell-footer__section-title mb-3 text-xs font-extrabold uppercase tracking-normal text-[rgba(247,244,235,0.54)]">Кабинет клиента</div>
            <FooterLinks items={CLIENT_FOOTER_ACCOUNT_ITEMS} />
          </section>
        )}

        <section className="rh-shell-footer__column">
          <div className="rh-footer__section-title rh-shell-footer__section-title mb-3 text-xs font-extrabold uppercase tracking-normal text-[rgba(247,244,235,0.54)]">Поддержка и условия</div>
          <div className="rh-footer__links rh-shell-footer__links grid gap-2">
            <Link to="/terms" className="rh-footer__link rh-shell-footer__link text-sm font-medium leading-relaxed text-[rgba(247,244,235,0.72)] transition hover:text-[#f7f4eb]">Условия использования</Link>
            <a href={`mailto:${PLATFORM_CONTACTS.supportEmail}`} className="rh-footer__link rh-shell-footer__link text-sm font-medium leading-relaxed text-[rgba(247,244,235,0.72)] transition hover:text-[#f7f4eb]">
              {PLATFORM_CONTACTS.supportEmail}
            </a>
            <a href="tel:+74955552121" className="rh-footer__link rh-shell-footer__link text-sm font-medium leading-relaxed text-[rgba(247,244,235,0.72)] transition hover:text-[#f7f4eb]">
              {PLATFORM_CONTACTS.supportPhone}
            </a>
            <a href={supportMessengerLink} className="rh-footer__link rh-shell-footer__link text-sm font-medium leading-relaxed text-[rgba(247,244,235,0.72)] transition hover:text-[#f7f4eb]" target="_blank" rel="noreferrer">
              {PLATFORM_CONTACTS.supportMessenger}
            </a>
          </div>
        </section>
      </div>

      <div className="rh-footer__bottom rh-shell-footer__bottom mt-8 flex flex-col gap-2 border-t border-white/10 pt-5 text-xs leading-relaxed text-[rgba(247,244,235,0.54)] sm:flex-row sm:items-center sm:justify-between">
        <span>© {PLATFORM_NAME} {new Date().getFullYear()}</span>
        <span>Платформа организует бронирование и оплату, а фактические услуги оказывает выбранный объект.</span>
      </div>
    </footer>
  )
}
