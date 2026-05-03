import { useState } from 'react'
import { DatePicker, Table, Tag, Typography, Card, Space } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { HistoryOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { useParams } from 'react-router-dom'
import { useGetMyBathhousesIdHistory } from '@/api/generated/bathhouses/bathhouses'
import type { InternalHandlerAuditLogResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'
import EmptyState from '@/components/EmptyState'

const { Title } = Typography
const { RangePicker } = DatePicker

const ACTION_CONFIG: Record<string, { color: string; text: string }> = {
  create: { color: 'green', text: 'Создание' },
  update: { color: 'blue', text: 'Изменение' },
  delete: { color: 'red', text: 'Удаление' },
}

const FIELD_LABELS: Record<string, string> = {
  name: 'Название',
  description: 'Описание',
  address: 'Адрес',
  city_id: 'Город',
  base_price: 'Базовая цена',
  capacity: 'Вместимость',
  status: 'Статус',
  is_active: 'Активность',
  photos: 'Фото',
  amenities: 'Удобства',
  rules: 'Правила',
  working_hours: 'Часы работы',
  cancellation_policy: 'Политика отмены',
  latitude: 'Широта',
  longitude: 'Долгота',
  slug: 'URL-идентификатор',
  type_id: 'Тип объекта',
  buffer_time: 'Буфер между бронями',
  lead_time: 'Минимальное время до брони',
  max_advance_days: 'Макс. дней бронирования вперёд',
  booking_mode: 'Режим бронирования',
  deposit_percent: 'Процент залога',
}

function renderChangedFields(fields: unknown): React.ReactNode {
  if (!fields) return '—'

  if (typeof fields === 'object' && fields !== null && !Array.isArray(fields)) {
    const entries = Object.entries(fields as Record<string, unknown>)
    if (entries.length === 0) return '—'

    return (
      <div style={{ maxWidth: 400 }}>
        {entries.map(([key, value]) => {
          const label = FIELD_LABELS[key] || key
          const change = value as { old?: unknown; new?: unknown } | unknown

          if (typeof change === 'object' && change !== null && 'old' in (change as Record<string, unknown>)) {
            const typedChange = change as { old?: unknown; new?: unknown }
            return (
              <div key={key} style={{ marginBottom: 4, fontSize: 13 }}>
                <strong>{label}:</strong>{' '}
                <span style={{ color: '#b42318', textDecoration: 'line-through' }}>
                  {formatValue(typedChange.old)}
                </span>
                {' → '}
                <span style={{ color: '#15803d' }}>{formatValue(typedChange.new)}</span>
              </div>
            )
          }

          return (
            <div key={key} style={{ marginBottom: 4, fontSize: 13 }}>
              <strong>{label}:</strong> {formatValue(change)}
            </div>
          )
        })}
      </div>
    )
  }

  if (Array.isArray(fields)) {
    return fields.join(', ')
  }

  return String(fields)
}

function formatValue(val: unknown): string {
  if (val === null || val === undefined) return '—'
  if (typeof val === 'boolean') return val ? 'Да' : 'Нет'
  if (typeof val === 'object') return JSON.stringify(val)
  return String(val)
}

export default function AuditLog() {
  const { id: bathhouseId } = useParams<{ id: string }>()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs | null, dayjs.Dayjs | null] | null>(null)

  const { data, isLoading } = useGetMyBathhousesIdHistory(bathhouseId ?? '', {
    page,
    page_size: pageSize,
  }, {
    query: {
      enabled: !!bathhouseId,
    },
  })

  const entries = data?.data ?? []
  const meta = data?.meta

  const filteredEntries = entries.filter((entry) => {
    if (dateRange?.[0] && dateRange?.[1] && entry.created_at) {
      const entryDate = dayjs(entry.created_at)
      if (entryDate.isBefore(dateRange[0], 'day') || entryDate.isAfter(dateRange[1], 'day')) {
        return false
      }
    }
    return true
  })

  const columns: ColumnsType<InternalHandlerAuditLogResponse> = [
    {
      title: 'Дата',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (date: string) => date ? formatDateTime(date) : '—',
      sorter: (a, b) => (a.created_at ?? '').localeCompare(b.created_at ?? ''),
      defaultSortOrder: 'descend',
    },
    {
      title: 'Действие',
      dataIndex: 'action',
      key: 'action',
      width: 120,
      render: (action: string) => {
        const config = ACTION_CONFIG[action] ?? { color: 'default', text: action }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: 'Пользователь',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 140,
      render: (userId: string) => userId ? userId.slice(0, 8) + '...' : '—',
      responsive: ['md'],
    },
    {
      title: 'Изменения',
      dataIndex: 'changed_fields',
      key: 'changed_fields',
      render: (fields: unknown) => renderChangedFields(fields),
    },
  ]

  return (
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>
        <HistoryOutlined style={{ marginRight: 8 }} />
        История изменений
      </Title>

      <Card style={{ marginBottom: 16 }}>
        <Space wrap>
          <span>Фильтр по дате:</span>
          <RangePicker
            value={dateRange}
            onChange={(dates) => setDateRange(dates)}
            format="DD.MM.YYYY"
            placeholder={['С', 'По']}
          />
        </Space>
      </Card>

      <Table
        columns={columns}
        dataSource={filteredEntries}
        rowKey="id"
        loading={isLoading}
        locale={{ emptyText: <EmptyState description="История изменений пуста" /> }}
        pagination={{
          current: page,
          pageSize: pageSize,
          total: meta?.total_count ?? 0,
          showSizeChanger: true,
          showTotal: (total) => `Всего: ${total}`,
          onChange: (p, ps) => {
            setPage(p)
            setPageSize(ps)
          },
        }}
      />
    </div>
  )
}
