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
} from '@/components/design/system'
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
} from '@/components/design/icons'
import dayjs from 'dayjs'
import type { Dayjs } from 'dayjs'
import isoWeek from 'dayjs/plugin/isoWeek'
import { App } from '@/components/design/system'
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
import PageHeader from '@/components/PageHeader'

dayjs.extend(isoWeek)

const { Text } = Typography
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
const STATUS_META: Record<string, { tone: string; label: string }> = {
  confirmed: { tone: 'success', label: 'Подтверждено' },
  pending: { tone: 'warning', label: 'Ожидает' },
  pending_owner: { tone: 'warning', label: 'Ожидает владельца' },
  cancelled: { tone: 'danger', label: 'Отменено' },
  rejected: { tone: 'danger', label: 'Отклонено' },
  completed: { tone: 'success', label: 'Завершено' },
  no_show: { tone: 'muted', label: 'Неявка' },
  force_majeure_cancelled: { tone: 'muted', label: 'Форс-мажор' },
}

const LEGEND_ITEMS = [
  { tone: 'success', label: 'Подтверждено' },
  { tone: 'warning', label: 'Ожидает' },
  { tone: 'danger', label: 'Отменено' },
  { tone: 'muted', label: 'Заблокировано' },
]

/** Color palette for multi-bathhouse consolidated view */
const BATHHOUSE_TONES = ['primary', 'teal', 'primary', 'warning', 'warning', 'success', 'primary', 'warning', 'success', 'danger']

function getBathhouseTone(index: number): string {
  return BATHHOUSE_TONES[index % BATHHOUSE_TONES.length]!
}

function getStatusColor(status: string) {
  return STATUS_META[status] ?? { tone: 'primary', label: status }
}

function useBookingBlockLayout(top: number, height: number) {
  return useCallback((node: HTMLDivElement | null) => {
    if (!node) return
    node.style.top = `${top}px`
    node.style.height = `${height}px`
  }, [top, height])
}

interface BookingBlock {
  booking: InternalHandlerBookingResponse
  top: number
  height: number
  dayIndex: number
  bathhouseName?: string
  bathhouseTone?: string
}

