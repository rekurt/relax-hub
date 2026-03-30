import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Col,
  Form,
  Input,
  InputNumber,
  List,
  Modal,
  Popconfirm,
  Row,
  Select,
  Table,
  Tag,
  Typography,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  TeamOutlined,
} from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import dayjs from 'dayjs'
import { axiosInstance } from '@/api/axios-instance'
import { formatPrice } from '@/lib/format'

const { Title, Text } = Typography

interface SegmentConditions {
  visit_count_min?: number
  visit_count_max?: number
  avg_check_min?: number
  avg_check_max?: number
  total_spent_min?: number
  total_spent_max?: number
  last_visit_days_min?: number
  last_visit_days_max?: number
  tags_include?: string[]
  tags_exclude?: string[]
  rfm_recency_min?: number
  rfm_recency_max?: number
  rfm_frequency_min?: number
  rfm_frequency_max?: number
  rfm_monetary_min?: number
  rfm_monetary_max?: number
}

interface CustomSegment {
  id: string
  owner_id: string
  bathhouse_id?: string
  name: string
  conditions: SegmentConditions
  guest_count: number
  created_at: string
  updated_at: string
}

interface GuestCard {
  id: string
  client_id: string
  bathhouse_id: string
  visit_count: number
  total_spent: number
  avg_check: number
  last_visit_at: string
  tags: string[]
}

function useCustomSegments() {
  return useQuery({
    queryKey: ['crm', 'custom-segments'],
    queryFn: async () => {
      const { data } = await axiosInstance.get<{ data: CustomSegment[] }>('/my/crm/segments/custom')
      return data.data
    },
  })
}

function useSegmentGuests(segmentId: string | undefined, page: number) {
  return useQuery({
    queryKey: ['crm', 'custom-segments', segmentId, 'guests', page],
    queryFn: async () => {
      const { data } = await axiosInstance.get(`/my/crm/segments/custom/${segmentId}/guests`, {
        params: { page, page_size: 20 },
      })
      return data
    },
    enabled: !!segmentId,
  })
}

const RFM_OPTIONS = [1, 2, 3, 4, 5].map((v) => ({ label: `${v}`, value: v }))

