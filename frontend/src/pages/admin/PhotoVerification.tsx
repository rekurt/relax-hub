import { useState } from 'react'
import {
  App,
  Button,
  Empty,
  Image,
  Input,
  Modal,
  Pagination,
  Space,
  Spin,
  Tag,
  Typography,
} from '@/components/design/system'
import {
  CheckOutlined,
  CloseOutlined,
} from '@/components/design/icons'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetAdminPhotosPending,
  getGetAdminPhotosPendingQueryKey,
  usePatchAdminPhotosIdVerify,
  usePatchAdminPhotosIdReject,
} from '@/api/generated/admin-photos/admin-photos'
import type { InternalHandlerPhotoResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

export default function PhotoVerification() {
  const { modal, message } = App.useApp()
  const queryClient = useQueryClient()

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(12)
  const [rejectModalOpen, setRejectModalOpen] = useState(false)
  const [rejectReason, setRejectReason] = useState('')
  const [rejectPhotoId, setRejectPhotoId] = useState<string>('')

  const { data, isLoading } = useGetAdminPhotosPending({ page, page_size: pageSize })
  const verifyMutation = usePatchAdminPhotosIdVerify()
  const rejectMutation = usePatchAdminPhotosIdReject()

  const photos: InternalHandlerPhotoResponse[] = data?.data ?? []
  const meta = data?.meta

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetAdminPhotosPendingQueryKey() })
  }

  const handleVerify = (photo: InternalHandlerPhotoResponse) => {
    modal.confirm({
      title: 'Подтвердить фото?',
      content: `Фото бани ${photo.bathhouse_id?.slice(0, 8)}... будет подтверждено.`,
      okText: 'Подтвердить',
      cancelText: 'Отмена',
      onOk: () =>
        verifyMutation.mutateAsync({ id: photo.id! }).then(() => {
          message.success('Фото подтверждено')
          invalidate()
        }),
    })
  }

  const openRejectModal = (photoId: string) => {
    setRejectPhotoId(photoId)
    setRejectReason('')
    setRejectModalOpen(true)
  }

  const handleRejectConfirm = async () => {
    if (!rejectPhotoId) return
    try {
      await rejectMutation.mutateAsync({
        id: rejectPhotoId,
        data: { reason: rejectReason || undefined },
      })
      message.success('Фото отклонено')
      setRejectModalOpen(false)
      invalidate()
    } catch {
      message.error('Не удалось отклонить фото')
    }
  }

  return (
    <div className="rh-stack">
      <PageHeader
        size="compact"
        eyebrow="Модерация"
        title="Верификация фото"
        description="Проверяйте кадры объектов рядом с позицией, датой загрузки и быстрым решением модератора."
      />

      {isLoading ? (
        <section className="rh-admin-panel" style={{ textAlign: 'center', padding: 48 }}>
          <Spin size="large" />
        </section>
      ) : photos.length === 0 ? (
        <section className="rh-admin-panel">
          <Empty description="Нет фото на рассмотрении" />
        </section>
      ) : (
        <section className="rh-admin-panel">
          <div className="rh-admin-toolbar" style={{ marginBottom: 18 }}>
            <div className="rh-admin-toolbar__copy">
              <h2 className="rh-admin-toolbar__title">Фото на проверке</h2>
              <div className="rh-admin-toolbar__hint">Ожидают проверки: {meta?.total_count ?? photos.length}</div>
            </div>
            <Tag color="orange">В очереди</Tag>
          </div>

          <Image.PreviewGroup>
            <div className="rh-admin-photo-grid">
              {photos.map((photo) => (
                <article className="rh-admin-photo-card" key={photo.id}>
                  <div className="rh-admin-photo-card__media">
                    <Image
                      src={photo.url}
                      alt="Фото бани"
                      height="100%"
                      width="100%"
                      style={{ objectFit: 'cover', display: 'block' }}
                      fallback="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgdmlld0JveD0iMCAwIDIwMCAyMDAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZyI+PHJlY3Qgd2lkdGg9IjIwMCIgaGVpZ2h0PSIyMDAiIGZpbGw9IiNmMGYwZjAiLz48dGV4dCB4PSIxMDAiIHk9IjEwMCIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iIGZpbGw9IiM5OTkiIGZvbnQtc2l6ZT0iMTQiPk5vIEltYWdlPC90ZXh0Pjwvc3ZnPg=="
                    />
                  </div>
                  <div className="rh-admin-photo-card__body">
                    <div className="rh-admin-photo-card__title">
                      <Space>
                        <Tag color="orange">Ожидает</Tag>
                        {photo.position !== undefined && (
                          <Text type="secondary">#{photo.position + 1}</Text>
                        )}
                      </Space>
                    </div>
                    <div className="rh-admin-photo-card__meta">
                      Баня: {photo.bathhouse_id?.slice(0, 8)}...
                    </div>
                    {photo.uploaded_at && (
                      <div className="rh-admin-photo-card__meta">
                        {formatDateTime(photo.uploaded_at)}
                      </div>
                    )}
                    <div className="rh-admin-photo-card__actions">
                      <Button
                        type="primary"
                        icon={<CheckOutlined />}
                        onClick={(e) => {
                          e.stopPropagation()
                          handleVerify(photo)
                        }}
                      >
                        Подтвердить
                      </Button>
                      <Button
                        icon={<CloseOutlined />}
                        danger
                        onClick={(e) => {
                          e.stopPropagation()
                          openRejectModal(photo.id!)
                        }}
                      >
                        Отклонить
                      </Button>
                    </div>
                  </div>
                </article>
              ))}
            </div>
          </Image.PreviewGroup>

          {meta && meta.total_pages! > 1 && (
            <div style={{ marginTop: 24, textAlign: 'center' }}>
              <Pagination
                current={page}
                pageSize={pageSize}
                total={meta.total_count}
                showSizeChanger
                pageSizeOptions={['12', '24', '48']}
                showTotal={(total) => `Всего: ${total}`}
                onChange={(p, ps) => {
                  setPage(p)
                  setPageSize(ps)
                }}
              />
            </div>
          )}
        </section>
      )}

      <Modal
        title="Отклонить фото"
        open={rejectModalOpen}
        onCancel={() => setRejectModalOpen(false)}
        onOk={handleRejectConfirm}
        okText="Отклонить"
        okType="danger"
        cancelText="Отмена"
        confirmLoading={rejectMutation.isPending}
      >
        <div style={{ marginBottom: 8 }}>
          <Text>Укажите причину отклонения (необязательно):</Text>
        </div>
        <Input.TextArea
          rows={3}
          placeholder="Причина отклонения"
          value={rejectReason}
          onChange={(e) => setRejectReason(e.target.value)}
        />
      </Modal>
    </div>
  )
}
