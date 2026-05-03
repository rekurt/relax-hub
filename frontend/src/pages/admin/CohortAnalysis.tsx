import { useState } from 'react'
import { Card, Segmented, Spin, Table, Typography } from '@/components/design/system'
import type { ColumnsType } from '@/components/design/types'
import { useGetAdminAnalyticsCohorts } from '@/api/generated/admin-analytics/admin-analytics'
import type { GithubComRekurtRelaxHubInternalDomainCohortRow } from '@/api/generated/model'
import { formatPrice } from '@/lib/format'

const { Title } = Typography

const MONTHS_OPTIONS = [
  { label: '3 мес', value: 3 },
  { label: '6 мес', value: 6 },
  { label: '12 мес', value: 12 },
  { label: '24 мес', value: 24 },
]

function retentionColor(value: number): string {
  if (value >= 80) return 'rgba(21, 128, 61, 0.08)'
  if (value >= 60) return '#fcffe6'
  if (value >= 40) return '#fff7e6'
  if (value >= 20) return 'rgba(180, 35, 24, 0.08)'
  return 'rgba(180, 35, 24, 0.08)'
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
          <div
            style={{
              background: retentionColor(val),
              padding: '2px 6px',
              borderRadius: 12,
              textAlign: 'center',
              fontWeight: val >= 50 ? 600 : 400,
            }}
          >
            {val.toFixed(0)}%
          </div>
        )
      },
    })),
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={3} style={{ margin: 0 }}>Когортный анализ</Title>
        <Segmented
          options={MONTHS_OPTIONS}
          value={months}
          onChange={(v) => setMonths(v as number)}
        />
      </div>

      <Spin spinning={isLoading}>
        <Card>
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
