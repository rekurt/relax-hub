import { useState } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import {
  Typography,
  Card,
  Button,
  DatePicker,
  InputNumber,
  Input,
  Space,
  Descriptions,
  Spin,
  Alert,
  Divider,
  Switch,
  App,
  Empty,
} from 'antd'
import {
  ArrowLeftOutlined,
  ClockCircleOutlined,
  TagOutlined,
  GiftOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import { useGetBathhousesId, useGetBathhousesIdAvailableSlots } from '@/api/generated/bathhouses/bathhouses'
import { useGetBathhousesIdPriceCalculator } from '@/api/generated/pricing/pricing'
import { usePostBookings } from '@/api/generated/bookings/bookings'
import { usePostPromoCodesValidate } from '@/api/generated/promo-codes/promo-codes'
import { useGetCertificatesCodeBalance } from '@/api/generated/certificates/certificates'
import { formatPrice } from '@/lib/format'

const { Title, Text } = Typography

export default function BookingCreate() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { message } = App.useApp()

  const bathhouseId = searchParams.get('bathhouse') ?? ''
  const initialDate = searchParams.get('date') ?? dayjs().format('YYYY-MM-DD')
  const initialFrom = searchParams.get('from') ?? ''
  const initialTo = searchParams.get('to') ?? ''

  const [selectedDate, setSelectedDate] = useState(initialDate)
  const [selectedSlot, setSelectedSlot] = useState<{ from: string; to: string } | null>(
    initialFrom && initialTo ? { from: initialFrom, to: initialTo } : null,
  )
  const [guestCount, setGuestCount] = useState(1)
  const [comment, setComment] = useState('')

  // Discounts
  const [promoCode, setPromoCode] = useState('')
  const [promoValidated, setPromoValidated] = useState<{ discount: number; type?: string } | null>(null)
  const [promoError, setPromoError] = useState('')

  const [certificateCode, setCertificateCode] = useState('')

  const [usePoints, setUsePoints] = useState(false)
  const [pointsAmount, setPointsAmount] = useState(0)
  const [useReferral, setUseReferral] = useState(false)
  const [referralAmount, setReferralAmount] = useState(0)

  const { data: bathhouseData, isLoading: bathhouseLoading } = useGetBathhousesId(bathhouseId, {
    query: { enabled: !!bathhouseId },
  })
  const bathhouse = bathhouseData?.data

  const { data: slotsData, isLoading: slotsLoading } = useGetBathhousesIdAvailableSlots(
    bathhouseId,
    { date: selectedDate },
    { query: { enabled: !!bathhouseId && !!selectedDate } },
  )
  const slots = slotsData?.data ?? []

  // Build RFC3339 timestamps for price calculator
  const startTime = selectedSlot ? `${selectedDate}T${selectedSlot.from}:00Z` : ''
  const endTime = selectedSlot ? `${selectedDate}T${selectedSlot.to}:00Z` : ''

  const { data: priceData, isLoading: priceLoading } = useGetBathhousesIdPriceCalculator(
    bathhouseId,
    { start: startTime, end: endTime },
    { query: { enabled: !!bathhouseId && !!selectedSlot } },
  )
  const priceInfo = priceData?.data

  const promoValidateMutation = usePostPromoCodesValidate({
    mutation: {
      onSuccess: (response) => {
        const data = response?.data
        if (data && data.discount != null) {
          setPromoValidated({ discount: data.discount, type: data.type })
          setPromoError('')
          message.success(`Промокод применён: скидка ${formatPrice(data.discount)}`)
        }
      },
      onError: () => {
        setPromoValidated(null)
        setPromoError('Недействительный промокод')
      },
    },
  })

  const handleValidatePromo = () => {
    if (!promoCode.trim()) return
    setPromoError('')
    promoValidateMutation.mutate({
      data: {
        code: promoCode.trim(),
        bathhouse_id: bathhouseId,
        amount: priceInfo?.final_price ?? priceInfo?.base_price ?? 0,
      },
    })
  }

  // Certificate balance check - enabled when code is entered
  const shouldCheckCertificate = certificateCode.trim().length >= 4
  const { data: certBalanceData } = useGetCertificatesCodeBalance(
    certificateCode.trim(),
    {
      query: {
        enabled: shouldCheckCertificate,
        retry: false,
      },
    },
  )

  // Derive certificate balance from query data
  const certificateBalance = certBalanceData?.data?.balance ?? null

  const createBookingMutation = usePostBookings({
    mutation: {
      onSuccess: (response) => {
        const booking = response?.data
        if (booking?.id) {
          message.success('Бронирование создано!')
          navigate(`/client/bookings/${booking.id}`)
        }
      },
      onError: (error) => {
        const errorMsg = (error as { error?: { message?: string } })?.error?.message ?? 'Не удалось создать бронирование'
        message.error(errorMsg)
      },
    },
  })

  const handleCreateBooking = () => {
    if (!selectedSlot || !bathhouseId) return

    createBookingMutation.mutate({
      data: {
        bathhouse_id: bathhouseId,
        start_time: startTime,
        end_time: endTime,
        guest_count: guestCount,
        comment: comment || undefined,
        promo_code: promoValidated ? promoCode.trim() : undefined,
        certificate_code: certificateBalance != null && certificateBalance > 0 ? certificateCode.trim() : undefined,
        use_points: usePoints ? pointsAmount : undefined,
        use_referral_bonus: useReferral ? referralAmount : undefined,
      },
    })
  }

  if (bathhouseLoading) {
    return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  }

  if (!bathhouse) {
    return <Empty description="Баня не найдена" />
  }

  return (
    <div style={{ maxWidth: 800, margin: '0 auto' }}>
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate(`/client/bathhouse/${bathhouseId}`)}
        style={{ marginBottom: 16 }}
      >
        Назад к бане
      </Button>

      <Title level={3}>Бронирование: {bathhouse.name}</Title>

      {/* Date and Slot Selection */}
      <Card title="Дата и время" style={{ marginBottom: 16 }}>
        <Space direction="vertical" style={{ width: '100%' }} size={16}>
          <div>
            <Text strong>Дата:</Text>
            <DatePicker
              value={dayjs(selectedDate)}
              onChange={(d) => {
                if (d) {
                  setSelectedDate(d.format('YYYY-MM-DD'))
                  setSelectedSlot(null)
                }
              }}
              disabledDate={(d) => d.isBefore(dayjs(), 'day')}
              style={{ width: '100%', marginTop: 8 }}
            />
          </div>

          <div>
            <Text strong>Доступные слоты:</Text>
            <Spin spinning={slotsLoading}>
              {slots.length === 0 ? (
                <Empty description="Нет доступных слотов" image={Empty.PRESENTED_IMAGE_SIMPLE} style={{ margin: '16px 0' }} />
              ) : (
                <Space wrap style={{ marginTop: 8 }}>
                  {slots.map((slot) => {
                    const isSelected = selectedSlot?.from === slot.startTime && selectedSlot?.to === slot.endTime
                    return (
                      <Button
                        key={`${slot.startTime}-${slot.endTime}`}
                        type={isSelected ? 'primary' : 'default'}
                        disabled={!slot.available}
                        icon={<ClockCircleOutlined />}
                        onClick={() => setSelectedSlot({ from: slot.startTime!, to: slot.endTime! })}
                      >
                        {slot.startTime} — {slot.endTime}
                        {slot.price != null && ` (${formatPrice(slot.price)})`}
                      </Button>
                    )
                  })}
                </Space>
              )}
            </Spin>
          </div>

          <div>
            <Text strong>Количество гостей:</Text>
            <InputNumber
              min={1}
              max={bathhouse.max_guests ?? 20}
              value={guestCount}
              onChange={(v) => setGuestCount(v ?? 1)}
              style={{ width: '100%', marginTop: 8 }}
            />
          </div>

          <div>
            <Text strong>Комментарий:</Text>
            <Input.TextArea
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              placeholder="Пожелания к бронированию"
              rows={2}
              style={{ marginTop: 8 }}
            />
          </div>
        </Space>
      </Card>

      {/* Discounts */}
      <Card title="Скидки и бонусы" style={{ marginBottom: 16 }}>
        <Space direction="vertical" style={{ width: '100%' }} size={16}>
          {/* Promo code */}
          <div>
            <Text strong><TagOutlined /> Промокод:</Text>
            <Space.Compact style={{ width: '100%', marginTop: 8 }}>
              <Input
                value={promoCode}
                onChange={(e) => {
                  setPromoCode(e.target.value)
                  setPromoValidated(null)
                  setPromoError('')
                }}
                placeholder="Введите промокод"
                status={promoError ? 'error' : promoValidated ? '' : undefined}
              />
              <Button
                onClick={handleValidatePromo}
                loading={promoValidateMutation.isPending}
                disabled={!promoCode.trim()}
              >
                Применить
              </Button>
            </Space.Compact>
            {promoError && <Text type="danger" style={{ fontSize: 12 }}>{promoError}</Text>}
            {promoValidated && (
              <Text type="success" style={{ fontSize: 12 }}>
                Скидка: {formatPrice(promoValidated.discount)}
              </Text>
            )}
          </div>

          {/* Gift certificate */}
          <div>
            <Text strong><GiftOutlined /> Подарочный сертификат:</Text>
            <Input
              value={certificateCode}
              onChange={(e) => setCertificateCode(e.target.value)}
              placeholder="BANI-XXXX-XXXX"
              style={{ marginTop: 8 }}
            />
            {certificateBalance != null && (
              <Text type="success" style={{ fontSize: 12 }}>
                Баланс сертификата: {formatPrice(certificateBalance)}
              </Text>
            )}
          </div>

          {/* Loyalty points */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <Switch checked={usePoints} onChange={setUsePoints} />
            <Text>Использовать баллы лояльности</Text>
            {usePoints && (
              <InputNumber
                min={0}
                value={pointsAmount}
                onChange={(v) => setPointsAmount(v ?? 0)}
                placeholder="Кол-во баллов"
                style={{ width: 150 }}
              />
            )}
          </div>

          {/* Referral balance */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <Switch checked={useReferral} onChange={setUseReferral} />
            <Text>Использовать реферальный бонус</Text>
            {useReferral && (
              <InputNumber
                min={0}
                value={referralAmount}
                onChange={(v) => setReferralAmount(v ?? 0)}
                placeholder="Сумма (коп.)"
                style={{ width: 150 }}
              />
            )}
          </div>
        </Space>
      </Card>

      {/* Price Summary */}
      <Card title="Итого" style={{ marginBottom: 16 }}>
        {priceLoading ? (
          <Spin />
        ) : priceInfo ? (
          <Descriptions column={1} size="small">
            <Descriptions.Item label="Базовая цена">
              {formatPrice(priceInfo.base_price ?? 0)}
            </Descriptions.Item>
            {priceInfo.hours != null && (
              <Descriptions.Item label="Часов">
                {priceInfo.hours}
              </Descriptions.Item>
            )}
            {(priceInfo.price_saving ?? 0) > 0 && (
              <Descriptions.Item label="Экономия (динамическое ценообразование)">
                <Text type="success">-{formatPrice(priceInfo.price_saving ?? 0)}</Text>
              </Descriptions.Item>
            )}
            {promoValidated && (
              <Descriptions.Item label="Скидка по промокоду">
                <Text type="success">-{formatPrice(promoValidated.discount)}</Text>
              </Descriptions.Item>
            )}
            {certificateBalance != null && certificateBalance > 0 && (
              <Descriptions.Item label="Сертификат">
                <Text type="success">до -{formatPrice(certificateBalance)}</Text>
              </Descriptions.Item>
            )}
            <Descriptions.Item label="Итого к оплате">
              <Text strong style={{ fontSize: 18 }}>
                {formatPrice(priceInfo.final_price ?? 0)}
              </Text>
            </Descriptions.Item>
          </Descriptions>
        ) : selectedSlot ? (
          <Text type="secondary">Загрузка цены...</Text>
        ) : (
          <Text type="secondary">Выберите слот для расчёта цены</Text>
        )}
      </Card>

      {/* Refund policy */}
      <Alert
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
        message="Политика отмены"
        description="Более 24ч до начала — 100% возврат. От 2 до 24ч — 50% возврат. Менее 2ч — без возврата."
      />

      <Divider />

      <Space>
        <Button
          type="primary"
          size="large"
          onClick={handleCreateBooking}
          loading={createBookingMutation.isPending}
          disabled={!selectedSlot || !bathhouseId}
        >
          Забронировать
        </Button>
        <Button size="large" onClick={() => navigate(`/client/bathhouse/${bathhouseId}`)}>
          Отмена
        </Button>
      </Space>
    </div>
  )
}
