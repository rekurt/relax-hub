import { useState } from 'react'
import {
  Badge,
  Card,
  Col,
  Empty,
  Pagination,
  Row,
  Segmented,
  Select,
  Space,
  Spin,
  Statistic,
  Table,
  Tag,
  Typography,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import {
  useGetAdminTickets,
  useGetAdminTicketsStats,
} from '@/api/generated/support-admin/support-admin'
import type { InternalHandlerTicketResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'
import { axiosInstance } from '@/api/axios-instance'

interface OperationMetrics {
  fcr_percent: number
  aht_seconds: number
  avg_csat: number
  sla_compliance_percent: number
  total_resolved: number
  total_tickets: number
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

function formatDuration(seconds: number): string {
  if (seconds < 3600) return `${Math.round(seconds / 60)} мин`
  const hours = Math.floor(seconds / 3600)
  const mins = Math.round((seconds % 3600) / 60)
  return mins > 0 ? `${hours} ч ${mins} мин` : `${hours} ч`
}

const { Title, Text } = Typography

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

  const tickets: InternalHandlerTicketResponse[] = data?.data ?? []
  const meta = data?.meta
  const stats = statsData?.data

  const columns: ColumnsType<InternalHandlerTicketResponse> = [
    {
      title: 'Тема',
      dataIndex: 'subject',
      key: 'subject',
      ellipsis: true,
      render: (text: string, record: InternalHandlerTicketResponse) => (
        <a onClick={() => navigate(`/admin/tickets/${record.id}`)}>
          {text || <Text type="secondary">Без темы</Text>}
        </a>
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
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>
        Управление обращениями
      </Title>

      {stats && (
        <Space wrap style={{ marginBottom: 16 }}>
          <Card size="small">
            <Statistic
              title="Открытые"
              value={stats.open ?? 0}
              valueStyle={{ color: '#1677ff' }}
              prefix={<Badge status="processing" />}
            />
          </Card>
          <Card size="small">
            <Statistic
              title="В работе"
              value={stats.in_progress ?? 0}
              valueStyle={{ color: '#1677ff' }}
            />
          </Card>
          <Card size="small">
            <Statistic
              title="Эскалированные"
              value={stats.escalated ?? 0}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
          <Card size="small">
            <Statistic
              title="Решённые"
              value={stats.resolved ?? 0}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
          <Card size="small">
            <Statistic title="Закрытые" value={stats.closed ?? 0} />
          </Card>
        </Space>
      )}

      {metrics && (
        <Card size="small" title="Операционные метрики" style={{ marginBottom: 16 }}>
          <Row gutter={[24, 16]}>
            <Col xs={12} sm={8} md={4}>
              <Statistic
                title="FCR"
                value={metrics.fcr_percent}
                precision={1}
                suffix="%"
                valueStyle={{ color: metrics.fcr_percent >= 70 ? '#52c41a' : '#fa8c16' }}
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
                title="CSAT"
                value={metrics.avg_csat}
                precision={1}
                suffix="/ 5"
                valueStyle={{ color: metrics.avg_csat >= 4 ? '#52c41a' : metrics.avg_csat >= 3 ? '#fa8c16' : '#f5222d' }}
              />
            </Col>
            <Col xs={12} sm={8} md={4}>
              <Statistic
                title="SLA (24ч)"
                value={metrics.sla_compliance_percent}
                precision={1}
                suffix="%"
                valueStyle={{ color: metrics.sla_compliance_percent >= 90 ? '#52c41a' : '#fa8c16' }}
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

      <Space direction="vertical" size={16} style={{ width: '100%', marginBottom: 16 }}>
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
            placeholder="Приоритет"
            options={PRIORITY_OPTIONS}
            value={priorityFilter}
            onChange={(val) => {
              setPriorityFilter(val)
              setPage(1)
            }}
            style={{ width: 160 }}
            allowClear
            onClear={() => {
              setPriorityFilter('')
              setPage(1)
            }}
          />
          <Select
            placeholder="Уровень"
            options={LEVEL_OPTIONS}
            value={levelFilter}
            onChange={(val) => {
              setLevelFilter(val)
              setPage(1)
            }}
            style={{ width: 140 }}
            allowClear
            onClear={() => {
              setLevelFilter('')
              setPage(1)
            }}
          />
          <Select
            placeholder="Категория"
            options={CATEGORY_OPTIONS}
            value={categoryFilter}
            onChange={(val) => {
              setCategoryFilter(val)
              setPage(1)
            }}
            style={{ width: 160 }}
            allowClear
            onClear={() => {
              setCategoryFilter('')
              setPage(1)
            }}
          />
        </Space>
      </Space>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: 48 }}>
          <Spin size="large" />
        </div>
      ) : tickets.length === 0 ? (
        <Empty description="Нет обращений" />
      ) : (
        <>
          <Table
            dataSource={tickets}
            columns={columns}
            rowKey="id"
            pagination={false}
            size="middle"
            locale={{ emptyText: 'Нет обращений' }}
            onRow={(record) => ({
              onClick: () => navigate(`/admin/tickets/${record.id}`),
              style: { cursor: 'pointer' },
            })}
          />
          {meta && meta.total_pages! > 1 && (
            <div style={{ marginTop: 16, textAlign: 'right' }}>
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
    </div>
  )
}

