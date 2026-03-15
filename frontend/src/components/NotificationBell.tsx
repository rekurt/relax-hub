import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Badge, Popover, List, Button, Typography, Space, Empty } from 'antd'
import { BellOutlined, CheckOutlined } from '@ant-design/icons'
import { App } from 'antd'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/ru'
import {
  useGetMyNotifications,
  useGetMyNotificationsUnreadCount,
  usePatchMyNotificationsIdRead,
  usePatchMyNotificationsReadAll,
} from '@/api/generated/notifications/notifications'
import { useQueryClient } from '@tanstack/react-query'

dayjs.extend(relativeTime)
dayjs.locale('ru')

const { Text } = Typography

const notificationTypeLabels: Record<string, string> = {
  booking_new: 'Новое бронирование',
  booking_confirmed: 'Бронирование подтверждено',
  booking_cancelled: 'Бронирование отменено',
  booking_completed: 'Бронирование завершено',
  review_new: 'Новый отзыв',
  payment_received: 'Оплата получена',
  promo_used: 'Промокод использован',
  chat_message: 'Новое сообщение',
}

export default function NotificationBell() {
  const [open, setOpen] = useState(false)
  const navigate = useNavigate()
  const { message } = App.useApp()
  const queryClient = useQueryClient()

  const { data: unreadData } = useGetMyNotificationsUnreadCount({
    query: { refetchInterval: 30000 },
  })
  const unreadCount = unreadData?.data?.unread_count ?? 0

  const { data: notificationsData, isLoading } = useGetMyNotifications(
    { page: 0, page_size: 5 },
    { query: { enabled: open } },
  )
  const notifications = notificationsData?.data ?? []

  const markOneRead = usePatchMyNotificationsIdRead({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['/my/notifications/unread-count'] })
        queryClient.invalidateQueries({ queryKey: ['/my/notifications'] })
      },
    },
  })

  const markAllRead = usePatchMyNotificationsReadAll({
    mutation: {
      onSuccess: () => {
        message.success('Все уведомления прочитаны')
        queryClient.invalidateQueries({ queryKey: ['/my/notifications/unread-count'] })
        queryClient.invalidateQueries({ queryKey: ['/my/notifications'] })
      },
      onError: () => message.error('Ошибка при отметке уведомлений'),
    },
  })

  const handleItemClick = (id: string, isRead: boolean) => {
    if (!isRead) {
      markOneRead.mutate({ id })
    }
  }

  const content = (
    <div style={{ width: 360 }}>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 8,
        }}
      >
        <Text strong>Уведомления</Text>
        {unreadCount > 0 && (
          <Button
            type="link"
            size="small"
            icon={<CheckOutlined />}
            onClick={() => markAllRead.mutate()}
            loading={markAllRead.isPending}
          >
            Прочитать все
          </Button>
        )}
      </div>

      <List
        loading={isLoading}
        dataSource={notifications}
        locale={{ emptyText: <Empty description="Нет уведомлений" image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
        renderItem={(item) => (
          <List.Item
            style={{
              cursor: 'pointer',
              background: item.is_read ? undefined : 'rgba(22, 119, 255, 0.04)',
              padding: '8px 12px',
              borderRadius: 6,
            }}
            onClick={() => handleItemClick(item.id!, item.is_read ?? false)}
          >
            <List.Item.Meta
              title={
                <Space size={4}>
                  {!item.is_read && (
                    <Badge status="processing" />
                  )}
                  <Text style={{ fontSize: 13 }}>
                    {item.title ?? notificationTypeLabels[item.type ?? ''] ?? 'Уведомление'}
                  </Text>
                </Space>
              }
              description={
                <div>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    {item.body}
                  </Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: 11 }}>
                    {item.created_at ? dayjs(item.created_at).fromNow() : ''}
                  </Text>
                </div>
              }
            />
          </List.Item>
        )}
      />

      <Button
        type="link"
        block
        style={{ marginTop: 8 }}
        onClick={() => {
          setOpen(false)
          navigate('/notifications')
        }}
      >
        Все уведомления
      </Button>
    </div>
  )

  return (
    <Popover
      content={content}
      trigger="click"
      open={open}
      onOpenChange={setOpen}
      placement="bottomRight"
      arrow={false}
    >
      <Badge count={unreadCount} size="small" offset={[-2, 2]}>
        <Button type="text" icon={<BellOutlined style={{ fontSize: 18 }} />} />
      </Badge>
    </Popover>
  )
}
