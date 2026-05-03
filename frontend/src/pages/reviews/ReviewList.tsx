import { useMemo, useState } from 'react'
import {
  App,
  Avatar,
  Button,
  Card,
  Col,
  Image,
  Input,
  List,
  Progress,
  Rate,
  Row,
  Select,
  Space,
  Statistic,
  Tag,
  Typography,
} from '@/components/design/system'
import { MessageOutlined, StarFilled, UserOutlined, WarningOutlined } from '@/components/design/icons'
import dayjs from 'dayjs'
import {
  useGetBathhousesIdReviews,
  usePostReviewsIdResponse,
} from '@/api/generated/reviews/reviews'
import type { InternalHandlerReviewResponse } from '@/api/generated/model'
import { useBathhouseStore } from '@/stores/bathhouse'
import { useQueryClient } from '@tanstack/react-query'
import ReportModal from '@/components/ReportModal'
import type { ReportTargetType } from '@/components/ReportModal'
import EmptyState from '@/components/EmptyState'

const { Title, Paragraph, Text } = Typography
const { TextArea } = Input

const FILTER_OPTIONS = [
  { value: 'all', label: 'Все отзывы' },
  { value: 'no_response', label: 'Без ответа' },
  { value: 'with_response', label: 'С ответом' },
]

function RatingDistribution({ reviews }: { reviews: InternalHandlerReviewResponse[] }) {
  const distribution = useMemo(() => {
    const counts = [0, 0, 0, 0, 0]
    for (const r of reviews) {
      if (r.rating && r.rating >= 1 && r.rating <= 5) {
        counts[r.rating - 1]!++
      }
    }
    return counts
  }, [reviews])

  const total = reviews.length
  const avgRating = total > 0
    ? reviews.reduce((sum, r) => sum + (r.rating ?? 0), 0) / total
    : 0

  return (
    <Card size="small" title={`Статистика отзывов (${total > 0 ? `на странице: ${total}` : 'нет данных'})`} style={{ marginBottom: 16 }}>
      <Row gutter={24}>
        <Col xs={24} sm={8} style={{ textAlign: 'center', marginBottom: 16 }}>
          <Statistic
            title="Средний рейтинг"
            value={avgRating}
            precision={1}
            prefix={<StarFilled style={{ color: '#d97706' }} />}
          />
          <div style={{ marginTop: 4 }}>
            <Rate disabled allowHalf value={avgRating} style={{ fontSize: 14 }} />
          </div>
          <Text type="secondary">{total} отзывов</Text>
        </Col>
        <Col xs={24} sm={16}>
          {[5, 4, 3, 2, 1].map((star) => {
            const count = distribution[star - 1] ?? 0
            const percent = total > 0 ? (count / total) * 100 : 0
            return (
              <Row key={star} align="middle" gutter={8} style={{ marginBottom: 4 }}>
                <Col span={3}>
                  <Space size={2}>
                    <span>{star}</span>
                    <StarFilled style={{ color: '#d97706', fontSize: 12 }} />
                  </Space>
                </Col>
                <Col span={17}>
                  <Progress
                    percent={percent}
                    showInfo={false}
                    size="small"
                    strokeColor="#d97706"
                  />
                </Col>
                <Col span={4}>
                  <Text type="secondary" style={{ fontSize: 12 }}>{count}</Text>
                </Col>
              </Row>
            )
          })}
        </Col>
      </Row>
    </Card>
  )
}

function ReviewResponseForm({
  reviewId,
  onSuccess,
}: {
  reviewId: string
  onSuccess: () => void
}) {
  const { message } = App.useApp()
  const [responseText, setResponseText] = useState('')

  const responseMutation = usePostReviewsIdResponse({
    mutation: {
      onSuccess: () => {
        message.success('Ответ отправлен')
        setResponseText('')
        onSuccess()
      },
      onError: () => message.error('Не удалось отправить ответ'),
    },
  })

  const handleSubmit = () => {
    if (!responseText.trim()) return
    responseMutation.mutate({
      id: reviewId,
      data: { response: responseText.trim() },
    })
  }

  return (
    <div style={{ marginTop: 12 }}>
      <TextArea
        value={responseText}
        onChange={(e) => setResponseText(e.target.value)}
        placeholder="Напишите ответ на отзыв..."
        rows={3}
        maxLength={1000}
        showCount
      />
      <Button
        type="primary"
        icon={<MessageOutlined />}
        onClick={handleSubmit}
        loading={responseMutation.isPending}
        disabled={!responseText.trim()}
        style={{ marginTop: 8 }}
        size="small"
      >
        Отправить ответ
      </Button>
    </div>
  )
}

