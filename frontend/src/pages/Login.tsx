import { useState } from 'react'
import { Form, Input, Button, Card, Typography, Space, App, Segmented } from 'antd'
import { MailOutlined, LockOutlined, PhoneOutlined } from '@ant-design/icons'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { postAuthLogin, postAuthLoginPhone, postAuthVerifyPhone } from '@/api/generated/auth/auth'
import { useAuthStore } from '@/stores/auth'
import { getRoleHomePath } from '@/stores/auth'
import OAuthButtons from '@/components/OAuthButtons'
import PhoneOTPInput from '@/components/PhoneOTPInput'
import TwoFactorChallenge from '@/components/TwoFactorChallenge'
import type { InternalHandlerLoginRequest } from '@/api/generated/model'
import type { InternalHandlerUserResponse } from '@/api/generated/model'
import type { AxiosError } from 'axios'
import type { InternalHandlerAPIResponse } from '@/api/generated/model'

const { Title, Text } = Typography

type AuthMethod = 'email' | 'phone'

const AUTH_METHOD_OPTIONS = [
  { label: 'Email', value: 'email', icon: <MailOutlined /> },
  { label: 'Телефон', value: 'phone', icon: <PhoneOutlined /> },
]

export default function Login() {
  const navigate = useNavigate()
  const location = useLocation()
  const setAuth = useAuthStore((s) => s.setAuth)
  const { message } = App.useApp()
  const [loading, setLoading] = useState(false)
  const [authMethod, setAuthMethod] = useState<AuthMethod>('email')
  const [twoFAState, setTwoFAState] = useState<{ partialToken: string } | null>(null)

  const navigateAfterLogin = (user: InternalHandlerUserResponse) => {
    const from = (location.state as { from?: { pathname: string } })?.from?.pathname
    const defaultPath = getRoleHomePath(user.role)
    navigate(from ?? defaultPath, { replace: true })
  }

  const onEmailFinish = async (values: InternalHandlerLoginRequest) => {
    setLoading(true)
    try {
      const response = await postAuthLogin(values)
      if (response.success && response.data) {
        if (response.data.requires_2fa && response.data.token) {
          setTwoFAState({ partialToken: response.data.token })
          return
        }
        if (response.data.token && response.data.user) {
          setAuth(response.data.token, response.data.user)
          message.success('Вы успешно вошли в систему')
          navigateAfterLogin(response.data.user)
        }
      } else {
        message.error(response.error?.message || 'Ошибка авторизации')
      }
    } catch (err) {
      const error = err as AxiosError<InternalHandlerAPIResponse>
      const msg = error.response?.data?.error?.message || 'Неверный email или пароль'
      message.error(msg)
    } finally {
      setLoading(false)
    }
  }

  const handlePhoneSendOTP = async (phone: string) => {
    await postAuthLoginPhone({ phone })
  }

  const handlePhoneVerified = async (phone: string, code: string) => {
    setLoading(true)
    try {
      const response = await postAuthVerifyPhone({ phone, code })
      if (response.success && response.data) {
        if (response.data.requires_2fa && response.data.token) {
          setTwoFAState({ partialToken: response.data.token })
          return
        }
        if (response.data.token && response.data.user) {
          setAuth(response.data.token, response.data.user)
          message.success('Вы успешно вошли в систему')
          navigateAfterLogin(response.data.user)
        }
      } else {
        message.error(response.error?.message || 'Ошибка авторизации')
      }
    } catch (err) {
      const error = err as AxiosError<InternalHandlerAPIResponse>
      message.error(error.response?.data?.error?.message || 'Неверный код')
    } finally {
      setLoading(false)
    }
  }

  const handle2FASuccess = (token: string, user: unknown) => {
    const typedUser = user as InternalHandlerUserResponse
    setAuth(token, typedUser)
    message.success('Вы успешно вошли в систему')
    setTwoFAState(null)
    navigateAfterLogin(typedUser)
  }

  if (twoFAState) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f5f5f5' }}>
        <Card style={{ width: 400 }}>
          <TwoFactorChallenge
            partialToken={twoFAState.partialToken}
            onSuccess={handle2FASuccess}
            onCancel={() => setTwoFAState(null)}
          />
        </Card>
      </div>
    )
  }

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f5f5f5' }}>
      <Card style={{ width: 400 }}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div style={{ textAlign: 'center' }}>
            <Title level={3}>Вход в личный кабинет</Title>
            <Text type="secondary">Управление банями и бронированиями</Text>
          </div>

          <Segmented
            options={AUTH_METHOD_OPTIONS}
            value={authMethod}
            onChange={(v) => setAuthMethod(v as AuthMethod)}
            block
          />

          {authMethod === 'email' ? (
            <Form layout="vertical" onFinish={onEmailFinish} autoComplete="off">
              <Form.Item
                name="email"
                rules={[
                  { required: true, message: 'Введите email' },
                  { type: 'email', message: 'Некорректный email' },
                ]}
              >
                <Input prefix={<MailOutlined />} placeholder="Email" size="large" />
              </Form.Item>

              <Form.Item
                name="password"
                rules={[{ required: true, message: 'Введите пароль' }]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder="Пароль" size="large" />
              </Form.Item>

              <Form.Item>
                <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 8 }}>
                  <Link to="/forgot-password">Забыли пароль?</Link>
                </div>
                <Button type="primary" htmlType="submit" loading={loading} block size="large">
                  Войти
                </Button>
              </Form.Item>
            </Form>
          ) : (
            <PhoneOTPInput
              onSendOTP={handlePhoneSendOTP}
              onVerified={handlePhoneVerified}
              loading={loading}
            />
          )}

          <OAuthButtons />

          <div style={{ textAlign: 'center' }}>
            <Text>Нет аккаунта? </Text>
            <Link to="/register">Зарегистрироваться</Link>
          </div>
        </Space>
      </Card>
    </div>
  )
}
