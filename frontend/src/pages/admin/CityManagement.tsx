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
  Table,
} from '@/components/design/system'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@/components/design/icons'
import type { ColumnsType } from '@/components/design/types'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetCities,
  getGetCitiesQueryKey,
} from '@/api/generated/cities/cities'
import {
  usePostAdminCities,
  usePutAdminCitiesId,
  useDeleteAdminCitiesId,
} from '@/api/generated/admin-cities/admin-cities'
import type { InternalHandlerCityResponse } from '@/api/generated/model'
import PageHeader from '@/components/PageHeader'

interface CityFormValues {
  name: string
  slug: string
  latitude: number
  longitude: number
}

export default function CityManagement() {
  const { modal, message } = App.useApp()
  const queryClient = useQueryClient()
  const [form] = Form.useForm<CityFormValues>()

  const [formModalOpen, setFormModalOpen] = useState(false)
  const [editingCity, setEditingCity] = useState<InternalHandlerCityResponse | null>(null)

  const { data, isLoading } = useGetCities()
  const createMutation = usePostAdminCities()
  const updateMutation = usePutAdminCitiesId()
  const deleteMutation = useDeleteAdminCitiesId()

  const cities: InternalHandlerCityResponse[] = data?.data ?? []

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetCitiesQueryKey() })
  }

  const openCreateModal = () => {
    setEditingCity(null)
    form.resetFields()
    setFormModalOpen(true)
  }

  const openEditModal = (city: InternalHandlerCityResponse) => {
    setEditingCity(city)
    form.setFieldsValue({
      name: city.name ?? '',
      slug: city.slug ?? '',
      latitude: city.latitude ?? 0,
      longitude: city.longitude ?? 0,
    })
    setFormModalOpen(true)
  }

  const handleFormSubmit = async () => {
    const values = await form.validateFields()
    if (editingCity) {
      await updateMutation.mutateAsync({
        id: editingCity.id!,
        data: values,
      })
      message.success('Город обновлён')
    } else {
      await createMutation.mutateAsync({ data: values })
      message.success('Город создан')
    }
    setFormModalOpen(false)
    invalidate()
  }

  const handleDelete = (city: InternalHandlerCityResponse) => {
    modal.confirm({
      title: 'Удалить город?',
      content: `Город "${city.name}" будет удалён. Это действие необратимо.`,
      okText: 'Удалить',
      okType: 'danger',
      cancelText: 'Отмена',
      onOk: () =>
        deleteMutation.mutateAsync({ id: city.id! }).then(() => {
          message.success('Город удалён')
          invalidate()
        }),
    })
  }

  const columns: ColumnsType<InternalHandlerCityResponse> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Slug',
      dataIndex: 'slug',
      key: 'slug',
    },
    {
      title: 'Широта',
      dataIndex: 'latitude',
      key: 'latitude',
      width: 120,
      render: (val: number) => val?.toFixed(4) ?? '-',
    },
    {
      title: 'Долгота',
      dataIndex: 'longitude',
      key: 'longitude',
      width: 120,
      render: (val: number) => val?.toFixed(4) ?? '-',
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 160,
      render: (_: unknown, record: InternalHandlerCityResponse) => (
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
        title="Управление городами"
        description="Города, координаты и slug теперь собраны в чистый рабочий каталог."
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
            Добавить город
          </Button>
        }
      />

      <Card className="rh-admin-reference-card" title="Каталог городов">
        {isLoading ? (
          <div className="rh-admin-state-card">
            <Spin size="large" />
            <span>Загружаем города</span>
          </div>
        ) : cities.length === 0 ? (
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Нет городов</div>
            <p className="rh-admin-empty-state__text">
              Добавьте первый город, чтобы связать каталог, фильтры и географические сценарии.
            </p>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
              Добавить город
            </Button>
          </div>
        ) : (
          <Table
            dataSource={cities}
            columns={columns}
            rowKey="id"
            pagination={false}
            size="middle"
            locale={{ emptyText: 'Нет городов' }}
          />
        )}
      </Card>

      <Modal
        title={editingCity ? 'Редактировать город' : 'Добавить город'}
        open={formModalOpen}
        onCancel={() => setFormModalOpen(false)}
        onOk={handleFormSubmit}
        okText={editingCity ? 'Сохранить' : 'Создать'}
        cancelText="Отмена"
        confirmLoading={createMutation.isPending || updateMutation.isPending}
      >
        <Form form={form} layout="vertical" className="rh-admin-modal-form">
          <Form.Item
            name="name"
            label="Название"
            rules={[{ required: true, message: 'Введите название города' }]}
          >
            <Input placeholder="Москва" />
          </Form.Item>
          <Form.Item
            name="slug"
            label="Slug"
            rules={[{ required: true, message: 'Введите slug' }]}
          >
            <Input placeholder="moscow" />
          </Form.Item>
          <Form.Item name="latitude" label="Широта">
            <InputNumber
              className="rh-admin-form-control"
              placeholder="55.7558"
              step={0.0001}
              min={-90}
              max={90}
            />
          </Form.Item>
          <Form.Item name="longitude" label="Долгота">
            <InputNumber
              className="rh-admin-form-control"
              placeholder="37.6173"
              step={0.0001}
              min={-180}
              max={180}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
