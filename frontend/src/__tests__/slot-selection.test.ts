import { describe, expect, it } from 'vitest'
import {
  buildSlotRangeBetween,
  getRangeHours,
  resolveSlotRangeSelection,
} from '@/lib/slot-selection'

const slots = [
  { startTime: '2026-04-20T10:00:00', endTime: '2026-04-20T11:00:00', available: true },
  { startTime: '2026-04-20T11:00:00', endTime: '2026-04-20T12:00:00', available: true },
  { startTime: '2026-04-20T12:00:00', endTime: '2026-04-20T13:00:00', available: false },
  { startTime: '2026-04-20T13:00:00', endTime: '2026-04-20T14:00:00', available: true },
]

describe('slot-selection', () => {
  it('собирает диапазон между стартовым и конечным слотом двумя кликами', () => {
    const selection = buildSlotRangeBetween(slots, '2026-04-20T10:00:00', '2026-04-20T11:00:00')

    expect(selection).toEqual({
      from: '2026-04-20T10:00:00',
      to: '2026-04-20T12:00:00',
    })
    expect(getRangeHours(selection?.from, selection?.to)).toBe(2)
  })

  it('не строит диапазон через недоступный разрыв', () => {
    expect(buildSlotRangeBetween(slots, '2026-04-20T10:00:00', '2026-04-20T13:00:00')).toBeNull()
  })

  it('валидирует уже выбранный from/to только для реального непрерывного интервала', () => {
    expect(resolveSlotRangeSelection(slots, {
      from: '2026-04-20T10:00:00',
      to: '2026-04-20T12:00:00',
    })).toEqual({
      from: '2026-04-20T10:00:00',
      to: '2026-04-20T12:00:00',
    })

    expect(resolveSlotRangeSelection(slots, {
      from: '2026-04-20T10:00:00',
      to: '2026-04-20T14:00:00',
    })).toBeNull()
  })
})
