import { startTransition, useDeferredValue, useMemo, useState } from 'react'
import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  GiftOutlined,
  MailOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons'
import { App, Button, Collapse, Form, Input, InputNumber, Spin, Tag } from 'antd'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import {
  useGetCertificatesOrdersId,
  usePostCertificatesOrders,
  usePostCertificatesOrdersIdPay,
} from '@/api/generated/certificates/certificates'
import type { InternalHandlerCertificateOrderResponse } from '@/api/generated/model/internalHandlerCertificateOrderResponse'
import type { InternalHandlerCreateCertificateOrderRequest } from '@/api/generated/model/internalHandlerCreateCertificateOrderRequest'
import type { InternalHandlerCertificateOrderPaymentRequest } from '@/api/generated/model/internalHandlerCertificateOrderPaymentRequest'
import CertificateGiftPreview from '@/components/certificates/CertificateGiftPreview'
import CertificateOrderSummary from '@/components/certificates/CertificateOrderSummary'
import CertificatePaymentMethodSelector, {
  type CertificatePaymentMethod,
} from '@/components/certificates/CertificatePaymentMethodSelector'
import { formatDateTime, formatPrice } from '@/lib/format'
import { useAuthStore } from '@/stores/auth'

const PRESET_AMOUNTS = [100000, 200000, 300000, 500000]

const HERO_META = [
  {
    label: 'Срок действия',
    value: '365 дней с сохранением остатка',
  },
  {
    label: 'Доставка',
    value: 'На email сразу после подтверждения оплаты',
  },
  {
    label: 'Использование',
    value: 'На любые бронирования внутри сервиса',
  },
  {
    label: 'Поддержка',
    value: 'Поможем с оплатой и активацией без лишних шагов',
  },
]

const HOW_IT_WORKS = [
  {
    title: 'Соберите подарок за минуту',
    description: 'Укажите сумму, контакт покупателя и, если нужно, данные получателя. Отдельного режима не требуется.',
  },
  {
    title: 'Оплатите удобным способом',
    description: 'Карта, СБП, Apple Pay и Google Pay ведут в единый order flow без ложного выпуска сертификата до оплаты.',
  },
  {
    title: 'Получите код после подтверждения',
    description: 'Когда платёж завершён, мы выпускаем сертификат, отправляем письмо и показываем success-state в этой же странице.',
  },
]

const FAQ_ITEMS = [
  {
    key: 'validity',
    label: 'Когда появляется код сертификата?',
    children: 'Код создаётся только после подтверждённой оплаты. До этого в системе существует только заказ со статусом оплаты.',
  },
  {
    key: 'balance',
    label: 'Что будет с остатком после частичного использования?',
    children: 'Неиспользованный баланс сохраняется до окончания срока действия и может быть списан на следующих бронированиях.',
  },
  {
    key: 'delivery',
    label: 'Можно ли отправить подарок другому человеку?',
    children: 'Да. Если заполнить поля получателя, письмо с сертификатом уйдёт на указанный email, а имя попадёт в превью подарка.',
  },
]

type CheckoutFormValues = {
  amount: number
  purchaser_email: string
  recipient_email?: string
  recipient_name?: string
  message?: string
}

function buildCreateOrderPayload(values: CheckoutFormValues): InternalHandlerCreateCertificateOrderRequest {
  return {
    amount: Math.round(Number(values.amount) * 100),
    purchaser_email: values.purchaser_email.trim(),
    recipient_email: values.recipient_email?.trim() || undefined,
    recipient_name: values.recipient_name?.trim() || undefined,
    message: values.message?.trim() || undefined,
  }
}

function buildPaymentPayload(
  paymentMethod: CertificatePaymentMethod,
  paymentToken?: string,
): InternalHandlerCertificateOrderPaymentRequest {
  return {
    payment_method: paymentMethod,
    payment_token: paymentToken || undefined,
  }
}

