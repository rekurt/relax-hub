import { useState } from 'react'
import { Form, Input, Button, Card, Typography, Space, App, Segmented, Checkbox } from '@/components/design/system'
import { MailOutlined, LockOutlined, UserOutlined, PhoneOutlined } from '@/components/design/icons'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { postAuthRegister, postAuthRegisterPhone, postAuthVerifyPhone } from '@/api/generated/auth/auth'
import { useAuthStore } from '@/stores/auth'
import AuthShell from '@/components/AuthShell'
import OAuthButtons from '@/components/OAuthButtons'
import PhoneOTPInput from '@/components/PhoneOTPInput'
import type { AxiosError } from 'axios'
import type { InternalHandlerAPIResponse } from '@/api/generated/model'
import { PLATFORM_NAME } from '@/content/support'
import { syncAntdFormFromDOM } from '@/lib/autofill'

const { Text } = Typography

type AuthMethod = 'email' | 'phone'

const ROLE_OPTIONS = [
  { label: 'Клиент', value: 'client' },
  { label: 'Владелец бани', value: 'owner' },
]

const ROLE_DESCRIPTIONS: Record<string, { title: string; subtitle: string }> = {
  client: { title: 'Регистрация клиента', subtitle: 'Создайте аккаунт для поиска и бронирования бань' },
  owner: { title: 'Регистрация владельца', subtitle: 'Создайте аккаунт для управления банями' },
}

const AUTH_METHOD_OPTIONS = [
  { label: 'Email', value: 'email', icon: <MailOutlined /> },
  { label: 'Телефон', value: 'phone', icon: <PhoneOutlined /> },
]

interface EmailRegisterFormValues {
  email: string
  password: string
  confirmPassword: string
  name: string
  phone: string
  ageConfirmed: boolean
}

interface PhoneRegisterFormValues {
  name: string
  ageConfirmed: boolean
}

