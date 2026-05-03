import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  CreditCardOutlined,
  GiftOutlined,
  MailOutlined,
  MessageOutlined,
} from '@/components/design/icons'
import { formatPrice } from '@/lib/format'
import { CERTIFICATE_PAYMENT_LABELS, type CertificatePaymentMethod } from '@/lib/certificate-payment'
import { PLATFORM_NAME } from '@/content/support'

interface CertificateGiftSummaryProps {
  amount: number
  purchaserEmail?: string
  recipientEmail?: string
  recipientName?: string
  message?: string
  paymentMethod: CertificatePaymentMethod
}

export default function CertificateGiftSummary({
  amount,
  purchaserEmail,
  recipientEmail,
  recipientName,
  message,
  paymentMethod,
}: CertificateGiftSummaryProps) {
  const safeAmount = amount > 0 ? amount : 300000
  const recipient = recipientName?.trim() || `Гость ${PLATFORM_NAME}`
  const note =
    message?.trim() ||
    'Подарок на отдых, который можно использовать на бронирование в удобный момент.'
  const deliveryEmail = recipientEmail?.trim() || purchaserEmail?.trim() || 'email после оплаты'

  return (
    <section className="rh-certificate-summary">
      <header className="rh-certificate-summary__heading">
        <div className="rh-certificate-summary__eyebrow">Превью и сводка</div>
        <h3 className="rh-certificate-summary__title">Так будет выглядеть подарок</h3>
        <p className="rh-certificate-summary__caption">
          Превью обновляется вместе с формой. После оплаты этот же макет уйдёт получателю на email.
        </p>
      </header>

      <div className="rh-certificate-summary__card">
        <div className="rh-certificate-summary__card-top">
          <div className="rh-certificate-summary__brand">{PLATFORM_NAME}</div>
          <div className="rh-certificate-summary__chip">
            <GiftOutlined />
            Подарочный сертификат
          </div>
        </div>

        <div className="rh-certificate-summary__amount">{formatPrice(safeAmount)}</div>
        <div className="rh-certificate-summary__recipient">Для {recipient}</div>

        <div className="rh-certificate-summary__message">
          <MessageOutlined />
          <span>{note}</span>
        </div>

        <div className="rh-certificate-summary__meta">
          <div className="rh-certificate-summary__meta-item">
            <span className="rh-certificate-summary__meta-label">Доставка</span>
            <span className="rh-certificate-summary__meta-value">
              <MailOutlined />
              {deliveryEmail}
            </span>
          </div>
          <div className="rh-certificate-summary__meta-item">
            <span className="rh-certificate-summary__meta-label">Срок</span>
            <span className="rh-certificate-summary__meta-value">365 дней</span>
          </div>
          <div className="rh-certificate-summary__meta-item">
            <span className="rh-certificate-summary__meta-label">Оплата</span>
            <span className="rh-certificate-summary__meta-value">
              <CreditCardOutlined />
              {CERTIFICATE_PAYMENT_LABELS[paymentMethod]}
            </span>
          </div>
        </div>
      </div>

      <ul className="rh-certificate-summary__trust">
        <li className="rh-certificate-summary__trust-item">
          <CheckCircleOutlined />
          <span>Остаток не сгорает: используйте сертификат сразу или частями.</span>
        </li>
        <li className="rh-certificate-summary__trust-item">
          <MailOutlined />
          <span>Письмо с кодом отправим автоматически после успешной оплаты.</span>
        </li>
        <li className="rh-certificate-summary__trust-item">
          <ClockCircleOutlined />
          <span>Если платёж в обработке — заказ сохранится и будет доступен по ссылке.</span>
        </li>
      </ul>
    </section>
  )
}
