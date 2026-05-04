import { useState } from 'react'
import {
  Table,
  Button,
  Badge,
  Tag,
  Space,
  Select,
  App,
  Card,
  Statistic,
} from '@/components/design/system'
import {
  CheckOutlined,
  ExclamationCircleOutlined,
  WarningOutlined,
  InfoCircleOutlined,
  CloseCircleOutlined,
} from '@/components/design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { axiosInstance } from '@/api/axios-instance'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/ru'
import PageHeader from '@/components/PageHeader'

dayjs.extend(relativeTime)
dayjs.locale('ru')

interface AdminNotification {
  id: string
  role: string
  severity: string
  type: string
  title: string
  body: string
  data?: Record<string, unknown>
  is_read: boolean
  read_at?: string
  read_by?: string
  created_at: string
}

interface ListResponse {
  success: boolean
  data: AdminNotification[]
  meta: { page: number; page_size: number; total_count: number; total_pages: number }
}

interface UnreadResponse {
  success: boolean
  data: { unread_count: number }
}

const SEVERITY_CONFIG: Record<string, { color: string; icon: React.ReactNode; label: string }> = {
  critical: { color: 'red', icon: <CloseCircleOutlined />, label: 'Критический' },
  error: { color: 'orange', icon: <ExclamationCircleOutlined />, label: 'Ошибка' },
  warning: { color: 'gold', icon: <WarningOutlined />, label: 'Предупреждение' },
  info: { color: 'blue', icon: <InfoCircleOutlined />, label: 'Информация' },
}

const TYPE_LABELS: Record<string, string> = {
  antifraud_flag: 'Антифрод',
  sla_violation: 'Нарушение SLA',
  reconciliation_mismatch: 'Расхождение сверки',
  float_drift: 'Дрифт баланса',
  ticket_escalation: 'Эскалация тикета',
  dispute_opened: 'Открыт спор',
  kyc_pending: 'Ожидание KYC',
  system_alert: 'Системное',
}

