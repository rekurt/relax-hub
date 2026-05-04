import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Col,
  Descriptions,
  Form,
  InputNumber,
  Modal,
  Popconfirm,
  Row,
  Select,
  Space,
  Statistic,
  Table,
  Tag,
  Typography,
} from '@/components/design/system'
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  CrownOutlined,
  RocketOutlined,
  StarOutlined,
} from '@/components/design/icons'
import dayjs from 'dayjs'
import {
  useGetMyBathhousesIdSubscription,
  usePostMyBathhousesIdSubscription,
  useDeleteMyBathhousesIdSubscription,
  useGetMySubscriptions,
  useGetMyBathhousesIdPromotion,
  usePostMyBathhousesIdPromotion,
} from '@/api/generated/subscriptions/subscriptions'
import { useGetCities } from '@/api/generated/cities/cities'
import type {
  InternalHandlerSubscriptionResponse,
  InternalHandlerPromotionResponse,
} from '@/api/generated/model'
import { useBathhouseStore } from '@/stores/bathhouse'
import { useQueryClient } from '@tanstack/react-query'
import { formatPrice } from '@/lib/format'
import EmptyState from '@/components/EmptyState'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

const PLANS = [
  {
    key: 'free',
    title: 'Бесплатный',
    icon: <StarOutlined className="rh-subscription-plan-icon" />,
    features: ['Базовый листинг', 'До 5 фото', 'Стандартный поиск'],
    price: 0,
  },
  {
    key: 'premium',
    title: 'Премиум',
    icon: <CrownOutlined className="rh-subscription-plan-icon" />,
    features: ['Приоритет в поиске (+10)', 'До 20 фото', 'Аналитика', 'Виджет бронирования'],
    price: 99900,
  },
  {
    key: 'promoted',
    title: 'Продвижение',
    icon: <RocketOutlined className="rh-subscription-plan-icon" />,
    features: ['Всё из Премиум', 'Первые позиции', 'Промо-кампании', 'Персональный менеджер'],
    price: 299900,
  },
]

const PLAN_LABELS: Record<string, string> = {
  free: 'Бесплатный',
  premium: 'Премиум',
  promoted: 'Продвижение',
}

const PLAN_COLORS: Record<string, string> = {
  free: 'default',
  premium: 'gold',
  promoted: 'blue',
}

const STATUS_LABELS: Record<string, string> = {
  active: 'Активна',
  cancelled: 'Отменена',
  expired: 'Истекла',
  pending: 'Ожидает',
}

const STATUS_COLORS: Record<string, string> = {
  active: 'green',
  cancelled: 'orange',
  expired: 'red',
  pending: 'processing',
}

interface PromotionFormValues {
  budget: number
  duration_days: number
  target_city_id?: number
}

