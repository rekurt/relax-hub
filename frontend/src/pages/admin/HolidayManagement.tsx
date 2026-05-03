import { useState } from 'react'
import {
  App,
  Button,
  Card,
  DatePicker,
  Form,
  Input,
  Modal,
  Select,
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
import dayjs from 'dayjs'
import PageHeader from '@/components/PageHeader'

interface Holiday {
  id: string
  name: string
  date: string
  region: string
  is_recurring: boolean
  created_at: string
  updated_at: string
}

interface HolidayFormValues {
  name: string
  date: dayjs.Dayjs
  region: string
  is_recurring: boolean
}

const QUERY_KEY = ['admin', 'holidays']

const regionLabels: Record<string, string> = {
  RU: 'Россия',
  BY: 'Беларусь',
}

function useHolidays(region?: string) {
  return useQuery({
    queryKey: [...QUERY_KEY, region],
    queryFn: async () => {
      const params = region ? { region } : {}
      const { data } = await axiosInstance.get<{ data: Holiday[] }>('/admin/holidays', { params })
      return data.data ?? []
    },
  })
}

export default function HolidayManagement() {
  const { modal, message } = App.useApp()
  const queryClient = useQueryClient()
  const [form] = Form.useForm<HolidayFormValues>()

  const [formModalOpen, setFormModalOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<Holiday | null>(null)
  const [regionFilter, setRegionFilter] = useState<string | undefined>(undefined)

  const { data: holidays = [], isLoading } = useHolidays(regionFilter)

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: QUERY_KEY })
  }

  const createMutation = useMutation({
    mutationFn: (values: { name: string; date: string; region: string; is_recurring: boolean }) =>
      axiosInstance.post('/admin/holidays', values),
    onSuccess: () => {
      invalidate()
      message.success('Праздник создан')
      setFormModalOpen(false)
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, values }: { id: string; values: { name: string; date: string; region: string; is_recurring: boolean } }) =>
      axiosInstance.put(`/admin/holidays/${id}`, values),
    onSuccess: () => {
      invalidate()
      message.success('Праздник обновлён')
      setFormModalOpen(false)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => axiosInstance.delete(`/admin/holidays/${id}`),
    onSuccess: () => {
      invalidate()
      message.success('Праздник удалён')
    },
  })

  const openCreateModal = () => {
    setEditingItem(null)
    form.resetFields()
    form.setFieldsValue({ region: 'RU', is_recurring: false })
    setFormModalOpen(true)
  }

  const openEditModal = (item: Holiday) => {
    setEditingItem(item)
    form.setFieldsValue({
      name: item.name,
      date: dayjs(item.date),
      region: item.region,
      is_recurring: item.is_recurring,
    })
    setFormModalOpen(true)
  }

  const handleFormSubmit = async () => {
    const values = await form.validateFields()
    const payload = {
      name: values.name,
      date: values.date.format('YYYY-MM-DD'),
      region: values.region,
      is_recurring: values.is_recurring,
    }
    if (editingItem) {
      await updateMutation.mutateAsync({ id: editingItem.id, values: payload })
    } else {
      await createMutation.mutateAsync(payload)
    }
  }

  const handleDelete = (item: Holiday) => {
    modal.confirm({
      title: 'Удалить праздник?',
      content: `"${item.name}" будет удалён. Это действие необратимо.`,
      okText: 'Удалить',
      okType: 'danger',
      cancelText: 'Отмена',
      onOk: () => deleteMutation.mutateAsync(item.id),
    })
  }

  const columns: ColumnsType<Holiday> = [
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Дата',
      dataIndex: 'date',
      key: 'date',
      width: 130,
      render: (val: string) => dayjs(val).format('DD.MM.YYYY'),
    },
    {
      title: 'Регион',
      dataIndex: 'region',
      key: 'region',
      width: 120,
      render: (val: string) => regionLabels[val] ?? val,
    },
    {
      title: 'Повторяется',
      dataIndex: 'is_recurring',
      key: 'is_recurring',
      width: 130,
      render: (val: boolean) =>
        val ? <Tag color="blue">Ежегодно</Tag> : <Tag>Разово</Tag>,
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 180,
      render: (_: unknown, record: Holiday) => (
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
        eyebrow="Календарь"
        title="Праздничные дни"
        description="Региональные праздничные дни и ежегодные исключения для расписаний платформы."
        extra={
          <Space wrap>
          <Select
            className="rh-admin-filter-select"
            allowClear
            placeholder="Регион"
            value={regionFilter}
            onChange={setRegionFilter}
            options={[
              { value: 'RU', label: 'Россия' },
              { value: 'BY', label: 'Беларусь' },
            ]}
          />
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
            Добавить праздник
          </Button>
          </Space>
        }
      />

      <Card className="rh-admin-reference-card" title="Календарь праздников">
        {isLoading ? (
          <div className="rh-admin-state-card">
            <Spin size="large" />
            <span>Загружаем праздники</span>
          </div>
        ) : holidays.length === 0 ? (
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Нет праздников</div>
            <p className="rh-admin-empty-state__text">
              Добавьте праздничный день, чтобы учитывать региональные исключения в календарях.
            </p>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
              Добавить праздник
            </Button>
          </div>
        ) : (
          <Table
            dataSource={holidays}
            columns={columns}
            rowKey="id"
            pagination={false}
            size="middle"
            locale={{ emptyText: 'Нет праздников' }}
          />
        )}
      </Card>

      <Modal
        title={editingItem ? 'Редактировать праздник' : 'Добавить праздник'}
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
            rules={[{ required: true, message: 'Введите название праздника' }]}
          >
            <Input placeholder="Новый год" />
          </Form.Item>
          <Form.Item
            name="date"
            label="Дата"
            rules={[{ required: true, message: 'Выберите дату' }]}
          >
            <DatePicker className="rh-admin-form-control" format="DD.MM.YYYY" />
          </Form.Item>
          <Form.Item
            name="region"
            label="Регион"
            rules={[{ required: true, message: 'Выберите регион' }]}
          >
            <Select
              options={[
                { value: 'RU', label: 'Россия' },
                { value: 'BY', label: 'Беларусь' },
              ]}
            />
          </Form.Item>
          <Form.Item name="is_recurring" label="Повторяется ежегодно" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
