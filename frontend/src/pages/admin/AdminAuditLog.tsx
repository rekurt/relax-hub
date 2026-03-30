import { useState } from 'react'
import { Typography, Table, Tag, Space, Input, Select, DatePicker, Card, Segmented } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useGetAdminAuditLog, useGetAdminAuditLogActions } from '@/api/generated/admin-audit/admin-audit'
import type { InternalHandlerAuditLogResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'
import dayjs from 'dayjs'

const { Title } = Typography
const { RangePicker } = DatePicker

function actionTag(action?: string) {
  switch (action) {
    case 'create': return <Tag color="green">Создание</Tag>
    case 'update': return <Tag color="blue">Обновление</Tag>
    case 'delete': return <Tag color="red">Удаление</Tag>
    default: return <Tag>{action ?? '—'}</Tag>
  }
}

const ENTITY_TYPES = [
  { label: 'Баня', value: 'bathhouse' },
  { label: 'Бронирование', value: 'booking' },
  { label: 'Пользователь', value: 'user' },
  { label: 'Отзыв', value: 'review' },
  { label: 'Платёж', value: 'payment' },
  { label: 'Кошелёк', value: 'wallet' },
  { label: 'Тикет', value: 'ticket' },
  { label: 'Спор', value: 'dispute' },
]

const ACTION_TYPES = [
  { label: 'Создание', value: 'create' },
  { label: 'Обновление', value: 'update' },
  { label: 'Удаление', value: 'delete' },
]

type TabKey = 'entities' | 'actions'

export default function AdminAuditLog() {
  const [tab, setTab] = useState<TabKey>('actions')
  const [page, setPage] = useState(1)
  const [entityType, setEntityType] = useState<string | undefined>()
  const [entityId, setEntityId] = useState<string | undefined>()
  const [userId, setUserId] = useState<string | undefined>()
  const [action, setAction] = useState<string | undefined>()
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null)

  const entityLogQuery = useGetAdminAuditLog(
    tab === 'entities' ? {
      page,
      page_size: 20,
      entity_type: entityType,
      entity_id: entityId || undefined,
      user_id: userId || undefined,
      action,
      from_date: dateRange?.[0]?.toISOString(),
      to_date: dateRange?.[1]?.toISOString(),
    } : undefined,
    { query: { enabled: tab === 'entities' } },
  )

  const adminLogQuery = useGetAdminAuditLogActions(
    tab === 'actions' ? {
      page,
      page_size: 20,
      admin_id: userId || undefined,
      action,
      from_date: dateRange?.[0]?.toISOString(),
      to_date: dateRange?.[1]?.toISOString(),
    } : undefined,
    { query: { enabled: tab === 'actions' } },
  )

  const activeQuery = tab === 'entities' ? entityLogQuery : adminLogQuery
  const entries = activeQuery.data?.data ?? []
  const meta = activeQuery.data?.meta

  const resetFilters = () => {
    setPage(1)
    setEntityType(undefined)
    setEntityId(undefined)
    setUserId(undefined)
    setAction(undefined)
    setDateRange(null)
  }

  const columns: ColumnsType<InternalHandlerAuditLogResponse> = [
    {
      title: 'Дата',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (v: string) => v ? formatDateTime(v) : '—',
      width: 160,
    },
    ...(tab === 'entities' ? [{
      title: 'Тип',
      dataIndex: 'entity_type',
      key: 'entity_type',
      width: 140,
    } as const] : []),
    ...(tab === 'entities' ? [{
      title: 'ID сущности',
      dataIndex: 'entity_id',
      key: 'entity_id',
      ellipsis: true as const,
      width: 200,
      render: (v: string) => v?.slice(0, 8) + '...' || '—',
    } as const] : []),
    {
      title: 'Действие',
      dataIndex: 'action',
      key: 'action',
      render: actionTag,
      width: 130,
    },
    {
      title: tab === 'actions' ? 'Администратор' : 'Пользователь',
      dataIndex: tab === 'actions' ? 'admin_id' : 'user_id',
      key: 'user',
      ellipsis: true,
      width: 200,
      render: (v: string) => v?.slice(0, 8) + '...' || '—',
    },
    ...(tab === 'entities' ? [{
      title: 'Изм. поля',
      dataIndex: 'changed_fields',
      key: 'changed_fields',
      render: (v: number[]) => v?.length ? `${v.length} полей` : '—',
      width: 100,
    } as const] : []),
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={3} style={{ margin: 0 }}>Журнал аудита</Title>
        <Segmented
          options={[
            { label: 'Действия админов', value: 'actions' },
            { label: 'Изменения сущностей', value: 'entities' },
          ]}
          value={tab}
          onChange={(v) => { setTab(v as TabKey); resetFilters() }}
        />
      </div>

      <Card style={{ marginBottom: 16 }}>
        <Space wrap>
          {tab === 'entities' && (
            <>
              <Select
                placeholder="Тип сущности"
                allowClear
                style={{ width: 180 }}
                value={entityType}
                onChange={(v) => { setEntityType(v); setPage(1) }}
                options={ENTITY_TYPES}
              />
              <Input
                placeholder="ID сущности"
                allowClear
                style={{ width: 240 }}
                value={entityId}
                onChange={(e) => { setEntityId(e.target.value); setPage(1) }}
              />
            </>
          )}
          <Input
            placeholder={tab === 'actions' ? 'ID администратора' : 'ID пользователя'}
            allowClear
            style={{ width: 240 }}
            value={userId}
            onChange={(e) => { setUserId(e.target.value); setPage(1) }}
          />
          <Select
            placeholder="Действие"
            allowClear
            style={{ width: 160 }}
            value={action}
            onChange={(v) => { setAction(v); setPage(1) }}
            options={ACTION_TYPES}
          />
          <RangePicker
            value={dateRange}
            onChange={(dates) => { setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null); setPage(1) }}
          />
        </Space>
      </Card>

      <Table
        columns={columns}
        dataSource={entries}
        rowKey="id"
        loading={activeQuery.isLoading}
        pagination={{
          current: page,
          pageSize: 20,
          total: meta?.total_count,
          onChange: setPage,
        }}
        locale={{ emptyText: 'Нет записей' }}
      />
    </div>
  )
}
