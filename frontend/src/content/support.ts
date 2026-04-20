export const PLATFORM_NAME = 'RelaxHUB'

export const PLATFORM_CONTACTS = {
  supportEmail: 'support@relaxhub.ru',
  supportPhone: '+7 (495) 555-21-21',
  supportMessenger: '@relaxhub_reserve',
  supportHours: 'Ежедневно с 09:00 до 22:00 по Москве',
} as const

export const PLATFORM_BRAND_STATEMENT = 'Платформа для аккуратного выбора и бронирования приватных бань без лишнего шума в интерфейсе.'

export const TERMS_LAST_UPDATED = '20 апреля 2026'

export const CONTACT_CARDS = [
  {
    key: 'support',
    title: 'Поддержка бронирований',
    description: 'Помогаем подобрать баню, разобраться с оплатой и быстро перевести заявку в рабочее состояние.',
    meta: PLATFORM_CONTACTS.supportEmail,
  },
  {
    key: 'sales',
    title: 'Для владельцев бань',
    description: 'Если вы подключаете объект, продвижение или хотите demo-показ кабинета владельца, это основной канал.',
    meta: PLATFORM_CONTACTS.supportPhone,
  },
  {
    key: 'messengers',
    title: 'Быстрый канал',
    description: 'Для уточнений по текущему бронированию и навигации по сервису удобнее всего писать в мессенджер.',
    meta: PLATFORM_CONTACTS.supportMessenger,
  },
] as const
