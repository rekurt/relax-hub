import { useState } from 'react'
import { Form, Input, Button, Card, Typography, Space, App, Segmented } from 'antd'
import { MailOutlined, LockOutlined, PhoneOutlined } from '@ant-design/icons'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { postAuthLogin, postAuthLoginPhone, postAuthVerifyPhone } from '@/api/generated/auth/auth'
import { useAuthStore } from '@/stores/auth'
import { getRoleHomePath } from '@/stores/auth'
import AuthShell from '@/components/AuthShell'
import OAuthButtons from '@/components/OAuthButtons'
import PhoneOTPInput from '@/components/PhoneOTPInput'
import TwoFactorChallenge from '@/components/TwoFactorChallenge'
import type { InternalHandlerLoginRequest } from '@/api/generated/model'
import type { InternalHandlerUserResponse } from '@/api/generated/model'
import type { AxiosError } from 'axios'
import type { InternalHandlerAPIResponse } from '@/api/generated/model'
import { PLATFORM_NAME } from '@/content/support'
import { syncAntdFormFromDOM } from '@/lib/autofill'

const { Text } = Typography

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
  const [emailForm] = Form.useForm<InternalHandlerLoginRequest>()

  // Chrome/Safari password-managers fill DOM values but may skip React change
  // events — this drains the DOM into form state before submit fires.
  const handleEmailSubmitMouseDown = () => {
    syncAntdFormFromDOM(emailForm, ['email', 'password'])
  }

  const navigateAfterLogin = (user: InternalHandlerUserResponse) => {
    const from = (location.state as { from?: { pathname: string } })?.from?.pathname
    const defaultPath = getRoleHomePath(user.role)
    navigate(from ?? defaultPath, { replace: true })
  }

  const onEmailFinish = async (values: InternalHandlerLoginRequest) => {
    // Final defensive sync — if the mouse-down handler was bypassed (keyboard
    // Enter, assistive tech) read the DOM values directly before posting.
    const domEmail = (document.getElementById('email') as HTMLInputElement | null)?.value
    const domPassword = (document.getElementById('password') as HTMLInputElement | null)?.value
    const payload: InternalHandlerLoginRequest = {
      email: values.email || domEmail || '',
      password: values.password || domPassword || '',
    }
    if (!payload.email || !payload.password) {
      message.error('Введите email и пароль')
      return
    }
    setLoading(true)
    try {
      const response = await postAuthLogin(payload)
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
      <AuthShell
        eyebrow="Авторизация"
        title="Подтверждение входа"
        description="Завершите вход одноразовым кодом из включенного второго фактора."
        asideTitle="Один короткий шаг до кабинета"
        asideDescription={`Если на аккаунте включена дополнительная защита, ${PLATFORM_NAME} просит подтвердить вход отдельным кодом. Это закрывает случайные входы с чужих устройств.`}
        highlights={[
          'Подтверждение занимает несколько секунд и не требует повторного ввода логина и пароля.',
          'После успешной проверки вы попадете в тот раздел, куда шли изначально.',
          'Если код не приходит, вернитесь назад и повторите вход удобным способом.',
        ]}
      >
        <Card bordered={false} className="bani-auth-surface">
          <TwoFactorChallenge
            partialToken={twoFAState.partialToken}
            onSuccess={handle2FASuccess}
            onCancel={() => setTwoFAState(null)}
          />
        </Card>
      </AuthShell>
    )
  }

  return (
    <AuthShell
      eyebrow="Авторизация"
      title="Вход в личный кабинет"
      description="Управление банями и бронированиями"
      asideTitle="Вход без лишнего трения"
      asideDescription="Каталог, бронирования, чаты и служебные разделы открываются через один аккуратный сценарий входа. Выберите привычный способ и продолжайте с того места, где остановились."
      highlights={[
        'Email и телефон работают как равноправные сценарии, без скрытых шагов и лишних редиректов.',
        `Если на аккаунте включен второй фактор, ${PLATFORM_NAME} запросит подтверждение только после успешной проверки основного входа.`,
        'Для клиентов, владельцев и администраторов используется один и тот же поток входа, дальше система сама направит в нужный кабинет.',
      ]}
      footer={(
        <>
          <Text>Нет аккаунта? </Text>
          <Link to="/register">Зарегистрироваться</Link>
        </>
      )}
    >
      <Card bordered={false} className="bani-auth-surface">
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <Segmented
            className="bani-auth-segmented"
            options={AUTH_METHOD_OPTIONS}
            value={authMethod}
            onChange={(v) => setAuthMethod(v as AuthMethod)}
            block
          />

          {authMethod === 'email' ? (
            <Form
              form={emailForm}
              className="bani-auth-form"
              layout="vertical"
              onFinish={onEmailFinish}
              autoComplete="on"
            >
              <Form.Item
                label="Адрес email"
                name="email"
                rules={[
                  { required: true, message: 'Введите email' },
                  { type: 'email', message: 'Некорректный email' },
                ]}
              >
                <Input
                  prefix={<MailOutlined />}
                  placeholder="Email"
                  size="large"
                  autoComplete="email"
                  name="email"
                  type="email"
                />
              </Form.Item>

              <Form.Item
                label="Пароль"
                name="password"
                rules={[{ required: true, message: 'Введите пароль' }]}
              >
                <Input.Password
                  prefix={<LockOutlined />}
                  placeholder="Пароль"
                  size="large"
                  autoComplete="current-password"
                  name="password"
                />
              </Form.Item>

              <Form.Item className="bani-auth-form__actions">
                <div className="bani-auth-form__link-row">
                  <Link to="/forgot-password">Забыли пароль?</Link>
                </div>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={loading}
                  onMouseDown={handleEmailSubmitMouseDown}
                  block
                  size="large"
                >
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
        </Space>
      </Card>
    </AuthShell>
  )
}
