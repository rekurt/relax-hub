import { useState, useCallback, useMemo } from 'react'
import { App, Button, Card, Descriptions, Empty, Input, Modal, Space, Spin, Tag, Typography } from 'antd'
import { CheckOutlined, CloseOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { axiosInstance } from '@/api/axios-instance'
import { formatPrice, formatDateTime } from '@/lib/format'

const { Text } = Typography
const { TextArea } = Input

interface ModificationRequest {
  id: string
  booking_id: string
  user_id: string
  bathhouse_id: string
  status: string
  old_start_time: string
  old_end_time: string
  old_guest_count: number
  old_total_price: number
  proposed_start_time: string
  proposed_end_time: string
  proposed_guest_count: number
  proposed_total_price: number
  rejection_reason?: string
  created_at: string
  expires_at: string
  resolved_at?: string
}

interface ModificationRequestsProps {
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

export default function ModificationRequests({ bookingId, open, onClose }: ModificationRequestsProps) {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [rejectingId, setRejectingId] = useState<string | null>(null)
  const [rejectReason, setRejectReason] = useState('')

  const queryKey = useMemo(() => ['modification-requests', bookingId], [bookingId])

  const { data, isLoading } = useQuery({
    queryKey,
    queryFn: async () => {
      const res = await axiosInstance.get<{ success: boolean; data: ModificationRequest[] }>(
        `/bookings/${bookingId}/modification-requests`
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
      await axiosInstance.patch(`/bookings/modification-requests/${requestId}/approve`)
    },
    onSuccess: () => {
      message.success('Изменение одобрено')
      invalidate()
    },
    onError: () => message.error('Не удалось одобрить изменение'),
  })

  const rejectMutation = useMutation({
    mutationFn: async ({ requestId, reason }: { requestId: string; reason: string }) => {
      await axiosInstance.patch(`/bookings/modification-requests/${requestId}/reject`, { reason })
    },
    onSuccess: () => {
      message.success('Изменение отклонено')
      setRejectingId(null)
      setRejectReason('')
      invalidate()
    },
    onError: () => message.error('Не удалось отклонить изменение'),
  })

  const requests = data ?? []
  const pendingRequests = requests.filter((r) => r.status === 'pending')

  return (
    <>
      <Modal
        title="Запросы на изменение"
        open={open}
        onCancel={onClose}
        footer={null}
        width={700}
      >
        {isLoading ? (
          <Spin />
        ) : requests.length === 0 ? (
          <Empty description="Нет запросов на изменение" />
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
                        <Text type="warning">
                          Истекает: {formatDateTime(req.expires_at)}
                        </Text>
                      )}
                    </Space>

                    <Descriptions column={2} bordered size="small">
                      <Descriptions.Item label="Текущее время">
                        {formatDateTime(req.old_start_time, 'DD.MM HH:mm')} – {formatDateTime(req.old_end_time, 'HH:mm')}
                      </Descriptions.Item>
                      <Descriptions.Item label="Предложенное время">
                        {formatDateTime(req.proposed_start_time, 'DD.MM HH:mm')} – {formatDateTime(req.proposed_end_time, 'HH:mm')}
                      </Descriptions.Item>
                      <Descriptions.Item label="Текущие гости">
                        {req.old_guest_count}
                      </Descriptions.Item>
                      <Descriptions.Item label="Предложенные гости">
                        {req.proposed_guest_count}
                      </Descriptions.Item>
                      <Descriptions.Item label="Текущая цена">
                        {formatPrice(req.old_total_price)}
                      </Descriptions.Item>
                      <Descriptions.Item label="Предложенная цена">
                        {formatPrice(req.proposed_total_price)}
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
          <div style={{ marginTop: 16, padding: 8, background: '#fffbe6', borderRadius: 4 }}>
            <Text type="warning">
              {pendingRequests.length} запрос(ов) ожидают вашего решения. Без ответа запрос автоматически отклоняется через 24 часа.
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

export { type ModificationRequest }
