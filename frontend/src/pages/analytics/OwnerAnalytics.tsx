import { useState, useMemo } from 'react'
import {
  Alert,
  Segmented,
  Skeleton,
  Table,
  Typography,
} from '@/components/design/system'
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
} from '@/components/design/icons'
import dayjs from 'dayjs'
import {
  useGetMyBathhousesIdAnalytics,
  useGetMyBathhousesIdAnalyticsDaily,
  useGetMyBathhousesIdAnalyticsPerformance,
} from '@/api/generated/analytics/analytics'
import { useBathhouseStore } from '@/stores/bathhouse'
import { formatPrice } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

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
      <div className="rh-stack">
        <PageHeader
          size="compact"
          eyebrow="Владелец"
          title="Аналитика"
          description="Метрики объекта появятся после выбора бани в верхнем меню."
        />
        <Alert
          title="Выберите баню"
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
    return change > 0 ? '#15803d' : '#b42318'
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

  const renderChange = (change?: number) => {
    if (change === undefined) return null

    return (
      <Text style={{ color: changeColor(change), fontSize: 13 }}>
        {changeIcon(change)} {change > 0 ? '+' : ''}
        {change.toFixed(1)}%
      </Text>
    )
  }

  const kpiTiles = [
    {
      label: 'Бронирования',
      value: dashboard?.bookings ?? 0,
      hint: renderChange(dashboard?.bookings_change),
      icon: <ShoppingOutlined />,
    },
    {
      label: 'Выручка',
      value: formatPrice(dashboard?.revenue ?? 0),
      hint: renderChange(dashboard?.revenue_change),
      icon: <DollarOutlined />,
    },
    {
      label: 'Просмотры',
      value: dashboard?.views ?? 0,
      hint: dashboard?.unique_views !== undefined ? `${dashboard.unique_views} уник.` : 'Уникальные просмотры появятся после накопления данных',
      icon: <EyeOutlined />,
    },
    {
      label: 'Конверсия',
      value: `${((dashboard?.conversion_rate ?? 0) * 100).toFixed(1)}%`,
      hint: 'Доля просмотров, которые дошли до бронирования',
      icon: <FunnelPlotOutlined />,
    },
    {
      label: 'Рейтинг',
      value: `${dashboard?.rating?.toFixed(1) ?? '0.0'} / 5`,
      hint: renderChange(dashboard?.rating_change),
      icon: <StarOutlined />,
    },
    {
      label: 'Средний чек',
      value: formatPrice(dashboard?.avg_check ?? 0),
      hint: 'Средняя сумма бронирования',
      icon: <CalculatorOutlined />,
    },
  ]

  return (
    <div className="rh-stack">
      <PageHeader
        size="compact"
        eyebrow="Владелец"
        title="Аналитика"
        description="Показывает спрос, выручку, конверсию и сравнение объекта с городом."
        extra={(
          <Segmented
            options={PERIOD_OPTIONS}
            value={period}
            onChange={(val) => setPeriod(val as Period)}
          />
        )}
      />

      <div className="rh-stat-grid">
        {kpiTiles.map((tile) => (
          <div className="rh-stat-tile" key={tile.label}>
            {dashLoading ? (
              <Skeleton active paragraph={{ rows: 1 }} />
            ) : (
              <>
                <span className="rh-stat-tile__eyebrow">{tile.icon} {tile.label}</span>
                <div className="rh-stat-tile__value">{tile.value}</div>
                <div className="rh-stat-tile__hint">{tile.hint}</div>
              </>
            )}
          </div>
        ))}
      </div>

      <section className="rh-admin-panel" aria-busy={perfLoading}>
        <div className="rh-admin-toolbar" style={{ marginBottom: 18 }}>
          <div className="rh-admin-toolbar__copy">
            <h2 className="rh-admin-toolbar__title">Сравнение с конкурентами</h2>
            <div className="rh-admin-toolbar__hint">Сравните загрузку, конверсию и рейтинг с городским средним.</div>
          </div>
        </div>
        {performance ? (
          <div className="rh-info-grid">
            <div className="rh-info-card">
              <span className="rh-info-card__label"><UserOutlined /> Ваша загрузка</span>
              <div className="rh-info-card__value">{`${((performance.occupancy_rate ?? 0) * 100).toFixed(1)}%`}</div>
              <div className="rh-info-card__hint">Среднее в городе: {((performance.avg_city_occupancy_rate ?? 0) * 100).toFixed(1)}%</div>
            </div>
            <div className="rh-info-card">
              <span className="rh-info-card__label">Ваша конверсия</span>
              <div className="rh-info-card__value">{`${((performance.conversion_rate ?? 0) * 100).toFixed(1)}%`}</div>
              <div className="rh-info-card__hint">Среднее в городе: {((performance.avg_city_conversion_rate ?? 0) * 100).toFixed(1)}%</div>
            </div>
            <div className="rh-info-card">
              <span className="rh-info-card__label">Ваш рейтинг</span>
              <div className="rh-info-card__value">{performance.avg_rating?.toFixed(1) ?? '—'} / 5</div>
              <div className="rh-info-card__hint">Среднее в городе: {performance.avg_city_rating?.toFixed(1) ?? '—'}</div>
            </div>
          </div>
        ) : (
          <Alert title="Нет данных о конкурентах" type="info" showIcon />
        )}
      </section>

      <section className="rh-admin-table-card">
        <div className="rh-admin-toolbar">
          <div className="rh-admin-toolbar__copy">
            <h2 className="rh-admin-toolbar__title">Динамика по дням</h2>
            <div className="rh-admin-toolbar__hint">Ежедневные бронирования, выручка, просмотры и рейтинг.</div>
          </div>
        </div>
        <Table
          dataSource={dailySnapshots}
          columns={dailyColumns}
          rowKey="date"
          loading={dailyLoading}
          pagination={{ pageSize: 10, showSizeChanger: true, pageSizeOptions: ['10', '30', '90'] }}
          locale={{ emptyText: 'Нет данных за выбранный период' }}
          size="small"
        />
      </section>
    </div>
  )
}
