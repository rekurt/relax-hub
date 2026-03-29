import { useState, useCallback } from 'react'
import {
  App,
  Button,
  Card,
  DatePicker,
  Empty,
  Form,
  Input,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  Typography,
} from 'antd'
import { SearchOutlined, CloseCircleOutlined, SwapOutlined } from '@ant-design/icons'
import { axiosInstance } from '@/api/axios-instance'
import { formatPrice } from '@/lib/format'
import dayjs from 'dayjs'
import type { ColumnsType } from 'antd/es/table'

const { Title } = Typography
const { RangePicker } = DatePicker

interface BookingData {
  id: string
  user_id: string
  bathhouse_id: string
  start_time: string
  end_time: string
  guest_count: number
  total_price: number
  status: string
  comment: string
  created_at: string
}

interface PaginatedResponse {
  success: boolean
  data: BookingData[]
  meta: {
    page: number
    page_size: number
    total_count: number
    total_pages: number
  }
}

const STATUS_OPTIONS = [
  { value: 'pending', label: 'Ожидание' },
  { value: 'pending_owner', label: 'Ожидание владельца' },
  { value: 'confirmed', label: 'Подтверждено' },
  { value: 'cancelled', label: 'Отменено' },
  { value: 'rejected', label: 'Отклонено' },
  { value: 'completed', label: 'Завершено' },
  { value: 'no_show', label: 'Неявка' },
  { value: 'force_majeure_cancelled', label: 'Форс-мажор' },
]

const STATUS_TAGS: Record<string, { color: string; text: string }> = {
  pending: { color: 'blue', text: 'Ожидание' },
  pending_owner: { color: 'orange', text: 'Ожидание владельца' },
  confirmed: { color: 'green', text: 'Подтверждено' },
  cancelled: { color: 'red', text: 'Отменено' },
  rejected: { color: 'volcano', text: 'Отклонено' },
  completed: { color: 'cyan', text: 'Завершено' },
  no_show: { color: 'default', text: 'Неявка' },
  force_majeure_cancelled: { color: 'purple', text: 'Форс-мажор' },
}

type ActionType = 'cancel' | 'change-status'

