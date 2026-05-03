import { Button, Card, Col, Row, Typography } from 'antd'
import { ClockCircleOutlined, MailOutlined, MessageOutlined, PhoneOutlined } from '@ant-design/icons'
import PageHeader from '@/components/PageHeader'
import { CONTACT_CARDS, PLATFORM_CONTACTS } from '@/content/support'

const { Text } = Typography

const CONTACT_ICONS = {
  support: <MailOutlined />,
  sales: <PhoneOutlined />,
  messengers: <MessageOutlined />,
} as const

export default function PublicContacts() {
  return (
    <div className="bani-stack">
      <PageHeader
        eyebrow="Контакты"
        title="Контур связи для клиента и владельца"
        description="На публичной части контакты должны отвечать на простой вопрос: куда писать прямо сейчас. Поэтому раздел разделён по задачам, а не по внутренней структуре компании."
      />

      <section className="bani-hero-panel bani-hero-panel--dark">
        <div className="bani-hero-panel__eyebrow">Связь</div>
        <h2 className="bani-hero-panel__title">Куда обращаться до входа, до оплаты и после бронирования</h2>
        <div className="bani-hero-panel__description">
          Пользователь не должен гадать, писать ли в поддержку, в продажи или в чат. Мы явно развели каналы по контексту использования и оставили самый быстрый путь в один клик.
        </div>
        <div className="bani-hero-panel__meta">
          <div className="bani-hero-panel__meta-item">
            <span className="bani-hero-panel__meta-label">Поддержка клиентов</span>
            <div className="bani-hero-panel__meta-value">{PLATFORM_CONTACTS.supportHours}</div>
          </div>
          <div className="bani-hero-panel__meta-item">
            <span className="bani-hero-panel__meta-label">Подключение объектов</span>
            <div className="bani-hero-panel__meta-value">Отдельный канал для владельцев</div>
          </div>
          <div className="bani-hero-panel__meta-item">
            <span className="bani-hero-panel__meta-label">Быстрые уточнения</span>
            <div className="bani-hero-panel__meta-value">Мессенджер без лишних переходов</div>
          </div>
        </div>
      </section>

      <Row gutter={[16, 16]}>
        {CONTACT_CARDS.map((card) => (
          <Col key={card.key} xs={24} md={8}>
            <Card className="bani-equal-card">
              <div className="bani-feature-item__icon" style={{ marginBottom: 16 }}>
                {CONTACT_ICONS[card.key]}
              </div>
              <div className="bani-section-card">
                <h2 className="bani-section-card__title" style={{ fontSize: 22 }}>
                  {card.title}
                </h2>
                <div className="bani-section-card__description">{card.description}</div>
                <Text strong>{card.meta}</Text>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      <div className="bani-grid bani-grid--content-aside">
        <Card>
          <div className="bani-section-card">
            <h2 className="bani-section-card__title">Как быстрее получить ответ</h2>
            <div className="bani-feature-list">
              <div className="bani-feature-item">
                <div className="bani-feature-item__icon"><MailOutlined /></div>
                <div className="bani-feature-item__copy">
                  <div className="bani-feature-item__title">Есть номер брони</div>
                  <div className="bani-feature-item__description">Сразу указывайте номер заказа, дату визита и канал оплаты. Это сокращает время уточнений.</div>
                </div>
              </div>
              <div className="bani-feature-item">
                <div className="bani-feature-item__icon"><PhoneOutlined /></div>
                <div className="bani-feature-item__copy">
                  <div className="bani-feature-item__title">Нужен живой контакт</div>
                  <div className="bani-feature-item__description">Для срочных кейсов вроде переноса, возврата или проблемы со входом в аккаунт лучше звонок, а не форма.</div>
                </div>
              </div>
              <div className="bani-feature-item">
                <div className="bani-feature-item__icon"><MessageOutlined /></div>
                <div className="bani-feature-item__copy">
                  <div className="bani-feature-item__title">Только короткое уточнение</div>
                  <div className="bani-feature-item__description">Если нужно быстро спросить про адрес, время или доступность, удобнее мессенджер без длинной переписки.</div>
                </div>
              </div>
            </div>
          </div>
        </Card>

        <section className="bani-hero-panel bani-hero-panel--dark">
          <div className="bani-hero-panel__eyebrow">Приоритетный канал</div>
          <h2 className="bani-hero-panel__title">Дежурная команда по бронированиям</h2>
          <div className="bani-hero-panel__description">
            Отдельный поток для клиентов, которые впервые выбирают баню и не хотят разбираться в интерфейсе самостоятельно.
          </div>
          <div className="bani-hero-panel__meta-item">
            <span className="bani-hero-panel__meta-label">
              <ClockCircleOutlined style={{ marginRight: 8 }} />
              Часы работы
            </span>
            <div className="bani-hero-panel__meta-value">{PLATFORM_CONTACTS.supportHours}</div>
          </div>
          <Button type="primary" size="large" href={`mailto:${PLATFORM_CONTACTS.supportEmail}`} block>
            Написать в поддержку
          </Button>
        </section>
      </div>
    </div>
  )
}
