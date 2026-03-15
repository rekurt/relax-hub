import { Descriptions, Modal, Spin, Tag } from 'antd'
import { useGetBookingsIdPayment } from '@/api/generated/payments/payments'
import type { InternalHandlerBookingResponse } from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'

const STATUS_CONFIG: Record<string, { color: string; text: string }> = {
  pending: { color: 'orange', text: 'Ожидает' },
  confirmed: { color: 'blue', text: 'Подтверждено' },
  completed: { color: 'green', text: 'Завершено' },
  cancelled: { color: 'default', text: 'Отменено' },
  rejected: { color: 'red', text: 'Отклонено' },
}

const PAYMENT_STATUS_CONFIG: Record<string, { color: string; text: string }> = {
  pending: { color: 'orange', text: 'Ожидает оплаты' },
  succeeded: { color: 'green', text: 'Оплачено' },
  canceled: { color: 'default', text: 'Отменён' },
  refunded: { color: 'purple', text: 'Возвращён' },
}

interface BookingDetailsProps {
  booking: InternalHandlerBookingResponse | null
  onClose: () => void
}

export default function BookingDetails({ booking, onClose }: BookingDetailsProps) {
  const { data: paymentData, isLoading: paymentLoading } = useGetBookingsIdPayment(
    booking?.id ?? '',
    { query: { enabled: !!booking?.id } },
  )

  const payment = paymentData?.data

  const statusConfig = STATUS_CONFIG[booking?.status ?? ''] ?? { color: 'default', text: booking?.status }
  const paymentStatusConfig = PAYMENT_STATUS_CONFIG[payment?.status ?? ''] ?? {
    color: 'default',
    text: payment?.status ?? '—',
  }

  return (
    <Modal
      title="Детали бронирования"
      open={!!booking}
      onCancel={onClose}
      footer={null}
      width={600}
    >
      {booking && (
        <>
          <Descriptions column={1} bordered size="small" style={{ marginBottom: 16 }}>
            <Descriptions.Item label="ID">
              {booking.id?.slice(0, 8)}...
            </Descriptions.Item>
            <Descriptions.Item label="Статус">
              <Tag color={statusConfig.color}>{statusConfig.text}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Начало">
              {booking.start_time ? formatDateTime(booking.start_time) : '—'}
            </Descriptions.Item>
            <Descriptions.Item label="Окончание">
              {booking.end_time ? formatDateTime(booking.end_time) : '—'}
            </Descriptions.Item>
            <Descriptions.Item label="Гостей">
              {booking.guest_count ?? '—'}
            </Descriptions.Item>
            <Descriptions.Item label="Комментарий">
              {booking.comment || '—'}
            </Descriptions.Item>
            <Descriptions.Item label="Исходная цена">
              {formatPrice(booking.original_price ?? 0)}
            </Descriptions.Item>
            {(booking.promo_discount ?? 0) > 0 && (
              <Descriptions.Item label="Скидка по промокоду">
                -{formatPrice(booking.promo_discount ?? 0)}
              </Descriptions.Item>
            )}
            {(booking.loyalty_discount ?? 0) > 0 && (
              <Descriptions.Item label="Скидка лояльности">
                -{formatPrice(booking.loyalty_discount ?? 0)}
              </Descriptions.Item>
            )}
            {(booking.certificate_discount ?? 0) > 0 && (
              <Descriptions.Item label="Скидка по сертификату">
                -{formatPrice(booking.certificate_discount ?? 0)}
              </Descriptions.Item>
            )}
            {(booking.points_spent ?? 0) > 0 && (
              <Descriptions.Item label="Баллы использованы">
                {booking.points_spent}
              </Descriptions.Item>
            )}
            {(booking.referral_bonus_used ?? 0) > 0 && (
              <Descriptions.Item label="Реферальный бонус">
                -{formatPrice(booking.referral_bonus_used ?? 0)}
              </Descriptions.Item>
            )}
            <Descriptions.Item label="Итого">
              <strong>{formatPrice(booking.total_price ?? 0)}</strong>
            </Descriptions.Item>
            {(booking.earned_points ?? 0) > 0 && (
              <Descriptions.Item label="Начислено баллов">
                +{booking.earned_points}
              </Descriptions.Item>
            )}
            <Descriptions.Item label="Создано">
              {booking.created_at ? formatDateTime(booking.created_at) : '—'}
            </Descriptions.Item>
          </Descriptions>

          <h4>Платёж</h4>
          {paymentLoading ? (
            <Spin size="small" />
          ) : payment ? (
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="ID платежа">
                {payment.id?.slice(0, 8)}...
              </Descriptions.Item>
              <Descriptions.Item label="Статус">
                <Tag color={paymentStatusConfig.color}>{paymentStatusConfig.text}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Сумма">
                {formatPrice(payment.amount ?? 0)}
              </Descriptions.Item>
              <Descriptions.Item label="Провайдер">
                {payment.provider ?? '—'}
              </Descriptions.Item>
              {(payment.refund_amount ?? 0) > 0 && (
                <Descriptions.Item label="Возвращено">
                  {formatPrice(payment.refund_amount ?? 0)}
                </Descriptions.Item>
              )}
              <Descriptions.Item label="Дата">
                {payment.created_at ? formatDateTime(payment.created_at) : '—'}
              </Descriptions.Item>
            </Descriptions>
          ) : (
            <div style={{ color: '#999' }}>Платёж не найден</div>
          )}
        </>
      )}
    </Modal>
  )
}
