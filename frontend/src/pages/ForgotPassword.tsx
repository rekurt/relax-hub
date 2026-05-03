import { useState } from 'react'
import { Form, Input, Button, Card, Space, App, Result } from '@/components/design/system'
import { MailOutlined } from '@/components/design/icons'
import { Link } from 'react-router-dom'
import { postAuthForgotPassword } from '@/api/generated/auth/auth'
import AuthShell from '@/components/AuthShell'
import type { AxiosError } from 'axios'
import type { InternalHandlerAPIResponse } from '@/api/generated/model'
import { syncAntdFormFromDOM } from '@/lib/autofill'

export default function ForgotPassword() {
  const { message } = App.useApp()
  const [loading, setLoading] = useState(false)
  const [sent, setSent] = useState(false)
  const [form] = Form.useForm<{ email: string }>()

  const handleSubmitMouseDown = () => {
    syncAntdFormFromDOM(form, ['email'])
  }

  const onFinish = async (values: { email: string }) => {
    const domEmail = (document.getElementById('email') as HTMLInputElement | null)?.value
    const email = values.email || domEmail || ''
    if (!email) {
      message.error('Введите email')
      return
    }
    setLoading(true)
    try {
      const response = await postAuthForgotPassword({ email })
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
      <AuthShell
        eyebrow="Восстановление доступа"
        title="Инструкция отправлена"
        description="Мы отправили инструкцию на указанный email."
        asideTitle="Доступ восстанавливается без поддержки"
        asideDescription="Страница сброса пароля должна закрывать задачу пользователя за пару минут, без ручных обращений и без лишних форм."
        highlights={[
          'Ссылка приходит на email, указанный при регистрации.',
          'После перехода по ссылке можно сразу задать новый пароль.',
          'Если письма нет, проверьте папку со спамом и повторите запрос.',
        ]}
      >
        <Card variant="borderless" className="rh-auth-surface">
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
      </AuthShell>
    )
  }

  return (
    <AuthShell
      eyebrow="Восстановление доступа"
      title="Восстановление пароля"
      description="Введите email, указанный при регистрации"
      asideTitle="Возврат доступа без долгого сценария"
      asideDescription="Восстановление пароля должно быть предсказуемым: один email, одно письмо, один переход обратно к входу."
      highlights={[
        'Ссылка для восстановления приходит только на подтверждённый email аккаунта.',
        'Пароль меняется по защищённой ссылке, без общения с поддержкой.',
        'После смены можно сразу вернуться ко входу и продолжить работу.',
      ]}
      footer={<Link to="/login">Вернуться к входу</Link>}
    >
      <Card variant="borderless" className="rh-auth-surface">
        <Space orientation="vertical" size="large" style={{ width: '100%' }}>
          <Form
            form={form}
            className="rh-auth-form"
            layout="vertical"
            onFinish={onFinish}
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

            <Form.Item className="rh-auth-form__actions">
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                onMouseDown={handleSubmitMouseDown}
                block
                size="large"
              >
                Отправить ссылку
              </Button>
            </Form.Item>
          </Form>
        </Space>
      </Card>
    </AuthShell>
  )
}
