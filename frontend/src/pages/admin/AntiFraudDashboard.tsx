import { useState } from 'react'
import {
  Typography,
  Table,
  Select,
  Tag,
  Button,
  Space,
  Descriptions,
  Drawer,
  App,
  Segmented,
  Statistic,
  Row,
  Col,
  Card,
} from 'antd'
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetAdminAntifraudFlags,
  usePatchAdminAntifraudFlagsId,
} from '@/api/generated/admin-antifraud/admin-antifraud'
import type { InternalHandlerFraudFlagResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'

const { Title, Text } = Typography

const STATUS_OPTIONS = [
  { value: '', label: 'Все' },
  { value: 'pending', label: 'Ожидают' },
  { value: 'reviewed', label: 'Рассмотрены' },
  { value: 'dismissed', label: 'Отклонены' },
]

const RULE_LABELS: Record<string, string> = {
  multi_card_topup: 'Мульти-карта пополнение',
  topup_cancel_cycle: 'Цикл пополнение/отмена',
  dormant_balance: 'Неактивный баланс',
  rapid_bookings: 'Быстрые бронирования',
  self_booking: 'Самобронирование',
  structuring: 'Дробление операций',
  fake_reviews: 'Фейковые отзывы',
}

const SEVERITY_COLORS: Record<string, string> = {
  low: 'blue',
  medium: 'orange',
  high: 'red',
  critical: 'volcano',
}

const SEVERITY_LABELS: Record<string, string> = {
  low: 'Низкий',
  medium: 'Средний',
  high: 'Высокий',
  critical: 'Критический',
}

const ACTION_LABELS: Record<string, string> = {
  flag: 'Пометка',
  freeze_wallet: 'Заморозка кошелька',
  block_user: 'Блокировка',
  notify_admin: 'Уведомление',
  block: 'Блокировка операции',
}

const STATUS_LABELS: Record<string, string> = {
  pending: 'Ожидает',
  reviewed: 'Рассмотрен',
  dismissed: 'Отклонён',
  action_taken: 'Действие выполнено',
}

const STATUS_COLORS: Record<string, string> = {
  pending: 'warning',
  reviewed: 'success',
  dismissed: 'default',
  action_taken: 'processing',
}

export default function AntiFraudDashboard() {
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState('')
  const [ruleFilter, setRuleFilter] = useState<string | undefined>()
  const [selectedFlag, setSelectedFlag] = useState<InternalHandlerFraudFlagResponse | null>(null)

  const { message } = App.useApp()
  const queryClient = useQueryClient()

  const { data, isLoading } = useGetAdminAntifraudFlags({
    page,
    page_size: 20,
    status: statusFilter || undefined,
    rule: ruleFilter,
  })
  const flags = data?.data ?? []
  const meta = data?.meta

  const updateMutation = usePatchAdminAntifraudFlagsId({
    mutation: {
      onSuccess: () => {
        message.success('Статус обновлён')
        queryClient.invalidateQueries({ queryKey: ['/admin/antifraud/flags'] })
        setSelectedFlag(null)
      },
      onError: () => message.error('Не удалось обновить статус'),
    },
  })

  const handleReview = (id: string, status: string) => {
    updateMutation.mutate({ id, data: { status } })
  }

  const pendingCount = flags.filter((f) => f.status === 'pending').length

  const columns = [
    {
      title: 'Правило',
      dataIndex: 'rule',
      key: 'rule',
      render: (rule: string) => RULE_LABELS[rule] || rule,
    },
    {
      title: 'Серьёзность',
      dataIndex: 'severity',
      key: 'severity',
      width: 130,
      render: (severity: string) => (
        <Tag color={SEVERITY_COLORS[severity] || 'default'}>
          {SEVERITY_LABELS[severity] || severity}
        </Tag>
      ),
    },
    {
      title: 'Действие',
      dataIndex: 'action',
      key: 'action',
      width: 180,
      render: (action: string) => ACTION_LABELS[action] || action,
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      width: 140,
      render: (status: string) => (
        <Tag color={STATUS_COLORS[status] || 'default'}>
          {STATUS_LABELS[status] || status}
        </Tag>
      ),
    },
    {
      title: 'Пользователь',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 120,
      render: (id: string) => (id ? `${id.slice(0, 8)}...` : '—'),
    },
    {
      title: 'Дата',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (date: string) => (date ? formatDateTime(date, 'DD.MM.YYYY HH:mm') : '—'),
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 220,
      render: (_: unknown, record: InternalHandlerFraudFlagResponse) => (
        <Space>
          <Button size="small" onClick={() => setSelectedFlag(record)}>
            Детали
          </Button>
          {record.status === 'pending' && (
            <>
              <Button
                size="small"
                type="primary"
                icon={<CheckCircleOutlined />}
                onClick={() => record.id && handleReview(record.id, 'reviewed')}
                loading={updateMutation.isPending}
              >
                Подтвердить
              </Button>
              <Button
                size="small"
                danger
                icon={<CloseCircleOutlined />}
                onClick={() => record.id && handleReview(record.id, 'dismissed')}
                loading={updateMutation.isPending}
              >
                Отклонить
              </Button>
            </>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Title level={3}>Антифрод</Title>

      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={8}>
          <Card size="small">
            <Statistic
              title="На рассмотрении"
              value={pendingCount}
              prefix={<ExclamationCircleOutlined />}
              valueStyle={{ color: pendingCount > 0 ? '#faad14' : '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card size="small">
            <Statistic title="Всего флагов" value={meta?.total_count ?? 0} />
          </Card>
        </Col>
      </Row>

      <Space style={{ marginBottom: 16 }} wrap>
        <Segmented
          options={STATUS_OPTIONS}
          value={statusFilter}
          onChange={(v) => {
            setStatusFilter(v as string)
            setPage(1)
          }}
        />
        <Select
          placeholder="Фильтр по правилу"
          allowClear
          style={{ width: 250 }}
          value={ruleFilter}
          onChange={(v) => {
            setRuleFilter(v)
            setPage(1)
          }}
          options={Object.entries(RULE_LABELS).map(([value, label]) => ({
            value,
            label,
          }))}
        />
      </Space>

      <Table
        dataSource={flags}
        columns={columns}
        loading={isLoading}
        rowKey="id"
        scroll={{ x: 'max-content' }}
        pagination={
          meta && meta.total_pages && meta.total_pages > 1
            ? {
                current: page,
                pageSize: 20,
                total: meta.total_count,
                onChange: setPage,
                showSizeChanger: false,
              }
            : false
        }
      />

      <Drawer
        title="Детали подозрительной операции"
        open={!!selectedFlag}
        onClose={() => setSelectedFlag(null)}
        width={480}
      >
        {selectedFlag && (
          <>
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="ID">
                {selectedFlag.id}
              </Descriptions.Item>
              <Descriptions.Item label="Пользователь">
                {selectedFlag.user_id}
              </Descriptions.Item>
              <Descriptions.Item label="Правило">
                {RULE_LABELS[selectedFlag.rule ?? ''] || selectedFlag.rule}
              </Descriptions.Item>
              <Descriptions.Item label="Серьёзность">
                <Tag color={SEVERITY_COLORS[selectedFlag.severity ?? ''] || 'default'}>
                  {SEVERITY_LABELS[selectedFlag.severity ?? ''] || selectedFlag.severity}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Действие">
                {ACTION_LABELS[selectedFlag.action ?? ''] || selectedFlag.action}
              </Descriptions.Item>
              <Descriptions.Item label="Статус">
                <Tag color={STATUS_COLORS[selectedFlag.status ?? ''] || 'default'}>
                  {STATUS_LABELS[selectedFlag.status ?? ''] || selectedFlag.status}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Создан">
                {selectedFlag.created_at
                  ? formatDateTime(selectedFlag.created_at, 'DD.MM.YYYY HH:mm:ss')
                  : '—'}
              </Descriptions.Item>
              {selectedFlag.reviewed_at && (
                <Descriptions.Item label="Рассмотрен">
                  {formatDateTime(selectedFlag.reviewed_at, 'DD.MM.YYYY HH:mm:ss')}
                </Descriptions.Item>
              )}
              {selectedFlag.reviewed_by && (
                <Descriptions.Item label="Рассмотрел">
                  {selectedFlag.reviewed_by}
                </Descriptions.Item>
              )}
            </Descriptions>

            {selectedFlag.details && (
              <div style={{ marginTop: 16 }}>
                <Text strong>Детали:</Text>
                <pre
                  style={{
                    background: '#f5f5f5',
                    padding: 12,
                    borderRadius: 8,
                    fontSize: 13,
                    marginTop: 8,
                    overflow: 'auto',
                    maxHeight: 300,
                  }}
                >
                  {JSON.stringify(selectedFlag.details, null, 2)}
                </pre>
              </div>
            )}

            {selectedFlag.status === 'pending' && (
              <Space style={{ marginTop: 24 }}>
                <Button
                  type="primary"
                  icon={<CheckCircleOutlined />}
                  onClick={() =>
                    selectedFlag.id && handleReview(selectedFlag.id, 'reviewed')
                  }
                  loading={updateMutation.isPending}
                >
                  Подтвердить
                </Button>
                <Button
                  danger
                  icon={<CloseCircleOutlined />}
                  onClick={() =>
                    selectedFlag.id && handleReview(selectedFlag.id, 'dismissed')
                  }
                  loading={updateMutation.isPending}
                >
                  Отклонить
                </Button>
              </Space>
            )}
          </>
        )}
      </Drawer>
    </div>
  )
}
