import { useState } from 'react'
import {
  Button,
  Typography,
  Card,
  Row,
  Col,
  Progress,
  Table,
  Tag,
  Spin,
  Pagination,
} from '@/components/design/system'
import {
  TrophyOutlined,
  StarOutlined,
  GiftOutlined,
  RiseOutlined,
} from '@/components/design/icons'
import { useGetMyLoyalty } from '@/api/generated/loyalty/loyalty'
import { useGetMyLoyaltyLevels } from '@/api/generated/loyalty/loyalty'
import { useGetMyLoyaltyTransactions } from '@/api/generated/loyalty/loyalty'
import type { InternalHandlerLoyaltyLevelResponse, InternalHandlerLoyaltyTransactionResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'
import { useNavigate } from 'react-router-dom'
import type { ColumnsType } from '@/components/design/types'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

const PAGE_SIZE = 10

const LEVEL_CONFIG: Record<string, { color: string; icon: React.ReactNode; label: string }> = {
  bronze: { color: '#cd7f32', icon: <TrophyOutlined />, label: 'Бронза' },
  silver: { color: '#c0c0c0', icon: <StarOutlined />, label: 'Серебро' },
  gold: { color: '#ffd700', icon: <GiftOutlined />, label: 'Золото' },
  platinum: { color: '#e5e4e2', icon: <RiseOutlined />, label: 'Платина' },
}

const DEFAULT_LEVEL_CFG = LEVEL_CONFIG['bronze'] as { color: string; icon: React.ReactNode; label: string }

const LEVEL_ORDER = ['bronze', 'silver', 'gold', 'platinum']

function getLevelProgress(currentLevel: string, visitCount: number, levels: InternalHandlerLoyaltyLevelResponse[]): number {
  const currentIdx = LEVEL_ORDER.indexOf(currentLevel)
  if (currentIdx === LEVEL_ORDER.length - 1) return 100

  const nextLevelName = LEVEL_ORDER[currentIdx + 1]
  const nextLevel = levels.find((l) => l.level === nextLevelName)
  const currentLevelDef = levels.find((l) => l.level === currentLevel)

  if (!nextLevel || !currentLevelDef) return 0

  const currentMin = currentLevelDef.min_visits ?? 0
  const nextMin = nextLevel.min_visits ?? 1
  const range = nextMin - currentMin
  if (range <= 0) return 100

  const progress = ((visitCount - currentMin) / range) * 100
  return Math.min(Math.max(progress, 0), 100)
}

export default function LoyaltyDashboard() {
  const navigate = useNavigate()
  const [page, setPage] = useState(1)

  const { data: loyaltyData, isLoading: loadingLoyalty } = useGetMyLoyalty()
  const { data: levelsData, isLoading: loadingLevels } = useGetMyLoyaltyLevels()
  const { data: txData, isLoading: loadingTx } = useGetMyLoyaltyTransactions(
    { page, page_size: PAGE_SIZE },
  )

  const loyalty = loyaltyData?.data
  const levels = levelsData?.data ?? []
  const transactions = txData?.data ?? []
  const txMeta = txData?.meta

  const currentLevel = loyalty?.level ?? 'bronze'
  const levelCfg = LEVEL_CONFIG[currentLevel] ?? DEFAULT_LEVEL_CFG
  const progress = getLevelProgress(currentLevel, loyalty?.visit_count ?? 0, levels)

  const columns: ColumnsType<InternalHandlerLoyaltyTransactionResponse> = [
    {
      title: 'Дата',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (val: string) => val ? formatDateTime(val) : '—',
      width: 160,
    },
    {
      title: 'Тип',
      dataIndex: 'type',
      key: 'type',
      render: (val: string) => {
        if (val === 'earned') return <Tag color="green">Начислено</Tag>
        if (val === 'spent') return <Tag color="red">Списано</Tag>
        return <Tag>{val}</Tag>
      },
      width: 120,
    },
    {
      title: 'Баллы',
      dataIndex: 'amount',
      key: 'amount',
      render: (val: number) => (
        <Text strong className={val > 0 ? 'rh-loyalty-amount rh-loyalty-amount--positive' : 'rh-loyalty-amount rh-loyalty-amount--negative'}>
          {val > 0 ? `+${val}` : val}
        </Text>
      ),
      width: 100,
    },
    {
      title: 'Описание',
      dataIndex: 'description',
      key: 'description',
    },
  ]

  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="Личный кабинет"
        title="Программа лояльности"
        description="Следите за уровнем, прогрессом, привилегиями и движением баллов."
      />

      <Spin spinning={loadingLoyalty || loadingLevels}>
        {loyalty ? (
          <>
            {/* Current level card */}
            <Card className={`rh-admin-detail-card rh-loyalty-current-card rh-loyalty-level-tone--${currentLevel}`}>
              <Row gutter={[24, 16]} align="middle">
                <Col xs={24} sm={8}>
                  <div className="rh-loyalty-current-badge">
                    <div className="rh-loyalty-current-badge__icon">{levelCfg.icon}</div>
                    <div className="rh-loyalty-current-badge__title">
                      {levelCfg.label}
                    </div>
                  </div>
                </Col>
                <Col xs={24} sm={16}>
                  <Row gutter={[16, 12]}>
                    <Col span={12}>
                      <Text type="secondary">Баллы</Text>
                      <div className="rh-loyalty-kpi-value">{loyalty.points ?? 0}</div>
                    </Col>
                    <Col span={12}>
                      <Text type="secondary">Визиты</Text>
                      <div className="rh-loyalty-kpi-value">{loyalty.visit_count ?? 0}</div>
                    </Col>
                    <Col span={12}>
                      <Text type="secondary">Всего заработано</Text>
                      <Text>{loyalty.total_earned ?? 0} баллов</Text>
                    </Col>
                    <Col span={12}>
                      <Text type="secondary">Всего потрачено</Text>
                      <Text>{loyalty.total_spent ?? 0} баллов</Text>
                    </Col>
                  </Row>

                  {loyalty.privileges?.next_level && (
                    <div className="rh-loyalty-progress">
                      <div className="rh-loyalty-progress__meta">
                        <Text type="secondary">
                          До уровня {LEVEL_CONFIG[loyalty.privileges.next_level]?.label ?? loyalty.privileges.next_level}
                        </Text>
                        <Text type="secondary">
                          {loyalty.privileges.visits_to_next != null
                            ? `ещё ${loyalty.privileges.visits_to_next} визитов`
                            : ''}
                        </Text>
                      </div>
                      <Progress
                        percent={Math.round(progress)}
                        strokeColor={LEVEL_CONFIG[loyalty.privileges.next_level]?.color ?? '#0f766e'}
                        size="small"
                      />
                    </div>
                  )}
                </Col>
              </Row>
            </Card>

            {/* Privileges */}
            {loyalty.privileges && (
              <Card title="Ваши привилегии" className="rh-admin-detail-card">
                <Row gutter={[16, 12]}>
                  <Col xs={24} sm={12}>
                    <Text type="secondary">Скидка на бронирования</Text>
                    <div className="rh-loyalty-benefit-value">
                      {loyalty.privileges.discount_percent ?? 0}%
                    </div>
                  </Col>
                  <Col xs={24} sm={12}>
                    <Text type="secondary">Множитель баллов</Text>
                    <div className="rh-loyalty-benefit-value">
                      x{loyalty.privileges.point_multiplier ?? 1}
                    </div>
                  </Col>
                </Row>
              </Card>
            )}

            {/* Level info cards */}
            {levels.length > 0 && (
              <>
                <h2 className="rh-section-card__title">Уровни программы</h2>
                <Row gutter={[16, 16]} className="rh-loyalty-levels-row">
                  {levels.map((level) => {
                    const cfg = LEVEL_CONFIG[level.level ?? ''] ?? DEFAULT_LEVEL_CFG
                    const isCurrentLevel = level.level === currentLevel
                    return (
                      <Col key={level.level} xs={24} sm={12} md={6}>
                        <Card
                          size="small"
                          className={`rh-loyalty-tier-card rh-loyalty-level-tone--${level.level ?? 'bronze'}${isCurrentLevel ? ' rh-loyalty-tier-card--current' : ''}`}
                        >
                          <div className="rh-loyalty-tier-card__header">
                            <div className="rh-loyalty-tier-card__icon">{cfg.icon}</div>
                            <Text strong className="rh-loyalty-tier-card__label">{cfg.label}</Text>
                            {isCurrentLevel && <Tag color="blue" className="rh-loyalty-current-tag">Текущий</Tag>}
                          </div>
                          <div>
                            <Text type="secondary" className="rh-loyalty-tier-card__meta">
                              От {level.min_visits ?? 0} визитов
                            </Text>
                          </div>
                          <div>
                            <Text type="secondary" className="rh-loyalty-tier-card__meta">
                              Скидка: {level.discount_percent ?? 0}%
                            </Text>
                          </div>
                          <div>
                            <Text type="secondary" className="rh-loyalty-tier-card__meta">
                              Множитель: x{level.point_multiplier ?? 1}
                            </Text>
                          </div>
                        </Card>
                      </Col>
                    )
                  })}
                </Row>
              </>
            )}

            {/* Transaction history */}
            <h2 className="rh-section-card__title">История операций</h2>
            <Spin spinning={loadingTx}>
              {transactions.length > 0 ? (
                <>
                  <Table
                    dataSource={transactions}
                    columns={columns}
                    rowKey="id"
                    pagination={false}
                    size="small"
                  />
                  {txMeta && txMeta.total_pages && txMeta.total_pages > 1 && (
                    <div className="rh-pagination-center">
                      <Pagination
                        current={page}
                        total={txMeta.total_count}
                        pageSize={PAGE_SIZE}
                        onChange={setPage}
                        showSizeChanger={false}
                      />
                    </div>
                  )}
                </>
              ) : (
                !loadingTx && (
                  <div className="rh-admin-empty-state">
                    <div className="rh-admin-empty-state__title">Пока нет операций с баллами</div>
                    <p className="rh-admin-empty-state__text">
                      Начисления и списания появятся после бронирований и оплаты бонусами.
                    </p>
                  </div>
                )
              )}
            </Spin>
          </>
        ) : (
          !loadingLoyalty && (
            <div className="rh-admin-empty-state">
              <div className="rh-admin-empty-state__title">
                Начните пользоваться сервисом — бронируйте бани и получайте баллы лояльности за каждый визит.
              </div>
              <p className="rh-admin-empty-state__text">
                После первого визита здесь появятся уровень, привилегии и история баллов.
              </p>
              <Button type="primary" onClick={() => navigate('/client')}>
                Найти баню
              </Button>
            </div>
          )
        )}
      </Spin>
    </div>
  )
}
