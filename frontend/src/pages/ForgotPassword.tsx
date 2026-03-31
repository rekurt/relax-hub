import { useState } from 'react'
import { Form, Input, Button, Card, Typography, Space, App, Result } from 'antd'
import { MailOutlined } from '@ant-design/icons'
import { Link } from 'react-router-dom'
import { postAuthForgotPassword } from '@/api/generated/auth/auth'
import type { AxiosError } from 'axios'
import type { InternalHandlerAPIResponse } from '@/api/generated/model'

const { Title, Text } = Typography

export default function ForgotPassword() {
  const { message } = App.useApp()
  const [loading, setLoading] = useState(false)
  const [sent, setSent] = useState(false)

  const onFinish = async (values: { email: string }) => {
    setLoading(true)
    try {
      const response = await postAuthForgotPassword({ email: values.email })
      if (response.success) {
        setSent(true)
      } else {
        message.error(response.error?.message || 'Ошибка отправки')
      }
    } catch (err) {
      const error = err as AxiosError<InternalHandlerAPIResponse>
      const msg = error.response?.data?.error?.message || 'Ошибка отправки ссылки для восстановления'
      message.error(msg)
    } finally {
      setLoading(false)
    }
  }

  if (sent) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f5f5f5' }}>
        <Card style={{ width: 400 }}>
          <Result
            status="success"
            title="Письмо отправлено"
            subTitle="Проверьте почту и перейдите по ссылке для восстановления пароля"
            extra={
              <Link to="/login">
                <Button type="primary">Вернуться к входу</Button>
              </Link>
            }
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
            <Title level={3}>Восстановление пароля</Title>
            <Text type="secondary">Введите email, указанный при регистрации</Text>
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

            <Form.Item>
              <Button type="primary" htmlType="submit" loading={loading} block size="large">
                Отправить ссылку
              </Button>
            </Form.Item>
          </Form>

          <div style={{ textAlign: 'center' }}>
            <Link to="/login">Вернуться к входу</Link>
          </div>
        </Space>
      </Card>
    </div>
  )
}
