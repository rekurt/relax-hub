import { useState } from 'react'
import { Button, Segmented, Spin, Table, Tag, Tooltip } from '@/components/design/system'
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
  InfoCircleOutlined,
  ShopOutlined,
  StarOutlined,
  TeamOutlined,
  UserOutlined,
} from '@/components/design/icons'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
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

function MetricHelp({ title }: { title: string }) {
  return (
    <Tooltip title={title}>
      <span className="rh-admin-help" aria-label={title} tabIndex={0}>
        <InfoCircleOutlined />
      </span>
    </Tooltip>
  )
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
  const navigate = useNavigate()
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
      tooltip: 'Все бронирования за выбранный период: оплаченные, ожидающие и завершённые.',
      icon: <CalendarOutlined />,
    },
    {
      label: 'Выручка',
      value: ((dashboard?.total_revenue ?? 0) / 100).toLocaleString('en-US'),
      hint: 'Валовая сумма по заказам, ₽',
      tooltip: 'GMV в рублях до вычета комиссий, возвратов и операционных расходов.',
      icon: <DollarOutlined />,
    },
    {
      label: 'Пользователи',
      value: (dashboard?.total_users ?? 0).toLocaleString('en-US'),
      hint: 'Все зарегистрированные клиенты и владельцы',
      tooltip: 'Общий размер пользовательской базы: клиенты, владельцы и операторы.',
      icon: <TeamOutlined />,
    },
    {
      label: 'Бани',
      value: (dashboard?.total_bathhouses ?? 0).toLocaleString('en-US'),
      hint: 'Активные объекты в каталоге',
      tooltip: 'Количество объектов, которые участвуют в каталоге и могут влиять на предложение.',
      icon: <ShopOutlined />,
    },
    {
      label: 'Просмотры',
      value: (dashboard?.total_views ?? 0).toLocaleString('en-US'),
      hint: 'Интерес к каталогу и карточкам объектов',
      tooltip: 'Сигнал спроса: просмотры каталога и страниц объектов за выбранный период.',
      icon: <EyeOutlined />,
    },
    {
      label: 'Средний рейтинг',
      value: (dashboard?.avg_rating ?? 0).toFixed(1),
      hint: 'Средняя оценка сервиса по платформе',
      tooltip: 'Среднее значение отзывов; падение ниже 4.4 стоит сверять с модерацией и поддержкой.',
      icon: <StarOutlined />,
    },
    {
      label: 'Новые пользователи',
      value: (dashboard?.new_users ?? 0).toLocaleString('en-US'),
      hint: 'Прирост за выбранный период',
      tooltip: 'Регистрации за активный период фильтра. Смотрите вместе с DAU/WAU/MAU.',
      icon: <UserOutlined />,
    },
    {
      label: 'DAU',
      value: (dashboard?.dau ?? 0).toLocaleString('en-US'),
      hint: `WAU: ${dashboard?.wau ?? 0} · MAU: ${dashboard?.mau ?? 0}`,
      tooltip: 'DAU показывает дневную активность; WAU/MAU помогают увидеть удержание и сезонность.',
      icon: <FieldTimeOutlined />,
    },
  ]

  const supportMetricTiles = supportMetrics
    ? [
        {
          label: 'FCR',
          value: `${supportMetrics.fcr_percent.toFixed(1)}%`,
          hint: 'Решено при первом обращении',
          tooltip: 'First Contact Resolution: доля обращений, закрытых без повторного контакта.',
          icon: <CustomerServiceOutlined />,
        },
        {
          label: 'AHT',
          value: formatDuration(supportMetrics.aht_seconds),
          hint: 'Среднее время обработки',
          tooltip: 'Average Handle Time: сколько в среднем уходит на обработку одного обращения.',
          icon: <FieldTimeOutlined />,
        },
        {
          label: 'SLA (24ч)',
          value: `${supportMetrics.sla_compliance_percent.toFixed(1)}%`,
          hint: 'Соблюдение регламента ответа',
          tooltip: 'Доля обращений, где поддержка уложилась в 24 часа.',
          icon: <CheckCircleOutlined />,
        },
        {
          label: 'В очереди',
          value: supportMetrics.queue_size.toLocaleString('en-US'),
          hint: 'Открытые обращения поддержки',
          tooltip: 'Сколько обращений сейчас открыто и требует внимания операторов.',
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
      to: '/admin/bathhouses',
    },
    {
      type: 'Обращения',
      title: `${moderationQueueSize} обращений ждут ответа`,
      time: supportMetrics ? formatDuration(supportMetrics.aht_seconds) : '—',
      action: 'Разобрать',
      to: '/admin/tickets',
    },
    {
      type: 'Финансы',
      title: `${((dashboard?.total_revenue ?? 0) / 100).toLocaleString('en-US')} ₽ оборота за период`,
      time: period,
      action: 'Сверить',
      to: '/admin/finance',
    },
    {
      type: 'Рост',
      title: `${dashboard?.new_users ?? 0} новых пользователей за выбранный период`,
      time: 'динамика',
      action: 'Смотреть',
      to: '/admin/analytics/funnels',
    },
  ]

  const topColumns: ColumnsType<GithubComRekurtRelaxHubInternalServiceTopBathhouseInfo> = [
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
      width: 260,
    },
    {
      title: 'Просмотры',
      dataIndex: 'views',
      key: 'views',
      width: 130,
      render: (value: number) => value?.toLocaleString('ru-RU') ?? '—',
      sorter: (a, b) => (a.views ?? 0) - (b.views ?? 0),
    },
    {
      title: 'Бронирования',
      dataIndex: 'bookings',
      key: 'bookings',
      width: 150,
      sorter: (a, b) => (a.bookings ?? 0) - (b.bookings ?? 0),
    },
    {
      title: 'Выручка',
      dataIndex: 'revenue',
      key: 'revenue',
      width: 150,
      render: (value: number) => formatPrice(value ?? 0),
      sorter: (a, b) => (a.revenue ?? 0) - (b.revenue ?? 0),
    },
    {
      title: 'Рейтинг',
      dataIndex: 'rating',
      key: 'rating',
      width: 120,
      render: (value: number) => value?.toFixed(1) ?? '—',
      sorter: (a, b) => (a.rating ?? 0) - (b.rating ?? 0),
    },
  ]

  return (
    <div className="rh-stack rh-admin-dashboard">
      <PageHeader
        size="compact"
        eyebrow="Администрирование"
        title="Панель администратора"
        description="Операционная сводка без маркетингового шума: очереди, риски, поддержка, финансы и топ объектов."
        extra={(
          <Tooltip title="Период влияет на KPI, риски и операционную сводку ниже.">
            <span className="rh-admin-dashboard__segmented">
              <Segmented
                options={PERIOD_OPTIONS}
                value={period}
                onChange={(value) => setPeriod(value as string)}
              />
            </span>
          </Tooltip>
        )}
      />

      <Spin spinning={analyticsLoading}>
        <div className="rh-admin-metric-grid">
          {metricTiles.map((tile) => (
            <div className="rh-admin-metric" key={tile.label}>
              <div className="rh-admin-metric__head">
                <span className="rh-admin-metric__label">{tile.label}<MetricHelp title={tile.tooltip} /></span>
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
              <h2 className="rh-admin-toolbar__title">Очередь модерации <MetricHelp title="Сводка задач, которые оператор должен проверить в первую очередь." /></h2>
              <div className="rh-admin-toolbar__hint">Сначала объекты, обращения и финансовые события с влиянием на сервис.</div>
            </div>
            <Tooltip title="Переход к общей очереди всех операционных задач.">
              <Button size="small" onClick={() => navigate('/admin/bathhouses')}>Открыть всё</Button>
            </Tooltip>
          </div>

          <div className="rh-admin-queue">
            {queueRows.map((row) => (
              <div className="rh-admin-queue__row" key={`${row.type}-${row.title}`} title={`${row.type}: ${row.title}`}>
                <Tooltip title={`Категория задачи: ${row.type}`}>
                  <Tag>{row.type}</Tag>
                </Tooltip>
                <div className="rh-admin-queue__title">{row.title}</div>
                <Tooltip title="Время, период или контекст для этой строки.">
                  <div className="rh-admin-queue__meta">{row.time}</div>
                </Tooltip>
                <Tooltip title={`Открыть связанный раздел: ${row.title}`}>
                  <Button size="small" onClick={() => navigate(row.to)}>{row.action}</Button>
                </Tooltip>
              </div>
            ))}
          </div>
        </section>

        <section className="rh-admin-panel">
          <div className="rh-admin-toolbar">
            <div className="rh-admin-toolbar__copy">
              <h2 className="rh-admin-toolbar__title">Риски <MetricHelp title="Автоматические сигналы, которые требуют внимания до того, как метрика станет инцидентом." /></h2>
              <div className="rh-admin-toolbar__hint">Сигналы, которые нужно видеть рядом с цифрами платформы.</div>
            </div>
          </div>

          <div className="rh-admin-risk-list">
            <div className="rh-admin-risk-note rh-admin-risk-note--warning" title="SLA ниже 90% означает, что поддержка не укладывается в целевой регламент.">
              <ExclamationCircleOutlined className="rh-admin-risk-note__icon" />
              <span className="rh-admin-risk-note__text">SLA поддержки: {slaCompliance.toFixed(1)} %. Ниже 90 % требует ручной проверки нагрузки.</span>
              <MetricHelp title="Порог 90% используется как операционный минимум для качества поддержки." />
            </div>
            <div className="rh-admin-risk-note" title="DAU нужно читать вместе с WAU и MAU, чтобы не перепутать разовый всплеск с устойчивым ростом.">
              <CheckCircleOutlined className="rh-admin-risk-note__icon" />
              <span className="rh-admin-risk-note__text">DAU: {dashboard?.dau ?? 0}. Сравните активность с WAU и MAU перед промо-решениями.</span>
              <MetricHelp title="Если DAU растёт без WAU/MAU, это может быть короткая акция, а не удержание." />
            </div>
            <div className="rh-admin-risk-note rh-admin-risk-note--error" title="Большая очередь поддержки быстро ухудшает SLA и отзывы.">
              <AlertOutlined className="rh-admin-risk-note__icon" />
              <span className="rh-admin-risk-note__text">Очередь поддержки: {moderationQueueSize}. При росте выше 20 нужно усилить первую линию.</span>
              <MetricHelp title="Порог 20 обращений помогает вовремя подключить дополнительных операторов." />
            </div>
          </div>
        </section>
      </div>

      {supportMetrics && (
        <section className="rh-admin-panel" data-testid="support-metrics-widget">
          <div className="rh-admin-toolbar" style={{ marginBottom: 18 }}>
            <div className="rh-admin-toolbar__copy">
              <h2 className="rh-admin-toolbar__title"><CustomerServiceOutlined /> Поддержка <MetricHelp title="Показывает качество и скорость обработки обращений поддержки." /></h2>
              <div className="rh-admin-toolbar__hint">Операционные показатели поддержки рядом с бизнес-метриками помогают вовремя заметить просадку качества сервиса.</div>
            </div>
          </div>
          <div className="rh-admin-metric-grid">
            {supportMetricTiles.map((tile) => (
              <div className="rh-admin-metric" key={tile.label}>
                <div className="rh-admin-metric__head">
                  <span className="rh-admin-metric__label">{tile.label}<MetricHelp title={tile.tooltip} /></span>
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
              <h2 className="rh-admin-toolbar__title">Топ бань <MetricHelp title="Рейтинг объектов по выбранной метрике помогает быстро найти лидеров и аномалии." /></h2>
              <div className="rh-admin-toolbar__hint">Сортировка по ключевому показателю позволяет быстро увидеть лидеров роста.</div>
            </div>
            <Tooltip title="Выберите показатель, по которому API вернёт топ объектов.">
              <span className="rh-admin-dashboard__segmented">
                <Segmented
                  options={METRIC_OPTIONS}
                  value={metric}
                  onChange={(value) => setMetric(value as string)}
                />
              </span>
            </Tooltip>
          </div>

          <Table
            columns={topColumns}
            dataSource={topBathhouses}
            rowKey="bathhouse_id"
            loading={topLoading}
            pagination={false}
            scroll={{ x: 810 }}
            locale={{ emptyText: 'Нет данных' }}
          />
        </div>
      </section>
    </div>
  )
}
