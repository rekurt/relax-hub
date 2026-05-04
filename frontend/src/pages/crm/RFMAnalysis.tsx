import { useState, useMemo } from 'react'
import { Card, Col, Row, Table, Tag, Tooltip, Typography, Spin } from '@/components/design/system'
import { useQuery } from '@tanstack/react-query'
import dayjs from 'dayjs'
import { axiosInstance } from '@/api/axios-instance'
import { formatPrice } from '@/lib/format'
import EmptyState from '@/components/EmptyState'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

interface RFMScore {
  recency: number
  frequency: number
  monetary: number
}

interface RFMGuest {
  id: string
  client_id: string
  bathhouse_id: string
  last_visit_at: string
  visit_count: number
  total_spent: number
  avg_check: number
  tags: string[]
  rfm: RFMScore
}

interface RFMMatrixCell {
  recency: number
  frequency: number
  count: number
}

interface RFMAnalysisData {
  guests: RFMGuest[]
  matrix: RFMMatrixCell[]
}

function useRFMAnalysis() {
  return useQuery({
    queryKey: ['crm', 'rfm'],
    queryFn: async () => {
      const { data } = await axiosInstance.get<{ data: RFMAnalysisData }>('/my/crm/rfm')
      return data.data
    },
  })
}

const SCORE_COLORS: Record<number, string> = {
  1: '#b42318',
  2: '#d97706',
  3: '#d97706',
  4: '#15803d',
  5: '#15803d',
}

const SCORE_LABELS: Record<number, string> = {
  1: 'Очень низкий',
  2: 'Низкий',
  3: 'Средний',
  4: 'Высокий',
  5: 'Очень высокий',
}

function ScoreTag({ value }: { value: number }) {
  return (
    <Tooltip title={SCORE_LABELS[value]}>
      <Tag color={SCORE_COLORS[value]} className="rh-score-tag">
        {value}
      </Tag>
    </Tooltip>
  )
}

