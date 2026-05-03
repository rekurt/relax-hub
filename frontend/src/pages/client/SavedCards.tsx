import { useState } from 'react'
import { Typography, Table, Card, Button, Tag, Space, Popconfirm, Spin } from '@/components/design/system'
import { CreditCardOutlined, DeleteOutlined, CheckCircleOutlined, StarOutlined } from '@/components/design/icons'
import { App } from '@/components/design/system'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetMySavedCards,
  useDeleteMySavedCardsId,
  usePostMySavedCardsIdDefault,
  getGetMySavedCardsQueryKey,
} from '@/api/generated/saved-cards/saved-cards'
import type { ColumnsType } from '@/components/design/types'
import type { InternalHandlerSavedCardResponse } from '@/api/generated/model'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

const BRAND_COLORS: Record<string, string> = {
  visa: '#1a1f71',
  mastercard: '#eb001b',
  mir: '#4db45e',
  belkart: '#e31e24',
}

function getBrandTone(brand?: string) {
  const normalizedBrand = brand?.toLowerCase() ?? ''
  return BRAND_COLORS[normalizedBrand] ? normalizedBrand : 'default'
}

function formatExpiry(month?: number, year?: number): string {
  if (!month || !year) return '—'
  return `${String(month).padStart(2, '0')}/${String(year).slice(-2)}`
}

export default function SavedCards() {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)

  const { data: cardsData, isLoading } = useGetMySavedCards({ page, page_size: 20 })
  const cards = cardsData?.data ?? []
  const meta = cardsData?.meta

  const defaultCard = cards.find((c) => c.is_default)
  const totalCards = meta?.total_count ?? cards.length

  const deleteMutation = useDeleteMySavedCardsId({
    mutation: {
      onSuccess: () => {
        message.success('Карта удалена')
        queryClient.invalidateQueries({ queryKey: getGetMySavedCardsQueryKey() })
      },
      onError: () => message.error('Не удалось удалить карту'),
    },
  })

  const setDefaultMutation = usePostMySavedCardsIdDefault({
    mutation: {
      onSuccess: () => {
        message.success('Карта установлена по умолчанию')
        queryClient.invalidateQueries({ queryKey: getGetMySavedCardsQueryKey() })
      },
      onError: () => message.error('Не удалось установить карту по умолчанию'),
    },
  })

  const columns: ColumnsType<InternalHandlerSavedCardResponse> = [
    {
      title: 'Карта',
      key: 'card',
      render: (_, record) => {
        const brandTone = getBrandTone(record.brand)
        return (
          <Space className={`rh-saved-card-brand rh-saved-card-brand--${brandTone}`}>
            <CreditCardOutlined className="rh-saved-card-brand__icon" />
            <span>
              <strong className="rh-saved-card-brand__name">
                {record.brand ?? 'Карта'}
              </strong>
              {' •••• '}
              {record.last4 ?? '****'}
            </span>
          </Space>
        )
      },
    },
    {
      title: 'Срок',
      key: 'expiry',
      width: 100,
      render: (_, record) => formatExpiry(record.expiry_month, record.expiry_year),
    },
    {
      title: 'Статус',
      key: 'status',
      width: 150,
      render: (_, record) =>
        record.is_default ? (
          <Tag color="green" icon={<CheckCircleOutlined />}>По умолчанию</Tag>
        ) : null,
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 240,
      render: (_, record) => (
        <Space>
          {!record.is_default && (
            <Button
              size="small"
              icon={<StarOutlined />}
              onClick={() => record.id && setDefaultMutation.mutate({ id: record.id })}
              loading={setDefaultMutation.isPending}
            >
              По умолчанию
            </Button>
          )}
          <Popconfirm
            title="Удалить карту?"
            description="Карта будет удалена из сохранённых способов оплаты."
            onConfirm={() => record.id && deleteMutation.mutate({ id: record.id })}
            okText="Удалить"
            cancelText="Отмена"
            okButtonProps={{ danger: true }}
          >
            <Button
              size="small"
              danger
              icon={<DeleteOutlined />}
              loading={deleteMutation.isPending}
            >
              Удалить
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  if (isLoading) {
    return (
      <Card className="rh-admin-state-card">
        <Spin size="large" />
      </Card>
    )
  }

  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="Личный кабинет"
        title="Сохранённые карты"
        description="Проверяйте карту по умолчанию и быстро управляйте способами оплаты без лишних действий."
      />

      <div className="rh-stat-grid">
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Сохранённых карт</span>
          <span className="rh-stat-tile__value">{totalCards}</span>
          <span className="rh-stat-tile__hint">/ 10</span>
        </div>
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Карта по умолчанию</span>
          <span className={defaultCard ? 'rh-stat-tile__value' : 'rh-stat-tile__value rh-saved-card-empty-value'}>
            {defaultCard ? `${defaultCard.brand ?? ''} •••• ${defaultCard.last4 ?? ''}` : 'Не выбрана'}
          </span>
          <span className="rh-stat-tile__hint">
            Используется первой при оплате, если не выбрана другая карта.
          </span>
        </div>
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Безопасность</span>
          <span className="rh-stat-tile__value">Токенизация</span>
          <span className="rh-stat-tile__hint">
            В интерфейсе отображаются только бренд и последние четыре цифры.
          </span>
        </div>
      </div>

      <Card
        title="Мои карты"
        extra={(
          <Text type="secondary">
            Список карт для быстрых повторных оплат
          </Text>
        )}
      >
        {cards.length === 0 ? (
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">У вас нет сохранённых карт</div>
            <p className="rh-admin-empty-state__text">
              Сохранённые способы оплаты появятся после успешной оплаты картой.
            </p>
          </div>
        ) : (
          <Table
            columns={columns}
            dataSource={cards}
            rowKey="id"
            pagination={
              (meta?.total_pages ?? 1) > 1
                ? {
                    current: page,
                    pageSize: 20,
                    total: meta?.total_count,
                    onChange: setPage,
                    showSizeChanger: false,
                  }
                : false
            }
          />
        )}
      </Card>
    </div>
  )
}
