import { formatPrice, formatDayOfWeek, formatDateTime, formatTime } from '../lib/format'

describe('formatPrice', () => {
  it('converts kopecks to rubles without decimals', () => {
    expect(formatPrice(15000)).toBe('150 ₽')
  })

  it('converts kopecks to rubles with decimals', () => {
    expect(formatPrice(15050)).toBe('150,50 ₽')
  })

  it('handles zero', () => {
    expect(formatPrice(0)).toBe('0 ₽')
  })

  it('handles single kopeck', () => {
    expect(formatPrice(1)).toBe('0,01 ₽')
  })

  it('handles large amounts', () => {
    expect(formatPrice(1000000)).toBe('10000 ₽')
  })
})

describe('formatDayOfWeek', () => {
  it('returns Monday for 0', () => {
    expect(formatDayOfWeek(0)).toBe('Понедельник')
  })

  it('returns Sunday for 6', () => {
    expect(formatDayOfWeek(6)).toBe('Воскресенье')
  })

  it('returns short form', () => {
    expect(formatDayOfWeek(0, true)).toBe('Пн')
    expect(formatDayOfWeek(4, true)).toBe('Пт')
  })

  it('handles out of range gracefully', () => {
    expect(formatDayOfWeek(7)).toBe('День 7')
    expect(formatDayOfWeek(-1)).toBe('День -1')
  })
})

describe('formatDateTime', () => {
  it('formats date string without timezone', () => {
    expect(formatDateTime('2025-03-15 14:30:00')).toBe('15.03.2025 14:30')
  })

  it('accepts custom format', () => {
    expect(formatDateTime('2025-03-15 14:30:00', 'YYYY-MM-DD')).toBe('2025-03-15')
  })
})

describe('formatTime', () => {
  it('formats time only', () => {
    expect(formatTime('2025-03-15 14:30:00')).toBe('14:30')
  })
})
