import { useState } from 'react'
import {
  Card,
  Rate,
  Image,
  Space,
  Tag,
  Typography,
  Button,
  Modal,
  Input,
  Select,
  App,
  Avatar,
} from '@/components/design/system'
import {
  UserOutlined,
  FlagOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@/components/design/icons'
import dayjs from 'dayjs'
import type { InternalHandlerReviewResponse, InternalHandlerMediaResponse } from '@/api/generated/model'
import { usePostReviewsIdReport } from '@/api/generated/complaints/complaints'
import { useDeleteReviewsId } from '@/api/generated/reviews/reviews'

const { Paragraph, Text } = Typography
const { TextArea } = Input

const REPORT_REASONS = [
  { value: 'spam', label: 'Спам' },
  { value: 'offensive', label: 'Оскорбительный контент' },
  { value: 'fake', label: 'Фейковый отзыв' },
  { value: 'fraud', label: 'Мошенничество' },
  { value: 'other', label: 'Другое' },
]

interface ReviewCardProps {
  review: InternalHandlerReviewResponse
  showActions?: boolean
  isAuthor?: boolean
  onEdit?: (reviewId: string) => void
  onDeleted?: () => void
}

export default function ReviewCard({
  review,
  showActions = true,
  isAuthor = false,
  onEdit,
  onDeleted,
}: ReviewCardProps) {
  const { message, modal } = App.useApp()
  const [reportOpen, setReportOpen] = useState(false)
  const [reportReason, setReportReason] = useState('')
  const [reportDescription, setReportDescription] = useState('')

  const reportMutation = usePostReviewsIdReport({
    mutation: {
      onSuccess: () => {
        message.success('Жалоба отправлена')
        setReportOpen(false)
        setReportReason('')
        setReportDescription('')
      },
      onError: () => message.error('Не удалось отправить жалобу'),
    },
  })

  const deleteMutation = useDeleteReviewsId({
    mutation: {
      onSuccess: () => {
        message.success('Отзыв удалён')
        onDeleted?.()
      },
      onError: () => message.error('Не удалось удалить отзыв'),
    },
  })

  const handleReport = () => {
    if (!reportReason || !review.id) return
    reportMutation.mutate({
      id: review.id,
      data: { reason: reportReason, description: reportDescription || undefined },
    })
  }

  const handleDelete = () => {
    if (!review.id) return
    modal.confirm({
      title: 'Удалить отзыв?',
      content: 'Это действие нельзя отменить.',
      okText: 'Удалить',
      okType: 'danger',
      cancelText: 'Отмена',
      onOk: () => deleteMutation.mutate({ id: review.id! }),
    })
  }

  const renderMedia = (media: InternalHandlerMediaResponse[]) => (
    <div style={{ marginTop: 8 }}>
      <Image.PreviewGroup>
        <Space wrap size={8}>
          {media.map((m) =>
            m.type === 'image' ? (
              <Image
                key={m.id}
                src={m.thumbnail_url ?? m.url}
                alt="Фото отзыва"
                width={80}
                height={80}
                style={{ borderRadius: 12, objectFit: 'cover' }}
                preview={{ src: m.url }}
              />
            ) : (
              <Tag key={m.id} color="blue" icon={<span>&#9654;</span>}>
                Видео
              </Tag>
            ),
          )}
        </Space>
      </Image.PreviewGroup>
    </div>
  )

  return (
    <Card size="small" style={{ marginBottom: 12 }}>
      <div style={{ display: 'flex', gap: 12 }}>
        <Avatar icon={<UserOutlined />} />
        <div style={{ flex: 1 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Space>
              <Rate disabled value={review.rating ?? 0} style={{ fontSize: 14 }} />
              <Text type="secondary" style={{ fontSize: 12 }}>
                {review.created_at ? dayjs(review.created_at).format('DD.MM.YYYY') : ''}
              </Text>
              {review.status && review.status !== 'approved' && (
                <Tag color={review.status === 'pending' ? 'orange' : 'red'}>
                  {review.status === 'pending' ? 'На модерации' : 'Отклонён'}
                </Tag>
              )}
            </Space>
            {showActions && (
              <Space size={4}>
                {isAuthor && onEdit && review.id && (
                  <Button
                    type="text"
                    size="small"
                    icon={<EditOutlined />}
                    onClick={() => onEdit(review.id!)}
                  />
                )}
                {isAuthor && review.id && (
                  <Button
                    type="text"
                    size="small"
                    danger
                    icon={<DeleteOutlined />}
                    onClick={handleDelete}
                    loading={deleteMutation.isPending}
                  />
                )}
                {!isAuthor && review.id && (
                  <Button
                    type="text"
                    size="small"
                    icon={<FlagOutlined />}
                    onClick={() => setReportOpen(true)}
                  >
                    Жалоба
                  </Button>
                )}
              </Space>
            )}
          </div>

          <Paragraph style={{ marginTop: 8, marginBottom: 4 }}>
            {review.text || <Text type="secondary">Без текста</Text>}
          </Paragraph>

          {review.media && review.media.length > 0 && renderMedia(review.media)}

          {/* Legacy images field */}
          {review.images && review.images.length > 0 && !review.media?.length && (
            <div style={{ marginTop: 8 }}>
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
            <Card size="small" style={{ marginTop: 8, background: 'rgba(21, 128, 61, 0.08)' }}>
              <Text strong>Ответ владельца:</Text>
              <Paragraph style={{ margin: '4px 0 0' }}>{review.owner_response}</Paragraph>
              {review.owner_response_at && (
                <Text type="secondary" style={{ fontSize: 11 }}>
                  {dayjs(review.owner_response_at).format('DD.MM.YYYY HH:mm')}
                </Text>
              )}
            </Card>
          )}
        </div>
      </div>

      <Modal
        title="Пожаловаться на отзыв"
        open={reportOpen}
        onCancel={() => setReportOpen(false)}
        onOk={handleReport}
        okText="Отправить"
        cancelText="Отмена"
        confirmLoading={reportMutation.isPending}
        okButtonProps={{ disabled: !reportReason }}
      >
        <Space orientation="vertical" style={{ width: '100%' }} size={12}>
          <div>
            <Text strong>Причина:</Text>
            <Select
              value={reportReason || undefined}
              onChange={setReportReason}
              options={REPORT_REASONS}
              placeholder="Выберите причину"
              style={{ width: '100%', marginTop: 4 }}
            />
          </div>
          <div>
            <Text strong>Описание (необязательно):</Text>
            <TextArea
              value={reportDescription}
              onChange={(e) => setReportDescription(e.target.value)}
              placeholder="Подробности жалобы..."
              rows={3}
              maxLength={500}
              showCount
              style={{ marginTop: 4 }}
            />
          </div>
        </Space>
      </Modal>
    </Card>
  )
}
