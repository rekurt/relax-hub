import { useState, useEffect, useCallback } from 'react'
import { Input, Button, App, Typography } from 'antd'
import { PhoneOutlined } from '@ant-design/icons'

const OTP_COOLDOWN_SECONDS = 60
const OTP_LENGTH = 6

interface PhoneOTPInputProps {
  onVerified: (phone: string, code: string) => void
  onSendOTP: (phone: string) => Promise<void>
  loading?: boolean
  phoneLabel?: string
}

const { Text } = Typography

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
    <div className="bani-auth-phone">
      <div className="bani-auth-phone__block">
        <div className="bani-auth-field">
          <Text className="bani-auth-field__label">{phoneLabel}</Text>
          <Input
            prefix={<PhoneOutlined />}
            placeholder={phoneLabel}
            size="large"
            autoComplete="tel"
            inputMode="tel"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            disabled={otpSent && cooldown > 0}
          />
        </div>
        <div className="bani-auth-phone__meta">
          <Text type="secondary" className="bani-auth-phone__hint">
            Отправим одноразовый код по SMS. Номер можно вводить в привычном формате, включая `+7`.
          </Text>
          {otpSent && cooldown > 0 ? (
            <Text className="bani-auth-phone__status">Повтор через {cooldown}с</Text>
          ) : null}
        </div>
        <Button
          type={phone.trim() && cooldown <= 0 && !otpSent ? 'primary' : 'default'}
          block
          size="large"
          onClick={handleSendOTP}
          loading={sending}
          disabled={cooldown > 0 || !phone.trim()}
        >
          {cooldown > 0 ? `${cooldown}с` : otpSent ? 'Отправить снова' : 'Получить код'}
        </Button>
      </div>

      {otpSent && (
        <div className="bani-auth-phone__block bani-auth-phone__block--confirm">
          <div className="bani-auth-field">
            <Text className="bani-auth-field__label">Код из SMS</Text>
            <Input
              placeholder="Введите код из SMS"
              size="large"
              autoComplete="one-time-code"
              inputMode="numeric"
              maxLength={OTP_LENGTH}
              value={code}
              onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
              onPressEnter={handleVerify}
            />
          </div>
          <Text type="secondary" className="bani-auth-phone__hint">
            Код состоит из {OTP_LENGTH} цифр. После подтверждения вы сразу попадёте в нужный кабинет.
          </Text>
          <Button
            type="primary"
            block
            size="large"
            onClick={handleVerify}
            loading={loading}
            disabled={code.length !== OTP_LENGTH}
          >
            Подтвердить
          </Button>
        </div>
      )}
    </div>
  )
}
