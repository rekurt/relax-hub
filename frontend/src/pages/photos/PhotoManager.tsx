import { useState, useCallback, useRef } from 'react'
import {
  App,
  Badge,
  Button,
  Card,
  Empty,
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
} from 'antd'
import {
  DeleteOutlined,
  DragOutlined,
  PlusOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
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

const { Title, Text } = Typography

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
      <div>
        <Title level={3}>Фотографии</Title>
        <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
          Выберите баню для управления фотографиями
        </div>
      </div>
    )
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>Фотографии</Title>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={invalidatePhotos}>
            Обновить
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); setUploadModalOpen(true) }}>
            Добавить фото
          </Button>
        </Space>
      </div>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: 60 }}>
          <Spin size="large" />
        </div>
      ) : photos.length === 0 ? (
        <Empty
          description="Нет фотографий"
          style={{ padding: 60 }}
        >
          <Button type="primary" onClick={() => { form.resetFields(); setUploadModalOpen(true) }}>
            Загрузить первое фото
          </Button>
        </Empty>
      ) : (
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))',
            gap: 16,
          }}
        >
          {photos.map((photo, index) => (
            <Card
              key={photo.id}
              size="small"
              draggable
              onDragStart={() => handleDragStart(index)}
              onDragEnter={() => handleDragEnter(index)}
              onDragLeave={handleDragLeave}
              onDragOver={handleDragOver}
              onDrop={() => handleDrop(index)}
              onDragEnd={handleDragEnd}
              style={{
                cursor: 'grab',
                opacity: dragIndex === index ? 0.4 : 1,
                border: dragOverIndex === index && dragIndex !== index ? '2px dashed #1677ff' : undefined,
                transition: 'opacity 0.2s, border 0.2s',
              }}
              cover={
                <div style={{ position: 'relative' }}>
                  <Image
                    src={photo.thumbnail_url || photo.url}
                    alt={`Фото ${index + 1}`}
                    style={{ height: 160, objectFit: 'cover', width: '100%' }}
                    preview={{ src: photo.url }}
                    fallback="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgZmlsbD0iI2YwZjBmMCIvPjx0ZXh0IHg9IjUwJSIgeT0iNTAlIiBkb21pbmFudC1iYXNlbGluZT0ibWlkZGxlIiB0ZXh0LWFuY2hvcj0ibWlkZGxlIiBmaWxsPSIjYmZiZmJmIiBmb250LXNpemU9IjE0Ij7QpNC+0YLQvjwvdGV4dD48L3N2Zz4="
                  />
                  <Badge
                    count={index + 1}
                    style={{
                      position: 'absolute',
                      top: 8,
                      left: 8,
                      backgroundColor: 'rgba(0,0,0,0.5)',
                    }}
                  />
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
                  <DeleteOutlined style={{ color: '#ff4d4f' }} />
                </Popconfirm>,
              ]}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                {getStatusBadge(photo.status)}
                <Text type="secondary" style={{ fontSize: 12 }}>
                  {photo.uploaded_at ? dayjs(photo.uploaded_at).format('DD.MM.YYYY') : ''}
                </Text>
              </div>
              {photo.status === 'rejected' && photo.rejection_reason && (
                <Text type="danger" style={{ fontSize: 12, display: 'block', marginTop: 4 }}>
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
