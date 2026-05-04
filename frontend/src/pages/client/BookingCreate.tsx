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

const { Text } = Typography

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
    return (
      <div className="rh-admin-empty-state">
        <div className="rh-admin-empty-state__title">Баня не найдена</div>
        <p className="rh-admin-empty-state__text">
          Вернитесь в каталог и выберите доступный объект для бронирования.
        </p>
      </div>
    )
  }

  // Slot conflict screen
  if (slotConflict) {
    const alternativeSlots = slots.filter(
      (s) => s.available && (s.startTime !== selectedSlot?.from || s.endTime !== selectedSlot?.to),
    )
    return (
      <div className="rh-booking-page">
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
          <Card title="Ближайшие доступные слоты" className="rh-admin-detail-card rh-alert-spaced">
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
    <div className="rh-stack rh-booking-page">
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate(`/client/bathhouse/${bathhouse?.slug ?? bathhouseId}`)}
        className="rh-admin-detail-back"
      >
        Назад к бане
      </Button>

      <h1 className="rh-page-title">Бронирование: {bathhouse.name}</h1>

      {rebookId && rebookData?.data && (
        <Alert
          type="info"
          showIcon
          icon={<ReloadOutlined />}
          className="rh-booking-alert"
          title="Повторное бронирование"
          description="Параметры предыдущего бронирования предзаполнены. Выберите удобную дату и время."
          data-testid="rebook-alert"
        />
      )}

      {isRequestMode && (
        <Alert
          type="warning"
          showIcon
          className="rh-booking-alert"
          title="Бронирование по заявке"
          description="Это заведение работает по заявкам. После оформления заявки владелец подтвердит бронирование в течение установленного времени. Средства будут заблокированы до подтверждения."
        />
      )}

      <Steps
        current={currentStep}
        items={stepItems}
        className="rh-booking-steps"
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
        <Card title="Дата и время" className="rh-admin-detail-card">
          <Space orientation="vertical" className="rh-full-width" size={16}>
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
                className="rh-booking-field-control"
              />
            </div>

            <div>
              <Text strong>Доступные слоты:</Text>
              <Spin spinning={slotsLoading}>
                {slots.length === 0 ? (
                  <div className="rh-admin-empty-state">
                    <div className="rh-admin-empty-state__title">Нет доступных слотов</div>
                    <p className="rh-admin-empty-state__text">
                      Выберите другую дату или вернитесь к объекту позже.
                    </p>
                  </div>
                ) : (
                  <Space orientation="vertical" className="rh-booking-slot-stack" size="middle">
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
                className="rh-booking-field-control"
              />
            </div>

            <div>
              <Text strong>Комментарий:</Text>
              <Input.TextArea
                value={comment}
                onChange={(e) => setComment(e.target.value)}
                placeholder="Пожелания к бронированию"
                rows={2}
                className="rh-booking-field-control"
              />
            </div>
          </Space>
        </Card>
      )}

      {/* Step 2: Add-ons */}
      {currentStep === 1 && (
        <Card title="Дополнительные услуги" className="rh-admin-detail-card">
          {addons.length === 0 ? (
            <div className="rh-admin-empty-state">
              <div className="rh-admin-empty-state__title">Нет доступных дополнительных услуг</div>
              <p className="rh-admin-empty-state__text">
                Можно перейти дальше и оформить бронирование без дополнений.
              </p>
            </div>
          ) : (
            <Space orientation="vertical" className="rh-full-width" size={12}>
              {addons.map((addon) => {
                const selected = selectedAddons.find((s) => s.addon_id === addon.id)
                const unitLabel = addon.unit === 'per_hour' ? '/час' : addon.unit === 'per_person' ? '/чел.' : '/шт.'
                return (
                  <Card
                    key={addon.id}
                    size="small"
                    className={selected ? 'rh-booking-addon-card rh-booking-addon-card--selected' : 'rh-booking-addon-card'}
                    onClick={() => addon.id && toggleAddon(addon.id)}
                  >
                    <div className="rh-booking-addon-card__row">
                      <div>
                        <Checkbox checked={!!selected} className="rh-booking-inline-control" />
                        <Text strong>{addon.name}</Text>
                        {addon.description && (
                          <Text type="secondary" className="rh-booking-inline-note">{addon.description}</Text>
                        )}
                      </div>
                      <div className="rh-booking-addon-card__controls">
                        <Tag color="blue">{formatPrice(addon.price ?? 0)}{unitLabel}</Tag>
                        {selected && (
                          <InputNumber
                            min={1}
                            max={10}
                            value={selected.quantity}
                            onChange={(v) => addon.id && updateAddonQuantity(addon.id, v ?? 1)}
                            onClick={(e) => e.stopPropagation()}
                            size="small"
                            className="rh-booking-addon-qty"
                          />
                        )}
                      </div>
                    </div>
                  </Card>
                )
              })}
              {selectedAddons.length > 0 && (
                <div className="rh-booking-addon-total">
                  <Text strong>Итого за доп. услуги: {formatPrice(addonsTotal)}</Text>
                </div>
              )}
            </Space>
          )}
        </Card>
      )}

      {/* Step 3: Promo/Certificate/Wallet discounts */}
      {currentStep === 2 && (
        <Card title="Скидки и бонусы" className="rh-admin-detail-card">
          <Space orientation="vertical" className="rh-full-width" size={16}>
            {/* Promo code */}
            <div>
              <Text strong><TagOutlined /> Промокод:</Text>
              <Space.Compact className="rh-booking-compact">
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
              {promoError && <Text type="danger" className="rh-booking-hint">{promoError}</Text>}
              {promoValidated && (
                <Text type="success" className="rh-booking-hint">
                  Скидка: {promoValidated.type === 'percentage' ? `${promoValidated.discount}%` : formatPrice(promoValidated.discount)}
                </Text>
              )}
            </div>

            {/* Gift certificate */}
            <div>
              <Text strong><GiftOutlined /> Подарочный сертификат:</Text>
              <Space.Compact className="rh-booking-compact">
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
                <div className="rh-booking-certificate-result">
                  <Text type="success" className="rh-booking-hint">
                    Баланс сертификата: {formatPrice(certificateBalance)}
                  </Text>
                  {certificateApplied > 0 && (
                    <Text type="success" className="rh-booking-inline-note">
                      (будет списано: {formatPrice(certificateApplied)})
                    </Text>
                  )}
                </div>
              )}
            </div>

            {/* Loyalty points */}
            <div className="rh-booking-toggle-row">
              <Switch checked={usePoints} onChange={setUsePoints} />
              <Text>Использовать баллы лояльности</Text>
              {usePoints && (
                <InputNumber
                  min={0}
                  value={pointsAmount}
                  onChange={(v) => setPointsAmount(v ?? 0)}
                  placeholder="Кол-во баллов"
                  className="rh-price-input"
                />
              )}
            </div>

            {/* Referral balance */}
            <div className="rh-booking-toggle-row">
              <Switch checked={useReferral} onChange={setUseReferral} />
              <Text>Использовать реферальный бонус</Text>
              {useReferral && (
                <InputNumber
                  min={0}
                  value={referralAmount}
                  onChange={(v) => setReferralAmount(v ?? 0)}
                  placeholder="Сумма"
                  className="rh-price-input"
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
          <Card title="Итого" className="rh-admin-detail-card">
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
              <div className="rh-booking-deposit-note">
                <Text type="warning">
                  <LockOutlined className="rh-booking-inline-control" />
                  Залог (возвратный): ~{formatPrice(Math.round((priceInfo.base_price ?? 0) * depositPercent / 100))}
                </Text>
              </div>
            )}
          </Card>

          {/* Payment method */}
          <Card title="Способ оплаты" className="rh-admin-detail-card">
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
              className="rh-full-width"
            >
              <Space orientation="vertical" className="rh-full-width" size={8}>
                {walletBalance > 0 && walletBalance >= totalPrice && (
                  <Radio.Button
                    value="wallet"
                    className={paymentMethodClass(paymentMethod === 'wallet')}
                  >
                    <WalletOutlined className="rh-booking-method-icon" />
                    Кошелёк ({formatPrice(walletBalance)})
                  </Radio.Button>
                )}
                <Radio.Button
                  value="card"
                  className={paymentMethodClass(paymentMethod === 'card')}
                >
                  <CreditCardOutlined className="rh-booking-method-icon" />
                  Банковская карта
                </Radio.Button>
                <Radio.Button
                  value="sbp"
                  className={paymentMethodClass(paymentMethod === 'sbp')}
                >
                  <BankOutlined className="rh-booking-method-icon" />
                  СБП
                </Radio.Button>
                <Radio.Button
                  value="apple_pay"
                  className={paymentMethodClass(paymentMethod === 'apple_pay')}
                >
                  <AppleOutlined className="rh-booking-method-icon" />
                  Apple Pay
                </Radio.Button>
                <Radio.Button
                  value="google_pay"
                  className={paymentMethodClass(paymentMethod === 'google_pay')}
                >
                  <GoogleOutlined className="rh-booking-method-icon" />
                  Google Pay
                </Radio.Button>
                {walletBalance > 0 && walletBalance < totalPrice && (
                  <Radio.Button
                    value="combo"
                    className={paymentMethodClass(paymentMethod === 'combo')}
                  >
                    <WalletOutlined className="rh-booking-method-icon rh-booking-method-icon--compact" />
                    +
                    <CreditCardOutlined className="rh-booking-method-icon" />
                    Кошелёк + Карта
                  </Radio.Button>
                )}
              </Space>
            </Radio.Group>

            {/* Saved cards selection */}
            {paymentMethod === 'card' && savedCards.length > 0 && (
              <div className="rh-booking-saved-cards" data-testid="saved-cards-section">
                <Text strong className="rh-catalog__field-label">Сохранённые карты:</Text>
                <Radio.Group
                  value={selectedSavedCardId}
                  onChange={(e) => setSelectedSavedCardId(e.target.value)}
                  className="rh-full-width"
                >
                  <Space orientation="vertical" className="rh-full-width" size={8}>
                    {savedCards.map((card) => (
                      <Radio key={card.id} value={card.id} className="rh-booking-card-radio">
                        <CreditCardOutlined className="rh-booking-method-icon" />
                        {card.brand ?? 'Карта'} •••• {card.last4}
                        {card.expiry_month != null && card.expiry_year != null && (
                          <Text type="secondary" className="rh-booking-inline-note">
                            {String(card.expiry_month).padStart(2, '0')}/{String(card.expiry_year).slice(-2)}
                          </Text>
                        )}
                        {card.is_default && <Tag color="blue" className="rh-booking-inline-note">Основная</Tag>}
                      </Radio>
                    ))}
                    <Radio value={null} className="rh-booking-card-radio">
                      <CreditCardOutlined className="rh-booking-method-icon" />
                      Новая карта
                    </Radio>
                  </Space>
                </Radio.Group>
              </div>
            )}

            {/* Combo payment slider */}
            {paymentMethod === 'combo' && maxWalletForCombo > 0 && (
              <div className="rh-booking-combo">
                <Text>Сумма из кошелька:</Text>
                <Slider
                  min={0}
                  max={maxWalletForCombo}
                  step={100}
                  value={comboWalletAmount}
                  onChange={setComboWalletAmount}
                  tooltip={{ formatter: (v) => v != null ? formatPrice(v) : '' }}
                />
                <div className="rh-booking-combo__summary">
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
            className="rh-booking-alert"
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
              className="rh-booking-alert"
              title="Средства будут заблокированы"
              description="Средства будут заблокированы на вашей карте до подтверждения владельцем. Если заявка будет отклонена или истечёт время ожидания, средства разблокируются автоматически."
              data-testid="hold-indicator"
            />
          )}
        </>
      )}

      <Divider />

      {/* Navigation buttons */}
      <div className="rh-booking-nav">
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

function paymentMethodClass(isSelected: boolean): string {
  return isSelected ? 'rh-booking-payment-method rh-booking-payment-method--selected' : 'rh-booking-payment-method'
}
