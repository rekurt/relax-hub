import { useState } from 'react'
import { Button, Segmented, Spin, Table, Tag } from '@/components/design/system'
import type { ColumnsType } from '@/components/design/types'
import {
  AlertOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  CustomerServiceOutlined,
  DollarOutlined,
  ExclamationCircleOutlined,
  EyeOutlined,
  FieldTimeOutlined,
  ShopOutlined,
  StarOutlined,
  TeamOutlined,
  UserOutlined,
} from '@/components/design/icons'
import { useQuery } from '@tanstack/react-query'
import { useGetAdminAnalytics, useGetAdminAnalyticsTop } from '@/api/generated/admin-analytics/admin-analytics'
import type { GithubComRekurtRelaxHubInternalServiceTopBathhouseInfo } from '@/api/generated/model'
import { axiosInstance } from '@/api/axios-instance'
import { formatPrice } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

interface SupportMetrics {
  fcr_percent: number
  aht_seconds: number
  sla_compliance_percent: number
  queue_size: number
}

function useSupportMetrics() {
  return useQuery({
    queryKey: ['admin', 'support', 'metrics'],
    queryFn: async () => {
      const { data } = await axiosInstance.get<{ data: SupportMetrics }>('/admin/tickets/metrics')
      return data.data
    },
    staleTime: 60_000,
  })
}

function formatDuration(seconds: number): string {
  if (seconds < 3600) return `${Math.round(seconds / 60)} мин`
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.round((seconds % 3600) / 60)
  return minutes > 0 ? `${hours} ч ${minutes} мин` : `${hours} ч`
}

const PERIOD_OPTIONS = [
  { label: 'День', value: '1d' },
  { label: 'Неделя', value: '7d' },
  { label: 'Месяц', value: '30d' },
  { label: '3 месяца', value: '90d' },
]

const METRIC_OPTIONS = [
  { label: 'Просмотры', value: 'views' },
  { label: 'Бронирования', value: 'bookings' },
  { label: 'Выручка', value: 'revenue' },
  { label: 'Рейтинг', value: 'rating' },
]

