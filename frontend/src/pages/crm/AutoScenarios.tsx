import {
  App,
  Card,
  Col,
  Input,
  InputNumber,
  Row,
  Select,
  Switch,
  Typography,
} from '@/components/design/system'
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  GiftOutlined,
  HeartOutlined,
  StarOutlined,
} from '@/components/design/icons'
import {
  useGetMyCrmAutoScenarios,
  usePutMyCrmAutoScenariosType,
} from '@/api/generated/crm/crm'
import type { InternalHandlerAutoScenarioResponse } from '@/api/generated/model'
import { useQueryClient } from '@tanstack/react-query'
import EmptyState from '@/components/EmptyState'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

const SCENARIO_ICONS: Record<string, React.ReactNode> = {
  thank_after_visit: <HeartOutlined className="rh-scenario-icon--warning" />,
  request_review: <StarOutlined className="rh-scenario-icon--warning" />,
  remind_revisit_30d: <ClockCircleOutlined className="rh-scenario-icon--primary" />,
  reactivate_lost_90d: <CheckCircleOutlined className="rh-scenario-icon--success" />,
  birthday_greeting: <GiftOutlined className="rh-scenario-icon--teal" />,
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
    <div className="rh-stack">
      <PageHeader
        eyebrow="CRM"
        title="Автоматические сценарии"
        description="Настройте автоматическую отправку сообщений гостям. Сценарии выполняются при выполнении условий."
        size="compact"
      />

      <Row gutter={[16, 16]}>
        {isLoading ? (
          <Col span={24}>
            <Card loading />
          </Col>
        ) : scenarios.length === 0 ? (
          <Col span={24}>
            <EmptyState
              description="Нет настроенных сценариев. Автоматические сценарии появятся после подключения CRM"
            />
          </Col>
        ) : (
          scenarios.map((scenario) => (
            <Col xs={24} key={scenario.type}>
              <Card
                className={scenario.enabled ? 'rh-scenario-card rh-scenario-card--enabled' : 'rh-scenario-card'}
              >
                <div className="rh-scenario-card__header">
                  <div className="rh-scenario-card__title-row">
                    <span className="rh-scenario-card__icon">
                      {SCENARIO_ICONS[scenario.type ?? ''] ?? <ClockCircleOutlined />}
                    </span>
                    <div>
                      <Text strong className="rh-scenario-card__title">{scenario.name}</Text>
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
                  <Row gutter={16} className="rh-section-offset">
                    <Col xs={24} sm={8}>
                      <Text type="secondary" className="rh-field-label">
                        Канал
                      </Text>
                      <Select
                        value={scenario.channel}
                        onChange={(v) => handleUpdate(scenario, 'channel', v)}
                        options={CHANNEL_OPTIONS}
                        className="rh-full-width"
                        size="small"
                      />
                    </Col>
                    <Col xs={24} sm={8}>
                      <Text type="secondary" className="rh-field-label">
                        Задержка (часов)
                      </Text>
                      <InputNumber
                        value={scenario.delay_hours}
                        onChange={(v) => v !== null && handleUpdate(scenario, 'delay_hours', v)}
                        min={0}
                        max={8760}
                        className="rh-full-width"
                        size="small"
                      />
                    </Col>
                    <Col xs={24} sm={24} className="rh-section-offset-sm">
                      <Text type="secondary" className="rh-field-label">
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
