import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Badge, Popover, List, Button, Typography, Space } from '@/components/design/system'
import { BellOutlined, CheckOutlined } from '@/components/design/icons'
import { App } from '@/components/design/system'
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
import { NOTIFICATION_TYPE_LABELS } from '@/lib/constants'
import { useAuthStore } from '@/stores/auth'
import EmptyState from '@/components/EmptyState'

dayjs.extend(relativeTime)
dayjs.locale('ru')

const { Text } = Typography

function getNotificationsPath(role?: string) {
  if (role === 'client') return '/client/notifications'
  if (role === 'admin') return '/admin/notifications'
  return '/notifications'
}

export default function NotificationBell() {
  const [open, setOpen] = useState(false)
  const navigate = useNavigate()
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const userRole = useAuthStore((s) => s.user?.role)

  const { data: unreadData } = useGetMyNotificationsUnreadCount({
    query: { refetchInterval: 30000 },
  })
  const unreadCount = unreadData?.data?.unread_count ?? 0

  const { data: notificationsData, isLoading } = useGetMyNotifications(
    { page: 1, page_size: 5 },
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
    <div className="rh-notification-popover">
      <div className="rh-notification-popover__head">
        <Text strong className="rh-notification-popover__title">Уведомления</Text>
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
        locale={{ emptyText: <EmptyState description="Нет уведомлений" /> }}
        renderItem={(item) => (
          <List.Item
            className={item.is_read ? 'rh-notification-popover__item' : 'rh-notification-popover__item rh-notification-popover__item--unread'}
            onClick={() => handleItemClick(item.id!, item.is_read ?? false)}
          >
            <List.Item.Meta
              title={
                <Space size={4}>
                  {!item.is_read && (
                    <Badge status="processing" />
                  )}
                  <Text className={item.is_read ? 'rh-notification-popover__item-title' : 'rh-notification-popover__item-title rh-notification-popover__item-title--unread'}>
                    {item.title ?? NOTIFICATION_TYPE_LABELS[item.type ?? ''] ?? 'Уведомление'}
                  </Text>
                </Space>
              }
              description={
                <div>
                  <Text type="secondary" className="rh-notification-popover__body">
                    {item.body}
                  </Text>
                  <br />
                  <Text type="secondary" className="rh-notification-popover__time">
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
        className="rh-section-offset-sm"
        onClick={() => {
          setOpen(false)
          navigate(getNotificationsPath(userRole))
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
        <Button type="text" icon={<BellOutlined className="rh-notification-bell-icon" />} />
      </Badge>
    </Popover>
  )
}
