import { useState } from 'react'
import { Typography, List, Button, Badge, Space, Empty, Card, App } from 'antd'
import { CheckOutlined, BellOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/ru'
import {
  useGetMyNotifications,
  usePatchMyNotificationsIdRead,
  usePatchMyNotificationsReadAll,
} from '@/api/generated/notifications/notifications'
import { useQueryClient } from '@tanstack/react-query'
import { NOTIFICATION_TYPE_LABELS } from '@/lib/constants'
import PageHeader from '@/components/PageHeader'

dayjs.extend(relativeTime)
dayjs.locale('ru')

const { Text } = Typography

export default function NotificationList() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const { message } = App.useApp()
  const queryClient = useQueryClient()

  const { data, isLoading } = useGetMyNotifications({
    page,
    page_size: pageSize,
  })

  const notifications = data?.data ?? []
  const meta = data?.meta

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

  return (
    <div>
      <PageHeader
        eyebrow="Лента"
        title="Уведомления"
        description="Все события собраны в одном месте. Непрочитанные помечены акцентом и доступны для быстрого чтения."
        extra={
          <Button
            icon={<CheckOutlined />}
            onClick={() => markAllRead.mutate()}
            loading={markAllRead.isPending}
          >
            Прочитать все
          </Button>
        }
      />

      <Card>
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
                background: item.is_read ? undefined : 'rgba(15, 118, 110, 0.05)',
                padding: '14px 16px',
                borderRadius: 14,
                border: '1px solid rgba(15, 23, 42, 0.06)',
                marginBottom: 8,
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
    </div>
  )
}