function RFMMatrix({ matrix }: { matrix: RFMMatrixCell[] }) {
  const matrixMap = useMemo(() => {
    const m: Record<string, number> = {}
    for (const cell of matrix) {
      m[`${cell.recency}-${cell.frequency}`] = cell.count
    }
    return m
  }, [matrix])

  const maxCount = useMemo(() => {
    return Math.max(1, ...matrix.map((c) => c.count))
  }, [matrix])

  return (
    <Card title="Матрица RFM (Давность x Частота)" size="small">
      <div className="rh-rfm-matrix">
        <table className="rh-rfm-matrix__table">
          <thead>
            <tr>
              <th className="rh-rfm-matrix__head">R \ F</th>
              {[1, 2, 3, 4, 5].map((f) => (
                <th key={f} className="rh-rfm-matrix__head">
                  F={f}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {[5, 4, 3, 2, 1].map((r) => (
              <tr key={r}>
                <td className="rh-rfm-matrix__head rh-rfm-matrix__head--row">
                  R={r}
                </td>
                {[1, 2, 3, 4, 5].map((f) => {
                  const count = matrixMap[`${r}-${f}`] ?? 0
                  const intensity = count / maxCount
                  const level = count > 0 ? Math.max(1, Math.ceil(intensity * 5)) : 0
                  return (
                    <td
                      key={f}
                      className={`rh-rfm-matrix__cell rh-rfm-matrix__cell--level-${level}`}
                    >
                      {count || ''}
                    </td>
                  )
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <Text type="secondary" className="rh-field-help-text">
        R = Давность (5 = недавний визит), F = Частота (5 = много визитов)
      </Text>
    </Card>
  )
}

export default function RFMAnalysis() {
  const { data, isLoading } = useRFMAnalysis()
  const [pageSize] = useState(20)

  const guests = useMemo(() => data?.guests ?? [], [data?.guests])
  const matrix = data?.matrix ?? []

  const segmentSummary = useMemo(() => {
    const champions = guests.filter((g) => g.rfm.recency >= 4 && g.rfm.frequency >= 4).length
    const loyal = guests.filter((g) => g.rfm.frequency >= 4 && g.rfm.recency < 4).length
    const atRisk = guests.filter((g) => g.rfm.recency <= 2 && g.rfm.frequency >= 3).length
    const lost = guests.filter((g) => g.rfm.recency <= 2 && g.rfm.frequency <= 2).length
    return { champions, loyal, atRisk, lost }
  }, [guests])

  const columns = [
    {
      title: 'ID клиента',
      dataIndex: 'client_id',
      key: 'client_id',
      render: (id: string) => <Tag className="rh-code-tag">{id?.slice(0, 8)}...</Tag>,
    },
    {
      title: 'R',
      dataIndex: ['rfm', 'recency'],
      key: 'recency',
      width: 70,
      sorter: (a: RFMGuest, b: RFMGuest) => a.rfm.recency - b.rfm.recency,
      render: (v: number) => <ScoreTag value={v} />,
    },
    {
      title: 'F',
      dataIndex: ['rfm', 'frequency'],
      key: 'frequency',
      width: 70,
      sorter: (a: RFMGuest, b: RFMGuest) => a.rfm.frequency - b.rfm.frequency,
      render: (v: number) => <ScoreTag value={v} />,
    },
    {
      title: 'M',
      dataIndex: ['rfm', 'monetary'],
      key: 'monetary',
      width: 70,
      sorter: (a: RFMGuest, b: RFMGuest) => a.rfm.monetary - b.rfm.monetary,
      render: (v: number) => <ScoreTag value={v} />,
    },
    {
      title: 'Визиты',
      dataIndex: 'visit_count',
      key: 'visit_count',
      width: 90,
      sorter: (a: RFMGuest, b: RFMGuest) => a.visit_count - b.visit_count,
    },
    {
      title: 'Общая сумма',
      dataIndex: 'total_spent',
      key: 'total_spent',
      width: 130,
      sorter: (a: RFMGuest, b: RFMGuest) => a.total_spent - b.total_spent,
      render: (v: number) => formatPrice(v ?? 0),
    },
    {
      title: 'Последний визит',
      dataIndex: 'last_visit_at',
      key: 'last_visit_at',
      width: 140,
      render: (d: string) => (d ? dayjs(d).format('DD.MM.YYYY') : '—'),
    },
    {
      title: 'Теги',
      dataIndex: 'tags',
      key: 'tags',
      render: (tags: string[]) => tags?.length ? tags.map((t) => <Tag key={t}>{t}</Tag>) : '—',
    },
  ]

  if (isLoading) {
    return (
      <div className="rh-loading-block">
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="CRM"
        title="RFM-анализ"
        description="Анализ гостей по давности визита (R), частоте (F) и сумме трат (M). Каждый показатель от 1 до 5."
        size="compact"
      />

      <Row gutter={[16, 16]} className="rh-section-spaced">
        <Col xs={12} sm={6}>
          <Card size="small">
            <Text type="secondary">Чемпионы</Text>
            <div className="rh-rfm-summary-value rh-rfm-summary-value--success">
              {segmentSummary.champions}
            </div>
            <Text type="secondary" className="rh-rfm-summary-note">R≥4, F≥4</Text>
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small">
            <Text type="secondary">Лояльные</Text>
            <div className="rh-rfm-summary-value rh-rfm-summary-value--primary">
              {segmentSummary.loyal}
            </div>
            <Text type="secondary" className="rh-rfm-summary-note">F≥4, R&lt;4</Text>
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small">
            <Text type="secondary">Под угрозой</Text>
            <div className="rh-rfm-summary-value rh-rfm-summary-value--warning">
              {segmentSummary.atRisk}
            </div>
            <Text type="secondary" className="rh-rfm-summary-note">R≤2, F≥3</Text>
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small">
            <Text type="secondary">Потерянные</Text>
            <div className="rh-rfm-summary-value rh-rfm-summary-value--danger">
              {segmentSummary.lost}
            </div>
            <Text type="secondary" className="rh-rfm-summary-note">R≤2, F≤2</Text>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} className="rh-section-spaced">
        <Col xs={24} lg={12}>
          <RFMMatrix matrix={matrix} />
        </Col>
        <Col xs={24} lg={12}>
          <Card title="Подсказка по интерпретации" size="small">
            <div className="rh-rfm-legend">
              <Tag color="#15803d">5</Tag> Лучший квинтиль (топ-20%)<br />
              <Tag color="#15803d">4</Tag> Выше среднего<br />
              <Tag color="#d97706">3</Tag> Средний<br />
              <Tag color="#d97706">2</Tag> Ниже среднего<br />
              <Tag color="#b42318">1</Tag> Нижний квинтиль (нижние 20%)<br />
            </div>
            <Text type="secondary" className="rh-field-help-text rh-section-offset">
              Чемпионы (R≥4, F≥4) - самые ценные гости. Потерянные (R≤2, F≤2) - давно не приходили.
              Используйте эти данные для таргетированных рассылок.
            </Text>
          </Card>
        </Col>
      </Row>

      <Card title={`Гости с RFM-скорами (${guests.length})`} size="small">
        <Table
          dataSource={guests}
          columns={columns}
          rowKey="id"
          size="small"
          pagination={{ pageSize, showSizeChanger: false }}
          locale={{ emptyText: <EmptyState description="Нет данных для RFM-анализа. Гости появятся после завершённых бронирований" /> }}
        />
      </Card>
    </div>
  )
}
