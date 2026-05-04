import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Space,
  Spin,
  Switch,
  Table,
  Tag,
} from '@/components/design/system'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@/components/design/icons'
import type { ColumnsType } from '@/components/design/types'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { axiosInstance } from '@/api/axios-instance'
import PageHeader from '@/components/PageHeader'

interface Amenity {
  id: string
  name: string
  icon: string
  sort_order: number
  is_active: boolean
  created_at: string
  updated_at: string
}

interface AmenityFormValues {
  name: string
  icon: string
  sort_order: number
  is_active: boolean
}

const QUERY_KEY = ['admin', 'amenities']

function useAmenities() {
  return useQuery({
    queryKey: QUERY_KEY,
    queryFn: async () => {
      const { data } = await axiosInstance.get<{ data: Amenity[] }>('/admin/amenities')
      return data.data ?? []
    },
  })
}

export default function AmenityManagement() {
  const { modal, message } = App.useApp()
  const queryClient = useQueryClient()
  const [form] = Form.useForm<AmenityFormValues>()

  const [formModalOpen, setFormModalOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<Amenity | null>(null)

  const { data: amenities = [], isLoading } = useAmenities()

  const createMutation = useMutation({
    mutationFn: (values: AmenityFormValues) =>
      axiosInstance.post('/admin/amenities', values),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY })
      message.success('Удобство создано')
      setFormModalOpen(false)
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, values }: { id: string; values: AmenityFormValues }) =>
      axiosInstance.put(`/admin/amenities/${id}`, values),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY })
      message.success('Удобство обновлено')
      setFormModalOpen(false)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => axiosInstance.delete(`/admin/amenities/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY })
      message.success('Удобство удалено')
    },
  })

  const openCreateModal = () => {
    setEditingItem(null)
    form.resetFields()
    form.setFieldsValue({ is_active: true, sort_order: 0 })
    setFormModalOpen(true)
  }

  const openEditModal = (item: Amenity) => {
    setEditingItem(item)
    form.setFieldsValue({
      name: item.name,
      icon: item.icon,
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

  const handleDelete = (item: Amenity) => {
    modal.confirm({
      title: 'Удалить удобство?',
      content: `"${item.name}" будет удалено. Это действие необратимо.`,
      okText: 'Удалить',
      okType: 'danger',
      cancelText: 'Отмена',
      onOk: () => deleteMutation.mutateAsync(item.id),
    })
  }

  const columns: ColumnsType<Amenity> = [
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Иконка',
      dataIndex: 'icon',
      key: 'icon',
      width: 120,
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
      render: (_: unknown, record: Amenity) => (
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
    <div className="rh-stack rh-admin-reference-page">
      <PageHeader
        eyebrow="Справочник"
        title="Управление удобствами"
        description="Единый каталог удобств с иконками, порядком сортировки и статусами доступности."
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
            Добавить удобство
          </Button>
        }
      />

      <Card className="rh-admin-reference-card" title="Каталог удобств">
        {isLoading ? (
          <div className="rh-admin-state-card">
            <Spin size="large" />
            <span>Загружаем удобства</span>
          </div>
        ) : amenities.length === 0 ? (
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Нет удобств</div>
            <p className="rh-admin-empty-state__text">
              Добавьте первое удобство, чтобы владельцы могли точнее описывать объекты.
            </p>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
              Добавить удобство
            </Button>
          </div>
        ) : (
          <Table
            dataSource={amenities}
            columns={columns}
            rowKey="id"
            pagination={false}
            size="middle"
            locale={{ emptyText: 'Нет удобств' }}
          />
        )}
      </Card>

      <Modal
        title={editingItem ? 'Редактировать удобство' : 'Добавить удобство'}
        open={formModalOpen}
        onCancel={() => setFormModalOpen(false)}
        onOk={handleFormSubmit}
        okText={editingItem ? 'Сохранить' : 'Создать'}
        cancelText="Отмена"
        confirmLoading={createMutation.isPending || updateMutation.isPending}
      >
        <Form form={form} layout="vertical" className="rh-admin-modal-form">
          <Form.Item
            name="name"
            label="Название"
            rules={[{ required: true, message: 'Введите название удобства' }]}
          >
            <Input placeholder="Бассейн" />
          </Form.Item>
          <Form.Item
            name="icon"
            label="Иконка (идентификатор)"
            tooltip="Например: pool, sauna, bbq"
          >
            <Input placeholder="pool" />
          </Form.Item>
          <Form.Item name="sort_order" label="Порядок сортировки">
            <InputNumber className="rh-admin-form-control" min={0} />
          </Form.Item>
          <Form.Item name="is_active" label="Активно" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
