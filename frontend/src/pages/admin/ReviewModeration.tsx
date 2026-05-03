import { useState } from 'react'
import {
  App,
  Badge,
  Button,
  Card,
  Checkbox,
  Descriptions,
  Drawer,
  Image,
  Input,
  Modal,
  Rate,
  Segmented,
  Space,
  Table,
  Tag,
  Typography,
} from '@/components/design/system'
import type { ColumnsType } from '@/components/design/types'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetAdminReviews,
  getGetAdminReviewsQueryKey,
  useGetAdminReviewsPendingCount,
  getGetAdminReviewsPendingCountQueryKey,
  usePatchAdminReviewsIdApprove,
  usePatchAdminReviewsIdReject,
  usePostAdminReviewsBatchApprove,
  usePostAdminReviewsBatchReject,
} from '@/api/generated/admin-reviews/admin-reviews'
import type { InternalHandlerAdminReviewResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography
const { Search } = Input

const STATUS_LABELS: Record<string, { color: string; text: string }> = {
  pending: { color: 'orange', text: 'На рассмотрении' },
  approved: { color: 'green', text: 'Одобрен' },
  rejected: { color: 'red', text: 'Отклонён' },
  hidden: { color: 'default', text: 'Скрыт' },
}

const STATUS_OPTIONS = [
  { label: 'Все', value: '' },
  { label: 'На рассмотрении', value: 'pending' },
  { label: 'Одобренные', value: 'approved' },
  { label: 'Отклонённые', value: 'rejected' },
  { label: 'Скрытые', value: 'hidden' },
]

export default function ReviewModeration() {
  const { modal, message } = App.useApp()
  const queryClient = useQueryClient()

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [statusFilter, setStatusFilter] = useState('')
  const [search, setSearch] = useState('')
  const [detailItem, setDetailItem] = useState<InternalHandlerAdminReviewResponse | null>(null)
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [rejectModalOpen, setRejectModalOpen] = useState(false)
  const [rejectReason, setRejectReason] = useState('')
  const [rejectTarget, setRejectTarget] = useState<'single' | 'batch'>('single')
  const [singleRejectId, setSingleRejectId] = useState<string>('')

  const { data, isLoading } = useGetAdminReviews({
    page,
    page_size: pageSize,
    ...(statusFilter ? { status: statusFilter } : {}),
  })
  const { data: pendingData } = useGetAdminReviewsPendingCount()
  const approveMutation = usePatchAdminReviewsIdApprove()
  const rejectMutation = usePatchAdminReviewsIdReject()
  const batchApproveMutation = usePostAdminReviewsBatchApprove()
  const batchRejectMutation = usePostAdminReviewsBatchReject()

  const reviews = data?.data ?? []
  const meta = data?.meta
  const pendingCount = pendingData?.data?.pending_count ?? 0

  const filteredReviews = search
    ? reviews.filter((r) => {
        const q = search.toLowerCase()
        return (
          r.text?.toLowerCase().includes(q) ||
          r.user_id?.toLowerCase().includes(q) ||
          r.bathhouse_id?.toLowerCase().includes(q)
        )
      })
    : reviews

  const invalidateAll = () => {
    queryClient.invalidateQueries({ queryKey: getGetAdminReviewsQueryKey() })
    queryClient.invalidateQueries({ queryKey: getGetAdminReviewsPendingCountQueryKey() })
  }

  const handleApprove = (item: InternalHandlerAdminReviewResponse) => {
    modal.confirm({
      title: 'Одобрить отзыв?',
      content: `Отзыв с рейтингом ${item.rating ?? 0} будет опубликован.`,
      okText: 'Одобрить',
      cancelText: 'Отмена',
      onOk: () =>
        approveMutation.mutateAsync({ id: item.id! }).then(() => {
          message.success('Отзыв одобрен')
          invalidateAll()
        }).catch(() => {
          message.error('Не удалось одобрить отзыв')
        }),
    })
  }

  const openRejectModal = (target: 'single' | 'batch', id?: string) => {
    setRejectTarget(target)
    setSingleRejectId(id ?? '')
    setRejectReason('')
    setRejectModalOpen(true)
  }

  const handleRejectConfirm = async () => {
    try {
      if (rejectTarget === 'single' && singleRejectId) {
        await rejectMutation.mutateAsync({
          id: singleRejectId,
          data: { reason: rejectReason || undefined },
        })
        message.success('Отзыв отклонён')
      } else if (rejectTarget === 'batch' && selectedIds.length > 0) {
        await batchRejectMutation.mutateAsync({
          data: { ids: selectedIds, reason: rejectReason || undefined },
        })
        message.success(`Отклонено отзывов: ${selectedIds.length}`)
        setSelectedIds([])
      }
      setRejectModalOpen(false)
      invalidateAll()
    } catch {
      message.error('Не удалось отклонить отзыв')
    }
  }

  const handleBatchApprove = () => {
    if (selectedIds.length === 0) return
    modal.confirm({
      title: `Одобрить ${selectedIds.length} отзывов?`,
      content: 'Все выбранные отзывы будут опубликованы.',
      okText: 'Одобрить',
      cancelText: 'Отмена',
      onOk: () =>
        batchApproveMutation.mutateAsync({ data: { ids: selectedIds } }).then(() => {
          message.success(`Одобрено отзывов: ${selectedIds.length}`)
          setSelectedIds([])
          invalidateAll()
        }).catch(() => {
          message.error('Не удалось одобрить отзывы')
        }),
    })
  }

  const toggleSelect = (id: string) => {
    setSelectedIds((prev) =>
      prev.includes(id) ? prev.filter((i) => i !== id) : [...prev, id],
    )
  }

  const toggleSelectAll = () => {
    const allIds = filteredReviews.map((r) => r.id!).filter(Boolean)
    if (selectedIds.length === allIds.length) {
      setSelectedIds([])
    } else {
      setSelectedIds(allIds)
    }
  }

  const columns: ColumnsType<InternalHandlerAdminReviewResponse> = [
    {
      title: (
        <Checkbox
          checked={selectedIds.length > 0 && selectedIds.length === filteredReviews.length}
          indeterminate={selectedIds.length > 0 && selectedIds.length < filteredReviews.length}
          onChange={toggleSelectAll}
        />
      ),
      key: 'select',
      width: 48,
      render: (_, record) => (
        <Checkbox
          checked={selectedIds.includes(record.id!)}
          onChange={() => toggleSelect(record.id!)}
        />
      ),
    },
    {
      title: 'Рейтинг',
      dataIndex: 'rating',
      key: 'rating',
      width: 160,
      render: (rating: number) => <Rate className="rh-admin-rating-small" disabled value={rating} />,
    },
    {
      title: 'Текст',
      dataIndex: 'text',
      key: 'text',
      ellipsis: true,
      render: (text: string, record) => (
        <a onClick={() => setDetailItem(record)}>{text || '—'}</a>
      ),
    },
    {
      title: 'Медиа',
      key: 'images',
      width: 80,
      responsive: ['lg'] as const,
      render: (_, record) => {
        const count = record.images?.length ?? 0
        return count > 0 ? <Tag>{count} фото</Tag> : '—'
      },
    },
    {
      title: 'Статус',
      key: 'status',
      dataIndex: 'status',
      width: 140,
      render: (status: string) => {
        const config = STATUS_LABELS[status] ?? { color: 'default', text: status }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: 'Дата',
      key: 'created_at',
      dataIndex: 'created_at',
      width: 140,
      responsive: ['md'] as const,
      render: (date: string) => (date ? formatDateTime(date) : '—'),
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 180,
      render: (_, record) => {
        const actions: React.ReactNode[] = []
        if (record.status !== 'approved') {
          actions.push(
            <Button key="approve" type="link" size="small" onClick={() => handleApprove(record)}>
              Одобрить
            </Button>,
          )
        }
        if (record.status !== 'rejected') {
          actions.push(
            <Button
              key="reject"
              type="link"
              size="small"
              danger
              onClick={() => openRejectModal('single', record.id)}
            >
              Отклонить
            </Button>,
          )
        }
        return <Space>{actions}</Space>
      },
    },
  ]

  const pendingLabel = (
    <Space>
      На рассмотрении
      {pendingCount > 0 && <Badge count={pendingCount} size="small" />}
    </Space>
  )

  const statusOptions = STATUS_OPTIONS.map((opt) =>
    opt.value === 'pending' ? { ...opt, label: pendingLabel } : opt,
  )

  return (
    <div className="rh-stack rh-admin-reference-page">
      <PageHeader
        eyebrow="Модерация"
        title="Модерация отзывов"
        description="Очередь отзывов, batch-действия и карточка деталей с медиа без старой разрозненной верстки."
      />

      <Card className="rh-admin-filter-card" title="Фильтры">
        <Segmented
          options={statusOptions}
          value={statusFilter}
          onChange={(val) => {
            setStatusFilter(val as string)
            setPage(1)
            setSelectedIds([])
          }}
        />
      </Card>

      <Card className="rh-admin-reference-card" title="Список отзывов">
        <div className="rh-admin-table-tools">
          <Search
            className="rh-admin-search-control"
            placeholder="Поиск по тексту или ID"
            allowClear
            onSearch={setSearch}
            onChange={(e) => !e.target.value && setSearch('')}
          />
          {selectedIds.length > 0 && (
            <Space className="rh-admin-selection-bar" wrap>
              <Text type="secondary">Выбрано: {selectedIds.length}</Text>
              <Button type="primary" size="small" onClick={handleBatchApprove}>
                Одобрить выбранные
              </Button>
              <Button
                danger
                size="small"
                onClick={() => openRejectModal('batch')}
              >
                Отклонить выбранные
              </Button>
            </Space>
          )}
        </div>

        <Table
          columns={columns}
          dataSource={filteredReviews}
          rowKey="id"
          loading={isLoading}
          locale={{ emptyText: 'Нет отзывов' }}
          pagination={{
            current: page,
            pageSize: pageSize,
            total: meta?.total_count ?? 0,
            showSizeChanger: true,
            showTotal: (total) => `Всего: ${total}`,
            onChange: (p, ps) => {
              setPage(p)
              setPageSize(ps)
              setSelectedIds([])
            },
          }}
        />
      </Card>

      <Drawer
        title="Детали отзыва"
        open={!!detailItem}
        onClose={() => setDetailItem(null)}
        size="large"
      >
        {detailItem && (
          <>
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="Рейтинг">
                <Rate disabled value={detailItem.rating ?? 0} />
              </Descriptions.Item>
              <Descriptions.Item label="Текст">
                {detailItem.text ?? '—'}
              </Descriptions.Item>
              <Descriptions.Item label="Статус">
                {(() => {
                  const config = STATUS_LABELS[detailItem.status ?? ''] ?? {
                    color: 'default',
                    text: detailItem.status,
                  }
                  return <Tag color={config.color}>{config.text}</Tag>
                })()}
              </Descriptions.Item>
              <Descriptions.Item label="Пользователь">
                <Text copyable>{detailItem.user_id ?? '—'}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Баня">
                <Text copyable>{detailItem.bathhouse_id ?? '—'}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Бронирование">
                {detailItem.booking_id ? (
                  <Text copyable>{detailItem.booking_id}</Text>
                ) : '—'}
              </Descriptions.Item>
              <Descriptions.Item label="Ответ владельца">
                {detailItem.owner_response ?? '—'}
              </Descriptions.Item>
              {detailItem.owner_response_at && (
                <Descriptions.Item label="Дата ответа">
                  {formatDateTime(detailItem.owner_response_at)}
                </Descriptions.Item>
              )}
              {detailItem.rejection_reasons && detailItem.rejection_reasons.length > 0 && (
                <Descriptions.Item label="Причины отклонения">
                  {detailItem.rejection_reasons.join(', ')}
                </Descriptions.Item>
              )}
              <Descriptions.Item label="Создан">
                {detailItem.created_at ? formatDateTime(detailItem.created_at) : '—'}
              </Descriptions.Item>
            </Descriptions>

            {detailItem.images && detailItem.images.length > 0 && (
              <div className="rh-admin-media-block">
                <h3 className="rh-admin-subsection-title">Медиа ({detailItem.images.length})</h3>
                <Image.PreviewGroup>
                  <Space wrap>
                    {detailItem.images.map((url, idx) => (
                      <Image
                        className="rh-admin-review-media"
                        key={idx}
                        src={url}
                        width={120}
                        height={120}
                        fallback="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTIwIiBoZWlnaHQ9IjEyMCIgdmlld0JveD0iMCAwIDEyMCAxMjAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHJlY3Qgd2lkdGg9IjEyMCIgaGVpZ2h0PSIxMjAiIGZpbGw9IiNmMGYwZjAiLz48dGV4dCB4PSI2MCIgeT0iNjAiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGR5PSIuM2VtIiBmaWxsPSIjOTk5IiBmb250LXNpemU9IjEyIj5ObyBJbWFnZTwvdGV4dD48L3N2Zz4="
                      />
                    ))}
                  </Space>
                </Image.PreviewGroup>
              </div>
            )}
          </>
        )}
      </Drawer>

      <Modal
        title="Отклонить отзыв"
        open={rejectModalOpen}
        onCancel={() => setRejectModalOpen(false)}
        onOk={handleRejectConfirm}
        okText="Отклонить"
        okType="danger"
        cancelText="Отмена"
        confirmLoading={rejectMutation.isPending || batchRejectMutation.isPending}
      >
        <div className="rh-admin-modal-description">
          <Text>
            {rejectTarget === 'batch'
              ? `Отклонить ${selectedIds.length} отзывов?`
              : 'Укажите причину отклонения (необязательно):'}
          </Text>
        </div>
        <Input.TextArea
          rows={3}
          placeholder="Причина отклонения"
          value={rejectReason}
          onChange={(e) => setRejectReason(e.target.value)}
        />
      </Modal>
    </div>
  )
}
