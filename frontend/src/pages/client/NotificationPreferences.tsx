import { useEffect, useState, type ReactNode } from 'react'
import { Typography, Card, Switch, Table, Tag, Button, Space, App, Spin, Divider } from '@/components/design/system'
import {
  BellOutlined,
  MailOutlined,
  MobileOutlined,
  MessageOutlined,
} from '@/components/design/icons'
import {
  useGetMyNotificationPreferences,
  usePutMyNotificationPreferences,
} from '@/api/generated/notifications/notifications'
import { axiosInstance } from '@/api/axios-instance'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

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

function ChannelToggleCard({
  icon,
  title,
  description,
  checked,
  disabled,
  onChange,
}: {
  icon: ReactNode
  title: string
  description: string
  checked: boolean
  disabled?: boolean
  onChange?: (checked: boolean) => void
}) {
  return (
    <div className="rh-toggle-card">
      <div className="rh-toggle-card__copy">
        <Space size={8}>
          {icon}
          <Text className="rh-toggle-card__title">{title}</Text>
        </Space>
        <Text className="rh-toggle-card__description">{description}</Text>
      </div>
      <Switch checked={checked} disabled={disabled} onChange={onChange} />
    </div>
  )
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
    return (
      <div className="rh-fullscreen-state">
        <Spin size="large" />
      </div>
    )
  }

  const eventPrefMap = new Map(localEventPrefs.map((p) => [p.event_type, p]))
  const activeChannels = [globalPrefs?.push, globalPrefs?.email, globalPrefs?.sms, globalPrefs?.telegram].filter(Boolean).length
  const mandatoryEvents = localEventPrefs.filter((event) => event.is_mandatory).length

  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="Личные настройки"
        title="Настройки уведомлений"
        description="Управляйте каналами доставки и типами событий без лишнего шума."
      />

      <div className="rh-stat-grid">
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Активных каналов</span>
          <span className="rh-stat-tile__value">{activeChannels}</span>
          <span className="rh-stat-tile__hint">Push, email, SMS и Telegram с текущим состоянием на аккаунте.</span>
        </div>
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Событий под настройку</span>
          <span className="rh-stat-tile__value">{localEventPrefs.length}</span>
          <span className="rh-stat-tile__hint">Можно гибко включать каналы для каждого типа события.</span>
        </div>
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Обязательных событий</span>
          <span className="rh-stat-tile__value">{mandatoryEvents}</span>
          <span className="rh-stat-tile__hint">Системные уведомления, которые нельзя полностью отключить.</span>
        </div>
      </div>

      <Card title="Каналы доставки">
        <Text type="secondary" className="rh-card-intro-text">
          Включите или выключите каналы доставки уведомлений
        </Text>
        <div className="rh-toggle-grid">
          <ChannelToggleCard
            icon={<BellOutlined />}
            title="Push-уведомления"
            description="Быстрые уведомления в интерфейсе и на устройстве."
            checked={globalPrefs?.push ?? false}
            onChange={(checked) => handleGlobalToggle('push', checked)}
          />
          <ChannelToggleCard
            icon={<MailOutlined />}
            title="Email"
            description="Письма для подтверждений, итогов и важных изменений."
            checked={globalPrefs?.email ?? true}
            onChange={(checked) => handleGlobalToggle('email', checked)}
          />
          <ChannelToggleCard
            icon={<MobileOutlined />}
            title="SMS"
            description="Короткие уведомления для критичных сценариев и напоминаний."
            checked={globalPrefs?.sms ?? false}
            onChange={(checked) => handleGlobalToggle('sms', checked)}
          />
          <ChannelToggleCard
            icon={<MessageOutlined />}
            title="Telegram"
            description="Канал подготовлен, но пока недоступен для самостоятельного включения."
            checked={globalPrefs?.telegram ?? false}
            disabled
          />
        </div>
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
        <Text type="secondary" className="rh-card-intro-text">
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
                        <Tag color="red" className="rh-micro-tag">
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
