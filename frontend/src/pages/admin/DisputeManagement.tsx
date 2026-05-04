import { useState } from 'react'
import {
  Card,
  Pagination,
  Segmented,
  Spin,
  Table,
  Tag,
  Typography,
} from '@/components/design/system'
import type { ColumnsType } from '@/components/design/types'
import { useNavigate } from 'react-router-dom'
import {
  useGetAdminDisputes,
} from '@/api/generated/disputes-admin/disputes-admin'
import type { InternalHandlerDisputeResponse } from '@/api/generated/model'
import { formatDateTime, formatPrice } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

const STATUS_OPTIONS = [
  { label: 'Все', value: '' },
  { label: 'Открытые', value: 'open' },
  { label: 'Сбор доказательств', value: 'evidence_collection' },
  { label: 'На рассмотрении', value: 'under_review' },
  { label: 'Решённые', value: 'resolved' },
  { label: 'Апелляция', value: 'appealed' },
  { label: 'Закрытые', value: 'closed' },
]

const statusLabel: Record<string, string> = {
  open: 'Открыт',
  evidence_collection: 'Сбор доказательств',
  under_review: 'На рассмотрении',
  resolved: 'Решён',
  appealed: 'Апелляция',
  closed: 'Закрыт',
}

const statusColor: Record<string, string> = {
  open: 'blue',
  evidence_collection: 'processing',
  under_review: 'orange',
  resolved: 'green',
  appealed: 'volcano',
  closed: 'default',
}

const reasonLabel: Record<string, string> = {
  service_not_provided: 'Услуга не оказана',
  poor_quality: 'Низкое качество',
  damage: 'Повреждение имущества',
  safety_issue: 'Проблема безопасности',
  billing_error: 'Ошибка в счёте',
  other: 'Другое',
}

const resolutionLabel: Record<string, string> = {
  full_refund: 'Полный возврат',
  partial_refund: 'Частичный возврат',
  no_refund: 'Без возврата',
}

export default function DisputeManagement() {
  const navigate = useNavigate()

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [statusFilter, setStatusFilter] = useState('')

  const { data, isLoading } = useGetAdminDisputes({
    page,
    page_size: pageSize,
    ...(statusFilter && { status: statusFilter }),
  })

  const disputes: InternalHandlerDisputeResponse[] = data?.data ?? []
  const meta = data?.meta

  const columns: ColumnsType<InternalHandlerDisputeResponse> = [
    {
      title: 'Причина',
      dataIndex: 'reason',
      key: 'reason',
      width: 180,
      render: (reason: string, record: InternalHandlerDisputeResponse) => (
        <a onClick={() => navigate(`/admin/disputes/${record.id}`)}>
          {reasonLabel[reason] ?? reason}
        </a>
      ),
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      width: 160,
      render: (status: string) => (
        <Tag color={statusColor[status] ?? 'default'}>
          {statusLabel[status] ?? status}
        </Tag>
      ),
    },
    {
      title: 'Решение',
      dataIndex: 'resolution',
      key: 'resolution',
      width: 150,
      render: (resolution: string) =>
        resolution ? (
          <Tag>{resolutionLabel[resolution] ?? resolution}</Tag>
        ) : (
          <Text type="secondary">-</Text>
        ),
    },
    {
      title: 'Возврат',
      dataIndex: 'refund_amount',
      key: 'refund_amount',
      width: 120,
      render: (amount: number) =>
        amount ? formatPrice(amount) : <Text type="secondary">-</Text>,
    },
    {
      title: 'Инициатор',
      dataIndex: 'initiator_id',
      key: 'initiator_id',
      width: 120,
      render: (uid: string) =>
        uid ? (
          <Text copyable={{ text: uid }}>{uid.slice(0, 8)}...</Text>
        ) : (
          '-'
        ),
    },
    {
      title: 'Медиатор',
      dataIndex: 'mediator_id',
      key: 'mediator_id',
      width: 120,
      render: (uid: string) =>
        uid ? (
          <Text copyable={{ text: uid }}>{uid.slice(0, 8)}...</Text>
        ) : (
          <Text type="secondary">Не назначен</Text>
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
        eyebrow="Модерация"
        title="Управление спорами"
        description="Фильтрация, медиаторы и статусы возвратов в едином операционном списке."
        extra={
          <Segmented
            options={STATUS_OPTIONS}
            value={statusFilter}
            onChange={(val) => {
              setStatusFilter(val as string)
              setPage(1)
            }}
          />
        }
      />

      <Card className="rh-admin-reference-card" title="Список споров">
        {isLoading ? (
          <div className="rh-admin-state-card">
            <Spin size="large" />
            <span>Загружаем споры</span>
          </div>
        ) : disputes.length === 0 ? (
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Нет споров</div>
            <p className="rh-admin-empty-state__text">
              Новые обращения появятся здесь, когда клиент или владелец откроет спор по бронированию.
            </p>
          </div>
        ) : (
          <>
            <Table
              dataSource={disputes}
              columns={columns}
              rowKey="id"
              pagination={false}
              size="middle"
              locale={{ emptyText: 'Нет споров' }}
              onRow={(record) => ({
                onClick: () => navigate(`/admin/disputes/${record.id}`),
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
