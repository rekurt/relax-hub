import { useState, useMemo } from 'react'
import {
  Badge,
  Card,
  Col,
  Pagination,
  Progress,
  Row,
  Segmented,
  Select,
  Space,
  Spin,
  Statistic,
  Table,
  Tag,
  Tooltip,
  Typography,
} from '@/components/design/system'
import type { ColumnsType } from '@/components/design/types'
import {
  ClockCircleOutlined,
  ArrowUpOutlined,
  WarningOutlined,
} from '@/components/design/icons'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import dayjs from 'dayjs'
import {
  useGetAdminTickets,
  useGetAdminTicketsStats,
} from '@/api/generated/support-admin/support-admin'
import type { InternalHandlerTicketResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'
import { axiosInstance } from '@/api/axios-instance'
import PageHeader from '@/components/PageHeader'

interface OperationMetrics {
  fcr_percent: number
  aht_seconds: number
  avg_csat: number
  sla_compliance_percent: number
  total_resolved: number
  total_tickets: number
}

interface AgentThroughput {
  agent_id: string
  agent_name: string
  resolved_count: number
  avg_resolution_seconds: number
  csat_avg: number
}

function useOperationMetrics() {
  return useQuery({
    queryKey: ['admin', 'tickets', 'metrics'],
    queryFn: async () => {
      const { data } = await axiosInstance.get<{ data: OperationMetrics }>('/admin/tickets/metrics')
      return data.data
    },
    staleTime: 60_000,
  })
}

function useAgentThroughput() {
  return useQuery({
    queryKey: ['admin', 'tickets', 'agent-throughput'],
    queryFn: async () => {
      const { data } = await axiosInstance.get<{ data: AgentThroughput[] }>('/admin/tickets/agent-throughput')
      return data.data
    },
    staleTime: 60_000,
  })
}

function formatDuration(seconds: number): string {
  if (seconds < 3600) return `${Math.round(seconds / 60)} мин`
  const hours = Math.floor(seconds / 3600)
  const mins = Math.round((seconds % 3600) / 60)
  return mins > 0 ? `${hours} ч ${mins} мин` : `${hours} ч`
}

function metricStatClass(tone: 'success' | 'warning' | 'danger') {
  return `rh-admin-metric-stat rh-admin-metric-stat--${tone}`
}

const SLA_THRESHOLDS: Record<string, number> = {
  L1: 24,
  L2: 48,
  L3: 72,
}

function getSlaInfo(ticket: InternalHandlerTicketResponse) {
  if (ticket.status === 'resolved' || ticket.status === 'closed') return null
  if (!ticket.created_at) return null

  const level = ticket.level ?? 'L1'
  const thresholdHours = SLA_THRESHOLDS[level] ?? 24
  const created = dayjs(ticket.created_at)
  const deadline = created.add(thresholdHours, 'hour')
  const now = dayjs()
  const remainingHours = deadline.diff(now, 'hour', true)

  return {
    deadline,
    remainingHours,
    isBreached: remainingHours <= 0,
    isWarning: remainingHours > 0 && remainingHours <= 4,
  }
}

const { Text } = Typography

const STATUS_OPTIONS = [
  { label: 'Все', value: '' },
  { label: 'Открытые', value: 'open' },
  { label: 'В работе', value: 'in_progress' },
  { label: 'Эскалированные', value: 'escalated' },
  { label: 'Решённые', value: 'resolved' },
  { label: 'Закрытые', value: 'closed' },
]

const PRIORITY_OPTIONS = [
  { label: 'Все приоритеты', value: '' },
  { label: 'Низкий', value: 'low' },
  { label: 'Средний', value: 'medium' },
  { label: 'Высокий', value: 'high' },
  { label: 'Критический', value: 'critical' },
]

const LEVEL_OPTIONS = [
  { label: 'Все уровни', value: '' },
  { label: 'L1', value: 'L1' },
  { label: 'L2', value: 'L2' },
  { label: 'L3', value: 'L3' },
]

const CATEGORY_OPTIONS = [
  { label: 'Все категории', value: '' },
  { label: 'Вопрос', value: 'question' },
  { label: 'Проблема', value: 'problem' },
  { label: 'Жалоба', value: 'complaint' },
  { label: 'Возврат', value: 'refund_request' },
  { label: 'Аккаунт', value: 'account_issue' },
]

const statusLabel: Record<string, string> = {
  open: 'Открыт',
  in_progress: 'В работе',
  escalated: 'Эскалирован',
  resolved: 'Решён',
  closed: 'Закрыт',
}

const statusColor: Record<string, string> = {
  open: 'blue',
  in_progress: 'processing',
  escalated: 'orange',
  resolved: 'green',
  closed: 'default',
}

const categoryLabel: Record<string, string> = {
  question: 'Вопрос',
  problem: 'Проблема',
  complaint: 'Жалоба',
  refund_request: 'Возврат',
  account_issue: 'Аккаунт',
}

const priorityLabel: Record<string, string> = {
  low: 'Низкий',
  medium: 'Средний',
  high: 'Высокий',
  critical: 'Критический',
}

const priorityColor: Record<string, string> = {
  low: 'default',
  medium: 'blue',
  high: 'orange',
  critical: 'red',
}

export default function TicketManagement() {
  const navigate = useNavigate()

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [statusFilter, setStatusFilter] = useState('')
  const [priorityFilter, setPriorityFilter] = useState('')
  const [levelFilter, setLevelFilter] = useState('')
  const [categoryFilter, setCategoryFilter] = useState('')

  const { data, isLoading } = useGetAdminTickets({
    page,
    page_size: pageSize,
    ...(statusFilter && { status: statusFilter }),
    ...(priorityFilter && { priority: priorityFilter }),
    ...(levelFilter && { level: levelFilter }),
    ...(categoryFilter && { category: categoryFilter }),
  })

  const { data: statsData } = useGetAdminTicketsStats()
  const { data: metrics } = useOperationMetrics()
  const { data: agentThroughput } = useAgentThroughput()
  const [showAgentStats, setShowAgentStats] = useState(false)

  const meta = data?.meta
  const stats = statsData?.data

  const ticketsWithSla = useMemo(() => {
    const items: InternalHandlerTicketResponse[] = data?.data ?? []
    return items.map((t) => ({
      ...t,
      _sla: getSlaInfo(t),
    }))
  }, [data?.data])

  type TicketWithSla = InternalHandlerTicketResponse & {
    _sla: ReturnType<typeof getSlaInfo>
  }

  const columns: ColumnsType<TicketWithSla> = [
    {
      title: 'Тема',
      dataIndex: 'subject',
      key: 'subject',
      ellipsis: true,
      render: (text: string, record: TicketWithSla) => (
        <Space>
          <a onClick={() => navigate(`/admin/tickets/${record.id}`)}>
            {text || <Text type="secondary">Без темы</Text>}
          </a>
          {record.status === 'escalated' && (
            <Tooltip title={`Эскалирован на ${record.level ?? 'L2'}`}>
              <ArrowUpOutlined className="rh-admin-icon-warning" data-testid="escalation-icon" />
            </Tooltip>
          )}
        </Space>
      ),
    },
    {
      title: 'Категория',
      dataIndex: 'category',
      key: 'category',
      width: 120,
      render: (cat: string) => categoryLabel[cat] ?? cat,
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      width: 130,
      render: (status: string) => (
        <Tag color={statusColor[status] ?? 'default'}>
          {statusLabel[status] ?? status}
        </Tag>
      ),
    },
    {
      title: 'Приоритет',
      dataIndex: 'priority',
      key: 'priority',
      width: 120,
      render: (priority: string) => (
        <Tag color={priorityColor[priority] ?? 'default'}>
          {priorityLabel[priority] ?? priority}
        </Tag>
      ),
    },
    {
      title: 'Уровень',
      dataIndex: 'level',
      key: 'level',
      width: 80,
      render: (level: string) => <Tag>{level}</Tag>,
    },
    {
      title: 'SLA',
      key: 'sla',
      width: 140,
      render: (_: unknown, record: TicketWithSla) => {
        const sla = record._sla
        if (!sla) return <Text type="secondary">-</Text>
        if (sla.isBreached) {
          return (
            <Tooltip title={`Дедлайн: ${sla.deadline.format('DD.MM HH:mm')}`}>
              <Tag color="red" icon={<WarningOutlined />} data-testid="sla-breached">
                Просрочен
              </Tag>
            </Tooltip>
          )
        }
        const hours = Math.floor(sla.remainingHours)
        const mins = Math.round((sla.remainingHours - hours) * 60)
        const label = hours > 0 ? `${hours}ч ${mins}м` : `${mins}м`
        const percent = Math.max(0, Math.min(100, (1 - sla.remainingHours / (SLA_THRESHOLDS[record.level ?? 'L1'] ?? 24)) * 100))
        return (
          <Tooltip title={`Дедлайн: ${sla.deadline.format('DD.MM HH:mm')}`}>
            <div data-testid="sla-timer">
              <Progress
                percent={percent}
                size="small"
                strokeColor={sla.isWarning ? '#d97706' : '#15803d'}
                format={() => (
                  <span className="rh-admin-sla-label">
                    <ClockCircleOutlined /> {label}
                  </span>
                )}
              />
            </div>
          </Tooltip>
        )
      },
    },
    {
      title: 'Пользователь',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 120,
      render: (uid: string) =>
        uid ? (
          <Text copyable={{ text: uid }}>{uid.slice(0, 8)}...</Text>
        ) : (
          '-'
        ),
    },
    {
      title: 'Дата',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (date: string) => (date ? formatDateTime(date) : '-'),
    },
  ]

  return (
    <div className="rh-stack rh-admin-reference-page">
      <PageHeader
        eyebrow="Поддержка"
        title="Управление обращениями"
        description="SLA, статусы и операционные метрики поддержки в едином рабочем интерфейсе."
      />

      {stats && (
        <div className="rh-stat-grid">
          <div className="rh-stat-tile">
            <span className="rh-stat-tile__eyebrow">Открытые</span>
            <span className="rh-stat-tile__value"><Badge status="processing" /> {stats.open ?? 0}</span>
            <span className="rh-stat-tile__hint">Новые обращения без финального решения.</span>
          </div>
          <div className="rh-stat-tile">
            <span className="rh-stat-tile__eyebrow">В работе</span>
            <span className="rh-stat-tile__value">{stats.in_progress ?? 0}</span>
            <span className="rh-stat-tile__hint">Назначены агентам поддержки.</span>
          </div>
          <div className="rh-stat-tile">
            <span className="rh-stat-tile__eyebrow">Эскалированные</span>
            <span className="rh-stat-tile__value">{stats.escalated ?? 0}</span>
            <span className="rh-stat-tile__hint">Требуют повышенного уровня обработки.</span>
          </div>
          <div className="rh-stat-tile">
            <span className="rh-stat-tile__eyebrow">Решённые</span>
            <span className="rh-stat-tile__value">{stats.resolved ?? 0}</span>
            <span className="rh-stat-tile__hint">Закрыты решением агента.</span>
          </div>
          <div className="rh-stat-tile">
            <span className="rh-stat-tile__eyebrow">Закрытые</span>
            <span className="rh-stat-tile__value">{stats.closed ?? 0}</span>
            <span className="rh-stat-tile__hint">Финально архивированы.</span>
          </div>
        </div>
      )}

      {metrics && (
        <Card className="rh-admin-action-card" size="small" title="Операционные метрики">
          <Row gutter={[24, 16]}>
            <Col xs={12} sm={8} md={4}>
              <Statistic
                className={metricStatClass(metrics.fcr_percent >= 70 ? 'success' : 'warning')}
                title="FCR"
                value={metrics.fcr_percent}
                precision={1}
                suffix="%"
              />
            </Col>
            <Col xs={12} sm={8} md={4}>
              <Statistic
                title="AHT"
                value={formatDuration(metrics.aht_seconds)}
              />
            </Col>
            <Col xs={12} sm={8} md={4}>
              <Statistic
                className={
                  metricStatClass(
                    metrics.avg_csat >= 4 ? 'success' : metrics.avg_csat >= 3 ? 'warning' : 'danger',
                  )
                }
                title="CSAT"
                value={metrics.avg_csat}
                precision={1}
                suffix="/ 5"
              />
            </Col>
            <Col xs={12} sm={8} md={4}>
              <Statistic
                className={metricStatClass(metrics.sla_compliance_percent >= 90 ? 'success' : 'warning')}
                title="SLA (24ч)"
                value={metrics.sla_compliance_percent}
                precision={1}
                suffix="%"
              />
            </Col>
            <Col xs={12} sm={8} md={4}>
              <Statistic title="Решено" value={metrics.total_resolved} />
            </Col>
            <Col xs={12} sm={8} md={4}>
              <Statistic title="Всего" value={metrics.total_tickets} />
            </Col>
          </Row>
        </Card>
      )}

      {agentThroughput && agentThroughput.length > 0 && (
        <Card
          className="rh-admin-action-card"
          size="small"
          title="Производительность агентов"
          extra={
            <a onClick={() => setShowAgentStats(!showAgentStats)}>
              {showAgentStats ? 'Скрыть' : 'Показать'}
            </a>
          }
          data-testid="agent-throughput-card"
        >
          {showAgentStats && (
            <Table
              dataSource={agentThroughput}
              rowKey="agent_id"
              pagination={false}
              size="small"
              columns={[
                { title: 'Агент', dataIndex: 'agent_name', key: 'agent_name' },
                { title: 'Решено', dataIndex: 'resolved_count', key: 'resolved_count', width: 100 },
                {
                  title: 'Среднее время',
                  dataIndex: 'avg_resolution_seconds',
                  key: 'avg_resolution_seconds',
                  width: 140,
                  render: (v: number) => formatDuration(v),
                },
                {
                  title: 'CSAT',
                  dataIndex: 'csat_avg',
                  key: 'csat_avg',
                  width: 80,
                  render: (v: number) => v?.toFixed(1) ?? '-',
                },
              ]}
            />
          )}
        </Card>
      )}

      <Card className="rh-admin-filter-card" title="Фильтры">
        <Space className="rh-admin-filter-column" orientation="vertical" size={16}>
          <Segmented
            options={STATUS_OPTIONS}
            value={statusFilter}
            onChange={(val) => {
              setStatusFilter(val as string)
              setPage(1)
            }}
          />
          <Space wrap>
            <Select
              className="rh-admin-filter-select"
              placeholder="Приоритет"
              options={PRIORITY_OPTIONS}
              value={priorityFilter}
              onChange={(val) => {
                setPriorityFilter(val)
                setPage(1)
              }}
              allowClear
              onClear={() => {
                setPriorityFilter('')
                setPage(1)
              }}
            />
            <Select
              className="rh-admin-filter-select"
              placeholder="Уровень"
              options={LEVEL_OPTIONS}
              value={levelFilter}
              onChange={(val) => {
                setLevelFilter(val)
                setPage(1)
              }}
              allowClear
              onClear={() => {
                setLevelFilter('')
                setPage(1)
              }}
            />
            <Select
              className="rh-admin-filter-select"
              placeholder="Категория"
              options={CATEGORY_OPTIONS}
              value={categoryFilter}
              onChange={(val) => {
                setCategoryFilter(val)
                setPage(1)
              }}
              allowClear
              onClear={() => {
                setCategoryFilter('')
                setPage(1)
              }}
            />
          </Space>
        </Space>
      </Card>

      <Card className="rh-admin-reference-card" title="Список обращений">
        {isLoading ? (
          <div className="rh-admin-state-card">
            <Spin size="large" />
            <span>Загружаем обращения</span>
          </div>
        ) : ticketsWithSla.length === 0 ? (
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Нет обращений</div>
            <p className="rh-admin-empty-state__text">
              Новые тикеты поддержки появятся здесь после обращения пользователя.
            </p>
          </div>
        ) : (
          <>
            <Table
              dataSource={ticketsWithSla}
              columns={columns}
              rowKey="id"
              pagination={false}
              size="middle"
              locale={{ emptyText: 'Нет обращений' }}
              onRow={(record) => ({
                onClick: () => navigate(`/admin/tickets/${record.id}`),
                className: 'rh-clickable-row',
              })}
            />
            {meta && meta.total_pages! > 1 && (
              <div className="rh-admin-pagination">
                <Pagination
                  current={page}
                  pageSize={pageSize}
                  total={meta.total_count}
                  showSizeChanger
                  pageSizeOptions={['10', '20', '50']}
                  showTotal={(total) => `Всего: ${total}`}
                  onChange={(p, ps) => {
                    setPage(p)
                    setPageSize(ps)
                  }}
                />
              </div>
            )}
          </>
        )}
      </Card>
    </div>
  )
}
