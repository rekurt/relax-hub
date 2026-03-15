import { useState, useMemo } from 'react'
import {
  Alert,
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
  Select,
  Space,
  Spin,
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
import { App } from 'antd'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetMyBathhousesIdCalendarToken,
  usePostMyBathhousesIdSlotBlocks,
  useGetMyBathhousesIdExternalCalendars,
  usePostMyBathhousesIdExternalCalendars,
  usePostMyBathhousesIdExternalCalendarsSync,
  useDeleteMyExternalCalendarsId,
} from '@/api/generated/calendar/calendar'
import { useGetBathhousesIdBookings } from '@/api/generated/bookings/bookings'
import type {
  InternalHandlerBookingResponse,
  InternalHandlerExternalCalendarResponse,
} from '@/api/generated/model'
import { useBathhouseStore } from '@/stores/bathhouse'
import { formatPrice, formatTime, formatDateTime } from '@/lib/format'
import { BOOKING_STATUS_CONFIG } from '@/lib/constants'

const { Title, Text } = Typography
const { RangePicker } = DatePicker

const HOURS = Array.from({ length: 24 }, (_, i) => i)
const DAYS_SHORT = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

const SOURCE_OPTIONS = [
  { value: 'google_calendar', label: 'Google Calendar' },
  { value: 'yandex_calendar', label: 'Яндекс.Календарь' },
]

interface BookingBlock {
  booking: InternalHandlerBookingResponse
  top: number
  height: number
  dayIndex: number
}

export default function CalendarPage() {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)

  const [weekStart, setWeekStart] = useState<Dayjs>(() => dayjs().startOf('week'))
  const [slotBlockModalOpen, setSlotBlockModalOpen] = useState(false)
  const [externalCalendarModalOpen, setExternalCalendarModalOpen] = useState(false)
  const [slotBlockForm] = Form.useForm()
  const [externalCalForm] = Form.useForm()

  const weekEnd = weekStart.add(6, 'day').endOf('day')

  const { data: bookingsData, isLoading: bookingsLoading } = useGetBathhousesIdBookings(
    selectedBathhouseId ?? '',
    { page: 1, page_size: 100 },
    { query: { enabled: !!selectedBathhouseId } },
  )

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

  const weekDays = useMemo(
    () => Array.from({ length: 7 }, (_, i) => weekStart.add(i, 'day')),
    [weekStart],
  )

  const bookingBlocks: BookingBlock[] = useMemo(() => {
    const blocks: BookingBlock[] = []
    const items = bookingsData?.data ?? []
    for (const booking of items) {
      if (!booking.start_time || !booking.end_time) continue
      if (booking.status === 'cancelled' || booking.status === 'rejected') continue

      const start = dayjs(booking.start_time)
      const end = dayjs(booking.end_time)

      for (let d = 0; d < 7; d++) {
        const day = weekDays[d]!
        const dayStart = day.startOf('day')
        const dayEnd = day.endOf('day')

        if (start.isBefore(dayEnd) && end.isAfter(dayStart)) {
          const effectiveStart = start.isAfter(dayStart) ? start : dayStart
          const effectiveEnd = end.isBefore(dayEnd) ? end : dayEnd

          const startMinutes = effectiveStart.hour() * 60 + effectiveStart.minute()
          const endMinutes = effectiveEnd.hour() * 60 + effectiveEnd.minute()

          const top = (startMinutes / 60) * 60
          const height = Math.max(((endMinutes - startMinutes) / 60) * 60, 20)

          blocks.push({ booking, top, height, dayIndex: d })
        }
      }
    }
    return blocks
  }, [bookingsData?.data, weekDays])

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

  const navigateWeek = (direction: number) => {
    setWeekStart((prev) => prev.add(direction * 7, 'day'))
  }

  const goToToday = () => {
    setWeekStart(dayjs().startOf('week'))
  }

  if (!selectedBathhouseId) {
    return (
      <div>
        <Title level={3}>Календарь</Title>
        <Alert
          message="Выберите баню"
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

      {/* Week navigation */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Space>
            <Button icon={<LeftOutlined />} onClick={() => navigateWeek(-1)} />
            <Button onClick={goToToday}>Сегодня</Button>
            <Button icon={<RightOutlined />} onClick={() => navigateWeek(1)} />
          </Space>
          <Text strong>
            {weekStart.format('DD MMM')} — {weekEnd.format('DD MMM YYYY')}
          </Text>
          <Space>
            <Tag color="blue">Подтверждено</Tag>
            <Tag color="orange">Ожидает</Tag>
            <Tag color="green">Завершено</Tag>
          </Space>
        </div>
      </Card>

      {/* Weekly calendar grid */}
      <Card
        size="small"
        style={{ marginBottom: 16, overflow: 'auto' }}
        styles={{ body: { padding: 0 } }}
      >
        {bookingsLoading ? (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <Spin size="large" />
          </div>
        ) : (
          <div style={{ display: 'flex', minWidth: 800 }}>
            {/* Time column */}
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

            {/* Day columns */}
            {weekDays.map((day, dayIndex) => {
              const isToday = day.isSame(dayjs(), 'day')
              const dayBlocks = bookingBlocks.filter((b) => b.dayIndex === dayIndex)

              return (
                <div
                  key={dayIndex}
                  style={{
                    flex: 1,
                    borderRight: dayIndex < 6 ? '1px solid #f0f0f0' : undefined,
                    minWidth: 100,
                  }}
                >
                  {/* Day header */}
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
                    <Text
                      type="secondary"
                      style={{ fontSize: 11, lineHeight: 1 }}
                    >
                      {DAYS_SHORT[dayIndex]}
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

                  {/* Hours grid with bookings */}
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
                    {dayBlocks.map((block, i) => {
                      const status = block.booking.status ?? 'pending'
                      const statusConf = BOOKING_STATUS_CONFIG[status]
                      const bgColor = statusConf?.hexColor ?? '#1677ff'
                      return (
                        <Tooltip
                          key={`${block.booking.id}-${i}`}
                          title={
                            <div>
                              <div>{statusConf?.text ?? status}</div>
                              <div>
                                {block.booking.start_time
                                  ? formatTime(block.booking.start_time)
                                  : ''}{' '}
                                –{' '}
                                {block.booking.end_time
                                  ? formatTime(block.booking.end_time)
                                  : ''}
                              </div>
                              <div>
                                Гостей: {block.booking.guest_count ?? '—'}
                              </div>
                              <div>
                                {formatPrice(block.booking.total_price ?? 0)}
                              </div>
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
                              {block.booking.start_time
                                ? formatTime(block.booking.start_time)
                                : ''}
                            </div>
                            {block.height > 30 && (
                              <div>{formatPrice(block.booking.total_price ?? 0)}</div>
                            )}
                          </div>
                        </Tooltip>
                      )
                    })}
                  </div>
                </div>
              )
            })}
          </div>
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
              <Space direction="vertical" style={{ width: '100%' }}>
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
                          loading={removeExternalCalMutation.isPending}
                        />
                      </Popconfirm>,
                    ]}
                  >
                    <List.Item.Meta
                      avatar={<CalendarOutlined />}
                      title={
                        <Space>
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
          message="Блокировка запретит бронирование на указанный период."
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
          message="Занятые слоты из внешнего календаря будут автоматически блокировать бронирования."
          type="info"
          showIcon
          style={{ marginTop: 8 }}
        />
      </Modal>
    </div>
  )
}