function OrderStatePanel({
  order,
  loading,
  isAuthenticated,
}: {
  order?: InternalHandlerCertificateOrderResponse
  loading: boolean
  isAuthenticated: boolean
}) {
  if (loading) {
    return (
      <section className="bani-certificates-status bani-certificates-status--processing">
        <Spin size="large" />
        <div className="bani-certificates-status__copy">
          <h2>Проверяем статус заказа</h2>
          <p>Забираем актуальное состояние оплаты и готовность сертификата.</p>
        </div>
      </section>
    )
  }

  if (!order) {
    return (
      <section className="bani-certificates-status bani-certificates-status--warning">
        <div className="bani-certificates-status__copy">
          <h2>Заказ не найден</h2>
          <p>Проверьте ссылку из письма или начните новый заказ, если оплата ещё не создавалась.</p>
        </div>
        <Link to="/certificates">
          <Button type="primary" size="large">Создать новый заказ</Button>
        </Link>
      </section>
    )
  }

  if (order.status === 'paid') {
    return (
      <section className="bani-certificates-status bani-certificates-status--success">
        <div className="bani-certificates-status__badge">
          <CheckCircleOutlined />
          Оплата подтверждена
        </div>
        <div className="bani-certificates-status__copy">
          <h2>Сертификат оплачен</h2>
          <p>Мы отправили подтверждение на email и выпустили код для дальнейших бронирований.</p>
        </div>

        <div className="bani-certificates-status__details">
          <div className="bani-certificates-status__detail-card">
            <span className="bani-certificates-status__detail-label">Код сертификата</span>
            <strong>{order.certificate?.code ?? '—'}</strong>
          </div>
          <div className="bani-certificates-status__detail-card">
            <span className="bani-certificates-status__detail-label">Номинал</span>
            <strong>{formatPrice(order.certificate?.amount ?? order.amount ?? 0)}</strong>
          </div>
          <div className="bani-certificates-status__detail-card">
            <span className="bani-certificates-status__detail-label">Действителен до</span>
            <strong>{order.certificate?.valid_until ? formatDateTime(order.certificate.valid_until, 'DD.MM.YYYY') : '—'}</strong>
          </div>
        </div>

        <div className="bani-certificates-status__actions">
          {isAuthenticated && (
            <Link to="/client/certificates">
              <Button size="large">Открыть кабинет</Button>
            </Link>
          )}
          <Link to="/certificates">
            <Button type="primary" size="large">Купить ещё сертификат</Button>
          </Link>
        </div>
      </section>
    )
  }

  if (order.status === 'pending_payment') {
    return (
      <section className="bani-certificates-status bani-certificates-status--processing">
        <div className="bani-certificates-status__badge">
          <ClockCircleOutlined />
          Статус заказа обновляется автоматически
        </div>
        <div className="bani-certificates-status__copy">
          <h2>Платёж обрабатывается</h2>
          <p>Подтверждаем оплату и выпуск сертификата. Как только статус сменится на `paid`, здесь появится код и success-state.</p>
        </div>
      </section>
    )
  }

  return (
    <section className="bani-certificates-status bani-certificates-status--warning">
      <div className="bani-certificates-status__copy">
        <h2>Оплата не завершена</h2>
        <p>Заказ сохранился. Можно вернуться к checkout и повторить платёж другим способом.</p>
      </div>
      <Link to="/certificates">
        <Button type="primary" size="large">Вернуться к checkout</Button>
      </Link>
    </section>
  )
}

