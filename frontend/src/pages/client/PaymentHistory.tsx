import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { DatePicker, Select, Space, Table, Tag, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { useGetMyPayments } from '@/api/generated/payments/payments'
import type { InternalHandlerPaymentResponse } from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'
import { PAYMENT_STATUS_CONFIG } from '@/lib/constants'

const { Title } = Typography
const { RangePicker } = DatePicker

const STATUS_OPTIONS = [
  { value: '', label: 'Все статусы' },
  ...Object.entries(PAYMENT_STATUS_CONFIG).map(([value, { text }]) => ({ value, label: text })),
]

export default function PaymentHistory() {
  const navigate = useNavigate()

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [statusFilter, setStatusFilter] = useState('')
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs | null, dayjs.Dayjs | null] | null>(null)

  const { data, isLoading } = useGetMyPayments({ page, page_size: pageSize })

  const payments = data?.data ?? []
  const meta = data?.meta

  const filteredPayments = payments.filter((p) => {
    if (statusFilter && p.status !== statusFilter) return false
    if (dateRange?.[0] && dateRange?.[1] && p.created_at) {
      const paymentDate = dayjs(p.created_at)
      if (paymentDate.isBefore(dateRange[0], 'day') || paymentDate.isAfter(dateRange[1], 'day')) {
        return false
      }
    }
    return true
  })

  const columns: ColumnsType<InternalHandlerPaymentResponse> = [
    {
      title: 'Дата',
      key: 'created_at',
      render: (_, record) => record.created_at ? formatDateTime(record.created_at, 'DD.MM.YYYY HH:mm') : '—',
      sorter: (a, b) => (a.created_at ?? '').localeCompare(b.created_at ?? ''),
    },
    {
      title: 'Сумма',
      dataIndex: 'amount',
      key: 'amount',
      render: (amount: number) => formatPrice(amount ?? 0),
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const config = PAYMENT_STATUS_CONFIG[status] ?? { color: 'default', text: status }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: 'Возврат',
      key: 'refund',
      render: (_, record) => {
        if (!record.refund_amount) return '—'
        return (
          <span>
            {formatPrice(record.refund_amount)}
            {record.refunded_at && (
              <div style={{ color: '#888', fontSize: 12 }}>
                {formatDateTime(record.refunded_at, 'DD.MM.YYYY')}
              </div>
            )}
          </span>
        )
      },
      responsive: ['md'] as const,
    },
    {
      title: 'Бронирование',
      key: 'booking',
      render: (_, record) =>
        record.booking_id ? (
          <a onClick={() => navigate(`/client/bookings/${record.booking_id}`)}>
            Открыть
          </a>
        ) : (
          '—'
        ),
    },
  ]

  return (
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>История платежей</Title>

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
        dataSource={filteredPayments}
        rowKey="id"
        loading={isLoading}
        locale={{ emptyText: 'Нет платежей' }}
        pagination={{
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
