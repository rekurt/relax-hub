export const AUTH_TOKEN_KEY = 'bani_token'

export const BOOKING_STATUS_CONFIG: Record<string, { color: string; text: string; hexColor: string }> = {
  pending: { color: 'orange', text: 'Ожидает', hexColor: '#faad14' },
  pending_owner: { color: 'orange', text: 'Ожидает подтверждения', hexColor: '#faad14' },
  confirmed: { color: 'blue', text: 'Подтверждено', hexColor: '#1677ff' },
  completed: { color: 'green', text: 'Завершено', hexColor: '#52c41a' },
  cancelled: { color: 'default', text: 'Отменено', hexColor: '#d9d9d9' },
  rejected: { color: 'red', text: 'Отклонено', hexColor: '#ff4d4f' },
  no_show: { color: 'volcano', text: 'Неявка', hexColor: '#fa541c' },
  force_majeure_cancelled: { color: 'purple', text: 'Форс-мажор', hexColor: '#722ed1' },
}

export const PAYMENT_STATUS_CONFIG: Record<string, { color: string; text: string }> = {
  pending: { color: 'orange', text: 'Ожидает оплаты' },
  succeeded: { color: 'green', text: 'Оплачено' },
  canceled: { color: 'default', text: 'Отменён' },
  refunded: { color: 'purple', text: 'Возвращён' },
}

export const NOTIFICATION_TYPE_LABELS: Record<string, string> = {
  booking_new: 'Новое бронирование',
  booking_confirmed: 'Бронирование подтверждено',
  booking_cancelled: 'Бронирование отменено',
  booking_completed: 'Бронирование завершено',
  review_new: 'Новый отзыв',
  payment_received: 'Оплата получена',
  promo_used: 'Промокод использован',
  chat_message: 'Новое сообщение',
}

export const PROVIDER_LABELS: Record<string, string> = {
  vk: 'ВКонтакте',
  yandex: 'Яндекс',
  google: 'Google',
}

export const PROVIDER_COLORS: Record<string, string> = {
  vk: '#4C75A3',
  yandex: '#FC3F1D',
  google: '#4285F4',
}
