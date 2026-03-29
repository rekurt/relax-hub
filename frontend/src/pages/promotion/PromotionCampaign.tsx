import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Col,
  DatePicker,
  Descriptions,
  Empty,
  Form,
  InputNumber,
  Modal,
  Popconfirm,
  Progress,
  Row,
  Select,
  Space,
  Statistic,
  Table,
  Tag,
  Tooltip,
  Typography,
} from 'antd'
import {
  PauseCircleOutlined,
  PlayCircleOutlined,
  PlusOutlined,
  RocketOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import type { Dayjs } from 'dayjs'
import { useBathhouseStore } from '@/stores/bathhouse'
import { useQueryClient } from '@tanstack/react-query'
import { formatPrice } from '@/lib/format'
import { axiosInstance } from '@/api/axios-instance'
import { useQuery, useMutation } from '@tanstack/react-query'
import { useGetCities } from '@/api/generated/cities/cities'

const { Title, Text } = Typography

const STATUS_LABELS: Record<string, string> = {
  active: 'Активна',
  paused: 'На паузе',
  exhausted: 'Бюджет исчерпан',
  expired: 'Завершена',
}

const STATUS_COLORS: Record<string, string> = {
  active: 'green',
  paused: 'orange',
  exhausted: 'red',
  expired: 'default',
}

interface Promotion {
  id: string
  bathhouse_id: string
  daily_bid_kopecks: number
  budget_kopecks: number
  spent_kopecks: number
  remaining_budget: number
  start_date: string
  end_date: string
  target_city_id?: number | null
  status: string
  impression_count: number
  click_count: number
  ctr: number
  cost_per_click: number
  created_at: string
  updated_at: string
}

interface CampaignFormValues {
  daily_bid: number
  budget: number
  duration_days: number
  target_city_id?: number
}

export default function PromotionCampaign() {
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)
  const queryClient = useQueryClient()
  const { message } = App.useApp()
  const [form] = Form.useForm<CampaignFormValues>()
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [page, setPage] = useState(1)
  const pageSize = 10

  const { data: citiesData } = useGetCities()
  const cities = (citiesData?.data ?? []) as { id?: number; name?: string }[]

  const { data: promotionsData, isLoading } = useQuery({
    queryKey: [`/my/bathhouses/${selectedBathhouseId}/promotions`, page, pageSize],
    queryFn: async () => {
      const res = await axiosInstance.get(
        `/api/v1/my/bathhouses/${selectedBathhouseId}/promotions`,
        { params: { page, page_size: pageSize } },
      )
      return res.data
    },
    enabled: !!selectedBathhouseId,
  })

  const promotions = (promotionsData?.data ?? []) as Promotion[]
  const totalCount = promotionsData?.meta?.total_count ?? 0
  const activePromotion = promotions.find((p) => p.status === 'active')

  const createMutation = useMutation({
    mutationFn: async (values: CampaignFormValues) => {
      const res = await axiosInstance.post(
        `/api/v1/my/bathhouses/${selectedBathhouseId}/promotion`,
        {
          daily_bid_kopecks: values.daily_bid * 100,
          budget_kopecks: values.budget * 100,
          duration_days: values.duration_days,
          target_city_id: values.target_city_id ?? null,
        },
      )
      return res.data
    },
    onSuccess: () => {
      message.success('Кампания создана')
      setCreateModalOpen(false)
      form.resetFields()
      invalidate()
    },
    onError: () => message.error('Не удалось создать кампанию'),
  })

  const pauseMutation = useMutation({
    mutationFn: async (id: string) => {
      await axiosInstance.post(`/api/v1/my/promotions/${id}/pause`)
    },
    onSuccess: () => {
      message.success('Кампания приостановлена')
      invalidate()
    },
    onError: () => message.error('Не удалось приостановить кампанию'),
  })

  const resumeMutation = useMutation({
    mutationFn: async (id: string) => {
      await axiosInstance.post(`/api/v1/my/promotions/${id}/resume`)
    },
    onSuccess: () => {
      message.success('Кампания возобновлена')
      invalidate()
    },
    onError: () => message.error('Не удалось возобновить кампанию'),
  })

  const invalidate = () => {
    queryClient.invalidateQueries({
      queryKey: [`/my/bathhouses/${selectedBathhouseId}/promotions`],
    })
    queryClient.invalidateQueries({
      queryKey: [`/my/bathhouses/${selectedBathhouseId}/promotion`],
    })
  }

  if (!selectedBathhouseId) {
    return (
      <Card>
        <Empty description="Выберите объект для управления продвижением" />
      </Card>
    )
  }

  const columns = [
    {
      title: 'Период',
      key: 'period',
      render: (_: unknown, record: Promotion) => (
        <Space direction="vertical" size={0}>
          <Text>{dayjs(record.start_date).format('DD.MM.YYYY')}</Text>
          <Text type="secondary">{dayjs(record.end_date).format('DD.MM.YYYY')}</Text>
        </Space>
      ),
    },
    {
      title: 'Ставка/день',
      dataIndex: 'daily_bid_kopecks',
      render: (v: number) => formatPrice(v),
    },
    {
      title: 'Бюджет',
      key: 'budget',
      render: (_: unknown, record: Promotion) => {
        const pct = record.budget_kopecks > 0
          ? Math.round((record.spent_kopecks / record.budget_kopecks) * 100)
          : 0
        return (
          <Tooltip title={`${formatPrice(record.spent_kopecks)} из ${formatPrice(record.budget_kopecks)}`}>
            <Progress
              percent={pct}
              size="small"
              status={record.status === 'exhausted' ? 'exception' : 'active'}
            />
          </Tooltip>
        )
      },
    },
    {
      title: 'Показы',
      dataIndex: 'impression_count',
      render: (v: number) => v.toLocaleString('ru-RU'),
    },
    {
      title: 'Клики',
      dataIndex: 'click_count',
      render: (v: number) => v.toLocaleString('ru-RU'),
    },
    {
      title: 'CTR',
      dataIndex: 'ctr',
      render: (v: number) => `${v.toFixed(2)}%`,
    },
    {
      title: 'Цена клика',
      dataIndex: 'cost_per_click',
      render: (v: number) => v > 0 ? formatPrice(Math.round(v)) : '—',
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      render: (status: string) => (
        <Tag color={STATUS_COLORS[status] ?? 'default'}>
          {STATUS_LABELS[status] ?? status}
        </Tag>
      ),
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_: unknown, record: Promotion) => (
        <Space>
          {record.status === 'active' && (
            <Popconfirm
              title="Приостановить кампанию?"
              onConfirm={() => pauseMutation.mutate(record.id)}
            >
              <Button
                icon={<PauseCircleOutlined />}
                size="small"
                loading={pauseMutation.isPending}
              >
                Пауза
              </Button>
            </Popconfirm>
          )}
          {record.status === 'paused' && (
            <Button
              icon={<PlayCircleOutlined />}
              size="small"
              type="primary"
              loading={resumeMutation.isPending}
              onClick={() => resumeMutation.mutate(record.id)}
            >
              Возобновить
            </Button>
          )}
        </Space>
      ),
    },
  ]

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={3} style={{ margin: 0 }}>
          <RocketOutlined /> Рекламные кампании
        </Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setCreateModalOpen(true)}
          disabled={!!activePromotion}
        >
          Создать кампанию
        </Button>
      </div>

      {activePromotion && (
        <Card title="Активная кампания" size="small">
          <Row gutter={[16, 16]}>
            <Col xs={12} sm={6}>
              <Statistic
                title="Ставка/день"
                value={activePromotion.daily_bid_kopecks / 100}
                suffix="₽"
              />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic
                title="Остаток бюджета"
                value={activePromotion.remaining_budget / 100}
                suffix="₽"
              />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic title="Показы" value={activePromotion.impression_count} />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic title="Клики" value={activePromotion.click_count} />
            </Col>
          </Row>
          <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
            <Col xs={12} sm={6}>
              <Statistic title="CTR" value={activePromotion.ctr.toFixed(2)} suffix="%" />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic
                title="Цена клика"
                value={activePromotion.cost_per_click > 0 ? (activePromotion.cost_per_click / 100).toFixed(2) : '—'}
                suffix={activePromotion.cost_per_click > 0 ? '₽' : ''}
              />
            </Col>
            <Col xs={12} sm={6}>
              <Descriptions size="small" column={1}>
                <Descriptions.Item label="Окончание">
                  {dayjs(activePromotion.end_date).format('DD.MM.YYYY')}
                </Descriptions.Item>
              </Descriptions>
            </Col>
            <Col xs={12} sm={6}>
              <Progress
                type="circle"
                size={60}
                percent={
                  activePromotion.budget_kopecks > 0
                    ? Math.round((activePromotion.spent_kopecks / activePromotion.budget_kopecks) * 100)
                    : 0
                }
              />
            </Col>
          </Row>
        </Card>
      )}

      <Table
        dataSource={promotions}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        pagination={{
          current: page,
          pageSize,
          total: totalCount,
          onChange: setPage,
          showSizeChanger: false,
        }}
        locale={{ emptyText: 'Нет кампаний. Создайте первую!' }}
      />

      <Modal
        title="Новая рекламная кампания"
        open={createModalOpen}
        onCancel={() => setCreateModalOpen(false)}
        onOk={() => form.submit()}
        confirmLoading={createMutation.isPending}
        okText="Создать"
        cancelText="Отмена"
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={(values) => createMutation.mutate(values)}
          initialValues={{ daily_bid: 50, budget: 1000, duration_days: 30 }}
        >
          <Form.Item
            name="daily_bid"
            label="Ставка в день (₽)"
            rules={[{ required: true, message: 'Укажите ставку' }]}
            extra="Минимум 50 ₽/день. Чем выше ставка, тем выше позиция в поиске."
          >
            <InputNumber min={50} style={{ width: '100%' }} addonAfter="₽/день" />
          </Form.Item>

          <Form.Item
            name="budget"
            label="Общий бюджет (₽)"
            rules={[
              { required: true, message: 'Укажите бюджет' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  const dailyBid = getFieldValue('daily_bid')
                  if (value && dailyBid && value < dailyBid) {
                    return Promise.reject(new Error('Бюджет должен быть не менее дневной ставки'))
                  }
                  return Promise.resolve()
                },
              }),
            ]}
            extra="Кампания приостанавливается при исчерпании бюджета."
          >
            <InputNumber min={50} style={{ width: '100%' }} addonAfter="₽" />
          </Form.Item>

          <Form.Item
            name="duration_days"
            label="Длительность (дней)"
            rules={[{ required: true, message: 'Укажите длительность' }]}
          >
            <InputNumber min={1} max={365} style={{ width: '100%' }} addonAfter="дн." />
          </Form.Item>

          <Form.Item name="target_city_id" label="Целевой город (опционально)">
            <Select allowClear placeholder="Все города">
              {cities.map((c) => (
                <Select.Option key={c.id} value={c.id}>
                  {c.name}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  )
}
