import { useState } from 'react'
import {
  Alert,
  App,
  Button,
  Card,
  Col,
  Empty,
  Input,
  Row,
  Select,
  Spin,
  Table,
  Tag,
  Typography,
} from 'antd'
import { ReloadOutlined, StopOutlined, PlayCircleOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetMySubscriptions,
  getGetMySubscriptionsQueryKey,
  useDeleteMyBathhousesIdSubscription,
  usePostMyBathhousesIdSubscription,
} from '@/api/generated/subscriptions/subscriptions'
import type { InternalHandlerSubscriptionResponse } from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'

const { Title, Text } = Typography
const { Search } = Input

const STATUS_MAP: Record<string, { label: string; color: string }> = {
  active: { label: 'Активна', color: 'green' },
  expired: { label: 'Истекла', color: 'default' },
  cancelled: { label: 'Отменена', color: 'red' },
  pending: { label: 'Ожидание', color: 'orange' },
}

const PLAN_MAP: Record<string, { label: string; color: string }> = {
  free: { label: 'Бесплатный', color: 'default' },
  premium: { label: 'Премиум', color: 'blue' },
  promoted: { label: 'Продвинутый', color: 'gold' },
}

export default function SubscriptionManagement() {
  const { message, modal } = App.useApp()
  const queryClient = useQueryClient()

  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState<string | undefined>(undefined)
  const [searchText, setSearchText] = useState('')

  const { data, isLoading } = useGetMySubscriptions({ page, page_size: 20 })
  const deactivateMutation = useDeleteMyBathhousesIdSubscription()
  const activateMutation = usePostMyBathhousesIdSubscription()

  const subscriptions: InternalHandlerSubscriptionResponse[] = data?.data ?? []
  const totalCount = data?.meta?.total_count ?? 0

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetMySubscriptionsQueryKey() })
  }

  const filteredSubscriptions = subscriptions.filter((sub) => {
    if (statusFilter && sub.status !== statusFilter) return false
    if (searchText) {
      const search = searchText.toLowerCase()
      return (
        sub.bathhouse_id?.toLowerCase().includes(search) ||
        sub.owner_id?.toLowerCase().includes(search) ||
        sub.id?.toLowerCase().includes(search)
      )
    }
    return true
  })

  const handleDeactivate = (sub: InternalHandlerSubscriptionResponse) => {
    modal.confirm({
      title: 'Деактивировать подписку?',
      content: `Подписка ${sub.plan} для бани ${sub.bathhouse_id} будет деактивирована.`,
      okText: 'Деактивировать',
      okType: 'danger',
      cancelText: 'Отмена',
      onOk: async () => {
        if (!sub.bathhouse_id) return
        await deactivateMutation.mutateAsync({ id: sub.bathhouse_id })
        message.success('Подписка деактивирована')
        invalidate()
      },
    })
  }

  const handleActivate = (sub: InternalHandlerSubscriptionResponse) => {
    modal.confirm({
      title: 'Активировать подписку?',
      content: `Подписка ${sub.plan} для бани ${sub.bathhouse_id} будет активирована.`,
      okText: 'Активировать',
      cancelText: 'Отмена',
      onOk: async () => {
        if (!sub.bathhouse_id) return
        await activateMutation.mutateAsync({
          id: sub.bathhouse_id,
          data: { plan: sub.plan ?? 'premium' },
        })
        message.success('Подписка активирована')
        invalidate()
      },
    })
  }

  const activeCount = subscriptions.filter((s) => s.status === 'active').length
  const premiumCount = subscriptions.filter((s) => s.plan === 'premium' && s.status === 'active').length
  const promotedCount = subscriptions.filter((s) => s.plan === 'promoted' && s.status === 'active').length

  const columns: ColumnsType<InternalHandlerSubscriptionResponse> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 100,
      ellipsis: true,
      render: (id: string) => <Text copyable={{ text: id }}>{id?.slice(0, 8)}...</Text>,
    },
    {
      title: 'Баня',
      dataIndex: 'bathhouse_id',
      key: 'bathhouse_id',
      width: 120,
      ellipsis: true,
      render: (id: string) => id?.slice(0, 8) + '...',
    },
    {
      title: 'Владелец',
      dataIndex: 'owner_id',
      key: 'owner_id',
      width: 120,
      ellipsis: true,
      render: (id: string) => id?.slice(0, 8) + '...',
    },
    {
      title: 'Тариф',
      dataIndex: 'plan',
      key: 'plan',
      width: 130,
      render: (plan: string) => {
        const info = PLAN_MAP[plan] ?? { label: plan, color: 'default' }
        return <Tag color={info.color}>{info.label}</Tag>
      },
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: string) => {
        const info = STATUS_MAP[status] ?? { label: status, color: 'default' }
        return <Tag color={info.color}>{info.label}</Tag>
      },
    },
    {
      title: 'Цена',
      dataIndex: 'price_kopecks',
      key: 'price_kopecks',
      width: 100,
      render: (val: number) => (val ? formatPrice(val) : '—'),
    },
    {
      title: 'Начало',
      dataIndex: 'start_date',
      key: 'start_date',
      width: 130,
      responsive: ['lg'] as const,
      render: (val: string) => (val ? formatDateTime(val) : '—'),
    },
    {
      title: 'Окончание',
      dataIndex: 'end_date',
      key: 'end_date',
      width: 130,
      responsive: ['lg'] as const,
      render: (val: string) => (val ? formatDateTime(val) : '—'),
    },
    {
      title: 'Автопродление',
      dataIndex: 'auto_renew',
      key: 'auto_renew',
      width: 120,
      render: (val: boolean) => (val ? <Tag color="blue">Да</Tag> : <Tag>Нет</Tag>),
    },
    {
      title: '',
      key: 'actions',
      width: 140,
      render: (_: unknown, record: InternalHandlerSubscriptionResponse) =>
        record.status === 'active' ? (
          <Button
            type="link"
            danger
            size="small"
            icon={<StopOutlined />}
            onClick={() => handleDeactivate(record)}
          >
            Деактивировать
          </Button>
        ) : (
          <Button
            type="link"
            size="small"
            icon={<PlayCircleOutlined />}
            onClick={() => handleActivate(record)}
          >
            Активировать
          </Button>
        ),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>
          Управление подписками
        </Title>
        <Button icon={<ReloadOutlined />} onClick={invalidate}>
          Обновить
        </Button>
      </div>

      <Alert
        type="warning"
        showIcon
        message="Раздел в разработке"
        description="Сейчас страница показывает только подписки, привязанные к текущему админу (через /api/v1/my/subscriptions). Платформенный admin-эндпоинт со списком всех подписок ещё не реализован — найдено в аудите A1.7."
        style={{ marginBottom: 16 }}
      />

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={8} sm={8}>
          <Card size="small">
            <Text type="secondary">Активных</Text>
            <div style={{ fontSize: 24, fontWeight: 600 }}>{activeCount}</div>
          </Card>
        </Col>
        <Col xs={8} sm={8}>
          <Card size="small">
            <Text type="secondary">Премиум</Text>
            <div style={{ fontSize: 24, fontWeight: 600, color: '#1677ff' }}>{premiumCount}</div>
          </Card>
        </Col>
        <Col xs={8} sm={8}>
          <Card size="small">
            <Text type="secondary">Продвинутых</Text>
            <div style={{ fontSize: 24, fontWeight: 600, color: '#faad14' }}>{promotedCount}</div>
          </Card>
        </Col>
      </Row>

      <div style={{ display: 'flex', gap: 12, marginBottom: 16, flexWrap: 'wrap' }}>
        <Search
          placeholder="Поиск по ID бани или владельца"
          allowClear
          style={{ width: 280 }}
          onSearch={setSearchText}
          onChange={(e) => !e.target.value && setSearchText('')}
        />
        <Select
          placeholder="Статус"
          allowClear
          style={{ width: 160 }}
          value={statusFilter}
          onChange={setStatusFilter}
          options={[
            { value: 'active', label: 'Активна' },
            { value: 'expired', label: 'Истекла' },
            { value: 'cancelled', label: 'Отменена' },
            { value: 'pending', label: 'Ожидание' },
          ]}
        />
      </div>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: 48 }}><Spin size="large" /></div>
      ) : filteredSubscriptions.length === 0 ? (
        <Empty description="Нет подписок" />
      ) : (
        <Table
          dataSource={filteredSubscriptions}
          columns={columns}
          rowKey="id"
          size="middle"
          scroll={{ x: 900 }}
          pagination={{
            current: page,
            pageSize: 20,
            total: totalCount,
            onChange: setPage,
            showTotal: (total) => `Всего: ${total}`,
          }}
          locale={{ emptyText: 'Нет подписок' }}
        />
      )}
    </div>
  )
}
