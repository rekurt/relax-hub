import { useCallback, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Alert,
  App,
  Button,
  Card,
  Col,
  DatePicker,
  Dropdown,
  Form,
  InputNumber,
  Row,
  Select,
  Space,
  Statistic,
  Table,
  Tag,
  Typography,
} from 'antd'
import {
  DownloadOutlined,
  LockOutlined,
  PlusOutlined,
  WalletOutlined,
  WarningOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { useGetMyWallet, useGetMyWalletTransactions, useGetMyWalletHolds, usePostMyWalletTopup } from '@/api/generated/wallet/wallet'
import type { InternalHandlerWalletTransactionResponse, InternalHandlerWalletHoldResponse } from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'
import { AUTH_TOKEN_KEY } from '@/lib/constants'
import EmptyState from '@/components/EmptyState'

const { Title, Text } = Typography
const { RangePicker } = DatePicker

const TX_TYPE_OPTIONS = [
  { value: '', label: 'Все типы' },
  { value: 'topup', label: 'Пополнение' },
  { value: 'spend', label: 'Списание' },
  { value: 'refund', label: 'Возврат' },
  { value: 'cashback', label: 'Кэшбэк' },
  { value: 'promo', label: 'Промо' },
  { value: 'referral', label: 'Реферальный' },
  { value: 'welcome_bonus', label: 'Бонус за регистрацию' },
  { value: 'gift_cert', label: 'Сертификат' },
]

const TX_TYPE_COLORS: Record<string, string> = {
  topup: 'green',
  spend: 'red',
  refund: 'blue',
  cashback: 'gold',
  promo: 'purple',
  referral: 'cyan',
  welcome_bonus: 'magenta',
  gift_cert: 'orange',
  bonus: 'gold',
}

const TX_TYPE_LABELS: Record<string, string> = {
  topup: 'Пополнение',
  spend: 'Списание',
  refund: 'Возврат',
  cashback: 'Кэшбэк',
  promo: 'Промо',
  referral: 'Реферальный',
  welcome_bonus: 'Приветственный',
  gift_cert: 'Сертификат',
  bonus: 'Бонус',
}

const MIN_TOPUP = 500
const MAX_TOPUP = 30000

export default function WalletDashboard() {
  const navigate = useNavigate()
  const { message } = App.useApp()
  const [form] = Form.useForm()

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [typeFilter, setTypeFilter] = useState('')
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs | null, dayjs.Dayjs | null] | null>(null)
  const [exporting, setExporting] = useState(false)

  const { data: walletData, isLoading: walletLoading } = useGetMyWallet()
  const wallet = walletData?.data

  const { data: txData, isLoading: txLoading } = useGetMyWalletTransactions({
    page,
    page_size: pageSize,
    ...(typeFilter ? { type: typeFilter } : {}),
  })
  const allTransactions = txData?.data ?? []
  const txMeta = txData?.meta

  const { data: holdsData } = useGetMyWalletHolds()
  const holds = holdsData?.data ?? []

  const topupMutation = usePostMyWalletTopup()

  const transactions = allTransactions.filter((tx) => {
    if (dateRange?.[0] && dateRange?.[1] && tx.created_at) {
      const txDate = dayjs(tx.created_at)
      if (txDate.isBefore(dateRange[0], 'day') || txDate.isAfter(dateRange[1], 'day')) {
        return false
      }
    }
    return true
  })

  const handleTopUp = (values: { amount: number }) => {
    topupMutation.mutate(
      { data: { amount: values.amount * 100 } },
      {
        onSuccess: () => {
          message.success('Пополнение инициировано')
          form.resetFields()
        },
        onError: () => {
          message.error('Ошибка при пополнении')
        },
      },
    )
  }

  const handleExport = useCallback(async (format: 'csv' | 'pdf') => {
    setExporting(true)
    try {
      const params = new URLSearchParams({ format })
      if (dateRange?.[0]) params.set('date_from', dateRange[0].format('YYYY-MM-DD'))
      if (dateRange?.[1]) params.set('date_to', dateRange[1].format('YYYY-MM-DD'))

      const token = localStorage.getItem(AUTH_TOKEN_KEY)
      const response = await fetch(`/api/v1/my/wallet/export?${params}`, {
        headers: { Authorization: `Bearer ${token}` },
      })

      if (!response.ok) throw new Error('Export failed')

      const blob = await response.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `wallet_transactions.${format}`
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      message.error('Ошибка при экспорте')
    } finally {
      setExporting(false)
    }
  }, [dateRange, message])

  const txColumns: ColumnsType<InternalHandlerWalletTransactionResponse> = [
    {
      title: 'Дата',
      key: 'created_at',
      render: (_, record) => record.created_at ? formatDateTime(record.created_at, 'DD.MM.YYYY HH:mm') : '—',
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
        const isPositive = record.type !== 'spend'
        return (
          <span style={{ color: isPositive ? '#52c41a' : '#ff4d4f', fontWeight: 500 }}>
            {isPositive ? '+' : ''}{formatPrice(amount ?? 0)}
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
      render: (val: number) => val != null ? formatPrice(val) : '—',
      responsive: ['md'] as const,
    },
    {
      title: 'Бронирование',
      key: 'reference',
      render: (_, record) =>
        record.reference_id && record.reference_type === 'booking' ? (
          <a onClick={() => navigate(`/client/bookings/${record.reference_id}`)}>
            Открыть
          </a>
        ) : (
          '—'
        ),
      responsive: ['lg'] as const,
    },
  ]

  const holdColumns: ColumnsType<InternalHandlerWalletHoldResponse> = [
    {
      title: 'Сумма',
      dataIndex: 'amount',
      key: 'amount',
      render: (amount: number) => formatPrice(amount ?? 0),
    },
    {
      title: 'Описание',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: 'Истекает',
      key: 'expires_at',
      render: (_, record) => record.expires_at ? formatDateTime(record.expires_at, 'DD.MM.YYYY HH:mm') : '—',
    },
  ]

  const expiringBonuses = transactions.filter(
    (tx) => tx.is_bonus && tx.expires_at && dayjs(tx.expires_at).diff(dayjs(), 'day') <= 30 && dayjs(tx.expires_at).isAfter(dayjs()),
  )

  return (
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>Кошелёк</Title>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={8}>
          <Card loading={walletLoading}>
            <Statistic
              title="Баланс"
              value={wallet?.balance ? wallet.balance / 100 : 0}
              suffix="₽"
              prefix={<WalletOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8}>
          <Card loading={walletLoading}>
            <Statistic
              title="Доступно"
              value={wallet?.available ? wallet.available / 100 : 0}
              suffix="₽"
              styles={{ content: { color: '#52c41a' } }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8}>
          <Card loading={walletLoading}>
            <Statistic
              title="Заморожено"
              value={wallet?.held_amount ? wallet.held_amount / 100 : 0}
              suffix="₽"
              prefix={<LockOutlined />}
              styles={{ content: { color: '#faad14' } }}
            />
          </Card>
        </Col>
      </Row>

      {wallet?.expiring_soon != null && wallet.expiring_soon > 0 && (
        <Alert
          type="warning"
          showIcon
          icon={<WarningOutlined />}
          message={`Бонусы на сумму ${formatPrice(wallet.expiring_soon)} скоро сгорят${wallet.earliest_expiry ? ` (до ${formatDateTime(wallet.earliest_expiry, 'DD.MM.YYYY')})` : ''}`}
          style={{ marginTop: 16 }}
        />
      )}

      {expiringBonuses.length > 0 && (
        <Card title="Бонусы, истекающие в ближайшие 30 дней" size="small" style={{ marginTop: 16 }}>
          {expiringBonuses.map((bonus) => (
            <div key={bonus.id} style={{ display: 'flex', justifyContent: 'space-between', padding: '4px 0' }}>
              <Text>{bonus.description ?? 'Бонус'}</Text>
              <Space>
                <Text strong>{formatPrice(bonus.amount ?? 0)}</Text>
                <Text type="secondary">до {bonus.expires_at ? formatDateTime(bonus.expires_at, 'DD.MM.YYYY') : '—'}</Text>
              </Space>
            </div>
          ))}
        </Card>
      )}

      {holds.length > 0 && (
        <Card title="Замороженные средства" size="small" style={{ marginTop: 16 }}>
          <Table
            columns={holdColumns}
            dataSource={holds}
            rowKey="id"
            pagination={false}
            size="small"
          />
        </Card>
      )}

      <Card title="Пополнить кошелёк" style={{ marginTop: 16 }}>
        <Form
          form={form}
          layout="inline"
          onFinish={handleTopUp}
        >
          <Form.Item
            name="amount"
            rules={[
              { required: true, message: 'Введите сумму' },
              { type: 'number', min: MIN_TOPUP, message: `Минимум ${MIN_TOPUP} ₽` },
              { type: 'number', max: MAX_TOPUP, message: `Максимум ${MAX_TOPUP} ₽` },
            ]}
          >
            <InputNumber
              placeholder="Сумма, ₽"
              min={MIN_TOPUP}
              max={MAX_TOPUP}
              style={{ width: 200 }}
              addonAfter="₽"
            />
          </Form.Item>
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              icon={<PlusOutlined />}
              loading={topupMutation.isPending}
            >
              Пополнить
            </Button>
          </Form.Item>
        </Form>
      </Card>

      <Title level={4} style={{ marginTop: 24, marginBottom: 12 }}>История операций</Title>

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
        <Dropdown
          menu={{
            items: [
              { key: 'csv', label: 'Экспорт CSV', onClick: () => handleExport('csv') },
              { key: 'pdf', label: 'Экспорт PDF', onClick: () => handleExport('pdf') },
            ],
          }}
        >
          <Button icon={<DownloadOutlined />} loading={exporting}>
            Экспорт
          </Button>
        </Dropdown>
      </Space>

      <Table
        columns={txColumns}
        dataSource={transactions}
        rowKey="id"
        loading={txLoading}
        locale={{ emptyText: <EmptyState description="Нет операций по кошельку" /> }}
        pagination={dateRange?.[0] && dateRange?.[1] ? false : {
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
