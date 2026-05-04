import { useState } from 'react'
import {
  App,
  Button,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
} from '@/components/design/system'
import {
  PlusOutlined,
  SendOutlined,
} from '@/components/design/icons'
import dayjs from 'dayjs'
import { useNavigate } from 'react-router-dom'
import {
  useGetMyCrmBroadcasts,
  usePostMyCrmBroadcastsIdSend,
} from '@/api/generated/crm/crm'
import type { InternalHandlerBroadcastResponse } from '@/api/generated/model'
import { useQueryClient } from '@tanstack/react-query'
import PageHeader from '@/components/PageHeader'
import EmptyState from '@/components/EmptyState'

const STATUS_COLORS: Record<string, string> = {
  draft: 'default',
  sending: 'processing',
  sent: 'success',
  failed: 'error',
}

const STATUS_LABELS: Record<string, string> = {
  draft: 'Черновик',
  sending: 'Отправляется',
  sent: 'Отправлено',
  failed: 'Ошибка',
}

const SEGMENT_LABELS: Record<string, string> = {
  new: 'Новые',
  regular: 'Постоянные',
  lost: 'Потерянные',
  vip: 'VIP',
  birthday_soon: 'День рождения',
}

export default function BroadcastList() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { message } = App.useApp()

  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState<string>()
  const pageSize = 20

  const { data, isLoading } = useGetMyCrmBroadcasts({
    page,
    page_size: pageSize,
    status: statusFilter,
  })

  const broadcasts = data?.data ?? []
  const totalCount = data?.meta?.total_count ?? 0

  const sendMutation = usePostMyCrmBroadcastsIdSend({
    mutation: {
      onSuccess: () => {
        message.success('Рассылка отправлена')
        queryClient.invalidateQueries({ queryKey: ['/my/crm/broadcasts'] })
      },
      onError: (error: { response?: { status?: number } }) => {
        if (error?.response?.status === 429) {
          message.error('Лимит рассылок исчерпан (макс. 3 в неделю)')
        } else {
          message.error('Не удалось отправить рассылку')
        }
      },
    },
  })

  const columns = [
    {
      title: 'Заголовок',
      dataIndex: 'title',
      key: 'title',
      ellipsis: true,
    },
    {
      title: 'Сегмент',
      dataIndex: 'segment',
      key: 'segment',
      width: 140,
      render: (seg: string) => SEGMENT_LABELS[seg] ?? seg,
    },
    {
      title: 'Каналы',
      dataIndex: 'channels',
      key: 'channels',
      width: 180,
      render: (channels: string[]) =>
        channels?.map((ch) => (
          <Tag key={ch} color="blue">{ch}</Tag>
        )) ?? '—',
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      width: 130,
      render: (status: string) => (
        <Tag color={STATUS_COLORS[status] ?? 'default'}>
          {STATUS_LABELS[status] ?? status}
        </Tag>
      ),
    },
    {
      title: 'Доставлено / Прочитано / Клики',
      key: 'stats',
      width: 210,
      render: (_: unknown, record: InternalHandlerBroadcastResponse) => (
        <span>{record.delivered ?? 0} / {record.read ?? 0} / {(record as Record<string, unknown>).clicked as number ?? 0}</span>
      ),
    },
    {
      title: 'Дата',
      key: 'date',
      width: 140,
      render: (_: unknown, record: InternalHandlerBroadcastResponse) => {
        const date = record.sent_at ?? record.created_at
        return date ? dayjs(date).format('DD.MM.YYYY HH:mm') : '—'
      },
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 100,
      render: (_: unknown, record: InternalHandlerBroadcastResponse) =>
        record.status === 'draft' ? (
          <Popconfirm
            title="Отправить рассылку?"
            description="Сообщение будет отправлено всем гостям в выбранном сегменте"
            onConfirm={() => record.id && sendMutation.mutate({ id: record.id })}
            okText="Отправить"
            cancelText="Отмена"
          >
            <Button type="link" icon={<SendOutlined />} size="small">
              Отправить
            </Button>
          </Popconfirm>
        ) : null,
    },
  ]

  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="CRM"
        title="Рассылки"
        description="Черновики и отправленные кампании по сегментам гостей."
        size="compact"
        extra={(
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/crm/broadcasts/new')}>
            Создать рассылку
          </Button>
        )}
      />

      <Space className="rh-page-toolbar">
        <Select
          placeholder="Фильтр по статусу"
          value={statusFilter}
          onChange={(v) => { setStatusFilter(v); setPage(1) }}
          allowClear
          className="rh-crm-status-filter"
          options={[
            { value: 'draft', label: 'Черновик' },
            { value: 'sending', label: 'Отправляется' },
            { value: 'sent', label: 'Отправлено' },
            { value: 'failed', label: 'Ошибка' },
          ]}
        />
      </Space>

      <Table
        dataSource={broadcasts}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        locale={{
          emptyText: (
            <EmptyState description="Нет рассылок. Создайте рассылку для информирования гостей об акциях и новостях." />
          ),
        }}
        pagination={
          totalCount > pageSize
            ? {
                current: page,
                pageSize,
                total: totalCount,
                onChange: setPage,
                showSizeChanger: false,
              }
            : false
        }
      />
    </div>
  )
}
