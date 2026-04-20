import { useState, useEffect } from 'react'
import { Typography, List, Button, Badge, Space, Empty, Card, App, Form, Switch, Skeleton } from 'antd'
import { CheckOutlined, BellOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/ru'
import {
  useGetMyNotifications,
  usePatchMyNotificationsIdRead,
  usePatchMyNotificationsReadAll,
  useGetMyNotificationPreferences,
  usePutMyNotificationPreferences,
} from '@/api/generated/notifications/notifications'
import { useQueryClient } from '@tanstack/react-query'
import { NOTIFICATION_TYPE_LABELS } from '@/lib/constants'
import PageHeader from '@/components/PageHeader'

dayjs.extend(relativeTime)
dayjs.locale('ru')

const { Text } = Typography

const PREFERENCE_CARDS = [
  {
    name: 'in_app',
    title: 'В приложении',
    description: 'Служебные события прямо в интерфейсе админки.',
  },
  {
    name: 'email',
    title: 'Email',
    description: 'Дублирование важных событий в почту.',
  },
  {
    name: 'push',
    title: 'Push-уведомления',
    description: 'Короткие сигналы по критичным операциям.',
  },
  {
    name: 'booking_events',
    title: 'Бронирования',
    description: 'Изменения по заказам и операционным действиям.',
  },
  {
    name: 'review_events',
    title: 'Отзывы',
    description: 'Новые отзывы, модерация и ответы.',
  },
  {
    name: 'promo_events',
    title: 'Промокоды',
    description: 'События по акциям и служебным рассылкам.',
  },
  {
    name: 'reminders',
    title: 'Напоминания',
    description: 'Дедлайны и системные триггеры платформы.',
  },
] as const

export default function AdminNotifications() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [prefsForm] = Form.useForm()

  const { data, isLoading } = useGetMyNotifications({
    page,
    page_size: pageSize,
  })

  const notifications = data?.data ?? []
  const meta = data?.meta
  const unreadCount = notifications.filter((item) => !item.is_read).length

  const { data: prefsData, isLoading: prefsLoading } = useGetMyNotificationPreferences()
  const prefs = prefsData?.data

  useEffect(() => {
    if (prefs) {
      prefsForm.setFieldsValue({
        in_app: prefs.in_app ?? true,
        email: prefs.email ?? false,
        push: prefs.push ?? false,
        booking_events: prefs.booking_events ?? true,
        review_events: prefs.review_events ?? true,
        promo_events: prefs.promo_events ?? true,
        reminders: prefs.reminders ?? true,
      })
    }
  }, [prefs, prefsForm])

  const markOneRead = usePatchMyNotificationsIdRead({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['/my/notifications'] })
        queryClient.invalidateQueries({ queryKey: ['/my/notifications/unread-count'] })
      },
    },
  })

  const markAllRead = usePatchMyNotificationsReadAll({
    mutation: {
      onSuccess: () => {
        message.success('Все уведомления прочитаны')
        queryClient.invalidateQueries({ queryKey: ['/my/notifications'] })
        queryClient.invalidateQueries({ queryKey: ['/my/notifications/unread-count'] })
      },
      onError: () => message.error('Ошибка при отметке уведомлений'),
    },
  })

  const updatePrefs = usePutMyNotificationPreferences({
    mutation: {
      onSuccess: () => {
        message.success('Настройки уведомлений сохранены')
        queryClient.invalidateQueries({ queryKey: ['/my/notification-preferences'] })
      },
      onError: () => message.error('Ошибка при сохранении настроек'),
    },
  })

  const handlePrefsSubmit = (values: Record<string, boolean>) => {
    updatePrefs.mutate({ data: values })
  }

  return (
    <div className="bani-stack">
      <PageHeader
        eyebrow="Служебные события"
        title="Уведомления"
        description="Рабочая лента уведомлений и базовые настройки каналов доставки."
      />

      <div className="bani-stat-grid">
        <div className="bani-stat-tile">
          <span className="bani-stat-tile__eyebrow">Всего в ленте</span>
          <span className="bani-stat-tile__value">{meta?.total_count ?? notifications.length}</span>
          <span className="bani-stat-tile__hint">Текущий объём уведомлений на выбранной странице.</span>
        </div>
        <div className="bani-stat-tile">
          <span className="bani-stat-tile__eyebrow">Непрочитанных</span>
          <span className="bani-stat-tile__value">{unreadCount}</span>
          <span className="bani-stat-tile__hint">Их стоит разбирать в первую очередь.</span>
        </div>
        <div className="bani-stat-tile">
          <span className="bani-stat-tile__eyebrow">Активные каналы</span>
          <span className="bani-stat-tile__value">
            {[prefs?.in_app, prefs?.email, prefs?.push].filter(Boolean).length}
          </span>
          <span className="bani-stat-tile__hint">Активные каналы для служебных событий.</span>
        </div>
      </div>

      <div className="bani-grid bani-grid--content-aside">
        <Card
          title="Список уведомлений"
          extra={
            <Button
              icon={<CheckOutlined />}
              onClick={() => markAllRead.mutate()}
              loading={markAllRead.isPending}
            >
              Прочитать все
            </Button>
          }
        >
          <List
            loading={isLoading}
            dataSource={notifications}
            locale={{
              emptyText: (
                <Empty
                  image={<BellOutlined style={{ fontSize: 48, color: '#d9d9d9' }} />}
                  description="Нет уведомлений"
                />
              ),
            }}
            pagination={{
              current: page,
              pageSize,
              total: meta?.total_count ?? 0,
              onChange: (nextPage, nextPageSize) => {
                setPage(nextPage)
                setPageSize(nextPageSize)
              },
              showSizeChanger: true,
              showTotal: (total) => `Всего: ${total}`,
            }}
            renderItem={(item) => (
              <List.Item
                className={`bani-feed-item ${item.is_read ? '' : 'bani-feed-item--unread'}`.trim()}
                onClick={() => {
                  if (!item.is_read && item.id) {
                    markOneRead.mutate({ id: item.id })
                  }
                }}
                actions={
                  !item.is_read
                    ? [
                        <Button
                          key="read"
                          type="link"
                          size="small"
                          onClick={(event) => {
                            event.stopPropagation()
                            if (item.id) {
                              markOneRead.mutate({ id: item.id })
                            }
                          }}
                        >
                          Прочитать
                        </Button>,
                      ]
                    : undefined
                }
              >
                <div className="bani-feed-item__main">
                  <span className="bani-feed-item__badge" />
                  <div className="bani-feed-item__copy">
                    <Text strong={!item.is_read}>
                      {item.title ?? NOTIFICATION_TYPE_LABELS[item.type ?? ''] ?? 'Уведомление'}
                    </Text>
                    <Text type="secondary">{item.body}</Text>
                    <Text className="bani-feed-item__meta">
                      {item.created_at ? dayjs(item.created_at).fromNow() : ''}
                    </Text>
                  </div>
                </div>
                <Space>
                  <Badge status={item.is_read ? 'default' : 'processing'} />
                </Space>
              </List.Item>
            )}
          />
        </Card>

        <Card title="Настройки уведомлений">
          {prefsLoading ? (
            <Skeleton active />
          ) : (
            <Form form={prefsForm} layout="vertical" onFinish={handlePrefsSubmit}>
              <Text type="secondary" style={{ display: 'block', marginBottom: 16 }}>
                Каналы доставки
              </Text>

              <div className="bani-toggle-grid" style={{ marginBottom: 20 }}>
                {PREFERENCE_CARDS.map((item) => (
                  <div key={item.name} className="bani-toggle-card">
                    <div className="bani-toggle-card__copy">
                      <Text className="bani-toggle-card__title">{item.title}</Text>
                      <Text className="bani-toggle-card__description">
                        {item.description}
                      </Text>
                    </div>
                    <Form.Item name={item.name} valuePropName="checked" style={{ marginBottom: 0 }}>
                      <Switch />
                    </Form.Item>
                  </div>
                ))}
              </div>

              <Text type="secondary" style={{ display: 'block', marginBottom: 12 }}>
                События
              </Text>

              <Form.Item>
                <Button type="primary" htmlType="submit" loading={updatePrefs.isPending}>
                  Сохранить настройки
                </Button>
              </Form.Item>
            </Form>
          )}
        </Card>
      </div>
    </div>
  )
}
