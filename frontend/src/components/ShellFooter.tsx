import { Link } from 'react-router-dom'
import BrandLockup from '@/components/BrandLockup'
import { CLIENT_FOOTER_ACCOUNT_ITEMS, PUBLIC_FOOTER_NAV_ITEMS, type NavigationItem } from '@/navigation/menu'
import { PLATFORM_BRAND_STATEMENT, PLATFORM_CONTACTS, PLATFORM_NAME } from '@/content/support'

interface ShellFooterProps {
  showClientSection?: boolean
}

function FooterLinks({ items }: { items: NavigationItem[] }) {
  return (
    <div className="bani-shell-footer__links">
      {items.map((item) => (
        <Link key={item.key} to={item.to} className="bani-shell-footer__link">
          {item.label}
        </Link>
      ))}
    </div>
  )
}

export default function ShellFooter({ showClientSection = false }: ShellFooterProps) {
  const supportMessengerLink = `https://t.me/${PLATFORM_CONTACTS.supportMessenger.replace(/^@/, '')}`

  return (
    <footer className="bani-shell-footer">
      <div className="bani-shell-footer__grid">
        <section className="bani-shell-footer__column bani-shell-footer__column--brand">
          <BrandLockup
            tone="inverse"
            size="footer"
            layout="stacked"
            subtitle="Премиальный маркетплейс бронирования"
            className="bani-shell-footer__lockup"
          />
          <h2 className="bani-shell-footer__title">Сервис бронирования с премиальной подачей и понятной навигацией.</h2>
          <p className="bani-shell-footer__description">{PLATFORM_BRAND_STATEMENT}</p>
          <div className="bani-shell-footer__support-hours">{PLATFORM_CONTACTS.supportHours}</div>
        </section>

        <section className="bani-shell-footer__column">
          <div className="bani-shell-footer__section-title">Навигация</div>
          <FooterLinks items={PUBLIC_FOOTER_NAV_ITEMS} />
        </section>

        {showClientSection && (
          <section className="bani-shell-footer__column">
            <div className="bani-shell-footer__section-title">Кабинет клиента</div>
            <FooterLinks items={CLIENT_FOOTER_ACCOUNT_ITEMS} />
          </section>
        )}

        <section className="bani-shell-footer__column">
          <div className="bani-shell-footer__section-title">Поддержка и условия</div>
          <div className="bani-shell-footer__links">
            <Link to="/terms" className="bani-shell-footer__link">Terms of Use</Link>
            <a href={`mailto:${PLATFORM_CONTACTS.supportEmail}`} className="bani-shell-footer__link">
              {PLATFORM_CONTACTS.supportEmail}
            </a>
            <a href="tel:+74955552121" className="bani-shell-footer__link">
              {PLATFORM_CONTACTS.supportPhone}
            </a>
            <a href={supportMessengerLink} className="bani-shell-footer__link" target="_blank" rel="noreferrer">
              {PLATFORM_CONTACTS.supportMessenger}
            </a>
          </div>
        </section>
      </div>

      <div className="bani-shell-footer__bottom">
        <span>© {PLATFORM_NAME} {new Date().getFullYear()}</span>
        <span>Платформа организует бронирование и оплату, а фактические услуги оказывает выбранный объект.</span>
      </div>
    </footer>
  )
}
