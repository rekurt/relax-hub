import { ClockCircleOutlined } from '@/components/design/icons'
import { Alert, Button, Typography } from '@/components/design/system'
import { useMemo, useState } from 'react'
import dayjs from 'dayjs'
import {
  buildSlotRangeBetween,
  formatSlotTimeLabel,
  isSlotWithinRange,
  resolveSlotRangeSelection,
  type SlotRangeSelection,
  type SlotSelectionItem,
} from '@/lib/slot-selection'

const { Text } = Typography

interface ContiguousSlotSelectorProps<T extends SlotSelectionItem> {
  slots: T[]
  value: SlotRangeSelection | null
  onChange: (value: SlotRangeSelection | null) => void
  minDurationHours?: number
  label?: React.ReactNode
  description?: React.ReactNode
  size?: 'small' | 'middle' | 'large'
}

export default function ContiguousSlotSelector<T extends SlotSelectionItem>({
  slots,
  value,
  onChange,
  minDurationHours = 1,
  label,
  description,
  size = 'large',
}: ContiguousSlotSelectorProps<T>) {
  const [pendingStartTime, setPendingStartTime] = useState<string | null>(null)

  const sortedSlots = useMemo(() => {
    return [...slots]
      .filter((slot): slot is T & { startTime: string; endTime: string } => Boolean(slot.startTime && slot.endTime))
      .sort((left, right) => dayjs(left.startTime).valueOf() - dayjs(right.startTime).valueOf())
  }, [slots])

  const resolvedValue = useMemo(() => resolveSlotRangeSelection(slots, value), [slots, value])
  const activePendingStartTime = resolvedValue ? null : pendingStartTime
  const pendingStartLabel = formatSlotTimeLabel(activePendingStartTime)

  return (
    <div>
      {label ? (
        <Text type="secondary">
          {label}
        </Text>
      ) : null}

      {description ? (
        <div style={{ marginTop: 8 }}>
          <Text type="secondary">{description}</Text>
        </div>
      ) : null}

      <div className="rh-slot-grid" style={{ marginTop: 12 }}>
        {sortedSlots.map((slot) => {
          const isDisabled = !slot.available
          const isSelected = !isDisabled && (
            isSlotWithinRange(slot, resolvedValue)
            || (activePendingStartTime != null && slot.startTime === activePendingStartTime)
          )

          return (
            <Button
              key={`${slot.startTime}-${slot.endTime}`}
              type={isSelected ? 'primary' : 'default'}
              size={size}
              disabled={isDisabled}
              icon={<ClockCircleOutlined />}
              onClick={() => {
                if (!slot.startTime) return

                if (resolvedValue) {
                  setPendingStartTime(slot.startTime)
                  onChange(null)
                  return
                }

                if (!pendingStartTime) {
                  setPendingStartTime(slot.startTime)
                  return
                }

                const nextRange = buildSlotRangeBetween(slots, pendingStartTime, slot.startTime)
                if (nextRange) {
                  setPendingStartTime(null)
                  onChange(nextRange)
                  return
                }

                setPendingStartTime(slot.startTime)
                onChange(null)
              }}
            >
              {formatSlotTimeLabel(slot.startTime)}
            </Button>
          )
        })}
      </div>

      {activePendingStartTime ? (
        <Alert
          style={{ marginTop: 12 }}
          type="info"
          showIcon={false}
          title={`Старт: ${pendingStartLabel}. Выберите конечный слот${minDurationHours > 1 ? ` для интервала от ${minDurationHours} ч` : ''}.`}
        />
      ) : null}
    </div>
  )
}
