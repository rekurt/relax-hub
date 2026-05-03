import { useState } from 'react'
import { Card, Segmented, Spin, Table } from '@/components/design/system'
import type { ColumnsType } from '@/components/design/types'
import { useGetAdminAnalyticsCohorts } from '@/api/generated/admin-analytics/admin-analytics'
import type { GithubComRekurtRelaxHubInternalDomainCohortRow } from '@/api/generated/model'
import { formatPrice } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

const MONTHS_OPTIONS = [
  { label: '3 мес', value: 3 },
  { label: '6 мес', value: 6 },
  { label: '12 мес', value: 12 },
  { label: '24 мес', value: 24 },
]

function retentionTone(value: number): string {
  if (value >= 80) return 'high'
  if (value >= 60) return 'good'
  if (value >= 40) return 'medium'
  return 'low'
}

export default function CohortAnalysis() {
  const [months, setMonths] = useState(6)

  const { data: cohortData, isLoading } = useGetAdminAnalyticsCohorts({ months })

  const cohorts = cohortData?.data?.cohorts ?? []

  const maxWeeks = cohorts.reduce(
    (max, row) => Math.max(max, (row.retention_weeks ?? []).length),
    0,
  )

  const columns: ColumnsType<GithubComRekurtRelaxHubInternalDomainCohortRow> = [
    {
      title: 'Когорта',
      dataIndex: 'cohort_month',
      key: 'cohort_month',
      fixed: 'left',
      width: 100,
      render: (v: string) => v ?? '—',
    },
    {
      title: 'Пользователей',
      dataIndex: 'users_count',
      key: 'users_count',
      width: 120,
      sorter: (a, b) => (a.users_count ?? 0) - (b.users_count ?? 0),
    },
    {
      title: 'Расход',
      dataIndex: 'total_spending',
      key: 'total_spending',
      width: 120,
      render: (v: number) => formatPrice(v ?? 0),
      sorter: (a, b) => (a.total_spending ?? 0) - (b.total_spending ?? 0),
    },
    ...Array.from({ length: maxWeeks }, (_, i) => ({
      title: `Н${i}`,
      key: `week_${i}`,
      width: 70,
      render: (_: unknown, record: GithubComRekurtRelaxHubInternalDomainCohortRow) => {
        const weeks = record.retention_weeks ?? []
        if (i >= weeks.length) return '—'
        const val = weeks[i] ?? 0
        return (
          <div className={`rh-cohort-retention rh-cohort-retention--${retentionTone(val)}`}>
            {val.toFixed(0)}%
          </div>
        )
      },
    })),
  ]

  return (
    <div className="rh-stack rh-admin-reference-page rh-cohort-page">
      <PageHeader
        eyebrow="Аналитика"
        title="Когортный анализ"
        description="Матрица удержания клиентов по неделям с аккуратной шкалой интенсивности."
        extra={
          <Segmented
            options={MONTHS_OPTIONS}
            value={months}
            onChange={(v) => setMonths(v as number)}
          />
        }
      />

      <Spin spinning={isLoading}>
        <Card className="rh-admin-reference-card" title="Матрица удержания">
          <Table
            columns={columns}
            dataSource={cohorts}
            rowKey="cohort_month"
            scroll={{ x: 'max-content' }}
            pagination={false}
            locale={{ emptyText: 'Нет данных' }}
            size="small"
          />
        </Card>
      </Spin>
    </div>
  )
}
