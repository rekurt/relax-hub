import { useNavigate, useSearchParams } from 'react-router-dom'
import {
  App,
  Button,
  Card,
  Form,
  Input,
  Select,
  Space,
  Typography,
} from 'antd'
import { ArrowLeftOutlined } from '@ant-design/icons'
import { useQueryClient } from '@tanstack/react-query'
import {
  usePostBookingsIdDispute,
} from '@/api/generated/disputes/disputes'
import { getGetMyDisputesQueryKey } from '@/api/generated/disputes/disputes'

const { Title } = Typography

const REASON_OPTIONS = [
  { label: 'Услуга не оказана', value: 'service_not_provided' },
  { label: 'Низкое качество', value: 'poor_quality' },
  { label: 'Повреждение имущества', value: 'damage' },
  { label: 'Проблема безопасности', value: 'safety_issue' },
  { label: 'Ошибка в счёте', value: 'billing_error' },
  { label: 'Другое', value: 'other' },
]

export default function DisputeCreate() {
  const navigate = useNavigate()
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [searchParams] = useSearchParams()
  const bookingId = searchParams.get('booking') ?? ''
  const [form] = Form.useForm()

  const createMutation = usePostBookingsIdDispute()

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const result = await createMutation.mutateAsync({
        id: bookingId,
        data: {
          reason: values.reason,
          description: values.description,
        },
      })
      message.success('Спор открыт')
      queryClient.invalidateQueries({ queryKey: getGetMyDisputesQueryKey() })
      const disputeId = result?.data?.id
      if (disputeId) {
        navigate(`/client/disputes/${disputeId}`)
      } else {
        navigate('/client/disputes')
      }
    } catch {
      if (createMutation.isError) {
        message.error('Не удалось открыть спор')
      }
    }
  }

  if (!bookingId) {
    return (
      <div style={{ maxWidth: 600, margin: '0 auto', textAlign: 'center', padding: '40px 0' }}>
        <Title level={4}>Бронирование не указано</Title>
        <Button onClick={() => navigate('/client/bookings')}>К бронированиям</Button>
      </div>
    )
  }

  return (
    <div style={{ maxWidth: 600, margin: '0 auto' }}>
      <Space style={{ marginBottom: 16 }}>
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate(-1)}
        >
          Назад
        </Button>
      </Space>

      <Title level={3} style={{ marginBottom: 16 }}>
        Открыть спор
      </Title>

      <Card>
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item label="Бронирование">
            <Input value={bookingId} disabled />
          </Form.Item>
          <Form.Item
            name="reason"
            label="Причина спора"
            rules={[{ required: true, message: 'Выберите причину' }]}
          >
            <Select
              placeholder="Выберите причину"
              options={REASON_OPTIONS}
            />
          </Form.Item>
          <Form.Item
            name="description"
            label="Описание"
            rules={[{ required: true, message: 'Опишите проблему' }]}
          >
            <Input.TextArea
              rows={5}
              placeholder="Подробно опишите вашу проблему. Что произошло, когда это случилось, какой результат вы ожидали."
              maxLength={5000}
              showCount
            />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={createMutation.isPending}
              >
                Открыть спор
              </Button>
              <Button onClick={() => navigate(-1)}>
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
