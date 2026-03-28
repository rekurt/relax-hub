import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Typography,
  Card,
  Descriptions,
  Tag,
  Button,
  Spin,
  Empty,
  Alert,
  Divider,
  Space,
  Popconfirm,
  App,
  Modal,
  DatePicker,
  InputNumber,
  Form,
} from 'antd'
import {
  ArrowLeftOutlined,
  StopOutlined,
  DollarOutlined,
  StarOutlined,
  EditOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import { useQueryClient } from '@tanstack/react-query'
import { useGetBookings, usePatchBookingsIdCancel } from '@/api/generated/bookings/bookings'
import { useGetBookingsIdPayment, usePostBookingsIdPay } from '@/api/generated/payments/payments'
import { formatPrice, formatDateTime } from '@/lib/format'
import { BOOKING_STATUS_CONFIG, PAYMENT_STATUS_CONFIG } from '@/lib/constants'
import { axiosInstance } from '@/api/axios-instance'

const { Title, Text } = Typography

export default function ClientBookingDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [modifyModalOpen, setModifyModalOpen] = useState(false)
  const [modifyLoading, setModifyLoading] = useState(false)
  const [form] = Form.useForm()

  // Fetch user's bookings and find the one we need
  const { data: bookingsData, isLoading } = useGetBookings(
    { page: 1, page_size: 999 },
  )
  const booking = (bookingsData?.data ?? []).find((b) => b.id === id)

  const { data: paymentData, isLoading: paymentLoading } = useGetBookingsIdPayment(
    id ?? '',
    { query: { enabled: !!id } },
  )
  const payment = paymentData?.data

  const cancelMutation = usePatchBookingsIdCancel({
    mutation: {
      onSuccess: () => {
        message.success('Бронирование отменено')
        queryClient.invalidateQueries({ queryKey: ['/bookings'] })
      },
      onError: () => message.error('Не удалось отменить бронирование'),
    },
  })

  const payMutation = usePostBookingsIdPay({
    mutation: {
      onSuccess: (response) => {
        const confirmationUrl = response?.data?.confirmation_url
        if (confirmationUrl) {
          window.location.href = confirmationUrl
        }
      },
      onError: () => message.error('Не удалось инициировать оплату'),
    },
  })

  const handleModify = async (values: { startTime: dayjs.Dayjs; endTime: dayjs.Dayjs; guestCount: number }) => {
    if (!id) return
    setModifyLoading(true)
    try {
      const resp = await axiosInstance.put(`/bookings/${id}/modify`, {
        start_time: values.startTime.toISOString(),
        end_time: values.endTime.toISOString(),
        guest_count: values.guestCount,
      })
      const result = resp.data?.data
      if (result?.price_diff > 0) {
        message.info(`Бронирование изменено. Доплата: ${formatPrice(result.price_diff)}`)
      } else if (result?.price_diff < 0) {
        message.success(`Бронирование изменено. Возврат: ${formatPrice(-result.price_diff)}`)
      } else {
        message.success('Бронирование изменено')
      }
      setModifyModalOpen(false)
      queryClient.invalidateQueries({ queryKey: ['/bookings'] })
    } catch (err: unknown) {
      const errorMsg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      message.error(errorMsg || 'Не удалось изменить бронирование')
    } finally {
      setModifyLoading(false)
    }
  }

  if (isLoading) {
    return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  }

  if (!booking) {
    return <Empty description="Бронирование не найдено" />
  }

  const statusConfig = BOOKING_STATUS_CONFIG[booking.status ?? ''] ?? { color: 'default', text: booking.status }
  const paymentStatusConfig = PAYMENT_STATUS_CONFIG[payment?.status ?? ''] ?? {
    color: 'default',
    text: payment?.status ?? '—',
  }

  const canCancel = booking.status === 'pending' || booking.status === 'confirmed'
  const canPay = booking.status === 'confirmed' && (!payment || payment.status === 'pending' || !payment.status)
  const canReview = booking.status === 'completed'
  const canDispute = booking.status === 'completed' || booking.status === 'no_show'
  const canModify =
    (booking.status === 'pending' || booking.status === 'confirmed' || booking.status === 'pending_owner') &&
    (booking.modification_count ?? 0) < 3 &&
    booking.start_time &&
    dayjs(booking.start_time).isAfter(dayjs())

  const getRefundInfo = () => {
    if (!booking.start_time) return ''
    const hoursUntil = dayjs(booking.start_time).diff(dayjs(), 'hour')
    if (hoursUntil > 24) return 'При отмене сейчас вы получите 100% возврат.'
    if (hoursUntil >= 2) return 'При отмене сейчас вы получите 50% возврат.'
    return 'При отмене менее чем за 2 часа возврат не предусмотрен.'
  }

  const openModifyModal = () => {
    form.setFieldsValue({
      startTime: booking.start_time ? dayjs(booking.start_time) : undefined,
      endTime: booking.end_time ? dayjs(booking.end_time) : undefined,
      guestCount: booking.guest_count ?? 1,
    })
    setModifyModalOpen(true)
  }

  return (
    <div style={{ maxWidth: 700, margin: '0 auto' }}>
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/client/bookings')}
        style={{ marginBottom: 16 }}
      >
        К бронированиям
      </Button>

      <Title level={3}>Детали бронирования</Title>

      <Card style={{ marginBottom: 16 }}>
        <Descriptions column={1} bordered size="small">
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
            <Text strong style={{ fontSize: 16 }}>{formatPrice(booking.total_price ?? 0)}</Text>
          </Descriptions.Item>
          {(booking.earned_points ?? 0) > 0 && (
            <Descriptions.Item label="Начислено баллов">
              +{booking.earned_points}
            </Descriptions.Item>
          )}
          {(booking.modification_count ?? 0) > 0 && (
            <Descriptions.Item label="Изменений">
              {booking.modification_count} / 3
            </Descriptions.Item>
          )}
          <Descriptions.Item label="Создано">
            {booking.created_at ? formatDateTime(booking.created_at) : '—'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* Payment section */}
      <Card title="Платёж" style={{ marginBottom: 16 }}>
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
      </Card>

      {/* Refund policy */}
      {canCancel && (
        <Alert
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
          message="Политика отмены"
          description={getRefundInfo()}
        />
      )}

      <Divider />

      <Space>
        {canModify && (
          <Button
            size="large"
            icon={<EditOutlined />}
            onClick={openModifyModal}
          >
            Изменить бронирование
          </Button>
        )}
        {canReview && booking.bathhouse_id && (
          <Button
            type="primary"
            size="large"
            icon={<StarOutlined />}
            onClick={() => navigate(`/client/review?bathhouse=${booking.bathhouse_id}&booking=${booking.id}`)}
          >
            Оставить отзыв
          </Button>
        )}
        {canDispute && (
          <Button
            size="large"
            icon={<ExclamationCircleOutlined />}
            danger
            onClick={() => navigate(`/client/disputes/new?booking=${booking.id}`)}
          >
            Открыть спор
          </Button>
        )}
        {canPay && (
          <Button
            type="primary"
            size="large"
            icon={<DollarOutlined />}
            onClick={() => id && payMutation.mutate({ id, data: {} })}
            loading={payMutation.isPending}
          >
            Оплатить
          </Button>
        )}
        {canCancel && (
          <Popconfirm
            title="Отменить бронирование?"
            description={getRefundInfo()}
            onConfirm={() => id && cancelMutation.mutate({ id, data: {} })}
            okText="Отменить"
            cancelText="Нет"
            okButtonProps={{ danger: true }}
          >
            <Button
              size="large"
              danger
              icon={<StopOutlined />}
              loading={cancelMutation.isPending}
            >
              Отменить бронирование
            </Button>
          </Popconfirm>
        )}
      </Space>

      {/* Modify booking modal */}
      <Modal
        title="Изменить бронирование"
        open={modifyModalOpen}
        onCancel={() => setModifyModalOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
          message={`Осталось изменений: ${3 - (booking.modification_count ?? 0)}`}
          description="Цена будет пересчитана автоматически. При увеличении стоимости потребуется доплата, при уменьшении — разница будет возвращена."
        />
        <Form
          form={form}
          layout="vertical"
          onFinish={handleModify}
        >
          <Form.Item
            name="startTime"
            label="Начало"
            rules={[{ required: true, message: 'Выберите время начала' }]}
          >
            <DatePicker
              showTime={{ format: 'HH:mm' }}
              format="DD.MM.YYYY HH:mm"
              style={{ width: '100%' }}
            />
          </Form.Item>
          <Form.Item
            name="endTime"
            label="Окончание"
            rules={[{ required: true, message: 'Выберите время окончания' }]}
          >
            <DatePicker
              showTime={{ format: 'HH:mm' }}
              format="DD.MM.YYYY HH:mm"
              style={{ width: '100%' }}
            />
          </Form.Item>
          <Form.Item
            name="guestCount"
            label="Количество гостей"
            rules={[{ required: true, message: 'Укажите количество гостей' }]}
          >
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={modifyLoading}>
                Сохранить изменения
              </Button>
              <Button onClick={() => setModifyModalOpen(false)}>
                Отмена
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
