import { useState } from 'react'
import { Form, Input, Button, Card, Typography, Space, App } from 'antd'
import { MailOutlined, LockOutlined } from '@ant-design/icons'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { postAuthLogin } from '@/api/generated/auth/auth'
import { useAuthStore } from '@/stores/auth'
import { getRoleHomePath } from '@/stores/auth'
import type { InternalHandlerLoginRequest } from '@/api/generated/model'
import type { AxiosError } from 'axios'
import type { InternalHandlerAPIResponse } from '@/api/generated/model'

const { Title, Text } = Typography

export default function Login() {
  const navigate = useNavigate()
  const location = useLocation()
  const setAuth = useAuthStore((s) => s.setAuth)
  const { message } = App.useApp()
  const [loading, setLoading] = useState(false)

  const onFinish = async (values: InternalHandlerLoginRequest) => {
    setLoading(true)
    try {
      const response = await postAuthLogin(values)
      if (response.success && response.data?.token && response.data.user) {
        setAuth(response.data.token, response.data.user)
        message.success('Вы успешно вошли в систему')
        const from = (location.state as { from?: { pathname: string } })?.from?.pathname
        const defaultPath = getRoleHomePath(response.data.user.role)
        navigate(from ?? defaultPath, { replace: true })
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

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f5f5f5' }}>
      <Card style={{ width: 400 }}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div style={{ textAlign: 'center' }}>
            <Title level={3}>Вход в личный кабинет</Title>
            <Text type="secondary">Управление банями и бронированиями</Text>
          </div>

          <Form layout="vertical" onFinish={onFinish} autoComplete="off">
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
              <Button type="primary" htmlType="submit" loading={loading} block size="large">
                Войти
              </Button>
            </Form.Item>
          </Form>

          <div style={{ textAlign: 'center' }}>
            <Text>Нет аккаунта? </Text>
            <Link to="/register">Зарегистрироваться</Link>
          </div>
        </Space>
      </Card>
    </div>
  )
}
