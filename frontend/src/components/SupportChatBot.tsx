import { useState, useRef, useEffect } from 'react'
import {
  Button,
  Card,
  Input,
  Space,
  Typography,
  Collapse,
  Tag,
  Divider,
} from '@/components/design/system'
import {
  QuestionCircleOutlined,
  SendOutlined,
  CustomerServiceOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@/components/design/icons'
import { axiosInstance } from '@/api/axios-instance'

const { Text, Paragraph } = Typography

interface FAQMatch {
  faq: {
    id: string
    category: string
    question: string
    answer: string
  }
  score: number
}

interface ChatMessage {
  type: 'user' | 'bot' | 'faq_results'
  text?: string
  matches?: FAQMatch[]
}

const CATEGORY_LABELS: Record<string, string> = {
  booking: 'Бронирование',
  payment: 'Оплата',
  cancellation: 'Отмена',
  wallet: 'Кошелёк',
  account: 'Аккаунт',
  general: 'Общее',
}

const CATEGORY_COLORS: Record<string, string> = {
  booking: 'green',
  payment: 'green',
  cancellation: 'orange',
  wallet: 'gold',
  account: 'green',
  general: 'default',
}

interface SupportChatBotProps {
  onEscalate: (subject: string, message: string) => void
}

export default function SupportChatBot({ onEscalate }: SupportChatBotProps) {
  const [messages, setMessages] = useState<ChatMessage[]>([
    {
      type: 'bot',
      text: 'Здравствуйте! Опишите вашу проблему, и я постараюсь помочь. Если мой ответ не подойдёт, я передам ваш вопрос специалисту.',
    },
  ])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [lastQuery, setLastQuery] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleSend = async () => {
    const query = input.trim()
    if (!query || loading) return

    setInput('')
    setLastQuery(query)
    setMessages((prev) => [...prev, { type: 'user', text: query }])
    setLoading(true)

    try {
      const response = await axiosInstance.post<{
        success: boolean
        data: FAQMatch[]
      }>('/my/support/faq-match', { query })
      const matches = response.data?.data || []

      if (matches.length > 0) {
        setMessages((prev) => [
          ...prev,
          {
            type: 'bot',
            text: 'Вот что я нашёл по вашему вопросу:',
          },
          { type: 'faq_results', matches },
        ])
      } else {
        setMessages((prev) => [
          ...prev,
          {
            type: 'bot',
            text: 'К сожалению, я не нашёл подходящего ответа. Хотите создать обращение в поддержку?',
          },
        ])
      }
    } catch {
      setMessages((prev) => [
        ...prev,
        {
          type: 'bot',
          text: 'Произошла ошибка при поиске. Попробуйте ещё раз или создайте обращение.',
        },
      ])
    } finally {
      setLoading(false)
    }
  }

  const handleNotHelpful = () => {
    setMessages((prev) => [
      ...prev,
      {
        type: 'bot',
        text: 'Создаю обращение в поддержку...',
      },
    ])
    onEscalate(lastQuery.slice(0, 200), lastQuery)
  }

  const handleHelpful = () => {
    setMessages((prev) => [
      ...prev,
      {
        type: 'bot',
        text: 'Рад, что смог помочь! Если возникнут ещё вопросы — обращайтесь.',
      },
    ])
  }

  return (
    <Card
      title={
        <Space>
          <CustomerServiceOutlined />
          <span>Помощник поддержки</span>
        </Space>
      }
      className="rh-support-chat-card"
    >
      <div className="rh-support-chat__messages">
        {messages.map((msg, idx) => {
          if (msg.type === 'user') {
            return (
              <div key={idx} className="rh-support-chat__row rh-support-chat__row--user">
                <div className="rh-support-chat__bubble rh-support-chat__bubble--user">
                  <Text className="rh-support-chat__user-text">{msg.text}</Text>
                </div>
              </div>
            )
          }

          if (msg.type === 'bot') {
            return (
              <div key={idx} className="rh-support-chat__row rh-support-chat__row--bot">
                <div className="rh-support-chat__bubble rh-support-chat__bubble--bot">
                  <Text>{msg.text}</Text>
                </div>
              </div>
            )
          }

          if (msg.type === 'faq_results' && msg.matches) {
            return (
              <div key={idx} className="rh-support-chat__faq">
                <Collapse
                  accordion
                  items={msg.matches.map((m, i) => ({
                    key: i.toString(),
                    label: (
                      <Space>
                        <QuestionCircleOutlined />
                        <Text strong>{m.faq.question}</Text>
                        <Tag color={CATEGORY_COLORS[m.faq.category]}>
                          {CATEGORY_LABELS[m.faq.category] || m.faq.category}
                        </Tag>
                      </Space>
                    ),
                    children: <Paragraph>{m.faq.answer}</Paragraph>,
                  }))}
                />
                <Divider className="rh-support-chat__divider" />
                <Space>
                  <Button
                    type="primary"
                    icon={<CheckCircleOutlined />}
                    onClick={handleHelpful}
                  >
                    Помогло
                  </Button>
                  <Button
                    icon={<CloseCircleOutlined />}
                    onClick={handleNotHelpful}
                  >
                    Не помогло — создать обращение
                  </Button>
                </Space>
              </div>
            )
          }

          return null
        })}
        <div ref={messagesEndRef} />
      </div>
      <div className="rh-support-chat__composer">
        <Space.Compact className="rh-full-width">
          <Input
            placeholder="Опишите вашу проблему..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onPressEnter={handleSend}
            disabled={loading}
          />
          <Button
            type="primary"
            icon={<SendOutlined />}
            onClick={handleSend}
            loading={loading}
          />
        </Space.Compact>
      </div>
    </Card>
  )
}
