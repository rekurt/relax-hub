import { useMemo, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Button, Card, Collapse, Row, Col, Typography, Alert, Spin } from 'antd'
import {
  AppstoreOutlined,
  CreditCardOutlined,
  MessageOutlined,
  PhoneOutlined,
  QuestionCircleOutlined,
  SafetyOutlined,
} from '@ant-design/icons'
import { axiosInstance } from '@/api/axios-instance'
import PageHeader from '@/components/PageHeader'
import {
  HELP_BOOKING_JOURNEY,
  HELP_CATEGORY_ORDER,
  HELP_ESCALATION_CASES,
  HELP_FALLBACK_FAQ,
  HELP_SCENARIO_CARDS,
  type HelpFAQItem,
} from '@/content/help'
import { PLATFORM_CONTACTS, PLATFORM_NAME } from '@/content/support'

const { Text } = Typography

const CATEGORY_LABELS: Record<string, { label: string; color: string; icon: ReactNode }> = {
  booking: { label: 'Бронирование', color: 'blue', icon: <AppstoreOutlined /> },
  payment: { label: 'Оплата', color: 'gold', icon: <CreditCardOutlined /> },
  cancellation: { label: 'Отмена', color: 'orange', icon: <QuestionCircleOutlined /> },
  wallet: { label: 'Кошелёк', color: 'green', icon: <CreditCardOutlined /> },
  account: { label: 'Аккаунт', color: 'purple', icon: <QuestionCircleOutlined /> },
  general: { label: 'Общее', color: 'geekblue', icon: <QuestionCircleOutlined /> },
}

