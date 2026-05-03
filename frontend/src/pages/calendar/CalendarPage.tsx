import { useState, useMemo, useCallback } from 'react'
import {
  Alert,
  Badge,
  Button,
  Card,
  Col,
  DatePicker,
  Form,
  Input,
  List,
  Modal,
  Popconfirm,
  Row,
  Segmented,
  Select,
  Space,
  Spin,
  Switch,
  Tag,
  Tooltip,
  Typography,
} from 'antd'
import {
  CalendarOutlined,
  CopyOutlined,
  DeleteOutlined,
  ExportOutlined,
  LeftOutlined,
  LinkOutlined,
  PlusOutlined,
  ReloadOutlined,
  RightOutlined,
  SyncOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import type { Dayjs } from 'dayjs'
import isoWeek from 'dayjs/plugin/isoWeek'
import { App } from 'antd'
import { useQueryClient, useQueries } from '@tanstack/react-query'
import {
  useGetMyBathhousesIdCalendarToken,
  usePostMyBathhousesIdSlotBlocks,
  useGetMyBathhousesIdExternalCalendars,
  usePostMyBathhousesIdExternalCalendars,
  usePostMyBathhousesIdExternalCalendarsSync,
  useDeleteMyExternalCalendarsId,
} from '@/api/generated/calendar/calendar'
import {
  useGetBathhousesIdBookings,
  getBathhousesIdBookings,
} from '@/api/generated/bookings/bookings'
import { useGetMyBathhouses } from '@/api/generated/bathhouses/bathhouses'
import type {
  InternalHandlerBookingResponse,
  InternalHandlerExternalCalendarResponse,
} from '@/api/generated/model'
import { useBathhouseStore } from '@/stores/bathhouse'
import { formatPrice, formatTime, formatDateTime } from '@/lib/format'

dayjs.extend(isoWeek)

const { Title, Text } = Typography
const { RangePicker } = DatePicker

const HOURS = Array.from({ length: 24 }, (_, i) => i)
const DAYS_SHORT = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

const SOURCE_OPTIONS = [
  { value: 'google_calendar', label: 'Google Calendar' },
  { value: 'yandex_calendar', label: 'Яндекс.Календарь' },
]

type CalendarView = 'day' | 'week' | 'month'

const VIEW_OPTIONS = [
  { label: 'День', value: 'day' as CalendarView },
  { label: 'Неделя', value: 'week' as CalendarView },
  { label: 'Месяц', value: 'month' as CalendarView },
]

/** Status-based color coding per BRD: green=confirmed, yellow=pending, red=cancelled, gray=blocked */
const STATUS_COLORS: Record<string, { bg: string; text: string; label: string }> = {
  confirmed: { bg: '#52c41a', text: '#fff', label: 'Подтверждено' },
  pending: { bg: '#faad14', text: '#fff', label: 'Ожидает' },
  pending_owner: { bg: '#faad14', text: '#fff', label: 'Ожидает владельца' },
  cancelled: { bg: '#ff4d4f', text: '#fff', label: 'Отменено' },
  rejected: { bg: '#ff4d4f', text: '#fff', label: 'Отклонено' },
  completed: { bg: '#52c41a', text: '#fff', label: 'Завершено' },
  no_show: { bg: '#d9d9d9', text: '#333', label: 'Неявка' },
  force_majeure_cancelled: { bg: '#d9d9d9', text: '#333', label: 'Форс-мажор' },
}

const LEGEND_ITEMS = [
  { color: '#52c41a', label: 'Подтверждено' },
  { color: '#faad14', label: 'Ожидает' },
  { color: '#ff4d4f', label: 'Отменено' },
  { color: '#d9d9d9', label: 'Заблокировано' },
]

/** Color palette for multi-bathhouse consolidated view */
const BATHHOUSE_COLORS = [
  '#1677ff', '#722ed1', '#13c2c2', '#eb2f96', '#fa8c16',
  '#52c41a', '#2f54eb', '#faad14', '#a0d911', '#f5222d',
]

function getBathhouseColor(index: number): string {
  return BATHHOUSE_COLORS[index % BATHHOUSE_COLORS.length]!
}

function getStatusColor(status: string) {
  return STATUS_COLORS[status] ?? { bg: '#1677ff', text: '#fff', label: status }
}

interface BookingBlock {
  booking: InternalHandlerBookingResponse
  top: number
  height: number
  dayIndex: number
  bathhouseName?: string
  bathhouseColor?: string
}

function computeBookingBlocks(
  bookings: InternalHandlerBookingResponse[],
  days: Dayjs[],
  bookingBathhouseMap?: Map<string, string>,
  bathhouseColorMap?: Map<string, { name: string; color: string }>,
): BookingBlock[] {
  const blocks: BookingBlock[] = []
  for (const booking of bookings) {
    if (!booking.start_time || !booking.end_time) continue

    const start = dayjs(booking.start_time)
    const end = dayjs(booking.end_time)

    // Resolve bathhouse info for consolidated view
    const bathhouseId = booking.id && bookingBathhouseMap?.get(booking.id)
    const bathhouseInfo = bathhouseId ? bathhouseColorMap?.get(bathhouseId) : undefined

    for (let d = 0; d < days.length; d++) {
      const day = days[d]!
      const dayStart = day.startOf('day')
      const dayEnd = day.endOf('day')

      if (start.isBefore(dayEnd) && end.isAfter(dayStart)) {
        const effectiveStart = start.isAfter(dayStart) ? start : dayStart
        const effectiveEnd = end.isBefore(dayEnd) ? end : dayEnd

        const startMinutes = effectiveStart.hour() * 60 + effectiveStart.minute()
        const endMinutes = effectiveEnd.hour() * 60 + effectiveEnd.minute()

        const top = (startMinutes / 60) * 60
        const height = Math.max(((endMinutes - startMinutes) / 60) * 60, 20)

        blocks.push({
          booking,
          top,
          height,
          dayIndex: d,
          bathhouseName: bathhouseInfo?.name,
          bathhouseColor: bathhouseInfo?.color,
        })
      }
    }
  }
  return blocks
}

interface MonthCellData {
  date: Dayjs
  isCurrentMonth: boolean
  counts: Record<string, number>
  total: number
}

function computeMonthCells(
  monthStart: Dayjs,
  bookings: InternalHandlerBookingResponse[],
): MonthCellData[] {
  const firstDay = monthStart.startOf('month')
  const lastDay = monthStart.endOf('month')

  // Start from Monday of first week
  const calendarStart = firstDay.startOf('isoWeek')
  // End on Sunday of last week
  const calendarEnd = lastDay.endOf('isoWeek')

  const cells: MonthCellData[] = []
  let current = calendarStart

  while (current.isBefore(calendarEnd) || current.isSame(calendarEnd, 'day')) {
    const dayStart = current.startOf('day')
    const dayEnd = current.endOf('day')
    const counts: Record<string, number> = {}
    let total = 0

    for (const booking of bookings) {
      if (!booking.start_time) continue
      const bStart = dayjs(booking.start_time)
      const bEnd = booking.end_time ? dayjs(booking.end_time) : bStart

      if (bStart.isBefore(dayEnd) && bEnd.isAfter(dayStart)) {
        const status = booking.status ?? 'pending'
        counts[status] = (counts[status] ?? 0) + 1
        total++
      }
    }

    cells.push({
      date: current,
      isCurrentMonth: current.month() === monthStart.month(),
      counts,
      total,
    })

    current = current.add(1, 'day')
  }

  return cells
}

/** Renders a single booking block for day/week grids */
function BookingBlockEl({ block }: { block: BookingBlock }) {
  const status = block.booking.status ?? 'pending'
  const sc = getStatusColor(status)
  const bgColor = block.bathhouseColor ?? sc.bg
  return (
    <Tooltip
      title={
        <div>
          {block.bathhouseName && <div style={{ fontWeight: 600 }}>{block.bathhouseName}</div>}
          <div>{sc.label}</div>
          <div>
            {block.booking.start_time ? formatTime(block.booking.start_time) : ''} –{' '}
            {block.booking.end_time ? formatTime(block.booking.end_time) : ''}
          </div>
          <div>Гостей: {block.booking.guest_count ?? '—'}</div>
          <div>{formatPrice(block.booking.total_price ?? 0)}</div>
        </div>
      }
    >
      <div
        style={{
          position: 'absolute',
          top: block.top,
          left: 2,
          right: 2,
          height: block.height,
          backgroundColor: bgColor,
          opacity: 0.85,
          borderRadius: 4,
          padding: '2px 4px',
          overflow: 'hidden',
          cursor: 'pointer',
          color: '#fff',
          fontSize: 11,
          lineHeight: '14px',
        }}
      >
        <div style={{ fontWeight: 500 }}>
          {block.booking.start_time ? formatTime(block.booking.start_time) : ''}
        </div>
        {block.height > 30 && (
          <div>{block.bathhouseName ?? formatPrice(block.booking.total_price ?? 0)}</div>
        )}
      </div>
    </Tooltip>
  )
}

/** Time column shared by day and week views */
function TimeColumn() {
  return (
    <div style={{ width: 60, flexShrink: 0, borderRight: '1px solid #f0f0f0' }}>
      <div
        style={{
          height: 40,
          borderBottom: '1px solid #f0f0f0',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Text type="secondary" style={{ fontSize: 12 }}>
          Время
        </Text>
      </div>
      {HOURS.map((hour) => (
        <div
          key={hour}
          style={{
            height: 60,
            borderBottom: '1px solid #f5f5f5',
            display: 'flex',
            alignItems: 'flex-start',
            justifyContent: 'center',
            paddingTop: 2,
          }}
        >
          <Text type="secondary" style={{ fontSize: 11 }}>
            {String(hour).padStart(2, '0')}:00
          </Text>
        </div>
      ))}
    </div>
  )
}

/** Single day column with hourly grid and booking blocks */
function DayColumn({
  day,
  dayIndex,
  blocks,
  isLast,
  showDayHeader = true,
}: {
  day: Dayjs
  dayIndex: number
  blocks: BookingBlock[]
  isLast?: boolean
  showDayHeader?: boolean
}) {
  const isToday = day.isSame(dayjs(), 'day')
  const dayBlocks = blocks.filter((b) => b.dayIndex === dayIndex)

  return (
    <div
      style={{
        flex: 1,
        borderRight: isLast ? undefined : '1px solid #f0f0f0',
        minWidth: 100,
      }}
    >
      {showDayHeader && (
        <div
          style={{
            height: 40,
            borderBottom: '1px solid #f0f0f0',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            backgroundColor: isToday ? '#e6f4ff' : undefined,
          }}
        >
          <Text type="secondary" style={{ fontSize: 11, lineHeight: 1 }}>
            {DAYS_SHORT[day.isoWeekday() - 1]}
          </Text>
          <Text
            strong={isToday}
            style={{
              fontSize: 14,
              color: isToday ? '#1677ff' : undefined,
              lineHeight: 1.2,
            }}
          >
            {day.format('DD')}
          </Text>
        </div>
      )}

      <div style={{ position: 'relative' }}>
        {HOURS.map((hour) => (
          <div
            key={hour}
            style={{
              height: 60,
              borderBottom: '1px solid #f5f5f5',
              backgroundColor: isToday ? '#fafcff' : undefined,
            }}
          />
        ))}
        {dayBlocks.map((block, i) => (
          <BookingBlockEl key={`${block.booking.id}-${i}`} block={block} />
        ))}
      </div>
    </div>
  )
}

/** Day view - single day with hourly grid */
function DayView({
  date,
  bookings,
  bookingBathhouseMap,
  bathhouseColorMap,
}: {
  date: Dayjs
  bookings: InternalHandlerBookingResponse[]
  bookingBathhouseMap?: Map<string, string>
  bathhouseColorMap?: Map<string, { name: string; color: string }>
}) {
  const days = useMemo(() => [date], [date])
  const blocks = useMemo(
    () => computeBookingBlocks(bookings, days, bookingBathhouseMap, bathhouseColorMap),
    [bookings, days, bookingBathhouseMap, bathhouseColorMap],
  )

  return (
    <div style={{ display: 'flex', minWidth: 300 }}>
      <TimeColumn />
      <DayColumn day={date} dayIndex={0} blocks={blocks} isLast showDayHeader={false} />
    </div>
  )
}

/** Week view - 7 days with hourly grid */
function WeekView({
  weekStart,
  bookings,
  bookingBathhouseMap,
  bathhouseColorMap,
}: {
  weekStart: Dayjs
  bookings: InternalHandlerBookingResponse[]
  bookingBathhouseMap?: Map<string, string>
  bathhouseColorMap?: Map<string, { name: string; color: string }>
}) {
  const weekDays = useMemo(
    () => Array.from({ length: 7 }, (_, i) => weekStart.add(i, 'day')),
    [weekStart],
  )
  const blocks = useMemo(
    () => computeBookingBlocks(bookings, weekDays, bookingBathhouseMap, bathhouseColorMap),
    [bookings, weekDays, bookingBathhouseMap, bathhouseColorMap],
  )

  return (
    <div style={{ display: 'flex', minWidth: 800 }}>
      <TimeColumn />
      {weekDays.map((day, idx) => (
        <DayColumn
          key={idx}
          day={day}
          dayIndex={idx}
          blocks={blocks}
          isLast={idx === 6}
        />
      ))}
    </div>
  )
}

/** Month view - grid of day cells with booking counts */
function MonthView({
  monthStart,
  bookings,
  onDayClick,
}: {
  monthStart: Dayjs
  bookings: InternalHandlerBookingResponse[]
  onDayClick: (date: Dayjs) => void
}) {
  const cells = useMemo(() => computeMonthCells(monthStart, bookings), [monthStart, bookings])

  const weeks: MonthCellData[][] = []
  for (let i = 0; i < cells.length; i += 7) {
    weeks.push(cells.slice(i, i + 7))
  }

  return (
    <div>
      {/* Day of week headers */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', borderBottom: '1px solid #f0f0f0' }}>
        {DAYS_SHORT.map((d) => (
          <div
            key={d}
            style={{
              textAlign: 'center',
              padding: '8px 0',
              fontWeight: 500,
              fontSize: 13,
              color: '#666',
            }}
          >
            {d}
          </div>
        ))}
      </div>

      {/* Week rows */}
      {weeks.map((week, wi) => (
        <div
          key={wi}
          style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)' }}
        >
          {week.map((cell, ci) => {
            const isToday = cell.date.isSame(dayjs(), 'day')
            return (
              <div
                key={ci}
                onClick={() => onDayClick(cell.date)}
                style={{
                  minHeight: 90,
                  padding: 6,
                  border: '1px solid #f0f0f0',
                  borderTop: 'none',
                  borderLeft: ci === 0 ? '1px solid #f0f0f0' : 'none',
                  backgroundColor: isToday
                    ? '#e6f4ff'
                    : cell.isCurrentMonth
                      ? '#fff'
                      : '#fafafa',
                  cursor: 'pointer',
                  transition: 'background-color 0.2s',
                }}
                onMouseEnter={(e) => {
                  if (!isToday) e.currentTarget.style.backgroundColor = '#f5f5ff'
                }}
                onMouseLeave={(e) => {
                  if (!isToday)
                    e.currentTarget.style.backgroundColor = cell.isCurrentMonth ? '#fff' : '#fafafa'
                }}
              >
                <div
                  style={{
                    fontSize: 14,
                    fontWeight: isToday ? 700 : 400,
                    color: cell.isCurrentMonth
                      ? isToday
                        ? '#1677ff'
                        : '#333'
                      : '#bbb',
                    marginBottom: 4,
                  }}
                >
                  {cell.date.format('D')}
                </div>

                {cell.total > 0 && (
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: 2 }}>
                    {(cell.counts['confirmed'] ?? 0) > 0 && (
                      <Badge
                        count={cell.counts['confirmed']}
                        color="#52c41a"
                        size="small"
                        title="Подтверждено"
                      />
                    )}
                    {(cell.counts['completed'] ?? 0) > 0 && (
                      <Badge
                        count={cell.counts['completed']}
                        color="#52c41a"
                        size="small"
                        title="Завершено"
                      />
                    )}
                    {((cell.counts['pending'] ?? 0) + (cell.counts['pending_owner'] ?? 0)) > 0 && (
                      <Badge
                        count={(cell.counts['pending'] ?? 0) + (cell.counts['pending_owner'] ?? 0)}
                        color="#faad14"
                        size="small"
                        title="Ожидает"
                      />
                    )}
                    {((cell.counts['cancelled'] ?? 0) + (cell.counts['rejected'] ?? 0)) > 0 && (
                      <Badge
                        count={(cell.counts['cancelled'] ?? 0) + (cell.counts['rejected'] ?? 0)}
                        color="#ff4d4f"
                        size="small"
                        title="Отменено"
                      />
                    )}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      ))}
    </div>
  )
}

export default function CalendarPage() {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)

  const [view, setView] = useState<CalendarView>('week')
  const [currentDate, setCurrentDate] = useState<Dayjs>(() => dayjs())
  const [consolidatedView, setConsolidatedView] = useState(false)
  const [slotBlockModalOpen, setSlotBlockModalOpen] = useState(false)
  const [externalCalendarModalOpen, setExternalCalendarModalOpen] = useState(false)
  const [slotBlockForm] = Form.useForm()
  const [externalCalForm] = Form.useForm()

  // Derive week/month start from currentDate
  const weekStart = useMemo(() => currentDate.startOf('isoWeek'), [currentDate])
  const monthStart = useMemo(() => currentDate.startOf('month'), [currentDate])

  // Fetch owner's bathhouses for consolidated view
  const { data: bathhousesData } = useGetMyBathhouses(
    { page: 1, page_size: 100 },
    { query: { enabled: consolidatedView } },
  )
  const allBathhouses = useMemo(() => bathhousesData?.data ?? [], [bathhousesData?.data])
  const hasMultipleBathhouses = allBathhouses.length > 1

  // Build bathhouse color map for consolidated view
  const bathhouseColorMap = useMemo(() => {
    const map = new Map<string, { name: string; color: string }>()
    allBathhouses.forEach((b, i) => {
      if (b.id) map.set(b.id, { name: b.name ?? `Объект ${i + 1}`, color: getBathhouseColor(i) })
    })
    return map
  }, [allBathhouses])

  const { data: bookingsData, isLoading: bookingsLoading } = useGetBathhousesIdBookings(
    selectedBathhouseId ?? '',
    { page: 1, page_size: 200 },
    { query: { enabled: !!selectedBathhouseId && !consolidatedView } },
  )

  // Fetch bookings from all bathhouses in consolidated view
  const consolidatedQueries = useQueries({
    queries: consolidatedView
      ? allBathhouses
          .filter((b) => !!b.id)
          .map((b) => ({
            queryKey: [`/bathhouses/${b.id}/bookings`, { page: 1, page_size: 200 }] as const,
            queryFn: () => getBathhousesIdBookings(b.id!, { page: 1, page_size: 200 }),
            enabled: consolidatedView && allBathhouses.length > 0,
          }))
      : [],
  })
  const consolidatedLoading = consolidatedView && consolidatedQueries.some((q) => q.isLoading)

  const { data: tokenData } = useGetMyBathhousesIdCalendarToken(
    selectedBathhouseId ?? '',
    { query: { enabled: !!selectedBathhouseId } },
  )

  const { data: externalCalendarsData, isLoading: externalCalLoading } =
    useGetMyBathhousesIdExternalCalendars(selectedBathhouseId ?? '', {
      query: { enabled: !!selectedBathhouseId },
    })

  const invalidateCalendarData = () => {
    queryClient.invalidateQueries({
      queryKey: [`/bathhouses/${selectedBathhouseId}/bookings`],
    })
    queryClient.invalidateQueries({
      queryKey: [`/my/bathhouses/${selectedBathhouseId}/external-calendars`],
    })
  }

  const createSlotBlockMutation = usePostMyBathhousesIdSlotBlocks({
    mutation: {
      onSuccess: () => {
        message.success('Блокировка создана')
        setSlotBlockModalOpen(false)
        slotBlockForm.resetFields()
        invalidateCalendarData()
      },
      onError: () => message.error('Не удалось создать блокировку'),
    },
  })

  const addExternalCalMutation = usePostMyBathhousesIdExternalCalendars({
    mutation: {
      onSuccess: () => {
        message.success('Календарь добавлен')
        setExternalCalendarModalOpen(false)
        externalCalForm.resetFields()
        invalidateCalendarData()
      },
      onError: () => message.error('Не удалось добавить календарь'),
    },
  })

  const syncMutation = usePostMyBathhousesIdExternalCalendarsSync({
    mutation: {
      onSuccess: () => message.success('Синхронизация запущена'),
      onError: () => message.error('Не удалось запустить синхронизацию'),
    },
  })

  const removeExternalCalMutation = useDeleteMyExternalCalendarsId({
    mutation: {
      onSuccess: () => {
        message.success('Календарь удалён')
        invalidateCalendarData()
      },
      onError: () => message.error('Не удалось удалить календарь'),
    },
  })

  const externalCalendars = (externalCalendarsData?.data ?? []) as InternalHandlerExternalCalendarResponse[]
  const calendarToken = tokenData?.data

  const bookings = useMemo(() => {
    if (consolidatedView) {
      const merged: InternalHandlerBookingResponse[] = []
      consolidatedQueries.forEach((q) => {
        if (q.data?.data) {
          merged.push(...(q.data.data as InternalHandlerBookingResponse[]))
        }
      })
      return merged
    }
    return (bookingsData?.data ?? []) as InternalHandlerBookingResponse[]
  }, [consolidatedView, consolidatedQueries, bookingsData?.data])

  // Create booking-to-bathhouse mapping for consolidated view color coding
  const bookingBathhouseMap = useMemo(() => {
    if (!consolidatedView) return new Map<string, string>()
    const map = new Map<string, string>()
    consolidatedQueries.forEach((q, idx) => {
      const bathhouseId = allBathhouses[idx]?.id
      if (bathhouseId && q.data?.data) {
        for (const b of q.data.data as InternalHandlerBookingResponse[]) {
          if (b.id) map.set(b.id, bathhouseId)
        }
      }
    })
    return map
  }, [consolidatedView, consolidatedQueries, allBathhouses])

  const handleCreateSlotBlock = () => {
    slotBlockForm.validateFields().then((values) => {
      if (!selectedBathhouseId) return
      const [startDate, endDate] = values.dateRange
      createSlotBlockMutation.mutate({
        id: selectedBathhouseId,
        data: {
          start_time: startDate.toISOString(),
          end_time: endDate.toISOString(),
          description: values.description || '',
        },
      })
    })
  }

  const handleAddExternalCalendar = () => {
    externalCalForm.validateFields().then((values) => {
      if (!selectedBathhouseId) return
      addExternalCalMutation.mutate({
        id: selectedBathhouseId,
        data: {
          url: values.url,
          source: values.source,
        },
      })
    })
  }

  const handleCopyIcalUrl = () => {
    if (calendarToken?.url) {
      const fullUrl = `${window.location.origin}${calendarToken.url}`
      navigator.clipboard.writeText(fullUrl).then(
        () => message.success('Ссылка скопирована'),
        () => message.error('Не удалось скопировать ссылку'),
      )
    }
  }

  const navigate = useCallback(
    (direction: number) => {
      setCurrentDate((prev) => {
        switch (view) {
          case 'day':
            return prev.add(direction, 'day')
          case 'week':
            return prev.add(direction * 7, 'day')
          case 'month':
            return prev.add(direction, 'month')
        }
      })
    },
    [view],
  )

  const goToToday = useCallback(() => {
    setCurrentDate(dayjs())
  }, [])

  const handleMonthDayClick = useCallback((date: Dayjs) => {
    setCurrentDate(date)
    setView('day')
  }, [])

  const dateLabel = useMemo(() => {
    switch (view) {
      case 'day':
        return currentDate.format('DD MMMM YYYY (dd)')
      case 'week': {
        const end = weekStart.add(6, 'day')
        return `${weekStart.format('DD MMM')} — ${end.format('DD MMM YYYY')}`
      }
      case 'month':
        return currentDate.format('MMMM YYYY')
    }
  }, [view, currentDate, weekStart])

  if (!selectedBathhouseId) {
    return (
      <div>
        <Title level={3}>Календарь</Title>
        <Alert
          title="Выберите баню"
          description="Для просмотра календаря выберите баню в верхнем меню."
          type="info"
          showIcon
        />
      </div>
    )
  }

  return (
    <div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 16,
          flexWrap: 'wrap',
          gap: 12,
        }}
      >
        <Title level={3} style={{ margin: 0 }}>
          Календарь
        </Title>
        <Space wrap>
          {hasMultipleBathhouses && (
            <Tooltip title="Показать бронирования всех объектов">
              <Space>
                <Text style={{ fontSize: 13 }}>Все объекты</Text>
                <Switch
                  size="small"
                  checked={consolidatedView}
                  onChange={setConsolidatedView}
                />
              </Space>
            </Tooltip>
          )}
          <Button
            icon={<PlusOutlined />}
            onClick={() => setSlotBlockModalOpen(true)}
          >
            Блокировка
          </Button>
          <Button icon={<LinkOutlined />} onClick={() => setExternalCalendarModalOpen(true)}>
            Внешний календарь
          </Button>
        </Space>
      </div>

      {/* Navigation + View Switcher */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            flexWrap: 'wrap',
            gap: 8,
          }}
        >
          <Space>
            <Button icon={<LeftOutlined />} onClick={() => navigate(-1)} />
            <Button onClick={goToToday}>Сегодня</Button>
            <Button icon={<RightOutlined />} onClick={() => navigate(1)} />
          </Space>
          <Text strong>{dateLabel}</Text>
          <Space>
            <Segmented
              options={VIEW_OPTIONS}
              value={view}
              onChange={(v) => setView(v as CalendarView)}
            />
          </Space>
        </div>
        {/* Legend */}
        <div style={{ display: 'flex', gap: 12, marginTop: 8, flexWrap: 'wrap' }}>
          {consolidatedView ? (
            // Show bathhouse color legend in consolidated mode
            Array.from(bathhouseColorMap.entries()).map(([id, info]) => (
              <Space key={id} size={4}>
                <div
                  style={{
                    width: 12,
                    height: 12,
                    borderRadius: 2,
                    backgroundColor: info.color,
                  }}
                />
                <Text style={{ fontSize: 12 }}>{info.name}</Text>
              </Space>
            ))
          ) : (
            LEGEND_ITEMS.map((item) => (
              <Space key={item.label} size={4}>
                <div
                  style={{
                    width: 12,
                    height: 12,
                    borderRadius: 2,
                    backgroundColor: item.color,
                  }}
                />
                <Text style={{ fontSize: 12 }}>{item.label}</Text>
              </Space>
            ))
          )}
        </div>
      </Card>

      {/* Calendar Grid */}
      <Card
        size="small"
        style={{ marginBottom: 16, overflow: 'auto' }}
        styles={{ body: { padding: 0 } }}
      >
        {(bookingsLoading || consolidatedLoading) ? (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <Spin size="large" />
          </div>
        ) : view === 'day' ? (
          <DayView
            date={currentDate}
            bookings={bookings}
            bookingBathhouseMap={consolidatedView ? bookingBathhouseMap : undefined}
            bathhouseColorMap={consolidatedView ? bathhouseColorMap : undefined}
          />
        ) : view === 'week' ? (
          <WeekView
            weekStart={weekStart}
            bookings={bookings}
            bookingBathhouseMap={consolidatedView ? bookingBathhouseMap : undefined}
            bathhouseColorMap={consolidatedView ? bathhouseColorMap : undefined}
          />
        ) : (
          <MonthView
            monthStart={monthStart}
            bookings={bookings}
            onDayClick={handleMonthDayClick}
          />
        )}
      </Card>

      {/* iCal Export & External Calendars */}
      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <Card
            title={
              <Space>
                <ExportOutlined />
                <span>Экспорт iCal</span>
              </Space>
            }
            size="small"
          >
            <Text type="secondary" style={{ display: 'block', marginBottom: 12 }}>
              Используйте ссылку для подключения календаря бронирований в Google Calendar,
              Apple Calendar или другие приложения.
            </Text>
            {calendarToken?.url ? (
              <Space orientation="vertical" style={{ width: '100%' }}>
                <Input
                  readOnly
                  value={`${window.location.origin}${calendarToken.url}`}
                  addonAfter={
                    <CopyOutlined
                      onClick={handleCopyIcalUrl}
                      style={{ cursor: 'pointer' }}
                    />
                  }
                />
              </Space>
            ) : (
              <Text type="secondary">Загрузка...</Text>
            )}
          </Card>
        </Col>

        <Col xs={24} md={12}>
          <Card
            title={
              <Space>
                <SyncOutlined />
                <span>Внешние календари</span>
              </Space>
            }
            size="small"
            extra={
              <Space>
                <Button
                  size="small"
                  icon={<ReloadOutlined />}
                  loading={syncMutation.isPending}
                  onClick={() =>
                    selectedBathhouseId &&
                    syncMutation.mutate({ id: selectedBathhouseId })
                  }
                >
                  Синхронизировать
                </Button>
                <Button
                  size="small"
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={() => setExternalCalendarModalOpen(true)}
                >
                  Добавить
                </Button>
              </Space>
            }
          >
            {externalCalLoading ? (
              <Spin size="small" />
            ) : externalCalendars.length === 0 ? (
              <Text type="secondary">
                Нет подключённых календарей. Добавьте Google Calendar или Яндекс.Календарь
                для автоматической блокировки занятых слотов.
              </Text>
            ) : (
              <List
                size="small"
                dataSource={externalCalendars}
                renderItem={(cal) => (
                  <List.Item
                    actions={[
                      <Popconfirm
                        key="delete"
                        title="Удалить календарь?"
                        onConfirm={() =>
                          cal.id && removeExternalCalMutation.mutate({ id: cal.id })
                        }
                      >
                        <Button
                          size="small"
                          danger
                          icon={<DeleteOutlined />}
                          loading={removeExternalCalMutation.isPending && removeExternalCalMutation.variables?.id === cal.id}
                        />
                      </Popconfirm>,
                    ]}
                  >
                    <List.Item.Meta
                      avatar={<CalendarOutlined />}
                      title={
                        <Space>
                          <Badge
                            status={cal.last_error ? 'error' : cal.last_sync_at ? 'success' : 'default'}
                            title={cal.last_error ? 'Ошибка синхронизации' : cal.last_sync_at ? 'Синхронизировано' : 'Не синхронизировано'}
                          />
                          <Tag>
                            {cal.source === 'google_calendar'
                              ? 'Google'
                              : 'Яндекс'}
                          </Tag>
                          <Text
                            ellipsis
                            style={{ maxWidth: 200 }}
                            title={cal.url}
                          >
                            {cal.url}
                          </Text>
                        </Space>
                      }
                      description={
                        <Space>
                          {cal.last_sync_at && (
                            <Text type="secondary" style={{ fontSize: 12 }}>
                              Синхр.: {formatDateTime(cal.last_sync_at)}
                            </Text>
                          )}
                          {cal.last_error && (
                            <Text type="danger" style={{ fontSize: 12 }}>
                              Ошибка: {cal.last_error}
                            </Text>
                          )}
                        </Space>
                      }
                    />
                  </List.Item>
                )}
              />
            )}
          </Card>
        </Col>
      </Row>

      {/* Create Slot Block Modal */}
      <Modal
        title="Создать блокировку"
        open={slotBlockModalOpen}
        onCancel={() => {
          setSlotBlockModalOpen(false)
          slotBlockForm.resetFields()
        }}
        onOk={handleCreateSlotBlock}
        confirmLoading={createSlotBlockMutation.isPending}
        okText="Создать"
        cancelText="Отмена"
      >
        <Form form={slotBlockForm} layout="vertical">
          <Form.Item
            name="dateRange"
            label="Период блокировки"
            rules={[{ required: true, message: 'Выберите период' }]}
          >
            <RangePicker
              showTime={{ format: 'HH:mm' }}
              format="DD.MM.YYYY HH:mm"
              style={{ width: '100%' }}
              placeholder={['Начало', 'Конец']}
            />
          </Form.Item>
          <Form.Item name="description" label="Описание (необязательно)">
            <Input.TextArea rows={2} placeholder="Например: Техобслуживание" />
          </Form.Item>
        </Form>
        <Alert
          title="Блокировка запретит бронирование на указанный период."
          type="info"
          showIcon
          style={{ marginTop: 8 }}
        />
      </Modal>

      {/* Add External Calendar Modal */}
      <Modal
        title="Добавить внешний календарь"
        open={externalCalendarModalOpen}
        onCancel={() => {
          setExternalCalendarModalOpen(false)
          externalCalForm.resetFields()
        }}
        onOk={handleAddExternalCalendar}
        confirmLoading={addExternalCalMutation.isPending}
        okText="Добавить"
        cancelText="Отмена"
      >
        <Form form={externalCalForm} layout="vertical">
          <Form.Item
            name="source"
            label="Источник"
            rules={[{ required: true, message: 'Выберите источник' }]}
          >
            <Select options={SOURCE_OPTIONS} placeholder="Выберите тип календаря" />
          </Form.Item>
          <Form.Item
            name="url"
            label="URL календаря (iCal)"
            rules={[
              { required: true, message: 'Введите URL' },
              { type: 'url', message: 'Введите корректный URL' },
            ]}
          >
            <Input placeholder="https://calendar.google.com/calendar/ical/.../basic.ics" />
          </Form.Item>
        </Form>
        <Alert
          title="Занятые слоты из внешнего календаря будут автоматически блокировать бронирования."
          type="info"
          showIcon
          style={{ marginTop: 8 }}
        />
      </Modal>
    </div>
  )
}
