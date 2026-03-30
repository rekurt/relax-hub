import { useState, useMemo } from 'react'
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
  Steps,
  Slider,
  Radio,
  Checkbox,
  Tag,
  Result,
} from 'antd'
import {
  ArrowLeftOutlined,
  ClockCircleOutlined,
  TagOutlined,
  GiftOutlined,
  WalletOutlined,
  CreditCardOutlined,
  BankOutlined,
  AppleOutlined,
  GoogleOutlined,
  ShoppingCartOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import { useGetBathhousesId, useGetBathhousesIdAvailableSlots } from '@/api/generated/bathhouses/bathhouses'
import { useGetBathhousesIdPriceCalculator } from '@/api/generated/pricing/pricing'
import { usePostBookings } from '@/api/generated/bookings/bookings'
import { usePostPromoCodesValidate } from '@/api/generated/promo-codes/promo-codes'
import { useGetCertificatesCodeBalance } from '@/api/generated/certificates/certificates'
import { useGetBathhousesIdAddons } from '@/api/generated/add-ons/add-ons'
import { useGetMyWallet } from '@/api/generated/wallet/wallet'
import { formatPrice } from '@/lib/format'

const { Title, Text } = Typography

type PaymentMethod = 'wallet' | 'card' | 'sbp' | 'apple_pay' | 'google_pay' | 'combo'

interface AddonSelection {
  addon_id: string
  quantity: number
}

export default function BookingCreate() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { message } = App.useApp()

  const bathhouseId = searchParams.get('bathhouse') ?? ''
  const initialDate = searchParams.get('date') ?? dayjs().format('YYYY-MM-DD')
  const initialFrom = searchParams.get('from') ?? ''
  const initialTo = searchParams.get('to') ?? ''

  // Step state
  const [currentStep, setCurrentStep] = useState(0)

  // Step 1: Date/Time/Duration/Guests
  const [selectedDate, setSelectedDate] = useState(initialDate)
  const [selectedSlot, setSelectedSlot] = useState<{ from: string; to: string } | null>(
    initialFrom && initialTo ? { from: initialFrom, to: initialTo } : null,
  )
  const [guestCount, setGuestCount] = useState(1)
  const [comment, setComment] = useState('')

  // Step 2: Add-ons
  const [selectedAddons, setSelectedAddons] = useState<AddonSelection[]>([])

  // Step 3: Discounts
  const [promoCode, setPromoCode] = useState('')
  const [promoValidated, setPromoValidated] = useState<{ discount: number; type?: string } | null>(null)
  const [promoError, setPromoError] = useState('')
  const [certificateCode, setCertificateCode] = useState('')
  const [usePoints, setUsePoints] = useState(false)
  const [pointsAmount, setPointsAmount] = useState(0)
  const [useReferral, setUseReferral] = useState(false)
  const [referralAmount, setReferralAmount] = useState(0)

  // Step 4: Payment
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('card')
  const [comboWalletAmount, setComboWalletAmount] = useState(0)

  // Slot conflict
  const [slotConflict, setSlotConflict] = useState(false)

  // Data fetching
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

  const startTime = selectedSlot?.from ?? ''
  const endTime = selectedSlot?.to ?? ''

  const { data: priceData, isLoading: priceLoading } = useGetBathhousesIdPriceCalculator(
    bathhouseId,
    { start: startTime, end: endTime },
    { query: { enabled: !!bathhouseId && !!selectedSlot } },
  )
  const priceInfo = priceData?.data

  const { data: addonsData } = useGetBathhousesIdAddons(bathhouseId, {
    query: { enabled: !!bathhouseId },
  })
  const addons = (addonsData?.data ?? []).filter((a) => a.is_active)

  const { data: walletData } = useGetMyWallet({ query: { retry: false } })
  const walletBalance = (walletData?.data as Record<string, unknown>)?.balance as number ?? 0

  const isRequestMode = (bathhouse as Record<string, unknown>)?.booking_mode === 'request'

  // Calculated final price with add-ons
  const addonsTotal = useMemo(() => {
    return selectedAddons.reduce((sum, sel) => {
      const addon = addons.find((a) => a.id === sel.addon_id)
      return sum + (addon?.price ?? 0) * sel.quantity
    }, 0)
  }, [selectedAddons, addons])

  const totalPrice = (priceInfo?.final_price ?? 0) + addonsTotal

  // Certificate balance
  const [checkCertificate, setCheckCertificate] = useState(false)
  const { data: certBalanceData } = useGetCertificatesCodeBalance(
    certificateCode.trim(),
    {
      query: {
        enabled: checkCertificate && certificateCode.trim().length >= 4,
        retry: false,
      },
    },
  )
  const certificateBalance = certBalanceData?.data?.balance ?? null

  // Promo validation
  const promoValidateMutation = usePostPromoCodesValidate({
    mutation: {
      onSuccess: (response) => {
        const data = response?.data
        if (data && data.discount != null) {
          setPromoValidated({ discount: data.discount, type: data.type })
          setPromoError('')
          message.success(`Промокод применён: скидка ${data.type === 'percentage' ? `${data.discount}%` : formatPrice(data.discount)}`)
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

  // Booking creation
  const createBookingMutation = usePostBookings({
    mutation: {
      onSuccess: (response) => {
        const booking = response?.data
        if (booking?.id) {
          message.success(isRequestMode ? 'Заявка на бронирование отправлена!' : 'Бронирование создано!')
          navigate(`/client/bookings/${booking.id}`)
        }
      },
      onError: (error) => {
        const axiosErr = error as { response?: { data?: { error?: { code?: string; message?: string } } } }
        const errData = axiosErr?.response?.data?.error
        if (errData?.code === 'slot_unavailable') {
          setSlotConflict(true)
          return
        }
        const errorMsg = errData?.message ?? 'Не удалось создать бронирование'
        message.error(errorMsg)
      },
    },
  })

  const handleCreateBooking = () => {
    if (!selectedSlot || !bathhouseId) return

    const addonPayload = selectedAddons.length > 0
      ? selectedAddons.map((s) => ({ addon_id: s.addon_id, quantity: s.quantity }))
      : undefined

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
        addons: addonPayload,
      },
    })
  }

  // Addon selection helpers
  const toggleAddon = (addonId: string) => {
    setSelectedAddons((prev) => {
      const existing = prev.find((s) => s.addon_id === addonId)
      if (existing) {
        return prev.filter((s) => s.addon_id !== addonId)
      }
      return [...prev, { addon_id: addonId, quantity: 1 }]
    })
  }

  const updateAddonQuantity = (addonId: string, quantity: number) => {
    setSelectedAddons((prev) =>
      prev.map((s) => (s.addon_id === addonId ? { ...s, quantity } : s)),
    )
  }

  // Step validation
  const canProceedFromStep = (step: number): boolean => {
    switch (step) {
      case 0:
        return !!selectedSlot && !!bathhouseId && guestCount >= 1
      case 1:
        return true // add-ons are optional
      case 2:
        return true // discounts are optional
      case 3:
        return !!paymentMethod
      default:
        return false
    }
  }

  const handleNext = () => {
    if (canProceedFromStep(currentStep) && currentStep < 3) {
      setCurrentStep(currentStep + 1)
    }
  }

  const handlePrev = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1)
    }
  }

  // Combo payment: max wallet amount is the lesser of wallet balance and total price
  const maxWalletForCombo = Math.min(walletBalance, totalPrice)
  const comboCardAmount = paymentMethod === 'combo' ? totalPrice - comboWalletAmount : 0

  if (bathhouseLoading) {
    return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  }

  if (!bathhouse) {
    return <Empty description="Баня не найдена" />
  }

  // Slot conflict screen
  if (slotConflict) {
    const alternativeSlots = slots.filter(
      (s) => s.available && (s.startTime !== selectedSlot?.from || s.endTime !== selectedSlot?.to),
    )
    return (
      <div style={{ maxWidth: 800, margin: '0 auto' }}>
        <Result
          status="warning"
          title="Слот только что занят"
          subTitle="К сожалению, выбранное время было забронировано другим гостем. Выберите другой слот."
          extra={[
            <Button
              key="back"
              onClick={() => {
                setSlotConflict(false)
                setSelectedSlot(null)
                setCurrentStep(0)
              }}
            >
              Выбрать другое время
            </Button>,
          ]}
        />
        {alternativeSlots.length > 0 && (
          <Card title="Ближайшие доступные слоты" style={{ marginTop: 16 }}>
            <Space wrap>
              {alternativeSlots.slice(0, 5).map((slot) => {
                const fromTime = slot.startTime?.slice(11, 16) ?? ''
                const toTime = slot.endTime?.slice(11, 16) ?? ''
                return (
                  <Button
                    key={`${slot.startTime}-${slot.endTime}`}
                    icon={<ClockCircleOutlined />}
                    onClick={() => {
                      if (slot.startTime && slot.endTime) {
                        setSelectedSlot({ from: slot.startTime, to: slot.endTime })
                        setSlotConflict(false)
                        setCurrentStep(0)
                      }
                    }}
                  >
                    {fromTime} — {toTime}
                    {slot.price != null && ` (${formatPrice(slot.price)})`}
                  </Button>
                )
              })}
            </Space>
          </Card>
        )}
      </div>
    )
  }

  const stepItems = [
    { title: 'Дата и время' },
    { title: 'Доп. услуги' },
    { title: 'Скидки' },
    { title: 'Оплата' },
  ]

  return (
    <div style={{ maxWidth: 800, margin: '0 auto' }}>
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate(`/client/bathhouse/${bathhouse?.slug ?? bathhouseId}`)}
        style={{ marginBottom: 16 }}
      >
        Назад к бане
      </Button>

      <Title level={3}>Бронирование: {bathhouse.name}</Title>

      {isRequestMode && (
        <Alert
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
          message="Бронирование по заявке"
          description="Это заведение работает по заявкам. После оформления заявки владелец подтвердит бронирование в течение установленного времени. Средства будут заблокированы до подтверждения."
        />
      )}

      <Steps
        current={currentStep}
        items={stepItems}
        style={{ marginBottom: 24 }}
        onChange={(step) => {
          // Allow clicking previous steps, only allow forward if valid
          if (step < currentStep) {
            setCurrentStep(step)
          } else if (step === currentStep + 1 && canProceedFromStep(currentStep)) {
            setCurrentStep(step)
          }
        }}
      />

      {/* Step 1: Date/Time/Duration/Guests */}
      {currentStep === 0 && (
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
                      const fromTime = slot.startTime?.slice(11, 16) ?? slot.startTime ?? ''
                      const toTime = slot.endTime?.slice(11, 16) ?? slot.endTime ?? ''
                      return (
                        <Button
                          key={`${slot.startTime}-${slot.endTime}`}
                          type={isSelected ? 'primary' : 'default'}
                          disabled={!slot.available || !slot.startTime || !slot.endTime}
                          icon={<ClockCircleOutlined />}
                          onClick={() => {
                            if (slot.startTime && slot.endTime) {
                              setSelectedSlot({ from: slot.startTime, to: slot.endTime })
                            }
                          }}
                        >
                          {fromTime} — {toTime}
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
      )}

      {/* Step 2: Add-ons */}
      {currentStep === 1 && (
        <Card title="Дополнительные услуги" style={{ marginBottom: 16 }}>
          {addons.length === 0 ? (
            <Empty description="Нет доступных дополнительных услуг" image={Empty.PRESENTED_IMAGE_SIMPLE} />
          ) : (
            <Space direction="vertical" style={{ width: '100%' }} size={12}>
              {addons.map((addon) => {
                const selected = selectedAddons.find((s) => s.addon_id === addon.id)
                const unitLabel = addon.unit === 'per_hour' ? '/час' : addon.unit === 'per_person' ? '/чел.' : '/шт.'
                return (
                  <Card
                    key={addon.id}
                    size="small"
                    style={{
                      border: selected ? '2px solid #1677ff' : '1px solid #d9d9d9',
                      cursor: 'pointer',
                    }}
                    onClick={() => addon.id && toggleAddon(addon.id)}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <div>
                        <Checkbox checked={!!selected} style={{ marginRight: 8 }} />
                        <Text strong>{addon.name}</Text>
                        {addon.description && (
                          <Text type="secondary" style={{ marginLeft: 8 }}>{addon.description}</Text>
                        )}
                      </div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <Tag color="blue">{formatPrice(addon.price ?? 0)}{unitLabel}</Tag>
                        {selected && (
                          <InputNumber
                            min={1}
                            max={10}
                            value={selected.quantity}
                            onChange={(v) => addon.id && updateAddonQuantity(addon.id, v ?? 1)}
                            onClick={(e) => e.stopPropagation()}
                            size="small"
                            style={{ width: 70 }}
                          />
                        )}
                      </div>
                    </div>
                  </Card>
                )
              })}
              {selectedAddons.length > 0 && (
                <div style={{ textAlign: 'right', marginTop: 8 }}>
                  <Text strong>Итого за доп. услуги: {formatPrice(addonsTotal)}</Text>
                </div>
              )}
            </Space>
          )}
        </Card>
      )}

      {/* Step 3: Promo/Certificate/Wallet discounts */}
      {currentStep === 2 && (
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
                  Скидка: {promoValidated.type === 'percentage' ? `${promoValidated.discount}%` : formatPrice(promoValidated.discount)}
                </Text>
              )}
            </div>

            {/* Gift certificate */}
            <div>
              <Text strong><GiftOutlined /> Подарочный сертификат:</Text>
              <Space.Compact style={{ width: '100%', marginTop: 8 }}>
                <Input
                  value={certificateCode}
                  onChange={(e) => {
                    setCertificateCode(e.target.value)
                    setCheckCertificate(false)
                  }}
                  placeholder="BANI-XXXX-XXXX"
                />
                <Button
                  onClick={() => setCheckCertificate(true)}
                  disabled={certificateCode.trim().length < 4}
                >
                  Проверить
                </Button>
              </Space.Compact>
              {checkCertificate && certificateBalance != null && (
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
                  placeholder="Сумма"
                  style={{ width: 150 }}
                />
              )}
            </div>
          </Space>
        </Card>
      )}

      {/* Step 4: Price breakdown + Payment method */}
      {currentStep === 3 && (
        <>
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
                {addonsTotal > 0 && (
                  <Descriptions.Item label="Дополнительные услуги">
                    {formatPrice(addonsTotal)}
                  </Descriptions.Item>
                )}
                {promoValidated && (
                  <Descriptions.Item label="Скидка по промокоду">
                    <Text type="success">
                      -{promoValidated.type === 'percentage' ? `${promoValidated.discount}%` : formatPrice(promoValidated.discount)}
                    </Text>
                  </Descriptions.Item>
                )}
                {certificateBalance != null && certificateBalance > 0 && (
                  <Descriptions.Item label="Сертификат">
                    <Text type="success">до -{formatPrice(certificateBalance)}</Text>
                  </Descriptions.Item>
                )}
                {(() => {
                  const depositPercent = (bathhouse as Record<string, unknown>)?.security_deposit_percent as number
                  return depositPercent > 0 ? (
                    <Descriptions.Item label="Залог (возвратный)">
                      <Text type="warning">
                        ~{formatPrice(Math.round((priceInfo.base_price ?? 0) * depositPercent / 100))}
                      </Text>
                    </Descriptions.Item>
                  ) : null
                })()}
                <Descriptions.Item label="К оплате">
                  <Text strong style={{ fontSize: 18 }}>
                    {formatPrice(totalPrice)}
                  </Text>
                </Descriptions.Item>
                {paymentMethod === 'combo' && comboWalletAmount > 0 && (
                  <>
                    <Descriptions.Item label="Из кошелька">
                      <Text type="success">{formatPrice(comboWalletAmount)}</Text>
                    </Descriptions.Item>
                    <Descriptions.Item label="Картой">
                      <Text>{formatPrice(comboCardAmount)}</Text>
                    </Descriptions.Item>
                  </>
                )}
              </Descriptions>
            ) : (
              <Text type="secondary">Выберите слот для расчёта цены</Text>
            )}
          </Card>

          {/* Payment method */}
          <Card title="Способ оплаты" style={{ marginBottom: 16 }}>
            <Radio.Group
              value={paymentMethod}
              onChange={(e) => {
                setPaymentMethod(e.target.value)
                if (e.target.value !== 'combo') {
                  setComboWalletAmount(0)
                }
              }}
              style={{ width: '100%' }}
            >
              <Space direction="vertical" style={{ width: '100%' }} size={8}>
                {walletBalance > 0 && walletBalance >= totalPrice && (
                  <Radio.Button
                    value="wallet"
                    style={paymentMethodStyle(paymentMethod === 'wallet')}
                  >
                    <WalletOutlined style={{ marginRight: 8 }} />
                    Кошелёк ({formatPrice(walletBalance)})
                  </Radio.Button>
                )}
                <Radio.Button
                  value="card"
                  style={paymentMethodStyle(paymentMethod === 'card')}
                >
                  <CreditCardOutlined style={{ marginRight: 8 }} />
                  Банковская карта
                </Radio.Button>
                <Radio.Button
                  value="sbp"
                  style={paymentMethodStyle(paymentMethod === 'sbp')}
                >
                  <BankOutlined style={{ marginRight: 8 }} />
                  СБП
                </Radio.Button>
                <Radio.Button
                  value="apple_pay"
                  style={paymentMethodStyle(paymentMethod === 'apple_pay')}
                >
                  <AppleOutlined style={{ marginRight: 8 }} />
                  Apple Pay
                </Radio.Button>
                <Radio.Button
                  value="google_pay"
                  style={paymentMethodStyle(paymentMethod === 'google_pay')}
                >
                  <GoogleOutlined style={{ marginRight: 8 }} />
                  Google Pay
                </Radio.Button>
                {walletBalance > 0 && walletBalance < totalPrice && (
                  <Radio.Button
                    value="combo"
                    style={paymentMethodStyle(paymentMethod === 'combo')}
                  >
                    <WalletOutlined style={{ marginRight: 4 }} />
                    +
                    <CreditCardOutlined style={{ marginLeft: 4, marginRight: 8 }} />
                    Кошелёк + Карта
                  </Radio.Button>
                )}
              </Space>
            </Radio.Group>

            {/* Combo payment slider */}
            {paymentMethod === 'combo' && maxWalletForCombo > 0 && (
              <div style={{ marginTop: 16 }}>
                <Text>Сумма из кошелька:</Text>
                <Slider
                  min={0}
                  max={maxWalletForCombo}
                  step={100}
                  value={comboWalletAmount}
                  onChange={setComboWalletAmount}
                  tooltip={{ formatter: (v) => v != null ? formatPrice(v) : '' }}
                />
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Text type="success">
                    <WalletOutlined /> {formatPrice(comboWalletAmount)} из кошелька
                  </Text>
                  <Text>
                    <CreditCardOutlined /> {formatPrice(comboCardAmount)} картой
                  </Text>
                </div>
              </div>
            )}
          </Card>

          {/* Refund policy */}
          <Alert
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
            message="Политика отмены"
            description={(() => {
              const policy = (bathhouse as Record<string, unknown>)?.cancellation_policy as string
              switch (policy) {
                case 'moderate':
                  return 'Умеренная: более 72ч — 100% возврат, 24-72ч — 50% возврат, менее 24ч — без возврата.'
                case 'strict':
                  return 'Строгая: более 7 дней — 100% возврат, 3-7 дней — 50% возврат, менее 3 дней — без возврата.'
                default:
                  return 'Гибкая: более 24ч до начала — 100% возврат, менее 24ч — 50% возврат.'
              }
            })()}
          />

          {isRequestMode && (
            <Alert
              type="warning"
              showIcon
              style={{ marginBottom: 16 }}
              message="Бронирование по заявке"
              description="Средства будут заблокированы на вашей карте до подтверждения владельцем. Если заявка будет отклонена или истечёт время ожидания, средства разблокируются автоматически."
            />
          )}
        </>
      )}

      <Divider />

      {/* Navigation buttons */}
      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
        <div>
          {currentStep > 0 && (
            <Button size="large" onClick={handlePrev}>
              Назад
            </Button>
          )}
          {currentStep === 0 && (
            <Button size="large" onClick={() => navigate(`/client/bathhouse/${bathhouse?.slug ?? bathhouseId}`)}>
              Отмена
            </Button>
          )}
        </div>
        <div>
          {currentStep < 3 ? (
            <Button
              type="primary"
              size="large"
              onClick={handleNext}
              disabled={!canProceedFromStep(currentStep)}
            >
              Далее
            </Button>
          ) : (
            <Button
              type="primary"
              size="large"
              icon={isRequestMode ? <CheckCircleOutlined /> : <ShoppingCartOutlined />}
              onClick={handleCreateBooking}
              loading={createBookingMutation.isPending}
              disabled={!selectedSlot || !bathhouseId}
            >
              {isRequestMode ? 'Отправить заявку' : 'Забронировать'}
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}

function paymentMethodStyle(isSelected: boolean): React.CSSProperties {
  return {
    display: 'block',
    width: '100%',
    height: 'auto',
    padding: '12px 16px',
    textAlign: 'left' as const,
    borderRadius: 8,
    border: isSelected ? '2px solid #1677ff' : '1px solid #d9d9d9',
    background: isSelected ? '#e6f4ff' : '#fff',
  }
}
