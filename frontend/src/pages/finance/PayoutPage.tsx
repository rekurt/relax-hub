import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Col,
  Form,
  InputNumber,
  Row,
  Select,
  Space,
  Statistic,
  Switch,
  Table,
  Tag,
  Typography,
} from 'antd'
import {
  BankOutlined,
  SendOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  useGetMyWallet,
  useGetMyWalletPayouts,
  usePostMyWalletPayout,
  usePutMyWalletAutoPayout,
} from '@/api/generated/wallet/wallet'
import type { InternalHandlerPayoutResponse } from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'
import EmptyState from '@/components/EmptyState'

const { Title, Text } = Typography

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
      { data: { ...{ amount: values.amount * 100 }, payout_method: values.method } as Parameters<typeof payoutMutation.mutate>[0]['data'] },
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

  return (
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>Выплаты</Title>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={8}>
          <Card loading={walletLoading}>
            <Statistic
              title="Доступно к выводу"
              value={availableRubles}
              suffix="\u20BD"
              styles={{ content: { color: '#52c41a' } }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8}>
          <Card>
            <Statistic
              title="Дневной лимит"
              value={DAILY_LIMIT / 100}
              suffix="\u20BD"
            />
            <Text type="secondary" style={{ fontSize: 12 }}>Максимум в день</Text>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8}>
          <Card>
            <Statistic
              title="Месячный лимит"
              value={MONTHLY_LIMIT / 100}
              suffix="\u20BD"
            />
            <Text type="secondary" style={{ fontSize: 12 }}>Максимум в месяц</Text>
          </Card>
        </Col>
      </Row>

      <Card title="Запросить выплату" style={{ marginTop: 16 }}>
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
              style={{ width: 200 }}
              addonAfter="\u20BD"
            />
          </Form.Item>
          <Form.Item
            name="method"
            rules={[{ required: true, message: 'Выберите метод' }]}
          >
            <Select
              options={PAYOUT_METHOD_OPTIONS}
              style={{ width: 220 }}
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
      </Card>

      <Card title="Автовыплата" size="small" style={{ marginTop: 16 }}>
        <Space>
          <Switch
            onChange={handleAutoPayoutToggle}
            loading={autoPayoutMutation.isPending}
          />
          <Text>Автоматически выводить при достижении порога (5 000 \u20BD)</Text>
        </Space>
      </Card>

      <Title level={4} style={{ marginTop: 24, marginBottom: 12 }}>История выплат</Title>

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
    </div>
  )
}
