import { useState } from 'react'
import {
  Typography,
  Table,
  Button,
  Empty,
  Tag,
  Space,
  Popconfirm,
  App,
} from 'antd'
import { DeleteOutlined, BellOutlined, BellFilled } from '@ant-design/icons'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetMySavedSearches,
  useDeleteMySavedSearchesId,
} from '@/api/generated/saved-searches/saved-searches'
import type { InternalHandlerSavedSearchResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'

const { Title } = Typography

export default function SavedSearches() {
  const [page, setPage] = useState(1)
  const { message } = App.useApp()
  const queryClient = useQueryClient()

  const { data, isLoading } = useGetMySavedSearches({ page, page_size: 20 })
  const searches = data?.data ?? []
  const meta = data?.meta

  const deleteMutation = useDeleteMySavedSearchesId({
    mutation: {
      onSuccess: () => {
        message.success('Поиск удалён')
        queryClient.invalidateQueries({ queryKey: ['/my/saved-searches'] })
      },
      onError: () => message.error('Не удалось удалить поиск'),
    },
  })

  const renderFilters = (filters: unknown) => {
    if (!filters || typeof filters !== 'object') return '—'
    const f = filters as Record<string, unknown>
    const parts: string[] = []
    if (f.q) parts.push(`"${f.q}"`)
    if (f.city_id) parts.push(`Город: ${f.city_id}`)
    if (f.price_min || f.price_max) {
      const min = typeof f.price_min === 'number' ? Math.round(f.price_min / 100) : 0
      const max = typeof f.price_max === 'number' ? Math.round(f.price_max / 100) : '∞'
      parts.push(`${min}–${max} ₽`)
    }
    if (f.min_rating) parts.push(`Рейтинг ≥ ${f.min_rating}`)
    if (f.guest_count) parts.push(`Гостей: ${f.guest_count}`)
    if (f.has_sauna) parts.push('Сауна')
    if (f.has_pool) parts.push('Бассейн')
    if (f.has_steam_room) parts.push('Парная')
    return parts.length > 0 ? parts.join(', ') : 'Без фильтров'
  }

  const columns = [
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => name || 'Без названия',
    },
    {
      title: 'Фильтры',
      dataIndex: 'filters',
      key: 'filters',
      render: (filters: unknown) => (
        <span style={{ fontSize: 13, color: '#666' }}>{renderFilters(filters)}</span>
      ),
    },
    {
      title: 'Уведомления',
      dataIndex: 'notify_on_new',
      key: 'notify_on_new',
      width: 140,
      render: (notify: boolean) =>
        notify ? (
          <Tag icon={<BellFilled />} color="blue">
            Включены
          </Tag>
        ) : (
          <Tag icon={<BellOutlined />}>Выключены</Tag>
        ),
    },
    {
      title: 'Создан',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (date: string) => (date ? formatDateTime(date, 'DD.MM.YYYY HH:mm') : '—'),
    },
    {
      title: '',
      key: 'actions',
      width: 80,
      render: (_: unknown, record: InternalHandlerSavedSearchResponse) => (
        <Popconfirm
          title="Удалить сохранённый поиск?"
          onConfirm={() => {
            if (record.id) deleteMutation.mutate({ id: record.id })
          }}
          okText="Удалить"
          cancelText="Отмена"
        >
          <Button
            type="text"
            danger
            icon={<DeleteOutlined />}
            loading={deleteMutation.isPending}
          />
        </Popconfirm>
      ),
    },
  ]

  return (
    <div>
      <Space
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          marginBottom: 24,
        }}
      >
        <Title level={3} style={{ margin: 0 }}>
          Сохранённые поиски
        </Title>
      </Space>

      {searches.length === 0 && !isLoading ? (
        <Empty description="У вас нет сохранённых поисков" />
      ) : (
        <Table
          dataSource={searches}
          columns={columns}
          loading={isLoading}
          rowKey="id"
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
      )}
    </div>
  )
}
