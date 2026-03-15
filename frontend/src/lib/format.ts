import dayjs from 'dayjs'

const DAYS_OF_WEEK = [
  'Понедельник',
  'Вторник',
  'Среда',
  'Четверг',
  'Пятница',
  'Суббота',
  'Воскресенье',
] as const

const DAYS_OF_WEEK_SHORT = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'] as const

/**
 * Конвертирует копейки в строку с рублями.
 * 15000 -> "150 ₽", 15050 -> "150,50 ₽"
 */
export function formatPrice(kopecks: number): string {
  const rubles = kopecks / 100
  const formatted = rubles % 1 === 0 ? rubles.toString() : rubles.toFixed(2).replace('.', ',')
  return `${formatted} ₽`
}

/**
 * Название дня недели по индексу (0=Пн, 6=Вс).
 */
export function formatDayOfWeek(day: number, short = false): string {
  const list = short ? DAYS_OF_WEEK_SHORT : DAYS_OF_WEEK
  return list[day] ?? `День ${day}`
}

/**
 * Форматирует ISO-дату в человекочитаемый вид.
 */
export function formatDateTime(
  date: string | Date,
  format = 'DD.MM.YYYY HH:mm',
): string {
  return dayjs(date).format(format)
}

/**
 * Форматирует только время.
 */
export function formatTime(date: string | Date): string {
  return dayjs(date).format('HH:mm')
}
