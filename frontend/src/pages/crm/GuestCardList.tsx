import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Col,
  Input,
  Row,
  Select,
  Space,
  Statistic,
  Table,
  Tag,
  Typography,
} from 'antd'
import {
  DownloadOutlined,
  SearchOutlined,
  UserOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import {
  useGetMyCrmGuests,
  useGetMyCrmStats,
} from '@/api/generated/crm/crm'
import type { InternalHandlerGuestCardResponse } from '@/api/generated/model'
import { formatPrice } from '@/lib/format'
import { useNavigate } from 'react-router-dom'
import { customInstance } from '@/api/axios-instance'

const { Title } = Typography

const SORT_OPTIONS = [
  { value: 'last_visit', label: 'Последний визит' },
  { value: 'total_spent', label: 'Общая сумма' },
  { value: 'visit_count', label: 'Кол-во визитов' },
  { value: 'avg_check', label: 'Средний чек' },
]

export default function GuestCardList() {
  const navigate = useNavigate()
  const { message } = App.useApp()

  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [tag, setTag] = useState<string>()
  const [sortBy, setSortBy] = useState('last_visit')
  const pageSize = 20

  const { data: statsData } = useGetMyCrmStats()
  const stats = statsData?.data

  const { data, isLoading } = useGetMyCrmGuests({
    page,
    page_size: pageSize,
    search: search || undefined,
    tag,
    sort_by: sortBy,
  })

  const guests = data?.data ?? []
  const totalCount = data?.meta?.total_count ?? 0

  const handleExportCSV = async () => {
    try {
      const response = await customInstance<Blob>({
        url: '/my/crm/guests/export',
        method: 'GET',
        responseType: 'blob',
      })
      const url = window.URL.createObjectURL(new Blob([response]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', `guests_${dayjs().format('YYYY-MM-DD')}.csv`)
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
      message.success('CSV экспортирован')
    } catch {
      message.error('Не удалось экспортировать')
    }
  }

  const columns = [
    {
      title: 'Клиент',
      dataIndex: 'client_id',
      key: 'client_id',
      render: (id: string) => (
        <Tag icon={<UserOutlined />} style={{ fontFamily: 'monospace' }}>
          {id?.slice(0, 8)}...
        </Tag>
      ),
    },
    {
      title: 'Визиты',
      dataIndex: 'visit_count',
      key: 'visit_count',
      width: 100,
      sorter: true,
      render: (count: number) => <Tag color="blue">{count ?? 0}</Tag>,
    },
    {
      title: 'Общая сумма',
      dataIndex: 'total_spent',
      key: 'total_spent',
      width: 140,
      render: (val: number) => formatPrice(val ?? 0),
    },
    {
      title: 'Средний чек',
      dataIndex: 'avg_check',
      key: 'avg_check',
      width: 140,
      render: (val: number) => formatPrice(val ?? 0),
    },
    {
      title: 'Последний визит',
      dataIndex: 'last_visit_at',
      key: 'last_visit_at',
      width: 160,
      render: (date: string) => date ? dayjs(date).format('DD.MM.YYYY') : '—',
    },
    {
      title: 'Теги',
      dataIndex: 'tags',
      key: 'tags',
      render: (tags: string[]) =>
        tags?.length ? tags.map((t) => <Tag key={t}>{t}</Tag>) : '—',
    },
    {
      title: 'Заметки',
      dataIndex: 'notes',
      key: 'notes',
      ellipsis: true,
      width: 200,
      render: (notes: string) => notes || '—',
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>Гостевые карточки</Title>
        <Button icon={<DownloadOutlined />} onClick={handleExportCSV}>
          Экспорт CSV
        </Button>
      </div>

      {stats && (
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col xs={12} sm={6}>
            <Card size="small">
              <Statistic title="Всего гостей" value={stats.total_guests ?? 0} />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card size="small">
              <Statistic title="Новых за месяц" value={stats.new_this_month ?? 0} />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card size="small">
              <Statistic title="Ср. визитов" value={stats.avg_visit_count ?? 0} />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card size="small">
              <Statistic title="Ср. чек" value={formatPrice(stats.avg_spent ?? 0)} />
            </Card>
          </Col>
        </Row>
      )}

      <Space style={{ marginBottom: 16 }} wrap>
        <Input
          placeholder="Поиск по имени, email, телефону"
          prefix={<SearchOutlined />}
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          style={{ width: 280 }}
          allowClear
        />
        <Input
          placeholder="Фильтр по тегу"
          value={tag}
          onChange={(e) => { setTag(e.target.value || undefined); setPage(1) }}
          style={{ width: 160 }}
          allowClear
        />
        <Select
          value={sortBy}
          onChange={(v) => { setSortBy(v); setPage(1) }}
          options={SORT_OPTIONS}
          style={{ width: 180 }}
        />
      </Space>

      <Table
        dataSource={guests}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        locale={{ emptyText: 'Нет гостевых карточек' }}
        onRow={(record: InternalHandlerGuestCardResponse) => ({
          onClick: () => record.id && navigate(`/crm/guests/${record.id}`),
          style: { cursor: 'pointer' },
        })}
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
