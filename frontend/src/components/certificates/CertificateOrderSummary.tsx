import { CheckCircleOutlined, ClockCircleOutlined, MailOutlined } from '@ant-design/icons'
import { formatPrice } from '@/lib/format'
import { CERTIFICATE_PAYMENT_LABELS, type CertificatePaymentMethod } from '@/lib/certificate-payment'

interface CertificateOrderSummaryProps {
  amount: number
  purchaserEmail?: string
  recipientEmail?: string
  paymentMethod: CertificatePaymentMethod
}

export default function CertificateOrderSummary({
  amount,
  purchaserEmail,
  recipientEmail,
  paymentMethod,
}: CertificateOrderSummaryProps) {
  const safeAmount = amount > 0 ? amount : 300000
  const deliveryEmail = recipientEmail?.trim() || purchaserEmail?.trim() || 'email после оплаты'

  return (
    <section className="bani-certificate-summary">
      <div className="bani-certificate-summary__eyebrow">Сводка заказа</div>
      <div className="bani-certificate-summary__amount">{formatPrice(safeAmount)}</div>
      <div className="bani-certificate-summary__caption">
        Сумма фиксируется в заказе и полностью идёт на будущие бронирования.
      </div>

      <div className="bani-certificate-summary__list">
        <div className="bani-certificate-summary__row">
          <span>Оплата</span>
          <strong>{CERTIFICATE_PAYMENT_LABELS[paymentMethod]}</strong>
        </div>
        <div className="bani-certificate-summary__row">
          <span>Email доставки</span>
          <strong>{deliveryEmail}</strong>
        </div>
        <div className="bani-certificate-summary__row">
          <span>Срок</span>
          <strong>365 дней</strong>
        </div>
      </div>

      <div className="bani-certificate-summary__trust">
        <div className="bani-certificate-summary__trust-item">
          <CheckCircleOutlined />
          <span>Остаток не сгорает после частичного использования.</span>
        </div>
        <div className="bani-certificate-summary__trust-item">
          <MailOutlined />
          <span>Подтверждение и код придут только после успешной оплаты.</span>
        </div>
        <div className="bani-certificate-summary__trust-item">
          <ClockCircleOutlined />
          <span>Если платёж в обработке, заказ сохранится и будет доступен по ссылке.</span>
        </div>
      </div>
    </section>
  )
}
