import { useState } from 'react'
import {
  Card,
  Col,
  Row,
  Skeleton,
  Segmented,
  Statistic,
  Alert,
  Typography,
  Space,
} from 'antd'
import {
  ShoppingOutlined,
  DollarOutlined,
  EyeOutlined,
  StarOutlined,
  CalculatorOutlined,
  FunnelPlotOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
} from '@ant-design/icons'
import { useGetMyBathhousesIdAnalytics } from '@/api/generated/analytics/analytics'
import { useBathhouseStore } from '@/stores/bathhouse'
import { formatPrice } from '@/lib/format'

const { Title, Text } = Typography

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
  icon: React.ReactNode
  loading?: boolean
  suffix?: string
}

function KpiCard({ title, value, change, icon, loading, suffix }: KpiCardProps) {
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
    change !== undefined && change !== 0 ? (
      change > 0 ? (
        <ArrowUpOutlined />
      ) : (
        <ArrowDownOutlined />
      )
    ) : null

  return (
    <Card>
      <Statistic
        title={
          <Space>
            {icon}
            <span>{title}</span>
          </Space>
        }
        value={value}
        suffix={suffix}
        styles={{ content: { fontSize: 28 } }}
      />
      {change !== undefined && (
        <Text
          style={{ color: changeColor, fontSize: 13, marginTop: 4, display: 'block' }}
        >
          {changeIcon} {change > 0 ? '+' : ''}
          {change.toFixed(1)}% к пред. периоду
        </Text>
      )}
    </Card>
  )
}

export default function Dashboard() {
  const [period, setPeriod] = useState<Period>('30d')
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)

  const { data, isLoading, isError, error } = useGetMyBathhousesIdAnalytics(
    selectedBathhouseId ?? '',
    { period },
    { query: { enabled: !!selectedBathhouseId } },
  )

  const dashboard = data?.data

  if (!selectedBathhouseId) {
    return (
      <div>
        <Title level={3}>Дашборд</Title>
        <Alert
          message="Выберите баню"
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
      <div>
        <Title level={3}>Дашборд</Title>
        <Alert
          message="Ошибка загрузки"
          description={errorMessage}
          type="error"
          showIcon
        />
      </div>
    )
  }

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
          Дашборд
        </Title>
        <Segmented
          options={PERIOD_OPTIONS}
          value={period}
          onChange={(val) => setPeriod(val as Period)}
        />
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={8}>
          <KpiCard
            title="Бронирования"
            value={dashboard?.bookings ?? 0}
            change={dashboard?.bookings_change}
            icon={<ShoppingOutlined />}
            loading={isLoading}
          />
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <KpiCard
            title="Выручка"
            value={formatPrice(dashboard?.revenue ?? 0)}
            change={dashboard?.revenue_change}
            icon={<DollarOutlined />}
            loading={isLoading}
          />
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <KpiCard
            title="Просмотры"
            value={dashboard?.views ?? 0}
            change={dashboard?.views_change}
            icon={<EyeOutlined />}
            loading={isLoading}
          />
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <KpiCard
            title="Рейтинг"
            value={dashboard?.rating?.toFixed(1) ?? '0.0'}
            change={dashboard?.rating_change}
            icon={<StarOutlined />}
            loading={isLoading}
            suffix="/ 5"
          />
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <KpiCard
            title="Средний чек"
            value={formatPrice(dashboard?.avg_check ?? 0)}
            icon={<CalculatorOutlined />}
            loading={isLoading}
          />
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <KpiCard
            title="Конверсия"
            value={`${((dashboard?.conversion_rate ?? 0) * 100).toFixed(1)}%`}
            icon={<FunnelPlotOutlined />}
            loading={isLoading}
          />
        </Col>
      </Row>
    </div>
  )
}
