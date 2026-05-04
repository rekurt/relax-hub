import { useState } from 'react'
import {
  App,
  Button,
  Form,
  InputNumber,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  Typography,
} from '@/components/design/system'
import {
  BankOutlined,
  SendOutlined,
  ThunderboltOutlined,
} from '@/components/design/icons'
import type { ColumnsType } from '@/components/design/types'
import {
  useGetMyWallet,
  useGetMyWalletPayouts,
  usePostMyWalletPayout,
  usePutMyWalletAutoPayout,
} from '@/api/generated/wallet/wallet'
import type { InternalHandlerPayoutResponse } from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'
import EmptyState from '@/components/EmptyState'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

const PAYOUT_STATUS_CONFIG: Record<string, { color: string; text: string }> = {
  pending: { color: 'orange', text: 'В обработке' },
  processing: { color: 'blue', text: 'Выполняется' },
  completed: { color: 'green', text: 'Выполнена' },
  failed: { color: 'red', text: 'Ошибка' },
  cancelled: { color: 'default', text: 'Отменена' },
}

const PAYOUT_METHOD_OPTIONS = [
  { value: 'sbp', label: 'СБП (мгновенный)' },
  { value: 'bank_transfer', label: 'Банковский перевод' },
]

const DAILY_LIMIT = 10000000
const MONTHLY_LIMIT = 100000000

