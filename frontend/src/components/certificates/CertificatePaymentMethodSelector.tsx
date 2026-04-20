import type { ReactNode } from 'react'
import { AppleOutlined, CreditCardOutlined, QrcodeOutlined } from '@ant-design/icons'
import { Button } from 'antd'
import ApplePayButton from '@/components/ApplePayButton'
import GooglePayButton from '@/components/GooglePayButton'

export type CertificatePaymentMethod = 'card' | 'sbp' | 'apple_pay' | 'google_pay'

interface CertificatePaymentMethodSelectorProps {
  amount: number
  value: CertificatePaymentMethod
  disabled?: boolean
  loading?: boolean
  onChange: (value: CertificatePaymentMethod) => void
  onSubmit: () => void
  onToken: (token: string) => void
}

const PAYMENT_METHODS: Array<{
  value: CertificatePaymentMethod
  label: string
  icon?: ReactNode
  description: string
}> = [
  {
    value: 'card',
    label: 'Карта',
    icon: <CreditCardOutlined />,
    description: 'Стандартная оплата с переходом на защищённую страницу банка.',
  },
  {
    value: 'sbp',
    label: 'СБП',
    icon: <QrcodeOutlined />,
    description: 'Быстрый платёж через банковское приложение и моментальное подтверждение.',
  },
  {
    value: 'apple_pay',
    label: 'Apple Pay',
    icon: <AppleOutlined />,
    description: 'Оплата в один жест на поддерживаемом устройстве Apple.',
  },
  {
    value: 'google_pay',
    label: 'Google Pay',
    description: 'Быстрый checkout через сохранённую карту Google Pay.',
  },
]

export const CERTIFICATE_PAYMENT_LABELS: Record<CertificatePaymentMethod, string> = {
  card: 'Банковская карта',
  sbp: 'СБП',
  apple_pay: 'Apple Pay',
  google_pay: 'Google Pay',
}

export default function CertificatePaymentMethodSelector({
  amount,
  value,
  disabled,
  loading,
  onChange,
  onSubmit,
  onToken,
}: CertificatePaymentMethodSelectorProps) {
  const selectedMethod = PAYMENT_METHODS.find((method) => method.value === value) ?? PAYMENT_METHODS[0]!
  const walletDisabled = disabled || amount < 10000

  return (
    <div className="bani-certificates-payment-selector">
      <div className="bani-certificates-payment-selector__grid">
        {PAYMENT_METHODS.map((method) => (
          <button
            key={method.value}
            type="button"
            className={`bani-certificates-payment-selector__option${method.value === value ? ' bani-certificates-payment-selector__option--active' : ''}`}
            onClick={() => onChange(method.value)}
            disabled={disabled}
          >
            <span className="bani-certificates-payment-selector__option-header">
              <span className="bani-certificates-payment-selector__option-label">
                {method.icon}
                {method.label}
              </span>
            </span>
            <span className="bani-certificates-payment-selector__option-description">{method.description}</span>
          </button>
        ))}
      </div>

      <div className="bani-certificates-payment-selector__footer">
        <div className="bani-certificates-payment-selector__note">
          {selectedMethod.description}
        </div>

        {(value === 'apple_pay' || value === 'google_pay') ? (
          <div className="bani-certificates-payment-selector__wallet">
            {value === 'apple_pay' ? (
              <ApplePayButton
                amount={amount}
                onToken={onToken}
                disabled={walletDisabled}
                loading={loading}
              />
            ) : (
              <GooglePayButton
                amount={amount}
                onToken={onToken}
                disabled={walletDisabled}
                loading={loading}
              />
            )}
          </div>
        ) : (
          <Button
            type="primary"
            size="large"
            className="bani-certificates-payment-selector__cta"
            loading={loading}
            disabled={disabled}
            onClick={onSubmit}
          >
            Перейти к оплате
          </Button>
        )}
      </div>
    </div>
  )
}
