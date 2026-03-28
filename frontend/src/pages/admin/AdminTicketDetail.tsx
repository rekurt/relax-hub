import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  App,
  Button,
  Card,
  Descriptions,
  Divider,
  Empty,
  Input,
  Modal,
  Space,
  Spin,
  Tag,
  Typography,
} from 'antd'
import {
  ArrowLeftOutlined,
  ArrowUpOutlined,
  CheckOutlined,
  SendOutlined,
  UserSwitchOutlined,
} from '@ant-design/icons'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetAdminTicketsId,
  getGetAdminTicketsIdQueryKey,
  getGetAdminTicketsQueryKey,
  usePatchAdminTicketsIdAssign,
  usePatchAdminTicketsIdEscalate,
  usePatchAdminTicketsIdResolve,
  usePostAdminTicketsIdMessages,
} from '@/api/generated/support-admin/support-admin'
import {
  useGetMyTicketsIdMessages,
  getGetMyTicketsIdMessagesQueryKey,
} from '@/api/generated/support/support'
import type { InternalHandlerTicketMessageResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'

const { Title, Text, Paragraph } = Typography

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

export default function AdminTicketDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { modal, message } = App.useApp()
  const queryClient = useQueryClient()

  const [messageText, setMessageText] = useState('')
  const [assignInput, setAssignInput] = useState('')
  const [assignModalOpen, setAssignModalOpen] = useState(false)

  const { data: ticketData, isLoading: ticketLoading } = useGetAdminTicketsId(id!)
  const { data: messagesData, isLoading: messagesLoading } = useGetMyTicketsIdMessages(id!)

  const sendMutation = usePostAdminTicketsIdMessages()
  const assignMutation = usePatchAdminTicketsIdAssign()
  const escalateMutation = usePatchAdminTicketsIdEscalate()
  const resolveMutation = usePatchAdminTicketsIdResolve()

  const ticket = ticketData?.data
  const messages: InternalHandlerTicketMessageResponse[] = messagesData?.data ?? []

  const invalidateAll = () => {
    queryClient.invalidateQueries({ queryKey: getGetAdminTicketsIdQueryKey(id) })
    queryClient.invalidateQueries({ queryKey: getGetMyTicketsIdMessagesQueryKey(id) })
    queryClient.invalidateQueries({ queryKey: getGetAdminTicketsQueryKey() })
  }

  const handleSendMessage = async () => {
    if (!messageText.trim()) return
    try {
      await sendMutation.mutateAsync({
        id: id!,
        data: { body: messageText.trim() },
      })
      setMessageText('')
      queryClient.invalidateQueries({ queryKey: getGetMyTicketsIdMessagesQueryKey(id) })
    } catch {
      message.error('Не удалось отправить сообщение')
    }
  }

  const handleEscalate = () => {
    modal.confirm({
      title: 'Эскалировать обращение?',
      content: 'Обращение будет передано на следующий уровень поддержки.',
      okText: 'Эскалировать',
      cancelText: 'Отмена',
      onOk: () =>
        escalateMutation
          .mutateAsync({ id: id! })
          .then(() => {
            message.success('Обращение эскалировано')
            invalidateAll()
          })
          .catch(() => {
            message.error('Не удалось эскалировать')
          }),
    })
  }

  const handleResolve = () => {
    modal.confirm({
      title: 'Отметить как решённое?',
      content: 'Обращение будет переведено в статус "Решённое". Пользователь получит возможность оценить качество поддержки.',
      okText: 'Решить',
      cancelText: 'Отмена',
      onOk: () =>
        resolveMutation
          .mutateAsync({ id: id! })
          .then(() => {
            message.success('Обращение решено')
            invalidateAll()
          })
          .catch(() => {
            message.error('Не удалось решить обращение')
          }),
    })
  }

  const handleAssign = async () => {
    if (!assignInput.trim()) {
      message.warning('Укажите ID администратора')
      return
    }
    try {
      await assignMutation.mutateAsync({
        id: id!,
        data: { assigned_to: assignInput.trim() },
      })
      message.success('Обращение назначено')
      setAssignModalOpen(false)
      setAssignInput('')
      invalidateAll()
    } catch {
      message.error('Не удалось назначить обращение')
    }
  }

  if (ticketLoading || messagesLoading) {
    return (
      <div style={{ textAlign: 'center', padding: 48 }}>
        <Spin size="large" />
      </div>
    )
  }

  if (!ticket) {
    return <Empty description="Обращение не найдено" />
  }

  const canAct = ticket.status !== 'closed' && ticket.status !== 'resolved'

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/admin/tickets')}
        >
          Назад
        </Button>
      </Space>

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16, flexWrap: 'wrap', gap: 8 }}>
        <Title level={3} style={{ margin: 0 }}>
          {ticket.subject}
        </Title>
        {canAct && (
          <Space wrap>
            <Button
              icon={<UserSwitchOutlined />}
              onClick={() => setAssignModalOpen(true)}
            >
              Назначить
            </Button>
            <Button
              icon={<ArrowUpOutlined />}
              onClick={handleEscalate}
              loading={escalateMutation.isPending}
            >
              Эскалировать
            </Button>
            <Button
              type="primary"
              icon={<CheckOutlined />}
              onClick={handleResolve}
              loading={resolveMutation.isPending}
            >
              Решить
            </Button>
          </Space>
        )}
      </div>

      <Card style={{ marginBottom: 16 }}>
        <Descriptions column={{ xs: 1, sm: 2, md: 3 }} size="small">
          <Descriptions.Item label="Статус">
            <Tag color={statusColor[ticket.status!] ?? 'default'}>
              {statusLabel[ticket.status!] ?? ticket.status}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Категория">
            {categoryLabel[ticket.category!] ?? ticket.category}
          </Descriptions.Item>
          <Descriptions.Item label="Приоритет">
            <Tag color={priorityColor[ticket.priority!] ?? 'default'}>
              {priorityLabel[ticket.priority!] ?? ticket.priority}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Уровень">
            <Tag>{ticket.level ?? '-'}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Пользователь">
            {ticket.user_id ? (
              <Text copyable={{ text: ticket.user_id }}>
                {ticket.user_id.slice(0, 8)}...
              </Text>
            ) : (
              '-'
            )}
          </Descriptions.Item>
          <Descriptions.Item label="Назначен">
            {ticket.assigned_to ? (
              <Text copyable={{ text: ticket.assigned_to }}>
                {ticket.assigned_to.slice(0, 8)}...
              </Text>
            ) : (
              <Text type="secondary">Не назначен</Text>
            )}
          </Descriptions.Item>
          <Descriptions.Item label="Создан">
            {ticket.created_at ? formatDateTime(ticket.created_at) : '-'}
          </Descriptions.Item>
          {ticket.resolved_at && (
            <Descriptions.Item label="Решён">
              {formatDateTime(ticket.resolved_at)}
            </Descriptions.Item>
          )}
          {ticket.booking_id && (
            <Descriptions.Item label="Бронирование">
              <Text copyable={{ text: ticket.booking_id }}>
                {ticket.booking_id.slice(0, 8)}...
              </Text>
            </Descriptions.Item>
          )}
          {ticket.csat_score != null && (
            <Descriptions.Item label="CSAT">
              {ticket.csat_score}/5
            </Descriptions.Item>
          )}
        </Descriptions>
      </Card>

      <Card title="Переписка" style={{ marginBottom: 16 }}>
        {messages.length === 0 ? (
          <Empty description="Нет сообщений" />
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            {messages.map((msg) => {
              const isAdmin = msg.sender_type === 'admin'
              return (
                <div
                  key={msg.id}
                  style={{
                    display: 'flex',
                    justifyContent: isAdmin ? 'flex-end' : 'flex-start',
                  }}
                >
                  <div
                    style={{
                      maxWidth: '70%',
                      padding: '10px 14px',
                      borderRadius: 8,
                      background: isAdmin ? '#f6ffed' : '#f0f0f0',
                      border: isAdmin ? '1px solid #b7eb8f' : '1px solid #d9d9d9',
                    }}
                  >
                    <div style={{ marginBottom: 4 }}>
                      <Text strong style={{ fontSize: 12 }}>
                        {isAdmin ? 'Поддержка' : 'Пользователь'}
                      </Text>
                      <Text type="secondary" style={{ fontSize: 11, marginLeft: 8 }}>
                        {msg.created_at ? formatDateTime(msg.created_at) : ''}
                      </Text>
                    </div>
                    <Paragraph style={{ margin: 0, whiteSpace: 'pre-wrap' }}>
                      {msg.body}
                    </Paragraph>
                    {msg.attachments && msg.attachments.length > 0 && (
                      <>
                        <Divider style={{ margin: '8px 0' }} />
                        <Space direction="vertical" size={2}>
                          {msg.attachments.map((url, idx) => (
                            <a
                              key={idx}
                              href={url}
                              target="_blank"
                              rel="noopener noreferrer"
                            >
                              Вложение {idx + 1}
                            </a>
                          ))}
                        </Space>
                      </>
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        )}

        {canAct && (
          <>
            <Divider />
            <Space.Compact style={{ width: '100%' }}>
              <Input.TextArea
                value={messageText}
                onChange={(e) => setMessageText(e.target.value)}
                placeholder="Ответ пользователю..."
                autoSize={{ minRows: 2, maxRows: 6 }}
                maxLength={5000}
                onPressEnter={(e) => {
                  if (e.ctrlKey || e.metaKey) {
                    handleSendMessage()
                  }
                }}
              />
              <Button
                type="primary"
                icon={<SendOutlined />}
                onClick={handleSendMessage}
                loading={sendMutation.isPending}
                disabled={!messageText.trim()}
                style={{ height: 'auto' }}
              >
                Отправить
              </Button>
            </Space.Compact>
            <Text type="secondary" style={{ fontSize: 11 }}>
              Ctrl+Enter для отправки
            </Text>
          </>
        )}
      </Card>

      <Modal
        title="Назначить обращение"
        open={assignModalOpen}
        onCancel={() => {
          setAssignModalOpen(false)
          setAssignInput('')
        }}
        onOk={handleAssign}
        okText="Назначить"
        cancelText="Отмена"
        confirmLoading={assignMutation.isPending}
      >
        <div style={{ marginBottom: 8 }}>
          <Text>ID администратора:</Text>
        </div>
        <Input
          placeholder="UUID администратора"
          value={assignInput}
          onChange={(e) => setAssignInput(e.target.value)}
        />
      </Modal>
    </div>
  )
}
