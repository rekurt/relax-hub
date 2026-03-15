export const AUTH_TOKEN_KEY = 'bani_token'

export const BOOKING_STATUS_CONFIG: Record<string, { color: string; text: string; hexColor: string }> = {
  pending: { color: 'orange', text: 'Ожидает', hexColor: '#faad14' },
  confirmed: { color: 'blue', text: 'Подтверждено', hexColor: '#1677ff' },
  completed: { color: 'green', text: 'Завершено', hexColor: '#52c41a' },
  cancelled: { color: 'default', text: 'Отменено', hexColor: '#d9d9d9' },
  rejected: { color: 'red', text: 'Отклонено', hexColor: '#ff4d4f' },
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
