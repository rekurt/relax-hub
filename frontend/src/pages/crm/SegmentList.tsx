import { useState } from 'react'
import {
  Card,
  Col,
  List,
  Row,
  Table,
  Tag,
  Typography,
} from '@/components/design/system'
import {
  TeamOutlined,
  UserAddOutlined,
  UserDeleteOutlined,
  CrownOutlined,
  GiftOutlined,
} from '@/components/design/icons'
import dayjs from 'dayjs'
import {
  useGetMyCrmSegments,
  useGetMyCrmSegmentsSlugGuests,
} from '@/api/generated/crm/crm'
import type { InternalHandlerSegmentResponse } from '@/api/generated/model'
import { formatPrice } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

const SEGMENT_ICONS: Record<string, React.ReactNode> = {
  new: <UserAddOutlined />,
  regular: <TeamOutlined />,
  lost: <UserDeleteOutlined />,
  vip: <CrownOutlined />,
  birthday_soon: <GiftOutlined />,
}

const SEGMENT_COLORS: Record<string, string> = {
  new: '#15803d',
  regular: '#0f766e',
  lost: '#b42318',
  vip: '#d97706',
  birthday_soon: '#0a5f59',
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
        <Tag className="rh-code-tag">{id?.slice(0, 8)}...</Tag>
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
    <div className="rh-stack">
      <PageHeader
        eyebrow="CRM"
        title="Сегменты гостей"
        description="Готовые сегменты и гости, попадающие в выбранную группу."
        size="compact"
      />

      <Row gutter={[16, 16]} className="rh-section-spaced">
        <Col xs={24} md={10}>
          <List
            loading={segmentsLoading}
            dataSource={segments}
            locale={{ emptyText: 'Нет сегментов' }}
            renderItem={(item: InternalHandlerSegmentResponse) => (
              <Card
                size="small"
                hoverable
                className={selectedSegment === item.slug ? 'rh-segment-card rh-segment-card--selected' : 'rh-segment-card'}
                onClick={() => { setSelectedSegment(item.slug); setPage(1) }}
              >
                <div className="rh-segment-card__row">
                  <div className="rh-segment-card__main">
                    <span className={`rh-segment-card__icon rh-segment-card__icon--${item.slug ?? 'default'}`}>
                      {SEGMENT_ICONS[item.slug ?? ''] ?? <TeamOutlined />}
                    </span>
                    <div>
                      <Text strong>{item.name}</Text>
                      <br />
                      <Text type="secondary" className="rh-segment-card__description">{item.description}</Text>
                    </div>
                  </div>
                  <Tag color={SEGMENT_COLORS[item.slug ?? '']} className="rh-segment-count-tag">
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
              <div className="rh-muted-empty">
                Выберите сегмент для просмотра гостей
              </div>
            </Card>
          )}
        </Col>
      </Row>
    </div>
  )
}
