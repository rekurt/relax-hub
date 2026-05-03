import { useState } from 'react'
import { Card, Col, Row, Segmented, Spin, Statistic, Table, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  ShopOutlined,
  SearchOutlined,
  CalendarOutlined,
  WalletOutlined,
  UserOutlined,
  DollarOutlined,
} from '@ant-design/icons'
import {
  useGetAdminAnalyticsGeo,
  useGetAdminAnalyticsWallet,
  useGetAdminAnalyticsBusinessMetrics,
  useGetAdminAnalyticsPnl,
} from '@/api/generated/admin-analytics/admin-analytics'
import type { GithubComRekurtRelaxHubInternalDomainGeoSupplyDemand } from '@/api/generated/model'

const { Title } = Typography

const PERIOD_OPTIONS = [
  { label: 'День', value: '1d' },
  { label: 'Неделя', value: '7d' },
  { label: 'Месяц', value: '30d' },
  { label: '3 месяца', value: '90d' },
]

export default function SupplyDemandMetrics() {
  const [period, setPeriod] = useState('30d')

  const { data: geoData, isLoading: geoLoading } = useGetAdminAnalyticsGeo({ period })
  const { data: walletData, isLoading: walletLoading } = useGetAdminAnalyticsWallet({ period })
  const { data: bizData, isLoading: bizLoading } = useGetAdminAnalyticsBusinessMetrics({ period })
  const { data: pnlData, isLoading: pnlLoading } = useGetAdminAnalyticsPnl({ period })

  const cities = geoData?.data?.cities ?? []
  const wallet = walletData?.data
  const biz = bizData?.data
  const pnl = pnlData?.data

  const isLoading = geoLoading || walletLoading || bizLoading || pnlLoading

  const totalListings = cities.reduce((sum, c) => sum + (c.listing_count ?? 0), 0)
  const totalSearches = cities.reduce((sum, c) => sum + (c.search_count ?? 0), 0)
  const totalBookings = cities.reduce((sum, c) => sum + (c.booking_count ?? 0), 0)

  const cityColumns: ColumnsType<GithubComRekurtRelaxHubInternalDomainGeoSupplyDemand> = [
    {
      title: 'Город',
      dataIndex: 'city_name',
      key: 'city_name',
      ellipsis: true,
    },
    {
      title: 'Листинги',
      dataIndex: 'listing_count',
      key: 'listing_count',
      sorter: (a, b) => (a.listing_count ?? 0) - (b.listing_count ?? 0),
    },
    {
      title: 'Поиски',
      dataIndex: 'search_count',
      key: 'search_count',
      sorter: (a, b) => (a.search_count ?? 0) - (b.search_count ?? 0),
    },
    {
      title: 'Бронирования',
      dataIndex: 'booking_count',
      key: 'booking_count',
      sorter: (a, b) => (a.booking_count ?? 0) - (b.booking_count ?? 0),
    },
    {
      title: 'Конверсия',
      key: 'conversion',
      render: (_, record) => {
        const searches = record.search_count ?? 0
        const bookings = record.booking_count ?? 0
        if (searches === 0) return '—'
        return `${((bookings / searches) * 100).toFixed(1)}%`
      },
      sorter: (a, b) => {
        const rateA = (a.search_count ?? 0) > 0 ? (a.booking_count ?? 0) / (a.search_count ?? 1) : 0
        const rateB = (b.search_count ?? 0) > 0 ? (b.booking_count ?? 0) / (b.search_count ?? 1) : 0
        return rateA - rateB
      },
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={3} style={{ margin: 0 }}>Метрики предложения и спроса</Title>
        <Segmented
          options={PERIOD_OPTIONS}
          value={period}
          onChange={(v) => setPeriod(v as string)}
        />
      </div>

      <Spin spinning={isLoading}>
        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Активные листинги"
                value={totalListings}
                prefix={<ShopOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Поисковые запросы"
                value={totalSearches}
                prefix={<SearchOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Бронирования"
                value={totalBookings}
                prefix={<CalendarOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Конверсия поиск→бронь"
                value={totalSearches > 0 ? (totalBookings / totalSearches) * 100 : 0}
                precision={1}
                suffix="%"
              />
            </Card>
          </Col>
        </Row>

        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="DAU"
                value={biz?.dau ?? 0}
                prefix={<UserOutlined />}
              />
              <div style={{ display: 'flex', gap: 16, marginTop: 8, fontSize: 12, color: 'var(--rh-text-soft)' }}>
                <span>MAU: {biz?.mau ?? 0}</span>
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="ADR"
                value={(biz?.adr ?? 0) / 100}
                precision={0}
                suffix="₽"
                prefix={<DollarOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="ARPU"
                value={(biz?.arpu ?? 0) / 100}
                precision={0}
                suffix="₽"
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Отток"
                value={biz?.churn_rate ?? 0}
                precision={1}
                suffix="%"
              />
            </Card>
          </Col>
        </Row>

        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="GMV"
                value={(pnl?.gmv ?? 0) / 100}
                precision={0}
                suffix="₽"
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Take Rate"
                value={pnl?.take_rate ?? 0}
                precision={2}
                suffix="%"
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Выручка платформы"
                value={(pnl?.platform_revenue ?? 0) / 100}
                precision={0}
                suffix="₽"
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Ср. выручка/бронь"
                value={(pnl?.revenue_per_booking ?? 0) / 100}
                precision={0}
                suffix="₽"
              />
            </Card>
          </Col>
        </Row>

        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Активные кошельки"
                value={wallet?.active_wallets ?? 0}
                prefix={<WalletOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Баланс клиентов"
                value={(wallet?.total_client_balance ?? 0) / 100}
                precision={0}
                suffix="₽"
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Баланс владельцев"
                value={(wallet?.total_owner_balance ?? 0) / 100}
                precision={0}
                suffix="₽"
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Оплата кошельком"
                value={wallet?.wallet_payment_share ?? 0}
                precision={1}
                suffix="%"
              />
            </Card>
          </Col>
        </Row>

        <Card title="Спрос и предложение по городам">
          <Table
            columns={cityColumns}
            dataSource={cities}
            rowKey="city_id"
            pagination={false}
            locale={{ emptyText: 'Нет данных' }}
          />
        </Card>
      </Spin>
    </div>
  )
}
