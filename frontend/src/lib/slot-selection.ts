import dayjs from 'dayjs'

export interface SlotSelectionItem {
  startTime?: string
  endTime?: string
  available?: boolean
  price?: number | null
}

export interface SlotRangeSelection {
  from: string
  to: string
}

function toTimestamp(value?: string) {
  if (!value) return null
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.valueOf() : null
}

function sortSlots<T extends SlotSelectionItem>(slots: T[]) {
  return [...slots].sort((left, right) => {
    const leftTime = toTimestamp(left.startTime) ?? 0
    const rightTime = toTimestamp(right.startTime) ?? 0
    return leftTime - rightTime
  })
}

function areConsecutive(left: SlotSelectionItem, right: SlotSelectionItem) {
  const leftEnd = toTimestamp(left.endTime)
  const rightStart = toTimestamp(right.startTime)
  return leftEnd != null && rightStart != null && leftEnd === rightStart
}

function getAvailableSlots<T extends SlotSelectionItem>(slots: T[]) {
  return sortSlots(
    slots.filter((slot): slot is T & { startTime: string; endTime: string } => {
      return Boolean(slot.available && slot.startTime && slot.endTime)
    }),
  )
}

function buildSelection<T extends SlotSelectionItem>(
  slots: Array<T & { startTime: string; endTime: string }>,
  startIndex: number,
  endIndex: number,
): SlotRangeSelection | null {
  const startSlot = slots[startIndex]
  const endSlot = slots[endIndex]

  if (!startSlot?.startTime || !endSlot?.endTime) return null

  return {
    from: startSlot.startTime,
    to: endSlot.endTime,
  }
}

function findSelectionBounds<T extends SlotSelectionItem>(
  slots: Array<T & { startTime: string; endTime: string }>,
  selection: SlotRangeSelection | null | undefined,
) {
  if (!selection) return null

  const startIndex = slots.findIndex((slot) => slot.startTime === selection.from)
  if (startIndex < 0) return null

  let endIndex = startIndex

  while (endIndex < slots.length) {
    const current = slots[endIndex]
    if (!current) return null
    if (current.endTime === selection.to) {
      return { startIndex, endIndex }
    }

    const next = slots[endIndex + 1]
    if (!next || !areConsecutive(current, next)) return null
    endIndex += 1
  }

  return null
}

export function resolveSlotRangeSelection<T extends SlotSelectionItem>(
  slots: T[],
  selection: SlotRangeSelection | null | undefined,
) {
  const availableSlots = getAvailableSlots(slots)
  if (!selection) return null

  const bounds = findSelectionBounds(availableSlots, selection)
  if (!bounds) return null

  return buildSelection(availableSlots, bounds.startIndex, bounds.endIndex)
}

export function buildSlotRangeBetween<T extends SlotSelectionItem>(
  slots: T[],
  startTime: string | null | undefined,
  endTime: string | null | undefined,
) {
  const availableSlots = getAvailableSlots(slots)
  if (!startTime || !endTime) {
    return null
  }

  const startIndex = availableSlots.findIndex((slot) => slot.startTime === startTime)
  const endIndex = availableSlots.findIndex((slot) => slot.startTime === endTime)
  if (startIndex < 0 || endIndex < 0) {
    return null
  }

  const fromIndex = Math.min(startIndex, endIndex)
  const toIndex = Math.max(startIndex, endIndex)

  for (let index = fromIndex; index < toIndex; index += 1) {
    const current = availableSlots[index]
    const next = availableSlots[index + 1]
    if (!current || !next || !areConsecutive(current, next)) {
      return null
    }
  }

  return buildSelection(availableSlots, fromIndex, toIndex)
}

export function isSlotWithinRange(
  slot: Pick<SlotSelectionItem, 'startTime' | 'endTime'>,
  selection: SlotRangeSelection | null | undefined,
) {
  const slotStart = toTimestamp(slot.startTime)
  const slotEnd = toTimestamp(slot.endTime)
  const rangeStart = toTimestamp(selection?.from)
  const rangeEnd = toTimestamp(selection?.to)

  if (slotStart == null || slotEnd == null || rangeStart == null || rangeEnd == null) {
    return false
  }

  return slotStart >= rangeStart && slotEnd <= rangeEnd
}

export function getRangeHours(from?: string | null, to?: string | null) {
  const fromTs = toTimestamp(from ?? undefined)
  const toTs = toTimestamp(to ?? undefined)

  if (fromTs == null || toTs == null || toTs <= fromTs) return 1

  const hours = (toTs - fromTs) / (60 * 60 * 1000)
  return Number.isFinite(hours) && hours >= 1 ? Math.round(hours) : 1
}

export function formatSlotTimeLabel(value?: string | null) {
  if (!value) return ''
  return dayjs(value).format('HH:mm')
}