export default function Register() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const referralCode = searchParams.get('referral_code') ?? undefined
  const setAuth = useAuthStore((s) => s.setAuth)
  const { message } = App.useApp()
  const [loading, setLoading] = useState(false)
  const [role, setRole] = useState<string>('client')
  const [authMethod, setAuthMethod] = useState<AuthMethod>('email')
  const [phoneRegData, setPhoneRegData] = useState<{ name: string; ageConfirmed: boolean } | null>(null)
  const [emailForm] = Form.useForm<EmailRegisterFormValues>()

  // Chrome/Safari password-managers fill DOM values but may skip React change
  // events — drain the DOM into form state before submit so antd validators
  // don't reject autofilled fields as "empty".
  const handleEmailSubmitMouseDown = () => {
    syncAntdFormFromDOM(emailForm, ['name', 'email', 'phone', 'password', 'confirmPassword'])
  }

  const getRoleHomePath = (r: string) => (r === 'client' ? '/client' : '/')

  const onEmailFinish = async (values: EmailRegisterFormValues) => {
    // Defensive DOM fallback for autofill scenarios that bypass React change
    // events (keyboard Enter, password-manager pickers).
    const dom = {
      name: (document.getElementById('name') as HTMLInputElement | null)?.value,
      email: (document.getElementById('email') as HTMLInputElement | null)?.value,
      phone: (document.getElementById('phone') as HTMLInputElement | null)?.value,
      password: (document.getElementById('password') as HTMLInputElement | null)?.value,
    }
    const payload = {
      email: values.email || dom.email || '',
      password: values.password || dom.password || '',
      name: values.name || dom.name || '',
      phone: values.phone || dom.phone || '',
      role,
      referral_code: referralCode,
      age_confirmed: values.ageConfirmed,
    }
    if (!payload.email || !payload.password || !payload.name) {
      message.error('Заполните имя, email и пароль')
      return
    }
    setLoading(true)
    try {
      const response = await postAuthRegister(payload)
      if (response.success && response.data?.token && response.data.user) {
        setAuth(response.data.token, response.data.user)
        message.success('Регистрация прошла успешно')
        navigate(getRoleHomePath(role), { replace: true })
      } else {
        message.error(response.error?.message || 'Ошибка регистрации')
      }
    } catch (err) {
      const error = err as AxiosError<InternalHandlerAPIResponse>
      const msg = error.response?.data?.error?.message || 'Ошибка регистрации'
      message.error(msg)
    } finally {
      setLoading(false)
    }
  }

  const [phoneForm] = Form.useForm<PhoneRegisterFormValues>()

  const handlePhoneSendOTP = async (phone: string) => {
    const values = await phoneForm.validateFields()
    setPhoneRegData(values)
    await postAuthRegisterPhone({
      phone,
      name: values.name,
      age_confirmed: values.ageConfirmed,
    })
  }

  const handlePhoneVerified = async (phone: string, code: string) => {
    setLoading(true)
    try {
      const name = phoneRegData?.name || ''
      const response = await postAuthVerifyPhone({ phone, code, name })
      if (response.success && response.data?.token && response.data.user) {
        setAuth(response.data.token, response.data.user)
        message.success('Регистрация прошла успешно')
        navigate(getRoleHomePath(role), { replace: true })
      } else {
        message.error(response.error?.message || 'Ошибка регистрации')
      }
    } catch (err) {
      const error = err as AxiosError<InternalHandlerAPIResponse>
      message.error(error.response?.data?.error?.message || 'Неверный код')
    } finally {
      setLoading(false)
    }
  }

  const roleDesc = ROLE_DESCRIPTIONS[role] ?? ROLE_DESCRIPTIONS['client']!
  const title = roleDesc.title
  const subtitle = roleDesc.subtitle

  return (
    <AuthShell
      eyebrow="Регистрация"
      title={title}
      description={subtitle}
      asideTitle="Создание аккаунта без перегруза"
      asideDescription={`${PLATFORM_NAME} не превращает регистрацию в длинную анкету. Сначала создаём рабочий аккаунт, а профиль и операционные настройки можно спокойно дозаполнить уже внутри кабинета.`}
      highlights={[
        'Один поток для клиента и владельца: различается только маршрут после входа и доступные разделы.',
        'Телефон и email доступны как альтернативные способы регистрации, а не как две разные формы жизни.',
        referralCode ? `Реферальный код ${referralCode} будет применён автоматически после завершения регистрации.` : 'Если вы пришли по рекомендации, код прикрепится автоматически через ссылку приглашения.',
      ]}
      footer={(
        <>
          <Text>Уже есть аккаунт? </Text>
          <Link to="/login">Войти</Link>
        </>
      )}
    >
      <Card variant="borderless" className="rh-auth-surface">
        <Space orientation="vertical" size="large" style={{ width: '100%' }}>
          <Segmented
            className="rh-auth-segmented"
            options={ROLE_OPTIONS}
            value={role}
            onChange={(v) => setRole(v as string)}
            block
          />

          <Segmented
            className="rh-auth-segmented"
            options={AUTH_METHOD_OPTIONS}
            value={authMethod}
            onChange={(v) => setAuthMethod(v as AuthMethod)}
            block
          />

          {authMethod === 'email' ? (
            <Form
              form={emailForm}
              className="rh-auth-form"
              layout="vertical"
              onFinish={onEmailFinish}
              autoComplete="on"
            >
              <Form.Item
                label="Имя"
                name="name"
                rules={[{ required: true, message: 'Введите имя' }]}
              >
                <Input
                  prefix={<UserOutlined />}
                  placeholder="Имя"
                  size="large"
                  autoComplete="name"
                  name="name"
                />
              </Form.Item>

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
                label="Номер телефона"
                name="phone"
                rules={[{ required: true, message: 'Введите телефон' }]}
              >
                <Input
                  prefix={<PhoneOutlined />}
                  placeholder="Телефон"
                  size="large"
                  autoComplete="tel"
                  name="phone"
                  type="tel"
                />
              </Form.Item>

              <Form.Item
                label="Пароль"
                name="password"
                rules={[
                  { required: true, message: 'Введите пароль' },
                  { min: 6, message: 'Минимум 6 символов' },
                ]}
              >
                <Input.Password
                  prefix={<LockOutlined />}
                  placeholder="Пароль"
                  size="large"
                  autoComplete="new-password"
                  name="password"
                />
              </Form.Item>

              <Form.Item
                label="Подтверждение пароля"
                name="confirmPassword"
                dependencies={['password']}
                rules={[
                  { required: true, message: 'Подтвердите пароль' },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('password') === value) {
                        return Promise.resolve()
                      }
                      return Promise.reject(new Error('Пароли не совпадают'))
                    },
                  }),
                ]}
              >
                <Input.Password
                  prefix={<LockOutlined />}
                  placeholder="Подтвердите пароль"
                  size="large"
                  autoComplete="new-password"
                  name="confirmPassword"
                />
              </Form.Item>

              <Form.Item
                name="ageConfirmed"
                valuePropName="checked"
                rules={[
                  {
                    validator: (_, value) =>
                      value ? Promise.resolve() : Promise.reject(new Error('Необходимо подтвердить возраст')),
                  },
                ]}
              >
                <Checkbox>Мне исполнилось 18 лет</Checkbox>
              </Form.Item>

              <Form.Item className="rh-auth-form__actions">
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={loading}
                  onMouseDown={handleEmailSubmitMouseDown}
                  block
                  size="large"
                >
                  Зарегистрироваться
                </Button>
              </Form.Item>
            </Form>
          ) : (
            <Form form={phoneForm} className="rh-auth-form" layout="vertical" autoComplete="off">
              <Form.Item
                label="Имя"
                name="name"
                rules={[{ required: true, message: 'Введите имя' }]}
              >
                <Input prefix={<UserOutlined />} placeholder="Имя" size="large" />
              </Form.Item>

              <Form.Item
                name="ageConfirmed"
                valuePropName="checked"
                rules={[
                  {
                    validator: (_, value) =>
                      value ? Promise.resolve() : Promise.reject(new Error('Необходимо подтвердить возраст')),
                  },
                ]}
              >
                <Checkbox>Мне исполнилось 18 лет</Checkbox>
              </Form.Item>

              <PhoneOTPInput
                onSendOTP={handlePhoneSendOTP}
                onVerified={handlePhoneVerified}
                loading={loading}
              />
            </Form>
          )}

          <OAuthButtons referralCode={referralCode} />
        </Space>
      </Card>
    </AuthShell>
  )
}
