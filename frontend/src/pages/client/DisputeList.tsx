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
  useGetMyDisputes,
} from '@/api/generated/disputes/disputes'
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

export default function DisputeList() {
  const navigate = useNavigate()

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [statusFilter, setStatusFilter] = useState('')

  const { data, isLoading } = useGetMyDisputes({ page, page_size: pageSize })

  const allDisputes: InternalHandlerDisputeResponse[] = data?.data ?? []
  const disputes = statusFilter
    ? allDisputes.filter((d) => d.status === statusFilter)
    : allDisputes
  const meta = data?.meta

  const columns: ColumnsType<InternalHandlerDisputeResponse> = [
    {
      title: 'Причина',
      dataIndex: 'reason',
      key: 'reason',
      width: 200,
      render: (reason: string, record: InternalHandlerDisputeResponse) => (
        <a onClick={() => navigate(`/client/disputes/${record.id}`)}>
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
      width: 160,
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
      title: 'Дата',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (date: string) => (date ? formatDateTime(date) : '-'),
    },
  ]

  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="Поддержка"
        title="Мои споры"
        description="Отслеживайте открытые споры, решения и возвраты."
      />

      <Segmented
        options={STATUS_OPTIONS}
        value={statusFilter}
        onChange={(val) => {
          setStatusFilter(val as string)
          setPage(1)
        }}
        className="rh-client-segmented-filter"
      />

      {isLoading ? (
        <Card className="rh-admin-state-card">
          <Spin size="large" />
        </Card>
      ) : disputes.length === 0 ? (
        <Card>
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Нет споров</div>
            <p className="rh-admin-empty-state__text">
              Спор можно создать из деталей завершённого или проблемного бронирования.
            </p>
          </div>
        </Card>
      ) : (
        <>
          <Table
            dataSource={disputes}
            columns={columns}
            rowKey="id"
            pagination={false}
            size="middle"
            locale={{
              emptyText: (
                <div className="rh-admin-empty-state">
                  <div className="rh-admin-empty-state__title">Нет споров с выбранным статусом</div>
                </div>
              ),
            }}
            onRow={(record) => ({
              onClick: () => navigate(`/client/disputes/${record.id}`),
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
    </div>
  )
}
