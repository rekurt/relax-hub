import { useState } from 'react'
import { Input, Button, Space, Typography, App } from '@/components/design/system'
import { SafetyOutlined, MessageOutlined } from '@/components/design/icons'
import { postAuth2faVerify } from '@/api/generated/2fa/2fa'
import type { AxiosError } from 'axios'
import type { InternalHandlerAPIResponse } from '@/api/generated/model'

const { Text } = Typography

const CODE_LENGTH = 6

interface TwoFactorChallengeProps {
  partialToken: string
  onSuccess: (token: string, user: unknown) => void
  onCancel: () => void
}

export default function TwoFactorChallenge({ partialToken, onSuccess, onCancel }: TwoFactorChallengeProps) {
  const { message } = App.useApp()
  const [code, setCode] = useState('')
  const [loading, setLoading] = useState(false)
  const [method, setMethod] = useState<'totp' | 'sms'>('totp')

  const handleVerify = async () => {
    if (code.length !== CODE_LENGTH) {
      message.warning(`Введите ${CODE_LENGTH}-значный код`)
      return
    }

    setLoading(true)
    try {
      const response = await postAuth2faVerify({ partial_token: partialToken, code })
      if (response.success && response.data?.token && response.data?.user) {
        onSuccess(response.data.token, response.data.user)
      } else {
        message.error(response.error?.message || 'Неверный код')
      }
    } catch (err) {
      const error = err as AxiosError<InternalHandlerAPIResponse>
      message.error(error.response?.data?.error?.message || 'Ошибка проверки кода')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Space orientation="vertical" size="large" style={{ width: '100%' }}>
      <div style={{ textAlign: 'center' }}>
        <SafetyOutlined style={{ fontSize: 48, color: '#0f766e', marginBottom: 16 }} />
        <Typography.Title level={4}>Двухфакторная аутентификация</Typography.Title>
        <Text type="secondary">
          {method === 'totp'
            ? 'Введите код из приложения-аутентификатора'
            : 'Введите код из SMS'}
        </Text>
      </div>

      <Input
        size="large"
        placeholder="000000"
        maxLength={CODE_LENGTH}
        value={code}
        onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
        onPressEnter={handleVerify}
        style={{ textAlign: 'center', fontSize: 24 }}
        autoFocus
      />

      <Button
        type="primary"
        size="large"
        block
        loading={loading}
        disabled={code.length !== CODE_LENGTH}
        onClick={handleVerify}
      >
        Подтвердить
      </Button>

      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
        <Button
          type="link"
          icon={<MessageOutlined />}
          onClick={() => {
            setMethod(method === 'totp' ? 'sms' : 'totp')
            setCode('')
          }}
        >
          {method === 'totp' ? 'Получить SMS' : 'Использовать приложение'}
        </Button>
        <Button type="link" onClick={onCancel}>
          Отмена
        </Button>
      </div>
    </Space>
  )
}