export default function ReviewList() {
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)
  const queryClient = useQueryClient()

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [filter, setFilter] = useState('all')
  const [reportTarget, setReportTarget] = useState<{ type: ReportTargetType; id: string } | null>(null)
  const hasActiveFilter = filter !== 'all'

  const { data, isLoading } = useGetBathhousesIdReviews(selectedBathhouseId ?? '', {
    page: hasActiveFilter ? 1 : page,
    page_size: hasActiveFilter ? 999 : pageSize,
  }, {
    query: {
      enabled: !!selectedBathhouseId,
    },
  })

  const reviews = data?.data ?? []
  const meta = data?.meta

  const filteredReviews = useMemo(() => {
    const items = data?.data ?? []
    if (filter === 'no_response') return items.filter((r) => !r.owner_response)
    if (filter === 'with_response') return items.filter((r) => !!r.owner_response)
    return items
  }, [data?.data, filter])

  const invalidateReviews = () => {
    queryClient.invalidateQueries({
      queryKey: [`/bathhouses/${selectedBathhouseId}/reviews`],
    })
  }

  if (!selectedBathhouseId) {
    return (
      <div>
        <Title level={3}>Отзывы</Title>
        <EmptyState description="Выберите баню для просмотра отзывов" />
      </div>
    )
  }

  return (
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>
        Отзывы
      </Title>

      <RatingDistribution reviews={reviews} />

      <Space wrap style={{ marginBottom: 16 }}>
        <Select
          value={filter}
          onChange={(val) => { setFilter(val); setPage(1) }}
          options={FILTER_OPTIONS}
          style={{ width: 180 }}
        />
      </Space>

      <List
        loading={isLoading}
        dataSource={filteredReviews}
        locale={{ emptyText: <EmptyState description="Отзывов пока нет. Забронируйте визит, чтобы оставить первый отзыв" /> }}
        pagination={hasActiveFilter ? {
          pageSize: pageSize,
          showSizeChanger: true,
          showTotal: (total) => `Всего: ${total}`,
        } : {
          current: page,
          pageSize: pageSize,
          total: meta?.total_count ?? 0,
          showSizeChanger: true,
          showTotal: (total) => `Всего: ${total}`,
          onChange: (p, ps) => {
            setPage(p)
            setPageSize(ps)
          },
        }}
        renderItem={(review) => (
          <List.Item key={review.id} style={{ alignItems: 'flex-start' }}>
            <List.Item.Meta
              avatar={<Avatar icon={<UserOutlined />} />}
              title={
                <Space>
                  <Rate disabled value={review.rating} style={{ fontSize: 14 }} />
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    {review.created_at ? dayjs(review.created_at).format('DD.MM.YYYY HH:mm') : ''}
                  </Text>
                  {review.status && review.status !== 'approved' && (
                    <Tag color={review.status === 'pending' ? 'orange' : 'red'}>
                      {review.status === 'pending' ? 'На модерации' : 'Отклонён'}
                    </Tag>
                  )}
                </Space>
              }
              description={
                <div>
                  <Paragraph style={{ marginBottom: 8 }}>
                    {review.text || <Text type="secondary">Без текста</Text>}
                  </Paragraph>

                  {review.media && review.media.length > 0 && (
                    <div style={{ marginBottom: 8 }}>
                      <Image.PreviewGroup>
                        <Space wrap>
                          {review.media.map((m) => (
                            m.type === 'image' ? (
                              <Image
                                key={m.id}
                                src={m.thumbnail_url ?? m.url}
                                width={80}
                                height={80}
                                style={{ objectFit: 'cover', borderRadius: 12 }}
                                preview={{ src: m.url }}
                              />
                            ) : (
                              <Tag key={m.id} color="blue">
                                Видео
                              </Tag>
                            )
                          ))}
                        </Space>
                      </Image.PreviewGroup>
                    </div>
                  )}

                  {review.images && review.images.length > 0 && !review.media?.length && (
                    <div style={{ marginBottom: 8 }}>
                      <Image.PreviewGroup>
                        <Space wrap>
                          {review.images.map((url, idx) => (
                            <Image
                              key={idx}
                              src={url}
                              width={80}
                              height={80}
                              style={{ objectFit: 'cover', borderRadius: 12 }}
                            />
                          ))}
                        </Space>
                      </Image.PreviewGroup>
                    </div>
                  )}

                  {review.owner_response && (
                    <Card
                      size="small"
                      style={{ marginTop: 8, background: 'rgba(21, 128, 61, 0.08)' }}
                    >
                      <Text strong>Ваш ответ:</Text>
                      <Paragraph style={{ marginBottom: 0, marginTop: 4 }}>
                        {review.owner_response}
                      </Paragraph>
                      {review.owner_response_at && (
                        <Text type="secondary" style={{ fontSize: 11 }}>
                          {dayjs(review.owner_response_at).format('DD.MM.YYYY HH:mm')}
                        </Text>
                      )}
                    </Card>
                  )}

                  {!review.owner_response && review.id && (
                    <ReviewResponseForm
                      reviewId={review.id}
                      onSuccess={invalidateReviews}
                    />
                  )}

                  {review.id && (
                    <Button
                      type="link"
                      size="small"
                      danger
                      icon={<WarningOutlined />}
                      style={{ marginTop: 8, padding: 0 }}
                      onClick={() => setReportTarget({ type: 'review', id: review.id! })}
                    >
                      Пожаловаться
                    </Button>
                  )}
                </div>
              }
            />
          </List.Item>
        )}
      />

      <ReportModal
        open={!!reportTarget}
        targetType={reportTarget?.type ?? 'review'}
        targetId={reportTarget?.id ?? ''}
        onClose={() => setReportTarget(null)}
      />
    </div>
  )
}