export default function BookingManagement() {
  const { message } = App.useApp()
  const [bookings, setBookings] = useState<BookingData[]>([])
  const [loading, setLoading] = useState(false)
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 })
  const [filterUserId, setFilterUserId] = useState('')
  const [filterBathhouseId, setFilterBathhouseId] = useState('')
  const [filterStatus, setFilterStatus] = useState<string | undefined>()
  const [filterDates, setFilterDates] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null)
  const [actionModal, setActionModal] = useState<{ type: ActionType; bookingId: string } | null>(null)
  const [form] = Form.useForm()

  const fetchBookings = useCallback(async (page = 1, pageSize = 20) => {
    setLoading(true)
    try {
      const params: Record<string, string> = {
        page: String(page),
        page_size: String(pageSize),
      }
      if (filterUserId) params.user_id = filterUserId
      if (filterBathhouseId) params.bathhouse_id = filterBathhouseId
      if (filterStatus) params.status = filterStatus
      if (filterDates) {
        params.from_date = filterDates[0].toISOString()
        params.to_date = filterDates[1].toISOString()
      }

      const { data } = await axiosInstance.get<PaginatedResponse>('/api/v1/admin/bookings', { params })
      setBookings(data.data || [])
      setPagination({
        current: data.meta.page,
        pageSize: data.meta.page_size,
        total: data.meta.total_count,
      })
    } catch {
      message.error('Ошибка загрузки бронирований')
    } finally {
      setLoading(false)
    }
  }, [filterUserId, filterBathhouseId, filterStatus, filterDates, message])

  const handleAction = async () => {
    if (!actionModal) return
    try {
      const values = await form.validateFields()
      if (actionModal.type === 'cancel') {
        await axiosInstance.post(`/api/v1/admin/bookings/${actionModal.bookingId}/cancel`, {
          reason: values.reason,
        })
        message.success('Бронирование отменено')
      } else {
        await axiosInstance.post(`/api/v1/admin/bookings/${actionModal.bookingId}/change-status`, {
          status: values.status,
          reason: values.reason,
        })
        message.success('Статус изменён')
      }
      setActionModal(null)
      form.resetFields()
      fetchBookings(pagination.current, pagination.pageSize)
    } catch {
      message.error('Ошибка выполнения операции')
    }
  }

  const columns: ColumnsType<BookingData> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 100,
      render: (id: string) => id.substring(0, 8) + '...',
    },
    {
      title: 'Клиент',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 100,
      render: (id: string) => id.substring(0, 8) + '...',
    },
    {
      title: 'Объект',
      dataIndex: 'bathhouse_id',
      key: 'bathhouse_id',
      width: 100,
      render: (id: string) => id.substring(0, 8) + '...',
    },
    {
      title: 'Дата начала',
      dataIndex: 'start_time',
      key: 'start_time',
      render: (t: string) => dayjs(t).format('DD.MM.YYYY HH:mm'),
    },
    {
      title: 'Гости',
      dataIndex: 'guest_count',
      key: 'guest_count',
      width: 80,
    },
    {
      title: 'Сумма',
      dataIndex: 'total_price',
      key: 'total_price',
      render: (p: number) => formatPrice(p),
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (s: string) => {
        const tag = STATUS_TAGS[s] || { color: 'default', text: s }
        return <Tag color={tag.color}>{tag.text}</Tag>
      },
    },
    {
      title: 'Создано',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (t: string) => dayjs(t).format('DD.MM.YYYY HH:mm'),
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_: unknown, record: BookingData) => (
        <Space size="small">
          <Button
            size="small"
            danger
            icon={<CloseCircleOutlined />}
            onClick={() => {
              setActionModal({ type: 'cancel', bookingId: record.id })
            }}
            disabled={record.status === 'cancelled' || record.status === 'rejected' || record.status === 'force_majeure_cancelled'}
          >
            Отменить
          </Button>
          <Button
            size="small"
            icon={<SwapOutlined />}
            onClick={() => {
              setActionModal({ type: 'change-status', bookingId: record.id })
            }}
          >
            Статус
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <Title level={3}>Управление бронированиями</Title>

      <Card style={{ marginBottom: 16 }}>
        <Space wrap>
          <Input
            placeholder="User ID"
            value={filterUserId}
            onChange={(e) => setFilterUserId(e.target.value)}
            style={{ width: 280 }}
            allowClear
          />
          <Input
            placeholder="Bathhouse ID"
            value={filterBathhouseId}
            onChange={(e) => setFilterBathhouseId(e.target.value)}
            style={{ width: 280 }}
            allowClear
          />
          <Select
            placeholder="Статус"
            value={filterStatus}
            onChange={setFilterStatus}
            options={STATUS_OPTIONS}
            style={{ width: 200 }}
            allowClear
          />
          <RangePicker
            value={filterDates}
            onChange={(dates) => setFilterDates(dates as [dayjs.Dayjs, dayjs.Dayjs] | null)}
          />
          <Button
            type="primary"
            icon={<SearchOutlined />}
            onClick={() => fetchBookings(1, pagination.pageSize)}
          >
            Поиск
          </Button>
        </Space>
      </Card>

      <Table
        columns={columns}
        dataSource={bookings}
        rowKey="id"
        loading={loading}
        locale={{ emptyText: <Empty description="Нет бронирований. Используйте фильтры и нажмите «Поиск» для загрузки данных" /> }}
        pagination={{
          ...pagination,
          showSizeChanger: true,
          showTotal: (total) => `Всего: ${total}`,
          onChange: (page, pageSize) => fetchBookings(page, pageSize),
        }}
      />

      <Modal
        title={actionModal?.type === 'cancel' ? 'Отмена бронирования' : 'Изменение статуса'}
        open={actionModal !== null}
        onOk={handleAction}
        onCancel={() => {
          setActionModal(null)
          form.resetFields()
        }}
        okText="Подтвердить"
        cancelText="Отмена"
      >
        <Form form={form} layout="vertical">
          {actionModal?.type === 'change-status' && (
            <Form.Item
              name="status"
              label="Новый статус"
              rules={[{ required: true, message: 'Выберите статус' }]}
            >
              <Select options={STATUS_OPTIONS} />
            </Form.Item>
          )}
          <Form.Item
            name="reason"
            label="Причина"
            rules={[{ required: true, message: 'Укажите причину' }]}
          >
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