function SegmentForm({
  open,
  onClose,
  initialValues,
}: {
  open: boolean
  onClose: () => void
  initialValues?: CustomSegment
}) {
  const [form] = Form.useForm()
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const isEdit = !!initialValues

  const createMutation = useMutation({
    mutationFn: (values: { name: string; conditions: SegmentConditions }) =>
      axiosInstance.post('/my/crm/segments/custom', values),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['crm', 'custom-segments'] })
      message.success('Сегмент создан')
      onClose()
    },
    onError: () => message.error('Ошибка при создании сегмента'),
  })

  const updateMutation = useMutation({
    mutationFn: (values: { name: string; conditions: SegmentConditions }) =>
      axiosInstance.put(`/my/crm/segments/custom/${initialValues?.id}`, values),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['crm', 'custom-segments'] })
      message.success('Сегмент обновлён')
      onClose()
    },
    onError: () => message.error('Ошибка при обновлении'),
  })

  const handleSubmit = () => {
    form.validateFields().then((values) => {
      const conditions: SegmentConditions = {}

      if (values.visit_count_min != null) conditions.visit_count_min = values.visit_count_min
      if (values.visit_count_max != null) conditions.visit_count_max = values.visit_count_max
      if (values.avg_check_min != null) conditions.avg_check_min = values.avg_check_min * 100
      if (values.avg_check_max != null) conditions.avg_check_max = values.avg_check_max * 100
      if (values.total_spent_min != null) conditions.total_spent_min = values.total_spent_min * 100
      if (values.total_spent_max != null) conditions.total_spent_max = values.total_spent_max * 100
      if (values.last_visit_days_min != null) conditions.last_visit_days_min = values.last_visit_days_min
      if (values.last_visit_days_max != null) conditions.last_visit_days_max = values.last_visit_days_max
      if (values.tags_include?.length) conditions.tags_include = values.tags_include
      if (values.tags_exclude?.length) conditions.tags_exclude = values.tags_exclude
      if (values.rfm_recency_min != null) conditions.rfm_recency_min = values.rfm_recency_min
      if (values.rfm_recency_max != null) conditions.rfm_recency_max = values.rfm_recency_max
      if (values.rfm_frequency_min != null) conditions.rfm_frequency_min = values.rfm_frequency_min
      if (values.rfm_frequency_max != null) conditions.rfm_frequency_max = values.rfm_frequency_max
      if (values.rfm_monetary_min != null) conditions.rfm_monetary_min = values.rfm_monetary_min
      if (values.rfm_monetary_max != null) conditions.rfm_monetary_max = values.rfm_monetary_max

      const payload = { name: values.name, conditions }

      if (isEdit) {
        updateMutation.mutate(payload)
      } else {
        createMutation.mutate(payload)
      }
    })
  }

  const initCond = initialValues?.conditions ?? {}

  return (
    <Modal
      title={isEdit ? 'Редактировать сегмент' : 'Новый сегмент'}
      open={open}
      onCancel={onClose}
      onOk={handleSubmit}
      okText={isEdit ? 'Сохранить' : 'Создать'}
      confirmLoading={createMutation.isPending || updateMutation.isPending}
      width={640}
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          name: initialValues?.name ?? '',
          visit_count_min: initCond.visit_count_min,
          visit_count_max: initCond.visit_count_max,
          avg_check_min: initCond.avg_check_min != null ? initCond.avg_check_min / 100 : undefined,
          avg_check_max: initCond.avg_check_max != null ? initCond.avg_check_max / 100 : undefined,
          total_spent_min: initCond.total_spent_min != null ? initCond.total_spent_min / 100 : undefined,
          total_spent_max: initCond.total_spent_max != null ? initCond.total_spent_max / 100 : undefined,
          last_visit_days_min: initCond.last_visit_days_min,
          last_visit_days_max: initCond.last_visit_days_max,
          tags_include: initCond.tags_include ?? [],
          tags_exclude: initCond.tags_exclude ?? [],
          rfm_recency_min: initCond.rfm_recency_min,
          rfm_recency_max: initCond.rfm_recency_max,
          rfm_frequency_min: initCond.rfm_frequency_min,
          rfm_frequency_max: initCond.rfm_frequency_max,
          rfm_monetary_min: initCond.rfm_monetary_min,
          rfm_monetary_max: initCond.rfm_monetary_max,
        }}
      >
        <Form.Item name="name" label="Название сегмента" rules={[{ required: true, message: 'Введите название' }]}>
          <Input placeholder="Например: VIP клиенты за последние 30 дней" />
        </Form.Item>

        <Title level={5}>Фильтры по визитам</Title>
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item name="visit_count_min" label="Визиты от">
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item name="visit_count_max" label="Визиты до">
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </Col>
        </Row>

        <Title level={5}>Фильтры по суммам (руб)</Title>
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item name="avg_check_min" label="Средний чек от">
              <InputNumber min={0} style={{ width: '100%' }} addonAfter="₽" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item name="avg_check_max" label="Средний чек до">
              <InputNumber min={0} style={{ width: '100%' }} addonAfter="₽" />
            </Form.Item>
          </Col>
        </Row>
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item name="total_spent_min" label="Общая сумма от">
              <InputNumber min={0} style={{ width: '100%' }} addonAfter="₽" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item name="total_spent_max" label="Общая сумма до">
              <InputNumber min={0} style={{ width: '100%' }} addonAfter="₽" />
            </Form.Item>
          </Col>
        </Row>

        <Title level={5}>Давность визита (дней)</Title>
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item name="last_visit_days_min" label="Минимум дней назад">
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item name="last_visit_days_max" label="Максимум дней назад">
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </Col>
        </Row>

        <Title level={5}>Теги</Title>
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item name="tags_include" label="Включить теги">
              <Select mode="tags" placeholder="Добавьте теги" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item name="tags_exclude" label="Исключить теги">
              <Select mode="tags" placeholder="Добавьте теги" />
            </Form.Item>
          </Col>
        </Row>

        <Title level={5}>RFM-скоры (1-5)</Title>
        <Row gutter={16}>
          <Col span={8}>
            <Form.Item name="rfm_recency_min" label="R от">
              <Select options={RFM_OPTIONS} allowClear placeholder="—" />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="rfm_frequency_min" label="F от">
              <Select options={RFM_OPTIONS} allowClear placeholder="—" />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="rfm_monetary_min" label="M от">
              <Select options={RFM_OPTIONS} allowClear placeholder="—" />
            </Form.Item>
          </Col>
        </Row>
        <Row gutter={16}>
          <Col span={8}>
            <Form.Item name="rfm_recency_max" label="R до">
              <Select options={RFM_OPTIONS} allowClear placeholder="—" />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="rfm_frequency_max" label="F до">
              <Select options={RFM_OPTIONS} allowClear placeholder="—" />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="rfm_monetary_max" label="M до">
              <Select options={RFM_OPTIONS} allowClear placeholder="—" />
            </Form.Item>
          </Col>
        </Row>
      </Form>
    </Modal>
  )
}

