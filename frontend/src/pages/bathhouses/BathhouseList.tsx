import { App, Badge, Button, Popconfirm, Rate, Space, Table, Tag, Tooltip } from '@/components/design/system'
import type { ColumnsType } from '@/components/design/types'
import { CopyOutlined, DeleteOutlined, EditOutlined, ImportOutlined, PlusOutlined } from '@/components/design/icons'
import { useNavigate } from 'react-router-dom'
import {
  useGetMyBathhouses,
  useDeleteBathhousesId,
  usePostMyBathhousesIdDuplicate,
} from '@/api/generated/bathhouses/bathhouses'
import type { InternalHandlerBathhouseResponse } from '@/api/generated/model'
import { useAuthStore } from '@/stores/auth'
import { formatPrice } from '@/lib/format'
import { useQueryClient } from '@tanstack/react-query'
import PageHeader from '@/components/PageHeader'

const STATUS_MAP: Record<string, { color: string; text: string }> = {
  active: { color: 'green', text: 'Активна' },
  pending: { color: 'orange', text: 'На модерации' },
  rejected: { color: 'red', text: 'Отклонена' },
  inactive: { color: 'default', text: 'Неактивна' },
}

export default function BathhouseList() {
  const navigate = useNavigate()
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const user = useAuthStore((s) => s.user)
  const isOwner = user?.role === 'owner'

  const { data, isLoading } = useGetMyBathhouses()

  const deleteMutation = useDeleteBathhousesId({
    mutation: {
      onSuccess: () => {
        message.success('Баня удалена')
        queryClient.invalidateQueries({ queryKey: ['/my/bathhouses'] })
      },
      onError: () => {
        message.error('Не удалось удалить баню')
      },
    },
  })

  const duplicateMutation = usePostMyBathhousesIdDuplicate({
    mutation: {
      onSuccess: (response) => {
        message.success('Объект дублирован')
        queryClient.invalidateQueries({ queryKey: ['/my/bathhouses'] })
        const newId = (response?.data as { id?: string })?.id
        if (newId) {
          navigate(`/bathhouses/${newId}/edit`)
        }
      },
      onError: () => {
        message.error('Не удалось дублировать объект')
      },
    },
  })

  const bathhouses = data?.data ?? []

  const columns: ColumnsType<InternalHandlerBathhouseResponse> = [
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => <strong>{name}</strong>,
    },
    {
      title: 'Адрес',
      dataIndex: 'address',
      key: 'address',
      responsive: ['md'],
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const info = STATUS_MAP[status] ?? { color: 'default', text: status }
        return <Badge status={info.color as 'success'} text={info.text} />
      },
    },
    {
      title: 'Рейтинг',
      dataIndex: 'rating',
      key: 'rating',
      render: (rating: number) => (
        <Space>
          <Rate disabled value={rating ?? 0} allowHalf className="rh-rate-compact" />
          <span>{rating?.toFixed(1) ?? '—'}</span>
        </Space>
      ),
      responsive: ['lg'],
    },
    {
      title: 'Отзывы',
      dataIndex: 'review_count',
      key: 'review_count',
      render: (count: number) => <Tag>{count ?? 0}</Tag>,
      responsive: ['md'],
    },
    {
      title: 'Цена/час',
      dataIndex: 'price_per_hour',
      key: 'price_per_hour',
      render: (price: number) => formatPrice(price ?? 0),
      responsive: ['sm'],
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => navigate(`/bathhouses/${record.id}/edit`)}
          >
            Изменить
          </Button>
          {isOwner && (
            <Tooltip title="Дублировать">
              <Button
                type="link"
                icon={<CopyOutlined />}
                loading={duplicateMutation.isPending && duplicateMutation.variables?.id === record.id}
                onClick={() => record.id && duplicateMutation.mutate({ id: record.id })}
              >
                Дублировать
              </Button>
            </Tooltip>
          )}
          {isOwner && (
            <Popconfirm
              title="Удалить баню?"
              description="Это действие необратимо. Баня со всеми данными будет удалена."
              onConfirm={() => record.id && deleteMutation.mutate({ id: record.id })}
              okText="Удалить"
              cancelText="Отмена"
              okButtonProps={{ danger: true }}
            >
              <Button
                type="link"
                danger
                icon={<DeleteOutlined />}
                loading={deleteMutation.isPending && deleteMutation.variables?.id === record.id}
              >
                Удалить
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="Объекты"
        title="Мои бани"
        description="Рабочий список объектов с быстрым переходом к редактированию, импорту и дублированию карточек."
        extra={(
          isOwner ? (
            <Space>
              <Button
                icon={<ImportOutlined />}
                onClick={() => navigate('/bathhouses/import')}
              >
                Импорт
              </Button>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => navigate('/bathhouses/new')}
              >
                Добавить баню
              </Button>
            </Space>
          ) : null
        )}
      />
      <Table
        columns={columns}
        dataSource={bathhouses}
        rowKey="id"
        loading={isLoading}
        pagination={false}
        locale={{
          emptyText: (
            <div className="rh-admin-empty-state">
              <div className="rh-admin-empty-state__title">
                У вас пока нет объектов. Добавьте первую баню, чтобы начать принимать бронирования.
              </div>
              <p className="rh-admin-empty-state__text">
                После модерации объект появится в каталоге и сможет принимать бронирования.
              </p>
              {isOwner && (
                <Button type="primary" onClick={() => navigate('/bathhouses/new')}>
                  Добавить баню
                </Button>
              )}
            </div>
          ),
        }}
      />
    </div>
  )
}
