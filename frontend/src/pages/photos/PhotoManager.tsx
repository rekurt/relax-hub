import { useState, useCallback, useRef } from 'react'
import {
  App,
  Button,
  Card,
  Form,
  Image,
  Input,
  Modal,
  Popconfirm,
  Space,
  Spin,
  Tag,
  Tooltip,
  Typography,
} from '@/components/design/system'
import {
  DeleteOutlined,
  DragOutlined,
  PlusOutlined,
  ReloadOutlined,
} from '@/components/design/icons'
import {
  useGetMyBathhousesIdPhotos,
  usePostMyBathhousesIdPhotos,
  usePutMyBathhousesIdPhotosReorder,
  useDeletePhotosId,
} from '@/api/generated/photos/photos'
import type { InternalHandlerPhotoResponse } from '@/api/generated/model'
import { useBathhouseStore } from '@/stores/bathhouse'
import { useQueryClient } from '@tanstack/react-query'
import dayjs from 'dayjs'
import { resolveAssetUrl } from '@/lib/asset-url'
import EmptyState from '@/components/EmptyState'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

const STATUS_CONFIG: Record<string, { label: string; color: string }> = {
  pending: { label: 'На модерации', color: 'processing' },
  verified: { label: 'Подтверждено', color: 'success' },
  rejected: { label: 'Отклонено', color: 'error' },
}

interface UploadFormValues {
  url: string
  thumbnail_url?: string
}

