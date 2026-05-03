import { useEffect, useState, useRef, useCallback } from 'react'
import { Button } from 'antd'
import { GoogleOutlined } from '@ant-design/icons'

interface GooglePayButtonProps {
  amount: number // kopecks
  onToken: (token: string) => void
  disabled?: boolean
  loading?: boolean
}

interface GooglePayClient {
  isReadyToPay(request: unknown): Promise<{ result: boolean }>
  loadPaymentData(request: unknown): Promise<{
    paymentMethodData: { tokenizationData: { token: string } }
  }>
}

declare global {
  interface Window {
    google?: {
      payments?: {
        api?: {
          PaymentsClient: new (config: { environment: string }) => GooglePayClient
        }
      }
    }
  }
}

const GOOGLE_PAY_SCRIPT_URL = 'https://pay.google.com/gp/p/js/pay.js'

const baseRequest = {
  apiVersion: 2,
  apiVersionMinor: 0,
}

const tokenizationSpecification = {
  type: 'PAYMENT_GATEWAY',
  parameters: {
    gateway: 'yookassa',
    gatewayMerchantId: import.meta.env.VITE_GPAY_MERCHANT_ID ?? '',
  },
}

const allowedCardNetworks = ['MASTERCARD', 'VISA', 'MIR']
const allowedCardAuthMethods = ['PAN_ONLY', 'CRYPTOGRAM_3DS']

const baseCardPaymentMethod = {
  type: 'CARD',
  parameters: {
    allowedAuthMethods: allowedCardAuthMethods,
    allowedCardNetworks: allowedCardNetworks,
  },
}

export default function GooglePayButton({ amount, onToken, disabled, loading }: GooglePayButtonProps) {
  const [available, setAvailable] = useState(false)
  const clientRef = useRef<GooglePayClient | null>(null)

  const initGooglePay = useCallback(async () => {
    if (!window.google?.payments?.api?.PaymentsClient) return

    const client = new window.google.payments.api.PaymentsClient({
      environment: import.meta.env.VITE_GPAY_ENVIRONMENT ?? 'TEST',
    })
    clientRef.current = client

    try {
      const response = await client.isReadyToPay({
        ...baseRequest,
        allowedPaymentMethods: [baseCardPaymentMethod],
      })
      setAvailable(response.result)
    } catch {
      setAvailable(false)
    }
  }, [])

  useEffect(() => {
    if (window.google?.payments?.api?.PaymentsClient) {
      initGooglePay() // eslint-disable-line react-hooks/set-state-in-effect -- async script load sets state after API check
      return
    }

    const script = document.createElement('script')
    script.src = GOOGLE_PAY_SCRIPT_URL
    script.async = true
    script.onload = () => initGooglePay()
    document.head.appendChild(script)

    return () => {
      if (script.parentNode) {
        script.parentNode.removeChild(script)
      }
    }
  }, [initGooglePay])

  const handleClick = async () => {
    if (!clientRef.current) return

    const paymentDataRequest = {
      ...baseRequest,
      allowedPaymentMethods: [
        {
          ...baseCardPaymentMethod,
          tokenizationSpecification,
        },
      ],
      transactionInfo: {
        totalPriceStatus: 'FINAL',
        totalPrice: (amount / 100).toFixed(2),
        currencyCode: 'RUB',
        countryCode: 'RU',
      },
      merchantInfo: {
        merchantName: 'RelaxHub',
      },
    }

    try {
      const paymentData = await clientRef.current.loadPaymentData(paymentDataRequest)
      const token = paymentData.paymentMethodData.tokenizationData.token
      onToken(token)
    } catch {
      // User cancelled or error - no action needed
    }
  }

  if (!available) return null

  return (
    <Button
      icon={<GoogleOutlined />}
      onClick={handleClick}
      disabled={disabled}
      loading={loading}
      size="large"
      style={{
        background: '#fff',
        color: '#3c4043',
        borderColor: '#dadce0',
        borderRadius: 20,
      }}
    >
      Google Pay
    </Button>
  )
}
