import BrandLockup from '@/components/BrandLockup'
import TopNavigationLayout from '@/components/TopNavigationLayout'
import { PLATFORM_NAME } from '@/content/support'
import BathhouseSelector from '@/components/BathhouseSelector'
import { OWNER_OVERFLOW_NAV_ITEMS, OWNER_PRIMARY_NAV_ITEMS } from '@/navigation/menu'
import { useDocumentTitle, type DocumentTitleEntry } from '@/lib/useDocumentTitle'

const OWNER_TITLES: readonly DocumentTitleEntry[] = [
  ['/dashboard', `${PLATFORM_NAME} — Дашборд владельца`],
  ['/bathhouses/new', `${PLATFORM_NAME} — Новый объект`],
  ['/bathhouses/import', `${PLATFORM_NAME} — Массовый импорт`],
  ['/bathhouses', `${PLATFORM_NAME} — Мои объекты`],
  ['/bookings/extensions', `${PLATFORM_NAME} — Запросы на продление`],
  ['/bookings/modifications', `${PLATFORM_NAME} — Запросы на изменение`],
  ['/bookings', `${PLATFORM_NAME} — Бронирования`],
  ['/calendar', `${PLATFORM_NAME} — Календарь`],
  ['/chat', `${PLATFORM_NAME} — Чат с клиентами`],
  ['/notifications', `${PLATFORM_NAME} — Уведомления`],
  ['/reviews', `${PLATFORM_NAME} — Отзывы`],
  ['/photo-order', `${PLATFORM_NAME} — Заказ фотосъёмки`],
  ['/photos', `${PLATFORM_NAME} — Фотогалерея`],
  ['/finance/payouts', `${PLATFORM_NAME} — Выплаты`],
  ['/finance/reports', `${PLATFORM_NAME} — Финансовые отчёты`],
  ['/finance', `${PLATFORM_NAME} — Финансы`],
  ['/analytics', `${PLATFORM_NAME} — Аналитика объекта`],
  ['/promotion', `${PLATFORM_NAME} — Рекламные кампании`],
  ['/promo', `${PLATFORM_NAME} — Промокоды`],
  ['/pricing', `${PLATFORM_NAME} — Правила цен`],
  ['/subscriptions', `${PLATFORM_NAME} — Подписка тарифа`],
  ['/representatives', `${PLATFORM_NAME} — Представители`],
  ['/widget', `${PLATFORM_NAME} — Виджет бронирования`],
  ['/crm/guests', `${PLATFORM_NAME} — CRM: гости`],
  ['/crm/segments/custom', `${PLATFORM_NAME} — CRM: сегменты`],
  ['/crm/segments', `${PLATFORM_NAME} — CRM: сегменты`],
  ['/crm/broadcasts/new', `${PLATFORM_NAME} — CRM: новая рассылка`],
  ['/crm/broadcasts', `${PLATFORM_NAME} — CRM: рассылки`],
  ['/crm/scenarios', `${PLATFORM_NAME} — CRM: автосценарии`],
  ['/crm/templates', `${PLATFORM_NAME} — CRM: шаблоны ответов`],
  ['/crm/rfm', `${PLATFORM_NAME} — CRM: RFM-анализ`],
  ['/settings/webhooks', `${PLATFORM_NAME} — Вебхуки`],
  ['/settings/pms', `${PLATFORM_NAME} — Интеграция PMS`],
  ['/settings/kyc', `${PLATFORM_NAME} — Верификация KYC`],
  ['/settings/offer', `${PLATFORM_NAME} — Договор-оферта`],
  ['/settings', `${PLATFORM_NAME} — Настройки`],
]

export default function AppLayout() {
  useDocumentTitle(OWNER_TITLES, `${PLATFORM_NAME} — Кабинет владельца`)
  return (
    <TopNavigationLayout
      brandTitle={(
        <BrandLockup
          size="header"
          subtitle="Управление объектами и продажами"
          className="bani-topnav__brand-lockup"
        />
      )}
      brandSubtitle={null}
      brandAriaLabel={PLATFORM_NAME}
      surface="owner"
      homeTo="/dashboard"
      primaryItems={OWNER_PRIMARY_NAV_ITEMS}
      overflowItems={OWNER_OVERFLOW_NAV_ITEMS}
      navigationMode="dropdown"
      profilePath="/settings"
      headerAccessory={<BathhouseSelector />}
    />
  )
}
