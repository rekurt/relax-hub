import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  App,
  Button,
  Card,
  Descriptions,
  Divider,
  Input,
  Modal,
  Space,
  Spin,
  Tag,
  Typography,
} from '@/components/design/system'
import {
  ArrowLeftOutlined,
  ArrowUpOutlined,
  CheckOutlined,
  SendOutlined,
  UserSwitchOutlined,
} from '@/components/design/icons'
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

const { Text, Paragraph } = Typography

function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(' ')
}

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
      <Card className="rh-admin-state-card">
        <Spin size="large" />
      </Card>
    )
  }

  if (!ticket) {
    return (
      <Card>
        <div className="rh-admin-empty-state">
          <div className="rh-admin-empty-state__title">Обращение не найдено</div>
          <p className="rh-admin-empty-state__text">
            Проверьте идентификатор обращения или вернитесь к списку тикетов.
          </p>
        </div>
      </Card>
    )
  }

  const canAct = ticket.status !== 'closed' && ticket.status !== 'resolved'

  return (
    <div className="rh-admin-detail-page">
      <div className="rh-admin-detail-back">
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/admin/tickets')}
        >
          Назад
        </Button>
      </div>

      <Card className="rh-admin-detail-hero">
        <div className="rh-admin-toolbar">
          <div className="rh-admin-toolbar__copy">
            <span className="rh-admin-toolbar__hint">Обращение поддержки</span>
            <h1 className="rh-admin-toolbar__title">{ticket.subject}</h1>
          </div>
          {canAct && (
            <Space className="rh-admin-toolbar__actions" wrap>
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
      </Card>

      <Card className="rh-admin-detail-card">
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

      <Card title="Переписка" className="rh-admin-detail-card rh-admin-message-card">
        {messages.length === 0 ? (
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Нет сообщений</div>
            <p className="rh-admin-empty-state__text">
              В этом обращении пока нет переписки с пользователем.
            </p>
          </div>
        ) : (
          <div className="rh-admin-message-list">
            {messages.map((msg) => {
              const isAdmin = msg.sender_type === 'admin'
              return (
                <div
                  key={msg.id}
                  className={cx('rh-admin-message-row', isAdmin && 'rh-admin-message-row--admin')}
                >
                  <div
                    className={cx('rh-admin-message-bubble', isAdmin && 'rh-admin-message-bubble--admin')}
                  >
                    <div className="rh-admin-message-meta">
                      <Text strong className="rh-admin-message-author">
                        {isAdmin ? 'Поддержка' : 'Пользователь'}
                      </Text>
                      <Text type="secondary" className="rh-admin-message-time">
                        {msg.created_at ? formatDateTime(msg.created_at) : ''}
                      </Text>
                    </div>
                    <Paragraph className="rh-admin-message-body">
                      {msg.body}
                    </Paragraph>
                    {msg.attachments && msg.attachments.length > 0 && (
                      <>
                        <Divider className="rh-admin-message-divider" />
                        <Space orientation="vertical" size={2}>
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
            <Space.Compact className="rh-admin-reply-box">
              <Input.TextArea
                value={messageText}
                onChange={(e) => setMessageText(e.target.value)}
                placeholder="Ответ пользователю..."
                rows={2}
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
                className="rh-admin-reply-button"
              >
                Отправить
              </Button>
            </Space.Compact>
            <Text type="secondary" className="rh-admin-keyboard-hint">
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
        <Text className="rh-admin-modal-label">ID администратора:</Text>
        <Input
          placeholder="UUID администратора"
          value={assignInput}
          onChange={(e) => setAssignInput(e.target.value)}
        />
      </Modal>
    </div>
  )
}
