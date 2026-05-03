import { useEffect, useState } from 'react'
import BrandLockup from '@/components/BrandLockup'
import OnboardingTour from '@/components/OnboardingTour'
import ShellFooter from '@/components/ShellFooter'
import TopNavigationLayout from '@/components/TopNavigationLayout'
import { PLATFORM_NAME } from '@/content/support'
import { useAuthStore } from '@/stores/auth'
import { CLIENT_DRAWER_SECTIONS, CLIENT_PRIMARY_NAV_ITEMS, CLIENT_PROFILE_MENU_ITEMS } from '@/navigation/menu'
import { useDocumentTitle, type DocumentTitleEntry } from '@/lib/useDocumentTitle'

const CLIENT_TITLES: readonly DocumentTitleEntry[] = [
  ['/client/search', `${PLATFORM_NAME} — Поиск бань`],
  ['/client/bathhouses/', `${PLATFORM_NAME} — Карточка бани`],
  ['/client/bookings/new', `${PLATFORM_NAME} — Новое бронирование`],
  ['/client/bookings/', `${PLATFORM_NAME} — Бронирование`],
  ['/client/bookings', `${PLATFORM_NAME} — Мои бронирования`],
  ['/client/favorites', `${PLATFORM_NAME} — Избранное`],
  ['/client/recommendations', `${PLATFORM_NAME} — Рекомендации`],
  ['/client/preferences', `${PLATFORM_NAME} — Предпочтения`],
  ['/client/loyalty', `${PLATFORM_NAME} — Программа лояльности`],
  ['/client/referral', `${PLATFORM_NAME} — Реферальная программа`],
  ['/client/certificates/purchase', `${PLATFORM_NAME} — Покупка сертификата`],
  ['/client/certificates', `${PLATFORM_NAME} — Сертификаты`],
  ['/client/payments', `${PLATFORM_NAME} — История платежей`],
  ['/client/profile', `${PLATFORM_NAME} — Профиль`],
  ['/client/chat', `${PLATFORM_NAME} — Чат с банями`],
  ['/client/notifications/preferences', `${PLATFORM_NAME} — Настройки уведомлений`],
  ['/client/notifications', `${PLATFORM_NAME} — Уведомления`],
  ['/client/saved-searches', `${PLATFORM_NAME} — Сохранённые поиски`],
  ['/client/comparison', `${PLATFORM_NAME} — Сравнение бань`],
  ['/client/tickets/', `${PLATFORM_NAME} — Тикет поддержки`],
  ['/client/tickets', `${PLATFORM_NAME} — Поддержка`],
  ['/client/disputes/new', `${PLATFORM_NAME} — Новый спор`],
  ['/client/disputes/', `${PLATFORM_NAME} — Спор`],
  ['/client/disputes', `${PLATFORM_NAME} — Споры`],
  ['/client/wallet', `${PLATFORM_NAME} — Кошелёк`],
  ['/client/saved-cards', `${PLATFORM_NAME} — Сохранённые карты`],
  ['/client/security', `${PLATFORM_NAME} — Безопасность`],
  ['/client/promo-codes', `${PLATFORM_NAME} — Активные промокоды`],
  ['/client', `${PLATFORM_NAME} — Кабинет клиента`],
]

export default function ClientLayout() {
  const [showTour, setShowTour] = useState(false)
  const { user, loadProfile } = useAuthStore()

  useDocumentTitle(CLIENT_TITLES, `${PLATFORM_NAME} — Кабинет клиента`)

  const shouldShowTour = !!user && !user.onboarding_completed

  useEffect(() => {
    if (!shouldShowTour) return
    const timer = setTimeout(() => setShowTour(true), 0)
    return () => clearTimeout(timer)
  }, [shouldShowTour])

  return (
    <>
      <TopNavigationLayout
        brandTitle={(
          <BrandLockup
            size="header"
            subtitle="Каталог, бронирование и личные поездки"
            className="bani-topnav__brand-lockup"
          />
        )}
        brandSubtitle={null}
        brandAriaLabel={PLATFORM_NAME}
        surface="client"
        homeTo="/"
        primaryItems={CLIENT_PRIMARY_NAV_ITEMS}
        drawerSections={CLIENT_DRAWER_SECTIONS}
        profileMenuItems={CLIENT_PROFILE_MENU_ITEMS}
        profilePath="/client/profile"
        footer={<ShellFooter showClientSection />}
      />
      {showTour && (
        <OnboardingTour
          open={showTour}
          onComplete={async () => {
            setShowTour(false)
            await loadProfile()
          }}
          region={user?.region}
        />
      )}
    </>
  )
}
