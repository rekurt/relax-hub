import { useState, useEffect } from 'react'
import { Typography, List, Button, Badge, Space, Empty, Card, App, Form, Switch, Divider, Skeleton } from 'antd'
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
    <div>
      <PageHeader
        eyebrow="Служебные события"
        title="Уведомления"
        description="Рабочая лента уведомлений и базовые настройки каналов доставки."
      />

      <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
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
              pageSize: pageSize,
              total: meta?.total_count ?? 0,
              onChange: (p, ps) => {
                setPage(p)
                setPageSize(ps)
              },
              showSizeChanger: true,
              showTotal: (total) => `Всего: ${total}`,
            }}
            renderItem={(item) => (
              <List.Item
                style={{
                  cursor: item.is_read ? 'default' : 'pointer',
                  background: item.is_read ? undefined : 'rgba(22, 119, 255, 0.04)',
                  padding: '12px 16px',
                  borderRadius: 6,
                }}
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
                          onClick={(e) => {
                            e.stopPropagation()
                            if (item.id) markOneRead.mutate({ id: item.id })
                          }}
                        >
                          Прочитать
                        </Button>,
                      ]
                    : undefined
                }
              >
                <List.Item.Meta
                  avatar={
                    <Space>
                      {!item.is_read && <Badge status="processing" />}
                      {item.is_read && <Badge status="default" />}
                    </Space>
                  }
                  title={
                    <Text strong={!item.is_read}>
                      {item.title ?? NOTIFICATION_TYPE_LABELS[item.type ?? ''] ?? 'Уведомление'}
                    </Text>
                  }
                  description={
                    <div>
                      <Text type="secondary">{item.body}</Text>
                      <br />
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        {item.created_at ? dayjs(item.created_at).fromNow() : ''}
                      </Text>
                    </div>
                  }
                />
              </List.Item>
            )}
          />
        </Card>

        <Card title="Настройки уведомлений">
          {prefsLoading ? (
            <Skeleton active />
          ) : (
            <Form
              form={prefsForm}
              layout="vertical"
              onFinish={handlePrefsSubmit}
              style={{ maxWidth: 500 }}
            >
              <Divider plain>Каналы доставки</Divider>
              <Form.Item label="В приложении" name="in_app" valuePropName="checked">
                <Switch />
              </Form.Item>
              <Form.Item label="Email" name="email" valuePropName="checked">
                <Switch />
              </Form.Item>
              <Form.Item label="Push-уведомления" name="push" valuePropName="checked">
                <Switch />
              </Form.Item>

              <Divider plain>События</Divider>
              <Form.Item label="Бронирования" name="booking_events" valuePropName="checked">
                <Switch />
              </Form.Item>
              <Form.Item label="Отзывы" name="review_events" valuePropName="checked">
                <Switch />
              </Form.Item>
              <Form.Item label="Промокоды" name="promo_events" valuePropName="checked">
                <Switch />
              </Form.Item>
              <Form.Item label="Напоминания" name="reminders" valuePropName="checked">
                <Switch />
              </Form.Item>

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
