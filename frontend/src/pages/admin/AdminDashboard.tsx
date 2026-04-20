import { useState } from 'react'
import { Card, Col, Row, Segmented, Spin, Statistic, Table } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  CalendarOutlined,
  CustomerServiceOutlined,
  DollarOutlined,
  EyeOutlined,
  ShopOutlined,
  StarOutlined,
  TeamOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { useQuery } from '@tanstack/react-query'
import { useGetAdminAnalytics, useGetAdminAnalyticsTop } from '@/api/generated/admin-analytics/admin-analytics'
import type { GithubComNikitaaldaevBaniInternalServiceTopBathhouseInfo } from '@/api/generated/model'
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
  const { data: supportMetrics } = useSupportMetrics()

  const dashboard = analyticsData?.data
  const topBathhouses = topData?.data?.bathhouses ?? []

  const topColumns: ColumnsType<GithubComNikitaaldaevBaniInternalServiceTopBathhouseInfo> = [
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
    <div className="bani-stack">
      <PageHeader
        eyebrow="Администрирование"
        title="Панель администратора"
        description="Первый экран собран как обзор платформы: объёмы, выручка, рост аудитории, операционные метрики поддержки и топ объектов по ключевому показателю."
        extra={(
          <Segmented
            options={PERIOD_OPTIONS}
            value={period}
            onChange={(value) => setPeriod(value as string)}
          />
        )}
      />

      <section className="bani-hero-panel">
        <div className="bani-hero-panel__eyebrow">Платформа</div>
        <h2 className="bani-hero-panel__title">Главные числа за выбранный период</h2>
        <div className="bani-hero-panel__description">
          Админу не нужен «красивый набор карточек». Нужен быстрый ответ: сколько бронирований, сколько денег, как растёт платформа и где могут появляться операционные риски.
        </div>

        <Spin spinning={analyticsLoading}>
          <div className="bani-stat-grid">
            <div className="bani-stat-tile">
              <span className="bani-stat-tile__eyebrow">Бронирования</span>
              <div className="bani-stat-tile__value">{(dashboard?.total_bookings ?? 0).toLocaleString('en-US')}</div>
              <span className="bani-stat-tile__hint"><CalendarOutlined /> Суммарный объём заказов по платформе</span>
            </div>
            <div className="bani-stat-tile">
              <span className="bani-stat-tile__eyebrow">Выручка</span>
              <div className="bani-stat-tile__value">{((dashboard?.total_revenue ?? 0) / 100).toLocaleString('en-US')}</div>
              <span className="bani-stat-tile__hint"><DollarOutlined /> Валовая сумма по заказам, ₽</span>
            </div>
            <div className="bani-stat-tile">
              <span className="bani-stat-tile__eyebrow">Пользователи</span>
              <div className="bani-stat-tile__value">{(dashboard?.total_users ?? 0).toLocaleString('en-US')}</div>
              <span className="bani-stat-tile__hint"><TeamOutlined /> Все зарегистрированные клиенты и владельцы</span>
            </div>
            <div className="bani-stat-tile">
              <span className="bani-stat-tile__eyebrow">Бани</span>
              <div className="bani-stat-tile__value">{(dashboard?.total_bathhouses ?? 0).toLocaleString('en-US')}</div>
              <span className="bani-stat-tile__hint"><ShopOutlined /> Активные объекты в каталоге</span>
            </div>
            <div className="bani-stat-tile">
              <span className="bani-stat-tile__eyebrow">Просмотры</span>
              <div className="bani-stat-tile__value">{(dashboard?.total_views ?? 0).toLocaleString('en-US')}</div>
              <span className="bani-stat-tile__hint"><EyeOutlined /> Интерес к каталогу и карточкам объектов</span>
            </div>
            <div className="bani-stat-tile">
              <span className="bani-stat-tile__eyebrow">Средний рейтинг</span>
              <div className="bani-stat-tile__value">{(dashboard?.avg_rating ?? 0).toFixed(1)}</div>
              <span className="bani-stat-tile__hint"><StarOutlined /> Средняя оценка сервиса по платформе</span>
            </div>
            <div className="bani-stat-tile">
              <span className="bani-stat-tile__eyebrow">Новые пользователи</span>
              <div className="bani-stat-tile__value">{(dashboard?.new_users ?? 0).toLocaleString('en-US')}</div>
              <span className="bani-stat-tile__hint"><UserOutlined /> Прирост за выбранный период</span>
            </div>
            <div className="bani-stat-tile">
              <span className="bani-stat-tile__eyebrow">DAU</span>
              <div className="bani-stat-tile__value">{(dashboard?.dau ?? 0).toLocaleString('en-US')}</div>
              <span className="bani-stat-tile__hint">WAU: {dashboard?.wau ?? 0} · MAU: {dashboard?.mau ?? 0}</span>
            </div>
          </div>
        </Spin>
      </section>

      {supportMetrics && (
        <Card data-testid="support-metrics-widget">
          <div className="bani-toolbar" style={{ marginBottom: 18 }}>
            <div>
              <h2 className="bani-section-card__title"><CustomerServiceOutlined /> Поддержка</h2>
              <div className="bani-section-card__description">
                Операционные показатели поддержки рядом с бизнес-метриками помогают вовремя заметить просадку качества сервиса.
              </div>
            </div>
          </div>
          <Row gutter={[16, 16]}>
            <Col xs={12} sm={6}>
              <Statistic
                title="FCR"
                value={supportMetrics.fcr_percent}
                precision={1}
                suffix="%"
                valueStyle={{ color: supportMetrics.fcr_percent >= 70 ? '#52c41a' : '#fa8c16' }}
              />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic title="AHT" value={formatDuration(supportMetrics.aht_seconds)} />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic
                title="SLA (24ч)"
                value={supportMetrics.sla_compliance_percent}
                precision={1}
                suffix="%"
                valueStyle={{ color: supportMetrics.sla_compliance_percent >= 90 ? '#52c41a' : '#fa8c16' }}
              />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic
                title="В очереди"
                value={supportMetrics.queue_size}
                valueStyle={{ color: supportMetrics.queue_size > 20 ? '#f5222d' : supportMetrics.queue_size > 10 ? '#fa8c16' : undefined }}
              />
            </Col>
          </Row>
        </Card>
      )}

      <Card>
        <div className="bani-table-shell">
          <div className="bani-toolbar">
            <div>
              <h2 className="bani-section-card__title">Топ бань</h2>
              <div className="bani-section-card__description">
                Сортировка по ключевому показателю позволяет быстро увидеть лидеров и понять, какие объекты тянут рост платформы.
              </div>
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
      </Card>
    </div>
  )
}
