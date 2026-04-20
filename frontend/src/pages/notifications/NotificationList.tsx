import { useMemo, useState } from 'react'
import { App, Button, Card, Empty, List, Pagination, Typography } from 'antd'
import { CheckOutlined, BellOutlined } from '@ant-design/icons'
import { useQueryClient } from '@tanstack/react-query'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/ru'
import {
  useGetMyNotifications,
  usePatchMyNotificationsIdRead,
  usePatchMyNotificationsReadAll,
} from '@/api/generated/notifications/notifications'
import PageHeader from '@/components/PageHeader'
import { NOTIFICATION_TYPE_LABELS } from '@/lib/constants'

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

  const unreadCount = useMemo(
    () => notifications.filter((notification) => !notification.is_read).length,
    [notifications],
  )

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
    <div className="bani-stack">
      <PageHeader
        eyebrow="Лента"
        title="Уведомления"
        description="Уведомления сгруппированы как рабочая лента: непрочитанные стоят выше по вниманию, а массовое чтение доступно без лишних переходов."
        extra={(
          <Button
            icon={<CheckOutlined />}
            onClick={() => markAllRead.mutate()}
            loading={markAllRead.isPending}
          >
            Прочитать все
          </Button>
        )}
      />

      <section className="bani-hero-panel">
        <div className="bani-hero-panel__eyebrow">Сводка</div>
        <h2 className="bani-hero-panel__title">Лента сделана как очередь внимания, а не как сырой список</h2>
        <div className="bani-hero-panel__description">
          Пользователь сразу понимает, сколько сообщений требуют реакции прямо сейчас и сколько уже обработано. Это удобнее, чем длинная плоская лента без приоритетов.
        </div>
        <div className="bani-stat-grid">
          <div className="bani-stat-tile">
            <span className="bani-stat-tile__eyebrow">Непрочитанные</span>
            <div className="bani-stat-tile__value">{unreadCount}</div>
            <span className="bani-stat-tile__hint">Сообщения, которые требуют просмотра</span>
          </div>
          <div className="bani-stat-tile">
            <span className="bani-stat-tile__eyebrow">Всего на странице</span>
            <div className="bani-stat-tile__value">{notifications.length}</div>
            <span className="bani-stat-tile__hint">Текущий блок ленты с пагинацией</span>
          </div>
          <div className="bani-stat-tile">
            <span className="bani-stat-tile__eyebrow">Всего в истории</span>
            <div className="bani-stat-tile__value">{meta?.total_count ?? 0}</div>
            <span className="bani-stat-tile__hint">Полный объём уведомлений аккаунта</span>
          </div>
        </div>
      </section>

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
          renderItem={(item) => (
            <List.Item style={{ padding: 0, border: 0, marginBottom: 12 }}>
              <div
                className={`bani-feed-item${item.is_read ? '' : ' bani-feed-item--unread'}`}
                style={{ width: '100%', cursor: item.is_read ? 'default' : 'pointer' }}
                onClick={() => {
                  if (!item.is_read && item.id) {
                    markOneRead.mutate({ id: item.id })
                  }
                }}
              >
                <div className="bani-feed-item__main">
                  <div className="bani-feed-item__badge" />
                  <div className="bani-feed-item__copy">
                    <Text strong={!item.is_read}>
                      {item.title ?? NOTIFICATION_TYPE_LABELS[item.type ?? ''] ?? 'Уведомление'}
                    </Text>
                    <Text type="secondary">{item.body}</Text>
                    <span className="bani-feed-item__meta">
                      {item.created_at ? dayjs(item.created_at).fromNow() : ''}
                    </span>
                  </div>
                </div>
                {!item.is_read && (
                  <div className="bani-feed-item__actions">
                    <Button
                      type="link"
                      size="small"
                      onClick={(event) => {
                        event.stopPropagation()
                        if (item.id) markOneRead.mutate({ id: item.id })
                      }}
                    >
                      Прочитать
                    </Button>
                  </div>
                )}
              </div>
            </List.Item>
          )}
        />

        <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 12 }}>
          <Pagination
            current={page}
            pageSize={pageSize}
            total={meta?.total_count ?? 0}
            onChange={(nextPage, nextPageSize) => {
              setPage(nextPage)
              setPageSize(nextPageSize)
            }}
            showSizeChanger
            showTotal={(total) => `Всего: ${total}`}
          />
        </div>
      </Card>
    </div>
  )
}
