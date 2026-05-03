import BrandLockup from '@/components/BrandLockup'
import TopNavigationLayout from '@/components/TopNavigationLayout'
import { Admin2FABanner, Admin2FAToastBridge } from '@/components/Admin2FANotice'
import { PLATFORM_NAME } from '@/content/support'
import { ADMIN_OVERFLOW_NAV_ITEMS, ADMIN_PRIMARY_NAV_ITEMS } from '@/navigation/menu'
import { useDocumentTitle, type DocumentTitleEntry } from '@/lib/useDocumentTitle'

const ADMIN_TITLES: readonly DocumentTitleEntry[] = [
  ['/admin/dashboard', `${PLATFORM_NAME} — Админ-дашборд`],
  ['/admin/users', `${PLATFORM_NAME} — Пользователи`],
  ['/admin/roles', `${PLATFORM_NAME} — Роли админов`],
  ['/admin/bathhouses', `${PLATFORM_NAME} — Модерация бань`],
  ['/admin/reviews', `${PLATFORM_NAME} — Модерация отзывов`],
  ['/admin/photos', `${PLATFORM_NAME} — Проверка фото`],
  ['/admin/complaints', `${PLATFORM_NAME} — Жалобы`],
  ['/admin/cities', `${PLATFORM_NAME} — Города`],
  ['/admin/amenities', `${PLATFORM_NAME} — Удобства`],
  ['/admin/object-types', `${PLATFORM_NAME} — Категории объектов`],
  ['/admin/holidays', `${PLATFORM_NAME} — Праздники`],
  ['/admin/promo-codes', `${PLATFORM_NAME} — Промокоды`],
  ['/admin/notification-center', `${PLATFORM_NAME} — Центр уведомлений`],
  ['/admin/notifications', `${PLATFORM_NAME} — Уведомления`],
  ['/admin/profile', `${PLATFORM_NAME} — Профиль администратора`],
  ['/admin/antifraud', `${PLATFORM_NAME} — Антифрод`],
  ['/admin/tickets', `${PLATFORM_NAME} — Тикеты`],
  ['/admin/disputes', `${PLATFORM_NAME} — Споры`],
  ['/admin/wallets', `${PLATFORM_NAME} — Кошельки`],
  ['/admin/bookings', `${PLATFORM_NAME} — Бронирования`],
  ['/admin/finance', `${PLATFORM_NAME} — Финансы`],
  ['/admin/audit-log', `${PLATFORM_NAME} — Журнал аудита`],
  ['/admin/bank-reconciliation', `${PLATFORM_NAME} — Банковская сверка`],
  ['/admin/platform-settings', `${PLATFORM_NAME} — Настройки платформы`],
  ['/admin/feature-flags', `${PLATFORM_NAME} — Feature flags`],
  ['/admin/service-fee', `${PLATFORM_NAME} — Сервисный сбор`],
  ['/admin/funnels', `${PLATFORM_NAME} — Конверсионные воронки`],
  ['/admin/cohorts', `${PLATFORM_NAME} — Когорты`],
  ['/admin/supply-demand', `${PLATFORM_NAME} — Спрос/предложение`],
  ['/admin/force-majeure', `${PLATFORM_NAME} — Форс-мажор`],
  ['/admin/subscriptions', `${PLATFORM_NAME} — Подписки`],
  ['/admin/loyalty', `${PLATFORM_NAME} — Лояльность`],
  ['/admin/certificates', `${PLATFORM_NAME} — Сертификаты`],
  ['/admin/heatmap', `${PLATFORM_NAME} — Гео-хитмап`],
  ['/admin/faq', `${PLATFORM_NAME} — FAQ`],
  ['/admin/photo-orders', `${PLATFORM_NAME} — Заявки на фотосъёмку`],
  ['/admin', `${PLATFORM_NAME} — Админ-панель`],
]

export default function AdminLayout() {
  useDocumentTitle(ADMIN_TITLES, `${PLATFORM_NAME} — Админ-панель`)
  return (
    <>
      <Admin2FAToastBridge />
      <TopNavigationLayout
        brandTitle={(
          <BrandLockup
            size="header"
            subtitle="Модерация, финансы и контроль платформы"
            className="rh-topnav__brand-lockup"
          />
        )}
        brandSubtitle={null}
        brandAriaLabel={PLATFORM_NAME}
        surface="admin"
        homeTo="/admin"
        primaryItems={ADMIN_PRIMARY_NAV_ITEMS}
        overflowItems={ADMIN_OVERFLOW_NAV_ITEMS}
        navigationMode="dropdown"
        profilePath="/admin/profile"
        topBanner={<Admin2FABanner />}
      />
    </>
  )
}