export default function SegmentBuilder() {
  const { message } = App.useApp()
  const [selectedSegment, setSelectedSegment] = useState<CustomSegment>()
  const [formOpen, setFormOpen] = useState(false)
  const [editSegment, setEditSegment] = useState<CustomSegment>()
  const [page, setPage] = useState(1)
  const queryClient = useQueryClient()

  const { data: segments, isLoading } = useCustomSegments()
  const { data: guestsData, isLoading: guestsLoading } = useSegmentGuests(selectedSegment?.id, page)

  const guests = (guestsData as { data: GuestCard[] })?.data ?? []
  const totalCount = (guestsData as { meta?: { total_count?: number } })?.meta?.total_count ?? 0

  const deleteMutation = useMutation({
    mutationFn: (id: string) => axiosInstance.delete(`/my/crm/segments/custom/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['crm', 'custom-segments'] })
      setSelectedSegment(undefined)
      message.success('Сегмент удалён')
    },
  })

  const guestColumns = [
    {
      title: 'ID клиента',
      dataIndex: 'client_id',
      key: 'client_id',
      render: (id: string) => <Tag style={{ fontFamily: 'monospace' }}>{id?.slice(0, 8)}...</Tag>,
    },
    {
      title: 'Визиты',
      dataIndex: 'visit_count',
      key: 'visit_count',
      width: 90,
      render: (v: number) => <Tag color="blue">{v ?? 0}</Tag>,
    },
    {
      title: 'Общая сумма',
      dataIndex: 'total_spent',
      key: 'total_spent',
      width: 130,
      render: (v: number) => formatPrice(v ?? 0),
    },
    {
      title: 'Последний визит',
      dataIndex: 'last_visit_at',
      key: 'last_visit_at',
      width: 140,
      render: (d: string) => (d ? dayjs(d).format('DD.MM.YYYY') : '—'),
    },
    {
      title: 'Теги',
      dataIndex: 'tags',
      key: 'tags',
      render: (tags: string[]) => tags?.length ? tags.map((t) => <Tag key={t}>{t}</Tag>) : '—',
    },
  ]

  function formatConditionsSummary(c: SegmentConditions): string {
    const parts: string[] = []
    if (c.visit_count_min != null || c.visit_count_max != null) {
      parts.push(`Визиты: ${c.visit_count_min ?? ''}–${c.visit_count_max ?? ''}`)
    }
    if (c.total_spent_min != null || c.total_spent_max != null) {
      const min = c.total_spent_min != null ? (c.total_spent_min / 100).toFixed(0) : ''
      const max = c.total_spent_max != null ? (c.total_spent_max / 100).toFixed(0) : ''
      parts.push(`Сумма: ${min}–${max} ₽`)
    }
    if (c.last_visit_days_min != null || c.last_visit_days_max != null) {
      parts.push(`Давность: ${c.last_visit_days_min ?? ''}–${c.last_visit_days_max ?? ''} дн.`)
    }
    if (c.rfm_recency_min != null || c.rfm_frequency_min != null || c.rfm_monetary_min != null) {
      parts.push('RFM фильтр')
    }
    if (c.tags_include?.length) parts.push(`Теги: ${c.tags_include.join(', ')}`)
    return parts.length ? parts.join(' | ') : 'Без условий'
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>Конструктор сегментов</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditSegment(undefined); setFormOpen(true) }}>
          Новый сегмент
        </Button>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} md={10}>
          <List
            loading={isLoading}
            dataSource={segments ?? []}
            locale={{ emptyText: 'Нет пользовательских сегментов. Создайте первый!' }}
            renderItem={(item: CustomSegment) => (
              <Card
                size="small"
                hoverable
                style={{
                  marginBottom: 8,
                  borderColor: selectedSegment?.id === item.id ? '#1677ff' : undefined,
                  borderWidth: selectedSegment?.id === item.id ? 2 : 1,
                }}
                onClick={() => { setSelectedSegment(item); setPage(1) }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <div style={{ flex: 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <TeamOutlined style={{ color: '#1677ff', fontSize: 18 }} />
                      <Text strong>{item.name}</Text>
                    </div>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {formatConditionsSummary(item.conditions)}
                    </Text>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <Tag color="blue" style={{ fontSize: 14, padding: '2px 10px' }}>
                      {item.guest_count}
                    </Tag>
                    <Button
                      type="text"
                      size="small"
                      icon={<EditOutlined />}
                      onClick={(e) => { e.stopPropagation(); setEditSegment(item); setFormOpen(true) }}
                    />
                    <Popconfirm
                      title="Удалить сегмент?"
                      onConfirm={(e) => { e?.stopPropagation(); deleteMutation.mutate(item.id) }}
                      onCancel={(e) => e?.stopPropagation()}
                    >
                      <Button
                        type="text"
                        size="small"
                        danger
                        icon={<DeleteOutlined />}
                        onClick={(e) => e.stopPropagation()}
                      />
                    </Popconfirm>
                  </div>
                </div>
              </Card>
            )}
          />
        </Col>

        <Col xs={24} md={14}>
          {selectedSegment ? (
            <Card title={`Гости сегмента "${selectedSegment.name}"`} size="small">
              <Table
                dataSource={guests}
                columns={guestColumns}
                rowKey="id"
                loading={guestsLoading}
                size="small"
                locale={{ emptyText: 'Нет гостей, соответствующих условиям' }}
                pagination={
                  totalCount > 20
                    ? { current: page, pageSize: 20, total: totalCount, onChange: setPage, showSizeChanger: false }
                    : false
                }
              />
            </Card>
          ) : (
            <Card>
              <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
                Выберите сегмент или создайте новый
              </div>
            </Card>
          )}
        </Col>
      </Row>

      <SegmentForm
        open={formOpen}
        onClose={() => { setFormOpen(false); setEditSegment(undefined) }}
        initialValues={editSegment}
      />
    </div>
  )
}
