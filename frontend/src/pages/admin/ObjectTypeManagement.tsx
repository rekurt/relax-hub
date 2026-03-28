import { useState } from 'react'
import {
  App,
  Button,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Space,
  Spin,
  Switch,
  Table,
  Tag,
  Typography,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { axiosInstance } from '@/api/axios-instance'

const { Title } = Typography
const { TextArea } = Input

interface ObjectType {
  id: string
  name: string
  description: string
  sort_order: number
  is_active: boolean
  created_at: string
  updated_at: string
}

interface ObjectTypeFormValues {
  name: string
  description: string
  sort_order: number
  is_active: boolean
}

const QUERY_KEY = ['admin', 'object-types']

function useObjectTypes() {
  return useQuery({
    queryKey: QUERY_KEY,
    queryFn: async () => {
      const { data } = await axiosInstance.get<{ data: ObjectType[] }>('/admin/object-types')
      return data.data ?? []
    },
  })
}

export default function ObjectTypeManagement() {
  const { modal, message } = App.useApp()
  const queryClient = useQueryClient()
  const [form] = Form.useForm<ObjectTypeFormValues>()

  const [formModalOpen, setFormModalOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<ObjectType | null>(null)

  const { data: objectTypes = [], isLoading } = useObjectTypes()

  const createMutation = useMutation({
    mutationFn: (values: ObjectTypeFormValues) =>
      axiosInstance.post('/admin/object-types', values),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY })
      message.success('Тип объекта создан')
      setFormModalOpen(false)
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, values }: { id: string; values: ObjectTypeFormValues }) =>
      axiosInstance.put(`/admin/object-types/${id}`, values),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY })
      message.success('Тип объекта обновлён')
      setFormModalOpen(false)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => axiosInstance.delete(`/admin/object-types/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY })
      message.success('Тип объекта удалён')
    },
  })

  const openCreateModal = () => {
    setEditingItem(null)
    form.resetFields()
    form.setFieldsValue({ is_active: true, sort_order: 0 })
    setFormModalOpen(true)
  }

  const openEditModal = (item: ObjectType) => {
    setEditingItem(item)
    form.setFieldsValue({
      name: item.name,
      description: item.description,
      sort_order: item.sort_order,
      is_active: item.is_active,
    })
    setFormModalOpen(true)
  }

  const handleFormSubmit = async () => {
    const values = await form.validateFields()
    if (editingItem) {
      await updateMutation.mutateAsync({ id: editingItem.id, values })
    } else {
      await createMutation.mutateAsync(values)
    }
  }

  const handleDelete = (item: ObjectType) => {
    modal.confirm({
      title: 'Удалить тип объекта?',
      content: `"${item.name}" будет удалён. Это действие необратимо.`,
      okText: 'Удалить',
      okType: 'danger',
      cancelText: 'Отмена',
      onOk: () => deleteMutation.mutateAsync(item.id),
    })
  }

  const columns: ColumnsType<ObjectType> = [
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Описание',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (val: string) => val || '-',
    },
    {
      title: 'Порядок',
      dataIndex: 'sort_order',
      key: 'sort_order',
      width: 100,
    },
    {
      title: 'Статус',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 100,
      render: (val: boolean) =>
        val ? <Tag color="green">Активно</Tag> : <Tag color="red">Неактивно</Tag>,
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 180,
      render: (_: unknown, record: ObjectType) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => openEditModal(record)}
          >
            Изменить
          </Button>
          <Button
            type="link"
            size="small"
            icon={<DeleteOutlined />}
            danger
            onClick={() => handleDelete(record)}
          >
            Удалить
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>
          Типы объектов
        </Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
          Добавить тип
        </Button>
      </div>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: 48 }}>
          <Spin size="large" />
        </div>
      ) : objectTypes.length === 0 ? (
        <Empty description="Нет типов объектов" />
      ) : (
        <Table
          dataSource={objectTypes}
          columns={columns}
          rowKey="id"
          pagination={false}
          size="middle"
          locale={{ emptyText: 'Нет типов объектов' }}
        />
      )}

      <Modal
        title={editingItem ? 'Редактировать тип объекта' : 'Добавить тип объекта'}
        open={formModalOpen}
        onCancel={() => setFormModalOpen(false)}
        onOk={handleFormSubmit}
        okText={editingItem ? 'Сохранить' : 'Создать'}
        cancelText="Отмена"
        confirmLoading={createMutation.isPending || updateMutation.isPending}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="name"
            label="Название"
            rules={[{ required: true, message: 'Введите название типа' }]}
          >
            <Input placeholder="Русская баня" />
          </Form.Item>
          <Form.Item name="description" label="Описание">
            <TextArea rows={3} placeholder="Традиционная русская баня с парной" />
          </Form.Item>
          <Form.Item name="sort_order" label="Порядок сортировки">
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item name="is_active" label="Активно" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
