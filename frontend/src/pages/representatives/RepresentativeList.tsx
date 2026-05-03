import { useState } from 'react'
import {
  App,
  Button,
  Empty,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Table,
  Tag,
  Typography,
} from 'antd'
import {
  DeleteOutlined,
  UserAddOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import {
  useGetBathhousesIdRepresentatives,
  usePostBathhousesIdRepresentatives,
  useDeleteRepresentativesId,
} from '@/api/generated/representatives/representatives'
import type { InternalHandlerRepresentativeResponse } from '@/api/generated/model'
import { useBathhouseStore } from '@/stores/bathhouse'
import { useAuthStore } from '@/stores/auth'
import { useQueryClient } from '@tanstack/react-query'

const { Title } = Typography

interface InviteFormValues {
  user_email: string
  role: string
}

const roleLabels: Record<string, string> = {
  manager: 'Менеджер',
  observer: 'Наблюдатель',
}

export default function RepresentativeList() {
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)
  const user = useAuthStore((s) => s.user)
  const queryClient = useQueryClient()
  const { message } = App.useApp()
  const [form] = Form.useForm<InviteFormValues>()

  const [modalOpen, setModalOpen] = useState(false)

  const isOwner = user?.role === 'owner'

  const { data, isLoading } = useGetBathhousesIdRepresentatives(
    selectedBathhouseId ?? '',
    { query: { enabled: !!selectedBathhouseId } },
  )

  const representatives = data?.data ?? []

  const invalidateRepresentatives = () => {
    queryClient.invalidateQueries({
      queryKey: [`/bathhouses/${selectedBathhouseId}/representatives`],
    })
  }

  const inviteMutation = usePostBathhousesIdRepresentatives({
    mutation: {
      onSuccess: () => {
        message.success('Представитель приглашён')
        setModalOpen(false)
        form.resetFields()
        invalidateRepresentatives()
      },
      onError: () => message.error('Не удалось пригласить представителя'),
    },
  })

  const deleteMutation = useDeleteRepresentativesId({
    mutation: {
      onSuccess: () => {
        message.success('Представитель удалён')
        invalidateRepresentatives()
      },
      onError: () => message.error('Не удалось удалить представителя'),
    },
  })

  const handleInvite = (values: InviteFormValues) => {
    if (!selectedBathhouseId) return
    inviteMutation.mutate({
      id: selectedBathhouseId,
      data: { user_email: values.user_email, role: values.role },
    })
  }

  const handleDelete = (repId: string) => {
    deleteMutation.mutate({ id: repId })
  }

  const columns = [
    {
      title: 'ID представителя',
      dataIndex: 'user_id',
      key: 'user_id',
      ellipsis: true,
      render: (userId: string) => (
        <Tag style={{ fontFamily: 'monospace' }}>{userId?.slice(0, 8)}...</Tag>
      ),
    },
    {
      title: 'Роль',
      dataIndex: 'role',
      key: 'role',
      width: 140,
      render: (role: string) => (
        <Tag color={role === 'manager' ? 'blue' : 'default'}>
          {roleLabels[role] ?? role}
        </Tag>
      ),
    },
    {
      title: 'Дата добавления',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => date ? dayjs(date).format('DD.MM.YYYY HH:mm') : '—',
    },
    ...(isOwner
      ? [
          {
            title: 'Действия',
            key: 'actions',
            width: 100,
            render: (_: unknown, record: InternalHandlerRepresentativeResponse) => (
              <Popconfirm
                title="Удалить представителя?"
                description="Доступ к управлению баней будет отозван"
                onConfirm={() => record.id && handleDelete(record.id)}
                okText="Удалить"
                cancelText="Отмена"
              >
                <Button type="text" danger icon={<DeleteOutlined />} size="small" />
              </Popconfirm>
            ),
          },
        ]
      : []),
  ]

  if (!selectedBathhouseId) {
    return (
      <div>
        <Title level={3}>Представители</Title>
        <div style={{ textAlign: 'center', padding: 40, color: 'var(--rh-text-muted)' }}>
          Выберите баню для управления представителями
        </div>
      </div>
    )
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>Представители</Title>
        {isOwner && (
          <Button
            type="primary"
            icon={<UserAddOutlined />}
            onClick={() => {
              form.resetFields()
              setModalOpen(true)
            }}
          >
            Пригласить
          </Button>
        )}
      </div>

      <Table
        dataSource={representatives}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        locale={{
          emptyText: (
            <Empty description="Нет представителей. Пригласите сотрудника для управления бронированиями и общением с клиентами." />
          ),
        }}
        pagination={false}
      />

      <Modal
        title="Пригласить представителя"
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        footer={null}
        destroyOnHidden
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleInvite}
        >
          <Form.Item
            name="user_email"
            label="Email пользователя"
            rules={[
              { required: true, message: 'Укажите email' },
              { type: 'email', message: 'Введите корректный email' },
            ]}
            extra="Пользователь должен быть зарегистрирован в системе"
          >
            <Input placeholder="user@example.com" />
          </Form.Item>

          <Form.Item
            name="role"
            label="Роль"
            initialValue="manager"
            rules={[{ required: true, message: 'Выберите роль' }]}
            extra="Менеджер — полный доступ, Наблюдатель — только просмотр"
          >
            <Select>
              <Select.Option value="manager">Менеджер</Select.Option>
              <Select.Option value="observer">Наблюдатель</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={inviteMutation.isPending}
              style={{ marginRight: 8 }}
            >
              Пригласить
            </Button>
            <Button onClick={() => setModalOpen(false)}>
              Отмена
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
