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
import { resolveAssetUrl } from '@/lib/asset-url'

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
    <div className="rh-review-media-row">
      <Image.PreviewGroup>
        <Space wrap size={8}>
          {media.map((m) =>
            m.type === 'image' ? (
              <Image
                key={m.id}
                src={resolveAssetUrl(m.thumbnail_url ?? m.url)}
                alt="Фото отзыва"
                width={80}
                height={80}
                className="rh-review-media-image"
                preview={{ src: resolveAssetUrl(m.url) }}
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
    <Card size="small" className="rh-review-card">
      <div className="rh-review-card__layout">
        <Avatar icon={<UserOutlined />} />
        <div className="rh-review-card__content">
          <div className="rh-review-card__head">
            <Space>
              <Rate disabled value={review.rating ?? 0} className="rh-rate-compact" />
              <Text type="secondary" className="rh-table-meta-text">
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

          <Paragraph className="rh-review-card__text">
            {review.text || <Text type="secondary">Без текста</Text>}
          </Paragraph>

          {review.media && review.media.length > 0 && renderMedia(review.media)}

          {/* Legacy images field */}
          {review.images && review.images.length > 0 && !review.media?.length && (
            <div className="rh-review-media-row">
              <Image.PreviewGroup>
                <Space wrap>
                  {review.images.map((url, idx) => (
                    <Image
                      key={idx}
                      src={resolveAssetUrl(url)}
                      width={80}
                      height={80}
                      className="rh-review-media-image"
                    />
                  ))}
                </Space>
              </Image.PreviewGroup>
            </div>
          )}

          {review.owner_response && (
            <Card size="small" className="rh-owner-response-card">
              <Text strong>Ответ владельца:</Text>
              <Paragraph className="rh-owner-response-text">{review.owner_response}</Paragraph>
              {review.owner_response_at && (
                <Text type="secondary" className="rh-table-meta-text">
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
        <Space orientation="vertical" className="rh-full-width" size={12}>
          <div>
            <Text strong>Причина:</Text>
            <Select
              value={reportReason || undefined}
              onChange={setReportReason}
              options={REPORT_REASONS}
              placeholder="Выберите причину"
              className="rh-modal-control-offset"
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
              className="rh-modal-control-offset"
            />
          </div>
        </Space>
      </Modal>
    </Card>
  )
}
