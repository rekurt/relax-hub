import {
  App,
  Button,
  Card,
  Checkbox,
  Form,
  Input,
  Select,
  Space,
  Tag,
  Tooltip,
  Typography,
} from 'antd'
import {
  ArrowLeftOutlined,
  SendOutlined,
} from '@ant-design/icons'
import { useRef } from 'react'
import type { TextAreaRef } from 'antd/es/input/TextArea'
import { useNavigate } from 'react-router-dom'
import {
  useGetMyCrmSegments,
  usePostMyCrmBroadcasts,
} from '@/api/generated/crm/crm'
import { useQueryClient } from '@tanstack/react-query'

const { Title, Text } = Typography

const CHANNEL_OPTIONS = [
  { label: 'Push-уведомление', value: 'push' },
  { label: 'Email', value: 'email' },
  { label: 'Telegram', value: 'telegram' },
  { label: 'SMS', value: 'sms' },
]

const PERSONALIZATION_TOKENS = [
  { token: '{{guest_name}}', label: 'Имя гостя', description: 'Имя клиента из профиля' },
  { token: '{{last_visit_date}}', label: 'Дата визита', description: 'Дата последнего визита (ДД.ММ.ГГГГ)' },
  { token: '{{visit_count}}', label: 'Кол-во визитов', description: 'Общее количество визитов гостя' },
  { token: '{{promo_code}}', label: 'Промокод', description: 'Промокод из привязанной акции' },
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
  const bodyRef = useRef<TextAreaRef>(null)

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

  const insertToken = (token: string) => {
    const textarea = bodyRef.current?.resizableTextArea?.textArea
    if (!textarea) return

    const start = textarea.selectionStart
    const end = textarea.selectionEnd
    const currentBody = form.getFieldValue('body') || ''
    const newBody = currentBody.slice(0, start) + token + currentBody.slice(end)
    form.setFieldValue('body', newBody)

    // Restore cursor after token
    requestAnimationFrame(() => {
      const pos = start + token.length
      textarea.focus()
      textarea.setSelectionRange(pos, pos)
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
              ref={bodyRef}
              rows={6}
              placeholder="Текст рассылки для гостей... Используйте токены персонализации ниже"
              maxLength={2000}
              showCount
            />
          </Form.Item>

          <div style={{ marginBottom: 24 }}>
            <Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
              Токены персонализации (нажмите для вставки в текст):
            </Text>
            <Space wrap>
              {PERSONALIZATION_TOKENS.map(({ token, label, description }) => (
                <Tooltip key={token} title={description}>
                  <Tag
                    color="blue"
                    style={{ cursor: 'pointer' }}
                    onClick={() => insertToken(token)}
                  >
                    {label}
                  </Tag>
                </Tooltip>
              ))}
            </Space>
          </div>

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