export default function PayoutPage() {
  const { message } = App.useApp()
  const [form] = Form.useForm()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [autoPayoutEnabled, setAutoPayoutEnabled] = useState(false)

  const { data: walletData, isLoading: walletLoading } = useGetMyWallet()
  const wallet = walletData?.data

  const { data: payoutsData, isLoading: payoutsLoading } = useGetMyWalletPayouts({
    page,
    page_size: pageSize,
  })
  const payouts = payoutsData?.data ?? []
  const payoutsMeta = payoutsData?.meta

  const payoutMutation = usePostMyWalletPayout()
  const autoPayoutMutation = usePutMyWalletAutoPayout()

  const handlePayout = (values: { amount: number; method: string }) => {
    payoutMutation.mutate(
      { data: { ...{ amount: Math.round(values.amount * 100) }, payout_method: values.method } as Parameters<typeof payoutMutation.mutate>[0]['data'] },
      {
        onSuccess: () => {
          message.success('Заявка на выплату создана')
          form.resetFields()
        },
        onError: () => {
          message.error('Ошибка при создании заявки на выплату')
        },
      },
    )
  }

  const handleAutoPayoutToggle = (checked: boolean) => {
    const threshold = checked ? 500000 : 0
    autoPayoutMutation.mutate(
      { data: { threshold } },
      {
        onSuccess: () => {
          setAutoPayoutEnabled(checked)
          message.success(checked ? 'Автовыплата включена' : 'Автовыплата отключена')
        },
        onError: () => {
          message.error('Ошибка при настройке автовыплаты')
        },
      },
    )
  }

  const payoutColumns: ColumnsType<InternalHandlerPayoutResponse> = [
    {
      title: 'Дата',
      key: 'requested_at',
      render: (_, record) => record.requested_at ? formatDateTime(record.requested_at, 'DD.MM.YYYY HH:mm') : '\u2014',
      sorter: (a, b) => (a.requested_at ?? '').localeCompare(b.requested_at ?? ''),
    },
    {
      title: 'Сумма',
      dataIndex: 'amount',
      key: 'amount',
      render: (amount: number) => formatPrice(amount ?? 0),
    },
    {
      title: 'Метод',
      dataIndex: 'payout_method',
      key: 'payout_method',
      render: (method: string) => {
        const label = PAYOUT_METHOD_OPTIONS.find(o => o.value === method)?.label ?? method
        return <Tag icon={method === 'sbp' ? <ThunderboltOutlined /> : <BankOutlined />}>{label}</Tag>
      },
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const config = PAYOUT_STATUS_CONFIG[status]
        return <Tag color={config?.color ?? 'default'}>{config?.text ?? status}</Tag>
      },
    },
    {
      title: 'Выполнена',
      key: 'processed_at',
      render: (_, record) => record.processed_at ? formatDateTime(record.processed_at, 'DD.MM.YYYY HH:mm') : '\u2014',
      responsive: ['md'] as const,
    },
    {
      title: 'Причина ошибки',
      dataIndex: 'failure_reason',
      key: 'failure_reason',
      ellipsis: true,
      responsive: ['lg'] as const,
      render: (val: string) => val || '\u2014',
    },
  ]

  const availableRubles = wallet?.available ? wallet.available / 100 : 0
  const payoutMetricTiles = [
    {
      label: 'Доступно к выводу',
      value: formatPrice(wallet?.available ?? 0),
      hint: 'Средства, которые можно отправить на выплату сейчас',
      icon: <SendOutlined />,
      loading: walletLoading,
    },
    {
      label: 'Дневной лимит',
      value: formatPrice(DAILY_LIMIT),
      hint: 'Максимум в день',
      icon: <ThunderboltOutlined />,
      loading: false,
    },
    {
      label: 'Месячный лимит',
      value: formatPrice(MONTHLY_LIMIT),
      hint: 'Максимум в месяц',
      icon: <BankOutlined />,
      loading: false,
    },
  ]

  return (
    <div className="rh-stack">
      <PageHeader
        size="compact"
        eyebrow="Финансы"
        title="Выплаты"
        description="Запросы на вывод, лимиты и история выплат владельца в одном рабочем экране."
      />

      <div className="rh-admin-metric-grid">
        {payoutMetricTiles.map((tile) => (
          <div className="rh-admin-metric" key={tile.label} aria-busy={tile.loading}>
            <div className="rh-admin-metric__head">
              <span className="rh-admin-metric__label">{tile.label}</span>
              <span className="rh-admin-metric__icon">{tile.icon}</span>
            </div>
            <div className="rh-admin-metric__value">{tile.value}</div>
            <div className="rh-admin-metric__hint">{tile.hint}</div>
          </div>
        ))}
      </div>

      <section className="rh-admin-panel">
        <div className="rh-admin-toolbar rh-admin-toolbar--spaced">
          <div className="rh-admin-toolbar__copy">
            <h2 className="rh-admin-toolbar__title">Запросить выплату</h2>
            <div className="rh-admin-toolbar__hint">Сумма не может превышать доступный баланс: {formatPrice(wallet?.available ?? 0)}.</div>
          </div>
        </div>
        <Form
          form={form}
          layout="inline"
          onFinish={handlePayout}
          initialValues={{ method: 'sbp' }}
        >
          <Form.Item
            name="amount"
            rules={[
              { required: true, message: 'Введите сумму' },
              { type: 'number', min: 100, message: 'Минимум 100 \u20BD' },
            ]}
          >
            <InputNumber
              placeholder="Сумма, \u20BD"
              min={100}
              max={availableRubles}
              className="rh-payout-amount-control"
            />
          </Form.Item>
          <Form.Item
            name="method"
            rules={[{ required: true, message: 'Выберите метод' }]}
          >
            <Select
              options={PAYOUT_METHOD_OPTIONS}
              className="rh-payout-method-control"
            />
          </Form.Item>
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              icon={<SendOutlined />}
              loading={payoutMutation.isPending}
            >
              Вывести
            </Button>
          </Form.Item>
        </Form>
      </section>

      <section className="rh-admin-panel">
        <div className="rh-admin-toolbar rh-admin-toolbar--compact-spaced">
          <div className="rh-admin-toolbar__copy">
            <h2 className="rh-admin-toolbar__title">Автовыплата</h2>
            <div className="rh-admin-toolbar__hint">Автоматический вывод при достижении заданного порога.</div>
          </div>
        </div>
        <Space>
          <Switch
            checked={autoPayoutEnabled}
            onChange={handleAutoPayoutToggle}
            loading={autoPayoutMutation.isPending}
          />
          <Text>Автоматически выводить при достижении порога (5 000 \u20BD)</Text>
        </Space>
      </section>

      <section className="rh-admin-table-card">
        <div className="rh-admin-toolbar">
          <div className="rh-admin-toolbar__copy">
            <h2 className="rh-admin-toolbar__title">История выплат</h2>
            <div className="rh-admin-toolbar__hint">Статусы, метод и причина ошибки по всем заявкам.</div>
          </div>
        </div>
        <Table
          columns={payoutColumns}
          dataSource={payouts}
          rowKey="id"
          loading={payoutsLoading}
          locale={{ emptyText: <EmptyState description="Нет выплат" /> }}
          pagination={{
            current: page,
            pageSize: pageSize,
            total: payoutsMeta?.total_count ?? 0,
            showSizeChanger: true,
            showTotal: (total) => `Всего: ${total}`,
            onChange: (p, ps) => {
              setPage(p)
              setPageSize(ps)
            },
          }}
        />
      </section>
    </div>
  )
}
