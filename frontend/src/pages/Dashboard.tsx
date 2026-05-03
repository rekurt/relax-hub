import { useState, type ReactNode } from 'react'
import { Alert, Card, Segmented, Skeleton, Space, Typography } from 'antd'
import {
  ArrowDownOutlined,
  ArrowUpOutlined,
  CalculatorOutlined,
  DollarOutlined,
  EyeOutlined,
  FunnelPlotOutlined,
  ShoppingOutlined,
  StarOutlined,
} from '@ant-design/icons'
import { useGetMyBathhousesIdAnalytics } from '@/api/generated/analytics/analytics'
import { useBathhouseStore } from '@/stores/bathhouse'
import { formatPrice } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

type Period = '1d' | '7d' | '30d' | '90d'

const PERIOD_OPTIONS = [
  { label: '1 день', value: '1d' as Period },
  { label: '7 дней', value: '7d' as Period },
  { label: '30 дней', value: '30d' as Period },
  { label: '90 дней', value: '90d' as Period },
]

interface KpiCardProps {
  title: string
  value: string | number
  change?: number
  icon: ReactNode
  loading?: boolean
  suffix?: string
  hint: string
}

function KpiCard({ title, value, change, icon, loading, suffix, hint }: KpiCardProps) {
  if (loading) {
    return (
      <Card>
        <Skeleton active paragraph={{ rows: 1 }} />
      </Card>
    )
  }

  const changeColor =
    change === undefined || change === 0
      ? undefined
      : change > 0
        ? '#52c41a'
        : '#ff4d4f'

  const changeIcon =
    change !== undefined && change !== 0
      ? (change > 0 ? <ArrowUpOutlined /> : <ArrowDownOutlined />)
      : null

  return (
    <div className="bani-stat-tile">
      <span className="bani-stat-tile__eyebrow">
        <Space size={8}>
          {icon}
          <span>{title}</span>
        </Space>
      </span>
      <div className="bani-stat-tile__value">
        {value}{suffix ?? ''}
      </div>
      {change !== undefined ? (
        <Text style={{ color: changeColor, fontSize: 13 }}>
          {changeIcon} {change > 0 ? '+' : ''}
          {change.toFixed(1)}% к пред. периоду
        </Text>
      ) : (
        <span className="bani-stat-tile__hint">{hint}</span>
      )}
    </div>
  )
}

export default function Dashboard() {
  const [period, setPeriod] = useState<Period>('30d')
  const selectedBathhouseId = useBathhouseStore((state) => state.selectedBathhouseId)

  const { data, isLoading, isError, error } = useGetMyBathhousesIdAnalytics(
    selectedBathhouseId ?? '',
    { period },
    { query: { enabled: !!selectedBathhouseId } },
  )

  const dashboard = data?.data

  if (!selectedBathhouseId) {
    return (
      <div className="bani-stack">
        <PageHeader
          eyebrow="Аналитика"
          title="Дашборд"
          description="Аналитика жёстко привязана к конкретной бане, поэтому сначала нужно выбрать объект в верхнем меню."
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

  if (isError) {
    const errorMessage =
      error && typeof error === 'object' && 'error' in error
        ? (error as { error?: { message?: string } }).error?.message
        : 'Не удалось загрузить аналитику'

    return (
      <div className="bani-stack">
        <PageHeader
          eyebrow="Аналитика"
          title="Дашборд"
          description="Если аналитика недоступна, пользователь всё равно должен видеть понятное состояние ошибки, а не пустой экран."
        />
        <Alert
          title="Ошибка загрузки"
          description={errorMessage}
          type="error"
          showIcon
        />
      </div>
    )
  }

  return (
    <div className="bani-stack">
      <PageHeader
        eyebrow="Аналитика"
        title="Дашборд"
        description="Это рабочий экран владельца: здесь быстро читаются объём бронирований, деньги, просмотры, конверсия и динамика относительно прошлого периода."
        extra={(
          <Segmented
            options={PERIOD_OPTIONS}
            value={period}
            onChange={(value) => setPeriod(value as Period)}
          />
        )}
      />

      <section className="bani-hero-panel">
        <div className="bani-hero-panel__eyebrow">Состояние объекта</div>
        <h2 className="bani-hero-panel__title">Главные метрики вынесены в первый экран</h2>
        <div className="bani-hero-panel__description">
          Владелец должен видеть не набор разрозненных карточек, а цельную картину: что происходит с трафиком, бронями, чеком и качеством сервиса за выбранный период.
        </div>
        <div className="bani-stat-grid">
          <KpiCard
            title="Бронирования"
            value={(dashboard?.bookings ?? 0).toLocaleString('en-US')}
            change={dashboard?.bookings_change}
            icon={<ShoppingOutlined />}
            loading={isLoading}
            hint="Количество подтверждённых бронирований"
          />
          <KpiCard
            title="Выручка"
            value={formatPrice(dashboard?.revenue ?? 0)}
            change={dashboard?.revenue_change}
            icon={<DollarOutlined />}
            loading={isLoading}
            hint="Оборот за выбранный период"
          />
          <KpiCard
            title="Просмотры"
            value={(dashboard?.views ?? 0).toLocaleString('en-US')}
            change={dashboard?.views_change}
            icon={<EyeOutlined />}
            loading={isLoading}
            hint="Интерес к карточке объекта"
          />
          <KpiCard
            title="Рейтинг"
            value={dashboard?.rating?.toFixed(1) ?? '0.0'}
            change={dashboard?.rating_change}
            icon={<StarOutlined />}
            loading={isLoading}
            suffix="/ 5"
            hint="Качество сервиса глазами клиента"
          />
          <KpiCard
            title="Средний чек"
            value={formatPrice(dashboard?.avg_check ?? 0)}
            icon={<CalculatorOutlined />}
            loading={isLoading}
            hint="Помогает быстро оценить ценовой профиль бронирований"
          />
          <KpiCard
            title="Конверсия"
            value={`${((dashboard?.conversion_rate ?? 0) * 100).toFixed(1)}%`}
            icon={<FunnelPlotOutlined />}
            loading={isLoading}
            hint="Доля просмотров, которая дошла до брони"
          />
        </div>
      </section>
    </div>
  )
}
