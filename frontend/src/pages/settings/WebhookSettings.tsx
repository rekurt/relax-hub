import { useState } from 'react'
import {
  Typography,
  Card,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  Switch,
  Space,
  Tag,
  Popconfirm,
  App,
  Empty,
  Collapse,
} from 'antd'
import {
  PlusOutlined,
  DeleteOutlined,
  EditOutlined,
  SendOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons'
import { axiosInstance } from '@/api/axios-instance'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'

const { Title, Text } = Typography

interface Webhook {
  id: string
  owner_id: string
  url: string
  events: string[]
  is_active: boolean
  created_at: string
  updated_at: string
}

interface WebhookDelivery {
  id: string
  webhook_id: string
  event_type: string
  status: string
  http_status: number
  error_message?: string
  attempt_count: number
  next_retry_at?: string
  created_at: string
}

const EVENT_OPTIONS = [
  { value: 'booking.created', label: 'Бронирование создано' },
  { value: 'booking.confirmed', label: 'Бронирование подтверждено' },
  { value: 'booking.cancelled', label: 'Бронирование отменено' },
  { value: 'booking.completed', label: 'Бронирование завершено' },
  { value: 'payment.received', label: 'Платёж получен' },
]

const STATUS_TAGS: Record<string, { color: string; icon: React.ReactNode; text: string }> = {
  success: { color: 'success', icon: <CheckCircleOutlined />, text: 'Доставлено' },
  failed: { color: 'error', icon: <CloseCircleOutlined />, text: 'Ошибка' },
  pending: { color: 'processing', icon: <ClockCircleOutlined />, text: 'Ожидание' },
}

function useWebhooks(page: number) {
  return useQuery({
    queryKey: ['webhooks', page],
    queryFn: async () => {
      const { data } = await axiosInstance.get('/my/webhooks', { params: { page, page_size: 20 } })
      return data
    },
  })
}

function useWebhookDeliveries(webhookId: string | null) {
  return useQuery({
    queryKey: ['webhook-deliveries', webhookId],
    queryFn: async () => {
      const { data } = await axiosInstance.get(`/my/webhooks/${webhookId}/deliveries`, { params: { page: 1, page_size: 20 } })
      return data
    },
    enabled: !!webhookId,
  })
}

export default function WebhookSettings() {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingWebhook, setEditingWebhook] = useState<Webhook | null>(null)
  const [selectedWebhookId, setSelectedWebhookId] = useState<string | null>(null)
  const [form] = Form.useForm()

  const { data: webhooksData, isLoading } = useWebhooks(page)
  const { data: deliveriesData, isLoading: deliveriesLoading } = useWebhookDeliveries(selectedWebhookId)

  const webhooks: Webhook[] = webhooksData?.data ?? []
  const meta = webhooksData?.meta

  const createMutation = useMutation({
    mutationFn: (values: { url: string; secret: string; events: string[] }) =>
      axiosInstance.post('/my/webhooks', values),
    onSuccess: () => {
      message.success('Вебхук создан')
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
      setModalOpen(false)
      form.resetFields()
    },
    onError: () => message.error('Ошибка при создании вебхука'),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, ...values }: { id: string; url: string; secret: string; events: string[]; is_active: boolean }) =>
      axiosInstance.put(`/my/webhooks/${id}`, values),
    onSuccess: () => {
      message.success('Вебхук обновлён')
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
      setModalOpen(false)
      setEditingWebhook(null)
      form.resetFields()
    },
    onError: () => message.error('Ошибка при обновлении вебхука'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => axiosInstance.delete(`/my/webhooks/${id}`),
    onSuccess: () => {
      message.success('Вебхук удалён')
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
    },
    onError: () => message.error('Ошибка при удалении вебхука'),
  })

  const testMutation = useMutation({
    mutationFn: (id: string) => axiosInstance.post(`/my/webhooks/${id}/test`),
    onSuccess: () => {
      message.success('Тестовое событие отправлено')
      if (selectedWebhookId) {
        queryClient.invalidateQueries({ queryKey: ['webhook-deliveries', selectedWebhookId] })
      }
    },
    onError: () => message.error('Ошибка отправки'),
  })

  const handleSubmit = (values: { url: string; secret: string; events: string[]; is_active?: boolean }) => {
    if (editingWebhook) {
      updateMutation.mutate({ id: editingWebhook.id, is_active: values.is_active ?? true, ...values })
    } else {
      createMutation.mutate(values)
    }
  }

  const openEdit = (webhook: Webhook) => {
    setEditingWebhook(webhook)
    form.setFieldsValue({
      url: webhook.url,
      secret: '',
      events: webhook.events,
      is_active: webhook.is_active,
    })
    setModalOpen(true)
  }

  const openCreate = () => {
    setEditingWebhook(null)
    form.resetFields()
    setModalOpen(true)
  }

  const deliveries: WebhookDelivery[] = deliveriesData?.data ?? []

  const columns = [
    {
      title: 'URL',
      dataIndex: 'url',
      key: 'url',
      ellipsis: true,
      render: (url: string) => <Text copyable style={{ maxWidth: 300 }}>{url}</Text>,
    },
    {
      title: 'События',
      dataIndex: 'events',
      key: 'events',
      render: (events: string[]) => (
        <Space wrap>
          {events.map(e => {
            const opt = EVENT_OPTIONS.find(o => o.value === e)
            return <Tag key={e}>{opt?.label ?? e}</Tag>
          })}
        </Space>
      ),
    },
    {
      title: 'Статус',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (active: boolean) => active ? <Tag color="success">Активен</Tag> : <Tag>Неактивен</Tag>,
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_: unknown, record: Webhook) => (
        <Space>
          <Button icon={<EditOutlined />} size="small" onClick={() => openEdit(record)} />
          <Button icon={<SendOutlined />} size="small" onClick={() => testMutation.mutate(record.id)} loading={testMutation.isPending}>
            Тест
          </Button>
          <Popconfirm title="Удалить вебхук?" onConfirm={() => deleteMutation.mutate(record.id)}>
            <Button icon={<DeleteOutlined />} size="small" danger />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const deliveryColumns = [
    { title: 'Событие', dataIndex: 'event_type', key: 'event_type' },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const cfg = STATUS_TAGS[status] ?? { color: 'default', icon: null, text: status }
        return <Tag color={cfg.color} icon={cfg.icon}>{cfg.text}</Tag>
      },
    },
    { title: 'HTTP', dataIndex: 'http_status', key: 'http_status' },
    { title: 'Попытки', dataIndex: 'attempt_count', key: 'attempt_count' },
    {
      title: 'Ошибка',
      dataIndex: 'error_message',
      key: 'error_message',
      ellipsis: true,
    },
    {
      title: 'Дата',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (d: string) => new Date(d).toLocaleString('ru-RU'),
    },
  ]

  return (
    <div style={{ padding: 24, maxWidth: 1200, margin: '0 auto' }}>
      <Space style={{ width: '100%', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>Вебхуки</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          Добавить вебхук
        </Button>
      </Space>

      <Card>
        <Table
          dataSource={webhooks}
          columns={columns}
          rowKey="id"
          loading={isLoading}
          locale={{ emptyText: <Empty description="Нет вебхуков. Добавьте первый вебхук для получения событий в вашу CRM." /> }}
          pagination={meta ? {
            current: meta.page,
            pageSize: meta.page_size,
            total: meta.total_count,
            onChange: setPage,
          } : false}
          expandable={{
            expandedRowRender: (record: Webhook) => (
              <Collapse
                ghost
                items={[{
                  key: 'deliveries',
                  label: 'История доставок',
                  children: (
                    <Table
                      dataSource={selectedWebhookId === record.id ? deliveries : []}
                      columns={deliveryColumns}
                      rowKey="id"
                      size="small"
                      loading={selectedWebhookId === record.id && deliveriesLoading}
                      pagination={false}
                      locale={{ emptyText: 'Нет доставок' }}
                    />
                  ),
                }]}
              />
            ),
            onExpand: (expanded, record) => {
              setSelectedWebhookId(expanded ? record.id : null)
            },
          }}
        />
      </Card>

      <Modal
        title={editingWebhook ? 'Редактировать вебхук' : 'Новый вебхук'}
        open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditingWebhook(null) }}
        onOk={() => form.submit()}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        okText={editingWebhook ? 'Сохранить' : 'Создать'}
        cancelText="Отмена"
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="url" label="URL" rules={[{ required: true, message: 'Укажите URL' }, { type: 'url', message: 'Некорректный URL' }]}>
            <Input placeholder="https://your-crm.com/webhooks" />
          </Form.Item>
          <Form.Item name="secret" label="Секретный ключ" rules={editingWebhook ? [] : [{ required: true, message: 'Укажите секрет' }]}>
            <Input.Password placeholder={editingWebhook ? 'Оставьте пустым, чтобы не менять' : 'Ключ для HMAC-SHA256 подписи'} />
          </Form.Item>
          <Form.Item name="events" label="События" rules={[{ required: true, message: 'Выберите хотя бы одно событие' }]}>
            <Select mode="multiple" placeholder="Выберите события" options={EVENT_OPTIONS} />
          </Form.Item>
          {editingWebhook && (
            <Form.Item name="is_active" label="Активен" valuePropName="checked">
              <Switch />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </div>
  )
}
