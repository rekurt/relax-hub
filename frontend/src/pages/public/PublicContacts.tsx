import { Card, Col, Row, Typography, Tag, Button } from 'antd'
import { ClockCircleOutlined, MailOutlined, MessageOutlined, PhoneOutlined } from '@ant-design/icons'

const { Title, Paragraph, Text } = Typography

const CONTACT_CARDS = [
  {
    key: 'support',
    title: 'Поддержка бронирований',
    description: 'Помогаем подобрать баню, разобраться с оплатой и быстро перевести заявку в рабочее состояние.',
    meta: 'support@bani.ru',
    icon: <MailOutlined />,
  },
  {
    key: 'sales',
    title: 'Для владельцев бань',
    description: 'Если вы подключаете объект, продвижение или хотите demo-показ кабинета владельца, это основной канал.',
    meta: '+7 (495) 555-21-21',
    icon: <PhoneOutlined />,
  },
  {
    key: 'messengers',
    title: 'Быстрый канал',
    description: 'Для уточнений по текущему бронированию и навигации по сервису удобнее всего писать в мессенджер.',
    meta: '@bani_reserve',
    icon: <MessageOutlined />,
  },
]

export default function PublicContacts() {
  return (
    <div style={{ display: 'grid', gap: 24 }}>
      <Card style={{ borderRadius: 24 }}>
        <Tag color="cyan">Контакты</Tag>
        <Title level={1} style={{ marginTop: 12, marginBottom: 12 }}>
          Контур связи для клиента и владельца
        </Title>
        <Paragraph style={{ maxWidth: 760, fontSize: 16, marginBottom: 0 }}>
          На публичной части должны быть ясные контакты до входа и до оплаты. Поэтому мы выводим отдельные каналы для бронирования, подключения объекта и быстрых уточнений.
        </Paragraph>
      </Card>

      <Row gutter={[16, 16]}>
        {CONTACT_CARDS.map((card) => (
          <Col key={card.key} xs={24} md={8}>
            <Card style={{ borderRadius: 24, height: '100%' }}>
              <div style={{ fontSize: 22, marginBottom: 16, color: '#155e63' }}>
                {card.icon}
              </div>
              <Title level={4}>{card.title}</Title>
              <Paragraph type="secondary">{card.description}</Paragraph>
              <Text strong>{card.meta}</Text>
            </Card>
          </Col>
        ))}
      </Row>

      <Card style={{ borderRadius: 24, background: 'linear-gradient(135deg, #17313b 0%, #4d2b21 100%)' }}>
        <Row gutter={[24, 24]} align="middle" justify="space-between">
          <Col xs={24} lg={16}>
            <Title level={3} style={{ color: '#fff', marginBottom: 12 }}>
              Дежурная команда по бронированиям
            </Title>
            <Paragraph style={{ color: 'rgba(255,255,255,0.82)', marginBottom: 8 }}>
              Мы держим отдельный поток для клиентов, которые впервые выбирают баню и не хотят разбираться в интерфейсе самостоятельно.
            </Paragraph>
            <Text style={{ color: 'rgba(255,255,255,0.72)' }}>
              <ClockCircleOutlined style={{ marginRight: 8 }} />
              Ежедневно с 09:00 до 22:00 по Москве
            </Text>
          </Col>
          <Col xs={24} lg={8}>
            <Button type="primary" size="large" href="mailto:support@bani.ru" block>
              Написать в поддержку
            </Button>
          </Col>
        </Row>
      </Card>
    </div>
  )
}
