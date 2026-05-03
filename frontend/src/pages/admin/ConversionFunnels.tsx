import { useState, type CSSProperties } from 'react'
import { Card, Col, Row, Segmented, Spin, Statistic, Typography } from '@/components/design/system'
import {
  FunnelPlotOutlined,
  ArrowDownOutlined,
} from '@/components/design/icons'
import { useGetAdminAnalyticsFunnel } from '@/api/generated/admin-analytics/admin-analytics'
import type { GithubComRekurtRelaxHubInternalDomainFunnelStep } from '@/api/generated/model'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

const PERIOD_OPTIONS = [
  { label: 'День', value: '1d' },
  { label: 'Неделя', value: '7d' },
  { label: 'Месяц', value: '30d' },
  { label: '3 месяца', value: '90d' },
]

const STEP_COLORS = [
  '#0f766e',
  '#0f766e',
  '#15803d',
  '#d97706',
  '#b42318',
  '#0a5f59',
]

function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(' ')
}

function clampFunnelBarWidth(percentage: number) {
  if (!Number.isFinite(percentage) || percentage <= 0) return 2
  return Math.min(Math.max(percentage, 2), 100)
}

function formatFunnelChange(currentCount: number, nextCount: number) {
  const delta = nextCount - currentCount
  const amount = Math.abs(delta)
  const rate = currentCount > 0
    ? (amount / currentCount) * 100
    : delta > 0
      ? 100
      : 0

  if (delta > 0) {
    return {
      kind: 'growth' as const,
      label: 'Прирост',
      amount: `+${amount.toLocaleString('ru-RU')}`,
      rate: rate.toFixed(1),
    }
  }

  if (delta < 0) {
    return {
      kind: 'drop' as const,
      label: 'Отсев',
      amount: amount.toLocaleString('ru-RU'),
      rate: rate.toFixed(1),
    }
  }

  return {
    kind: 'flat' as const,
    label: 'Без изменений',
    amount: '0',
    rate: '0.0',
  }
}

export default function ConversionFunnels() {
  const [period, setPeriod] = useState('30d')

  const { data: funnelData, isLoading } = useGetAdminAnalyticsFunnel({ period })

  const funnel = funnelData?.data
  const steps = funnel?.steps ?? []

  return (
    <div className="rh-funnel-page">
      <PageHeader
        title="Воронка конверсии"
        description="Контроль переходов между ключевыми этапами клиентского пути."
        size="compact"
        extra={(
          <Segmented
            className="rh-funnel-page__periods"
            options={PERIOD_OPTIONS}
            value={period}
            onChange={(v) => setPeriod(v as string)}
          />
        )}
      />

      <Spin spinning={isLoading}>
        {steps.length > 0 && (
          <Row gutter={[16, 16]} className="rh-funnel-page__summary">
            <Col xs={24} sm={12} lg={6}>
              <Card className="rh-funnel-page__summary-card">
                <Statistic
                  title="Начало воронки"
                  value={steps[0]?.count ?? 0}
                  prefix={<FunnelPlotOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card className="rh-funnel-page__summary-card">
                <Statistic
                  title="Конец воронки"
                  value={steps[steps.length - 1]?.count ?? 0}
                  prefix={<FunnelPlotOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card className="rh-funnel-page__summary-card">
                <Statistic
                  title="Общая конверсия"
                  value={steps[steps.length - 1]?.percentage ?? 0}
                  precision={1}
                  suffix="%"
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card className="rh-funnel-page__summary-card">
                <Statistic
                  title="Этапов"
                  value={steps.length}
                />
              </Card>
            </Col>
          </Row>
        )}

        <Card title="Этапы воронки" className="rh-funnel-card">
          {steps.length === 0 ? (
            <Text type="secondary">Нет данных</Text>
          ) : (
            <div className="rh-funnel__list">
              {steps.map((step: GithubComRekurtRelaxHubInternalDomainFunnelStep, index: number) => {
                const percentage = step.percentage ?? 0
                const barWidth = clampFunnelBarWidth(percentage)
                const color = STEP_COLORS[index % STEP_COLORS.length]
                const currentCount = step.count ?? 0
                const nextStep = steps[index + 1]
                const change = nextStep
                  ? formatFunnelChange(currentCount, nextStep.count ?? 0)
                  : null
                const stepName = step.name ?? `Этап ${index + 1}`

                return (
                  <div className="rh-funnel__step" key={step.name ?? index}>
                    <div className="rh-funnel__row">
                      <div className="rh-funnel__label">
                        {stepName}
                      </div>
                      <div
                        className="rh-funnel__track"
                        aria-label={`${stepName}: ${percentage.toFixed(1)}%`}
                      >
                        <div
                          className="rh-funnel__bar"
                          data-testid={`funnel-bar-${index}`}
                          role="meter"
                          aria-valuemin={0}
                          aria-valuemax={100}
                          aria-valuenow={barWidth}
                          style={{
                            '--rh-funnel-bar-width': `${barWidth}%`,
                            '--rh-funnel-bar-color': color,
                          } as CSSProperties}
                        >
                          <span className="rh-funnel__count">
                            {currentCount.toLocaleString('ru-RU')}
                          </span>
                        </div>
                      </div>
                      <div className="rh-funnel__percentage">
                        {percentage.toFixed(1)}%
                      </div>
                    </div>

                    {change && (
                      <div className={cx('rh-funnel__change', `rh-funnel__change--${change.kind}`)}>
                        <ArrowDownOutlined className="rh-funnel__change-icon" />
                        <span>
                          {change.label}: {change.amount} ({change.rate}%)
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
