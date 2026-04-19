import dayjs from 'dayjs'

export interface NavigationItem {
  key: string
  label: string
  to: string
  isActive: (pathname: string, searchParams: URLSearchParams) => boolean
}

export interface PublicShortcutCard {
  key: string
  title: string
  description: string
  eyebrow: string
  to: string
}

function exactItem(key: string, label: string, to: string): NavigationItem {
  return {
    key,
    label,
    to,
    isActive: (pathname) => pathname === to,
  }
}

function prefixItem(key: string, label: string, to: string, prefixes: string[] = [to]): NavigationItem {
  return {
    key,
    label,
    to,
    isActive: (pathname) => prefixes.some((prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`)),
  }
}

function buildCatalogPath(params: Record<string, string>): string {
  const search = new URLSearchParams(params)
  return `/catalog?${search.toString()}`
}

function nextSaturdayString() {
  let cursor = dayjs().startOf('day')
  while (cursor.day() !== 6) {
    cursor = cursor.add(1, 'day')
  }
  return cursor.format('YYYY-MM-DD')
}

function catalogItem(key: string, label: string, params?: Record<string, string>): NavigationItem {
  const to = params ? buildCatalogPath(params) : '/catalog'
  return {
    key,
    label,
    to,
    isActive: (pathname, searchParams) => {
      if (pathname !== '/catalog') return false
      if (!params) {
        return !searchParams.get('guest_count')
          && !searchParams.get('available_date')
          && !searchParams.get('has_hot_tub')
          && !searchParams.get('has_pool')
      }
      return Object.entries(params).every(([name, value]) => searchParams.get(name) === value)
    },
  }
}

export const PUBLIC_NAV_ITEMS: NavigationItem[] = [
  catalogItem('catalog-all', 'Все бани'),
  catalogItem('catalog-couple', 'Для двоих', { guest_count: '2' }),
  catalogItem('catalog-company', 'Для компании', { guest_count: '6' }),
  catalogItem('catalog-weekend', 'На выходные', { available_date: nextSaturdayString() }),
  catalogItem('catalog-hot-tub', 'С чаном', { has_hot_tub: 'true' }),
  catalogItem('catalog-pool', 'С бассейном', { has_pool: 'true' }),
  exactItem('certificates', 'Сертификаты', '/certificates'),
]

export const PUBLIC_SHORTCUT_CARDS: PublicShortcutCard[] = [
  {
    key: 'couple',
    eyebrow: 'Сценарий',
    title: 'Для двоих',
    description: 'Камерные приватные бани для спокойного вечера без лишнего шума.',
    to: buildCatalogPath({ guest_count: '2' }),
  },
  {
    key: 'company',
    eyebrow: 'Компания',
    title: 'Для компании',
    description: 'Большие объекты на 6+ гостей для дней рождений, сборов и выездов.',
    to: buildCatalogPath({ guest_count: '6' }),
  },
  {
    key: 'weekend',
    eyebrow: 'Выходные',
    title: 'На выходные',
    description: 'Подборка с доступностью на ближайшую субботу без ручной фильтрации.',
    to: buildCatalogPath({ available_date: nextSaturdayString() }),
  },
  {
    key: 'hot-tub',
    eyebrow: 'Комфорт',
    title: 'С чаном',
    description: 'Витрина объектов с чаном, уличным парением и вечерним сценарием отдыха.',
    to: buildCatalogPath({ has_hot_tub: 'true' }),
  },
]

export const CLIENT_PRIMARY_NAV_ITEMS: NavigationItem[] = [
  ...PUBLIC_NAV_ITEMS,
  prefixItem('client-bookings', 'Мои брони', '/client/bookings'),
]

export const CLIENT_OVERFLOW_NAV_ITEMS: NavigationItem[] = [
  prefixItem('client-favorites', 'Избранное', '/client/favorites'),
  prefixItem('client-wallet', 'Кошелек', '/client/wallet'),
  prefixItem('client-payments', 'Платежи', '/client/payments'),
  prefixItem('client-cards', 'Карты', '/client/cards'),
  prefixItem('client-chat', 'Чат', '/client/chat'),
  prefixItem('client-tickets', 'Поддержка', '/client/tickets'),
  prefixItem('client-disputes', 'Споры', '/client/disputes'),
  prefixItem('client-notifications', 'Уведомления', '/client/notifications'),
  prefixItem('client-security', 'Безопасность', '/client/security'),
  prefixItem('client-profile', 'Профиль', '/client/profile'),
]

export const OWNER_PRIMARY_NAV_ITEMS: NavigationItem[] = [
  exactItem('owner-dashboard', 'Обзор', '/dashboard'),
  prefixItem('owner-bathhouses', 'Объекты', '/bathhouses'),
  prefixItem('owner-bookings', 'Брони', '/bookings'),
  prefixItem('owner-calendar', 'Календарь', '/calendar'),
  prefixItem('owner-pricing', 'Цены', '/pricing'),
  prefixItem('owner-finance', 'Финансы', '/finance'),
  prefixItem('owner-crm', 'CRM', '/crm/guests', ['/crm']),
  prefixItem('owner-promotion', 'Продвижение', '/promotion', ['/promotion', '/promo']),
]

export const OWNER_OVERFLOW_NAV_ITEMS: NavigationItem[] = [
  prefixItem('owner-photos', 'Фото', '/photos', ['/photos', '/photo-order']),
  prefixItem('owner-widget', 'Виджет', '/widget'),
  prefixItem('owner-representatives', 'Представители', '/representatives'),
  prefixItem('owner-subscriptions', 'Подписки', '/subscriptions'),
  prefixItem('owner-notifications', 'Уведомления', '/notifications'),
  prefixItem('owner-pms', 'PMS', '/settings/pms'),
  prefixItem('owner-webhooks', 'Вебхуки', '/settings/webhooks'),
  prefixItem('owner-kyc', 'KYC', '/settings/kyc'),
  prefixItem('owner-offer', 'Оферта', '/settings/offer'),
  prefixItem('owner-profile', 'Профиль', '/settings'),
]

export const ADMIN_PRIMARY_NAV_ITEMS: NavigationItem[] = [
  exactItem('admin-dashboard', 'Обзор', '/admin'),
  prefixItem('admin-moderation', 'Модерация', '/admin/bathhouses', ['/admin/bathhouses', '/admin/reviews', '/admin/photos', '/admin/complaints']),
  prefixItem('admin-users', 'Пользователи', '/admin/users'),
  prefixItem('admin-bookings', 'Брони', '/admin/bookings', ['/admin/bookings', '/admin/disputes', '/admin/tickets']),
  prefixItem('admin-finance', 'Финансы', '/admin/finance'),
  prefixItem('admin-analytics', 'Аналитика', '/admin/analytics/funnels', ['/admin/analytics', '/admin/heatmap']),
  prefixItem('admin-settings', 'Настройки', '/admin/settings'),
]

export const ADMIN_OVERFLOW_NAV_ITEMS: NavigationItem[] = [
  prefixItem('admin-cities', 'Города', '/admin/cities'),
  prefixItem('admin-promos', 'Промокоды', '/admin/promos'),
  prefixItem('admin-antifraud', 'Антифрод', '/admin/antifraud'),
  prefixItem('admin-amenities', 'Удобства', '/admin/amenities'),
  prefixItem('admin-object-types', 'Типы объектов', '/admin/object-types'),
  prefixItem('admin-holidays', 'Праздники', '/admin/holidays'),
  prefixItem('admin-wallets', 'Кошельки', '/admin/wallets'),
  prefixItem('admin-roles', 'Роли', '/admin/roles'),
  prefixItem('admin-feature-flags', 'Флаги', '/admin/feature-flags'),
  prefixItem('admin-service-fees', 'Комиссии', '/admin/service-fees'),
  prefixItem('admin-faq', 'FAQ', '/admin/faq'),
  prefixItem('admin-profile', 'Профиль', '/admin/profile'),
]

export function getProfilePath(role?: string): string {
  switch (role) {
    case 'client':
      return '/client/profile'
    case 'admin':
      return '/admin/profile'
    default:
      return '/settings'
  }
}

export function getNotificationsPath(role?: string): string {
  if (role === 'client') return '/client/notifications'
  if (role === 'admin') return '/admin/notifications'
  return '/notifications'
}
