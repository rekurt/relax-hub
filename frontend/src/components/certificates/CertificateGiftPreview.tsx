import { GiftOutlined, MailOutlined, MessageOutlined } from '@ant-design/icons'
import { formatPrice } from '@/lib/format'

interface CertificateGiftPreviewProps {
  amount: number
  recipientName?: string
  recipientEmail?: string
  message?: string
}

export default function CertificateGiftPreview({
  amount,
  recipientName,
  recipientEmail,
  message,
}: CertificateGiftPreviewProps) {
  const safeAmount = amount > 0 ? amount : 300000
  const recipient = recipientName?.trim() || 'Гость BANI'
  const note = message?.trim() || 'Подарок на отдых, который можно использовать на бронирование в удобный момент.'

  return (
    <section className="bani-certificate-preview">
      <div className="bani-certificate-preview__header">
        <div>
          <div className="bani-certificate-preview__eyebrow">Превью сертификата</div>
          <h3 className="bani-certificate-preview__title">Как будет выглядеть подарок до оплаты</h3>
        </div>
      </div>

      <div className="bani-certificate-preview__card">
        <div className="bani-certificate-preview__card-top">
          <div className="bani-certificate-preview__brand">BANI</div>
          <div className="bani-certificate-preview__chip">
            <GiftOutlined />
            Gift Certificate
          </div>
        </div>

        <div className="bani-certificate-preview__amount">{formatPrice(safeAmount)}</div>
        <div className="bani-certificate-preview__recipient">Для {recipient}</div>

        <div className="bani-certificate-preview__message">
          <MessageOutlined />
          <span>{note}</span>
        </div>

        <div className="bani-certificate-preview__meta">
          <div className="bani-certificate-preview__meta-item">
            <span className="bani-certificate-preview__meta-label">Доставка на email</span>
            <span className="bani-certificate-preview__meta-value">
              <MailOutlined />
              {recipientEmail?.trim() || 'На email после оплаты'}
            </span>
          </div>
          <div className="bani-certificate-preview__meta-item">
            <span className="bani-certificate-preview__meta-label">Действует</span>
            <span className="bani-certificate-preview__meta-value">365 дней</span>
          </div>
        </div>
      </div>
    </section>
  )
}
