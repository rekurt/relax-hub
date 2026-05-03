import { useState, useCallback, useMemo } from 'react'
import { App, Button, Card, Descriptions, Empty, Input, Modal, Space, Spin, Tag, Typography } from '@/components/design/system'
import { CheckOutlined, CloseOutlined, ClockCircleOutlined } from '@/components/design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { axiosInstance } from '@/api/axios-instance'
import { formatPrice, formatDateTime } from '@/lib/format'

const { Text } = Typography
const { TextArea } = Input

interface ExtensionRequest {
  id: string
  booking_id: string
  user_id: string
  bathhouse_id: string
  status: string
  extra_hours: number
  extension_price: number
  new_end_time: string
  rejection_reason?: string
  created_at: string
  expires_at: string
  resolved_at?: string
}

interface ExtensionRequestsProps {
  bookingId: string
  open: boolean
  onClose: () => void
}

const STATUS_MAP: Record<string, { color: string; text: string }> = {
  pending: { color: 'processing', text: 'Ожидает' },
  approved: { color: 'success', text: 'Одобрено' },
  rejected: { color: 'error', text: 'Отклонено' },
  expired: { color: 'default', text: 'Истекло' },
}

function formatMinutesLeft(expiresAt: string): string {
  const diff = new Date(expiresAt).getTime() - Date.now()
  if (diff <= 0) return 'истекло'
  const minutes = Math.ceil(diff / 60000)
  return `${minutes} мин`
}

export default function ExtensionRequests({ bookingId, open, onClose }: ExtensionRequestsProps) {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [rejectingId, setRejectingId] = useState<string | null>(null)
  const [rejectReason, setRejectReason] = useState('')

  const queryKey = useMemo(() => ['extension-requests', bookingId], [bookingId])

  const { data, isLoading } = useQuery({
    queryKey,
    queryFn: async () => {
      const res = await axiosInstance.get<{ success: boolean; data: ExtensionRequest[] }>(
        `/bookings/${bookingId}/extension-requests`
      )
      return res.data.data ?? []
    },
    enabled: open && !!bookingId,
  })

  const invalidate = useCallback(() => {
    queryClient.invalidateQueries({ queryKey })
  }, [queryClient, queryKey])

  const approveMutation = useMutation({
    mutationFn: async (requestId: string) => {
      await axiosInstance.patch(`/bookings/extension-requests/${requestId}/approve`)
    },
    onSuccess: () => {
      message.success('Продление одобрено')
      invalidate()
    },
    onError: () => message.error('Не удалось одобрить продление'),
  })

  const rejectMutation = useMutation({
    mutationFn: async ({ requestId, reason }: { requestId: string; reason: string }) => {
      await axiosInstance.patch(`/bookings/extension-requests/${requestId}/reject`, { reason })
    },
    onSuccess: () => {
      message.success('Продление отклонено')
      setRejectingId(null)
      setRejectReason('')
      invalidate()
    },
    onError: () => message.error('Не удалось отклонить продление'),
  })

  const requests = data ?? []
  const pendingRequests = requests.filter((r) => r.status === 'pending')

  return (
    <>
      <Modal
        title="Запросы на продление"
        open={open}
        onCancel={onClose}
        footer={null}
        width={600}
      >
        {isLoading ? (
          <Spin />
        ) : requests.length === 0 ? (
          <Empty description="Нет запросов на продление" />
        ) : (
          <Space orientation="vertical" style={{ width: '100%' }} size="middle">
            {requests.map((req) => {
              const statusConfig = STATUS_MAP[req.status] ?? { color: 'default', text: req.status }
              const isPending = req.status === 'pending'

              return (
                <Card key={req.id} size="small">
                  <Space orientation="vertical" style={{ width: '100%' }}>
                    <Space>
                      <Tag color={statusConfig.color}>{statusConfig.text}</Tag>
                      <Text type="secondary">
                        Создан: {formatDateTime(req.created_at)}
                      </Text>
                      {isPending && (
                        <Tag icon={<ClockCircleOutlined />} color="warning">
                          Осталось: {formatMinutesLeft(req.expires_at)}
                        </Tag>
                      )}
                    </Space>

                    <Descriptions column={1} bordered size="small">
                      <Descriptions.Item label="Доп. часов">
                        +{req.extra_hours} ч
                      </Descriptions.Item>
                      <Descriptions.Item label="Стоимость продления">
                        {formatPrice(req.extension_price)}
                      </Descriptions.Item>
                      <Descriptions.Item label="Новое время окончания">
                        {formatDateTime(req.new_end_time, 'DD.MM.YYYY HH:mm')}
                      </Descriptions.Item>
                    </Descriptions>

                    {req.rejection_reason && (
                      <Text type="danger">Причина отклонения: {req.rejection_reason}</Text>
                    )}

                    {isPending && (
                      <Space>
                        <Button
                          type="primary"
                          icon={<CheckOutlined />}
                          loading={approveMutation.isPending}
                          onClick={() => approveMutation.mutate(req.id)}
                        >
                          Одобрить
                        </Button>
                        <Button
                          danger
                          icon={<CloseOutlined />}
                          onClick={() => setRejectingId(req.id)}
                        >
                          Отклонить
                        </Button>
                      </Space>
                    )}
                  </Space>
                </Card>
              )
            })}
          </Space>
        )}

        {pendingRequests.length > 0 && (
          <div style={{ marginTop: 16, padding: 8, background: 'rgba(180, 35, 24, 0.08)', borderRadius: 12 }}>
            <Text type="danger">
              {pendingRequests.length} запрос(ов) ожидают вашего решения. Без ответа запрос автоматически отклоняется через 30 минут.
            </Text>
          </div>
        )}
      </Modal>

      <Modal
        title="Причина отклонения"
        open={!!rejectingId}
        onCancel={() => { setRejectingId(null); setRejectReason('') }}
        onOk={() => rejectingId && rejectMutation.mutate({ requestId: rejectingId, reason: rejectReason })}
        okText="Отклонить"
        cancelText="Отмена"
        okButtonProps={{ danger: true, loading: rejectMutation.isPending }}
      >
        <TextArea
          value={rejectReason}
          onChange={(e) => setRejectReason(e.target.value)}
          placeholder="Укажите причину отклонения (необязательно)"
          rows={3}
        />
      </Modal>
    </>
  )
}

export { type ExtensionRequest }
