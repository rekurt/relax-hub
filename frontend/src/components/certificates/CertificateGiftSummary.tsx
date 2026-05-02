import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  CreditCardOutlined,
  GiftOutlined,
  MailOutlined,
  MessageOutlined,
} from '@ant-design/icons'
import { formatPrice } from '@/lib/format'
import { CERTIFICATE_PAYMENT_LABELS, type CertificatePaymentMethod } from '@/lib/certificate-payment'

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
  const recipient = recipientName?.trim() || 'Гость BANI'
  const note =
    message?.trim() ||
    'Подарок на отдых, который можно использовать на бронирование в удобный момент.'
  const deliveryEmail = recipientEmail?.trim() || purchaserEmail?.trim() || 'email после оплаты'

  return (
    <section className="bani-certificate-summary">
      <header className="bani-certificate-summary__heading">
        <div className="bani-certificate-summary__eyebrow">Превью и сводка</div>
        <h3 className="bani-certificate-summary__title">Так будет выглядеть подарок</h3>
        <p className="bani-certificate-summary__caption">
          Превью обновляется вместе с формой. После оплаты этот же макет уйдёт получателю на email.
        </p>
      </header>

      <div className="bani-certificate-summary__card">
        <div className="bani-certificate-summary__card-top">
          <div className="bani-certificate-summary__brand">BANI</div>
          <div className="bani-certificate-summary__chip">
            <GiftOutlined />
            Gift Certificate
          </div>
        </div>

        <div className="bani-certificate-summary__amount">{formatPrice(safeAmount)}</div>
        <div className="bani-certificate-summary__recipient">Для {recipient}</div>

        <div className="bani-certificate-summary__message">
          <MessageOutlined />
          <span>{note}</span>
        </div>

        <div className="bani-certificate-summary__meta">
          <div className="bani-certificate-summary__meta-item">
            <span className="bani-certificate-summary__meta-label">Доставка</span>
            <span className="bani-certificate-summary__meta-value">
              <MailOutlined />
              {deliveryEmail}
            </span>
          </div>
          <div className="bani-certificate-summary__meta-item">
            <span className="bani-certificate-summary__meta-label">Срок</span>
            <span className="bani-certificate-summary__meta-value">365 дней</span>
          </div>
          <div className="bani-certificate-summary__meta-item">
            <span className="bani-certificate-summary__meta-label">Оплата</span>
            <span className="bani-certificate-summary__meta-value">
              <CreditCardOutlined />
              {CERTIFICATE_PAYMENT_LABELS[paymentMethod]}
            </span>
          </div>
        </div>
      </div>

      <ul className="bani-certificate-summary__trust">
        <li className="bani-certificate-summary__trust-item">
          <CheckCircleOutlined />
          <span>Остаток не сгорает: используйте сертификат сразу или частями.</span>
        </li>
        <li className="bani-certificate-summary__trust-item">
          <MailOutlined />
          <span>Письмо с кодом отправим автоматически после успешной оплаты.</span>
        </li>
        <li className="bani-certificate-summary__trust-item">
          <ClockCircleOutlined />
          <span>Если платёж в обработке — заказ сохранится и будет доступен по ссылке.</span>
        </li>
      </ul>
    </section>
  )
}
