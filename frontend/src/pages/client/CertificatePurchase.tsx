import { useState } from 'react'
import { Typography, Card, InputNumber, Input, Button, Form, Result, Descriptions, Space, Alert } from 'antd'
import { GiftOutlined, ArrowLeftOutlined } from '@ant-design/icons'
import { App } from 'antd'
import { Link, useNavigate } from 'react-router-dom'
import { usePostCertificatesPurchase } from '@/api/generated/certificates/certificates'
import { formatPrice } from '@/lib/format'
import type { InternalHandlerCertificateResponse } from '@/api/generated/model'
import { useAuthStore } from '@/stores/auth'

const { Title, Text } = Typography

const PRESET_AMOUNTS = [100000, 200000, 300000, 500000] // in kopecks

export default function CertificatePurchase() {
  const { message } = App.useApp()
  const navigate = useNavigate()
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const [form] = Form.useForm()
  const [purchasedCertificate, setPurchasedCertificate] = useState<InternalHandlerCertificateResponse | null>(null)

  const purchaseMutation = usePostCertificatesPurchase()

  const handleSubmit = (values: {
    amount: number
    purchaser_email: string
    recipient_email?: string
    recipient_name?: string
    message?: string
  }) => {
    const amountKopecks = Math.round(values.amount * 100)
    purchaseMutation.mutate(
      {
        data: {
          amount: amountKopecks,
          purchaser_email: values.purchaser_email,
          recipient_email: values.recipient_email || undefined,
          recipient_name: values.recipient_name || undefined,
          message: values.message || undefined,
        },
      },
      {
        onSuccess: (response) => {
          message.success('Сертификат успешно создан!')
          setPurchasedCertificate(response.data ?? null)
        },
        onError: () => {
          message.error('Не удалось создать сертификат')
        },
      },
    )
  }

  const handlePresetAmount = (kopecks: number) => {
    form.setFieldsValue({ amount: kopecks / 100 })
  }

  if (purchasedCertificate) {
    return (
      <div>
        <Result
          status="success"
          title="Сертификат создан!"
          subTitle="Сохраните код сертификата — он понадобится для использования при бронировании"
          extra={[
            isAuthenticated && (
              <Button key="list" onClick={() => navigate('/client/certificates')}>
                Мои сертификаты
              </Button>
            ),
            <Button key="new" type="primary" onClick={() => {
              setPurchasedCertificate(null)
              form.resetFields()
            }}>
              Купить ещё
            </Button>,
          ].filter(Boolean)}
        />
        <Card style={{ maxWidth: 500, margin: '0 auto' }}>
          <Descriptions bordered column={1}>
            <Descriptions.Item label="Код">
              <Text copyable strong style={{ fontSize: 18 }}>
                {purchasedCertificate.code}
              </Text>
            </Descriptions.Item>
            <Descriptions.Item label="Номинал">
              {formatPrice(purchasedCertificate.amount ?? 0)}
            </Descriptions.Item>
            <Descriptions.Item label="Действителен до">
              {purchasedCertificate.valid_until
                ? new Date(purchasedCertificate.valid_until).toLocaleDateString('ru-RU')
                : '—'}
            </Descriptions.Item>
            {purchasedCertificate.recipient_name && (
              <Descriptions.Item label="Получатель">
                {purchasedCertificate.recipient_name}
              </Descriptions.Item>
            )}
            {purchasedCertificate.message && (
              <Descriptions.Item label="Сообщение">
                {purchasedCertificate.message}
              </Descriptions.Item>
            )}
          </Descriptions>
        </Card>
      </div>
    )
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        {isAuthenticated ? (
          <Link to="/client/certificates">
            <Button icon={<ArrowLeftOutlined />}>Назад к сертификатам</Button>
          </Link>
        ) : (
          <Link to="/">
            <Button icon={<ArrowLeftOutlined />}>На главную</Button>
          </Link>
        )}
      </Space>

      <Title level={2}>
        <GiftOutlined /> Купить подарочный сертификат
      </Title>

      <Alert
        message="Подарочный сертификат"
        description="Сертификат можно использовать для оплаты бронирований. Срок действия — 365 дней. Неиспользованный остаток сохраняется."
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <Card style={{ maxWidth: 600 }}>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            name="amount"
            label="Сумма (в рублях)"
            rules={[
              { required: true, message: 'Укажите сумму' },
              { type: 'number', min: 100, message: 'Минимальная сумма — 100 ₽' },
              { type: 'number', max: 100000, message: 'Максимальная сумма — 100 000 ₽' },
            ]}
          >
            <InputNumber
              style={{ width: '100%' }}
              placeholder="Введите сумму"
              min={100}
              max={100000}
              addonAfter="₽"
            />
          </Form.Item>

          <Space wrap style={{ marginBottom: 16 }}>
            {PRESET_AMOUNTS.map((amount) => (
              <Button key={amount} onClick={() => handlePresetAmount(amount)}>
                {formatPrice(amount)}
              </Button>
            ))}
          </Space>

          <Form.Item
            name="purchaser_email"
            label="Ваш email"
            rules={[
              { required: true, message: 'Укажите email' },
              { type: 'email', message: 'Введите корректный email' },
            ]}
          >
            <Input placeholder="your@email.com" />
          </Form.Item>

          <Form.Item
            name="recipient_name"
            label="Имя получателя"
            extra="Необязательно — если хотите подарить сертификат"
          >
            <Input placeholder="Имя получателя" />
          </Form.Item>

          <Form.Item
            name="recipient_email"
            label="Email получателя"
            rules={[
              { type: 'email', message: 'Введите корректный email' },
            ]}
            extra="Получатель получит код сертификата на этот email"
          >
            <Input placeholder="recipient@email.com" />
          </Form.Item>

          <Form.Item
            name="message"
            label="Сообщение"
            extra="Необязательно — будет отправлено получателю"
          >
            <Input.TextArea rows={3} placeholder="Поздравляю! Желаю приятного отдыха!" maxLength={500} showCount />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              size="large"
              icon={<GiftOutlined />}
              loading={purchaseMutation.isPending}
              block
            >
              Купить сертификат
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
