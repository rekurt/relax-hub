import { useEffect, useState } from 'react'
import { Typography, Card, Switch, Table, Tag, Button, Space, App, Spin, Divider } from 'antd'
import {
  BellOutlined,
  MailOutlined,
  MobileOutlined,
  MessageOutlined,
} from '@ant-design/icons'
import {
  useGetMyNotificationPreferences,
  usePutMyNotificationPreferences,
} from '@/api/generated/notifications/notifications'
import { axiosInstance } from '@/api/axios-instance'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import PageHeader from '@/components/PageHeader'

const { Title, Text } = Typography

interface EventPreference {
  event_type: string
  push_enabled: boolean
  email_enabled: boolean
  sms_enabled: boolean
  is_mandatory: boolean
}

const EVENT_LABELS: Record<string, string> = {
  booking_confirmed: 'Подтверждение бронирования',
  booking_cancelled: 'Отмена бронирования',
  booking_rejected: 'Отклонение бронирования',
  booking_request: 'Запрос на бронирование',
  booking_reminder_24h: 'Напоминание за 24 часа',
  booking_reminder_2h: 'Напоминание за 2 часа',
  booking_checked_in: 'Отметка заезда',
  booking_no_show: 'Неявка',
  booking_extended: 'Продление бронирования',
  new_review: 'Новый отзыв',
  review_response: 'Ответ на отзыв',
  review_approved: 'Одобрение отзыва',
  review_rejected: 'Отклонение отзыва',
  review_request: 'Запрос отзыва',
  promo: 'Промоакции',
  broadcast: 'Рассылки',
  auto_scenario: 'Автосценарии',
  new_message: 'Новое сообщение',
  loyalty_upgrade: 'Повышение уровня лояльности',
  referral_bonus: 'Реферальный бонус',
  bonus_expiring: 'Истечение бонусов',
  subscription_expiring: 'Истечение подписки',
  photo_verified: 'Фото верифицировано',
  photo_rejected: 'Фото отклонено',
  saved_search_match: 'Совпадение по сохранённому поиску',
  dispute_resolution: 'Решение по спору',
  payment_received: 'Получение платежа',
  system: 'Системные уведомления',
}

const EVENT_CATEGORIES: Record<string, string[]> = {
  'Бронирования': [
    'booking_confirmed', 'booking_cancelled', 'booking_rejected',
    'booking_request', 'booking_reminder_24h', 'booking_reminder_2h',
    'booking_checked_in', 'booking_no_show', 'booking_extended',
  ],
  'Отзывы': [
    'new_review', 'review_response', 'review_approved',
    'review_rejected', 'review_request',
  ],
  'Маркетинг': ['promo', 'broadcast', 'auto_scenario'],
  'Сообщения': ['new_message'],
  'Бонусы и лояльность': [
    'loyalty_upgrade', 'referral_bonus', 'bonus_expiring',
    'subscription_expiring',
  ],
  'Фото': ['photo_verified', 'photo_rejected'],
  'Прочее': [
    'saved_search_match', 'dispute_resolution', 'payment_received', 'system',
  ],
}

function useEventPreferences() {
  return useQuery({
    queryKey: ['notification-event-preferences'],
    queryFn: async () => {
      const { data } = await axiosInstance.get<{ success: boolean; data: EventPreference[] }>(
        '/my/notification-preferences/events'
      )
      return data.data
    },
  })
}

function useUpdateEventPreferences() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (preferences: EventPreference[]) => {
      const { data } = await axiosInstance.put('/my/notification-preferences/events', {
        preferences: preferences.map((p) => ({
          event_type: p.event_type,
          push_enabled: p.push_enabled,
          email_enabled: p.email_enabled,
          sms_enabled: p.sms_enabled,
        })),
      })
      return data.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notification-event-preferences'] })
    },
  })
}

