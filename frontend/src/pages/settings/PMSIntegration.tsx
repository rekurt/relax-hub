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
  Space,
  Tag,
  Popconfirm,
  App,
  Empty,
  Collapse,
  InputNumber,
  Descriptions,
} from '@/components/design/system'
import {
  PlusOutlined,
  DeleteOutlined,
  EditOutlined,
  SyncOutlined,
  ApiOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
} from '@/components/design/icons'
import { axiosInstance } from '@/api/axios-instance'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'

const { Title, Text } = Typography

interface PMSConnection {
  id: string
  owner_id: string
  bathhouse_id: string
  provider: string
  sync_direction: string
  sync_interval_minutes: number
  status: string
  last_sync_at?: string
  last_sync_error?: string
  external_id: string
  created_at: string
  updated_at: string
}

interface PMSSyncLog {
  id: string
  connection_id: string
  direction: string
  status: string
  items_synced: number
  error_message?: string
  started_at: string
  completed_at: string
}

const PROVIDER_OPTIONS = [
  { value: 'yclients', label: 'Yclients' },
  { value: 'restoplace', label: 'Restoplace' },
]

const SYNC_DIRECTION_OPTIONS = [
  { value: 'inbound', label: 'Только импорт (PMS → RelaxHub)' },
  { value: 'outbound', label: 'Только экспорт (RelaxHub → PMS)' },
  { value: 'both', label: 'Двусторонняя синхронизация' },
]

const STATUS_CONFIG: Record<string, { color: string; icon: React.ReactNode; text: string }> = {
  active: { color: 'success', icon: <CheckCircleOutlined />, text: 'Активно' },
  inactive: { color: 'default', icon: <CloseCircleOutlined />, text: 'Неактивно' },
  error: { color: 'error', icon: <ExclamationCircleOutlined />, text: 'Ошибка' },
}

const SYNC_STATUS_CONFIG: Record<string, { color: string; text: string }> = {
  success: { color: 'success', text: 'Успешно' },
  error: { color: 'error', text: 'Ошибка' },
}

function usePMSConnections(page: number) {
  return useQuery({
    queryKey: ['pms-connections', page],
    queryFn: async () => {
      const { data } = await axiosInstance.get('/my/pms-connections', { params: { page, page_size: 20 } })
      return data
    },
  })
}

function usePMSSyncLogs(connectionId: string | null) {
  return useQuery({
    queryKey: ['pms-sync-logs', connectionId],
    queryFn: async () => {
      const { data } = await axiosInstance.get(`/my/pms-connections/${connectionId}/logs`, { params: { page: 1, page_size: 20 } })
      return data
    },
    enabled: !!connectionId,
  })
}

