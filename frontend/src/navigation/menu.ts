import dayjs from 'dayjs'

export interface NavigationItem {
  key: string
  label: string
  to: string
  section?: string
  isActive: (pathname: string, searchParams: URLSearchParams) => boolean
}

export interface NavigationSection {
  key: string
  title: string
  items: NavigationItem[]
}

export interface PublicShortcutCard {
  key: string
  title: string
  description: string
  eyebrow: string
  to: string
  params: Record<string, string>
}

function exactItem(key: string, label: string, to: string, section?: string): NavigationItem {
  return {
    key,
    label,
    to,
    section,
    isActive: (pathname) => pathname === to,
  }
}

function prefixItem(key: string, label: string, to: string, prefixes: string[] = [to], section?: string): NavigationItem {
  return {
    key,
    label,
    to,
    section,
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

function catalogItem(key: string, label: string, params?: Record<string, string>, section?: string): NavigationItem {
  const to = params ? buildCatalogPath(params) : '/catalog'
  return {
    key,
    label,
    to,
    section,
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

const PUBLIC_FILTER_NAV_ITEMS: NavigationItem[] = [
  catalogItem('catalog-couple', 'Для двоих', { guest_count: '2' }, 'Подборки'),
  catalogItem('catalog-company', 'Для компании', { guest_count: '6' }, 'Подборки'),
  catalogItem('catalog-weekend', 'На выходные', { available_date: nextSaturdayString() }, 'Подборки'),
  catalogItem('catalog-hot-tub', 'С чаном', { has_hot_tub: 'true' }, 'Подборки'),
  catalogItem('catalog-pool', 'С бассейном', { has_pool: 'true' }, 'Подборки'),
]

const PUBLIC_SUPPORT_NAV_ITEMS: NavigationItem[] = [
  exactItem('faq', 'Помощь', '/faq'),
  exactItem('contacts', 'Контакты', '/contacts'),
]

export const PUBLIC_PRIMARY_NAV_ITEMS: NavigationItem[] = [
  catalogItem('catalog-all', 'Каталог'),
  exactItem('certificates', 'Сертификаты', '/certificates'),
  ...PUBLIC_SUPPORT_NAV_ITEMS,
]

export const PUBLIC_OVERFLOW_NAV_ITEMS: NavigationItem[] = [
  ...PUBLIC_FILTER_NAV_ITEMS,
]

export const PUBLIC_NAV_ITEMS: NavigationItem[] = [
  ...PUBLIC_PRIMARY_NAV_ITEMS,
  ...PUBLIC_OVERFLOW_NAV_ITEMS,
]

export const PUBLIC_SHORTCUT_CARDS: PublicShortcutCard[] = [
  {
    key: 'couple',
    eyebrow: 'Сценарий',
    title: 'Для двоих',
    description: 'Камерные приватные бани для спокойного вечера без лишнего шума.',
    params: { guest_count: '2' },
    to: buildCatalogPath({ guest_count: '2' }),
  },
  {
    key: 'company',
    eyebrow: 'Компания',
    title: 'Для компании',
    description: 'Большие объекты на 6+ гостей для дней рождений, сборов и выездов.',
    params: { guest_count: '6' },
    to: buildCatalogPath({ guest_count: '6' }),
  },
  {
    key: 'weekend',
    eyebrow: 'Выходные',
    title: 'На выходные',
    description: 'Подборка с доступностью на ближайшую субботу без ручной фильтрации.',
    params: { available_date: nextSaturdayString() },
    to: buildCatalogPath({ available_date: nextSaturdayString() }),
  },
  {
    key: 'hot-tub',
    eyebrow: 'Комфорт',
    title: 'С чаном',
    description: 'Витрина объектов с чаном, уличным парением и вечерним сценарием отдыха.',
    params: { has_hot_tub: 'true' },
    to: buildCatalogPath({ has_hot_tub: 'true' }),
  },
  {
    key: 'pool',
    eyebrow: 'Комфорт',
    title: 'С бассейном',
    description: 'Объекты с бассейном для долгого маршрута отдыха без отдельного поиска по удобствам.',
    params: { has_pool: 'true' },
    to: buildCatalogPath({ has_pool: 'true' }),
  },
]

export const CLIENT_PRIMARY_NAV_ITEMS: NavigationItem[] = [
  catalogItem('catalog-all', 'Каталог'),
  prefixItem('client-bookings', 'Мои брони', '/client/bookings'),
  exactItem('client-certificates', 'Сертификаты', '/certificates'),
  exactItem('client-faq', 'Помощь', '/faq'),
  exactItem('client-contacts', 'Контакты', '/contacts'),
]

export const CLIENT_OVERFLOW_NAV_ITEMS: NavigationItem[] = [
  ...PUBLIC_FILTER_NAV_ITEMS,
]

export const CLIENT_PROFILE_MENU_ITEMS: NavigationItem[] = [
  prefixItem('client-profile', 'Профиль', '/client/profile', ['/client/profile'], 'Аккаунт'),
  prefixItem('client-notifications', 'Уведомления', '/client/notification-preferences', ['/client/notification-preferences'], 'Аккаунт'),
  prefixItem('client-security', 'Безопасность', '/client/security', ['/client/security'], 'Аккаунт'),
  prefixItem('client-wallet', 'Кошелёк', '/client/wallet', ['/client/wallet'], 'Финансы'),
  prefixItem('client-payments', 'Платежи', '/client/payments', ['/client/payments'], 'Финансы'),
  prefixItem('client-cards', 'Карты', '/client/cards', ['/client/cards'], 'Финансы'),
]

export const CLIENT_MOBILE_SERVICE_ITEMS: NavigationItem[] = [
  prefixItem('client-favorites', 'Избранное', '/client/favorites', ['/client/favorites']),
  prefixItem('client-chat', 'Чат', '/client/chat', ['/client/chat']),
  prefixItem('client-tickets', 'Поддержка', '/client/tickets', ['/client/tickets']),
  prefixItem('client-disputes', 'Споры', '/client/disputes', ['/client/disputes']),
]

export const PUBLIC_DRAWER_SECTIONS: NavigationSection[] = [
  {
    key: 'public-navigation',
    title: 'Навигация',
    items: PUBLIC_PRIMARY_NAV_ITEMS,
  },
  {
    key: 'public-picks',
    title: 'Подборки',
    items: PUBLIC_FILTER_NAV_ITEMS,
  },
]

export const CLIENT_DRAWER_SECTIONS: NavigationSection[] = [
  {
    key: 'client-navigation',
    title: 'Навигация',
    items: CLIENT_PRIMARY_NAV_ITEMS,
  },
  {
    key: 'client-account',
    title: 'Аккаунт и финансы',
    items: CLIENT_PROFILE_MENU_ITEMS,
  },
  {
    key: 'client-services',
    title: 'Сервисы',
    items: CLIENT_MOBILE_SERVICE_ITEMS,
  },
]

export const PUBLIC_FOOTER_NAV_ITEMS: NavigationItem[] = [
  catalogItem('footer-catalog', 'Каталог'),
  exactItem('footer-certificates', 'Сертификаты', '/certificates'),
  exactItem('footer-help', 'Помощь', '/faq'),
  exactItem('footer-contacts', 'Контакты', '/contacts'),
]

export const CLIENT_FOOTER_ACCOUNT_ITEMS: NavigationItem[] = [
  prefixItem('footer-client-bookings', 'Мои брони', '/client/bookings', ['/client/bookings']),
  prefixItem('footer-client-wallet', 'Кошелёк', '/client/wallet', ['/client/wallet']),
  prefixItem('footer-client-payments', 'Платежи', '/client/payments', ['/client/payments']),
  prefixItem('footer-client-cards', 'Карты', '/client/cards', ['/client/cards']),
]

export const OWNER_PRIMARY_NAV_ITEMS: NavigationItem[] = [
  exactItem('owner-dashboard', 'Обзор', '/dashboard'),
  prefixItem('owner-bathhouses', 'Объекты', '/bathhouses'),
  prefixItem('owner-bookings', 'Брони', '/bookings'),
]

export const OWNER_OVERFLOW_NAV_ITEMS: NavigationItem[] = [
  prefixItem('owner-calendar', 'Календарь', '/calendar', ['/calendar'], 'Операции'),
  prefixItem('owner-pricing', 'Цены', '/pricing', ['/pricing'], 'Операции'),
  prefixItem('owner-finance', 'Финансы', '/finance', ['/finance'], 'Операции'),
  prefixItem('owner-crm', 'CRM', '/crm/guests', ['/crm'], 'Операции'),
  prefixItem('owner-promotion', 'Продвижение', '/promotion', ['/promotion', '/promo'], 'Рост'),
  prefixItem('owner-photos', 'Фото', '/photos', ['/photos', '/photo-order'], 'Рост'),
  prefixItem('owner-widget', 'Виджет', '/widget', ['/widget'], 'Рост'),
  prefixItem('owner-representatives', 'Представители', '/representatives', ['/representatives'], 'Рост'),
  prefixItem('owner-subscriptions', 'Подписки', '/subscriptions', ['/subscriptions'], 'Аккаунт'),
  prefixItem('owner-notifications', 'Уведомления', '/notifications', ['/notifications'], 'Аккаунт'),
  prefixItem('owner-pms', 'PMS', '/settings/pms', ['/settings/pms'], 'Настройки'),
  prefixItem('owner-webhooks', 'Вебхуки', '/settings/webhooks', ['/settings/webhooks'], 'Настройки'),
  prefixItem('owner-kyc', 'KYC', '/settings/kyc', ['/settings/kyc'], 'Настройки'),
  prefixItem('owner-offer', 'Оферта', '/settings/offer', ['/settings/offer'], 'Настройки'),
  prefixItem('owner-profile', 'Профиль', '/settings', ['/settings'], 'Настройки'),
]

export const ADMIN_PRIMARY_NAV_ITEMS: NavigationItem[] = [
  exactItem('admin-dashboard', 'Обзор', '/admin'),
  prefixItem('admin-moderation', 'Модерация', '/admin/bathhouses', ['/admin/bathhouses', '/admin/reviews', '/admin/photos', '/admin/complaints']),
  prefixItem('admin-users', 'Пользователи', '/admin/users'),
]

export const ADMIN_OVERFLOW_NAV_ITEMS: NavigationItem[] = [
  prefixItem('admin-bookings', 'Брони', '/admin/bookings', ['/admin/bookings', '/admin/disputes', '/admin/tickets'], 'Операции'),
  prefixItem('admin-finance', 'Финансы', '/admin/finance', ['/admin/finance'], 'Операции'),
  prefixItem('admin-analytics', 'Аналитика', '/admin/analytics/funnels', ['/admin/analytics', '/admin/heatmap'], 'Операции'),
  prefixItem('admin-settings', 'Настройки', '/admin/settings', ['/admin/settings'], 'Операции'),
  prefixItem('admin-cities', 'Города', '/admin/cities', ['/admin/cities'], 'Справочники'),
  prefixItem('admin-promos', 'Промокоды', '/admin/promos', ['/admin/promos'], 'Справочники'),
  prefixItem('admin-antifraud', 'Антифрод', '/admin/antifraud', ['/admin/antifraud'], 'Контроль'),
  prefixItem('admin-amenities', 'Удобства', '/admin/amenities', ['/admin/amenities'], 'Справочники'),
  prefixItem('admin-object-types', 'Типы объектов', '/admin/object-types', ['/admin/object-types'], 'Справочники'),
  prefixItem('admin-holidays', 'Праздники', '/admin/holidays', ['/admin/holidays'], 'Справочники'),
  prefixItem('admin-wallets', 'Кошельки', '/admin/wallets', ['/admin/wallets'], 'Контроль'),
  prefixItem('admin-roles', 'Роли', '/admin/roles', ['/admin/roles'], 'Контроль'),
  prefixItem('admin-feature-flags', 'Флаги', '/admin/feature-flags', ['/admin/feature-flags'], 'Контроль'),
  prefixItem('admin-service-fees', 'Комиссии', '/admin/service-fees', ['/admin/service-fees'], 'Контроль'),
  prefixItem('admin-faq', 'FAQ', '/admin/faq', ['/admin/faq'], 'Контроль'),
  prefixItem('admin-profile', 'Профиль', '/admin/profile', ['/admin/profile'], 'Аккаунт'),
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
