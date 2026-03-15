import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Col,
  Empty,
  Image,
  Input,
  Modal,
  Pagination,
  Row,
  Space,
  Spin,
  Tag,
  Typography,
} from 'antd'
import {
  CheckOutlined,
  CloseOutlined,
} from '@ant-design/icons'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetAdminPhotosPending,
  getGetAdminPhotosPendingQueryKey,
  usePatchAdminPhotosIdVerify,
  usePatchAdminPhotosIdReject,
} from '@/api/generated/admin-photos/admin-photos'
import type { InternalHandlerPhotoResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'

const { Title, Text } = Typography

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
    await rejectMutation.mutateAsync({
      id: rejectPhotoId,
      data: { reason: rejectReason || undefined },
    })
    message.success('Фото отклонено')
    setRejectModalOpen(false)
    invalidate()
  }

  return (
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>
        Верификация фото
      </Title>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: 48 }}>
          <Spin size="large" />
        </div>
      ) : photos.length === 0 ? (
        <Empty description="Нет фото на рассмотрении" />
      ) : (
        <>
          <Text type="secondary" style={{ display: 'block', marginBottom: 16 }}>
            Ожидают проверки: {meta?.total_count ?? photos.length}
          </Text>

          <Image.PreviewGroup>
            <Row gutter={[16, 16]}>
              {photos.map((photo) => (
                <Col key={photo.id} xs={24} sm={12} md={8} lg={6}>
                  <Card
                    hoverable
                    cover={
                      <Image
                        src={photo.url}
                        alt="Фото бани"
                        height={200}
                        style={{ objectFit: 'cover' }}
                        fallback="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgdmlld0JveD0iMCAwIDIwMCAyMDAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHJlY3Qgd2lkdGg9IjIwMCIgaGVpZ2h0PSIyMDAiIGZpbGw9IiNmMGYwZjAiLz48dGV4dCB4PSIxMDAiIHk9IjEwMCIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iIGZpbGw9IiM5OTkiIGZvbnQtc2l6ZT0iMTQiPk5vIEltYWdlPC90ZXh0Pjwvc3ZnPg=="
                      />
                    }
                    actions={[
                      <Button
                        key="verify"
                        type="text"
                        icon={<CheckOutlined />}
                        style={{ color: '#52c41a' }}
                        onClick={(e) => {
                          e.stopPropagation()
                          handleVerify(photo)
                        }}
                      >
                        Подтвердить
                      </Button>,
                      <Button
                        key="reject"
                        type="text"
                        icon={<CloseOutlined />}
                        danger
                        onClick={(e) => {
                          e.stopPropagation()
                          openRejectModal(photo.id!)
                        }}
                      >
                        Отклонить
                      </Button>,
                    ]}
                  >
                    <Card.Meta
                      title={
                        <Space>
                          <Tag color="orange">Ожидает</Tag>
                          {photo.position !== undefined && (
                            <Text type="secondary">#{photo.position + 1}</Text>
                          )}
                        </Space>
                      }
                      description={
                        <Space direction="vertical" size={4} style={{ width: '100%' }}>
                          <Text type="secondary" ellipsis>
                            Баня: {photo.bathhouse_id?.slice(0, 8)}...
                          </Text>
                          {photo.uploaded_at && (
                            <Text type="secondary" style={{ fontSize: 12 }}>
                              {formatDateTime(photo.uploaded_at)}
                            </Text>
                          )}
                        </Space>
                      }
                    />
                  </Card>
                </Col>
              ))}
            </Row>
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
        </>
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
