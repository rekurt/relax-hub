import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { App, Button, DatePicker, Popconfirm, Select, Space, Table, Tag, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { EyeOutlined, SearchOutlined, StopOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { useGetBookings, usePatchBookingsIdCancel } from '@/api/generated/bookings/bookings'
import type { InternalHandlerBookingResponse } from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'
import { BOOKING_STATUS_CONFIG } from '@/lib/constants'
import { useQueryClient } from '@tanstack/react-query'
import EmptyState from '@/components/EmptyState'

const { Title } = Typography
const { RangePicker } = DatePicker

const STATUS_OPTIONS = [
  { value: '', label: 'Все статусы' },
  { value: 'pending', label: 'Ожидает' },
  { value: 'confirmed', label: 'Подтверждено' },
  { value: 'completed', label: 'Завершено' },
  { value: 'cancelled', label: 'Отменено' },
  { value: 'rejected', label: 'Отклонено' },
]

export default function ClientBookingList() {
  const navigate = useNavigate()
  const { message } = App.useApp()
  const queryClient = useQueryClient()

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [statusFilter, setStatusFilter] = useState('')
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs | null, dayjs.Dayjs | null] | null>(null)

  const hasActiveFilter = !!statusFilter || !!(dateRange?.[0] && dateRange?.[1])

  const { data, isLoading } = useGetBookings({
    page: hasActiveFilter ? 1 : page,
    page_size: hasActiveFilter ? 999 : pageSize,
  })

  const cancelMutation = usePatchBookingsIdCancel({
    mutation: {
      onSuccess: () => {
        message.success('Бронирование отменено')
        queryClient.invalidateQueries({ queryKey: ['/bookings'] })
      },
      onError: () => message.error('Не удалось отменить бронирование'),
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

  const canCancel = (status?: string) => status === 'pending' || status === 'confirmed'

  const getRefundInfo = (startTime?: string) => {
    if (!startTime) return ''
    const hoursUntil = dayjs(startTime).diff(dayjs(), 'hour')
    if (hoursUntil > 24) return '(100% возврат)'
    if (hoursUntil >= 2) return '(50% возврат)'
    return '(без возврата)'
  }

  const columns: ColumnsType<InternalHandlerBookingResponse> = [
    {
      title: 'Дата/время',
      key: 'datetime',
      render: (_, record) => (
        <div>
          <div>{record.start_time ? formatDateTime(record.start_time, 'DD.MM.YYYY') : '—'}</div>
          <div style={{ color: '#888', fontSize: 12 }}>
            {record.start_time ? formatDateTime(record.start_time, 'HH:mm') : ''}
            {record.end_time ? ` – ${formatDateTime(record.end_time, 'HH:mm')}` : ''}
          </div>
        </div>
      ),
      sorter: (a, b) => (a.start_time ?? '').localeCompare(b.start_time ?? ''),
    },
    {
      title: 'Сумма',
      dataIndex: 'total_price',
      key: 'total_price',
      render: (price: number) => formatPrice(price ?? 0),
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
      title: 'Оплата',
      dataIndex: 'payment_status',
      key: 'payment_status',
      render: (status: string) => {
        if (!status) return <Tag>Не оплачено</Tag>
        const config: Record<string, { color: string; text: string }> = {
          pending: { color: 'orange', text: 'Ожидает' },
          succeeded: { color: 'green', text: 'Оплачено' },
          canceled: { color: 'default', text: 'Отменён' },
          refunded: { color: 'purple', text: 'Возвращён' },
        }
        const c = config[status] ?? { color: 'default', text: status }
        return <Tag color={c.color}>{c.text}</Tag>
      },
      responsive: ['md'],
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_, record) => (
        <Space wrap>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/client/bookings/${record.id}`)}
          >
            Детали
          </Button>
          {canCancel(record.status) && (
            <Popconfirm
              title="Отменить бронирование?"
              description={`Это действие нельзя отменить. ${getRefundInfo(record.start_time)}`}
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
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>Мои бронирования</Title>

      <Space wrap style={{ marginBottom: 16 }}>
        <Select
          value={statusFilter}
          onChange={setStatusFilter}
          options={STATUS_OPTIONS}
          style={{ width: 160 }}
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
        locale={{ emptyText: <EmptyState description="У вас пока нет бронирований" actionText="Найти баню" actionLink="/client/search" icon={<SearchOutlined />} /> }}
        pagination={hasActiveFilter ? {
          pageSize: 999,
          hideOnSinglePage: true,
          showTotal: (total) => `Найдено: ${total}`,
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
    </div>
  )
}
