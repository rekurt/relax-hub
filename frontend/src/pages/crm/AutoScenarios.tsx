import {
  App,
  Card,
  Col,
  Empty,
  Input,
  InputNumber,
  Row,
  Select,
  Switch,
  Typography,
} from 'antd'
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  GiftOutlined,
  HeartOutlined,
  StarOutlined,
} from '@ant-design/icons'
import {
  useGetMyCrmAutoScenarios,
  usePutMyCrmAutoScenariosType,
} from '@/api/generated/crm/crm'
import type { InternalHandlerAutoScenarioResponse } from '@/api/generated/model'
import { useQueryClient } from '@tanstack/react-query'

const { Title, Text, Paragraph } = Typography

const SCENARIO_ICONS: Record<string, React.ReactNode> = {
  thank_after_visit: <HeartOutlined style={{ color: '#eb2f96' }} />,
  request_review: <StarOutlined style={{ color: '#faad14' }} />,
  remind_revisit_30d: <ClockCircleOutlined style={{ color: '#1677ff' }} />,
  reactivate_lost_90d: <CheckCircleOutlined style={{ color: '#52c41a' }} />,
  birthday_greeting: <GiftOutlined style={{ color: '#722ed1' }} />,
}

const CHANNEL_OPTIONS = [
  { value: 'push', label: 'Push' },
  { value: 'email', label: 'Email' },
  { value: 'telegram', label: 'Telegram' },
]

export default function AutoScenarios() {
  const queryClient = useQueryClient()
  const { message } = App.useApp()

  const { data, isLoading } = useGetMyCrmAutoScenarios()
  const scenarios = data?.data ?? []

  const updateMutation = usePutMyCrmAutoScenariosType({
    mutation: {
      onSuccess: () => {
        message.success('Сценарий обновлён')
        queryClient.invalidateQueries({ queryKey: ['/my/crm/auto-scenarios'] })
      },
      onError: () => message.error('Не удалось обновить сценарий'),
    },
  })

  const handleUpdate = (
    scenario: InternalHandlerAutoScenarioResponse,
    field: string,
    value: unknown,
  ) => {
    if (!scenario.type) return
    updateMutation.mutate({
      type: scenario.type,
      data: {
        enabled: scenario.enabled,
        custom_text: scenario.custom_text,
        channel: scenario.channel,
        delay_hours: scenario.delay_hours,
        ...(field === 'enabled' && { enabled: value as boolean }),
        ...(field === 'custom_text' && { custom_text: value as string }),
        ...(field === 'channel' && { channel: value as string }),
        ...(field === 'delay_hours' && { delay_hours: value as number }),
      },
    })
  }

  return (
    <div>
      <Title level={3}>Автоматические сценарии</Title>
      <Paragraph type="secondary" style={{ marginBottom: 24 }}>
        Настройте автоматическую отправку сообщений гостям. Сценарии выполняются автоматически при выполнении условий.
      </Paragraph>

      <Row gutter={[16, 16]}>
        {isLoading ? (
          <Col span={24}>
            <Card loading />
          </Col>
        ) : scenarios.length === 0 ? (
          <Col span={24}>
            <Empty
              description="Нет настроенных сценариев. Автоматические сценарии появятся после подключения CRM"
              style={{ padding: 48 }}
            />
          </Col>
        ) : (
          scenarios.map((scenario) => (
            <Col xs={24} key={scenario.type}>
              <Card
                style={{
                  borderLeft: `4px solid ${scenario.enabled ? '#52c41a' : '#d9d9d9'}`,
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 12 }}>
                  <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
                    <span style={{ fontSize: 24 }}>
                      {SCENARIO_ICONS[scenario.type ?? ''] ?? <ClockCircleOutlined />}
                    </span>
                    <div>
                      <Text strong style={{ fontSize: 16 }}>{scenario.name}</Text>
                      <br />
                      <Text type="secondary">{scenario.description}</Text>
                    </div>
                  </div>
                  <Switch
                    checked={scenario.enabled}
                    onChange={(checked) => handleUpdate(scenario, 'enabled', checked)}
                  />
                </div>

                {scenario.enabled && (
                  <Row gutter={16} style={{ marginTop: 16 }}>
                    <Col xs={24} sm={8}>
                      <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>
                        Канал
                      </Text>
                      <Select
                        value={scenario.channel}
                        onChange={(v) => handleUpdate(scenario, 'channel', v)}
                        options={CHANNEL_OPTIONS}
                        style={{ width: '100%' }}
                        size="small"
                      />
                    </Col>
                    <Col xs={24} sm={8}>
                      <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>
                        Задержка (часов)
                      </Text>
                      <InputNumber
                        value={scenario.delay_hours}
                        onChange={(v) => v !== null && handleUpdate(scenario, 'delay_hours', v)}
                        min={0}
                        max={8760}
                        style={{ width: '100%' }}
                        size="small"
                      />
                    </Col>
                    <Col xs={24} sm={24} style={{ marginTop: 8 }}>
                      <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>
                        Кастомный текст
                      </Text>
                      <Input.TextArea
                        value={scenario.custom_text}
                        onBlur={(e) => handleUpdate(scenario, 'custom_text', e.target.value)}
                        rows={2}
                        placeholder="Оставьте пустым для текста по умолчанию"
                        size="small"
                      />
                    </Col>
                  </Row>
                )}
              </Card>
            </Col>
          ))
        )}
      </Row>
    </div>
  )
}
