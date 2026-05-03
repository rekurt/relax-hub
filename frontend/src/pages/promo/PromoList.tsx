import { useMemo, useState } from 'react'
import {
  App,
  Button,
  Card,
  DatePicker,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  Tooltip,
} from '@/components/design/system'
import { CopyOutlined, DeleteOutlined, PlusOutlined } from '@/components/design/icons'
import dayjs from 'dayjs'
import { useQueryClient } from '@tanstack/react-query'
import {
  useDeletePromoCodesId,
  useGetMyBathhousesIdPromoCodes,
  usePostMyBathhousesIdPromoCodes,
} from '@/api/generated/promo-codes/promo-codes'
import type { InternalHandlerPromoResponse } from '@/api/generated/model'
import { formatPrice } from '@/lib/format'
import PageHeader from '@/components/PageHeader'
import { useBathhouseStore } from '@/stores/bathhouse'
import EmptyState from '@/components/EmptyState'

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
  const selectedBathhouseId = useBathhouseStore((state) => state.selectedBathhouseId)
  const queryClient = useQueryClient()
  const { message } = App.useApp()
  const [form] = Form.useForm<PromoFormValues>()
  const [modalOpen, setModalOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<InternalHandlerPromoResponse | null>(null)
  const [page, setPage] = useState(1)
  const pageSize = 20

  const { data, isLoading } = useGetMyBathhousesIdPromoCodes(
    selectedBathhouseId ?? '',
    { page, page_size: pageSize },
    { query: { enabled: !!selectedBathhouseId } },
  )

  const promos = useMemo(() => data?.data ?? [], [data?.data])
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

  const selectedType = Form.useWatch('type', form)

  const stats = useMemo(() => {
    return promos.reduce(
      (acc, promo) => {
        if (promo.is_active && !isPromoExpired(promo) && !isPromoMaxed(promo)) acc.active += 1
        if (isPromoMaxed(promo)) acc.maxed += 1
        if (!promo.valid_until) acc.openEnded += 1
        return acc
      },
      { active: 0, maxed: 0, openEnded: 0 },
    )
  }, [promos])

  const handleOpenCreate = () => {
    form.resetFields()
    setModalOpen(true)
  }

  const handleSubmit = (values: PromoFormValues) => {
    if (!selectedBathhouseId) return

    createMutation.mutate({
      id: selectedBathhouseId,
      data: {
        code: values.code,
        type: values.type,
        value: values.type === 'fixed_amount' ? Math.round(values.value * 100) : values.value,
        max_uses: values.max_uses,
        min_amount: values.min_amount ? Math.round(values.min_amount * 100) : undefined,
        valid_from: values.validity?.[0]?.toISOString(),
        valid_until: values.validity?.[1]?.toISOString(),
      },
    })
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

  const columns = [
    {
      title: 'Код',
      dataIndex: 'code',
      key: 'code',
      render: (code: string) => (
        <Space>
          <Tag className="rh-code-tag">{code}</Tag>
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
        const label = PROMO_TYPES.find((promoType) => promoType.value === type)?.label ?? type
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
      render: (value: number | undefined) => value ? formatPrice(value) : '—',
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
        <Button
          type="text"
          danger
          icon={<DeleteOutlined />}
          size="small"
          onClick={() => setDeleteTarget(record)}
        />
      ),
    },
  ]

  if (!selectedBathhouseId) {
    return (
      <div className="rh-stack">
        <PageHeader
          eyebrow="Маркетинг"
          title="Промокоды"
          description="Скидки должны управляться в контексте конкретной бани, иначе пользователь не понимает, где код реально применится."
        />
        <Card>
          <div className="rh-feed-empty">
            Выберите баню для управления промокодами
          </div>
        </Card>
      </div>
    )
  }

  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="Маркетинг"
        title="Промокоды"
        description="Экран собран как операционный центр: сверху краткая картина по акциям, ниже рабочая таблица с копированием кодов, сроками и статусом."
        extra={(
          <Button type="primary" icon={<PlusOutlined />} onClick={handleOpenCreate}>
            Создать промокод
          </Button>
        )}
      />

      <section className="rh-hero-panel">
        <div className="rh-hero-panel__eyebrow">Сводка</div>
        <h2 className="rh-hero-panel__title">Какие промо реально работают прямо сейчас</h2>
        <div className="rh-hero-panel__description">
          Вместо сухой таблицы владелец сначала видит активные, исчерпанные и бессрочные коды. Это ускоряет решение: продлевать, выключать или запускать новую кампанию.
        </div>
        <div className="rh-stat-grid">
          <div className="rh-stat-tile">
            <span className="rh-stat-tile__eyebrow">Активны</span>
            <div className="rh-stat-tile__value">{stats.active}</div>
            <span className="rh-stat-tile__hint">Доступны клиенту без ручных проверок</span>
          </div>
          <div className="rh-stat-tile">
            <span className="rh-stat-tile__eyebrow">Исчерпаны</span>
            <div className="rh-stat-tile__value">{stats.maxed}</div>
            <span className="rh-stat-tile__hint">Нуждаются в продлении или замене</span>
          </div>
          <div className="rh-stat-tile">
            <span className="rh-stat-tile__eyebrow">Бессрочные</span>
            <div className="rh-stat-tile__value">{stats.openEnded}</div>
            <span className="rh-stat-tile__hint">Следите, чтобы они не жили вечно без причины</span>
          </div>
        </div>
      </section>

      <Card className="rh-admin-detail-card">
        <div className="rh-table-shell">
          <div className="rh-toolbar">
            <div>
              <h2 className="rh-section-card__title">Все промокоды</h2>
              <div className="rh-section-card__description">
                Код, тип скидки, лимиты использования и сроки показываются в одной строке, чтобы не приходилось открывать детали для базовых действий.
              </div>
            </div>
            <div className="rh-inline-note">Всего кодов: {totalCount}</div>
          </div>

          <Table
            dataSource={promos}
            columns={columns}
            rowKey="id"
            loading={isLoading}
            locale={{
              emptyText: (
                <EmptyState description="Нет промокодов. Создайте промокод для привлечения клиентов." />
              ),
            }}
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
        </div>
      </Card>

      <Modal
        title="Новый промокод"
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        footer={null}
        destroyOnHidden
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item
            name="code"
            label="Код промокода"
            rules={[{ required: true, message: 'Укажите код' }]}
            normalize={(value: string) => (typeof value === 'string' ? value.toUpperCase() : value)}
            extra="Латинские буквы и цифры, например: SUMMER20"
          >
            <Input placeholder="SUMMER20" className="rh-monospace-control" />
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
              className="rh-full-width"
            />
          </Form.Item>

          <Form.Item name="max_uses" label="Максимум использований" extra="Оставьте пустым для неограниченного">
            <InputNumber min={1} className="rh-full-width" placeholder="Без ограничений" />
          </Form.Item>

          <Form.Item name="min_amount" label="Минимальная сумма заказа (₽)" extra="Оставьте пустым без ограничения">
            <InputNumber min={0} step={100} className="rh-full-width" placeholder="Без ограничений" />
          </Form.Item>

          <Form.Item name="validity" label="Период действия" extra="Оставьте пустым для бессрочного">
            <DatePicker.RangePicker className="rh-full-width" format="DD.MM.YYYY" />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={createMutation.isPending}>
                Создать
              </Button>
              <Button onClick={() => setModalOpen(false)}>
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      <Card>
        <div className="rh-toolbar">
          <div>
            <h2 className="rh-section-card__title">Деактивировать промокод?</h2>
            <div className="rh-section-card__description">
              {deleteTarget
                ? `Код ${deleteTarget.code} станет недоступен для использования`
                : 'Выберите код в таблице, если хотите отключить его без удаления истории'}
            </div>
          </div>
          <div className="rh-toolbar__group">
            <Button
              danger
              disabled={!deleteTarget}
              onClick={() => {
                if (deleteTarget?.id) handleDelete(deleteTarget.id)
              }}
            >
              Деактивировать
            </Button>
            <Button disabled={!deleteTarget} onClick={() => setDeleteTarget(null)}>
              Отмена
            </Button>
          </div>
        </div>
      </Card>
    </div>
  )
}

function formatPromoValue(promo: InternalHandlerPromoResponse): string {
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

function isPromoExpired(promo: InternalHandlerPromoResponse): boolean {
  if (!promo.valid_until) return false
  return dayjs(promo.valid_until).isBefore(dayjs())
}

function isPromoMaxed(promo: InternalHandlerPromoResponse): boolean {
  if (!promo.max_uses) return false
  return (promo.current_uses ?? 0) >= promo.max_uses
}

function getPromoStatus(promo: InternalHandlerPromoResponse): { label: string; color: string } {
  if (!promo.is_active) return { label: 'Неактивен', color: 'default' }
  if (isPromoExpired(promo)) return { label: 'Истёк', color: 'red' }
  if (isPromoMaxed(promo)) return { label: 'Исчерпан', color: 'orange' }
  return { label: 'Активен', color: 'green' }
}