export default function AdminDashboard() {
  const [period, setPeriod] = useState('7d')
  const [metric, setMetric] = useState('bookings')

  const { data: analyticsData, isLoading: analyticsLoading } = useGetAdminAnalytics({ period })
  const { data: topData, isLoading: topLoading } = useGetAdminAnalyticsTop({ metric, limit: 10 })
  const { data: supportMetricsData } = useSupportMetrics()
  const supportMetrics = supportMetricsData
    ? {
        fcr_percent: supportMetricsData.fcr_percent ?? 0,
        aht_seconds: supportMetricsData.aht_seconds ?? 0,
        sla_compliance_percent: supportMetricsData.sla_compliance_percent ?? 0,
        queue_size: supportMetricsData.queue_size ?? 0,
      }
    : undefined

  const dashboard = analyticsData?.data
  const topBathhouses = topData?.data?.bathhouses ?? []
  const moderationQueueSize = supportMetrics?.queue_size ?? 0
  const slaCompliance = supportMetrics?.sla_compliance_percent ?? 0

  const metricTiles = [
    {
      label: 'Бронирования',
      value: (dashboard?.total_bookings ?? 0).toLocaleString('en-US'),
      hint: 'Суммарный объём заказов по платформе',
      icon: <CalendarOutlined />,
    },
    {
      label: 'Выручка',
      value: ((dashboard?.total_revenue ?? 0) / 100).toLocaleString('en-US'),
      hint: 'Валовая сумма по заказам, ₽',
      icon: <DollarOutlined />,
    },
    {
      label: 'Пользователи',
      value: (dashboard?.total_users ?? 0).toLocaleString('en-US'),
      hint: 'Все зарегистрированные клиенты и владельцы',
      icon: <TeamOutlined />,
    },
    {
      label: 'Бани',
      value: (dashboard?.total_bathhouses ?? 0).toLocaleString('en-US'),
      hint: 'Активные объекты в каталоге',
      icon: <ShopOutlined />,
    },
    {
      label: 'Просмотры',
      value: (dashboard?.total_views ?? 0).toLocaleString('en-US'),
      hint: 'Интерес к каталогу и карточкам объектов',
      icon: <EyeOutlined />,
    },
    {
      label: 'Средний рейтинг',
      value: (dashboard?.avg_rating ?? 0).toFixed(1),
      hint: 'Средняя оценка сервиса по платформе',
      icon: <StarOutlined />,
    },
    {
      label: 'Новые пользователи',
      value: (dashboard?.new_users ?? 0).toLocaleString('en-US'),
      hint: 'Прирост за выбранный период',
      icon: <UserOutlined />,
    },
    {
      label: 'DAU',
      value: (dashboard?.dau ?? 0).toLocaleString('en-US'),
      hint: `WAU: ${dashboard?.wau ?? 0} · MAU: ${dashboard?.mau ?? 0}`,
      icon: <FieldTimeOutlined />,
    },
  ]

  const supportMetricTiles = supportMetrics
    ? [
        {
          label: 'FCR',
          value: `${supportMetrics.fcr_percent.toFixed(1)}%`,
          hint: 'Решено при первом обращении',
          icon: <CustomerServiceOutlined />,
        },
        {
          label: 'AHT',
          value: formatDuration(supportMetrics.aht_seconds),
          hint: 'Среднее время обработки',
          icon: <FieldTimeOutlined />,
        },
        {
          label: 'SLA (24ч)',
          value: `${supportMetrics.sla_compliance_percent.toFixed(1)}%`,
          hint: 'Соблюдение регламента ответа',
          icon: <CheckCircleOutlined />,
        },
        {
          label: 'В очереди',
          value: supportMetrics.queue_size.toLocaleString('en-US'),
          hint: 'Открытые обращения поддержки',
          icon: <AlertOutlined />,
        },
      ]
    : []

  const queueRows = [
    {
      type: 'Модерация',
      title: `${dashboard?.total_bathhouses ?? 0} объектов в контуре каталога`,
      time: 'сейчас',
      action: 'Открыть',
    },
    {
      type: 'Обращения',
      title: `${moderationQueueSize} обращений ждут ответа`,
      time: supportMetrics ? formatDuration(supportMetrics.aht_seconds) : '—',
      action: 'Разобрать',
    },
    {
      type: 'Финансы',
      title: `${((dashboard?.total_revenue ?? 0) / 100).toLocaleString('en-US')} ₽ оборота за период`,
      time: period,
      action: 'Сверить',
    },
    {
      type: 'Рост',
      title: `${dashboard?.new_users ?? 0} новых пользователей за выбранный период`,
      time: 'динамика',
      action: 'Смотреть',
    },
  ]

  const topColumns: ColumnsType<GithubComRekurtRelaxHubInternalServiceTopBathhouseInfo> = [
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: 'Просмотры',
      dataIndex: 'views',
      key: 'views',
      render: (value: number) => value?.toLocaleString('ru-RU') ?? '—',
      sorter: (a, b) => (a.views ?? 0) - (b.views ?? 0),
    },
    {
      title: 'Бронирования',
      dataIndex: 'bookings',
      key: 'bookings',
      sorter: (a, b) => (a.bookings ?? 0) - (b.bookings ?? 0),
    },
    {
      title: 'Выручка',
      dataIndex: 'revenue',
      key: 'revenue',
      render: (value: number) => formatPrice(value ?? 0),
      sorter: (a, b) => (a.revenue ?? 0) - (b.revenue ?? 0),
    },
    {
      title: 'Рейтинг',
      dataIndex: 'rating',
      key: 'rating',
      render: (value: number) => value?.toFixed(1) ?? '—',
      sorter: (a, b) => (a.rating ?? 0) - (b.rating ?? 0),
    },
  ]

  return (
    <div className="rh-stack">
      <PageHeader
        size="compact"
        eyebrow="Администрирование"
        title="Панель администратора"
        description="Операционная сводка без маркетингового шума: очереди, риски, поддержка, финансы и топ объектов."
        extra={(
          <Segmented
            options={PERIOD_OPTIONS}
            value={period}
            onChange={(value) => setPeriod(value as string)}
          />
        )}
      />

      <Spin spinning={analyticsLoading}>
        <div className="rh-admin-metric-grid">
          {metricTiles.map((tile) => (
            <div className="rh-admin-metric" key={tile.label}>
              <div className="rh-admin-metric__head">
                <span className="rh-admin-metric__label">{tile.label}</span>
                <span className="rh-admin-metric__icon">{tile.icon}</span>
              </div>
              <div className="rh-admin-metric__value">{tile.value}</div>
              <div className="rh-admin-metric__hint">{tile.hint}</div>
            </div>
          ))}
        </div>
      </Spin>

      <div className="rh-admin-grid rh-admin-grid--split">
        <section className="rh-admin-panel">
          <div className="rh-admin-toolbar">
            <div className="rh-admin-toolbar__copy">
              <h2 className="rh-admin-toolbar__title">Очередь модерации</h2>
              <div className="rh-admin-toolbar__hint">Сначала объекты, обращения и финансовые события с влиянием на сервис.</div>
            </div>
            <Button size="small">Открыть всё</Button>
          </div>

          <div className="rh-admin-queue">
            {queueRows.map((row) => (
              <div className="rh-admin-queue__row" key={`${row.type}-${row.title}`}>
                <Tag>{row.type}</Tag>
                <div className="rh-admin-queue__title">{row.title}</div>
                <div className="rh-admin-queue__meta">{row.time}</div>
                <Button size="small">{row.action}</Button>
              </div>
            ))}
          </div>
        </section>

        <section className="rh-admin-panel">
          <div className="rh-admin-toolbar">
            <div className="rh-admin-toolbar__copy">
              <h2 className="rh-admin-toolbar__title">Риски</h2>
              <div className="rh-admin-toolbar__hint">Сигналы, которые нужно видеть рядом с цифрами платформы.</div>
            </div>
          </div>

          <div className="rh-admin-risk-list">
            <div className="rh-admin-risk-note rh-admin-risk-note--warning">
              <ExclamationCircleOutlined className="rh-admin-risk-note__icon" />
              <span className="rh-admin-risk-note__text">SLA поддержки: {slaCompliance.toFixed(1)} %. Ниже 90 % требует ручной проверки нагрузки.</span>
            </div>
            <div className="rh-admin-risk-note">
              <CheckCircleOutlined className="rh-admin-risk-note__icon" />
              <span className="rh-admin-risk-note__text">DAU: {dashboard?.dau ?? 0}. Сравните активность с WAU и MAU перед промо-решениями.</span>
            </div>
            <div className="rh-admin-risk-note rh-admin-risk-note--error">
              <AlertOutlined className="rh-admin-risk-note__icon" />
              <span className="rh-admin-risk-note__text">Очередь поддержки: {moderationQueueSize}. При росте выше 20 нужно усилить первую линию.</span>
            </div>
          </div>
        </section>
      </div>

      {supportMetrics && (
        <section className="rh-admin-panel" data-testid="support-metrics-widget">
          <div className="rh-admin-toolbar" style={{ marginBottom: 18 }}>
            <div className="rh-admin-toolbar__copy">
              <h2 className="rh-admin-toolbar__title"><CustomerServiceOutlined /> Поддержка</h2>
              <div className="rh-admin-toolbar__hint">Операционные показатели поддержки рядом с бизнес-метриками помогают вовремя заметить просадку качества сервиса.</div>
            </div>
          </div>
          <div className="rh-admin-metric-grid">
            {supportMetricTiles.map((tile) => (
              <div className="rh-admin-metric" key={tile.label}>
                <div className="rh-admin-metric__head">
                  <span className="rh-admin-metric__label">{tile.label}</span>
                  <span className="rh-admin-metric__icon">{tile.icon}</span>
                </div>
                <div className="rh-admin-metric__value">{tile.value}</div>
                <div className="rh-admin-metric__hint">{tile.hint}</div>
              </div>
            ))}
          </div>
        </section>
      )}

      <section className="rh-admin-table-card">
        <div className="rh-table-shell">
          <div className="rh-admin-toolbar">
            <div className="rh-admin-toolbar__copy">
              <h2 className="rh-admin-toolbar__title">Топ бань</h2>
              <div className="rh-admin-toolbar__hint">Сортировка по ключевому показателю позволяет быстро увидеть лидеров роста.</div>
            </div>
            <Segmented
              options={METRIC_OPTIONS}
              value={metric}
              onChange={(value) => setMetric(value as string)}
            />
          </div>

          <Table
            columns={topColumns}
            dataSource={topBathhouses}
            rowKey="bathhouse_id"
            loading={topLoading}
            pagination={false}
            locale={{ emptyText: 'Нет данных' }}
          />
        </div>
      </section>
    </div>
  )
}