export default function SubscriptionPage() {
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)
  const queryClient = useQueryClient()
  const { message } = App.useApp()
  const [promoForm] = Form.useForm<PromotionFormValues>()

  const [promoModalOpen, setPromoModalOpen] = useState(false)
  const [subsPage, setSubsPage] = useState(1)
  const subsPageSize = 10

  const { data: subscriptionData, isLoading: subscriptionLoading } =
    useGetMyBathhousesIdSubscription(selectedBathhouseId ?? '', {
      query: { enabled: !!selectedBathhouseId },
    })

  const { data: allSubscriptionsData, isLoading: allSubsLoading } = useGetMySubscriptions(
    { page: subsPage, page_size: subsPageSize },
  )

  const { data: promotionData } = useGetMyBathhousesIdPromotion(
    selectedBathhouseId ?? '',
    { query: { enabled: !!selectedBathhouseId } },
  )

  const { data: citiesData } = useGetCities()
  const cities = (citiesData?.data ?? []) as { id?: number; name?: string }[]

  const currentSubscription = subscriptionData?.data as InternalHandlerSubscriptionResponse | undefined
  const allSubscriptions = (allSubscriptionsData?.data ?? []) as InternalHandlerSubscriptionResponse[]
  const totalSubs = allSubscriptionsData?.meta?.total_count ?? 0
  const promotion = promotionData?.data as InternalHandlerPromotionResponse | undefined

  const invalidateSubscription = () => {
    queryClient.invalidateQueries({
      queryKey: [`/my/bathhouses/${selectedBathhouseId}/subscription`],
    })
    queryClient.invalidateQueries({ queryKey: ['/my/subscriptions'] })
  }

  const invalidatePromotion = () => {
    queryClient.invalidateQueries({
      queryKey: [`/my/bathhouses/${selectedBathhouseId}/promotion`],
    })
  }

  const subscribeMutation = usePostMyBathhousesIdSubscription({
    mutation: {
      onSuccess: () => {
        message.success('Подписка оформлена')
        invalidateSubscription()
      },
      onError: () => message.error('Не удалось оформить подписку'),
    },
  })

  const cancelMutation = useDeleteMyBathhousesIdSubscription({
    mutation: {
      onSuccess: () => {
        message.success('Автопродление отменено')
        invalidateSubscription()
      },
      onError: () => message.error('Не удалось отменить подписку'),
    },
  })

  const promotionMutation = usePostMyBathhousesIdPromotion({
    mutation: {
      onSuccess: () => {
        message.success('Промо-кампания создана')
        setPromoModalOpen(false)
        promoForm.resetFields()
        invalidatePromotion()
      },
      onError: () => message.error('Не удалось создать кампанию'),
    },
  })

  const handleSubscribe = (plan: string) => {
    if (!selectedBathhouseId) return
    subscribeMutation.mutate({ id: selectedBathhouseId, data: { plan } })
  }

  const handleCancelSubscription = () => {
    if (!selectedBathhouseId) return
    cancelMutation.mutate({ id: selectedBathhouseId })
  }

  const handleCreatePromotion = (values: PromotionFormValues) => {
    if (!selectedBathhouseId) return
    promotionMutation.mutate({
      id: selectedBathhouseId,
      data: {
        budget_kopecks: Math.round(values.budget * 100),
        duration_days: values.duration_days,
        target_city_id: values.target_city_id,
      },
    })
  }

  const subsColumns = [
    {
      title: 'Баня',
      dataIndex: 'bathhouse_id',
      key: 'bathhouse_id',
      ellipsis: true,
      render: (id: string) => <Text copyable={{ text: id }}>{id.slice(0, 8)}...</Text>,
    },
    {
      title: 'План',
      dataIndex: 'plan',
      key: 'plan',
      render: (plan: string) => (
        <Tag color={PLAN_COLORS[plan] ?? 'default'}>{PLAN_LABELS[plan] ?? plan}</Tag>
      ),
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={STATUS_COLORS[status] ?? 'default'}>{STATUS_LABELS[status] ?? status}</Tag>
      ),
    },
    {
      title: 'Стоимость',
      dataIndex: 'price_kopecks',
      key: 'price_kopecks',
      render: (v: number) => v ? formatPrice(v) : 'Бесплатно',
    },
    {
      title: 'Период',
      key: 'period',
      render: (_: unknown, record: InternalHandlerSubscriptionResponse) => {
        const from = record.start_date ? dayjs(record.start_date).format('DD.MM.YYYY') : '—'
        const to = record.end_date ? dayjs(record.end_date).format('DD.MM.YYYY') : '—'
        return `${from} – ${to}`
      },
    },
    {
      title: 'Автопродление',
      dataIndex: 'auto_renew',
      key: 'auto_renew',
      render: (val: boolean) =>
        val ? (
          <Tag icon={<CheckCircleOutlined />} color="green">Да</Tag>
        ) : (
          <Tag icon={<CloseCircleOutlined />} color="default">Нет</Tag>
        ),
    },
  ]

  if (!selectedBathhouseId) {
    return (
      <div className="rh-page-stack">
        <PageHeader
          title="Подписки"
          description="Тарифные планы, промо-кампании и история подписок объекта."
          size="compact"
        />
        <Card className="rh-admin-detail-card">
          <EmptyState description="Выберите баню для управления подписками" />
        </Card>
      </div>
    )
  }

  return (
    <div className="rh-page-stack">
      <PageHeader
        title="Подписки"
        description="Управляйте тарифом, продвижением и автопродлением выбранного объекта."
        size="compact"
      />

      {/* Current subscription */}
      {currentSubscription && currentSubscription.plan !== 'free' && (
        <Card
          title="Текущая подписка"
          className="rh-admin-detail-card"
          loading={subscriptionLoading}
          extra={
            currentSubscription.auto_renew && (
              <Popconfirm
                title="Отменить автопродление?"
                description="Подписка будет активна до конца оплаченного периода"
                onConfirm={handleCancelSubscription}
                okText="Отменить"
                cancelText="Назад"
              >
                <Button danger loading={cancelMutation.isPending}>
                  Отменить подписку
                </Button>
              </Popconfirm>
            )
          }
        >
          <Descriptions column={{ xs: 1, sm: 2, md: 3 }}>
            <Descriptions.Item label="План">
              <Tag color={PLAN_COLORS[currentSubscription.plan ?? ''] ?? 'default'}>
                {PLAN_LABELS[currentSubscription.plan ?? ''] ?? currentSubscription.plan}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Статус">
              <Tag color={STATUS_COLORS[currentSubscription.status ?? ''] ?? 'default'}>
                {STATUS_LABELS[currentSubscription.status ?? ''] ?? currentSubscription.status}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Стоимость">
              {currentSubscription.price_kopecks ? formatPrice(currentSubscription.price_kopecks) : 'Бесплатно'}
            </Descriptions.Item>
            <Descriptions.Item label="Начало">
              {currentSubscription.start_date ? dayjs(currentSubscription.start_date).format('DD.MM.YYYY') : '—'}
            </Descriptions.Item>
            <Descriptions.Item label="Окончание">
              {currentSubscription.end_date ? dayjs(currentSubscription.end_date).format('DD.MM.YYYY') : '—'}
            </Descriptions.Item>
            <Descriptions.Item label="Автопродление">
              {currentSubscription.auto_renew ? 'Да' : 'Нет'}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      )}

      {/* Plans */}
      <h2 className="rh-inline-title">Выбор плана</h2>
      <Row gutter={[16, 16]} className="rh-subscription-plan-grid">
        {PLANS.map((plan) => {
          const isCurrent = currentSubscription?.plan === plan.key
          return (
            <Col xs={24} sm={8} key={plan.key}>
              <Card
                hoverable={!isCurrent}
                className={[
                  'rh-subscription-plan-card',
                  `rh-subscription-plan-card--${plan.key}`,
                  isCurrent ? 'rh-subscription-plan-card--current' : '',
                ].filter(Boolean).join(' ')}
              >
                <div className="rh-subscription-plan-head">
                  <div className="rh-subscription-plan-icon-wrap">{plan.icon}</div>
                  <h3 className="rh-subscription-plan-title">{plan.title}</h3>
                  <div className="rh-subscription-plan-price">
                    {plan.price === 0 ? 'Бесплатно' : `${formatPrice(plan.price)}/мес`}
                  </div>
                </div>
                <ul className="rh-subscription-feature-list">
                  {plan.features.map((f) => (
                    <li key={f}>{f}</li>
                  ))}
                </ul>
                <Button
                  type={isCurrent ? 'default' : 'primary'}
                  block
                  disabled={isCurrent}
                  loading={subscribeMutation.isPending}
                  onClick={() => handleSubscribe(plan.key)}
                >
                  {isCurrent ? 'Текущий план' : 'Выбрать'}
                </Button>
              </Card>
            </Col>
          )
        })}
      </Row>

      {/* Promotions section (only for promoted plan) */}
      {currentSubscription?.plan === 'promoted' && (
        <>
          <div className="rh-section-toolbar">
            <h2 className="rh-inline-title">Промо-кампании</h2>
            <Button type="primary" onClick={() => { promoForm.resetFields(); setPromoModalOpen(true) }}>
              Создать кампанию
            </Button>
          </div>

          {promotion && (
            <Card className="rh-admin-detail-card rh-subscription-promotion-card">
              <Row gutter={16}>
                <Col xs={12} sm={6}>
                  <Statistic
                    title="Бюджет"
                    value={promotion.budget_kopecks ? formatPrice(promotion.budget_kopecks) : '—'}
                  />
                </Col>
                <Col xs={12} sm={6}>
                  <Statistic title="Потрачено" value={promotion.spent_kopecks ? formatPrice(promotion.spent_kopecks) : '0 ₽'} />
                </Col>
                <Col xs={12} sm={6}>
                  <Statistic title="Показы" value={promotion.impression_count ?? 0} />
                </Col>
                <Col xs={12} sm={6}>
                  <Statistic title="Клики" value={promotion.click_count ?? 0} />
                </Col>
              </Row>
              <div className="rh-subscription-promotion-period">
                <Text type="secondary">
                  Период: {promotion.start_date ? dayjs(promotion.start_date).format('DD.MM.YYYY') : '—'}
                  {' – '}
                  {promotion.end_date ? dayjs(promotion.end_date).format('DD.MM.YYYY') : '—'}
                </Text>
                {promotion.status && (
                  <Tag color={STATUS_COLORS[promotion.status] ?? 'default'}>
                    {STATUS_LABELS[promotion.status] ?? promotion.status}
                  </Tag>
                )}
              </div>
            </Card>
          )}
        </>
      )}

      {/* All subscriptions */}
      <h2 className="rh-inline-title">Все подписки</h2>
      <Table
        dataSource={allSubscriptions}
        columns={subsColumns}
        rowKey="id"
        loading={allSubsLoading}
        locale={{ emptyText: <EmptyState description="Нет подписок" /> }}
        pagination={
          totalSubs > subsPageSize
            ? {
                current: subsPage,
                pageSize: subsPageSize,
                total: totalSubs,
                onChange: setSubsPage,
                showSizeChanger: false,
              }
            : false
        }
      />

      {/* Create promotion modal */}
      <Modal
        title="Новая промо-кампания"
        open={promoModalOpen}
        onCancel={() => setPromoModalOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form
          form={promoForm}
          layout="vertical"
          onFinish={handleCreatePromotion}
          initialValues={{ duration_days: 30 }}
        >
          <Form.Item
            name="budget"
            label="Бюджет (₽)"
            rules={[{ required: true, message: 'Укажите бюджет' }]}
            extra="Сумма в рублях на рекламную кампанию"
          >
            <InputNumber min={100} step={500} className="rh-full-width" placeholder="5000" />
          </Form.Item>

          <Form.Item
            name="duration_days"
            label="Длительность (дней)"
            rules={[{ required: true, message: 'Укажите длительность' }]}
          >
            <Select
              options={[
                { value: 7, label: '7 дней' },
                { value: 14, label: '14 дней' },
                { value: 30, label: '30 дней' },
                { value: 60, label: '60 дней' },
                { value: 90, label: '90 дней' },
              ]}
            />
          </Form.Item>

          <Form.Item
            name="target_city_id"
            label="Целевой город"
            extra="Оставьте пустым для показа во всех городах"
          >
            <Select allowClear placeholder="Все города" className="rh-full-width">
              {cities.map((city) => (
                <Select.Option key={city.id} value={city.id}>
                  {city.name}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={promotionMutation.isPending}>
                Создать
              </Button>
              <Button onClick={() => setPromoModalOpen(false)}>Отмена</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
