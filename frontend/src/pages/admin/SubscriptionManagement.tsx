import { useState } from 'react'
import {
  Alert,
  App,
  Button,
  Card,
  Input,
  Select,
  Spin,
  Table,
  Tag,
  Typography,
} from '@/components/design/system'
import { ReloadOutlined, StopOutlined, PlayCircleOutlined } from '@/components/design/icons'
import type { ColumnsType } from '@/components/design/types'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetMySubscriptions,
  getGetMySubscriptionsQueryKey,
  useDeleteMyBathhousesIdSubscription,
  usePostMyBathhousesIdSubscription,
} from '@/api/generated/subscriptions/subscriptions'
import type { InternalHandlerSubscriptionResponse } from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography
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
    <div className="rh-stack rh-admin-reference-page">
      <PageHeader
        eyebrow="Монетизация"
        title="Управление подписками"
        description="Тарифы, статусы и ручное управление подписками в едином админском списке."
        extra={
        <Button icon={<ReloadOutlined />} onClick={invalidate}>
          Обновить
        </Button>
        }
      />

      <Alert
        type="warning"
        showIcon
        title="Раздел в разработке"
        description="Сейчас страница показывает только подписки, привязанные к текущему админу (через /api/v1/my/subscriptions). Платформенный admin-эндпоинт со списком всех подписок ещё не реализован — найдено в аудите A1.7."
        className="rh-admin-inline-alert"
      />

      <div className="rh-stat-grid">
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Активных</span>
          <span className="rh-stat-tile__value">{activeCount}</span>
          <span className="rh-stat-tile__hint">Подписки в рабочем статусе.</span>
        </div>
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Премиум</span>
          <span className="rh-stat-tile__value">{premiumCount}</span>
          <span className="rh-stat-tile__hint">Активные премиальные тарифы.</span>
        </div>
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Продвинутых</span>
          <span className="rh-stat-tile__value">{promotedCount}</span>
          <span className="rh-stat-tile__hint">Активные promoted-размещения.</span>
        </div>
      </div>

      <Card className="rh-admin-filter-card" title="Фильтры">
        <div className="rh-admin-filter-row rh-admin-filter-row--native">
          <Search
            className="rh-admin-filter-input"
            placeholder="Поиск по ID бани или владельца"
            allowClear
            onSearch={setSearchText}
            onChange={(e) => !e.target.value && setSearchText('')}
          />
          <Select
            className="rh-admin-filter-select"
            placeholder="Статус"
            allowClear
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
      </Card>

      <Card className="rh-admin-reference-card" title="Список подписок">
        {isLoading ? (
          <div className="rh-admin-state-card">
            <Spin size="large" />
            <span>Загружаем подписки</span>
          </div>
        ) : filteredSubscriptions.length === 0 ? (
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Нет подписок</div>
            <p className="rh-admin-empty-state__text">
              Измените фильтры или обновите список, чтобы увидеть доступные подписки.
            </p>
          </div>
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
      </Card>
    </div>
  )
}