export default function PMSIntegration() {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingConnection, setEditingConnection] = useState<PMSConnection | null>(null)
  const [selectedConnectionId, setSelectedConnectionId] = useState<string | null>(null)
  const [form] = Form.useForm()

  const { data: connectionsData, isLoading } = usePMSConnections(page)
  const { data: logsData, isLoading: logsLoading } = usePMSSyncLogs(selectedConnectionId)

  const connections: PMSConnection[] = connectionsData?.data ?? []
  const meta = connectionsData?.meta

  const createMutation = useMutation({
    mutationFn: (values: Record<string, unknown>) =>
      axiosInstance.post('/my/pms-connections', values),
    onSuccess: () => {
      message.success('PMS-подключение создано')
      queryClient.invalidateQueries({ queryKey: ['pms-connections'] })
      setModalOpen(false)
      form.resetFields()
    },
    onError: () => message.error('Ошибка при создании подключения'),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, ...values }: { id: string } & Record<string, unknown>) =>
      axiosInstance.put(`/my/pms-connections/${id}`, values),
    onSuccess: () => {
      message.success('PMS-подключение обновлено')
      queryClient.invalidateQueries({ queryKey: ['pms-connections'] })
      setModalOpen(false)
      setEditingConnection(null)
      form.resetFields()
    },
    onError: () => message.error('Ошибка при обновлении подключения'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => axiosInstance.delete(`/my/pms-connections/${id}`),
    onSuccess: () => {
      message.success('PMS-подключение удалено')
      queryClient.invalidateQueries({ queryKey: ['pms-connections'] })
    },
    onError: () => message.error('Ошибка при удалении подключения'),
  })

  const testMutation = useMutation({
    mutationFn: (id: string) => axiosInstance.post(`/my/pms-connections/${id}/test`),
    onSuccess: () => message.success('Подключение к PMS работает'),
    onError: () => message.error('Не удалось подключиться к PMS. Проверьте учётные данные.'),
  })

  const syncMutation = useMutation({
    mutationFn: (id: string) => axiosInstance.post(`/my/pms-connections/${id}/sync`),
    onSuccess: () => {
      message.success('Синхронизация выполнена')
      queryClient.invalidateQueries({ queryKey: ['pms-connections'] })
      queryClient.invalidateQueries({ queryKey: ['pms-sync-logs'] })
    },
    onError: () => message.error('Ошибка синхронизации'),
  })

  const handleSubmit = (values: Record<string, unknown>) => {
    if (editingConnection) {
      updateMutation.mutate({ id: editingConnection.id, ...values })
    } else {
      createMutation.mutate(values)
    }
  }

  const openEdit = (conn: PMSConnection) => {
    setEditingConnection(conn)
    form.setFieldsValue({
      provider: conn.provider,
      credentials: '',
      sync_direction: conn.sync_direction,
      sync_interval_minutes: conn.sync_interval_minutes,
      external_id: conn.external_id,
      status: conn.status,
    })
    setModalOpen(true)
  }

  const openCreate = () => {
    setEditingConnection(null)
    form.resetFields()
    setModalOpen(true)
  }

  const logs: PMSSyncLog[] = logsData?.data ?? []

  const columns = [
    {
      title: 'Провайдер',
      dataIndex: 'provider',
      key: 'provider',
      render: (provider: string) => {
        const opt = PROVIDER_OPTIONS.find(o => o.value === provider)
        return <Tag color="blue">{opt?.label ?? provider}</Tag>
      },
    },
    {
      title: 'Направление',
      dataIndex: 'sync_direction',
      key: 'sync_direction',
      render: (dir: string) => {
        const opt = SYNC_DIRECTION_OPTIONS.find(o => o.value === dir)
        return opt?.label ?? dir
      },
    },
    {
      title: 'Интервал',
      dataIndex: 'sync_interval_minutes',
      key: 'sync_interval_minutes',
      render: (min: number) => `${min} мин`,
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const cfg = STATUS_CONFIG[status] ?? { color: 'default', icon: null, text: status }
        return <Tag color={cfg.color} icon={cfg.icon}>{cfg.text}</Tag>
      },
    },
    {
      title: 'Последняя синхр.',
      dataIndex: 'last_sync_at',
      key: 'last_sync_at',
      render: (d: string | undefined) => d ? new Date(d).toLocaleString('ru-RU') : '—',
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_: unknown, record: PMSConnection) => (
        <Space>
          <Button icon={<EditOutlined />} size="small" onClick={() => openEdit(record)} />
          <Button
            icon={<ApiOutlined />}
            size="small"
            onClick={() => testMutation.mutate(record.id)}
            loading={testMutation.isPending}
          >
            Тест
          </Button>
          <Button
            icon={<SyncOutlined />}
            size="small"
            onClick={() => syncMutation.mutate(record.id)}
            loading={syncMutation.isPending}
          >
            Синхр.
          </Button>
          <Popconfirm title="Удалить подключение?" onConfirm={() => deleteMutation.mutate(record.id)}>
            <Button icon={<DeleteOutlined />} size="small" danger />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const logColumns = [
    {
      title: 'Направление',
      dataIndex: 'direction',
      key: 'direction',
      render: (dir: string) => {
        const opt = SYNC_DIRECTION_OPTIONS.find(o => o.value === dir)
        return opt?.label ?? dir
      },
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const cfg = SYNC_STATUS_CONFIG[status] ?? { color: 'default', text: status }
        return <Tag color={cfg.color}>{cfg.text}</Tag>
      },
    },
    { title: 'Объектов', dataIndex: 'items_synced', key: 'items_synced' },
    {
      title: 'Ошибка',
      dataIndex: 'error_message',
      key: 'error_message',
      ellipsis: true,
    },
    {
      title: 'Начало',
      dataIndex: 'started_at',
      key: 'started_at',
      render: (d: string) => new Date(d).toLocaleString('ru-RU'),
    },
    {
      title: 'Длительность',
      key: 'duration',
      render: (_: unknown, record: PMSSyncLog) => {
        const start = new Date(record.started_at).getTime()
        const end = new Date(record.completed_at).getTime()
        const diffMs = end - start
        return diffMs < 1000 ? `${diffMs} мс` : `${(diffMs / 1000).toFixed(1)} сек`
      },
    },
  ]

  return (
    <div style={{ padding: 24, maxWidth: 1200, margin: '0 auto' }}>
      <Space style={{ width: '100%', justifyContent: 'space-between', marginBottom: 16 }}>
        <div>
          <Title level={3} style={{ margin: 0 }}>Интеграция с PMS</Title>
          <Text type="secondary">Подключите Yclients или Restoplace для синхронизации бронирований</Text>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          Подключить PMS
        </Button>
      </Space>

      <Card>
        <Table
          dataSource={connections}
          columns={columns}
          rowKey="id"
          loading={isLoading}
          locale={{ emptyText: <Empty description="Нет подключений к PMS. Подключите Yclients или Restoplace для автоматической синхронизации." /> }}
          pagination={meta ? {
            current: meta.page,
            pageSize: meta.page_size,
            total: meta.total_count,
            onChange: setPage,
          } : false}
          expandable={{
            expandedRowRender: (record: PMSConnection) => (
              <div>
                {record.last_sync_error && (
                  <Descriptions size="small" style={{ marginBottom: 16 }}>
                    <Descriptions.Item label="Последняя ошибка">
                      <Text type="danger">{record.last_sync_error}</Text>
                    </Descriptions.Item>
                  </Descriptions>
                )}
                <Collapse
                  ghost
                  items={[{
                    key: 'logs',
                    label: 'История синхронизации',
                    children: (
                      <Table
                        dataSource={selectedConnectionId === record.id ? logs : []}
                        columns={logColumns}
                        rowKey="id"
                        size="small"
                        loading={selectedConnectionId === record.id && logsLoading}
                        pagination={false}
                        locale={{ emptyText: 'Нет записей синхронизации' }}
                      />
                    ),
                  }]}
                />
              </div>
            ),
            onExpand: (expanded, record) => {
              setSelectedConnectionId(expanded ? record.id : null)
            },
          }}
        />
      </Card>

      <Modal
        title={editingConnection ? 'Редактировать подключение' : 'Новое подключение к PMS'}
        open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditingConnection(null) }}
        onOk={() => form.submit()}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        okText={editingConnection ? 'Сохранить' : 'Подключить'}
        cancelText="Отмена"
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          {!editingConnection && (
            <Form.Item name="bathhouse_id" label="ID бани" rules={[{ required: true, message: 'Укажите ID бани' }]}>
              <Input placeholder="UUID бани для подключения" />
            </Form.Item>
          )}
          <Form.Item name="provider" label="Провайдер" rules={[{ required: true, message: 'Выберите провайдера' }]}>
            <Select placeholder="Выберите PMS" options={PROVIDER_OPTIONS} />
          </Form.Item>
          <Form.Item
            name="credentials"
            label="Учётные данные (JSON)"
            rules={editingConnection ? [] : [{ required: true, message: 'Укажите учётные данные' }]}
            extra={
              form.getFieldValue('provider') === 'yclients'
                ? 'Формат: {"partner_token": "...", "user_token": "..."}'
                : 'Формат: {"api_key": "..."}'
            }
          >
            <Input.TextArea
              rows={3}
              placeholder={editingConnection ? 'Оставьте пустым, чтобы не менять' : '{"partner_token": "...", "user_token": "..."}'}
            />
          </Form.Item>
          <Form.Item name="external_id" label="ID объекта в PMS" rules={[{ required: true, message: 'Укажите ID в PMS' }]}>
            <Input placeholder="ID вашего объекта в системе PMS" />
          </Form.Item>
          <Form.Item name="sync_direction" label="Направление синхронизации" initialValue="both">
            <Select options={SYNC_DIRECTION_OPTIONS} />
          </Form.Item>
          <Form.Item name="sync_interval_minutes" label="Интервал синхронизации (мин)" initialValue={15}>
            <InputNumber min={5} max={1440} style={{ width: '100%' }} />
          </Form.Item>
          {editingConnection && (
            <Form.Item name="status" label="Статус">
              <Select options={[
                { value: 'active', label: 'Активно' },
                { value: 'inactive', label: 'Неактивно' },
              ]} />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </div>
  )
}