export default function AdminNotificationCenter() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [severityFilter, setSeverityFilter] = useState<string | undefined>()
  const [typeFilter, setTypeFilter] = useState<string | undefined>()
  const [readFilter, setReadFilter] = useState<string | undefined>()
  const { message } = App.useApp()
  const queryClient = useQueryClient()

  const queryKey = ['admin-notifications', page, pageSize, severityFilter, typeFilter, readFilter]

  const { data, isLoading } = useQuery<ListResponse>({
    queryKey,
    queryFn: () => {
      const params: Record<string, string | number> = { page, page_size: pageSize }
      if (severityFilter) params.severity = severityFilter
      if (typeFilter) params.type = typeFilter
      if (readFilter !== undefined) params.is_read = readFilter
      return axiosInstance.get('/admin/notifications', { params }).then((r) => r.data)
    },
  })

  const { data: unreadData } = useQuery<UnreadResponse>({
    queryKey: ['admin-notifications-unread'],
    queryFn: () => axiosInstance.get('/admin/notifications/unread-count').then((r) => r.data),
    refetchInterval: 30_000,
  })

  const markRead = useMutation({
    mutationFn: (id: string) =>
      axiosInstance.put(`/admin/notifications/${id}/read`).then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-notifications'] })
      queryClient.invalidateQueries({ queryKey: ['admin-notifications-unread'] })
    },
  })

  const markAllRead = useMutation({
    mutationFn: () => axiosInstance.put('/admin/notifications/read-all').then((r) => r.data),
    onSuccess: () => {
      message.success('Все уведомления отмечены прочитанными')
      queryClient.invalidateQueries({ queryKey: ['admin-notifications'] })
      queryClient.invalidateQueries({ queryKey: ['admin-notifications-unread'] })
    },
    onError: () => message.error('Ошибка при отметке уведомлений'),
  })

  const notifications = data?.data ?? []
  const meta = data?.meta
  const unreadCount = unreadData?.data?.unread_count ?? 0

  const columns = [
    {
      title: 'Важность',
      dataIndex: 'severity',
      key: 'severity',
      width: 140,
      render: (severity: string) => {
        const cfg = SEVERITY_CONFIG[severity] || { color: 'blue', icon: <InfoCircleOutlined />, label: 'Информация' }
        return (
          <Tag icon={cfg.icon} color={cfg.color}>
            {cfg.label}
          </Tag>
        )
      },
    },
    {
      title: 'Тип',
      dataIndex: 'type',
      key: 'type',
      width: 160,
      render: (type: string) => TYPE_LABELS[type] ?? type,
    },
    {
      title: 'Заголовок',
      dataIndex: 'title',
      key: 'title',
      render: (title: string, record: AdminNotification) => (
        <div className="rh-admin-notification-title">
          <span className={record.is_read ? 'rh-admin-notification-title__text' : 'rh-admin-notification-title__text rh-admin-notification-title__text--unread'}>
            {title}
          </span>
          <span className="rh-admin-notification-title__body">{record.body}</span>
        </div>
      ),
    },
    {
      title: 'Дата',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (date: string) => (
        <span title={dayjs(date).format('DD.MM.YYYY HH:mm')}>{dayjs(date).fromNow()}</span>
      ),
    },
    {
      title: '',
      key: 'actions',
      width: 100,
      render: (_: unknown, record: AdminNotification) =>
        !record.is_read ? (
          <Button
            type="link"
            size="small"
            icon={<CheckOutlined />}
            loading={markRead.isPending}
            onClick={() => markRead.mutate(record.id)}
          >
            Прочитать
          </Button>
        ) : (
          <Tag color="default">Прочитано</Tag>
        ),
    },
  ]

  return (
    <div className="rh-admin-notifications-page">
      <PageHeader
        eyebrow="Мониторинг"
        title="Центр уведомлений"
        description="Операционная лента критичных и информационных событий платформы."
        extra={
          <Button
            icon={<CheckOutlined />}
            onClick={() => markAllRead.mutate()}
            loading={markAllRead.isPending}
            disabled={unreadCount === 0}
          >
            Прочитать все
          </Button>
        }
      />

      <Card size="small" className="rh-admin-notification-stat">
        <Statistic
          title="Непрочитанных"
          value={unreadCount}
          prefix={<Badge status={unreadCount > 0 ? 'processing' : 'default'} />}
        />
      </Card>

      <Card
        className="rh-admin-notification-table"
        extra={
          <Space className="rh-admin-table-tools" wrap>
            <Select
              allowClear
              placeholder="Важность"
              className="rh-admin-notification-filter rh-admin-notification-filter--severity"
              value={severityFilter}
              onChange={(v) => {
                setSeverityFilter(v)
                setPage(1)
              }}
              options={Object.entries(SEVERITY_CONFIG).map(([value, cfg]) => ({
                value,
                label: cfg.label,
              }))}
            />
            <Select
              allowClear
              placeholder="Тип"
              className="rh-admin-notification-filter rh-admin-notification-filter--type"
              value={typeFilter}
              onChange={(v) => {
                setTypeFilter(v)
                setPage(1)
              }}
              options={Object.entries(TYPE_LABELS).map(([value, label]) => ({
                value,
                label,
              }))}
            />
            <Select
              allowClear
              placeholder="Статус"
              className="rh-admin-notification-filter rh-admin-notification-filter--status"
              value={readFilter}
              onChange={(v) => {
                setReadFilter(v)
                setPage(1)
              }}
              options={[
                { value: 'false', label: 'Непрочитанные' },
                { value: 'true', label: 'Прочитанные' },
              ]}
            />
          </Space>
        }
      >
        <Table
          rowKey="id"
          columns={columns}
          dataSource={notifications}
          loading={isLoading}
          locale={{
            emptyText: (
              <div className="rh-admin-empty-state">
                <div className="rh-admin-empty-state__title">Нет уведомлений</div>
                <p className="rh-admin-empty-state__text">
                  События платформы появятся здесь после обработки.
                </p>
              </div>
            ),
          }}
          rowClassName={(record) => (record.is_read ? '' : 'rh-table-row-unread')}
          pagination={{
            current: page,
            pageSize,
            total: meta?.total_count ?? 0,
            onChange: (p, ps) => {
              setPage(p)
              setPageSize(ps)
            },
            showSizeChanger: true,
            showTotal: (total) => `Всего: ${total}`,
          }}
        />
      </Card>
    </div>
  )
}
