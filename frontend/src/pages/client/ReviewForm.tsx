import { useState } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import {
  Typography,
  Card,
  Button,
  Rate,
  Input,
  Space,
  App,
  Spin,
  Alert,
} from '@/components/design/system'
import { ArrowLeftOutlined, SendOutlined } from '@/components/design/icons'
import { useGetBathhousesId } from '@/api/generated/bathhouses/bathhouses'
import { usePostBathhousesIdReviews, usePutReviewsId } from '@/api/generated/reviews/reviews'
import { usePostReviewsIdMedia, useDeleteMediaId } from '@/api/generated/review-media/review-media'
import MediaUploader, { type MediaFile } from '@/components/MediaUploader'
import { useQueryClient } from '@tanstack/react-query'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography
const { TextArea } = Input

export default function ReviewForm() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { message } = App.useApp()
  const queryClient = useQueryClient()

  const bathhouseId = searchParams.get('bathhouse') ?? ''
  const bookingId = searchParams.get('booking') ?? ''
  const reviewId = searchParams.get('review') ?? ''
  const isEdit = !!reviewId

  const [rating, setRating] = useState(0)
  const [text, setText] = useState('')
  const [mediaFiles, setMediaFiles] = useState<MediaFile[]>([])
  const [submitting, setSubmitting] = useState(false)

  const { data: bathhouseData, isLoading: bathhouseLoading } = useGetBathhousesId(bathhouseId, {
    query: { enabled: !!bathhouseId },
  })
  const bathhouse = bathhouseData?.data

  const createMutation = usePostBathhousesIdReviews()
  const updateMutation = usePutReviewsId()
  const uploadMediaMutation = usePostReviewsIdMedia()
  const deleteMediaMutation = useDeleteMediaId()

  // We need deleteMediaMutation for edit mode (removing existing media)
  void deleteMediaMutation

  const uploadMediaFiles = async (targetReviewId: string) => {
    for (const mf of mediaFiles) {
      await uploadMediaMutation.mutateAsync({
        id: targetReviewId,
        data: { file: mf.file },
      })
    }
  }

  const handleSubmit = async () => {
    if (rating === 0) {
      message.error('Поставьте оценку')
      return
    }

    setSubmitting(true)

    try {
      if (isEdit) {
        await updateMutation.mutateAsync({
          id: reviewId,
          data: { rating, text: text || undefined },
        })
        if (mediaFiles.length > 0) {
          await uploadMediaFiles(reviewId)
        }
        message.success('Отзыв обновлён')
      } else {
        if (!bookingId) {
          message.error('Не указан ID бронирования')
          setSubmitting(false)
          return
        }
        const result = await createMutation.mutateAsync({
          id: bathhouseId,
          data: { rating, text: text || undefined, booking_id: bookingId },
        })
        const newReviewId = result?.data?.id
        if (newReviewId && mediaFiles.length > 0) {
          await uploadMediaFiles(newReviewId)
        }
        message.success('Отзыв отправлен!')
      }

      queryClient.invalidateQueries({ queryKey: [`/bathhouses/${bathhouseId}/reviews`] })
      navigate(`/bathhouses/${bathhouse?.slug ?? bathhouseId}`)
    } catch {
      message.error(isEdit ? 'Не удалось обновить отзыв' : 'Не удалось отправить отзыв')
    } finally {
      setSubmitting(false)
    }
  }

  if (bathhouseLoading) {
    return (
      <div className="rh-fullscreen-state">
        <Spin size="large" />
      </div>
    )
  }

  if (!bathhouseId) {
    return (
      <div className="rh-admin-empty-state">
        <div className="rh-admin-empty-state__title">Не указана баня</div>
        <p className="rh-admin-empty-state__text">
          Вернитесь к карточке бани и откройте форму отзыва из завершённого бронирования.
        </p>
      </div>
    )
  }

  return (
    <div className="rh-stack rh-client-narrow-page">
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate(`/bathhouses/${bathhouse?.slug ?? bathhouseId}`)}
        className="rh-admin-detail-back"
      >
        Назад к бане
      </Button>

      <PageHeader
        eyebrow="Отзыв"
        title={`${isEdit ? 'Редактировать отзыв' : 'Оставить отзыв'}${bathhouse?.name ? `: ${bathhouse.name}` : ''}`}
        description="Оцените визит, добавьте короткое описание и приложите медиа, если это поможет другим гостям."
        size="compact"
      />

      {!isEdit && !bookingId && (
        <Alert
          type="warning"
          showIcon
          title="Для написания отзыва нужно завершённое бронирование"
          className="rh-alert-spaced"
        />
      )}

      <Card className="rh-admin-detail-card">
        <Space orientation="vertical" className="rh-full-width" size={16}>
          <div>
            <Text strong>Оценка:</Text>
            <div className="rh-review-rating-row">
              <Rate
                value={rating}
                onChange={setRating}
                className="rh-review-rating"
              />
              {rating > 0 && (
                <Text className="rh-review-rating-label">
                  {['', 'Ужасно', 'Плохо', 'Нормально', 'Хорошо', 'Отлично'][rating]}
                </Text>
              )}
            </div>
          </div>

          <div>
            <Text strong>Текст отзыва:</Text>
            <TextArea
              value={text}
              onChange={(e) => setText(e.target.value)}
              placeholder="Расскажите о вашем опыте..."
              rows={4}
              maxLength={2000}
              showCount
              className="rh-review-field-control"
            />
          </div>

          <div>
            <Text strong>Фото и видео:</Text>
            <div className="rh-review-field-control">
              <MediaUploader
                files={mediaFiles}
                onChange={setMediaFiles}
                disabled={submitting}
              />
            </div>
          </div>
        </Space>
      </Card>

      <Space>
        <Button
          type="primary"
          size="large"
          icon={<SendOutlined />}
          onClick={handleSubmit}
          loading={submitting}
          disabled={rating === 0 || (!bookingId && !isEdit)}
        >
          {isEdit ? 'Сохранить' : 'Отправить отзыв'}
        </Button>
        <Button
          size="large"
          onClick={() => navigate(`/bathhouses/${bathhouse?.slug ?? bathhouseId}`)}
        >
          Отмена
        </Button>
      </Space>
    </div>
  )
}