export default function PhotoManager() {
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)
  const queryClient = useQueryClient()
  const { message } = App.useApp()
  const [form] = Form.useForm<UploadFormValues>()

  const [uploadModalOpen, setUploadModalOpen] = useState(false)
  const [dragIndex, setDragIndex] = useState<number | null>(null)
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null)
  const dragCounter = useRef(0)

  const { data, isLoading } = useGetMyBathhousesIdPhotos(
    selectedBathhouseId ?? '',
    { query: { enabled: !!selectedBathhouseId } },
  )

  const photos: InternalHandlerPhotoResponse[] = data?.data ?? []

  const invalidatePhotos = useCallback(() => {
    queryClient.invalidateQueries({
      queryKey: [`/my/bathhouses/${selectedBathhouseId}/photos`],
    })
  }, [queryClient, selectedBathhouseId])

  const uploadMutation = usePostMyBathhousesIdPhotos({
    mutation: {
      onSuccess: () => {
        message.success('Фото загружено и отправлено на модерацию')
        setUploadModalOpen(false)
        form.resetFields()
        invalidatePhotos()
      },
      onError: () => message.error('Не удалось загрузить фото'),
    },
  })

  const reorderMutation = usePutMyBathhousesIdPhotosReorder({
    mutation: {
      onSuccess: () => {
        message.success('Порядок обновлён')
        invalidatePhotos()
      },
      onError: () => message.error('Не удалось обновить порядок'),
    },
  })

  const deleteMutation = useDeletePhotosId({
    mutation: {
      onSuccess: () => {
        message.success('Фото удалено')
        invalidatePhotos()
      },
      onError: () => message.error('Не удалось удалить фото'),
    },
  })

  const handleUpload = (values: UploadFormValues) => {
    if (!selectedBathhouseId) return
    uploadMutation.mutate({
      id: selectedBathhouseId,
      data: { url: values.url, thumbnail_url: values.thumbnail_url || undefined },
    })
  }

  const handleDelete = (photoId: string) => {
    deleteMutation.mutate({ id: photoId })
  }

  const handleDragStart = (index: number) => {
    setDragIndex(index)
  }

  const handleDragEnter = (index: number) => {
    dragCounter.current++
    setDragOverIndex(index)
  }

  const handleDragLeave = () => {
    dragCounter.current--
    if (dragCounter.current === 0) {
      setDragOverIndex(null)
    }
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
  }

  const handleDrop = (targetIndex: number) => {
    dragCounter.current = 0
    if (dragIndex === null || dragIndex === targetIndex || !selectedBathhouseId) {
      setDragIndex(null)
      setDragOverIndex(null)
      return
    }

    const reordered = [...photos]
    const moved = reordered.splice(dragIndex, 1)[0]
    if (!moved) return
    reordered.splice(targetIndex, 0, moved)

    const photoIds = reordered.map((p) => p.id).filter(Boolean) as string[]
    reorderMutation.mutate({
      id: selectedBathhouseId,
      data: { photo_ids: photoIds },
    })

    setDragIndex(null)
    setDragOverIndex(null)
  }

  const handleDragEnd = () => {
    dragCounter.current = 0
    setDragIndex(null)
    setDragOverIndex(null)
  }

  const getStatusBadge = (status?: string) => {
    const config = STATUS_CONFIG[status ?? ''] ?? { label: status ?? 'Неизвестно', color: 'default' }
    return <Tag color={config.color}>{config.label}</Tag>
  }

  if (!selectedBathhouseId) {
    return (
      <div className="rh-page-stack">
        <PageHeader
          title="Фотографии"
          description="Галерея объекта, порядок показа и статусы модерации."
          size="compact"
        />
        <Card className="rh-admin-detail-card">
          <EmptyState description="Выберите баню для управления фотографиями" />
        </Card>
      </div>
    )
  }

  return (
    <div className="rh-page-stack">
      <PageHeader
        title="Фотографии"
        description="Управляйте изображениями, статусами модерации и порядком показа в карточке объекта."
        size="compact"
        extra={
          <Space wrap>
          <Button icon={<ReloadOutlined />} onClick={invalidatePhotos}>
            Обновить
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); setUploadModalOpen(true) }}>
            Добавить фото
          </Button>
          </Space>
        }
      />

      {isLoading ? (
        <div className="rh-loading-block rh-loading-block--large">
          <Spin size="large" />
        </div>
      ) : photos.length === 0 ? (
        <Card className="rh-admin-detail-card">
          <EmptyState
          description="Нет фотографий"
          actionText="Загрузить первое фото"
          onAction={() => { form.resetFields(); setUploadModalOpen(true) }}
        />
        </Card>
      ) : (
        <div className="rh-photo-grid">
          {photos.map((photo, index) => (
            <Card
              key={photo.id}
              size="small"
              className={[
                'rh-photo-card',
                dragIndex === index ? 'rh-photo-card--dragging' : '',
                dragOverIndex === index && dragIndex !== index ? 'rh-photo-card--drop-target' : '',
              ].filter(Boolean).join(' ')}
              draggable
              onDragStart={() => handleDragStart(index)}
              onDragEnter={() => handleDragEnter(index)}
              onDragLeave={handleDragLeave}
              onDragOver={handleDragOver}
              onDrop={() => handleDrop(index)}
              onDragEnd={handleDragEnd}
              cover={
                <div className="rh-photo-image-frame">
                  <Image
                    src={resolveAssetUrl(photo.thumbnail_url || photo.url)}
                    alt={`Фото ${index + 1}`}
                    className="rh-photo-image"
                    preview={{ src: resolveAssetUrl(photo.url) }}
                    fallback="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgZmlsbD0iI2YwZjBmMCIvPjx0ZXh0IHg9IjUwJSIgeT0iNTAlIiBkb21pbmFudC1iYXNlbGluZT0ibWlkZGxlIiB0ZXh0LWFuY2hvcj0ibWlkZGxlIiBmaWxsPSIjYmZiZmJmIiBmb250LXNpemU9IjE0Ij7QpNC+0YLQvjwvdGV4dD48L3N2Zz4="
                  />
                  <span className="rh-photo-order-badge">{index + 1}</span>
                </div>
              }
              actions={[
                <Tooltip title="Перетащите для сортировки" key="drag">
                  <DragOutlined />
                </Tooltip>,
                <Popconfirm
                  key="delete"
                  title="Удалить фото?"
                  description="Это действие нельзя отменить"
                  onConfirm={() => photo.id && handleDelete(photo.id)}
                  okText="Удалить"
                  cancelText="Отмена"
                >
                  <DeleteOutlined className="rh-danger-icon" />
                </Popconfirm>,
              ]}
            >
              <div className="rh-photo-card-footer">
                {getStatusBadge(photo.status)}
                <Text type="secondary" className="rh-table-meta-text">
                  {photo.uploaded_at ? dayjs(photo.uploaded_at).format('DD.MM.YYYY') : ''}
                </Text>
              </div>
              {photo.status === 'rejected' && photo.rejection_reason && (
                <Text type="danger" className="rh-photo-rejection-text">
                  Причина: {photo.rejection_reason}
                </Text>
              )}
            </Card>
          ))}
        </div>
      )}

      <Modal
        title="Добавить фото"
        open={uploadModalOpen}
        onCancel={() => setUploadModalOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleUpload}
        >
          <Form.Item
            name="url"
            label="URL фотографии"
            rules={[
              { required: true, message: 'Укажите URL фотографии' },
              { type: 'url', message: 'Введите корректный URL' },
            ]}
          >
            <Input placeholder="https://example.com/photo.jpg" />
          </Form.Item>

          <Form.Item
            name="thumbnail_url"
            label="URL превью (необязательно)"
            rules={[
              { type: 'url', message: 'Введите корректный URL' },
            ]}
            extra="Уменьшенная версия фото для быстрой загрузки"
          >
            <Input placeholder="https://example.com/photo-thumb.jpg" />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={uploadMutation.isPending}
              >
                Загрузить
              </Button>
              <Button onClick={() => setUploadModalOpen(false)}>
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
