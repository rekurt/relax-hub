import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Descriptions,
  Drawer,
  Input,
  Modal,
  Pagination,
  Segmented,
  Space,
  Spin,
  Table,
  Tag,
  Typography,
} from '@/components/design/system'
import {
  CheckOutlined,
  CloseOutlined,
  EyeOutlined,
} from '@/components/design/icons'
import type { ColumnsType } from '@/components/design/types'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetAdminComplaints,
  getGetAdminComplaintsQueryKey,
  usePatchAdminComplaintsIdResolve,
  usePatchAdminComplaintsIdDismiss,
} from '@/api/generated/admin-complaints/admin-complaints'
import type { InternalHandlerComplaintResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

const STATUS_OPTIONS = [
  { label: 'Все', value: '' },
  { label: 'На рассмотрении', value: 'pending' },
  { label: 'Решённые', value: 'resolved' },
  { label: 'Отклонённые', value: 'dismissed' },
]

const TARGET_TYPE_OPTIONS = [
  { label: 'Все типы', value: '' },
  { label: 'Отзыв', value: 'review' },
  { label: 'Баня', value: 'bathhouse' },
  { label: 'Пользователь', value: 'user' },
]

const REASON_OPTIONS = [
  { label: 'Все причины', value: '' },
  { label: 'Спам', value: 'spam' },
  { label: 'Оскорбление', value: 'offensive' },
  { label: 'Фейк', value: 'fake' },
  { label: 'Мошенничество', value: 'fraud' },
  { label: 'Другое', value: 'other' },
]

const statusLabel: Record<string, string> = {
  pending: 'На рассмотрении',
  resolved: 'Решена',
  dismissed: 'Отклонена',
}

const statusColor: Record<string, string> = {
  pending: 'orange',
  resolved: 'green',
  dismissed: 'default',
}

const targetTypeLabel: Record<string, string> = {
  review: 'Отзыв',
  bathhouse: 'Баня',
  user: 'Пользователь',
}

const reasonLabel: Record<string, string> = {
  spam: 'Спам',
  offensive: 'Оскорбление',
  fake: 'Фейк',
  fraud: 'Мошенничество',
  other: 'Другое',
}

export default function ComplaintManagement() {
  const { modal, message } = App.useApp()
  const queryClient = useQueryClient()

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [statusFilter, setStatusFilter] = useState('')
  const [targetTypeFilter, setTargetTypeFilter] = useState('')
  const [reasonFilter, setReasonFilter] = useState('')
  const [selectedComplaint, setSelectedComplaint] = useState<InternalHandlerComplaintResponse | null>(null)
  const [resolveModalOpen, setResolveModalOpen] = useState(false)
  const [resolveNote, setResolveNote] = useState('')
  const [resolveComplaintId, setResolveComplaintId] = useState<string>('')

  const { data, isLoading } = useGetAdminComplaints({
    page,
    page_size: pageSize,
    ...(statusFilter && { status: statusFilter }),
    ...(targetTypeFilter && { target_type: targetTypeFilter }),
    ...(reasonFilter && { reason: reasonFilter }),
  })

  const resolveMutation = usePatchAdminComplaintsIdResolve()
  const dismissMutation = usePatchAdminComplaintsIdDismiss()

  const complaints: InternalHandlerComplaintResponse[] = data?.data ?? []
  const meta = data?.meta

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetAdminComplaintsQueryKey() })
  }

  const openResolveModal = (id: string) => {
    setResolveComplaintId(id)
    setResolveNote('')
    setResolveModalOpen(true)
  }

  const handleResolveConfirm = async () => {
    if (!resolveComplaintId) return
    try {
      await resolveMutation.mutateAsync({
        id: resolveComplaintId,
        data: { resolution: resolveNote || undefined },
      })
      message.success('Жалоба решена')
      setResolveModalOpen(false)
      setSelectedComplaint(null)
      invalidate()
    } catch {
      message.error('Не удалось решить жалобу')
    }
  }

  const handleDismiss = (complaint: InternalHandlerComplaintResponse) => {
    modal.confirm({
      title: 'Отклонить жалобу?',
      content: 'Жалоба будет отклонена без действий.',
      okText: 'Отклонить',
      okType: 'danger',
      cancelText: 'Отмена',
      onOk: () =>
        dismissMutation.mutateAsync({ id: complaint.id! }).then(() => {
          message.success('Жалоба отклонена')
          setSelectedComplaint(null)
          invalidate()
        }).catch(() => {
          message.error('Не удалось отклонить жалобу')
        }),
    })
  }

  const columns: ColumnsType<InternalHandlerComplaintResponse> = [
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      width: 140,
      render: (status: string) => (
        <Tag color={statusColor[status] ?? 'default'}>
          {statusLabel[status] ?? status}
        </Tag>
      ),
    },
    {
      title: 'Тип цели',
      dataIndex: 'target_type',
      key: 'target_type',
      width: 130,
      render: (type: string) => targetTypeLabel[type] ?? type,
    },
    {
      title: 'Причина',
      dataIndex: 'reason',
      key: 'reason',
      width: 140,
      render: (reason: string) => (
        <Tag>{reasonLabel[reason] ?? reason}</Tag>
      ),
    },
    {
      title: 'Описание',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (text: string, record: InternalHandlerComplaintResponse) => (
        <a onClick={() => setSelectedComplaint(record)}>
          {text || <Text type="secondary">Без описания</Text>}
        </a>
      ),
    },
    {
      title: 'Дата',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (date: string) => date ? formatDateTime(date) : '-',
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 200,
      render: (_: unknown, record: InternalHandlerComplaintResponse) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => setSelectedComplaint(record)}
          >
            Детали
          </Button>
          {record.status === 'pending' && (
            <>
              <Button
                type="link"
                size="small"
                icon={<CheckOutlined />}
                className="rh-admin-action-success"
                onClick={() => openResolveModal(record.id!)}
              >
                Решить
              </Button>
              <Button
                type="link"
                size="small"
                icon={<CloseOutlined />}
                danger
                onClick={() => handleDismiss(record)}
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
    <div className="rh-stack rh-admin-reference-page">
      <PageHeader
        eyebrow="Модерация"
        title="Управление жалобами"
        description="Очередь жалоб, фильтры по типам и решения модераторов в одном аккуратном списке."
      />

      <Card className="rh-admin-filter-card" title="Фильтры">
        <Space className="rh-admin-filter-column" orientation="vertical" size={16}>
          <Space wrap>
            <Segmented
              options={STATUS_OPTIONS}
              value={statusFilter}
              onChange={(val) => {
                setStatusFilter(val as string)
                setPage(1)
              }}
            />
          </Space>
          <Space wrap>
            <Segmented
              options={TARGET_TYPE_OPTIONS}
              value={targetTypeFilter}
              onChange={(val) => {
                setTargetTypeFilter(val as string)
                setPage(1)
              }}
            />
            <Segmented
              options={REASON_OPTIONS}
              value={reasonFilter}
              onChange={(val) => {
                setReasonFilter(val as string)
                setPage(1)
              }}
            />
          </Space>
        </Space>
      </Card>

      <Card className="rh-admin-reference-card" title="Список жалоб">
        {isLoading ? (
          <div className="rh-admin-state-card">
            <Spin size="large" />
            <span>Загружаем жалобы</span>
          </div>
        ) : complaints.length === 0 ? (
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Нет жалоб</div>
            <p className="rh-admin-empty-state__text">
              Новые жалобы пользователей появятся здесь после отправки формы модерации.
            </p>
          </div>
        ) : (
          <>
            <Table
              dataSource={complaints}
              columns={columns}
              rowKey="id"
              pagination={false}
              size="middle"
              locale={{ emptyText: 'Нет жалоб' }}
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

      <Drawer
        title="Детали жалобы"
        open={!!selectedComplaint}
        onClose={() => setSelectedComplaint(null)}
        size="large"
        forceRender
        extra={
          selectedComplaint?.status === 'pending' && (
            <Space>
              <Button
                type="primary"
                icon={<CheckOutlined />}
                onClick={() => openResolveModal(selectedComplaint.id!)}
              >
                Решить
              </Button>
              <Button
                danger
                icon={<CloseOutlined />}
                onClick={() => handleDismiss(selectedComplaint)}
              >
                Отклонить
              </Button>
            </Space>
          )
        }
      >
        {selectedComplaint && (
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="Статус">
              <Tag color={statusColor[selectedComplaint.status!] ?? 'default'}>
                {statusLabel[selectedComplaint.status!] ?? selectedComplaint.status}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Тип цели">
              {targetTypeLabel[selectedComplaint.target_type!] ?? selectedComplaint.target_type}
            </Descriptions.Item>
            <Descriptions.Item label="ID цели">
              <Text copyable={{ text: selectedComplaint.target_id }}>
                {selectedComplaint.target_id?.slice(0, 8)}...
              </Text>
            </Descriptions.Item>
            <Descriptions.Item label="Причина">
              <Tag>{reasonLabel[selectedComplaint.reason!] ?? selectedComplaint.reason}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Описание">
              {selectedComplaint.description || <Text type="secondary">Не указано</Text>}
            </Descriptions.Item>
            <Descriptions.Item label="Автор жалобы">
              <Text copyable={{ text: selectedComplaint.reporter_id }}>
                {selectedComplaint.reporter_id?.slice(0, 8)}...
              </Text>
            </Descriptions.Item>
            <Descriptions.Item label="Дата">
              {selectedComplaint.created_at ? formatDateTime(selectedComplaint.created_at) : '-'}
            </Descriptions.Item>
            {selectedComplaint.resolution && (
              <Descriptions.Item label="Решение">
                {selectedComplaint.resolution}
              </Descriptions.Item>
            )}
            {selectedComplaint.resolved_at && (
              <Descriptions.Item label="Дата решения">
                {formatDateTime(selectedComplaint.resolved_at)}
              </Descriptions.Item>
            )}
            {selectedComplaint.resolved_by_id && (
              <Descriptions.Item label="Решил">
                <Text copyable={{ text: selectedComplaint.resolved_by_id }}>
                  {selectedComplaint.resolved_by_id.slice(0, 8)}...
                </Text>
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Drawer>

      <Modal
        title="Решить жалобу"
        open={resolveModalOpen}
        onCancel={() => setResolveModalOpen(false)}
        onOk={handleResolveConfirm}
        okText="Решить"
        cancelText="Отмена"
        confirmLoading={resolveMutation.isPending}
      >
        <div className="rh-admin-modal-description">
          <Text>Укажите примечание к решению (необязательно):</Text>
        </div>
        <Input.TextArea
          rows={3}
          placeholder="Примечание к решению"
          value={resolveNote}
          onChange={(e) => setResolveNote(e.target.value)}
        />
      </Modal>
    </div>
  )
}
