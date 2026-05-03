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
} from '@/components/design/system'
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
  LockOutlined,
  ReloadOutlined,
} from '@/components/design/icons'
import dayjs from 'dayjs'
import { useGetBathhousesId, useGetBathhousesIdAvailableSlots } from '@/api/generated/bathhouses/bathhouses'
import { useGetBathhousesIdPriceCalculator } from '@/api/generated/pricing/pricing'
import { usePostBookings, useGetBookingsIdRebookData } from '@/api/generated/bookings/bookings'
import { usePostPromoCodesValidate } from '@/api/generated/promo-codes/promo-codes'
import { useGetCertificatesCodeBalance } from '@/api/generated/certificates/certificates'
import { useGetBathhousesIdAddons } from '@/api/generated/add-ons/add-ons'
import { useGetMyWallet } from '@/api/generated/wallet/wallet'
import { useGetMySavedCards } from '@/api/generated/saved-cards/saved-cards'
import ContiguousSlotSelector from '@/components/ContiguousSlotSelector'
import PriceBreakdown from '@/components/PriceBreakdown'
import { formatPrice } from '@/lib/format'
import { formatSlotTimeLabel, getRangeHours, resolveSlotRangeSelection, type SlotRangeSelection } from '@/lib/slot-selection'

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
  const rebookId = searchParams.get('rebook') ?? ''

  // Step state
  const [currentStep, setCurrentStep] = useState(0)

  // Step 1: Date/Time/Duration/Guests
  const [selectedDate, setSelectedDate] = useState(initialDate)
  const [selectedSlotRange, setSelectedSlotRange] = useState<SlotRangeSelection | null>(
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
  const [selectedSavedCardId, setSelectedSavedCardId] = useState<string | null>(null)
  const [comboWalletAmount, setComboWalletAmount] = useState(0)

  // Slot conflict
  const [slotConflict, setSlotConflict] = useState(false)

  // Re-booking pre-fill applied flag
  const [rebookApplied, setRebookApplied] = useState(false)

  // Data fetching
  const { data: bathhouseData, isLoading: bathhouseLoading } = useGetBathhousesId(bathhouseId, {
    query: { enabled: !!bathhouseId },
  })
  const bathhouse = bathhouseData?.data
  const minDurationHours = Math.max(1, bathhouse?.min_duration ?? 1)

  const { data: slotsData, isLoading: slotsLoading } = useGetBathhousesIdAvailableSlots(
    bathhouseId,
    { date: selectedDate },
    { query: { enabled: !!bathhouseId && !!selectedDate } },
  )
  const slots = useMemo(() => slotsData?.data ?? [], [slotsData?.data])
  const resolvedSlotRange = resolveSlotRangeSelection(slots, selectedSlotRange)
  const selectedRangeHours = resolvedSlotRange ? getRangeHours(resolvedSlotRange.from, resolvedSlotRange.to) : 0
  const selectedSlot = resolvedSlotRange && selectedRangeHours >= minDurationHours ? resolvedSlotRange : null
  const selectedRangeNeedsHours = resolvedSlotRange ? Math.max(0, minDurationHours - selectedRangeHours) : 0

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

  // Saved cards
  const { data: savedCardsData } = useGetMySavedCards(undefined, {
    query: { retry: false },
  })
  const savedCards = (savedCardsData?.data ?? []) as Array<{
    id?: string
    last4?: string
    brand?: string
    expiry_month?: number
    expiry_year?: number
    is_default?: boolean
  }>

  // Re-booking data — apply pre-fill via onSuccess (runs once when data first arrives)
  const { data: rebookData } = useGetBookingsIdRebookData(rebookId, {
    query: {
      enabled: !!rebookId && !rebookApplied,
      retry: false,
    },
  })

  // Apply rebook pre-fill from fetched data
  const applyRebookData = () => {
    if (rebookApplied || !rebookData?.data) return
    const rebook = rebookData.data
    if (rebook.guest_count) setGuestCount(rebook.guest_count)
    if (rebook.addons && rebook.addons.length > 0) {
      setSelectedAddons(
        rebook.addons.map((a) => ({
          addon_id: a.addon_id ?? '',
          quantity: a.quantity ?? 1,
        })).filter((a) => a.addon_id),
      )
    }
    setRebookApplied(true)
  }

  // Trigger rebook apply when data becomes available
  if (rebookData?.data && !rebookApplied) {
    // Schedule for next microtask to avoid setState during render
    queueMicrotask(applyRebookData)
  }

  const isRequestMode = (bathhouse as Record<string, unknown>)?.booking_mode === 'request'

  // Calculate booking duration in hours for per_hour add-ons
  const durationHours = (() => {
    if (!startTime || !endTime) return 1
    const sParts = startTime.split(':').map(Number)
    const eParts = endTime.split(':').map(Number)
    let diff = ((eParts[0] ?? 0) * 60 + (eParts[1] ?? 0)) - ((sParts[0] ?? 0) * 60 + (sParts[1] ?? 0))
    if (diff <= 0) diff += 24 * 60 // wraparound midnight
    return Math.max(1, diff / 60)
  })()

  // Calculated final price with add-ons (accounting for unit type)
  const addonsTotal = useMemo(() => {
    return selectedAddons.reduce((sum, sel) => {
      const addon = addons.find((a) => a.id === sel.addon_id)
      if (!addon) return sum
      let multiplier = sel.quantity
      if (addon.unit === 'per_hour') multiplier = sel.quantity * durationHours
      else if (addon.unit === 'per_person') multiplier = sel.quantity * guestCount
      return sum + (addon.price ?? 0) * multiplier
    }, 0)
  }, [selectedAddons, addons, durationHours, guestCount])

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

  // Certificate amount to apply (min of balance and total)
  const certificateApplied = certificateBalance != null && certificateBalance > 0
    ? Math.min(certificateBalance, totalPrice)
    : 0

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
          navigate(`/client/bookings/${booking.id}`, {
            state: {
              paymentMethod,
              ...(paymentMethod === 'combo' && comboWalletAmount > 0 ? { walletApplied: comboWalletAmount } : {}),
            },
          })
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
        certificate_code: certificateApplied > 0 ? certificateCode.trim() : undefined,
        use_points: usePoints ? pointsAmount : undefined,
        use_referral_bonus: useReferral ? referralAmount : undefined,
        addons: addonPayload,
        saved_card_id: selectedSavedCardId ?? undefined,
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

  // Build add-on line items for PriceBreakdown
  const addonLineItems = useMemo(() => {
    return selectedAddons.map((sel) => {
      const addon = addons.find((a) => a.id === sel.addon_id)
      if (!addon) return { name: 'Доп. услуга', price: 0 }
      let multiplier = sel.quantity
      if (addon.unit === 'per_hour') multiplier = sel.quantity * durationHours
      else if (addon.unit === 'per_person') multiplier = sel.quantity * guestCount
      return { name: addon.name ?? 'Доп. услуга', price: (addon.price ?? 0) * multiplier }
    }).filter((item) => item.price > 0)
  }, [selectedAddons, addons, durationHours, guestCount])

  // Promo discount in absolute amount
  const promoDiscountAmount = useMemo(() => {
    if (!promoValidated) return 0
    if (promoValidated.type === 'percentage') {
      return Math.round((priceInfo?.base_price ?? 0) * promoValidated.discount / 100)
    }
    return promoValidated.discount
  }, [promoValidated, priceInfo])

  if (bathhouseLoading) {
    return (
      <div className="rh-fullscreen-state">
        <Spin size="large" />
      </div>
    )
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
                setSelectedSlotRange(null)
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
                      if (!slot.startTime || !slot.endTime) return
                      setSelectedSlotRange({ from: slot.startTime, to: slot.endTime })
                      setSlotConflict(false)
                      setCurrentStep(0)
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

  const depositPercent = (bathhouse as Record<string, unknown>)?.security_deposit_percent as number | undefined
  const areaAvgPrice = (bathhouse as Record<string, unknown>)?.area_average_price as number | undefined

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

      {rebookId && rebookData?.data && (
        <Alert
          type="info"
          showIcon
          icon={<ReloadOutlined />}
          style={{ marginBottom: 16 }}
          title="Повторное бронирование"
          description="Параметры предыдущего бронирования предзаполнены. Выберите удобную дату и время."
          data-testid="rebook-alert"
        />
      )}

      {isRequestMode && (
        <Alert
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
          title="Бронирование по заявке"
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

      {/* Step 1: Date/Time/Guests */}
      {currentStep === 0 && (
        <Card title="Дата и время" style={{ marginBottom: 16 }}>
          <Space orientation="vertical" style={{ width: '100%' }} size={16}>
            <div>
              <Text strong>Дата:</Text>
              <DatePicker
                value={dayjs(selectedDate)}
                onChange={(d) => {
                  if (d) {
                    setSelectedDate(d.format('YYYY-MM-DD'))
                    setSelectedSlotRange(null)
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
                  <Space orientation="vertical" style={{ width: '100%', marginTop: 8 }} size="middle">
                    <ContiguousSlotSelector
                      slots={slots}
                      value={selectedSlotRange}
                      onChange={setSelectedSlotRange}
                      minDurationHours={minDurationHours}
                      label={`Выберите непрерывный интервал${minDurationHours > 1 ? ` (от ${minDurationHours} ч)` : ''}`}
                      description="Выбор времени и длительности объединён: отмечайте соседние слоты подряд в одном блоке."
                      size="middle"
                    />

                    {resolvedSlotRange && !selectedSlot && (
                      <Alert
                        type="info"
                        showIcon={false}
                        title={`${formatSlotTimeLabel(resolvedSlotRange.from)} - ${formatSlotTimeLabel(resolvedSlotRange.to)} · ${selectedRangeHours} ч`}
                        description={`Добавьте ещё ${selectedRangeNeedsHours} ч, чтобы продолжить оформление.`}
                      />
                    )}
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
            <Space orientation="vertical" style={{ width: '100%' }} size={12}>
              {addons.map((addon) => {
                const selected = selectedAddons.find((s) => s.addon_id === addon.id)
                const unitLabel = addon.unit === 'per_hour' ? '/час' : addon.unit === 'per_person' ? '/чел.' : '/шт.'
                return (
                  <Card
                    key={addon.id}
                    size="small"
                    style={{
                      border: selected ? '2px solid #0f766e' : '1px solid #c9c1b5',
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
          <Space orientation="vertical" style={{ width: '100%' }} size={16}>
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
                <div style={{ marginTop: 4 }}>
                  <Text type="success" style={{ fontSize: 12 }}>
                    Баланс сертификата: {formatPrice(certificateBalance)}
                  </Text>
                  {certificateApplied > 0 && (
                    <Text type="success" style={{ fontSize: 12, marginLeft: 8 }}>
                      (будет списано: {formatPrice(certificateApplied)})
                    </Text>
                  )}
                </div>
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
          {/* Price Breakdown */}
          <Card title="Итого" style={{ marginBottom: 16 }}>
            {priceLoading ? (
              <Spin />
            ) : priceInfo ? (
              <PriceBreakdown
                basePrice={priceInfo.base_price ?? 0}
                addOns={addonLineItems.length > 0 ? addonLineItems : undefined}
                longSessionDiscount={priceInfo.price_saving}
                promoDiscount={promoDiscountAmount > 0 ? promoDiscountAmount : undefined}
                certificateDiscount={certificateApplied > 0 ? certificateApplied : undefined}
                serviceFee={(priceInfo as Record<string, unknown>)?.service_fee as number | undefined}
                extraGuestSurcharge={(priceInfo as Record<string, unknown>)?.extra_guest_surcharge as number | undefined}
                holidaySurcharge={(priceInfo as Record<string, unknown>)?.holiday_surcharge as number | undefined}
                holidayName={(priceInfo as Record<string, unknown>)?.holiday_name as string | undefined}
                seasonalTariffMultiplier={(priceInfo as Record<string, unknown>)?.seasonal_tariff_multiplier as number | undefined}
                seasonalTariffName={(priceInfo as Record<string, unknown>)?.seasonal_tariff_name as string | undefined}
                walletPayment={paymentMethod === 'combo' && comboWalletAmount > 0 ? comboWalletAmount : paymentMethod === 'wallet' ? walletBalance : undefined}
                areaAveragePrice={areaAvgPrice}
              />
            ) : (
              <Text type="secondary">Выберите слот для расчёта цены</Text>
            )}

            {depositPercent != null && depositPercent > 0 && priceInfo && (
              <div style={{ marginTop: 8 }}>
                <Text type="warning">
                  <LockOutlined style={{ marginRight: 4 }} />
                  Залог (возвратный): ~{formatPrice(Math.round((priceInfo.base_price ?? 0) * depositPercent / 100))}
                </Text>
              </div>
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
                if (e.target.value !== 'card') {
                  setSelectedSavedCardId(null)
                }
              }}
              style={{ width: '100%' }}
            >
              <Space orientation="vertical" style={{ width: '100%' }} size={8}>
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

            {/* Saved cards selection */}
            {paymentMethod === 'card' && savedCards.length > 0 && (
              <div style={{ marginTop: 16 }} data-testid="saved-cards-section">
                <Text strong style={{ display: 'block', marginBottom: 8 }}>Сохранённые карты:</Text>
                <Radio.Group
                  value={selectedSavedCardId}
                  onChange={(e) => setSelectedSavedCardId(e.target.value)}
                  style={{ width: '100%' }}
                >
                  <Space orientation="vertical" style={{ width: '100%' }} size={8}>
                    {savedCards.map((card) => (
                      <Radio key={card.id} value={card.id} style={{ display: 'block' }}>
                        <CreditCardOutlined style={{ marginRight: 8 }} />
                        {card.brand ?? 'Карта'} •••• {card.last4}
                        {card.expiry_month != null && card.expiry_year != null && (
                          <Text type="secondary" style={{ marginLeft: 8 }}>
                            {String(card.expiry_month).padStart(2, '0')}/{String(card.expiry_year).slice(-2)}
                          </Text>
                        )}
                        {card.is_default && <Tag color="blue" style={{ marginLeft: 8 }}>Основная</Tag>}
                      </Radio>
                    ))}
                    <Radio value={null} style={{ display: 'block' }}>
                      <CreditCardOutlined style={{ marginRight: 8 }} />
                      Новая карта
                    </Radio>
                  </Space>
                </Radio.Group>
              </div>
            )}

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
            title="Политика отмены"
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

          {/* Hold indicator for request-based bookings */}
          {isRequestMode && (
            <Alert
              type="warning"
              showIcon
              icon={<LockOutlined />}
              style={{ marginBottom: 16 }}
              title="Средства будут заблокированы"
              description="Средства будут заблокированы на вашей карте до подтверждения владельцем. Если заявка будет отклонена или истечёт время ожидания, средства разблокируются автоматически."
              data-testid="hold-indicator"
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
    borderRadius: 20,
    border: isSelected ? '2px solid #0f766e' : '1px solid #c9c1b5',
    background: isSelected ? 'rgba(15, 118, 110, 0.08)' : '#fff',
  }
}
