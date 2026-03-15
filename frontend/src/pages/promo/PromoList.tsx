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
  Table,
  Tag,
  Tooltip,
  Typography,
} from 'antd'
import {
  CopyOutlined,
  DeleteOutlined,
  PlusOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import {
  useGetMyBathhousesIdPromoCodes,
  usePostMyBathhousesIdPromoCodes,
  useDeletePromoCodesId,
} from '@/api/generated/promo-codes/promo-codes'
import type { InternalHandlerPromoResponse } from '@/api/generated/model'
import { useBathhouseStore } from '@/stores/bathhouse'
import { useQueryClient } from '@tanstack/react-query'
import { formatPrice } from '@/lib/format'

const { Title } = Typography

const PROMO_TYPES = [
  { value: 'percentage', label: 'Процент' },
  { value: 'fixed_amount', label: 'Фиксированная сумма' },
  { value: 'free_hour', label: 'Бесплатный час' },
]

const PROMO_TYPE_COLORS: Record<string, string> = {
  percentage: 'blue',
  fixed_amount: 'green',
  free_hour: 'purple',
}

interface PromoFormValues {
  code: string
  type: string
  value: number
  max_uses?: number
  min_amount?: number
  validity?: [dayjs.Dayjs, dayjs.Dayjs]
}

export default function PromoList() {
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)
  const queryClient = useQueryClient()
  const { message } = App.useApp()
  const [form] = Form.useForm<PromoFormValues>()

  const [modalOpen, setModalOpen] = useState(false)
  const [page, setPage] = useState(1)
  const pageSize = 20

  const { data, isLoading } = useGetMyBathhousesIdPromoCodes(
    selectedBathhouseId ?? '',
    { page, page_size: pageSize },
    { query: { enabled: !!selectedBathhouseId } },
  )

  const promos = data?.data ?? []
  const totalCount = data?.meta?.total_count ?? 0

  const invalidatePromos = () => {
    queryClient.invalidateQueries({
      queryKey: [`/my/bathhouses/${selectedBathhouseId}/promo-codes`],
    })
  }

  const createMutation = usePostMyBathhousesIdPromoCodes({
    mutation: {
      onSuccess: () => {
        message.success('Промокод создан')
        setModalOpen(false)
        form.resetFields()
        invalidatePromos()
      },
      onError: () => message.error('Не удалось создать промокод'),
    },
  })

  const deleteMutation = useDeletePromoCodesId({
    mutation: {
      onSuccess: () => {
        message.success('Промокод деактивирован')
        invalidatePromos()
      },
      onError: () => message.error('Не удалось деактивировать промокод'),
    },
  })

  const handleOpenCreate = () => {
    form.resetFields()
    setModalOpen(true)
  }

  const handleSubmit = (values: PromoFormValues) => {
    if (!selectedBathhouseId) return

    const payload = {
      code: values.code,
      type: values.type,
      value: values.type === 'fixed_amount' ? values.value * 100 : values.value,
      max_uses: values.max_uses,
      min_amount: values.min_amount ? values.min_amount * 100 : undefined,
      valid_from: values.validity?.[0]?.toISOString(),
      valid_until: values.validity?.[1]?.toISOString(),
    }

    createMutation.mutate({ id: selectedBathhouseId, data: payload })
  }

  const handleDelete = (promoId: string) => {
    deleteMutation.mutate({ id: promoId })
  }

  const handleCopyCode = (code: string) => {
    navigator.clipboard.writeText(code).then(
      () => message.success('Код скопирован'),
      () => message.error('Не удалось скопировать'),
    )
  }

  const formatPromoValue = (promo: InternalHandlerPromoResponse): string => {
    switch (promo.type) {
      case 'percentage':
        return `${promo.value}%`
      case 'fixed_amount':
        return formatPrice(promo.value ?? 0)
      case 'free_hour':
        return `${promo.value} ч.`
      default:
        return String(promo.value ?? '')
    }
  }

  const isPromoExpired = (promo: InternalHandlerPromoResponse): boolean => {
    if (!promo.valid_until) return false
    return dayjs(promo.valid_until).isBefore(dayjs())
  }

  const isPromoMaxed = (promo: InternalHandlerPromoResponse): boolean => {
    if (!promo.max_uses) return false
    return (promo.current_uses ?? 0) >= promo.max_uses
  }

  const getPromoStatus = (promo: InternalHandlerPromoResponse): { label: string; color: string } => {
    if (!promo.is_active) return { label: 'Неактивен', color: 'default' }
    if (isPromoExpired(promo)) return { label: 'Истёк', color: 'red' }
    if (isPromoMaxed(promo)) return { label: 'Исчерпан', color: 'orange' }
    return { label: 'Активен', color: 'green' }
  }

  const selectedType = Form.useWatch('type', form)

  const columns = [
    {
      title: 'Код',
      dataIndex: 'code',
      key: 'code',
      render: (code: string) => (
        <Space>
          <Tag style={{ fontFamily: 'monospace', fontSize: 14 }}>{code}</Tag>
          <Tooltip title="Копировать">
            <Button
              type="text"
              size="small"
              icon={<CopyOutlined />}
              onClick={() => handleCopyCode(code)}
            />
          </Tooltip>
        </Space>
      ),
    },
    {
      title: 'Тип скидки',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => {
        const label = PROMO_TYPES.find((t) => t.value === type)?.label ?? type
        return <Tag color={PROMO_TYPE_COLORS[type] ?? 'default'}>{label}</Tag>
      },
    },
    {
      title: 'Значение',
      key: 'value',
      render: (_: unknown, record: InternalHandlerPromoResponse) => formatPromoValue(record),
    },
    {
      title: 'Использовано',
      key: 'uses',
      render: (_: unknown, record: InternalHandlerPromoResponse) => {
        const current = record.current_uses ?? 0
        const max = record.max_uses
        return max ? `${current} / ${max}` : `${current} / ∞`
      },
    },
    {
      title: 'Мин. сумма',
      dataIndex: 'min_amount',
      key: 'min_amount',
      render: (val: number | undefined) => val ? formatPrice(val) : '—',
    },
    {
      title: 'Период',
      key: 'period',
      render: (_: unknown, record: InternalHandlerPromoResponse) => {
        if (!record.valid_from && !record.valid_until) return 'Бессрочно'
        const from = record.valid_from ? dayjs(record.valid_from).format('DD.MM.YYYY') : '...'
        const until = record.valid_until ? dayjs(record.valid_until).format('DD.MM.YYYY') : '...'
        return `${from} – ${until}`
      },
    },
    {
      title: 'Статус',
      key: 'status',
      render: (_: unknown, record: InternalHandlerPromoResponse) => {
        const status = getPromoStatus(record)
        return <Tag color={status.color}>{status.label}</Tag>
      },
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 80,
      render: (_: unknown, record: InternalHandlerPromoResponse) => (
        <Popconfirm
          title="Деактивировать промокод?"
          description="Промокод станет недоступен для использования"
          onConfirm={() => record.id && handleDelete(record.id)}
          okText="Деактивировать"
          cancelText="Отмена"
        >
          <Button type="text" danger icon={<DeleteOutlined />} size="small" />
        </Popconfirm>
      ),
    },
  ]

  if (!selectedBathhouseId) {
    return (
      <div>
        <Title level={3}>Промокоды</Title>
        <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
          Выберите баню для управления промокодами
        </div>
      </div>
    )
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>Промокоды</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleOpenCreate}>
          Создать промокод
        </Button>
      </div>

      <Table
        dataSource={promos}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        locale={{ emptyText: 'Нет промокодов' }}
        pagination={
          totalCount > pageSize
            ? {
                current: page,
                pageSize,
                total: totalCount,
                onChange: setPage,
                showSizeChanger: false,
              }
            : false
        }
      />

      <Modal
        title="Новый промокод"
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            name="code"
            label="Код промокода"
            rules={[{ required: true, message: 'Укажите код' }]}
            extra="Латинские буквы и цифры, например: SUMMER20"
          >
            <Input
              placeholder="SUMMER20"
              style={{ fontFamily: 'monospace' }}
              onChange={(e) => form.setFieldValue('code', e.target.value.toUpperCase())}
            />
          </Form.Item>

          <Form.Item
            name="type"
            label="Тип скидки"
            rules={[{ required: true, message: 'Выберите тип скидки' }]}
          >
            <Select options={PROMO_TYPES} placeholder="Выберите тип" />
          </Form.Item>

          <Form.Item
            name="value"
            label={
              selectedType === 'percentage' ? 'Процент скидки'
                : selectedType === 'free_hour' ? 'Количество часов'
                  : 'Сумма скидки (₽)'
            }
            rules={[{ required: true, message: 'Укажите значение' }]}
            extra={
              selectedType === 'percentage' ? 'От 1 до 100'
                : selectedType === 'free_hour' ? 'Количество бесплатных часов'
                  : 'Сумма в рублях'
            }
          >
            <InputNumber
              min={selectedType === 'percentage' ? 1 : 0.01}
              max={selectedType === 'percentage' ? 100 : undefined}
              step={selectedType === 'percentage' ? 1 : selectedType === 'free_hour' ? 1 : 100}
              style={{ width: '100%' }}
            />
          </Form.Item>

          <Form.Item name="max_uses" label="Максимум использований" extra="Оставьте пустым для неограниченного">
            <InputNumber min={1} style={{ width: '100%' }} placeholder="Без ограничений" />
          </Form.Item>

          <Form.Item name="min_amount" label="Минимальная сумма заказа (₽)" extra="Оставьте пустым без ограничения">
            <InputNumber min={0} step={100} style={{ width: '100%' }} placeholder="Без ограничений" />
          </Form.Item>

          <Form.Item name="validity" label="Период действия" extra="Оставьте пустым для бессрочного">
            <DatePicker.RangePicker style={{ width: '100%' }} format="DD.MM.YYYY" />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={createMutation.isPending}
              >
                Создать
              </Button>
              <Button onClick={() => setModalOpen(false)}>
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
