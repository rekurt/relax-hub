import { useState, useMemo } from 'react'
import {
  Alert,
  Card,
  Col,
  Row,
  Segmented,
  Skeleton,
  Statistic,
  Table,
  Typography,
} from 'antd'
import {
  ArrowUpOutlined,
  ArrowDownOutlined,
  ShoppingOutlined,
  DollarOutlined,
  EyeOutlined,
  StarOutlined,
  CalculatorOutlined,
  FunnelPlotOutlined,
  UserOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import {
  useGetMyBathhousesIdAnalytics,
  useGetMyBathhousesIdAnalyticsDaily,
  useGetMyBathhousesIdAnalyticsPerformance,
} from '@/api/generated/analytics/analytics'
import { useBathhouseStore } from '@/stores/bathhouse'
import { formatPrice } from '@/lib/format'

const { Title, Text } = Typography

type Period = '7d' | '30d' | '90d'

const PERIOD_OPTIONS = [
  { label: '7 дней', value: '7d' as Period },
  { label: '30 дней', value: '30d' as Period },
  { label: '90 дней', value: '90d' as Period },
]

function periodToDates(period: Period) {
  const to = dayjs().format('YYYY-MM-DD')
  const from = dayjs()
    .subtract(period === '7d' ? 7 : period === '30d' ? 30 : 90, 'day')
    .format('YYYY-MM-DD')
  return { from, to }
}

export default function OwnerAnalytics() {
  const [period, setPeriod] = useState<Period>('30d')
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)

  const { data: dashData, isLoading: dashLoading } = useGetMyBathhousesIdAnalytics(
    selectedBathhouseId ?? '',
    { period },
    { query: { enabled: !!selectedBathhouseId } },
  )

  const dateRange = useMemo(() => periodToDates(period), [period])

  const { data: dailyData, isLoading: dailyLoading } = useGetMyBathhousesIdAnalyticsDaily(
    selectedBathhouseId ?? '',
    dateRange,
    { query: { enabled: !!selectedBathhouseId } },
  )

  const { data: perfData, isLoading: perfLoading } = useGetMyBathhousesIdAnalyticsPerformance(
    selectedBathhouseId ?? '',
    { period },
    { query: { enabled: !!selectedBathhouseId } },
  )

  const dashboard = dashData?.data
  const dailySnapshots = dailyData?.data ?? []
  const performance = perfData?.data

  if (!selectedBathhouseId) {
    return (
      <div>
        <Title level={3}>Аналитика</Title>
        <Alert
          message="Выберите баню"
          description="Для просмотра аналитики выберите баню в верхнем меню."
          type="info"
          showIcon
        />
      </div>
    )
  }

  const changeIcon = (change?: number) => {
    if (change === undefined || change === 0) return null
    return change > 0 ? <ArrowUpOutlined /> : <ArrowDownOutlined />
  }

  const changeColor = (change?: number) => {
    if (change === undefined || change === 0) return undefined
    return change > 0 ? '#52c41a' : '#ff4d4f'
  }

  const dailyColumns = [
    {
      title: 'Дата',
      dataIndex: 'date',
      key: 'date',
      render: (val: string) => dayjs(val).format('DD.MM.YYYY'),
    },
    {
      title: 'Бронирования',
      dataIndex: 'bookings',
      key: 'bookings',
    },
    {
      title: 'Выручка',
      dataIndex: 'revenue',
      key: 'revenue',
      render: (val: number) => formatPrice(val ?? 0),
    },
    {
      title: 'Просмотры',
      dataIndex: 'views',
      key: 'views',
    },
    {
      title: 'Уник. просмотры',
      dataIndex: 'unique_views',
      key: 'unique_views',
    },
    {
      title: 'Рейтинг',
      dataIndex: 'avg_rating',
      key: 'avg_rating',
      render: (val: number) => val?.toFixed(1) ?? '—',
    },
  ]

  return (
    <div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 24,
          flexWrap: 'wrap',
          gap: 12,
        }}
      >
        <Title level={3} style={{ margin: 0 }}>
          Аналитика
        </Title>
        <Segmented
          options={PERIOD_OPTIONS}
          value={period}
          onChange={(val) => setPeriod(val as Period)}
        />
      </div>

      {/* KPI Cards */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            {dashLoading ? (
              <Skeleton active paragraph={{ rows: 1 }} />
            ) : (
              <>
                <Statistic
                  title={<><ShoppingOutlined /> Бронирования</>}
                  value={dashboard?.bookings ?? 0}
                />
                {dashboard?.bookings_change !== undefined && (
                  <Text style={{ color: changeColor(dashboard.bookings_change), fontSize: 13 }}>
                    {changeIcon(dashboard.bookings_change)} {dashboard.bookings_change > 0 ? '+' : ''}
                    {dashboard.bookings_change.toFixed(1)}%
                  </Text>
                )}
              </>
            )}
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            {dashLoading ? (
              <Skeleton active paragraph={{ rows: 1 }} />
            ) : (
              <>
                <Statistic
                  title={<><DollarOutlined /> Выручка</>}
                  value={formatPrice(dashboard?.revenue ?? 0)}
                />
                {dashboard?.revenue_change !== undefined && (
                  <Text style={{ color: changeColor(dashboard.revenue_change), fontSize: 13 }}>
                    {changeIcon(dashboard.revenue_change)} {dashboard.revenue_change > 0 ? '+' : ''}
                    {dashboard.revenue_change.toFixed(1)}%
                  </Text>
                )}
              </>
            )}
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            {dashLoading ? (
              <Skeleton active paragraph={{ rows: 1 }} />
            ) : (
              <Statistic
                title={<><EyeOutlined /> Просмотры</>}
                value={dashboard?.views ?? 0}
                suffix={
                  dashboard?.unique_views !== undefined
                    ? <Text type="secondary" style={{ fontSize: 14 }}> / {dashboard.unique_views} уник.</Text>
                    : undefined
                }
              />
            )}
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            {dashLoading ? (
              <Skeleton active paragraph={{ rows: 1 }} />
            ) : (
              <>
                <Statistic
                  title={<><FunnelPlotOutlined /> Конверсия</>}
                  value={`${((dashboard?.conversion_rate ?? 0) * 100).toFixed(1)}%`}
                />
              </>
            )}
          </Card>
        </Col>
      </Row>

      {/* Second row: rating, avg check */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            {dashLoading ? (
              <Skeleton active paragraph={{ rows: 1 }} />
            ) : (
              <>
                <Statistic
                  title={<><StarOutlined /> Рейтинг</>}
                  value={dashboard?.rating?.toFixed(1) ?? '0.0'}
                  suffix="/ 5"
                />
                {dashboard?.rating_change !== undefined && (
                  <Text style={{ color: changeColor(dashboard.rating_change), fontSize: 13 }}>
                    {changeIcon(dashboard.rating_change)} {dashboard.rating_change > 0 ? '+' : ''}
                    {dashboard.rating_change.toFixed(1)}%
                  </Text>
                )}
              </>
            )}
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            {dashLoading ? (
              <Skeleton active paragraph={{ rows: 1 }} />
            ) : (
              <Statistic
                title={<><CalculatorOutlined /> Средний чек</>}
                value={formatPrice(dashboard?.avg_check ?? 0)}
              />
            )}
          </Card>
        </Col>
      </Row>

      {/* Competitor Comparison */}
      <Card title="Сравнение с конкурентами" style={{ marginTop: 24 }} loading={perfLoading}>
        {performance ? (
          <Row gutter={[24, 16]}>
            <Col xs={24} sm={8}>
              <Statistic
                title="Ваша загрузка"
                value={`${((performance.occupancy_rate ?? 0) * 100).toFixed(1)}%`}
                prefix={<UserOutlined />}
              />
              <Text type="secondary">
                Среднее в городе: {((performance.avg_city_occupancy_rate ?? 0) * 100).toFixed(1)}%
              </Text>
            </Col>
            <Col xs={24} sm={8}>
              <Statistic
                title="Ваша конверсия"
                value={`${((performance.conversion_rate ?? 0) * 100).toFixed(1)}%`}
              />
              <Text type="secondary">
                Среднее в городе: {((performance.avg_city_conversion_rate ?? 0) * 100).toFixed(1)}%
              </Text>
            </Col>
            <Col xs={24} sm={8}>
              <Statistic
                title="Ваш рейтинг"
                value={performance.avg_rating?.toFixed(1) ?? '—'}
                suffix="/ 5"
              />
              <Text type="secondary">
                Среднее в городе: {performance.avg_city_rating?.toFixed(1) ?? '—'}
              </Text>
            </Col>
          </Row>
        ) : (
          <Alert message="Нет данных о конкурентах" type="info" showIcon />
        )}
      </Card>

      {/* Daily Breakdown Table */}
      <Card title="Динамика по дням" style={{ marginTop: 24 }}>
        <Table
          dataSource={dailySnapshots}
          columns={dailyColumns}
          rowKey="date"
          loading={dailyLoading}
          pagination={{ pageSize: 10, showSizeChanger: true, pageSizeOptions: ['10', '30', '90'] }}
          locale={{ emptyText: 'Нет данных за выбранный период' }}
          size="small"
        />
      </Card>
    </div>
  )
}
