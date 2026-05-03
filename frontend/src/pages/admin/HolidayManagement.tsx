import { useState } from 'react'
import {
  App,
  Button,
  DatePicker,
  Empty,
  Form,
  Input,
  Modal,
  Select,
  Space,
  Spin,
  Switch,
  Table,
  Tag,
  Typography,
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

const { Title } = Typography

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
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>
          Праздничные дни
        </Title>
        <Space>
          <Select
            allowClear
            placeholder="Регион"
            style={{ width: 150 }}
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
      </div>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: 48 }}>
          <Spin size="large" />
        </div>
      ) : holidays.length === 0 ? (
        <Empty description="Нет праздников" />
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

      <Modal
        title={editingItem ? 'Редактировать праздник' : 'Добавить праздник'}
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
            rules={[{ required: true, message: 'Введите название праздника' }]}
          >
            <Input placeholder="Новый год" />
          </Form.Item>
          <Form.Item
            name="date"
            label="Дата"
            rules={[{ required: true, message: 'Выберите дату' }]}
          >
            <DatePicker style={{ width: '100%' }} format="DD.MM.YYYY" />
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
