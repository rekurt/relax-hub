export const AUTH_TOKEN_KEY = 'rh_token'

export const BOOKING_STATUS_CONFIG: Record<string, { color: string; text: string; hexColor: string }> = {
  pending: { color: 'orange', text: 'Ожидает', hexColor: '#d97706' },
  pending_owner: { color: 'orange', text: 'Ожидает подтверждения', hexColor: '#d97706' },
  confirmed: { color: 'blue', text: 'Подтверждено', hexColor: '#0f766e' },
  completed: { color: 'green', text: 'Завершено', hexColor: '#15803d' },
  cancelled: { color: 'default', text: 'Отменено', hexColor: '#c9c1b5' },
  rejected: { color: 'red', text: 'Отклонено', hexColor: '#b42318' },
  no_show: { color: 'volcano', text: 'Неявка', hexColor: '#d97706' },
  force_majeure_cancelled: { color: 'purple', text: 'Форс-мажор', hexColor: '#0a5f59' },
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
