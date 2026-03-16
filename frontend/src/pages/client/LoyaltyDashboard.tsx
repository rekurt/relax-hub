import { useState } from 'react'
import {
  Typography,
  Card,
  Row,
  Col,
  Progress,
  Table,
  Tag,
  Spin,
  Empty,
  Pagination,
} from 'antd'
import {
  TrophyOutlined,
  StarOutlined,
  GiftOutlined,
  RiseOutlined,
} from '@ant-design/icons'
import { useGetMyLoyalty } from '@/api/generated/loyalty/loyalty'
import { useGetMyLoyaltyLevels } from '@/api/generated/loyalty/loyalty'
import { useGetMyLoyaltyTransactions } from '@/api/generated/loyalty/loyalty'
import type { InternalHandlerLoyaltyLevelResponse, InternalHandlerLoyaltyTransactionResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'
import type { ColumnsType } from 'antd/es/table'

const { Title, Text } = Typography

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
        <Text strong style={{ color: val > 0 ? '#52c41a' : '#ff4d4f' }}>
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
    <div>
      <Title level={3}>Программа лояльности</Title>

      <Spin spinning={loadingLoyalty || loadingLevels}>
        {loyalty ? (
          <>
            {/* Current level card */}
            <Card style={{ marginBottom: 24 }}>
              <Row gutter={[24, 16]} align="middle">
                <Col xs={24} sm={8}>
                  <div style={{ textAlign: 'center' }}>
                    <div style={{ fontSize: 48, color: levelCfg.color }}>{levelCfg.icon}</div>
                    <Title level={4} style={{ margin: '8px 0 0', color: levelCfg.color }}>
                      {levelCfg.label}
                    </Title>
                  </div>
                </Col>
                <Col xs={24} sm={16}>
                  <Row gutter={[16, 12]}>
                    <Col span={12}>
                      <Text type="secondary">Баллы</Text>
                      <Title level={4} style={{ margin: 0 }}>{loyalty.points ?? 0}</Title>
                    </Col>
                    <Col span={12}>
                      <Text type="secondary">Визиты</Text>
                      <Title level={4} style={{ margin: 0 }}>{loyalty.visit_count ?? 0}</Title>
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
                    <div style={{ marginTop: 16 }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
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
                        strokeColor={LEVEL_CONFIG[loyalty.privileges.next_level]?.color ?? '#1890ff'}
                        size="small"
                      />
                    </div>
                  )}
                </Col>
              </Row>
            </Card>

            {/* Privileges */}
            {loyalty.privileges && (
              <Card title="Ваши привилегии" style={{ marginBottom: 24 }}>
                <Row gutter={[16, 12]}>
                  <Col xs={24} sm={12}>
                    <Text type="secondary">Скидка на бронирования</Text>
                    <Title level={5} style={{ margin: 0 }}>
                      {loyalty.privileges.discount_percent ?? 0}%
                    </Title>
                  </Col>
                  <Col xs={24} sm={12}>
                    <Text type="secondary">Множитель баллов</Text>
                    <Title level={5} style={{ margin: 0 }}>
                      x{loyalty.privileges.point_multiplier ?? 1}
                    </Title>
                  </Col>
                </Row>
              </Card>
            )}

            {/* Level info cards */}
            {levels.length > 0 && (
              <>
                <Title level={4} style={{ marginBottom: 16 }}>Уровни программы</Title>
                <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
                  {levels.map((level) => {
                    const cfg = LEVEL_CONFIG[level.level ?? ''] ?? DEFAULT_LEVEL_CFG
                    const isCurrentLevel = level.level === currentLevel
                    return (
                      <Col key={level.level} xs={24} sm={12} md={6}>
                        <Card
                          size="small"
                          style={{
                            borderColor: isCurrentLevel ? cfg.color : undefined,
                            borderWidth: isCurrentLevel ? 2 : 1,
                          }}
                        >
                          <div style={{ textAlign: 'center', marginBottom: 8 }}>
                            <div style={{ fontSize: 28, color: cfg.color }}>{cfg.icon}</div>
                            <Text strong style={{ color: cfg.color }}>{cfg.label}</Text>
                            {isCurrentLevel && <Tag color="blue" style={{ marginLeft: 8 }}>Текущий</Tag>}
                          </div>
                          <div>
                            <Text type="secondary" style={{ fontSize: 12 }}>
                              От {level.min_visits ?? 0} визитов
                            </Text>
                          </div>
                          <div>
                            <Text type="secondary" style={{ fontSize: 12 }}>
                              Скидка: {level.discount_percent ?? 0}%
                            </Text>
                          </div>
                          <div>
                            <Text type="secondary" style={{ fontSize: 12 }}>
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
            <Title level={4}>История операций</Title>
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
                    <div style={{ textAlign: 'center', marginTop: 16 }}>
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
                  <Empty
                    description="Пока нет операций с баллами"
                    style={{ marginTop: 24 }}
                  />
                )
              )}
            </Spin>
          </>
        ) : (
          !loadingLoyalty && (
            <Empty
              description="Программа лояльности пока недоступна"
              style={{ marginTop: 48 }}
            />
          )
        )}
      </Spin>
    </div>
  )
}
