import { Card, Col, Row, Typography } from '@/components/design/system'
import PageHeader from '@/components/PageHeader'
import { PLATFORM_CONTACTS, PLATFORM_NAME, TERMS_LAST_UPDATED } from '@/content/support'

const { Text } = Typography

const TERMS_SECTIONS = [
  {
    key: 'platform',
    title: 'Назначение платформы',
    description: `${PLATFORM_NAME} помогает выбрать объект, проверить доступность, оформить бронирование и сопровождать коммуникацию между клиентом и площадкой до завершения визита.`,
  },
  {
    key: 'account',
    title: 'Аккаунт, SMS и вход',
    description: 'Для доступа к части сценариев может потребоваться подтверждение номера телефона. Пользователь отвечает за корректность номера и безопасность входа в аккаунт.',
  },
  {
    key: 'booking',
    title: 'Бронирование и оплата',
    description: 'До подтверждения брони клиент видит состав заказа, интервал, стоимость и доступный способ оплаты. Оплата через платформу подтверждает согласие с правилами объекта.',
  },
  {
    key: 'cancellation',
    title: 'Отмена и ответственность объекта',
    description: 'Политика отмены зависит от условий конкретной площадки. Платформа отображает эти условия заранее, но ответственность за оказание услуги несёт объект размещения.',
  },
  {
    key: 'reviews',
    title: 'Отзывы и пользовательский контент',
    description: 'Оставляя отзыв, текст или фотографии, пользователь подтверждает, что контент относится к реальному визиту, не нарушает закон и может быть использован в интерфейсе платформы.',
  },
  {
    key: 'liability',
    title: 'Ограничение ответственности',
    description: `${PLATFORM_NAME} отвечает за работу платформы как цифрового сервиса, но не гарантирует действия третьих лиц, форс-мажорные обстоятельства и качество офлайн-услуг вне своей операционной зоны.`,
  },
] as const

export default function PublicTerms() {
  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="Юридическая рамка"
        title={`Условия использования ${PLATFORM_NAME}`}
        description="Эта версия условий описывает роль платформы, порядок бронирования, правила пользовательского контента и базовые ограничения ответственности."
      />

      <section className="rh-hero-panel rh-hero-panel--dark">
        <div className="rh-hero-panel__eyebrow">Условия</div>
        <h2 className="rh-hero-panel__title">Понятные правила использования сервиса до первой оплаты.</h2>
        <div className="rh-hero-panel__description">
          Мы не прячем важные условия в мелкий шрифт. Документ фиксирует базовые договорённости между пользователем, платформой и объектом, где проходит бронирование.
        </div>
        <div className="rh-hero-panel__meta">
          <div className="rh-hero-panel__meta-item">
            <span className="rh-hero-panel__meta-label">Версия</span>
            <div className="rh-hero-panel__meta-value">{TERMS_LAST_UPDATED}</div>
          </div>
          <div className="rh-hero-panel__meta-item">
            <span className="rh-hero-panel__meta-label">Поддержка</span>
            <div className="rh-hero-panel__meta-value">{PLATFORM_CONTACTS.supportEmail}</div>
          </div>
          <div className="rh-hero-panel__meta-item">
            <span className="rh-hero-panel__meta-label">Часы связи</span>
            <div className="rh-hero-panel__meta-value">{PLATFORM_CONTACTS.supportHours}</div>
          </div>
        </div>
      </section>

      <Row gutter={[16, 16]}>
        {TERMS_SECTIONS.map((section) => (
          <Col key={section.key} xs={24} md={12}>
            <Card className="rh-equal-card">
              <div className="rh-section-card">
                <h2 className="rh-section-card__title">{section.title}</h2>
                <div className="rh-section-card__description">{section.description}</div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      <Card>
        <div className="rh-section-card">
          <h2 className="rh-section-card__title">Поддержка и применимое общение</h2>
          <div className="rh-section-card__description">
            Если у пользователя возникает спор по брони, оплате или контенту, первичный канал связи проходит через поддержку {PLATFORM_NAME}. Для срочных кейсов используйте {PLATFORM_CONTACTS.supportEmail}, {PLATFORM_CONTACTS.supportPhone} или {PLATFORM_CONTACTS.supportMessenger}.
          </div>
          <Text type="secondary">Документ действует с даты публикации и может обновляться при изменении продукта или регуляторных требований.</Text>
        </div>
      </Card>
    </div>
  )
}
