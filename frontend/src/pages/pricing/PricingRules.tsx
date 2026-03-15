import { useState } from 'react'
import {
  App,
  Button,
  DatePicker,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  TimePicker,
  Typography,
} from 'antd'
import {
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import {
  useGetMyBathhousesIdPricingRules,
  usePostMyBathhousesIdPricingRules,
  usePutPricingRulesId,
  useDeletePricingRulesId,
} from '@/api/generated/pricing/pricing'
import type { InternalHandlerPricingRuleResponse } from '@/api/generated/model'
import { useBathhouseStore } from '@/stores/bathhouse'
import { useQueryClient } from '@tanstack/react-query'
import { formatDayOfWeek } from '@/lib/format'

const { Title } = Typography

const RULE_TYPES = [
  { value: 'weekday', label: 'Будние дни' },
  { value: 'weekend', label: 'Выходные' },
  { value: 'holiday', label: 'Праздники' },
  { value: 'time_range', label: 'Диапазон времени' },
  { value: 'season', label: 'Сезон' },
]

const RULE_TYPE_COLORS: Record<string, string> = {
  weekday: 'blue',
  weekend: 'green',
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

export default function PricingRules() {
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)
  const queryClient = useQueryClient()
  const { message } = App.useApp()
  const [form] = Form.useForm<RuleFormValues>()

  const [modalOpen, setModalOpen] = useState(false)
  const [editingRule, setEditingRule] = useState<InternalHandlerPricingRuleResponse | null>(null)

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
          <span style={{ color: isDiscount ? '#52c41a' : val > 1 ? '#f5222d' : undefined }}>
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

  if (!selectedBathhouseId) {
    return (
      <div>
        <Title level={3}>Правила ценообразования</Title>
        <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
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

      <Table
        dataSource={rules}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        locale={{ emptyText: 'Нет правил ценообразования' }}
        pagination={false}
      />

      <Modal
        title={editingRule ? 'Редактировать правило' : 'Новое правило'}
        open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditingRule(null) }}
        footer={null}
        destroyOnClose
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
    </div>
  )
}
