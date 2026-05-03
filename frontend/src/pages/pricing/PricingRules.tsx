import { useState } from 'react'
import {
  Alert,
  App,
  Button,
  Card,
  DatePicker,
  Divider,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Statistic,
  Switch,
  Table,
  Tag,
  TimePicker,
  Typography,
} from '@/components/design/system'
import {
  ArrowUpOutlined,
  ArrowDownOutlined,
  BulbOutlined,
  CheckOutlined,
  CloseOutlined,
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
} from '@/components/design/icons'
import dayjs from 'dayjs'
import {
  useGetMyBathhousesIdPricingRules,
  usePostMyBathhousesIdPricingRules,
  usePutPricingRulesId,
  useDeletePricingRulesId,
} from '@/api/generated/pricing/pricing'
import type { InternalHandlerPricingRuleResponse } from '@/api/generated/model'
import { useBathhouseStore } from '@/stores/bathhouse'
import { useQueryClient, useQuery, useMutation } from '@tanstack/react-query'
import { formatDayOfWeek } from '@/lib/format'
import { axiosInstance } from '@/api/axios-instance'

const { Title } = Typography

const RULE_TYPES = [
  { value: 'weekday', label: 'Будние дни' },
  { value: 'weekend', label: 'Выходные' },
  { value: 'per_day', label: 'По дням недели' },
  { value: 'holiday', label: 'Праздники' },
  { value: 'time_range', label: 'Диапазон времени' },
  { value: 'season', label: 'Сезон' },
]

const RULE_TYPE_COLORS: Record<string, string> = {
  weekday: 'blue',
  weekend: 'green',
  per_day: 'cyan',
  holiday: 'red',
  time_range: 'orange',
  season: 'purple',
}

const WEEKDAY_OPTIONS = [0, 1, 2, 3, 4].map((d) => ({
  value: d,
  label: formatDayOfWeek(d),
}))

const WEEKEND_OPTIONS = [5, 6].map((d) => ({
  value: d,
  label: formatDayOfWeek(d),
}))

const ALL_DAY_OPTIONS = [0, 1, 2, 3, 4, 5, 6].map((d) => ({
  value: d,
  label: formatDayOfWeek(d),
}))

interface RuleFormValues {
  name: string
  type: string
  multiplier: number
  priority: number
  is_active: boolean
  days_of_week?: number[]
  time_from?: dayjs.Dayjs
  time_to?: dayjs.Dayjs
  date_range?: [dayjs.Dayjs, dayjs.Dayjs]
}

interface SeasonalTariff {
  id: string
  bathhouse_id: string
  name: string
  date_from: string
  date_to: string
  multiplier: number
  is_active: boolean
  created_at: string
  updated_at: string
}

interface TariffFormValues {
  name: string
  date_range: [dayjs.Dayjs, dayjs.Dayjs]
  multiplier: number
  is_active: boolean
}