export default function PublicFAQ() {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['/public/faq'],
    queryFn: async () => {
      const response = await axiosInstance.get<{ data?: HelpFAQItem[] }>('/faq')
      return response.data.data ?? []
    },
    retry: false,
  })

  const items = data && data.length > 0 ? data : HELP_FALLBACK_FAQ

  const grouped = useMemo(() => {
    const groups = items.reduce<Record<string, HelpFAQItem[]>>((acc, item) => {
      if (!acc[item.category]) {
        acc[item.category] = []
      }
      acc[item.category]!.push(item)
      return acc
    }, {})
    return HELP_CATEGORY_ORDER
      .filter((category) => groups[category]?.length)
      .map((category) => ({
        category,
        items: groups[category]!,
        meta: CATEGORY_LABELS[category] ?? CATEGORY_LABELS.general!,
      }))
  }, [items])

  if (isLoading) {
    return <Spin size="large" style={{ display: 'block', margin: '96px auto' }} />
  }

  return (
    <div className="bani-stack">
      <PageHeader
        eyebrow="Поддержка"
        title="Чем можем помочь прямо сейчас"
        description={`Это не декоративный FAQ, а рабочий help-раздел по реальным сценариям ${PLATFORM_NAME}: выбор слотов, checkout, оплата, отмена, кошелёк, отзывы и поддержка.`}
      />

      <section className="bani-hero-panel">
        <div className="bani-hero-panel__eyebrow">Help Center</div>
        <h2 className="bani-hero-panel__title">Понятные ответы до брони, во время оплаты и после визита.</h2>
        <div className="bani-hero-panel__description">
          Когда пользователь не понимает, как выбрать слот, оплатить, отменить или оставить отзыв, он не должен искать ответ по всему продукту. Экран собирает самые частые сценарии в одном месте и сразу показывает, куда идти дальше.
        </div>
        <div className="bani-hero-panel__meta">
          {grouped.map(({ category, items: categoryItems, meta }) => {
            return (
              <div key={category} className="bani-hero-panel__meta-item">
                <span className="bani-hero-panel__meta-label">{meta.label}</span>
                <div className="bani-hero-panel__meta-value">{categoryItems.length} ответов</div>
              </div>
            )
          })}
        </div>
      </section>

      {isError && (
        <Alert
          type="info"
          showIcon
          title={`Показываем встроенную базу знаний ${PLATFORM_NAME}`}
          description="Публичный FAQ API временно недоступен, поэтому экран использует встроенные ответы по реальным сценариям продукта: бронирование, оплата, отмена, кошелёк, аккаунт и отзывы."
        />
      )}

      <Row gutter={[16, 16]}>
        {HELP_SCENARIO_CARDS.map((card) => (
          <Col key={card.key} xs={24} md={12}>
            <Card style={{ borderRadius: 24, height: '100%' }}>
              <div className="bani-section-card">
                <div className="bani-shell-footer__eyebrow" style={{ color: 'var(--bani-text-soft)' }}>{card.eyebrow}</div>
                <h2 className="bani-section-card__title">{card.title}</h2>
                <div className="bani-section-card__description">{card.description}</div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      <Row gutter={[16, 16]}>
        {grouped.map(({ category, items: categoryItems, meta }) => {
          return (
            <Col key={category} xs={24} lg={12}>
              <Card
                title={(
                  <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
                    {meta.icon}
                    {meta.label}
                  </span>
                )}
                extra={<Text type="secondary">{categoryItems.length} ответов</Text>}
                style={{ borderRadius: 24, height: '100%' }}
              >
                <Collapse
                  ghost
                  items={categoryItems.map((item) => ({
                    key: item.id,
                    label: item.question,
                    children: (
                      <div className="bani-inline-note">
                        <Text type="secondary">{item.answer}</Text>
                      </div>
                    ),
                  }))}
                />
              </Card>
            </Col>
          )
        })}
      </Row>

      <div className="bani-grid bani-grid--content-aside">
        <Card>
          <div className="bani-section-card">
            <h2 className="bani-section-card__title">Как проходит бронирование</h2>
            <div className="bani-kv">
              {HELP_BOOKING_JOURNEY.map((step) => (
                <div key={step.key} className="bani-kv__row">
                  <div className="bani-kv__label">{step.label}</div>
                  <div className="bani-kv__value">{step.value}</div>
                </div>
              ))}
            </div>
          </div>
        </Card>

        <section className="bani-hero-panel bani-hero-panel--dark">
          <div className="bani-hero-panel__eyebrow">Живой контакт</div>
          <h2 className="bani-hero-panel__title">Если нужен не FAQ, а действие поддержки.</h2>
          <div className="bani-hero-panel__description">
            Когда вопрос касается оплаты, переноса визита или конфликта после посещения, лучше сразу перейти в канал, где проблему можно довести до решения.
          </div>
          <div className="bani-kv">
            <div className="bani-kv__row">
              <div className="bani-kv__label"><MessageOutlined /> Email</div>
              <div className="bani-kv__value">{PLATFORM_CONTACTS.supportEmail}</div>
            </div>
            <div className="bani-kv__row">
              <div className="bani-kv__label"><PhoneOutlined /> Телефон</div>
              <div className="bani-kv__value">{PLATFORM_CONTACTS.supportPhone}</div>
            </div>
            <div className="bani-kv__row">
              <div className="bani-kv__label"><SafetyOutlined /> Часы связи</div>
              <div className="bani-kv__value">{PLATFORM_CONTACTS.supportHours}</div>
            </div>
          </div>
          <Button type="primary" size="large" href="/contacts">
            Все контакты
          </Button>
        </section>
      </div>

      <Card>
        <div className="bani-section-card">
          <h2 className="bani-section-card__title">Когда лучше сразу идти в поддержку</h2>
          <div className="bani-feature-list">
            {HELP_ESCALATION_CASES.map((item) => (
              <div key={item.key} className="bani-feature-item">
                <div className="bani-feature-item__icon"><QuestionCircleOutlined /></div>
                <div className="bani-feature-item__copy">
                  <div className="bani-feature-item__title">{item.title}</div>
                  <div className="bani-feature-item__description">{item.description}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </Card>
    </div>
  )
}
