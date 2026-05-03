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
  Rate,
  Space,
  Spin,
  Tag,
  Typography,
} from 'antd'
import {
  ArrowLeftOutlined,
  SendOutlined,
  SmileOutlined,
} from '@ant-design/icons'
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
      <div style={{ textAlign: 'center', padding: 48 }}>
        <Spin size="large" />
      </div>
    )
  }

  if (!ticket) {
    return <Empty description="Обращение не найдено" />
  }

  const isResolved = ticket.status === 'resolved'
  const isClosed = ticket.status === 'closed'
  const canSendMessage = !isClosed
  const canSubmitCSAT = isResolved && !ticket.csat_score

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/client/tickets')}
        >
          Назад
        </Button>
      </Space>

      <Title level={3} style={{ marginBottom: 16 }}>
        {ticket.subject}
      </Title>

      <Card style={{ marginBottom: 16 }}>
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
          style={{ marginBottom: 16 }}
          title={
            <Space>
              <SmileOutlined />
              <span>Оцените качество поддержки</span>
            </Space>
          }
        >
          <Space orientation="vertical" align="center" style={{ width: '100%' }}>
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
                    justifyContent: isAdmin ? 'flex-start' : 'flex-end',
                  }}
                >
                  <div
                    style={{
                      maxWidth: '70%',
                      padding: '10px 14px',
                      borderRadius: 8,
                      background: isAdmin ? '#f0f0f0' : '#e6f4ff',
                      border: isAdmin ? '1px solid #d9d9d9' : '1px solid #91caff',
                    }}
                  >
                    <div style={{ marginBottom: 4 }}>
                      <Text strong style={{ fontSize: 12 }}>
                        {isAdmin ? 'Поддержка' : 'Вы'}
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
            <Space.Compact style={{ width: '100%' }}>
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
    </div>
  )
}
