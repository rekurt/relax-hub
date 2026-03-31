import { useState, useEffect, useCallback } from 'react'
import { Input, Button, Space, App } from 'antd'
import { PhoneOutlined } from '@ant-design/icons'

const OTP_COOLDOWN_SECONDS = 60
const OTP_LENGTH = 6

interface PhoneOTPInputProps {
  onVerified: (phone: string, code: string) => void
  onSendOTP: (phone: string) => Promise<void>
  loading?: boolean
  phoneLabel?: string
}

export default function PhoneOTPInput({ onVerified, onSendOTP, loading, phoneLabel = 'Телефон' }: PhoneOTPInputProps) {
  const { message } = App.useApp()
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [otpSent, setOtpSent] = useState(false)
  const [cooldown, setCooldown] = useState(0)
  const [sending, setSending] = useState(false)

  useEffect(() => {
    if (cooldown <= 0) return
    const timer = setTimeout(() => setCooldown((c) => c - 1), 1000)
    return () => clearTimeout(timer)
  }, [cooldown])

  const handleSendOTP = useCallback(async () => {
    const cleaned = phone.replace(/\s/g, '')
    if (!cleaned || cleaned.length < 10) {
      message.warning('Введите корректный номер телефона')
      return
    }
    setSending(true)
    try {
      await onSendOTP(cleaned)
      setOtpSent(true)
      setCooldown(OTP_COOLDOWN_SECONDS)
      message.success('Код отправлен на ваш телефон')
    } catch {
      message.error('Не удалось отправить код')
    } finally {
      setSending(false)
    }
  }, [phone, onSendOTP, message])

  const handleVerify = useCallback(() => {
    if (code.length !== OTP_LENGTH) {
      message.warning(`Введите ${OTP_LENGTH}-значный код`)
      return
    }
    onVerified(phone.replace(/\s/g, ''), code)
  }, [phone, code, onVerified, message])

  return (
    <Space direction="vertical" style={{ width: '100%' }} size="middle">
      <Space.Compact style={{ width: '100%' }}>
        <Input
          prefix={<PhoneOutlined />}
          placeholder={phoneLabel}
          size="large"
          value={phone}
          onChange={(e) => setPhone(e.target.value)}
          disabled={otpSent && cooldown > 0}
          style={{ flex: 1 }}
        />
        <Button
          size="large"
          onClick={handleSendOTP}
          loading={sending}
          disabled={cooldown > 0 || !phone.trim()}
        >
          {cooldown > 0 ? `${cooldown}с` : otpSent ? 'Отправить снова' : 'Получить код'}
        </Button>
      </Space.Compact>

      {otpSent && (
        <Space.Compact style={{ width: '100%' }}>
          <Input
            placeholder="Введите код из SMS"
            size="large"
            maxLength={OTP_LENGTH}
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
            onPressEnter={handleVerify}
            style={{ flex: 1 }}
          />
          <Button
            type="primary"
            size="large"
            onClick={handleVerify}
            loading={loading}
            disabled={code.length !== OTP_LENGTH}
          >
            Подтвердить
          </Button>
        </Space.Compact>
      )}
    </Space>
  )
}
