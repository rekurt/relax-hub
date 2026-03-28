import { useState } from 'react'
import { App, Button, Checkbox, Descriptions, Drawer, Input, Segmented, Space, Table, Tag, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetAdminBathhouses,
  getGetAdminBathhousesQueryKey,
  usePatchAdminBathhousesIdApprove,
  usePatchAdminBathhousesIdReject,
} from '@/api/generated/admin-bathhouses/admin-bathhouses'
import type { InternalHandlerBathhouseResponse } from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'
import { axiosInstance } from '@/api/axios-instance'

const { Title } = Typography
const { Search } = Input

const STATUS_LABELS: Record<string, { color: string; text: string }> = {
  pending: { color: 'orange', text: 'На рассмотрении' },
  active: { color: 'green', text: 'Активна' },
  rejected: { color: 'red', text: 'Отклонена' },
  blocked: { color: 'default', text: 'Заблокирована' },
}

const STATUS_OPTIONS = [
  { label: 'Все', value: '' },
  { label: 'На рассмотрении', value: 'pending' },
  { label: 'Активные', value: 'active' },
  { label: 'Отклонённые', value: 'rejected' },
  { label: 'Заблокированные', value: 'blocked' },
]

const AMENITY_LABELS: Record<string, string> = {
  has_sauna: 'Сауна',
  has_steam_room: 'Парная',
  has_pool: 'Бассейн',
  has_bbq: 'Мангал',
  has_hot_tub: 'Купель',
  has_karaoke: 'Караоке',
}