export default function PricingRules() {
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)
  const queryClient = useQueryClient()
  const { message } = App.useApp()
  const [form] = Form.useForm<RuleFormValues>()
  const [tariffForm] = Form.useForm<TariffFormValues>()

  const [modalOpen, setModalOpen] = useState(false)
  const [editingRule, setEditingRule] = useState<InternalHandlerPricingRuleResponse | null>(null)
  const [tariffModalOpen, setTariffModalOpen] = useState(false)
  const [editingTariff, setEditingTariff] = useState<SeasonalTariff | null>(null)

  // --- Pricing Rules ---
  const { data, isLoading } = useGetMyBathhousesIdPricingRules(selectedBathhouseId ?? '', {
    query: { enabled: !!selectedBathhouseId },
  })

  const rules = data?.data ?? []

  const invalidateRules = () => {
    queryClient.invalidateQueries({
      queryKey: [`/my/bathhouses/${selectedBathhouseId}/pricing-rules`],
    })
  }

  const createMutation = usePostMyBathhousesIdPricingRules({
    mutation: {
      onSuccess: () => {
        message.success('Правило создано')
        setModalOpen(false)
        form.resetFields()
        invalidateRules()
      },
      onError: () => message.error('Не удалось создать правило'),
    },
  })

  const updateMutation = usePutPricingRulesId({
    mutation: {
      onSuccess: () => {
        message.success('Правило обновлено')
        setModalOpen(false)
        setEditingRule(null)
        form.resetFields()
        invalidateRules()
      },
      onError: () => message.error('Не удалось обновить правило'),
    },
  })

  const deleteMutation = useDeletePricingRulesId({
    mutation: {
      onSuccess: () => {
        message.success('Правило удалено')
        invalidateRules()
      },
      onError: () => message.error('Не удалось удалить правило'),
    },
  })

  const handleOpenCreate = () => {
    setEditingRule(null)
    form.resetFields()
    form.setFieldsValue({ is_active: true, priority: 0, multiplier: 1.0 })
    setModalOpen(true)
  }

  const handleOpenEdit = (rule: InternalHandlerPricingRuleResponse) => {
    setEditingRule(rule)
    form.setFieldsValue({
      name: rule.name,
      type: rule.type,
      multiplier: rule.multiplier,
      priority: rule.priority ?? 0,
      is_active: rule.is_active ?? true,
      days_of_week: rule.days_of_week,
      time_from: rule.time_from ? dayjs(rule.time_from, 'HH:mm') : undefined,
      time_to: rule.time_to ? dayjs(rule.time_to, 'HH:mm') : undefined,
      date_range: rule.date_from && rule.date_to
        ? [dayjs(rule.date_from), dayjs(rule.date_to)]
        : undefined,
    })
    setModalOpen(true)
  }

  const handleSubmit = (values: RuleFormValues) => {
    const payload = {
      name: values.name,
      type: values.type,
      multiplier: values.multiplier,
      priority: values.priority,
      is_active: values.is_active,
      days_of_week: values.days_of_week,
      time_from: values.time_from ? values.time_from.format('HH:mm') : undefined,
      time_to: values.time_to ? values.time_to.format('HH:mm') : undefined,
      date_from: values.date_range?.[0]?.format('YYYY-MM-DD'),
      date_to: values.date_range?.[1]?.format('YYYY-MM-DD'),
    }

    if (editingRule?.id) {
      updateMutation.mutate({ id: editingRule.id, data: payload })
    } else if (selectedBathhouseId) {
      createMutation.mutate({ id: selectedBathhouseId, data: payload })
    }
  }

  const handleDelete = (ruleId: string) => {
    deleteMutation.mutate({ id: ruleId })
  }

  const handleToggleActive = (rule: InternalHandlerPricingRuleResponse) => {
    if (!rule.id) return
    updateMutation.mutate({
      id: rule.id,
      data: {
        name: rule.name,
        type: rule.type,
        multiplier: rule.multiplier,
        priority: rule.priority,
        is_active: !rule.is_active,
        days_of_week: rule.days_of_week,
        time_from: rule.time_from,
        time_to: rule.time_to,
        date_from: rule.date_from,
        date_to: rule.date_to,
      },
    })
  }

  const selectedType = Form.useWatch('type', form)

  const formatConditions = (rule: InternalHandlerPricingRuleResponse): string => {
    const parts: string[] = []
    if (rule.days_of_week && rule.days_of_week.length > 0) {
      parts.push(rule.days_of_week.map((d) => formatDayOfWeek(d, true)).join(', '))
    }
    if (rule.time_from && rule.time_to) {
      parts.push(`${rule.time_from}–${rule.time_to}`)
    }
    if (rule.date_from && rule.date_to) {
      parts.push(`${dayjs(rule.date_from).format('DD.MM.YYYY')}–${dayjs(rule.date_to).format('DD.MM.YYYY')}`)
    }
    return parts.join(' | ') || '—'
  }

  const columns = [
    {
      title: 'Приоритет',
      dataIndex: 'priority',
      key: 'priority',
      width: 100,
      sorter: (a: InternalHandlerPricingRuleResponse, b: InternalHandlerPricingRuleResponse) =>
        (a.priority ?? 0) - (b.priority ?? 0),
      defaultSortOrder: 'descend' as const,
    },
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Тип',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => {
        const label = RULE_TYPES.find((t) => t.value === type)?.label ?? type
        return <Tag color={RULE_TYPE_COLORS[type] ?? 'default'}>{label}</Tag>
      },
    },
    {
      title: 'Множитель',
      dataIndex: 'multiplier',
      key: 'multiplier',
      width: 120,
      render: (val: number) => {
        const isDiscount = val < 1
        const percent = Math.round(Math.abs(val - 1) * 100)
        return (
          <span style={{ color: isDiscount ? '#15803d' : val > 1 ? '#b42318' : undefined }}>
            x{val} {percent > 0 && `(${isDiscount ? '-' : '+'}${percent}%)`}
          </span>
        )
      },
    },
    {
      title: 'Условия',
      key: 'conditions',
      render: (_: unknown, record: InternalHandlerPricingRuleResponse) => formatConditions(record),
    },
    {
      title: 'Активно',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 90,
      render: (active: boolean, record: InternalHandlerPricingRuleResponse) => (
        <Switch
          checked={active}
          onChange={() => handleToggleActive(record)}
          size="small"
        />
      ),
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 100,
      render: (_: unknown, record: InternalHandlerPricingRuleResponse) => (
        <Space>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => handleOpenEdit(record)}
            size="small"
          />
          <Popconfirm
            title="Удалить правило?"
            description="Это действие нельзя отменить"
            onConfirm={() => record.id && handleDelete(record.id)}
            okText="Удалить"
            cancelText="Отмена"
          >
            <Button type="text" danger icon={<DeleteOutlined />} size="small" />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  // --- Smart Pricing Recommendation ---
  interface PriceRecommendation {
    current_price: number
    recommended_price: number
    coefficient: number
    avg_area_price: number
    occupancy_rate: number
    demand_trend: string
    recommendation_basis: string
  }

  const { data: recommendationData } = useQuery<PriceRecommendation>({
    queryKey: ['price-recommendation', selectedBathhouseId],
    queryFn: async () => {
      const res = await axiosInstance.get<{ success: boolean; data: PriceRecommendation }>(
        `/my/bathhouses/${selectedBathhouseId}/price-recommendation`,
      )
      return res.data.data
    },
    enabled: !!selectedBathhouseId,
    staleTime: 5 * 60 * 1000,
  })

  const [recommendationDismissed, setRecommendationDismissed] = useState(false)

  const acceptRecommendationMutation = useMutation({
    mutationFn: async (recommendedPrice: number) => {
      await axiosInstance.put(`/my/bathhouses/${selectedBathhouseId}`, {
        base_price: recommendedPrice,
      })
    },
    onSuccess: () => {
      message.success('Рекомендованная цена применена')
      setRecommendationDismissed(true)
      queryClient.invalidateQueries({ queryKey: ['price-recommendation', selectedBathhouseId] })
    },
    onError: () => {
      message.error('Не удалось применить рекомендацию')
    },
  })

  const formatPrice = (kopecks: number) => `${(kopecks / 100).toLocaleString('ru-RU')} ₽`
  const demandTrendLabel: Record<string, string> = {
    growing: 'Растущий',
    declining: 'Снижающийся',
    stable: 'Стабильный',
  }

  // --- Seasonal Tariffs ---
  const tariffQueryKey = ['seasonal-tariffs', selectedBathhouseId]

  const { data: tariffsData, isLoading: tariffsLoading } = useQuery({
    queryKey: tariffQueryKey,
    queryFn: async () => {
      const res = await axiosInstance.get<{ success: boolean; data: SeasonalTariff[] }>(
        `/my/bathhouses/${selectedBathhouseId}/seasonal-tariffs`,
      )
      return res.data.data ?? []
    },
    enabled: !!selectedBathhouseId,
  })

  const tariffs = tariffsData ?? []

  const invalidateTariffs = () => {
    queryClient.invalidateQueries({ queryKey: tariffQueryKey })
  }

  const createTariffMutation = useMutation({
    mutationFn: async (data: { name: string; date_from: string; date_to: string; multiplier: number; is_active: boolean }) => {
      return axiosInstance.post(`/my/bathhouses/${selectedBathhouseId}/seasonal-tariffs`, data)
    },
    onSuccess: () => {
      message.success('Сезонный тариф создан')
      setTariffModalOpen(false)
      tariffForm.resetFields()
      invalidateTariffs()
    },
    onError: () => message.error('Не удалось создать тариф'),
  })

  const updateTariffMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: { name: string; date_from: string; date_to: string; multiplier: number; is_active: boolean } }) => {
      return axiosInstance.put(`/seasonal-tariffs/${id}`, data)
    },
    onSuccess: () => {
      message.success('Сезонный тариф обновлён')
      setTariffModalOpen(false)
      setEditingTariff(null)
      tariffForm.resetFields()
      invalidateTariffs()
    },
    onError: () => message.error('Не удалось обновить тариф'),
  })

  const deleteTariffMutation = useMutation({
    mutationFn: async (id: string) => {
      return axiosInstance.delete(`/seasonal-tariffs/${id}`)
    },
    onSuccess: () => {
      message.success('Сезонный тариф удалён')
      invalidateTariffs()
    },
    onError: () => message.error('Не удалось удалить тариф'),
  })

  const handleOpenCreateTariff = () => {
    setEditingTariff(null)
    tariffForm.resetFields()
    tariffForm.setFieldsValue({ is_active: true, multiplier: 1.0 })
    setTariffModalOpen(true)
  }

  const handleOpenEditTariff = (tariff: SeasonalTariff) => {
    setEditingTariff(tariff)
    tariffForm.setFieldsValue({
      name: tariff.name,
      date_range: [dayjs(tariff.date_from), dayjs(tariff.date_to)],
      multiplier: tariff.multiplier,
      is_active: tariff.is_active,
    })
    setTariffModalOpen(true)
  }

  const handleTariffSubmit = (values: TariffFormValues) => {
    const payload = {
      name: values.name,
      date_from: values.date_range[0].format('YYYY-MM-DD'),
      date_to: values.date_range[1].format('YYYY-MM-DD'),
      multiplier: values.multiplier,
      is_active: values.is_active,
    }

    if (editingTariff) {
      updateTariffMutation.mutate({ id: editingTariff.id, data: payload })
    } else {
      createTariffMutation.mutate(payload)
    }
  }

  const handleToggleTariffActive = (tariff: SeasonalTariff) => {
    updateTariffMutation.mutate({
      id: tariff.id,
      data: {
        name: tariff.name,
        date_from: tariff.date_from,
        date_to: tariff.date_to,
        multiplier: tariff.multiplier,
        is_active: !tariff.is_active,
      },
    })
  }

  const tariffColumns = [
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Период',
      key: 'period',
      render: (_: unknown, record: SeasonalTariff) =>
        `${dayjs(record.date_from).format('DD.MM.YYYY')} – ${dayjs(record.date_to).format('DD.MM.YYYY')}`,
    },
    {
      title: 'Множитель',
      dataIndex: 'multiplier',
      key: 'multiplier',
      width: 120,
      render: (val: number) => {
        const isDiscount = val < 1
        const percent = Math.round(Math.abs(val - 1) * 100)
        return (
          <span style={{ color: isDiscount ? '#15803d' : val > 1 ? '#b42318' : undefined }}>
            x{val} {percent > 0 && `(${isDiscount ? '-' : '+'}${percent}%)`}
          </span>
        )
      },
    },
    {
      title: 'Активно',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 90,
      render: (active: boolean, record: SeasonalTariff) => (
        <Switch
          checked={active}
          onChange={() => handleToggleTariffActive(record)}
          size="small"
        />
      ),
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 100,
      render: (_: unknown, record: SeasonalTariff) => (
        <Space>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => handleOpenEditTariff(record)}
            size="small"
          />
          <Popconfirm
            title="Удалить тариф?"
            description="Это действие нельзя отменить"
            onConfirm={() => deleteTariffMutation.mutate(record.id)}
            okText="Удалить"
            cancelText="Отмена"
          >
            <Button type="text" danger icon={<DeleteOutlined />} size="small" />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  if (!selectedBathhouseId) {
    return (
      <div>
        <Title level={3}>Правила ценообразования</Title>
        <div style={{ textAlign: 'center', padding: 40, color: 'var(--rh-text-muted)' }}>
          Выберите баню для управления ценами
        </div>
      </div>
    )
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>Правила ценообразования</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleOpenCreate}>
          Добавить правило
        </Button>
      </div>

      {recommendationData && !recommendationDismissed && (
        <Card
          size="small"
          style={{ marginBottom: 16 }}
          title={<><BulbOutlined /> Рекомендация по цене</>}
          extra={
            <Space>
              <Button
                type="primary"
                size="small"
                icon={<CheckOutlined />}
                loading={acceptRecommendationMutation.isPending}
                onClick={() => acceptRecommendationMutation.mutate(recommendationData.recommended_price)}
              >
                Применить
              </Button>
              <Button
                size="small"
                icon={<CloseOutlined />}
                onClick={() => setRecommendationDismissed(true)}
              >
                Скрыть
              </Button>
            </Space>
          }
        >
          <div style={{ display: 'flex', gap: 24, flexWrap: 'wrap', alignItems: 'center' }}>
            <Statistic title="Текущая цена/час" value={formatPrice(recommendationData.current_price)} />
            <Statistic
              title="Рекомендуемая цена/час"
              value={formatPrice(recommendationData.recommended_price)}
              styles={{ content: { color: recommendationData.coefficient > 1 ? '#15803d' : recommendationData.coefficient < 1 ? '#b42318' : undefined } }}
              prefix={recommendationData.coefficient > 1 ? <ArrowUpOutlined /> : recommendationData.coefficient < 1 ? <ArrowDownOutlined /> : undefined}
            />
            <Statistic title="Коэффициент" value={`x${recommendationData.coefficient}`} />
            <Statistic title="Средняя цена в районе" value={formatPrice(recommendationData.avg_area_price)} />
            <Statistic title="Загрузка" value={`${Math.round(recommendationData.occupancy_rate * 100)}%`} />
            <Statistic title="Спрос" value={demandTrendLabel[recommendationData.demand_trend] ?? recommendationData.demand_trend} />
          </div>
          <Alert
            style={{ marginTop: 12 }}
            type="info"
            showIcon
            title={recommendationData.recommendation_basis}
          />
        </Card>
      )}

      <Table
        dataSource={rules}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        locale={{ emptyText: 'Нет правил ценообразования' }}
        pagination={false}
      />

      <Divider />

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>Сезонные тарифы</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleOpenCreateTariff}>
          Добавить тариф
        </Button>
      </div>

      <Card size="small" styles={{ body: { padding: 0 } }}>
        <Table
          dataSource={tariffs}
          columns={tariffColumns}
          rowKey="id"
          loading={tariffsLoading}
          locale={{ emptyText: 'Нет сезонных тарифов' }}
          pagination={false}
          size="small"
        />
      </Card>
      <Typography.Text type="secondary" style={{ display: 'block', marginTop: 8 }}>
        Сезонные тарифы умножают базовую цену на период дат. При пересечении нескольких тарифов применяется наибольший множитель.
      </Typography.Text>

      {/* Pricing Rule Modal */}
      <Modal
        title={editingRule ? 'Редактировать правило' : 'Новое правило'}
        open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditingRule(null) }}
        footer={null}
        destroyOnHidden
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{ is_active: true, priority: 0, multiplier: 1.0 }}
        >
          <Form.Item name="name" label="Название" rules={[{ required: true, message: 'Укажите название' }]}>
            <Input placeholder="Например: Выходные повышение" />
          </Form.Item>

          <Form.Item name="type" label="Тип правила" rules={[{ required: true, message: 'Выберите тип' }]}>
            <Select options={RULE_TYPES} placeholder="Выберите тип" />
          </Form.Item>

          <Form.Item
            name="multiplier"
            label="Множитель цены"
            rules={[{ required: true, message: 'Укажите множитель' }]}
            extra="1.0 = без изменений, 1.5 = +50%, 0.8 = -20%"
          >
            <InputNumber min={0.01} max={999.99} step={0.1} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="priority"
            label="Приоритет"
            extra="Правила с более высоким приоритетом применяются первыми"
          >
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>

          {(selectedType === 'weekday') && (
            <Form.Item
              name="days_of_week"
              label="Дни недели"
              rules={[{ required: true, message: 'Выберите дни' }]}
            >
              <Select mode="multiple" options={WEEKDAY_OPTIONS} placeholder="Выберите дни" />
            </Form.Item>
          )}

          {(selectedType === 'weekend') && (
            <Form.Item
              name="days_of_week"
              label="Дни недели"
              rules={[{ required: true, message: 'Выберите дни' }]}
            >
              <Select mode="multiple" options={WEEKEND_OPTIONS} placeholder="Выберите дни" />
            </Form.Item>
          )}

          {(selectedType === 'per_day') && (
            <Form.Item
              name="days_of_week"
              label="Дни недели"
              rules={[{ required: true, message: 'Выберите дни' }]}
              extra="Выберите любые дни недели для применения правила"
            >
              <Select mode="multiple" options={ALL_DAY_OPTIONS} placeholder="Выберите дни" />
            </Form.Item>
          )}

          {(selectedType === 'time_range') && (
            <Space>
              <Form.Item
                name="time_from"
                label="Время от"
                rules={[{ required: true, message: 'Укажите время' }]}
              >
                <TimePicker format="HH:mm" minuteStep={30} />
              </Form.Item>
              <Form.Item
                name="time_to"
                label="Время до"
                rules={[{ required: true, message: 'Укажите время' }]}
              >
                <TimePicker format="HH:mm" minuteStep={30} />
              </Form.Item>
            </Space>
          )}

          {(selectedType === 'holiday' || selectedType === 'season') && (
            <Form.Item
              name="date_range"
              label="Диапазон дат"
              rules={[{ required: true, message: 'Укажите даты' }]}
            >
              <DatePicker.RangePicker style={{ width: '100%' }} format="DD.MM.YYYY" />
            </Form.Item>
          )}

          <Form.Item name="is_active" label="Активно" valuePropName="checked">
            <Switch />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={createMutation.isPending || updateMutation.isPending}
              >
                {editingRule ? 'Сохранить' : 'Создать'}
              </Button>
              <Button onClick={() => { setModalOpen(false); setEditingRule(null) }}>
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Seasonal Tariff Modal */}
      <Modal
        title={editingTariff ? 'Редактировать сезонный тариф' : 'Новый сезонный тариф'}
        open={tariffModalOpen}
        onCancel={() => { setTariffModalOpen(false); setEditingTariff(null) }}
        footer={null}
        destroyOnHidden
      >
        <Form
          form={tariffForm}
          layout="vertical"
          onFinish={handleTariffSubmit}
          initialValues={{ is_active: true, multiplier: 1.0 }}
        >
          <Form.Item name="name" label="Название" rules={[{ required: true, message: 'Укажите название' }]}>
            <Input placeholder="Например: Летний сезон, Новогодние каникулы" />
          </Form.Item>

          <Form.Item
            name="date_range"
            label="Период действия"
            rules={[{ required: true, message: 'Укажите период' }]}
          >
            <DatePicker.RangePicker style={{ width: '100%' }} format="DD.MM.YYYY" />
          </Form.Item>

          <Form.Item
            name="multiplier"
            label="Множитель цены"
            rules={[{ required: true, message: 'Укажите множитель' }]}
            extra="1.0 = без изменений, 1.3 = +30%, 0.8 = -20%. Макс: 10.0"
          >
            <InputNumber min={0.01} max={10.0} step={0.1} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item name="is_active" label="Активно" valuePropName="checked">
            <Switch />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={createTariffMutation.isPending || updateTariffMutation.isPending}
              >
                {editingTariff ? 'Сохранить' : 'Создать'}
              </Button>
              <Button onClick={() => { setTariffModalOpen(false); setEditingTariff(null) }}>
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
