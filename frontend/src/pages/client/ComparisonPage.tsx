import { useState, useEffect } from 'react'
import {
  Typography,
  Table,
  Button,
  Empty,
  Image,
  Rate,
  Tag,
  Space,
  App,
} from 'antd'
import { DeleteOutlined } from '@ant-design/icons'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { usePostApiV1BathhousesCompare } from '@/api/generated/bathhouses/bathhouses'
import type { InternalHandlerComparisonItem, InternalHandlerCompareResponse } from '@/api/generated/model'
import { formatPrice } from '@/lib/format'

const { Title, Text } = Typography

const AMENITY_LABELS: Record<string, string> = {
  has_sauna: 'Сауна',
  has_steam_room: 'Парная',
  has_pool: 'Бассейн',
  has_hot_tub: 'Джакузи',
  has_bbq: 'Мангал',
  has_karaoke: 'Караоке',
}

interface ComparisonRow {
  key: string
  label: string
  values: (string | React.ReactNode)[]
}

function buildRows(items: InternalHandlerComparisonItem[]): ComparisonRow[] {
  const rows: ComparisonRow[] = [
    {
      key: 'image',
      label: 'Фото',
      values: items.map((item) =>
        item.images?.[0] ? (
          <Image
            src={item.images[0]}
            alt={item.name}
            height={120}
            style={{ objectFit: 'cover', borderRadius: 8 }}
            preview={false}
          />
        ) : (
          <div
            style={{
              height: 120,
              background: '#f5f5f5',
              borderRadius: 8,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Text type="secondary">Нет фото</Text>
          </div>
        ),
      ),
    },
    {
      key: 'price',
      label: 'Цена за час',
      values: items.map((item) =>
        item.price_per_hour ? (
          <Text strong style={{ fontSize: 16 }}>
            {formatPrice(item.price_per_hour)}
          </Text>
        ) : (
          '—'
        ),
      ),
    },
    {
      key: 'rating',
      label: 'Рейтинг',
      values: items.map((item) => (
        <Space>
          <Rate disabled allowHalf value={item.rating ?? 0} style={{ fontSize: 14 }} />
          <Text type="secondary">
            {item.rating?.toFixed(1)} ({item.review_count ?? 0})
          </Text>
        </Space>
      )),
    },
    {
      key: 'address',
      label: 'Адрес',
      values: items.map((item) => item.address || '—'),
    },
    {
      key: 'max_guests',
      label: 'Макс. гостей',
      values: items.map((item) => (item.max_guests ? `до ${item.max_guests}` : '—')),
    },
    {
      key: 'min_duration',
      label: 'Мин. время',
      values: items.map((item) =>
        item.min_duration ? `${item.min_duration} ч` : '—',
      ),
    },
    {
      key: 'distance',
      label: 'Расстояние',
      values: items.map((item) =>
        item.distance != null ? `${item.distance.toFixed(1)} км` : '—',
      ),
    },
  ]

  const amenityKeys = Object.keys(AMENITY_LABELS)
  for (const amenityKey of amenityKeys) {
    rows.push({
      key: amenityKey,
      label: AMENITY_LABELS[amenityKey] ?? amenityKey,
      values: items.map((item) =>
        item[amenityKey as keyof InternalHandlerComparisonItem] ? (
          <Tag color="green">Есть</Tag>
        ) : (
          <Tag>Нет</Tag>
        ),
      ),
    })
  }

  return rows
}

export default function ComparisonPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const navigate = useNavigate()
  const { message } = App.useApp()
  const [items, setItems] = useState<InternalHandlerComparisonItem[]>([])

  const ids = searchParams.get('ids')?.split(',').filter(Boolean) ?? []

  const compareMutation = usePostApiV1BathhousesCompare({
    mutation: {
      onSuccess: (data) => {
        const compareData = data?.data as InternalHandlerCompareResponse | undefined
        const responseItems = compareData?.items ?? []
        setItems(responseItems)
      },
      onError: () => {
        message.error('Не удалось загрузить данные для сравнения')
      },
    },
  })

  useEffect(() => {
    if (ids.length >= 2 && ids.length <= 3) {
      compareMutation.mutate({ data: { ids } })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams.toString()])

  const handleRemove = (index: number) => {
    const newIds = ids.filter((_, i) => i !== index)
    if (newIds.length < 2) {
      message.info('Для сравнения нужно минимум 2 бани')
      navigate('/client/search')
      return
    }
    setSearchParams({ ids: newIds.join(',') })
  }

  if (ids.length < 2) {
    return (
      <div style={{ textAlign: 'center', padding: '48px 0' }}>
        <Empty description="Выберите минимум 2 бани для сравнения" />
        <Button
          type="primary"
          style={{ marginTop: 16 }}
          onClick={() => navigate('/client/search')}
        >
          Перейти к поиску
        </Button>
      </div>
    )
  }

  const rows = buildRows(items)

  const columns = [
    {
      title: 'Параметр',
      dataIndex: 'label',
      key: 'label',
      width: 160,
      fixed: 'left' as const,
      render: (label: string) => <Text strong>{label}</Text>,
    },
    ...items.map((item, index) => ({
      title: (
        <Space orientation="vertical" size={4} align="center">
          <Button
            type="link"
            style={{ padding: 0, fontSize: 16, fontWeight: 600 }}
            onClick={() => navigate(`/client/bathhouse/${item.slug ?? item.id}`)}
          >
            {item.name}
          </Button>
          <Button
            type="text"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleRemove(index)}
          >
            Убрать
          </Button>
        </Space>
      ),
      dataIndex: ['values', index],
      key: item.id ?? `col-${index}`,
      width: 260,
      render: (_: unknown, record: ComparisonRow) => record.values[index],
    })),
  ]

  return (
    <div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 24,
        }}
      >
        <Title level={3} style={{ margin: 0 }}>
          Сравнение бань
        </Title>
        <Button onClick={() => navigate('/client/search')}>Назад к поиску</Button>
      </div>

      <Table
        dataSource={rows}
        columns={columns}
        pagination={false}
        bordered
        scroll={{ x: 'max-content' }}
        loading={compareMutation.isPending}
        rowKey="key"
      />
    </div>
  )
}
