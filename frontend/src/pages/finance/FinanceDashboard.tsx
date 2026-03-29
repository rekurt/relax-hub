import { useState } from 'react'
import {
  Card,
  Col,
  DatePicker,
  Row,
  Select,
  Space,
  Statistic,
  Table,
  Tag,
  Typography,
} from 'antd'
import {
  LockOutlined,
  WalletOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { useGetMyWallet, useGetMyWalletTransactions } from '@/api/generated/wallet/wallet'
import type { InternalHandlerWalletTransactionResponse } from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'
import EmptyState from '@/components/EmptyState'

const { Title } = Typography
const { RangePicker } = DatePicker

const TX_TYPE_OPTIONS = [
  { value: '', label: 'Все типы' },
  { value: 'income', label: 'Доход' },
  { value: 'payout', label: 'Выплата' },
  { value: 'service_fee', label: 'Комиссия' },
  { value: 'refund', label: 'Возврат' },
  { value: 'cashback', label: 'Кэшбэк' },
  { value: 'promo', label: 'Промо' },
]

const TX_TYPE_COLORS: Record<string, string> = {
  income: 'green',
  payout: 'blue',
  service_fee: 'orange',
  refund: 'red',
  cashback: 'gold',
  promo: 'purple',
  spend: 'red',
  topup: 'green',
}

const TX_TYPE_LABELS: Record<string, string> = {
  income: 'Доход',
  payout: 'Выплата',
  service_fee: 'Комиссия',
  refund: 'Возврат',
  cashback: 'Кэшбэк',
  promo: 'Промо',
  spend: 'Списание',
  topup: 'Пополнение',
}

const PERIOD_OPTIONS = [
  { value: 'week', label: 'Неделя' },
  { value: 'month', label: 'Месяц' },
  { value: 'year', label: 'Год' },
]

export default function FinanceDashboard() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [typeFilter, setTypeFilter] = useState('')
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs | null, dayjs.Dayjs | null] | null>(null)
  const [period, setPeriod] = useState('month')

  const { data: walletData, isLoading: walletLoading } = useGetMyWallet()
  const wallet = walletData?.data

  const { data: txData, isLoading: txLoading } = useGetMyWalletTransactions({
    page,
    page_size: pageSize,
    ...(typeFilter ? { type: typeFilter } : {}),
  })
  const transactions = txData?.data ?? []
  const txMeta = txData?.meta

  const filteredTransactions = transactions.filter((tx) => {
    if (dateRange?.[0] && dateRange?.[1] && tx.created_at) {
      const txDate = dayjs(tx.created_at)
      if (txDate.isBefore(dateRange[0], 'day') || txDate.isAfter(dateRange[1], 'day')) {
        return false
      }
    }
    return true
  })

  const txColumns: ColumnsType<InternalHandlerWalletTransactionResponse> = [
    {
      title: 'Дата',
      key: 'created_at',
      render: (_, record) => record.created_at ? formatDateTime(record.created_at, 'DD.MM.YYYY HH:mm') : '\u2014',
      sorter: (a, b) => (a.created_at ?? '').localeCompare(b.created_at ?? ''),
    },
    {
      title: 'Тип',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => (
        <Tag color={TX_TYPE_COLORS[type] ?? 'default'}>
          {TX_TYPE_LABELS[type] ?? type}
        </Tag>
      ),
    },
    {
      title: 'Сумма',
      dataIndex: 'amount',
      key: 'amount',
      render: (amount: number, record) => {
        const isPositive = record.type !== 'spend' && record.type !== 'service_fee' && record.type !== 'payout'
        return (
          <span style={{ color: isPositive ? '#52c41a' : '#ff4d4f', fontWeight: 500 }}>
            {isPositive ? '+' : '-'}{formatPrice(amount ?? 0)}
          </span>
        )
      },
    },
    {
      title: 'Описание',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: 'Баланс после',
      dataIndex: 'balance_after',
      key: 'balance_after',
      render: (val: number) => val != null ? formatPrice(val) : '\u2014',
      responsive: ['md'] as const,
    },
  ]

  return (
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>Финансы</Title>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={8}>
          <Card loading={walletLoading}>
            <Statistic
              title="Баланс"
              value={wallet?.balance ? wallet.balance / 100 : 0}
              suffix="\u20BD"
              prefix={<WalletOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8}>
          <Card loading={walletLoading}>
            <Statistic
              title="Доступно"
              value={wallet?.available ? wallet.available / 100 : 0}
              suffix="\u20BD"
              styles={{ content: { color: '#52c41a' } }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8}>
          <Card loading={walletLoading}>
            <Statistic
              title="Заморожено"
              value={wallet?.held_amount ? wallet.held_amount / 100 : 0}
              suffix="\u20BD"
              prefix={<LockOutlined />}
              styles={{ content: { color: '#faad14' } }}
            />
          </Card>
        </Col>
      </Row>

      <Card style={{ marginTop: 16 }}>
        <Title level={5} style={{ marginBottom: 12 }}>Доходы за период</Title>
        <Select
          value={period}
          onChange={setPeriod}
          options={PERIOD_OPTIONS}
          style={{ width: 150 }}
        />
        <div style={{ marginTop: 16, padding: 24, background: '#fafafa', borderRadius: 8, textAlign: 'center', color: '#999' }}>
          График доходов ({PERIOD_OPTIONS.find(o => o.value === period)?.label})
        </div>
      </Card>

      <Title level={4} style={{ marginTop: 24, marginBottom: 12 }}>Последние операции</Title>

      <Space wrap style={{ marginBottom: 16 }}>
        <Select
          value={typeFilter}
          onChange={(val) => { setTypeFilter(val); setPage(1) }}
          options={TX_TYPE_OPTIONS}
          style={{ width: 200 }}
          placeholder="Тип операции"
        />
        <RangePicker
          value={dateRange}
          onChange={(dates) => setDateRange(dates)}
          format="DD.MM.YYYY"
          placeholder={['С', 'По']}
        />
      </Space>

      <Table
        columns={txColumns}
        dataSource={filteredTransactions}
        rowKey="id"
        loading={txLoading}
        locale={{ emptyText: <EmptyState description="Нет операций" /> }}
        pagination={{
          current: page,
          pageSize: pageSize,
          total: txMeta?.total_count ?? 0,
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
