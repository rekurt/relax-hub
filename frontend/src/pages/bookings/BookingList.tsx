import { useState } from 'react'
import { App, Button, DatePicker, Popconfirm, Select, Space, Table, Tag } from '@/components/design/system'
import type { ColumnsType } from '@/components/design/types'
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  StopOutlined,
  CheckOutlined,
  EyeOutlined,
  DollarOutlined,
  SwapOutlined,
  LoginOutlined,
  LogoutOutlined,
  ClockCircleOutlined,
} from '@/components/design/icons'
import dayjs from 'dayjs'
import {
  useGetBathhousesIdBookings,
  usePatchBookingsIdConfirm,
  usePatchBookingsIdReject,
  usePatchBookingsIdCancel,
  usePatchBookingsIdComplete,
  usePatchBookingsIdCheckIn,
  usePatchBookingsIdCheckOut,
} from '@/api/generated/bookings/bookings'
import { usePostBookingsIdPay } from '@/api/generated/payments/payments'
import type { InternalHandlerBookingResponse } from '@/api/generated/model'
import { useBathhouseStore } from '@/stores/bathhouse'
import { formatPrice, formatDateTime } from '@/lib/format'
import { BOOKING_STATUS_CONFIG, PAYMENT_STATUS_CONFIG } from '@/lib/constants'
import { useQueryClient } from '@tanstack/react-query'
import BookingDetails from './BookingDetails'
import ModificationRequests from './ModificationRequests'
import ExtensionRequests from './ExtensionRequests'
import EmptyState from '@/components/EmptyState'
import PageHeader from '@/components/PageHeader'

const { RangePicker } = DatePicker

const STATUS_OPTIONS = [
  { value: '', label: 'Все статусы' },
  { value: 'pending', label: 'Ожидает' },
  { value: 'pending_owner', label: 'Ожидает подтверждения' },
  { value: 'confirmed', label: 'Подтверждено' },
  { value: 'completed', label: 'Завершено' },
  { value: 'cancelled', label: 'Отменено' },
  { value: 'rejected', label: 'Отклонено' },
  { value: 'no_show', label: 'Неявка' },
]