export default function BathhouseModeration() {
  const { modal, message } = App.useApp()
  const queryClient = useQueryClient()

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [statusFilter, setStatusFilter] = useState('')
  const [search, setSearch] = useState('')
  const [detailItem, setDetailItem] = useState<InternalHandlerBathhouseResponse | null>(null)
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [batchLoading, setBatchLoading] = useState(false)

  const { data, isLoading } = useGetAdminBathhouses({
    page,
    page_size: pageSize,
    ...(statusFilter ? { status: statusFilter } : {}),
  })
  const approveMutation = usePatchAdminBathhousesIdApprove()
  const rejectMutation = usePatchAdminBathhousesIdReject()

  const bathhouses = data?.data ?? []
  const meta = data?.meta

  const filteredBathhouses = search
    ? bathhouses.filter((b) => {
        const q = search.toLowerCase()
        return (
          b.name?.toLowerCase().includes(q) ||
          b.address?.toLowerCase().includes(q)
        )
      })
    : bathhouses

  const handleApprove = (item: InternalHandlerBathhouseResponse) => {
    modal.confirm({
      title: 'Одобрить баню?',
      content: `«${item.name ?? 'Баня'}» будет активирована и станет видна клиентам.`,
      okText: 'Одобрить',
      cancelText: 'Отмена',
      onOk: () =>
        approveMutation.mutateAsync({ id: item.id! }).then(() => {
          message.success('Баня одобрена')
          queryClient.invalidateQueries({ queryKey: getGetAdminBathhousesQueryKey() })
        }).catch(() => {
          message.error('Не удалось одобрить баню')
        }),
    })
  }

  const handleReject = (item: InternalHandlerBathhouseResponse) => {
    modal.confirm({
      title: 'Отклонить баню?',
      content: `«${item.name ?? 'Баня'}» будет отклонена.`,
      okText: 'Отклонить',
      okType: 'danger',
      cancelText: 'Отмена',
      onOk: () =>
        rejectMutation.mutateAsync({ id: item.id! }).then(() => {
          message.success('Баня отклонена')
          queryClient.invalidateQueries({ queryKey: getGetAdminBathhousesQueryKey() })
        }).catch(() => {
          message.error('Не удалось отклонить баню')
        }),
    })
  }

  const handleBatchAction = (action: 'approve' | 'reject') => {
    const label = action === 'approve' ? 'одобрить' : 'отклонить'
    modal.confirm({
      title: `${action === 'approve' ? 'Одобрить' : 'Отклонить'} выбранные бани (${selectedIds.length})?`,
      okText: action === 'approve' ? 'Одобрить' : 'Отклонить',
      okType: action === 'reject' ? 'danger' : 'primary',
      cancelText: 'Отмена',
      onOk: async () => {
        setBatchLoading(true)
        try {
          const { data: result } = await axiosInstance.post('/admin/listings/batch', {
            action,
            ids: selectedIds,
          })
          const succeeded = result?.data?.succeeded?.length ?? 0
          const failed = result?.data?.failed?.length ?? 0
          if (failed > 0) {
            message.warning(`Обработано: ${succeeded} успешно, ${failed} с ошибками`)
          } else {
            message.success(`Успешно: ${succeeded} бань`)
          }
          setSelectedIds([])
          queryClient.invalidateQueries({ queryKey: getGetAdminBathhousesQueryKey() })
        } catch {
          message.error(`Не удалось ${label} бани`)
        } finally {
          setBatchLoading(false)
        }
      },
    })
  }

  const getAmenities = (item: InternalHandlerBathhouseResponse): string[] => {
    const result: string[] = []
    for (const [key, label] of Object.entries(AMENITY_LABELS)) {
      if (item[key as keyof InternalHandlerBathhouseResponse]) {
        result.push(label)
      }
    }
    return result
  }

  const columns: ColumnsType<InternalHandlerBathhouseResponse> = [
    {
      title: (
        <Checkbox
          checked={selectedIds.length > 0 && selectedIds.length === filteredBathhouses.length}
          indeterminate={selectedIds.length > 0 && selectedIds.length < filteredBathhouses.length}
          onChange={(e) =>
            setSelectedIds(e.target.checked ? filteredBathhouses.map((b) => b.id!).filter(Boolean) : [])
          }
        />
      ),
      key: 'select',
      width: 48,
      render: (_, record) => (
        <Checkbox
          checked={selectedIds.includes(record.id!)}
          onChange={(e) =>
            setSelectedIds((prev) =>
              e.target.checked ? [...prev, record.id!] : prev.filter((id) => id !== record.id!),
            )
          }
        />
      ),
    },
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
      render: (name: string, record) => (
        <a onClick={() => setDetailItem(record)}>{name || '—'}</a>
      ),
    },
    {
      title: 'Адрес',
      dataIndex: 'address',
      key: 'address',
      ellipsis: true,
      responsive: ['md'] as const,
      render: (address: string) => address || '—',
    },
    {
      title: 'Цена/час',
      dataIndex: 'price_per_hour',
      key: 'price_per_hour',
      responsive: ['lg'] as const,
      render: (price: number) => (price ? formatPrice(price) : '—'),
    },
    {
      title: 'Рейтинг',
      dataIndex: 'rating',
      key: 'rating',
      responsive: ['lg'] as const,
      render: (rating: number) => (rating ? rating.toFixed(1) : '—'),
    },
    {
      title: 'Статус',
      key: 'status',
      dataIndex: 'status',
      render: (status: string) => {
        const config = STATUS_LABELS[status] ?? { color: 'default', text: status }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_, record) => {
        const actions: React.ReactNode[] = []
        if (record.status !== 'active') {
          actions.push(
            <a key="approve" onClick={() => handleApprove(record)}>
              Одобрить
            </a>,
          )
        }
        if (record.status !== 'rejected') {
          actions.push(
            <a
              key="reject"
              onClick={() => handleReject(record)}
              style={{ color: '#ff4d4f' }}
            >
              Отклонить
            </a>,
          )
        }
        return <Space>{actions}</Space>
      },
    },
  ]

  return (
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>
        Модерация бань
      </Title>

      <div style={{ marginBottom: 16 }}>
        <Segmented
          options={STATUS_OPTIONS}
          value={statusFilter}
          onChange={(val) => {
            setStatusFilter(val as string)
            setPage(1)
            setSelectedIds([])
          }}
        />
      </div>

      <Search
        placeholder="Поиск по названию или адресу"
        allowClear
        onSearch={setSearch}
        onChange={(e) => !e.target.value && setSearch('')}
        style={{ maxWidth: 400, marginBottom: 16 }}
      />

      {selectedIds.length > 0 && (
        <Space style={{ marginBottom: 16 }}>
          <span>Выбрано: {selectedIds.length}</span>
          <Button type="primary" loading={batchLoading} onClick={() => handleBatchAction('approve')}>
            Одобрить выбранные
          </Button>
          <Button danger loading={batchLoading} onClick={() => handleBatchAction('reject')}>
            Отклонить выбранные
          </Button>
        </Space>
      )}

      <Table
        columns={columns}
        dataSource={filteredBathhouses}
        rowKey="id"
        loading={isLoading}
        locale={{ emptyText: 'Нет бань' }}
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

      <Drawer
        title="Информация о бане"
        open={!!detailItem}
        onClose={() => setDetailItem(null)}
        size="large"
      >
        {detailItem && (
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="Название">
              {detailItem.name ?? '—'}
            </Descriptions.Item>
            <Descriptions.Item label="Адрес">
              {detailItem.address ?? '—'}
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
            <Descriptions.Item label="Цена за час">
              {detailItem.price_per_hour
                ? formatPrice(detailItem.price_per_hour)
                : '—'}
            </Descriptions.Item>
            <Descriptions.Item label="Макс. гостей">
              {detailItem.max_guests ?? '—'}
            </Descriptions.Item>
            <Descriptions.Item label="Рейтинг">
              {detailItem.rating ? detailItem.rating.toFixed(1) : '—'}
            </Descriptions.Item>
            <Descriptions.Item label="Отзывов">
              {detailItem.review_count ?? 0}
            </Descriptions.Item>
            <Descriptions.Item label="Удобства">
              {getAmenities(detailItem).join(', ') || '—'}
            </Descriptions.Item>
            <Descriptions.Item label="Фото проверены">
              {detailItem.is_photo_verified ? 'Да' : 'Нет'}
            </Descriptions.Item>
            <Descriptions.Item label="Создана">
              {detailItem.created_at
                ? formatDateTime(detailItem.created_at)
                : '—'}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Drawer>
    </div>
  )
}
