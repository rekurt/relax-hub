import { useMemo, useState } from 'react'
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
  Spin,
  Table,
  Tag,
  Typography,
} from '@/components/design/system'
import { PlusOutlined, RobotOutlined } from '@/components/design/icons'
import type { ColumnsType } from '@/components/design/types'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import {
  getGetMyTicketsQueryKey,
  useGetMyTickets,
  usePostMyTickets,
} from '@/api/generated/support/support'
import type { InternalHandlerTicketResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'
import SupportChatBot from '@/components/SupportChatBot'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

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

  const allTickets: InternalHandlerTicketResponse[] = useMemo(() => data?.data ?? [], [data?.data])
  const tickets = statusFilter
    ? allTickets.filter((ticket) => ticket.status === statusFilter)
    : allTickets
  const meta = data?.meta

  const stats = useMemo(() => {
    return allTickets.reduce(
      (acc, ticket) => {
        if (ticket.status === 'open') acc.open += 1
        if (ticket.status === 'in_progress') acc.inProgress += 1
        if (ticket.status === 'resolved') acc.resolved += 1
        return acc
      },
      { open: 0, inProgress: 0, resolved: 0 },
    )
  }, [allTickets])

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
      render: (category: string) => categoryLabel[category] ?? category,
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
      width: 170,
      render: (value: string) => (value ? formatDateTime(value) : '-'),
    },
  ]

  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="Поддержка"
        title="Мои обращения"
        description="Экран показывает состояние текущих кейсов без лишнего кликанья: сверху быстрый обзор, ниже фильтр по статусу и рабочая таблица."
        extra={(
          <div className="rh-toolbar__group">
            <Button icon={<RobotOutlined />} onClick={() => setFaqBotOpen(true)}>
              Быстрая помощь
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateModalOpen(true)}>
              Новое обращение
            </Button>
          </div>
        )}
      />

      <section className="rh-hero-panel">
        <div className="rh-hero-panel__eyebrow">Состояние поддержки</div>
        <h2 className="rh-hero-panel__title">Пользователь видит статус вопроса за один взгляд</h2>
        <div className="rh-hero-panel__description">
          Вместо пустого списка с таблицей сначала показываем текущую картину по обращениям. Это снижает тревогу: пользователь понимает, что уже открыто, что в работе и что закрыто.
        </div>
          <div className="rh-stat-grid">
            <div className="rh-stat-tile">
            <span className="rh-stat-tile__eyebrow">Открыто сейчас</span>
            <div className="rh-stat-tile__value">{stats.open}</div>
            <span className="rh-stat-tile__hint">Требуют реакции поддержки</span>
          </div>
          <div className="rh-stat-tile">
            <span className="rh-stat-tile__eyebrow">Активные кейсы</span>
            <div className="rh-stat-tile__value">{stats.inProgress}</div>
            <span className="rh-stat-tile__hint">По ним уже идёт коммуникация или разбор</span>
          </div>
          <div className="rh-stat-tile">
            <span className="rh-stat-tile__eyebrow">Решено</span>
            <div className="rh-stat-tile__value">{stats.resolved}</div>
            <span className="rh-stat-tile__hint">История кейсов, к которой можно вернуться</span>
          </div>
        </div>
      </section>

      <Card>
        <div className="rh-table-shell">
          <div className="rh-toolbar">
            <div>
              <h2 className="rh-section-card__title">Лента обращений</h2>
              <div className="rh-section-card__description">
                Фильтр по статусу вынесен наверх и не мешает чтению таблицы. Клик по строке открывает конкретный диалог с поддержкой.
              </div>
            </div>
          </div>

          <Segmented
            options={STATUS_OPTIONS}
            value={statusFilter}
            onChange={(value) => {
              setStatusFilter(value as string)
              setPage(1)
            }}
          />

          {isLoading ? (
            <div className="rh-feed-empty">
              <Spin size="large" />
            </div>
          ) : tickets.length === 0 ? (
            <div className="rh-feed-empty">
              <Empty description="Нет обращений">
                <Button type="primary" onClick={() => setCreateModalOpen(true)}>
                  Создать обращение
                </Button>
              </Empty>
            </div>
          ) : (
            <>
              <Table
                dataSource={tickets}
                columns={columns}
                rowKey="id"
                pagination={false}
                size="middle"
                locale={{ emptyText: <Empty description="Нет обращений с выбранным статусом." /> }}
                onRow={(record) => ({
                  onClick: () => navigate(`/client/tickets/${record.id}`),
                  style: { cursor: 'pointer' },
                })}
              />

              {meta && meta.total_pages! > 1 && (
                <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                  <Pagination
                    current={page}
                    pageSize={pageSize}
                    total={meta.total_count}
                    showSizeChanger
                    pageSizeOptions={['10', '20', '50']}
                    showTotal={(total) => `Всего: ${total}`}
                    onChange={(nextPage, nextPageSize) => {
                      setPage(nextPage)
                      setPageSize(nextPageSize)
                    }}
                  />
                </div>
              )}
            </>
          )}
        </div>
      </Card>

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
        destroyOnHidden
      >
        <SupportChatBot
          onEscalate={(subject, messageText) => {
            setFaqBotOpen(false)
            form.setFieldsValue({
              subject,
              message: messageText,
              category: 'question',
            })
            setCreateModalOpen(true)
          }}
        />
      </Modal>
    </div>
  )
}