function computeBookingBlocks(
  bookings: InternalHandlerBookingResponse[],
  days: Dayjs[],
  bookingBathhouseMap?: Map<string, string>,
  bathhouseColorMap?: Map<string, { name: string; tone: string }>,
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
          bathhouseTone: bathhouseInfo?.tone,
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
  const tone = block.bathhouseTone ?? sc.tone
  const blockRef = useBookingBlockLayout(block.top, block.height)
  return (
    <Tooltip
      title={
        <div className="rh-calendar-tooltip">
          {block.bathhouseName && <div className="rh-calendar-tooltip__title">{block.bathhouseName}</div>}
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
        ref={blockRef}
        className={`rh-calendar-booking-block rh-calendar-tone--${tone}`}
      >
        <div className="rh-calendar-booking-block__time">
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
    <div className="rh-calendar-time-column">
      <div className="rh-calendar-time-column__header">
        <Text type="secondary" className="rh-calendar-time-column__label">
          Время
        </Text>
      </div>
      {HOURS.map((hour) => (
        <div
          key={hour}
          className="rh-calendar-hour-label"
        >
          <Text type="secondary" className="rh-calendar-hour-label__text">
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
      className={isLast ? 'rh-calendar-day-column rh-calendar-day-column--last' : 'rh-calendar-day-column'}
    >
      {showDayHeader && (
        <div
          className={isToday ? 'rh-calendar-day-header rh-calendar-day-header--today' : 'rh-calendar-day-header'}
        >
          <Text type="secondary" className="rh-calendar-day-header__weekday">
            {DAYS_SHORT[day.isoWeekday() - 1]}
          </Text>
          <Text
            strong={isToday}
            className={isToday ? 'rh-calendar-day-header__date rh-calendar-day-header__date--today' : 'rh-calendar-day-header__date'}
          >
            {day.format('DD')}
          </Text>
        </div>
      )}

      <div className="rh-calendar-day-body">
        {HOURS.map((hour) => (
          <div
            key={hour}
            className={isToday ? 'rh-calendar-hour-cell rh-calendar-hour-cell--today' : 'rh-calendar-hour-cell'}
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
  bathhouseColorMap?: Map<string, { name: string; tone: string }>
}) {
  const days = useMemo(() => [date], [date])
  const blocks = useMemo(
    () => computeBookingBlocks(bookings, days, bookingBathhouseMap, bathhouseColorMap),
    [bookings, days, bookingBathhouseMap, bathhouseColorMap],
  )

  return (
    <div className="rh-calendar-day-view">
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
  bathhouseColorMap?: Map<string, { name: string; tone: string }>
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
    <div className="rh-calendar-week-view">
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
    <div className="rh-calendar-month">
      {/* Day of week headers */}
      <div className="rh-calendar-month__header">
        {DAYS_SHORT.map((d) => (
          <div
            key={d}
            className="rh-calendar-month__weekday"
          >
            {d}
          </div>
        ))}
      </div>

      {/* Week rows */}
      {weeks.map((week, wi) => (
        <div
          key={wi}
          className="rh-calendar-month__week"
        >
          {week.map((cell, ci) => {
            const isToday = cell.date.isSame(dayjs(), 'day')
            const cellClassName = [
              'rh-calendar-month__cell',
              ci === 0 ? 'rh-calendar-month__cell--week-start' : '',
              isToday ? 'rh-calendar-month__cell--today' : '',
              cell.isCurrentMonth ? '' : 'rh-calendar-month__cell--muted',
            ].filter(Boolean).join(' ')
            return (
              <div
                key={ci}
                onClick={() => onDayClick(cell.date)}
                className={cellClassName}
              >
                <div
                  className={isToday ? 'rh-calendar-month__date rh-calendar-month__date--today' : 'rh-calendar-month__date'}
                >
                  {cell.date.format('D')}
                </div>

                {cell.total > 0 && (
                  <div className="rh-calendar-month__badges">
                    {(cell.counts['confirmed'] ?? 0) > 0 && (
                      <Badge
                        count={cell.counts['confirmed']}
                        color="#15803d"
                        size="small"
                        title="Подтверждено"
                      />
                    )}
                    {(cell.counts['completed'] ?? 0) > 0 && (
                      <Badge
                        count={cell.counts['completed']}
                        color="#15803d"
                        size="small"
                        title="Завершено"
                      />
                    )}
                    {((cell.counts['pending'] ?? 0) + (cell.counts['pending_owner'] ?? 0)) > 0 && (
                      <Badge
                        count={(cell.counts['pending'] ?? 0) + (cell.counts['pending_owner'] ?? 0)}
                        color="#d97706"
                        size="small"
                        title="Ожидает"
                      />
                    )}
                    {((cell.counts['cancelled'] ?? 0) + (cell.counts['rejected'] ?? 0)) > 0 && (
                      <Badge
                        count={(cell.counts['cancelled'] ?? 0) + (cell.counts['rejected'] ?? 0)}
                        color="#b42318"
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
    const map = new Map<string, { name: string; tone: string }>()
    allBathhouses.forEach((b, i) => {
      if (b.id) map.set(b.id, { name: b.name ?? `Объект ${i + 1}`, tone: getBathhouseTone(i) })
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
      <div className="rh-stack">
        <PageHeader
          eyebrow="Расписание"
          title="Календарь"
          description="Выберите объект, чтобы открыть сетку бронирований и синхронизации."
          size="compact"
        />
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
    <div className="rh-stack">
      <PageHeader
        eyebrow="Расписание"
        title="Календарь"
        description="Управляйте бронированиями, блокировками и внешней синхронизацией в единой сетке."
        size="compact"
        extra={(
        <Space wrap>
          {hasMultipleBathhouses && (
            <Tooltip title="Показать бронирования всех объектов">
              <Space>
                <Text className="rh-calendar-toggle-label">Все объекты</Text>
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
        )}
      />

      {/* Navigation + View Switcher */}
      <Card size="small" className="rh-admin-detail-card rh-calendar-nav-card">
        <div className="rh-calendar-nav">
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
        <div className="rh-calendar-legend">
          {consolidatedView ? (
            // Show bathhouse color legend in consolidated mode
            Array.from(bathhouseColorMap.entries()).map(([id, info]) => (
              <Space key={id} size={4}>
                <span className={`rh-calendar-legend__dot rh-calendar-tone--${info.tone}`} />
                <Text className="rh-calendar-legend__label">{info.name}</Text>
              </Space>
            ))
          ) : (
            LEGEND_ITEMS.map((item) => (
              <Space key={item.label} size={4}>
                <span className={`rh-calendar-legend__dot rh-calendar-tone--${item.tone}`} />
                <Text className="rh-calendar-legend__label">{item.label}</Text>
              </Space>
            ))
          )}
        </div>
      </Card>

      {/* Calendar Grid */}
      <Card
        size="small"
        className="rh-calendar-grid-card"
      >
        {(bookingsLoading || consolidatedLoading) ? (
          <div className="rh-calendar-loading">
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
            className="rh-admin-detail-card"
          >
            <Text type="secondary" className="rh-card-intro-text">
              Используйте ссылку для подключения календаря бронирований в Google Calendar,
              Apple Calendar или другие приложения.
            </Text>
            {calendarToken?.url ? (
              <Space orientation="vertical" className="rh-full-width">
                <Space.Compact className="rh-compact-control">
                  <Input
                    readOnly
                    value={`${window.location.origin}${calendarToken.url}`}
                  />
                  <Button
                    aria-label="Скопировать ссылку iCal"
                    className="rh-input-addon-button"
                    icon={<CopyOutlined />}
                    onClick={handleCopyIcalUrl}
                  />
                </Space.Compact>
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
            className="rh-admin-detail-card"
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
                            className="rh-calendar-url"
                            title={cal.url}
                          >
                            {cal.url}
                          </Text>
                        </Space>
                      }
                      description={
                        <Space>
                          {cal.last_sync_at && (
                            <Text type="secondary" className="rh-calendar-sync-meta">
                              Синхр.: {formatDateTime(cal.last_sync_at)}
                            </Text>
                          )}
                          {cal.last_error && (
                            <Text type="danger" className="rh-calendar-sync-meta">
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
              className="rh-full-width"
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
          className="rh-modal-alert"
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
          className="rh-modal-alert"
        />
      </Modal>
    </div>
  )
}
