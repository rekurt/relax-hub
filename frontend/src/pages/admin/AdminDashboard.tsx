import { useState } from 'react'
import { Card, Col, Row, Segmented, Spin, Statistic, Table, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  TeamOutlined,
  ShopOutlined,
  CalendarOutlined,
  DollarOutlined,
  EyeOutlined,
  StarOutlined,
  UserOutlined,
  CustomerServiceOutlined,
} from '@ant-design/icons'
import { useQuery } from '@tanstack/react-query'
import { useGetAdminAnalytics, useGetAdminAnalyticsTop } from '@/api/generated/admin-analytics/admin-analytics'
import type { GithubComNikitaaldaevBaniInternalServiceTopBathhouseInfo } from '@/api/generated/model'
import { formatPrice } from '@/lib/format'
import { axiosInstance } from '@/api/axios-instance'

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
  const mins = Math.round((seconds % 3600) / 60)
  return mins > 0 ? `${hours} ч ${mins} мин` : `${hours} ч`
}

const { Title } = Typography

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
      render: (v: number) => v?.toLocaleString('ru-RU') ?? '—',
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
      render: (v: number) => formatPrice(v ?? 0),
      sorter: (a, b) => (a.revenue ?? 0) - (b.revenue ?? 0),
    },
    {
      title: 'Рейтинг',
      dataIndex: 'rating',
      key: 'rating',
      render: (v: number) => v?.toFixed(1) ?? '—',
      sorter: (a, b) => (a.rating ?? 0) - (b.rating ?? 0),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={3} style={{ margin: 0 }}>Панель администратора</Title>
        <Segmented
          options={PERIOD_OPTIONS}
          value={period}
          onChange={(v) => setPeriod(v as string)}
        />
      </div>

      <Spin spinning={analyticsLoading}>
        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Бронирования"
                value={dashboard?.total_bookings ?? 0}
                prefix={<CalendarOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Выручка"
                value={dashboard?.total_revenue ? dashboard.total_revenue / 100 : 0}
                precision={0}
                suffix="₽"
                prefix={<DollarOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Пользователи"
                value={dashboard?.total_users ?? 0}
                prefix={<TeamOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Бани"
                value={dashboard?.total_bathhouses ?? 0}
                prefix={<ShopOutlined />}
              />
            </Card>
          </Col>
        </Row>

        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Просмотры"
                value={dashboard?.total_views ?? 0}
                prefix={<EyeOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Средний рейтинг"
                value={dashboard?.avg_rating ?? 0}
                precision={1}
                prefix={<StarOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Новые пользователи"
                value={dashboard?.new_users ?? 0}
                prefix={<UserOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={8} lg={6}>
            <Card>
              <Statistic title="DAU" value={dashboard?.dau ?? 0} />
              <div style={{ display: 'flex', gap: 16, marginTop: 8, fontSize: 12, color: '#888' }}>
                <span>WAU: {dashboard?.wau ?? 0}</span>
                <span>MAU: {dashboard?.mau ?? 0}</span>
              </div>
            </Card>
          </Col>
        </Row>
      </Spin>

      {supportMetrics && (
        <Card
          size="small"
          title={<><CustomerServiceOutlined /> Поддержка</>}
          style={{ marginBottom: 24 }}
          data-testid="support-metrics-widget"
        >
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
              <Statistic
                title="AHT"
                value={formatDuration(supportMetrics.aht_seconds)}
              />
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

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>Топ бань</Title>
        <Segmented
          options={METRIC_OPTIONS}
          value={metric}
          onChange={(v) => setMetric(v as string)}
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
  )
}
