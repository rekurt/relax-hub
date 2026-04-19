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
  Empty,
  Alert,
} from 'antd'
import { ArrowLeftOutlined, SendOutlined } from '@ant-design/icons'
import { useGetBathhousesId } from '@/api/generated/bathhouses/bathhouses'
import { usePostBathhousesIdReviews, usePutReviewsId } from '@/api/generated/reviews/reviews'
import { usePostReviewsIdMedia, useDeleteMediaId } from '@/api/generated/review-media/review-media'
import MediaUploader, { type MediaFile } from '@/components/MediaUploader'
import { useQueryClient } from '@tanstack/react-query'

const { Title, Text } = Typography
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
    return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  }

  if (!bathhouseId) {
    return <Empty description="Не указана баня" />
  }

  return (
    <div style={{ maxWidth: 600, margin: '0 auto' }}>
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate(`/bathhouses/${bathhouse?.slug ?? bathhouseId}`)}
        style={{ marginBottom: 16 }}
      >
        Назад к бане
      </Button>

      <Title level={3}>
        {isEdit ? 'Редактировать отзыв' : 'Оставить отзыв'}
        {bathhouse?.name && `: ${bathhouse.name}`}
      </Title>

      {!isEdit && !bookingId && (
        <Alert
          type="warning"
          showIcon
          message="Для написания отзыва нужно завершённое бронирование"
          style={{ marginBottom: 16 }}
        />
      )}

      <Card style={{ marginBottom: 16 }}>
        <Space direction="vertical" style={{ width: '100%' }} size={16}>
          <div>
            <Text strong>Оценка:</Text>
            <div style={{ marginTop: 8 }}>
              <Rate
                value={rating}
                onChange={setRating}
                style={{ fontSize: 32 }}
              />
              {rating > 0 && (
                <Text style={{ marginLeft: 12 }}>
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
              style={{ marginTop: 8 }}
            />
          </div>

          <div>
            <Text strong>Фото и видео:</Text>
            <div style={{ marginTop: 8 }}>
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
