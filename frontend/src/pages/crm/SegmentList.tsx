import { useState } from 'react'
import {
  Card,
  Col,
  List,
  Row,
  Table,
  Tag,
  Typography,
} from 'antd'
import {
  TeamOutlined,
  UserAddOutlined,
  UserDeleteOutlined,
  CrownOutlined,
  GiftOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import {
  useGetMyCrmSegments,
  useGetMyCrmSegmentsSlugGuests,
} from '@/api/generated/crm/crm'
import type { InternalHandlerGuestCardResponse, InternalHandlerSegmentResponse } from '@/api/generated/model'
import { formatPrice } from '@/lib/format'

const { Title, Text } = Typography

const SEGMENT_ICONS: Record<string, React.ReactNode> = {
  new: <UserAddOutlined />,
  regular: <TeamOutlined />,
  lost: <UserDeleteOutlined />,
  vip: <CrownOutlined />,
  birthday_soon: <GiftOutlined />,
}

const SEGMENT_COLORS: Record<string, string> = {
  new: '#52c41a',
  regular: '#1677ff',
  lost: '#ff4d4f',
  vip: '#faad14',
  birthday_soon: '#722ed1',
}

export default function SegmentList() {
  const [selectedSegment, setSelectedSegment] = useState<string>()
  const [page, setPage] = useState(1)
  const pageSize = 20

  const { data: segmentsData, isLoading: segmentsLoading } = useGetMyCrmSegments()
  const segments = segmentsData?.data ?? []

  const { data: guestsData, isLoading: guestsLoading } = useGetMyCrmSegmentsSlugGuests(
    selectedSegment ?? '',
    { page, page_size: pageSize },
    { query: { enabled: !!selectedSegment } },
  )

  const guests = guestsData?.data ?? []
  const totalCount = guestsData?.meta?.total_count ?? 0

  const columns = [
    {
      title: 'ID клиента',
      dataIndex: 'client_id',
      key: 'client_id',
      render: (id: string) => (
        <Tag style={{ fontFamily: 'monospace' }}>{id?.slice(0, 8)}...</Tag>
      ),
    },
    {
      title: 'Визиты',
      dataIndex: 'visit_count',
      key: 'visit_count',
      width: 100,
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
  ]

  return (
    <div>
      <Title level={3}>Сегменты гостей</Title>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} md={10}>
          <List
            loading={segmentsLoading}
            dataSource={segments}
            locale={{ emptyText: 'Нет сегментов' }}
            renderItem={(item: InternalHandlerSegmentResponse) => (
              <Card
                size="small"
                hoverable
                style={{
                  marginBottom: 8,
                  borderColor: selectedSegment === item.slug ? SEGMENT_COLORS[item.slug ?? ''] : undefined,
                  borderWidth: selectedSegment === item.slug ? 2 : 1,
                }}
                onClick={() => { setSelectedSegment(item.slug); setPage(1) }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <span style={{ color: SEGMENT_COLORS[item.slug ?? ''], fontSize: 20 }}>
                      {SEGMENT_ICONS[item.slug ?? ''] ?? <TeamOutlined />}
                    </span>
                    <div>
                      <Text strong>{item.name}</Text>
                      <br />
                      <Text type="secondary" style={{ fontSize: 12 }}>{item.description}</Text>
                    </div>
                  </div>
                  <Tag color={SEGMENT_COLORS[item.slug ?? '']} style={{ fontSize: 16, padding: '2px 12px' }}>
                    {item.count ?? 0}
                  </Tag>
                </div>
              </Card>
            )}
          />
        </Col>

        <Col xs={24} md={14}>
          {selectedSegment ? (
            <Table
              dataSource={guests}
              columns={columns}
              rowKey="id"
              loading={guestsLoading}
              locale={{ emptyText: 'Нет гостей в этом сегменте' }}
              size="small"
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
          ) : (
            <Card>
              <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
                Выберите сегмент для просмотра гостей
              </div>
            </Card>
          )}
        </Col>
      </Row>
    </div>
  )
}
