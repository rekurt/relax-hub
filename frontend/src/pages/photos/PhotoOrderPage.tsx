import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Input,
  Modal,
  Space,
  Spin,
  Steps,
  Table,
  Tag,
  Typography,
} from '@/components/design/system'
import { PlusOutlined, ReloadOutlined } from '@/components/design/icons'
import { useBathhouseStore } from '@/stores/bathhouse'
import { axiosInstance } from '@/api/axios-instance'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { formatPrice } from '@/lib/format'
import dayjs from 'dayjs'
import type { ColumnsType } from '@/components/design/types'
import EmptyState from '@/components/EmptyState'
import PageHeader from '@/components/PageHeader'

const { Text, Paragraph } = Typography

interface PhotoOrder {
  id: string
  owner_id: string
  bathhouse_id: string
  region: string
  status: string
  photographer_name?: string
  price: number
  scheduled_at?: string
  notes?: string
  admin_notes?: string
  created_at: string
  updated_at: string
}

interface PhotoOrderListResponse {
  success: boolean
  data: PhotoOrder[]
  meta?: {
    page: number
    page_size: number
    total_count: number
    total_pages: number
  }
}

const STATUS_CONFIG: Record<string, { label: string; color: string; step: number }> = {
  requested: { label: 'Заявка отправлена', color: 'processing', step: 0 },
  confirmed: { label: 'Подтверждено', color: 'warning', step: 1 },
  completed: { label: 'Выполнено', color: 'success', step: 2 },
  cancelled: { label: 'Отменено', color: 'default', step: -1 },
}

export default function PhotoOrderPage() {
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)
  const queryClient = useQueryClient()
  const { message, modal } = App.useApp()
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [notes, setNotes] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading, refetch } = useQuery<PhotoOrderListResponse>({
    queryKey: ['photo-orders', page],
    queryFn: () =>
      axiosInstance
        .get('/my/photo-orders', { params: { page, page_size: 10 } })
        .then((r) => r.data),
  })

  const createMutation = useMutation({
    mutationFn: (payload: { notes: string }) =>
      axiosInstance
        .post(`/my/bathhouses/${selectedBathhouseId}/photo-order`, {
          region: 'RU',
          notes: payload.notes,
        })
        .then((r) => r.data),
    onSuccess: () => {
      message.success('Заявка на фотосъёмку отправлена')
      setCreateModalOpen(false)
      setNotes('')
      queryClient.invalidateQueries({ queryKey: ['photo-orders'] })
    },
    onError: () => {
      message.error('Не удалось создать заявку')
    },
  })

  const cancelMutation = useMutation({
    mutationFn: (id: string) =>
      axiosInstance.post(`/my/photo-orders/${id}/cancel`).then((r) => r.data),
    onSuccess: () => {
      message.success('Заявка отменена')
      queryClient.invalidateQueries({ queryKey: ['photo-orders'] })
    },
    onError: () => {
      message.error('Не удалось отменить заявку')
    },
  })

  const handleCancel = (id: string) => {
    modal.confirm({
      title: 'Отменить заявку на фотосъёмку?',
      content: 'Это действие нельзя отменить.',
      okText: 'Да, отменить',
      cancelText: 'Нет',
      onOk: () => cancelMutation.mutateAsync(id),
    })
  }

  const columns: ColumnsType<PhotoOrder> = [
    {
      title: 'Дата заявки',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (v: string) => dayjs(v).format('DD.MM.YYYY HH:mm'),
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const cfg = STATUS_CONFIG[status] ?? { label: status, color: 'default' }
        return <Tag color={cfg.color}>{cfg.label}</Tag>
      },
    },
    {
      title: 'Фотограф',
      dataIndex: 'photographer_name',
      key: 'photographer_name',
      render: (v: string) => v || <Text type="secondary">Не назначен</Text>,
    },
    {
      title: 'Дата съёмки',
      dataIndex: 'scheduled_at',
      key: 'scheduled_at',
      render: (v?: string) =>
        v ? dayjs(v).format('DD.MM.YYYY HH:mm') : <Text type="secondary">-</Text>,
    },
    {
      title: 'Стоимость',
      dataIndex: 'price',
      key: 'price',
      render: (v: number) => (v > 0 ? formatPrice(v) : <Text type="secondary">-</Text>),
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_: unknown, record: PhotoOrder) =>
        record.status === 'requested' ? (
          <Button size="small" danger onClick={() => handleCancel(record.id)}>
            Отменить
          </Button>
        ) : null,
    },
  ]

  if (!selectedBathhouseId) {
    return (
      <div className="rh-page-stack">
        <PageHeader
          title="Профессиональная фотосъёмка"
          description="Заявки, статусы и согласование съемки для выбранного объекта."
          size="compact"
        />
        <Card className="rh-admin-detail-card">
          <EmptyState description="Выберите объект для управления фотосъёмкой" />
        </Card>
      </div>
    )
  }

  const orders = data?.data ?? []

  return (
    <div className="rh-page-stack">
      <PageHeader
        eyebrow="Контент объекта"
        title="Профессиональная фотосъёмка"
        description="Закажите съемку, отслеживайте назначение фотографа и завершение работ."
        size="compact"
        extra={
          <Space wrap>
          <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
            Обновить
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateModalOpen(true)}>
            Заказать съёмку
          </Button>
          </Space>
        }
      />

      <Card className="rh-admin-detail-card rh-photo-order-intro">
        <Paragraph className="rh-card-intro-text">
          Закажите профессиональную фотосъёмку вашего объекта. Качественные фотографии увеличивают
          конверсию в бронирования до 40%. После отправки заявки администратор назначит фотографа
          и согласует дату съёмки. Оплата списывается с кошелька после завершения работы.
        </Paragraph>
        <Steps
          size="small"
          className="rh-photo-order-steps"
          items={[
            { title: 'Заявка' },
            { title: 'Подтверждение' },
            { title: 'Съёмка' },
            { title: 'Готово' },
          ]}
        />
      </Card>

      <Spin spinning={isLoading}>
        <Table
          dataSource={orders}
          columns={columns}
          rowKey="id"
          pagination={
            data?.meta
              ? {
                  current: data.meta.page,
                  pageSize: data.meta.page_size,
                  total: data.meta.total_count,
                  onChange: (p) => setPage(p),
                  showSizeChanger: false,
                }
              : false
          }
          expandable={{
            expandedRowRender: (record) => (
              <Space orientation="vertical" size="small">
                {record.notes && (
                  <Text>
                    <strong>Ваши пожелания:</strong> {record.notes}
                  </Text>
                )}
                {record.admin_notes && (
                  <Text>
                    <strong>Комментарий администратора:</strong> {record.admin_notes}
                  </Text>
                )}
              </Space>
            ),
            rowExpandable: (record) => !!(record.notes || record.admin_notes),
          }}
          locale={{
            emptyText: <EmptyState description="У вас пока нет заявок на фотосъёмку" />,
          }}
        />
      </Spin>

      <Modal
        title="Заказать профессиональную фотосъёмку"
        open={createModalOpen}
        onCancel={() => setCreateModalOpen(false)}
        onOk={() => createMutation.mutate({ notes })}
        confirmLoading={createMutation.isPending}
        okText="Отправить заявку"
        cancelText="Отмена"
      >
        <Paragraph>
          Опишите ваши пожелания: предпочтительное время, количество помещений, особые требования.
        </Paragraph>
        <Input.TextArea
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          placeholder="Например: съёмка парной, комнаты отдыха и бассейна. Предпочтительно в будний день."
          className="rh-full-width"
          rows={4}
        />
      </Modal>
    </div>
  )
}