export default function CertificatePurchase() {
  const { message } = App.useApp()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const [form] = Form.useForm<CheckoutFormValues>()
  const [paymentMethod, setPaymentMethod] = useState<CertificatePaymentMethod>('card')

  const createOrderMutation = usePostCertificatesOrders()
  const payOrderMutation = usePostCertificatesOrdersIdPay()

  const orderId = searchParams.get('order_id') ?? ''
  const orderQuery = useGetCertificatesOrdersId(orderId, {
    query: {
      enabled: Boolean(orderId),
      refetchInterval: orderId ? 3000 : false,
      retry: false,
    },
  })

  const amountValue = Form.useWatch('amount', form) ?? 3000
  const deferredAmountValue = useDeferredValue(amountValue)
  const purchaserEmail = Form.useWatch('purchaser_email', form) ?? ''
  const recipientEmail = Form.useWatch('recipient_email', form) ?? ''
  const recipientName = Form.useWatch('recipient_name', form) ?? ''
  const messageValue = Form.useWatch('message', form) ?? ''

  const order = orderQuery.data?.data
  const amountKopecks = Math.max(0, Math.round(Number(deferredAmountValue || 0) * 100))
  const isSubmitting = createOrderMutation.isPending || payOrderMutation.isPending
  const isStatusMode = Boolean(orderId)

  const paymentHint = useMemo(() => {
    switch (paymentMethod) {
      case 'sbp':
        return 'Откроем страницу оплаты СБП и после подтверждения вернём вас в этот order flow.'
      case 'apple_pay':
        return 'Apple Pay создаёт заказ и сразу передаёт токен оплаты без дополнительной формы.'
      case 'google_pay':
        return 'Google Pay работает через токенизированную оплату и затем возвращает вас в статус заказа.'
      default:
        return 'Классический redirect checkout с защищённой платёжной страницей.'
    }
  }, [paymentMethod])

  const handlePresetAmount = (presetAmount: number) => {
    startTransition(() => {
      form.setFieldValue('amount', presetAmount / 100)
    })
  }

  const handleCheckout = async (selectedMethod: CertificatePaymentMethod, paymentToken?: string) => {
    try {
      const values = await form.validateFields()
      const orderResponse = await createOrderMutation.mutateAsync({
        data: buildCreateOrderPayload(values),
      })
      const createdOrder = orderResponse.data

      if (!createdOrder?.id) {
        throw new Error('missing_order_id')
      }

      const payResponse = await payOrderMutation.mutateAsync({
        id: createdOrder.id,
        data: buildPaymentPayload(selectedMethod, paymentToken),
      })

      const nextUrl = `/certificates?order_id=${createdOrder.id}`
      startTransition(() => {
        navigate(nextUrl, { replace: true })
      })

      if (payResponse.data?.confirmation_url && import.meta.env.MODE !== 'test') {
        window.location.assign(payResponse.data.confirmation_url)
        return
      }

      message.success('Заказ создан. Статус оплаты будет доступен на этой странице.')
    } catch (error) {
      if (error instanceof Error && error.message === 'missing_order_id') {
        message.error('Не удалось получить идентификатор заказа')
        return
      }

      message.error('Не удалось создать заказ или инициировать оплату')
    }
  }

  return (
    <div className="bani-stack bani-certificates-page">
      <div className="bani-certificates-page__back">
        <Link to={isAuthenticated ? '/client/certificates' : '/'}>
          <Button icon={<ArrowLeftOutlined />} className="bani-certificates-page__back-button">
            {isAuthenticated ? 'К сертификатам' : 'На главную'}
          </Button>
        </Link>
      </div>

      <section className="bani-hero-panel bani-hero-panel--dark bani-certificates-hero">
        <div className="bani-hero-panel__eyebrow">Gift checkout</div>
        <h1 className="bani-hero-panel__title">Подарочный сертификат BANI</h1>
        <div className="bani-hero-panel__description">
          Оплатите подарок один раз, а использовать его можно позже и частями. Сертификат работает как премиальный запас времени на отдых, а не как одноразовый промокод.
        </div>

        <div className="bani-certificates-hero__trust">
          <div className="bani-certificates-hero__trust-item">
            <SafetyCertificateOutlined />
            <span>Код выпускается только после статуса `paid`.</span>
          </div>
          <div className="bani-certificates-hero__trust-item">
            <MailOutlined />
            <span>Подтверждение и инструкция приходят на email без ручной переписки с поддержкой.</span>
          </div>
          <div className="bani-certificates-hero__trust-item">
            <GiftOutlined />
            <span>Сценарий “себе” и “в подарок” живёт в одной форме и переключается заполнением полей.</span>
          </div>
        </div>

        <div className="bani-hero-panel__meta">
          {HERO_META.map((item) => (
            <div key={item.label} className="bani-hero-panel__meta-item">
              <span className="bani-hero-panel__meta-label">{item.label}</span>
              <div className="bani-hero-panel__meta-value">{item.value}</div>
            </div>
          ))}
        </div>
      </section>

      {isStatusMode ? (
        <OrderStatePanel
          order={order}
          loading={orderQuery.isLoading}
          isAuthenticated={isAuthenticated}
        />
      ) : (
        <div className="bani-grid bani-grid--content-aside bani-certificates-layout">
          <section className="bani-stack">
            <section className="bani-section-card bani-certificates-workspace">
              <div className="bani-section-card__surface bani-certificates-workspace__surface">
                <div className="bani-toolbar">
                  <div>
                    <div className="bani-certificates-section-eyebrow">Checkout</div>
                    <h2 className="bani-section-card__title">Соберите сертификат под конкретный подарок</h2>
                    <div className="bani-section-card__description">
                      Сначала собираем заказ, затем создаём оплату выбранным методом. Платёж и выпуск сертификата больше не смешаны в один ложный шаг.
                    </div>
                  </div>
                  <Tag color="cyan">Premium service flow</Tag>
                </div>

                <Form
                  form={form}
                  layout="vertical"
                  initialValues={{
                    amount: 3000,
                  }}
                  className="bani-certificates-form"
                >
                  <Form.Item
                    name="amount"
                    label="Сумма (в рублях)"
                    rules={[
                      { required: true, message: 'Укажите сумму' },
                      { type: 'number', min: 100, message: 'Минимальная сумма — 100 ₽' },
                      { type: 'number', max: 100000, message: 'Максимальная сумма — 100 000 ₽' },
                    ]}
                  >
                    <InputNumber
                      style={{ width: '100%' }}
                      min={100}
                      max={100000}
                      placeholder="Введите сумму"
                      controls={false}
                      addonAfter="₽"
                    />
                  </Form.Item>

                  <div className="bani-certificates-amount-pills">
                    {PRESET_AMOUNTS.map((preset) => {
                      const active = amountKopecks === preset
                      return (
                        <button
                          key={preset}
                          type="button"
                          className={`bani-certificates-amount-pill${active ? ' bani-certificates-amount-pill--active' : ''}`}
                          onClick={() => handlePresetAmount(preset)}
                        >
                          {formatPrice(preset)}
                        </button>
                      )
                    })}
                  </div>

                  <div className="bani-grid bani-grid--two bani-certificates-form__grid">
                    <Form.Item
                      name="purchaser_email"
                      label="Ваш email"
                      rules={[
                        { required: true, message: 'Укажите email' },
                        { type: 'email', message: 'Введите корректный email' },
                      ]}
                    >
                      <Input placeholder="your@email.com" />
                    </Form.Item>

                    <Form.Item
                      name="recipient_email"
                      label="Email получателя"
                      rules={[{ type: 'email', message: 'Введите корректный email' }]}
                    >
                      <Input placeholder="recipient@email.com" />
                    </Form.Item>
                  </div>

                  <div className="bani-grid bani-grid--two bani-certificates-form__grid">
                    <Form.Item name="recipient_name" label="Имя получателя">
                      <Input placeholder="Имя получателя" />
                    </Form.Item>

                    <div className="bani-certificates-inline-note">
                      Если поля получателя пустые, сертификат остаётся на вашем email и работает как личный баланс для будущих бронирований.
                    </div>
                  </div>

                  <Form.Item name="message" label="Сообщение">
                    <Input.TextArea
                      rows={4}
                      maxLength={500}
                      showCount
                      placeholder="Поздравляю! Желаю приятного отдыха!"
                    />
                  </Form.Item>

                  <div className="bani-certificates-payment-block">
                    <div className="bani-certificates-payment-block__header">
                      <div>
                        <div className="bani-certificates-section-eyebrow">Payment</div>
                        <h3>Способ оплаты</h3>
                      </div>
                      <span>{paymentHint}</span>
                    </div>

                    <CertificatePaymentMethodSelector
                      amount={amountKopecks}
                      value={paymentMethod}
                      disabled={isSubmitting}
                      loading={isSubmitting}
                      onChange={setPaymentMethod}
                      onSubmit={() => {
                        void handleCheckout(paymentMethod)
                      }}
                      onToken={(token) => {
                        void handleCheckout(paymentMethod, token)
                      }}
                    />
                  </div>
                </Form>
              </div>
            </section>

            <div className="bani-grid bani-grid--two bani-certificates-detail-grid">
              <section className="bani-section-card">
                <div className="bani-section-card__surface">
                  <div className="bani-certificates-section-eyebrow">How it works</div>
                  <h3 className="bani-section-card__title">Процесс покупки без серых зон</h3>
                  <div className="bani-feature-list">
                    {HOW_IT_WORKS.map((item, index) => (
                      <div key={item.title} className="bani-feature-item">
                        <div className="bani-feature-item__icon">{index + 1}</div>
                        <div className="bani-feature-item__copy">
                          <div className="bani-feature-item__title">{item.title}</div>
                          <div className="bani-feature-item__description">{item.description}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </section>

              <section className="bani-section-card">
                <div className="bani-section-card__surface">
                  <div className="bani-certificates-section-eyebrow">FAQ</div>
                  <h3 className="bani-section-card__title">Важные детали перед оплатой</h3>
                  <Collapse
                    ghost
                    items={FAQ_ITEMS.map((item) => ({
                      key: item.key,
                      label: item.label,
                      children: item.children,
                    }))}
                  />
                </div>
              </section>
            </div>
          </section>

          <aside className="bani-stack bani-certificates-sidebar">
            <CertificateOrderSummary
              amount={amountKopecks}
              purchaserEmail={purchaserEmail}
              recipientEmail={recipientEmail}
              paymentMethod={paymentMethod}
            />
            <CertificateGiftPreview
              amount={amountKopecks}
              recipientName={recipientName}
              recipientEmail={recipientEmail}
              message={messageValue}
            />
          </aside>
        </div>
      )}
    </div>
  )
}
