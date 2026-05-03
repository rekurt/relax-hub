import { useState } from 'react'
import {
  App,
  Button,
  Card,
  DatePicker,
  Descriptions,
  Empty,
  Form,
  Input,
  InputNumber,
  Select,
  Space,
  Tag,
  Typography,
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { usePostAdminPromoCodes } from '@/api/generated/promo-codes/promo-codes'
import type { InternalHandlerPromoResponse } from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'
import dayjs from 'dayjs'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

const PROMO_TYPES = [
  { label: 'Процент', value: 'percentage' },
  { label: 'Фиксированная сумма', value: 'fixed_amount' },
  { label: 'Бесплатный час', value: 'free_hour' },
]

const typeLabel: Record<string, string> = {
  percentage: 'Процент',
  fixed_amount: 'Фиксированная сумма',
  free_hour: 'Бесплатный час',
}

const typeColor: Record<string, string> = {
  percentage: 'blue',
  fixed_amount: 'green',
  free_hour: 'purple',
}

interface PromoFormValues {
  code: string
  type: string
  value: number
  max_uses?: number
  min_amount?: number
  valid_from?: dayjs.Dayjs
  valid_until?: dayjs.Dayjs
}

export default function GlobalPromoCodes() {
  const { message } = App.useApp()
  const [form] = Form.useForm<PromoFormValues>()
  const [formVisible, setFormVisible] = useState(false)
  const [createdPromos, setCreatedPromos] = useState<InternalHandlerPromoResponse[]>([])

  const createMutation = usePostAdminPromoCodes()

  const selectedType = Form.useWatch('type', form)

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const result = await createMutation.mutateAsync({
        data: {
          code: values.code,
          type: values.type,
          value: values.type === 'free_hour' ? 1 : values.type === 'fixed_amount' ? Math.round(values.value * 100) : values.value,
          max_uses: values.max_uses,
          min_amount: values.min_amount ? Math.round(values.min_amount * 100) : undefined,
          valid_from: values.valid_from?.toISOString(),
          valid_until: values.valid_until?.toISOString(),
        },
      })
      message.success('Промокод создан')
      if (result.data) {
        setCreatedPromos((prev) => [result.data as InternalHandlerPromoResponse, ...prev])
      }
      form.resetFields()
      setFormVisible(false)
    } catch {
      message.error('Не удалось создать промокод')
    }
  }

  const formatValue = (promo: InternalHandlerPromoResponse) => {
    if (promo.type === 'percentage') return `${promo.value}%`
    if (promo.type === 'fixed_amount') return formatPrice(promo.value ?? 0)
    return '1 час'
  }

  return (
    <div>
      <PageHeader
        eyebrow="Маркетинг"
        title="Глобальные промокоды"
        description="Глобальные промокоды действуют на все бани и создаются администратором. Единый экран для создания и просмотра промокодов."
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              form.resetFields()
              setFormVisible(true)
            }}
          >
            Создать промокод
          </Button>
        }
      />

      {formVisible && (
        <Card
          title="Новый промокод"
          style={{ marginBottom: 24 }}
          extra={
            <Button type="text" onClick={() => setFormVisible(false)}>
              Отмена
            </Button>
          }
        >
          <Form form={form} layout="vertical" onFinish={handleSubmit}>
            <Form.Item
              name="code"
              label="Код"
              rules={[{ required: true, message: 'Введите код промокода' }]}
            >
              <Input placeholder="SUMMER2026" style={{ textTransform: 'uppercase' }} />
            </Form.Item>

            <Form.Item
              name="type"
              label="Тип скидки"
              rules={[{ required: true, message: 'Выберите тип скидки' }]}
            >
              <Select options={PROMO_TYPES} placeholder="Выберите тип" />
            </Form.Item>

            {selectedType && selectedType !== 'free_hour' && (
              <Form.Item
                name="value"
                label={selectedType === 'percentage' ? 'Скидка (%)' : 'Скидка (руб.)'}
                rules={[{ required: true, message: 'Введите значение скидки' }]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={1}
                  max={selectedType === 'percentage' ? 100 : undefined}
                  placeholder={selectedType === 'percentage' ? '10' : '500'}
                />
              </Form.Item>
            )}

            <Form.Item name="max_uses" label="Максимум использований">
              <InputNumber style={{ width: '100%' }} min={1} placeholder="Без ограничений" />
            </Form.Item>

            <Form.Item name="min_amount" label="Минимальная сумма заказа (руб.)">
              <InputNumber style={{ width: '100%' }} min={0} placeholder="0" />
            </Form.Item>

            <Space>
              <Form.Item name="valid_from" label="Действует с">
                <DatePicker showTime format="DD.MM.YYYY HH:mm" placeholder="Начало" />
              </Form.Item>
              <Form.Item name="valid_until" label="Действует до">
                <DatePicker showTime format="DD.MM.YYYY HH:mm" placeholder="Окончание" />
              </Form.Item>
            </Space>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                loading={createMutation.isPending}
              >
                Создать
              </Button>
            </Form.Item>
          </Form>
        </Card>
      )}

      {createdPromos.length === 0 && !formVisible ? (
        <Empty description="Нет созданных промокодов в этой сессии" />
      ) : (
        <Space orientation="vertical" size={16} style={{ width: '100%' }}>
          {createdPromos.map((promo) => (
            <Card key={promo.id} size="small">
              <Descriptions column={{ xs: 1, sm: 2, md: 3 }} size="small">
                <Descriptions.Item label="Код">
                  <Text strong copyable>{promo.code}</Text>
                </Descriptions.Item>
                <Descriptions.Item label="Тип">
                  <Tag color={typeColor[promo.type ?? ''] ?? 'default'}>
                    {typeLabel[promo.type ?? ''] ?? promo.type}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="Значение">
                  {formatValue(promo)}
                </Descriptions.Item>
                {promo.max_uses !== undefined && promo.max_uses > 0 && (
                  <Descriptions.Item label="Макс. использований">
                    {promo.max_uses}
                  </Descriptions.Item>
                )}
                {promo.min_amount !== undefined && promo.min_amount > 0 && (
                  <Descriptions.Item label="Мин. сумма">
                    {formatPrice(promo.min_amount)}
                  </Descriptions.Item>
                )}
                {promo.valid_from && (
                  <Descriptions.Item label="С">
                    {formatDateTime(promo.valid_from)}
                  </Descriptions.Item>
                )}
                {promo.valid_until && (
                  <Descriptions.Item label="До">
                    {formatDateTime(promo.valid_until)}
                  </Descriptions.Item>
                )}
                <Descriptions.Item label="Создан">
                  {promo.created_at ? formatDateTime(promo.created_at) : '-'}
                </Descriptions.Item>
              </Descriptions>
            </Card>
          ))}
        </Space>
      )}
    </div>
  )
}