export default function BookingList() {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [statusFilter, setStatusFilter] = useState('')
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs | null, dayjs.Dayjs | null] | null>(null)
  const [detailsBooking, setDetailsBooking] = useState<InternalHandlerBookingResponse | null>(null)
  const [modRequestsBookingId, setModRequestsBookingId] = useState<string | null>(null)
  const [extRequestsBookingId, setExtRequestsBookingId] = useState<string | null>(null)

  const hasActiveFilter = !!statusFilter || !!(dateRange?.[0] && dateRange?.[1])

  const { data, isLoading } = useGetBathhousesIdBookings(selectedBathhouseId ?? '', {
    page: hasActiveFilter ? 1 : page,
    page_size: hasActiveFilter ? 999 : pageSize,
  }, {
    query: {
      enabled: !!selectedBathhouseId,
    },
  })

  const invalidateBookings = () => {
    queryClient.invalidateQueries({
      queryKey: [`/bathhouses/${selectedBathhouseId}/bookings`],
    })
  }

  const confirmMutation = usePatchBookingsIdConfirm({
    mutation: {
      onSuccess: () => {
        message.success('Бронирование подтверждено')
        invalidateBookings()
      },
      onError: () => message.error('Не удалось подтвердить бронирование'),
    },
  })

  const rejectMutation = usePatchBookingsIdReject({
    mutation: {
      onSuccess: () => {
        message.success('Бронирование отклонено')
        invalidateBookings()
      },
      onError: () => message.error('Не удалось отклонить бронирование'),
    },
  })

  const cancelMutation = usePatchBookingsIdCancel({
    mutation: {
      onSuccess: () => {
        message.success('Бронирование отменено')
        invalidateBookings()
      },
      onError: () => message.error('Не удалось отменить бронирование'),
    },
  })

  const completeMutation = usePatchBookingsIdComplete({
    mutation: {
      onSuccess: () => {
        message.success('Бронирование завершено')
        invalidateBookings()
      },
      onError: () => message.error('Не удалось завершить бронирование'),
    },
  })

  const checkInMutation = usePatchBookingsIdCheckIn({
    mutation: {
      onSuccess: () => {
        message.success('Гость отмечен как прибывший')
        invalidateBookings()
      },
      onError: () => message.error('Не удалось отметить прибытие гостя'),
    },
  })

  const checkOutMutation = usePatchBookingsIdCheckOut({
    mutation: {
      onSuccess: () => {
        message.success('Гость отмечен как ушедший')
        invalidateBookings()
      },
      onError: () => message.error('Не удалось отметить уход гостя'),
    },
  })

  const payMutation = usePostBookingsIdPay({
    mutation: {
      onSuccess: (response) => {
        const url = response?.data?.confirmation_url
        if (url) {
          window.open(url, '_blank')
        }
        message.success('Платёж инициирован')
        invalidateBookings()
      },
      onError: () => message.error('Не удалось инициировать оплату'),
    },
  })

  const bookings = data?.data ?? []
  const meta = data?.meta

  const filteredBookings = bookings.filter((b) => {
    if (statusFilter && b.status !== statusFilter) return false
    if (dateRange?.[0] && dateRange?.[1] && b.start_time) {
      const bookingDate = dayjs(b.start_time)
      if (bookingDate.isBefore(dateRange[0], 'day') || bookingDate.isAfter(dateRange[1], 'day')) {
        return false
      }
    }
    return true
  })

  const renderActions = (record: InternalHandlerBookingResponse) => {
    const actions: React.ReactNode[] = []

    actions.push(
      <Button
        key="details"
        type="link"
        size="small"
        icon={<EyeOutlined />}
        onClick={() => setDetailsBooking(record)}
      >
        Детали
      </Button>,
    )

    if (record.status === 'confirmed' || record.status === 'pending') {
      actions.push(
        <Button
          key="mod-requests"
          type="link"
          size="small"
          icon={<SwapOutlined />}
          onClick={() => record.id && setModRequestsBookingId(record.id)}
        >
          Изменения
        </Button>,
      )
    }

    if (record.status === 'confirmed') {
      actions.push(
        <Button
          key="ext-requests"
          type="link"
          size="small"
          icon={<ClockCircleOutlined />}
          onClick={() => record.id && setExtRequestsBookingId(record.id)}
        >
          Продление
        </Button>,
      )
    }

    if (record.status === 'pending' || record.status === 'pending_owner') {
      actions.push(
        <Button
          key="confirm"
          type="link"
          size="small"
          icon={<CheckCircleOutlined />}
          loading={confirmMutation.isPending && confirmMutation.variables?.id === record.id}
          onClick={() => record.id && confirmMutation.mutate({ id: record.id })}
        >
          Подтвердить
        </Button>,
      )
      actions.push(
        <Popconfirm
          key="reject"
          title="Отклонить бронирование?"
          description="Это действие нельзя отменить."
          onConfirm={() => record.id && rejectMutation.mutate({ id: record.id, data: {} })}
          okText="Отклонить"
          cancelText="Нет"
          okButtonProps={{ danger: true }}
        >
          <Button
            type="link"
            size="small"
            danger
            icon={<CloseCircleOutlined />}
            loading={rejectMutation.isPending && rejectMutation.variables?.id === record.id}
          >
            Отклонить
          </Button>
        </Popconfirm>,
      )
    }

    if (record.status === 'confirmed') {
      if (!record.payment_status || record.payment_status === 'pending') {
        actions.push(
          <Button
            key="pay"
            type="link"
            size="small"
            icon={<DollarOutlined />}
            loading={payMutation.isPending && payMutation.variables?.id === record.id}
            onClick={() => record.id && payMutation.mutate({ id: record.id, data: {} })}
          >
            Оплатить
          </Button>,
        )
      }

      if (!record.checked_in_at) {
        const now = dayjs()
        const start = record.start_time ? dayjs(record.start_time) : null
        const canCheckIn = start && now.isAfter(start.subtract(15, 'minute')) && now.isBefore(start.add(30, 'minute'))
        actions.push(
          <Button
            key="checkin"
            type="link"
            size="small"
            icon={<LoginOutlined />}
            disabled={!canCheckIn}
            title={canCheckIn ? 'Отметить прибытие гостя' : 'Регистрация доступна за 15 мин до и 30 мин после начала'}
            loading={checkInMutation.isPending && checkInMutation.variables?.id === record.id}
            onClick={() => record.id && checkInMutation.mutate({ id: record.id })}
          >
            Гость прибыл
          </Button>,
        )
      }

      if (record.checked_in_at && !record.checked_out_at) {
        actions.push(
          <Button
            key="checkout"
            type="link"
            size="small"
            icon={<LogoutOutlined />}
            loading={checkOutMutation.isPending && checkOutMutation.variables?.id === record.id}
            onClick={() => record.id && checkOutMutation.mutate({ id: record.id })}
          >
            Гость ушёл
          </Button>,
        )
      }

      if (record.checked_in_at && record.checked_out_at) {
        actions.push(
          <Button
            key="complete"
            type="link"
            size="small"
            icon={<CheckOutlined />}
            loading={completeMutation.isPending && completeMutation.variables?.id === record.id}
            onClick={() => record.id && completeMutation.mutate({ id: record.id })}
          >
            Завершить
          </Button>,
        )
      }
      actions.push(
        <Popconfirm
          key="cancel"
          title="Отменить бронирование?"
          description="Это может повлечь автоматический возврат средств."
          onConfirm={() => record.id && cancelMutation.mutate({ id: record.id, data: {} })}
          okText="Отменить"
          cancelText="Нет"
          okButtonProps={{ danger: true }}
        >
          <Button
            type="link"
            size="small"
            danger
            icon={<StopOutlined />}
            loading={cancelMutation.isPending && cancelMutation.variables?.id === record.id}
          >
            Отменить
          </Button>
        </Popconfirm>,
      )
    }

    return <Space wrap>{actions}</Space>
  }

  const columns: ColumnsType<InternalHandlerBookingResponse> = [
    {
      title: 'Дата/время',
      key: 'datetime',
      render: (_, record) => (
        <div>
          <div>{record.start_time ? formatDateTime(record.start_time, 'DD.MM.YYYY') : '—'}</div>
          <div className="rh-table-meta-text">
            {record.start_time ? formatDateTime(record.start_time, 'HH:mm') : ''}
            {record.end_time ? ` – ${formatDateTime(record.end_time, 'HH:mm')}` : ''}
          </div>
        </div>
      ),
      sorter: (a, b) => (a.start_time ?? '').localeCompare(b.start_time ?? ''),
    },
    {
      title: 'Гость',
      dataIndex: 'user_id',
      key: 'user_id',
      render: (userId: string) => userId ? userId.slice(0, 8) + '...' : '—',
      responsive: ['md'],
    },
    {
      title: 'Гостей',
      dataIndex: 'guest_count',
      key: 'guest_count',
      render: (count: number) => count ?? '—',
      responsive: ['sm'],
    },
    {
      title: 'Сумма',
      dataIndex: 'total_price',
      key: 'total_price',
      render: (price: number) => formatPrice(price ?? 0),
    },
    {
      title: 'Оплата',
      dataIndex: 'payment_status',
      key: 'payment_status',
      render: (status: string) => {
        if (!status) return <Tag>Не оплачено</Tag>
        const config = PAYMENT_STATUS_CONFIG[status] ?? { color: 'default', text: status }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: 'Регистрация',
      key: 'checkin_status',
      responsive: ['lg'] as const,
      render: (_: unknown, record: InternalHandlerBookingResponse) => {
        if (record.status !== 'confirmed' && record.status !== 'completed') return null
        if (record.checked_out_at) {
          return <Tag color="green" icon={<LogoutOutlined />}>Ушёл</Tag>
        }
        if (record.checked_in_at) {
          return <Tag color="blue" icon={<LoginOutlined />}>Прибыл</Tag>
        }
        return <Tag>Ожидает</Tag>
      },
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const config = BOOKING_STATUS_CONFIG[status] ?? { color: 'default', text: status }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_, record) => renderActions(record),
    },
  ]

  if (!selectedBathhouseId) {
    return (
      <div className="rh-stack">
        <PageHeader
          eyebrow="Операции"
          title="Бронирования"
          description="Выберите объект, чтобы управлять заявками и визитами."
          size="compact"
        />
        <EmptyState description="Выберите баню для просмотра бронирований" />
      </div>
    )
  }

  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="Операции"
        title="Бронирования"
        description="Подтверждайте заявки, отмечайте визиты и контролируйте оплату."
        size="compact"
      />

      <Space wrap className="rh-page-toolbar">
        <Select
          value={statusFilter}
          onChange={setStatusFilter}
          options={STATUS_OPTIONS}
          className="rh-client-filter-control"
          placeholder="Статус"
        />
        <RangePicker
          value={dateRange}
          onChange={(dates) => setDateRange(dates)}
          format="DD.MM.YYYY"
          placeholder={['С', 'По']}
        />
      </Space>

      <Table
        columns={columns}
        dataSource={filteredBookings}
        rowKey="id"
        loading={isLoading}
        locale={{ emptyText: <EmptyState description="Пока нет бронирований. Убедитесь, что ваш объект активен и заполнен" /> }}
        pagination={hasActiveFilter ? {
          pageSize: pageSize,
          showSizeChanger: true,
          showTotal: (total) => `Всего: ${total}`,
        } : {
          current: page,
          pageSize: pageSize,
          total: meta?.total_count ?? 0,
          showSizeChanger: true,
          showTotal: (total) => `Всего: ${total}`,
          onChange: (p, ps) => {
            setPage(p)
            setPageSize(ps)
          },
        }}
      />

      <BookingDetails
        booking={detailsBooking}
        onClose={() => setDetailsBooking(null)}
      />

      <ModificationRequests
        bookingId={modRequestsBookingId ?? ''}
        open={!!modRequestsBookingId}
        onClose={() => setModRequestsBookingId(null)}
      />

      <ExtensionRequests
        bookingId={extRequestsBookingId ?? ''}
        open={!!extRequestsBookingId}
        onClose={() => setExtRequestsBookingId(null)}
      />
    </div>
  )
}
