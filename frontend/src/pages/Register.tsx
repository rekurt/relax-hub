import { useState } from 'react'
import { Form, Input, Button, Card, Typography, Space, App, Segmented, Checkbox } from 'antd'
import { MailOutlined, LockOutlined, UserOutlined, PhoneOutlined } from '@ant-design/icons'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { postAuthRegister } from '@/api/generated/auth/auth'
import { useAuthStore } from '@/stores/auth'
import OAuthButtons from '@/components/OAuthButtons'
import type { AxiosError } from 'axios'
import type { InternalHandlerAPIResponse } from '@/api/generated/model'

const { Title, Text } = Typography

const ROLE_OPTIONS = [
  { label: 'Клиент', value: 'client' },
  { label: 'Владелец бани', value: 'owner' },
]

const ROLE_DESCRIPTIONS: Record<string, { title: string; subtitle: string }> = {
  client: { title: 'Регистрация клиента', subtitle: 'Создайте аккаунт для поиска и бронирования бань' },
  owner: { title: 'Регистрация владельца', subtitle: 'Создайте аккаунт для управления банями' },
}

interface RegisterFormValues {
  email: string
  password: string
  confirmPassword: string
  name: string
  phone: string
}

export default function Register() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const referralCode = searchParams.get('referral_code') ?? undefined
  const setAuth = useAuthStore((s) => s.setAuth)
  const { message } = App.useApp()
  const [loading, setLoading] = useState(false)
  const [role, setRole] = useState<string>('client')

  const getRoleHomePath = (r: string) => (r === 'client' ? '/client' : '/')

  const onFinish = async (values: RegisterFormValues) => {
    setLoading(true)
    try {
      const response = await postAuthRegister({
        email: values.email,
        password: values.password,
        name: values.name,
        phone: values.phone,
        role,
        referral_code: referralCode,
        age_confirmed: true,
      })
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

  const roleDesc = ROLE_DESCRIPTIONS[role] ?? ROLE_DESCRIPTIONS['client']!
  const title = roleDesc.title
  const subtitle = roleDesc.subtitle

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f5f5f5' }}>
      <Card style={{ width: 440 }}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div style={{ textAlign: 'center' }}>
            <Title level={3}>{title}</Title>
            <Text type="secondary">{subtitle}</Text>
          </div>

          <Segmented
            options={ROLE_OPTIONS}
            value={role}
            onChange={(v) => setRole(v as string)}
            block
          />

          <Form layout="vertical" onFinish={onFinish} autoComplete="off">
            <Form.Item
              name="name"
              rules={[{ required: true, message: 'Введите имя' }]}
            >
              <Input prefix={<UserOutlined />} placeholder="Имя" size="large" />
            </Form.Item>

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
              name="phone"
              rules={[{ required: true, message: 'Введите телефон' }]}
            >
              <Input prefix={<PhoneOutlined />} placeholder="Телефон" size="large" />
            </Form.Item>

            <Form.Item
              name="password"
              rules={[
                { required: true, message: 'Введите пароль' },
                { min: 6, message: 'Минимум 6 символов' },
              ]}
            >
              <Input.Password prefix={<LockOutlined />} placeholder="Пароль" size="large" />
            </Form.Item>

            <Form.Item
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
              <Input.Password prefix={<LockOutlined />} placeholder="Подтвердите пароль" size="large" />
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

            <Form.Item>
              <Button type="primary" htmlType="submit" loading={loading} block size="large">
                Зарегистрироваться
              </Button>
            </Form.Item>
          </Form>

          <OAuthButtons referralCode={referralCode} />

          <div style={{ textAlign: 'center' }}>
            <Text>Уже есть аккаунт? </Text>
            <Link to="/login">Войти</Link>
          </div>
        </Space>
      </Card>
    </div>
  )
}
