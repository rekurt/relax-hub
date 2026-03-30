import { useState } from 'react'
import { Card, Col, Row, Segmented, Spin, Statistic, Typography } from 'antd'
import {
  FunnelPlotOutlined,
  ArrowDownOutlined,
} from '@ant-design/icons'
import { useGetAdminAnalyticsFunnel } from '@/api/generated/admin-analytics/admin-analytics'
import type { GithubComNikitaaldaevBaniInternalDomainFunnelStep } from '@/api/generated/model'

const { Title, Text } = Typography

const PERIOD_OPTIONS = [
  { label: 'День', value: '1d' },
  { label: 'Неделя', value: '7d' },
  { label: 'Месяц', value: '30d' },
  { label: '3 месяца', value: '90d' },
]

const STEP_COLORS = [
  '#1677ff',
  '#13c2c2',
  '#52c41a',
  '#faad14',
  '#f5222d',
  '#722ed1',
]

export default function ConversionFunnels() {
  const [period, setPeriod] = useState('30d')

  const { data: funnelData, isLoading } = useGetAdminAnalyticsFunnel({ period })

  const funnel = funnelData?.data
  const steps = funnel?.steps ?? []

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={3} style={{ margin: 0 }}>Воронка конверсии</Title>
        <Segmented
          options={PERIOD_OPTIONS}
          value={period}
          onChange={(v) => setPeriod(v as string)}
        />
      </div>

      <Spin spinning={isLoading}>
        {steps.length > 0 && (
          <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
            <Col xs={24} sm={12} lg={6}>
              <Card>
                <Statistic
                  title="Начало воронки"
                  value={steps[0]?.count ?? 0}
                  prefix={<FunnelPlotOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card>
                <Statistic
                  title="Конец воронки"
                  value={steps[steps.length - 1]?.count ?? 0}
                  prefix={<FunnelPlotOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card>
                <Statistic
                  title="Общая конверсия"
                  value={steps[steps.length - 1]?.percentage ?? 0}
                  precision={1}
                  suffix="%"
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card>
                <Statistic
                  title="Этапов"
                  value={steps.length}
                />
              </Card>
            </Col>
          </Row>
        )}

        <Card title="Этапы воронки">
          {steps.length === 0 ? (
            <Text type="secondary">Нет данных</Text>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {steps.map((step: GithubComNikitaaldaevBaniInternalDomainFunnelStep, index: number) => {
                const percentage = step.percentage ?? 0
                const color = STEP_COLORS[index % STEP_COLORS.length]
                const nextStep = steps[index + 1]
                const dropOff = nextStep
                  ? ((step.count ?? 0) - (nextStep.count ?? 0))
                  : null

                return (
                  <div key={step.name ?? index}>
                    <div
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 16,
                        padding: '12px 16px',
                        borderRadius: 8,
                        background: '#fafafa',
                      }}
                    >
                      <div style={{ minWidth: 160, fontWeight: 500 }}>
                        {step.name ?? `Этап ${index + 1}`}
                      </div>
                      <div style={{ flex: 1 }}>
                        <div
                          style={{
                            height: 32,
                            width: `${Math.max(percentage, 2)}%`,
                            background: color,
                            borderRadius: 4,
                            display: 'flex',
                            alignItems: 'center',
                            paddingLeft: 8,
                            color: '#fff',
                            fontWeight: 500,
                            fontSize: 13,
                            minWidth: 60,
                            transition: 'width 0.3s ease',
                          }}
                        >
                          {(step.count ?? 0).toLocaleString('ru-RU')}
                        </div>
                      </div>
                      <div style={{ minWidth: 60, textAlign: 'right', fontWeight: 500 }}>
                        {percentage.toFixed(1)}%
                      </div>
                    </div>

                    {dropOff !== null && (
                      <div
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: 8,
                          padding: '4px 0 4px 40px',
                          color: '#999',
                          fontSize: 12,
                        }}
                      >
                        <ArrowDownOutlined />
                        <span>
                          Отсев: {dropOff.toLocaleString('ru-RU')} (
                          {((step.count ?? 0) > 0
                            ? ((dropOff / (step.count ?? 1)) * 100).toFixed(1)
                            : '0.0'
                          )}%)
                        </span>
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </Card>
      </Spin>
    </div>
  )
}
