import { Button } from 'antd'
import { AppleOutlined } from '@ant-design/icons'

interface ApplePayButtonProps {
  amount: number // kopecks
  onToken: (token: string) => void
  disabled?: boolean
  loading?: boolean
}

declare global {
  interface Window {
    ApplePaySession?: {
      new (version: number, request: unknown): ApplePaySessionInstance
      canMakePayments(): boolean
      STATUS_SUCCESS: number
      STATUS_FAILURE: number
    }
  }
}

interface ApplePaySessionInstance {
  begin(): void
  completeMerchantValidation(merchantSession: unknown): void
  completePayment(result: { status: number }): void
  onvalidatemerchant: ((event: { validationURL: string }) => void) | null
  onpaymentauthorized: ((event: { payment: { token: { paymentData: unknown } } }) => void) | null
  oncancel: (() => void) | null
}

export default function ApplePayButton({ amount, onToken, disabled, loading }: ApplePayButtonProps) {
  if (!window.ApplePaySession || !window.ApplePaySession.canMakePayments()) return null

  const handleClick = () => {
    if (!window.ApplePaySession) return

    const request = {
      countryCode: 'RU',
      currencyCode: 'RUB',
      supportedNetworks: ['visa', 'masterCard', 'mir'],
      merchantCapabilities: ['supports3DS'],
      total: {
        label: 'RelaxHub',
        amount: (amount / 100).toFixed(2),
      },
    }

    const session = new window.ApplePaySession(3, request)

    session.onvalidatemerchant = async (event: { validationURL: string }) => {
      try {
        const resp = await fetch('/api/v1/apple-pay/validate-merchant', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ validation_url: event.validationURL }),
        })
        const merchantSession = await resp.json()
        session.completeMerchantValidation(merchantSession.data)
      } catch {
        session.completePayment({ status: window.ApplePaySession!.STATUS_FAILURE })
      }
    }

    session.onpaymentauthorized = (event: { payment: { token: { paymentData: unknown } } }) => {
      const token = JSON.stringify(event.payment.token.paymentData)
      onToken(token)
      session.completePayment({ status: window.ApplePaySession!.STATUS_SUCCESS })
    }

    session.oncancel = () => {
      // User cancelled - no action needed
    }

    session.begin()
  }

  return (
    <Button
      icon={<AppleOutlined />}
      onClick={handleClick}
      disabled={disabled}
      loading={loading}
      size="large"
      style={{
        background: '#000',
        color: '#fff',
        borderColor: '#000',
        borderRadius: 8,
      }}
    >
      Apple Pay
    </Button>
  )
}
