import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  App,
  Button,
  Card,
  Descriptions,
  Divider,
  Input,
  Rate,
  Space,
  Spin,
  Tag,
  Typography,
} from '@/components/design/system'
import {
  ArrowLeftOutlined,
  SendOutlined,
  SmileOutlined,
} from '@/components/design/icons'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetMyTicketsId,
  useGetMyTicketsIdMessages,
  getGetMyTicketsIdMessagesQueryKey,
  getGetMyTicketsIdQueryKey,
  usePostMyTicketsIdMessages,
  usePostMyTicketsIdCsat,
} from '@/api/generated/support/support'
import type {
  InternalHandlerTicketMessageResponse,
} from '@/api/generated/model'
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

const csatLabels: Record<number, string> = {
  1: 'Ужасно',
  2: 'Плохо',
  3: 'Нормально',
  4: 'Хорошо',
  5: 'Отлично',
}

export default function TicketDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { message } = App.useApp()
  const queryClient = useQueryClient()

  const [messageText, setMessageText] = useState('')
  const [csatScore, setCsatScore] = useState(0)

  const { data: ticketData, isLoading: ticketLoading } = useGetMyTicketsId(id!)
  const { data: messagesData, isLoading: messagesLoading } = useGetMyTicketsIdMessages(id!)

  const sendMutation = usePostMyTicketsIdMessages()
  const csatMutation = usePostMyTicketsIdCsat()

  const ticket = ticketData?.data
  const messages: InternalHandlerTicketMessageResponse[] = messagesData?.data ?? []

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

  const handleSubmitCSAT = async () => {
    if (csatScore === 0) {
      message.warning('Выберите оценку')
      return
    }
    try {
      await csatMutation.mutateAsync({
        id: id!,
        data: { score: csatScore },
      })
      message.success('Спасибо за оценку!')
      queryClient.invalidateQueries({ queryKey: getGetMyTicketsIdQueryKey(id) })
    } catch {
      message.error('Не удалось отправить оценку')
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
            Проверьте ссылку или вернитесь к списку обращений.
          </p>
        </div>
      </Card>
    )
  }

  const isResolved = ticket.status === 'resolved'
  const isClosed = ticket.status === 'closed'
  const canSendMessage = !isClosed
  const canSubmitCSAT = isResolved && !ticket.csat_score

  return (
    <div className="rh-admin-detail-page">
      <div className="rh-admin-detail-back">
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/client/tickets')}
        >
          Назад
        </Button>
      </div>

      <Card className="rh-admin-detail-hero">
        <div className="rh-admin-toolbar">
          <div className="rh-admin-toolbar__copy">
            <span className="rh-admin-toolbar__hint">Обращение в поддержку</span>
            <h1 className="rh-admin-toolbar__title">{ticket.subject}</h1>
          </div>
        </div>
      </Card>

      <Card className="rh-admin-detail-card">
        <Descriptions column={{ xs: 1, sm: 2 }} size="small">
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
            {ticket.level ?? '-'}
          </Descriptions.Item>
          <Descriptions.Item label="Создан">
            {ticket.created_at ? formatDateTime(ticket.created_at) : '-'}
          </Descriptions.Item>
          {ticket.resolved_at && (
            <Descriptions.Item label="Решён">
              {formatDateTime(ticket.resolved_at)}
            </Descriptions.Item>
          )}
          {ticket.csat_score && (
            <Descriptions.Item label="Ваша оценка">
              <Rate disabled value={ticket.csat_score} />
            </Descriptions.Item>
          )}
        </Descriptions>
      </Card>

      {canSubmitCSAT && (
        <Card
          className="rh-admin-detail-card"
          title={
            <Space>
              <SmileOutlined />
              <span>Оцените качество поддержки</span>
            </Space>
          }
        >
          <Space orientation="vertical" align="center" className="rh-client-csat-panel">
            <Rate
              value={csatScore}
              onChange={setCsatScore}
              tooltips={Object.values(csatLabels)}
            />
            {csatScore > 0 && (
              <Text type="secondary">{csatLabels[csatScore]}</Text>
            )}
            <Button
              type="primary"
              onClick={handleSubmitCSAT}
              loading={csatMutation.isPending}
              disabled={csatScore === 0}
            >
              Отправить оценку
            </Button>
          </Space>
        </Card>
      )}

      <Card title="Переписка" className="rh-admin-detail-card rh-admin-message-card">
        {messages.length === 0 ? (
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Нет сообщений</div>
            <p className="rh-admin-empty-state__text">
              История обращения пока пуста.
            </p>
          </div>
        ) : (
          <div className="rh-admin-message-list">
            {messages.map((msg) => {
              const isAdmin = msg.sender_type === 'admin'
              const isOwnMessage = !isAdmin
              return (
                <div
                  key={msg.id}
                  className={cx('rh-admin-message-row', isOwnMessage && 'rh-client-message-row--own')}
                >
                  <div
                    className={cx('rh-admin-message-bubble', isOwnMessage && 'rh-client-message-bubble--own')}
                  >
                    <div className="rh-admin-message-meta">
                      <Text strong className="rh-admin-message-author">
                        {isAdmin ? 'Поддержка' : 'Вы'}
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

        {canSendMessage && (
          <>
            <Divider />
            <Space.Compact className="rh-admin-reply-box">
              <Input.TextArea
                value={messageText}
                onChange={(e) => setMessageText(e.target.value)}
                placeholder="Введите сообщение..."
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
    </div>
  )
}
