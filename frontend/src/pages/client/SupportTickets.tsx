import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Empty,
  Form,
  Input,
  Modal,
  Pagination,
  Select,
  Segmented,
  Space,
  Spin,
  Table,
  Tag,
  Typography,
} from 'antd'
import { PlusOutlined, RobotOutlined } from '@ant-design/icons'
import SupportChatBot from '@/components/SupportChatBot'
import type { ColumnsType } from 'antd/es/table'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetMyTickets,
  getGetMyTicketsQueryKey,
  usePostMyTickets,
} from '@/api/generated/support/support'
import type { InternalHandlerTicketResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'

const { Title, Text } = Typography

const STATUS_OPTIONS = [
  { label: 'Все', value: '' },
  { label: 'Открытые', value: 'open' },
  { label: 'В работе', value: 'in_progress' },
  { label: 'Эскалированные', value: 'escalated' },
  { label: 'Решённые', value: 'resolved' },
  { label: 'Закрытые', value: 'closed' },
]

const CATEGORY_OPTIONS = [
  { label: 'Вопрос', value: 'question' },
  { label: 'Проблема', value: 'problem' },
  { label: 'Жалоба', value: 'complaint' },
  { label: 'Запрос возврата', value: 'refund_request' },
  { label: 'Проблема с аккаунтом', value: 'account_issue' },
]

const statusLabel: Record<string, string> = {
  open: 'Открыт',
  in_progress: 'В работе',
  escalated: 'Эскалирован',
  resolved: 'Решён',
  closed: 'Закрыт',
}

const statusColor: Record<string, string> = {
  open: 'blue',
  in_progress: 'processing',
  escalated: 'orange',
  resolved: 'green',
  closed: 'default',
}

const categoryLabel: Record<string, string> = {
  question: 'Вопрос',
  problem: 'Проблема',
  complaint: 'Жалоба',
  refund_request: 'Запрос возврата',
  account_issue: 'Проблема с аккаунтом',
}

const priorityLabel: Record<string, string> = {
  low: 'Низкий',
  medium: 'Средний',
  high: 'Высокий',
  critical: 'Критический',
}

const priorityColor: Record<string, string> = {
  low: 'default',
  medium: 'blue',
  high: 'orange',
  critical: 'red',
}

export default function SupportTickets() {
  const { message } = App.useApp()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [statusFilter, setStatusFilter] = useState('')
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [faqBotOpen, setFaqBotOpen] = useState(false)
  const [form] = Form.useForm()

  const { data, isLoading } = useGetMyTickets({
    page,
    page_size: pageSize,
  })

  const createMutation = usePostMyTickets()

  const allTickets: InternalHandlerTicketResponse[] = data?.data ?? []
  const tickets = statusFilter
    ? allTickets.filter((t) => t.status === statusFilter)
    : allTickets
  const meta = data?.meta

  const handleCreate = async () => {
    try {
      const values = await form.validateFields()
      await createMutation.mutateAsync({
        data: {
          category: values.category,
          subject: values.subject,
          message: values.message,
          booking_id: values.booking_id || undefined,
        },
      })
      message.success('Обращение создано')
      setCreateModalOpen(false)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: getGetMyTicketsQueryKey() })
    } catch {
      if (createMutation.isError) {
        message.error('Не удалось создать обращение')
      }
    }
  }

  const columns: ColumnsType<InternalHandlerTicketResponse> = [
    {
      title: 'Тема',
      dataIndex: 'subject',
      key: 'subject',
      ellipsis: true,
      render: (text: string, record: InternalHandlerTicketResponse) => (
        <a onClick={() => navigate(`/client/tickets/${record.id}`)}>
          {text || <Text type="secondary">Без темы</Text>}
        </a>
      ),
    },
    {
      title: 'Категория',
      dataIndex: 'category',
      key: 'category',
      width: 160,
      render: (cat: string) => categoryLabel[cat] ?? cat,
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      width: 140,
      render: (status: string) => (
        <Tag color={statusColor[status] ?? 'default'}>
          {statusLabel[status] ?? status}
        </Tag>
      ),
    },
    {
      title: 'Приоритет',
      dataIndex: 'priority',
      key: 'priority',
      width: 120,
      render: (priority: string) => (
        <Tag color={priorityColor[priority] ?? 'default'}>
          {priorityLabel[priority] ?? priority}
        </Tag>
      ),
    },
    {
      title: 'Дата',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (date: string) => (date ? formatDateTime(date) : '-'),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>
          Мои обращения
        </Title>
        <Space>
          <Button
            icon={<RobotOutlined />}
            onClick={() => setFaqBotOpen(true)}
          >
            Быстрая помощь
          </Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalOpen(true)}
          >
            Новое обращение
          </Button>
        </Space>
      </div>

      <Space direction="vertical" size={16} style={{ width: '100%', marginBottom: 16 }}>
        <Segmented
          options={STATUS_OPTIONS}
          value={statusFilter}
          onChange={(val) => {
            setStatusFilter(val as string)
            setPage(1)
          }}
        />
      </Space>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: 48 }}>
          <Spin size="large" />
        </div>
      ) : tickets.length === 0 ? (
        <Card>
          <Empty description="Нет обращений">
            <Button type="primary" onClick={() => setCreateModalOpen(true)}>
              Создать обращение
            </Button>
          </Empty>
        </Card>
      ) : (
        <>
          <Table
            dataSource={tickets}
            columns={columns}
            rowKey="id"
            pagination={false}
            size="middle"
            locale={{
              emptyText: (
                <Empty description="Нет обращений с выбранным статусом." />
              ),
            }}
            onRow={(record) => ({
              onClick: () => navigate(`/client/tickets/${record.id}`),
              style: { cursor: 'pointer' },
            })}
          />
          {meta && meta.total_pages! > 1 && (
            <div style={{ marginTop: 16, textAlign: 'right' }}>
              <Pagination
                current={page}
                pageSize={pageSize}
                total={meta.total_count}
                showSizeChanger
                pageSizeOptions={['10', '20', '50']}
                showTotal={(total) => `Всего: ${total}`}
                onChange={(p, ps) => {
                  setPage(p)
                  setPageSize(ps)
                }}
              />
            </div>
          )}
        </>
      )}

      <Modal
        title="Новое обращение"
        open={createModalOpen}
        onCancel={() => {
          setCreateModalOpen(false)
          form.resetFields()
        }}
        onOk={handleCreate}
        okText="Отправить"
        cancelText="Отмена"
        confirmLoading={createMutation.isPending}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="category"
            label="Категория"
            rules={[{ required: true, message: 'Выберите категорию' }]}
          >
            <Select
              placeholder="Выберите категорию"
              options={CATEGORY_OPTIONS}
            />
          </Form.Item>
          <Form.Item
            name="subject"
            label="Тема"
            rules={[{ required: true, message: 'Укажите тему' }]}
          >
            <Input placeholder="Кратко опишите проблему" maxLength={500} />
          </Form.Item>
          <Form.Item
            name="message"
            label="Сообщение"
            rules={[{ required: true, message: 'Введите сообщение' }]}
          >
            <Input.TextArea
              rows={4}
              placeholder="Подробно опишите вашу проблему"
              maxLength={5000}
              showCount
            />
          </Form.Item>
          <Form.Item name="booking_id" label="ID бронирования (необязательно)">
            <Input placeholder="UUID бронирования, если связано" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={null}
        open={faqBotOpen}
        onCancel={() => setFaqBotOpen(false)}
        footer={null}
        width={600}
        destroyOnClose
      >
        <SupportChatBot
          onEscalate={(subject, msg) => {
            setFaqBotOpen(false)
            form.setFieldsValue({
              subject,
              message: msg,
              category: 'question',
            })
            setCreateModalOpen(true)
          }}
        />
      </Modal>
    </div>
  )
}
