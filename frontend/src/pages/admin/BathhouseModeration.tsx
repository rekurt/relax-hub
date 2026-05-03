import { useState } from 'react'
import { App, Button, Checkbox, Descriptions, Drawer, Input, Segmented, Space, Table, Tag } from '@/components/design/system'
import type { ColumnsType } from '@/components/design/types'
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
import PageHeader from '@/components/PageHeader'

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

  const updateStatusFilter = (value: string) => {
    setStatusFilter(value)
    setPage(1)
    setSelectedIds([])
  }

  const filterRailItems = STATUS_OPTIONS.map((option) => {
    const count = option.value
      ? bathhouses.filter((bathhouse) => bathhouse.status === option.value).length
      : meta?.total_count ?? bathhouses.length
    const railLabel = option.value === ''
      ? 'Все заявки'
      : option.value === 'pending'
        ? 'Ожидают проверки'
        : option.value === 'active'
          ? 'Активные объекты'
          : option.value === 'rejected'
            ? 'Отклонённые заявки'
            : 'Заблокированные'

    return {
      ...option,
      railLabel,
      count,
    }
  })

  const filteredBathhouses = search
    ? bathhouses.filter((b) => {
        const q = search.toLowerCase()
        return (
          b.name?.toLowerCase().includes(q) ||
          b.address?.toLowerCase().includes(q)
        )
      })
    : bathhouses
  const pendingBathhouses = filteredBathhouses.filter((b) => b.status === 'pending')
  const pendingIds = pendingBathhouses.map((b) => b.id!).filter(Boolean)

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
          checked={pendingIds.length > 0 && selectedIds.length === pendingIds.length}
          indeterminate={selectedIds.length > 0 && selectedIds.length < pendingIds.length}
          disabled={pendingIds.length === 0}
          onChange={(e) =>
            setSelectedIds(e.target.checked ? pendingIds : [])
          }
        />
      ),
      key: 'select',
      width: 48,
      render: (_, record) => (
        <Checkbox
          checked={selectedIds.includes(record.id!)}
          disabled={record.status !== 'pending'}
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
        if (record.status === 'pending') {
          return (
            <Space className="rh-admin-row-actions" wrap>
              <Button type="text" size="small" onClick={() => setDetailItem(record)}>
                Детали
              </Button>
              <Button type="primary" size="small" onClick={() => handleApprove(record)}>
              Одобрить
              </Button>
              <Button danger size="small" onClick={() => handleReject(record)}>
                Отклонить
              </Button>
            </Space>
          )
        }

        const statusHint = record.status === 'active'
          ? 'Объект уже опубликован'
          : record.status === 'rejected'
            ? 'Заявка уже отклонена'
            : 'Недоступно для текущего статуса'

        return (
          <Space className="rh-admin-row-actions" wrap>
            <Button type="text" size="small" onClick={() => setDetailItem(record)}>
              Детали
            </Button>
            <Tag color="default">{statusHint}</Tag>
          </Space>
        )
      },
    },
  ]

  return (
    <div className="rh-stack">
      <PageHeader
        size="compact"
        eyebrow="Модерация"
        title="Модерация бань"
        description="Проверка объектов, публичных обещаний и статусов в плотном рабочем интерфейсе."
      />

      <div className="rh-admin-grid rh-admin-grid--filters">
        <aside className="rh-admin-filter-rail">
          <span className="rh-admin-filter-rail__title">Фильтры</span>
          {filterRailItems.map((item) => (
            <button
              className={`rh-admin-filter-rail__item${statusFilter === item.value ? ' rh-admin-filter-rail__item--active' : ''}`}
              key={item.value || 'all'}
              type="button"
              onClick={() => updateStatusFilter(item.value)}
            >
              <span>{item.railLabel}</span>
              <span className="rh-admin-filter-rail__count">{item.count}</span>
            </button>
          ))}
        </aside>

        <section className="rh-admin-table-card">
          <div className="rh-admin-toolbar">
            <div className="rh-admin-toolbar__copy">
              <h2 className="rh-admin-toolbar__title">Реестр объектов</h2>
              <div className="rh-admin-toolbar__hint">Фильтры, поиск и решение модератора находятся рядом с таблицей.</div>
            </div>
            <div className="rh-admin-toolbar__actions">
              <Search
                placeholder="Поиск по названию или адресу"
                allowClear
                onSearch={setSearch}
                onChange={(e) => !e.target.value && setSearch('')}
                style={{ width: 320 }}
              />
            </div>
          </div>

          <Segmented
            options={STATUS_OPTIONS}
            value={statusFilter}
            onChange={(val) => updateStatusFilter(val as string)}
          />

          {selectedIds.length > 0 && (
            <Space wrap>
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
        </section>
      </div>

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
