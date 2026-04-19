import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, Collapse, Row, Col, Typography, Tag, Alert, Spin } from 'antd'
import { AppstoreOutlined, CreditCardOutlined, QuestionCircleOutlined } from '@ant-design/icons'
import { axiosInstance } from '@/api/axios-instance'

const { Title, Paragraph, Text } = Typography

interface PublicFAQItem {
  id: string
  category: string
  question: string
  answer: string
}

const CATEGORY_LABELS: Record<string, { label: string; color: string; icon: React.ReactNode }> = {
  booking: { label: 'Бронирование', color: 'blue', icon: <AppstoreOutlined /> },
  payment: { label: 'Оплата', color: 'gold', icon: <CreditCardOutlined /> },
  cancellation: { label: 'Отмена', color: 'orange', icon: <QuestionCircleOutlined /> },
  wallet: { label: 'Кошелек', color: 'green', icon: <CreditCardOutlined /> },
  account: { label: 'Аккаунт', color: 'purple', icon: <QuestionCircleOutlined /> },
  general: { label: 'Общее', color: 'geekblue', icon: <QuestionCircleOutlined /> },
}

const FALLBACK_FAQ: PublicFAQItem[] = [
  {
    id: 'faq-booking-1',
    category: 'booking',
    question: 'Как быстро подтверждается бронь?',
    answer: 'Мгновенные объекты подтверждаются сразу после создания брони. Для режима "по запросу" срок ответа зависит от владельца, но карточка бани показывает скорость ответа заранее.',
  },
  {
    id: 'faq-payment-1',
    category: 'payment',
    question: 'Нужно ли платить сразу?',
    answer: 'После создания брони вы переходите в кабинет клиента и оплачиваете бронирование стандартным способом. Итоговая цена и условия видны до подтверждения.',
  },
  {
    id: 'faq-cancel-1',
    category: 'cancellation',
    question: 'Где посмотреть условия отмены?',
    answer: 'У каждой бани в карточке и на детальной странице показана политика отмены: гибкая, умеренная или строгая. До бронирования мы показываем ее рядом с ценой.',
  },
]

export default function PublicFAQ() {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['/public/faq'],
    queryFn: async () => {
      const response = await axiosInstance.get<{ data?: PublicFAQItem[] }>('/faq')
      return response.data.data ?? []
    },
    retry: false,
  })

  const items = data && data.length > 0 ? data : FALLBACK_FAQ

  const grouped = useMemo(() => {
    return items.reduce<Record<string, PublicFAQItem[]>>((acc, item) => {
      if (!acc[item.category]) {
        acc[item.category] = []
      }
      acc[item.category]!.push(item)
      return acc
    }, {})
  }, [items])

  if (isLoading) {
    return <Spin size="large" style={{ display: 'block', margin: '96px auto' }} />
  }

  return (
    <div style={{ display: 'grid', gap: 24 }}>
      <Card style={{ borderRadius: 24 }}>
        <Tag color="gold">Поддержка</Tag>
        <Title level={1} style={{ marginTop: 12, marginBottom: 12 }}>
          Вопросы до первой брони
        </Title>
        <Paragraph style={{ maxWidth: 760, fontSize: 16, marginBottom: 0 }}>
          Мы собрали короткие ответы по бронированию, оплате и условиям отмены, чтобы клиент видел понятные правила еще до входа в кабинет.
        </Paragraph>
      </Card>

      {isError && (
        <Alert
          type="info"
          showIcon
          title="Показываем базовую подборку ответов"
          description="Публичный FAQ временно недоступен, поэтому страница использует безопасный демо-набор ответов."
        />
      )}

      <Row gutter={[16, 16]}>
        {Object.entries(grouped).map(([category, categoryItems]) => {
          const meta = CATEGORY_LABELS[category] ?? CATEGORY_LABELS.general ?? {
            label: 'Общее',
            color: 'geekblue',
            icon: <QuestionCircleOutlined />,
          }
          return (
            <Col key={category} xs={24} lg={12}>
              <Card
                title={(
                  <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
                    {meta.icon}
                    {meta.label}
                  </span>
                )}
                extra={<Tag color={meta.color}>{categoryItems.length}</Tag>}
                style={{ borderRadius: 24, height: '100%' }}
              >
                <Collapse
                  ghost
                  items={categoryItems.map((item) => ({
                    key: item.id,
                    label: item.question,
                    children: <Text type="secondary">{item.answer}</Text>,
                  }))}
                />
              </Card>
            </Col>
          )
        })}
      </Row>
    </div>
  )
}