export default function NotificationPreferences() {
  const { message } = App.useApp()
  const queryClient = useQueryClient()

  const { data: globalPrefsData, isLoading: globalLoading } = useGetMyNotificationPreferences()
  const globalPrefs = (globalPrefsData as { data?: Record<string, boolean> })?.data
  const updateGlobalMutation = usePutMyNotificationPreferences({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['/my/notification-preferences'] })
        message.success('Настройки каналов сохранены')
      },
    },
  })

  const { data: eventPrefs, isLoading: eventsLoading } = useEventPreferences()
  const updateEventsMutation = useUpdateEventPreferences()

  const [localEventPrefs, setLocalEventPrefs] = useState<EventPreference[]>([])
  const [dirty, setDirty] = useState(false)

  useEffect(() => {
    if (eventPrefs) {
      setLocalEventPrefs(eventPrefs) // eslint-disable-line react-hooks/set-state-in-effect -- sync local state from server data
      setDirty(false)
    }
  }, [eventPrefs])

  const handleGlobalToggle = (field: string, checked: boolean) => {
    if (!globalPrefs) return
    updateGlobalMutation.mutate({
      data: { [field]: checked },
    })
  }

  const handleEventToggle = (
    eventType: string,
    channel: 'push_enabled' | 'email_enabled' | 'sms_enabled',
    checked: boolean
  ) => {
    setLocalEventPrefs((prev) =>
      prev.map((p) =>
        p.event_type === eventType ? { ...p, [channel]: checked } : p
      )
    )
    setDirty(true)
  }

  const handleSaveEvents = () => {
    updateEventsMutation.mutate(localEventPrefs, {
      onSuccess: () => {
        message.success('Настройки уведомлений сохранены')
        setDirty(false)
      },
      onError: () => {
        message.error('Ошибка сохранения настроек')
      },
    })
  }

  if (globalLoading || eventsLoading) {
    return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  }

  const eventPrefMap = new Map(localEventPrefs.map((p) => [p.event_type, p]))

  return (
    <div>
      <PageHeader
        eyebrow="Личные настройки"
        title="Настройки уведомлений"
        description="Управляйте каналами доставки и типами событий без лишнего шума."
      />

      <Card title="Каналы доставки" style={{ marginBottom: 24 }}>
        <Text type="secondary" style={{ display: 'block', marginBottom: 16 }}>
          Включите или выключите каналы доставки уведомлений
        </Text>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Space>
              <BellOutlined />
              <Text>Push-уведомления</Text>
            </Space>
            <Switch
              checked={globalPrefs?.push ?? false}
              onChange={(checked) => handleGlobalToggle('push', checked)}
            />
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Space>
              <MailOutlined />
              <Text>Email</Text>
            </Space>
            <Switch
              checked={globalPrefs?.email ?? true}
              onChange={(checked) => handleGlobalToggle('email', checked)}
            />
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Space>
              <MobileOutlined />
              <Text>SMS</Text>
            </Space>
            <Switch
              checked={globalPrefs?.sms ?? false}
              onChange={(checked) => handleGlobalToggle('sms', checked)}
            />
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Space>
              <MessageOutlined />
              <Text>Telegram</Text>
            </Space>
            <Switch
              checked={globalPrefs?.telegram ?? false}
              disabled
            />
          </div>
        </Space>
      </Card>

      <Card
        title="Настройки по типам событий"
        extra={
          dirty && (
            <Button
              type="primary"
              onClick={handleSaveEvents}
              loading={updateEventsMutation.isPending}
            >
              Сохранить
            </Button>
          )
        }
      >
        <Text type="secondary" style={{ display: 'block', marginBottom: 16 }}>
          Настройте каналы для каждого типа уведомлений. Обязательные уведомления нельзя отключить.
        </Text>

        {Object.entries(EVENT_CATEGORIES).map(([category, events]) => (
          <div key={category}>
            <Divider>{category}</Divider>
            <Table
              dataSource={events
                .map((et) => eventPrefMap.get(et))
                .filter((p): p is EventPreference => !!p)}
              rowKey="event_type"
              pagination={false}
              size="small"
              columns={[
                {
                  title: 'Событие',
                  dataIndex: 'event_type',
                  render: (et: string, record: EventPreference) => (
                    <Space>
                      <Text>{EVENT_LABELS[et] ?? et}</Text>
                      {record.is_mandatory && (
                        <Tag color="red" style={{ fontSize: 10 }}>
                          обязательно
                        </Tag>
                      )}
                    </Space>
                  ),
                },
                {
                  title: 'Push',
                  dataIndex: 'push_enabled',
                  width: 80,
                  align: 'center' as const,
                  render: (_: boolean, record: EventPreference) => (
                    <Switch
                      size="small"
                      checked={record.push_enabled}
                      disabled={record.is_mandatory}
                      onChange={(checked) =>
                        handleEventToggle(record.event_type, 'push_enabled', checked)
                      }
                    />
                  ),
                },
                {
                  title: 'Email',
                  dataIndex: 'email_enabled',
                  width: 80,
                  align: 'center' as const,
                  render: (_: boolean, record: EventPreference) => (
                    <Switch
                      size="small"
                      checked={record.email_enabled}
                      disabled={record.is_mandatory}
                      onChange={(checked) =>
                        handleEventToggle(record.event_type, 'email_enabled', checked)
                      }
                    />
                  ),
                },
                {
                  title: 'SMS',
                  dataIndex: 'sms_enabled',
                  width: 80,
                  align: 'center' as const,
                  render: (_: boolean, record: EventPreference) => (
                    <Switch
                      size="small"
                      checked={record.sms_enabled}
                      onChange={(checked) =>
                        handleEventToggle(record.event_type, 'sms_enabled', checked)
                      }
                    />
                  ),
                },
              ]}
            />
          </div>
        ))}
      </Card>
    </div>
  )
}
