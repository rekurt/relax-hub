import {
  App,
  Button,
  Card,
  Checkbox,
  Form,
  Input,
  Select,
  Space,
  Typography,
} from 'antd'
import {
  ArrowLeftOutlined,
  SendOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import {
  useGetMyCrmSegments,
  usePostMyCrmBroadcasts,
} from '@/api/generated/crm/crm'
import { useQueryClient } from '@tanstack/react-query'

const { Title } = Typography

const CHANNEL_OPTIONS = [
  { label: 'Push-уведомление', value: 'push' },
  { label: 'Email', value: 'email' },
  { label: 'Telegram', value: 'telegram' },
]

interface BroadcastFormValues {
  segment: string
  title: string
  body: string
  image_url?: string
  channels: string[]
}

export default function BroadcastCreate() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { message } = App.useApp()
  const [form] = Form.useForm<BroadcastFormValues>()

  const { data: segmentsData } = useGetMyCrmSegments()
  const segments = segmentsData?.data ?? []

  const createMutation = usePostMyCrmBroadcasts({
    mutation: {
      onSuccess: () => {
        message.success('Рассылка создана')
        queryClient.invalidateQueries({ queryKey: ['/my/crm/broadcasts'] })
        navigate('/crm/broadcasts')
      },
      onError: () => message.error('Не удалось создать рассылку'),
    },
  })

  const handleSubmit = (values: BroadcastFormValues) => {
    createMutation.mutate({
      data: {
        segment: values.segment,
        title: values.title,
        body: values.body,
        image_url: values.image_url || undefined,
        channels: values.channels,
      },
    })
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/crm/broadcasts')}>
          Назад
        </Button>
        <Title level={3} style={{ margin: 0 }}>Новая рассылка</Title>
      </Space>

      <Card>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{ channels: ['push'] }}
          style={{ maxWidth: 600 }}
        >
          <Form.Item
            name="segment"
            label="Сегмент получателей"
            rules={[{ required: true, message: 'Выберите сегмент' }]}
          >
            <Select placeholder="Выберите сегмент">
              {segments.map((seg) => (
                <Select.Option key={seg.slug} value={seg.slug}>
                  {seg.name} ({seg.count ?? 0} гостей)
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="title"
            label="Заголовок"
            rules={[{ required: true, message: 'Введите заголовок' }]}
          >
            <Input placeholder="Заголовок рассылки" maxLength={200} showCount />
          </Form.Item>

          <Form.Item
            name="body"
            label="Текст сообщения"
            rules={[{ required: true, message: 'Введите текст' }]}
          >
            <Input.TextArea
              rows={6}
              placeholder="Текст рассылки для гостей..."
              maxLength={2000}
              showCount
            />
          </Form.Item>

          <Form.Item
            name="image_url"
            label="Ссылка на изображение"
            extra="Необязательно. URL изображения для рассылки"
          >
            <Input placeholder="https://..." />
          </Form.Item>

          <Form.Item
            name="channels"
            label="Каналы отправки"
            rules={[{ required: true, message: 'Выберите хотя бы один канал' }]}
          >
            <Checkbox.Group options={CHANNEL_OPTIONS} />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                icon={<SendOutlined />}
                loading={createMutation.isPending}
              >
                Создать черновик
              </Button>
              <Button onClick={() => navigate('/crm/broadcasts')}>
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
