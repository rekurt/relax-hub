import { useEffect, useMemo, useState } from 'react'
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
  App,
  Empty,
  Steps,
  Radio,
  Checkbox,
  Tag,
  Collapse,
} from '@/components/design/system'
import {
  ArrowLeftOutlined,
  WalletOutlined,
  CreditCardOutlined,
  BankOutlined,
  AppleOutlined,
  GoogleOutlined,
  ShoppingCartOutlined,
  CheckCircleOutlined,
  ReloadOutlined,
} from '@/components/design/icons'
import dayjs from 'dayjs'
import { useGetBathhousesId, useGetBathhousesIdAvailableSlots } from '@/api/generated/bathhouses/bathhouses'
import { useGetBathhousesIdPriceCalculator } from '@/api/generated/pricing/pricing'
import { usePostBookings, useGetBookingsIdRebookData } from '@/api/generated/bookings/bookings'
import { useGetBathhousesIdAddons } from '@/api/generated/add-ons/add-ons'
import { useGetMyWallet } from '@/api/generated/wallet/wallet'
import { useGetMySavedCards } from '@/api/generated/saved-cards/saved-cards'
import ContiguousSlotSelector from '@/components/ContiguousSlotSelector'
import PriceBreakdown from '@/components/PriceBreakdown'
import DiscountInput, { EMPTY_DISCOUNT_STATE, type DiscountState } from '@/components/DiscountInput'
import { formatPrice } from '@/lib/format'
import { formatSlotTimeLabel, getRangeHours, resolveSlotRangeSelection, type SlotRangeSelection } from '@/lib/slot-selection'
import { useDeviceToken } from '@/lib/useDeviceToken'

const { Title, Text } = Typography

type PaymentMethod = 'wallet' | 'card' | 'sbp' | 'apple_pay' | 'google_pay'

interface AddonSelection {
  addon_id: string
  quantity: number
}

const PUSH_PROMPT_KEY = 'rh_push_prompted'

export default function BookingCreate() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { message } = App.useApp()
  const { requestPushPermission } = useDeviceToken()

  const maybeRequestPushPermission = () => {
    if (typeof window === 'undefined') return
    if (localStorage.getItem(PUSH_PROMPT_KEY)) return
    localStorage.setItem(PUSH_PROMPT_KEY, '1')
    void requestPushPermission()
  }

  const bathhouseId = searchParams.get('bathhouse') ?? ''
  const initialDate = searchParams.get('date') ?? dayjs().format('YYYY-MM-DD')
  const initialFrom = searchParams.get('from') ?? ''
  const initialTo = searchParams.get('to') ?? ''
  const rebookId = searchParams.get('rebook') ?? ''

  const [currentStep, setCurrentStep] = useState(0)

  const [selectedDate, setSelectedDate] = useState(initialDate)
  const [selectedSlotRange, setSelectedSlotRange] = useState<SlotRangeSelection | null>(
    initialFrom && initialTo ? { from: initialFrom, to: initialTo } : null,
  )
  const [guestCount, setGuestCount] = useState(1)
  const [comment, setComment] = useState('')
  const [selectedAddons, setSelectedAddons] = useState<AddonSelection[]>([])
  const [discounts, setDiscounts] = useState<DiscountState>(EMPTY_DISCOUNT_STATE)
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('card')
  const [selectedSavedCardId, setSelectedSavedCardId] = useState<string | null>(null)
  const [autoResolvedSlot, setAutoResolvedSlot] = useState<{ from: string; to: string } | null>(null)
  const [rebookApplied, setRebookApplied] = useState(false)

  const { data: bathhouseData, isLoading: bathhouseLoading } = useGetBathhousesId(bathhouseId, {
    query: { enabled: !!bathhouseId },
  })
  const bathhouse = bathhouseData?.data
  const minDurationHours = Math.max(1, bathhouse?.min_duration ?? 1)

  const { data: slotsData, isLoading: slotsLoading, refetch: refetchSlots } = useGetBathhousesIdAvailableSlots(
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
  const walletInfo = walletData?.data as Record<string, unknown> | undefined
  const walletBalance = (walletInfo?.balance as number) ?? 0
  const availablePoints = (walletInfo?.loyalty_points as number) ?? 0
  const availableReferralBalance = (walletInfo?.referral_balance as number) ?? 0

  const { data: savedCardsData } = useGetMySavedCards(undefined, { query: { retry: false } })
  const savedCards = (savedCardsData?.data ?? []) as Array<{
    id?: string
    last4?: string
    brand?: string
    expiry_month?: number
    expiry_year?: number
    is_default?: boolean
  }>

  const { data: rebookData } = useGetBookingsIdRebookData(rebookId, {
    query: { enabled: !!rebookId && !rebookApplied, retry: false },
  })

  if (rebookData?.data && !rebookApplied) {
    queueMicrotask(() => {
      if (rebookApplied || !rebookData.data) return
      const rebook = rebookData.data
      if (rebook.guest_count) setGuestCount(rebook.guest_count)
      if (rebook.addons && rebook.addons.length > 0) {
        setSelectedAddons(
          rebook.addons
            .map((a) => ({ addon_id: a.addon_id ?? '', quantity: a.quantity ?? 1 }))
            .filter((a) => a.addon_id),
        )
      }
      setRebookApplied(true)
    })
  }

  const isRequestMode = (bathhouse as Record<string, unknown>)?.booking_mode === 'request'

  const durationHours = useMemo(() => {
    if (!startTime || !endTime) return 1
    const sParts = startTime.split(':').map(Number)
    const eParts = endTime.split(':').map(Number)
    let diff = ((eParts[0] ?? 0) * 60 + (eParts[1] ?? 0)) - ((sParts[0] ?? 0) * 60 + (sParts[1] ?? 0))
    if (diff <= 0) diff += 24 * 60
    return Math.max(1, diff / 60)
    // eslint-disable-next-line react-hooks/preserve-manual-memoization
  }, [startTime, endTime])

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

  const promoDiscountAmount = useMemo(() => {
    if (discounts.promoDiscount <= 0) return 0
    if (discounts.promoDiscountType === 'percentage') {
      return Math.round((priceInfo?.base_price ?? 0) * discounts.promoDiscount / 100)
    }
    return discounts.promoDiscount
  }, [discounts.promoDiscount, discounts.promoDiscountType, priceInfo?.base_price])

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

  // Net amount the user actually has to pay after server-side discounts.
  // Wallet deduction must be capped by this — not by gross totalPrice — otherwise
  // we'd send wallet_amount > booking.total_price after the server applies discounts.
  const netPayable = Math.max(
    0,
    totalPrice
      - promoDiscountAmount
      - (discounts.certificateAmount || 0)
      - (discounts.pointsAmount || 0)
      - (discounts.referralAmount || 0),
  )
  const walletApplied = walletBalance > 0 ? Math.min(walletBalance, netPayable) : 0
  const cardOnlyPath = paymentMethod !== 'wallet' && walletApplied < netPayable
  const walletWillCoverFullPrice = walletApplied >= netPayable && netPayable > 0

  const createBookingMutation = usePostBookings({
    mutation: {
      onSuccess: (response) => {
        const booking = response?.data
        if (!booking?.id) return
        message.success(isRequestMode ? 'Заявка отправлена!' : 'Бронирование создано!')
        maybeRequestPushPermission()
        navigate(`/client/bookings/${booking.id}`, {
          state: { paymentMethod, walletApplied: walletApplied > 0 ? walletApplied : undefined },
        })
      },
      onError: (error) => {
        const axiosErr = error as { response?: { data?: { error?: { code?: string; message?: string } } } }
        const errData = axiosErr?.response?.data?.error
        if (errData?.code === 'slot_unavailable') {
          // Auto-resolve to nearest available slot instead of showing a separate "conflict" page
          refetchSlots().then((res) => {
            const fresh = res.data?.data ?? []
            const candidate = fresh.find((s) => s.available && (s.startTime !== startTime || s.endTime !== endTime))
            if (candidate?.startTime && candidate?.endTime) {
              setSelectedSlotRange({ from: candidate.startTime, to: candidate.endTime })
              setAutoResolvedSlot({ from: candidate.startTime, to: candidate.endTime })
              setCurrentStep(0)
            } else {
              setSelectedSlotRange(null)
              setCurrentStep(0)
              message.warning('Слот занят, выберите другое время')
            }
          })
          return
        }
        message.error(errData?.message ?? 'Не удалось создать бронирование')
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
        promo_code: discounts.promoCode || undefined,
        certificate_code: discounts.certificateAmount > 0 ? discounts.certificateCode : undefined,
        use_points: discounts.pointsAmount > 0 ? discounts.pointsAmount : undefined,
        use_referral_bonus: discounts.referralAmount > 0 ? discounts.referralAmount : undefined,
        addons: selectedAddons.length > 0
          ? selectedAddons.map((s) => ({ addon_id: s.addon_id, quantity: s.quantity }))
          : undefined,
        saved_card_id: selectedSavedCardId ?? undefined,
      },
    })
  }

  // Clear auto-resolved alert once user picks something else
  useEffect(() => {
    if (!autoResolvedSlot) return
    if (selectedSlotRange?.from !== autoResolvedSlot.from || selectedSlotRange?.to !== autoResolvedSlot.to) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- clear stale auto-resolved alert when user picks a different slot
      setAutoResolvedSlot(null)
    }
  }, [selectedSlotRange, autoResolvedSlot])

  const toggleAddon = (addonId: string) => {
    setSelectedAddons((prev) => {
      const existing = prev.find((s) => s.addon_id === addonId)
      if (existing) return prev.filter((s) => s.addon_id !== addonId)
      return [...prev, { addon_id: addonId, quantity: 1 }]
    })
  }

  const updateAddonQuantity = (addonId: string, quantity: number) => {
    setSelectedAddons((prev) => prev.map((s) => (s.addon_id === addonId ? { ...s, quantity } : s)))
  }

  const canProceedToPayment = !!selectedSlot && !!bathhouseId && guestCount >= 1

  if (bathhouseLoading) {
    return (
      <div className="rh-fullscreen-state">
        <Spin size="large" />
      </div>
    )
  }
  if (!bathhouse) return <Empty description="Баня не найдена" />

  const stepItems = [
    { title: 'Когда и сколько' },
    { title: 'Оплата' },
  ]

  const depositPercent = (bathhouse as Record<string, unknown>)?.security_deposit_percent as number | undefined
  const securityDeposit = depositPercent != null && depositPercent > 0 && priceInfo
    ? Math.round((priceInfo.base_price ?? 0) * depositPercent / 100)
    : undefined
  const areaAvgPrice = (bathhouse as Record<string, unknown>)?.area_average_price as number | undefined

  const addonsCollapseItems = addons.length === 0 ? [] : [{
    key: 'addons',
    label: selectedAddons.length > 0
      ? `Дополнительные услуги (${selectedAddons.length}, +${formatPrice(addonsTotal)})`
      : 'Дополнительные услуги (необязательно)',
    children: (
      <Space orientation="vertical" style={{ width: '100%' }} size={8}>
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
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 8 }}>
                <div style={{ flex: 1, minWidth: 0 }}>
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
      </Space>
    ),
  }]

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
          description="Параметры предыдущей брони предзаполнены. Выберите удобное время."
          data-testid="rebook-alert"
        />
      )}

      {isRequestMode && (
        <Alert
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
          title="Бронирование по заявке"
          description="После оформления владелец подтвердит бронирование. Средства будут заблокированы до подтверждения, при отказе вернутся автоматически."
        />
      )}

      {autoResolvedSlot && (
        <Alert
          type="warning"
          showIcon
          closable
          onClose={() => setAutoResolvedSlot(null)}
          style={{ marginBottom: 16 }}
          title="Время изменилось"
          description={`Предыдущий слот только что заняли. Мы выбрали ближайший свободный: ${formatSlotTimeLabel(autoResolvedSlot.from)} – ${formatSlotTimeLabel(autoResolvedSlot.to)}. Можете изменить вручную.`}
        />
      )}

      <Steps
        current={currentStep}
        items={stepItems}
        style={{ marginBottom: 24 }}
        onChange={(step) => {
          if (step < currentStep) setCurrentStep(step)
          else if (step === 1 && canProceedToPayment) setCurrentStep(1)
        }}
      />

      {currentStep === 0 && (
        <>
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
                        description="Отмечайте соседние слоты подряд."
                        size="middle"
                      />
                      {resolvedSlotRange && !selectedSlot && (
                        <Alert
                          type="info"
                          showIcon={false}
                          title={`${formatSlotTimeLabel(resolvedSlotRange.from)} - ${formatSlotTimeLabel(resolvedSlotRange.to)} · ${selectedRangeHours} ч`}
                          description={`Добавьте ещё ${selectedRangeNeedsHours} ч, чтобы продолжить.`}
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
                <Text strong>Комментарий (необязательно):</Text>
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

          {addonsCollapseItems.length > 0 && (
            <Card style={{ marginBottom: 16 }}>
              <Collapse ghost items={addonsCollapseItems} />
            </Card>
          )}
        </>
      )}

      {currentStep === 1 && (
        <>
          <Card title="Итого" style={{ marginBottom: 16 }}>
            {priceLoading ? (
              <Spin />
            ) : priceInfo ? (
              <PriceBreakdown
                basePrice={priceInfo.base_price ?? 0}
                addOns={addonLineItems.length > 0 ? addonLineItems : undefined}
                longSessionDiscount={priceInfo.price_saving}
                promoDiscount={promoDiscountAmount > 0 ? promoDiscountAmount : undefined}
                certificateDiscount={discounts.certificateAmount > 0 ? discounts.certificateAmount : undefined}
                pointsDiscount={discounts.pointsAmount > 0 ? discounts.pointsAmount : undefined}
                referralDiscount={discounts.referralAmount > 0 ? discounts.referralAmount : undefined}
                serviceFee={(priceInfo as Record<string, unknown>)?.service_fee as number | undefined}
                extraGuestSurcharge={(priceInfo as Record<string, unknown>)?.extra_guest_surcharge as number | undefined}
                holidaySurcharge={(priceInfo as Record<string, unknown>)?.holiday_surcharge as number | undefined}
                holidayName={(priceInfo as Record<string, unknown>)?.holiday_name as string | undefined}
                seasonalTariffMultiplier={(priceInfo as Record<string, unknown>)?.seasonal_tariff_multiplier as number | undefined}
                seasonalTariffName={(priceInfo as Record<string, unknown>)?.seasonal_tariff_name as string | undefined}
                walletPayment={walletApplied > 0 ? walletApplied : undefined}
                securityDeposit={securityDeposit}
                areaAveragePrice={areaAvgPrice}
              />
            ) : (
              <Text type="secondary">Выберите слот для расчёта цены</Text>
            )}
          </Card>

          <Card title="Скидка и бонусы" style={{ marginBottom: 16 }}>
            <DiscountInput
              bathhouseId={bathhouseId}
              basePrice={priceInfo?.base_price ?? 0}
              totalPrice={totalPrice}
              availablePoints={availablePoints}
              availableReferralBalance={availableReferralBalance}
              value={discounts}
              onChange={setDiscounts}
            />
          </Card>

          <Card title="Способ оплаты" style={{ marginBottom: 16 }}>
            {walletApplied > 0 && !walletWillCoverFullPrice && (
              <Alert
                type="info"
                style={{ marginBottom: 12 }}
                message={
                  <span>
                    <WalletOutlined /> С кошелька будет списано <b>{formatPrice(walletApplied)}</b>,
                    остаток <b>{formatPrice(netPayable - walletApplied)}</b> — выбранным методом ниже.
                  </span>
                }
              />
            )}
            <Radio.Group
              value={paymentMethod}
              onChange={(e) => {
                setPaymentMethod(e.target.value)
                if (e.target.value !== 'card') setSelectedSavedCardId(null)
              }}
              style={{ width: '100%' }}
            >
              <Space orientation="vertical" style={{ width: '100%' }} size={8}>
                {walletWillCoverFullPrice && (
                  <Radio.Button value="wallet" style={paymentMethodStyle(paymentMethod === 'wallet')}>
                    <WalletOutlined style={{ marginRight: 8 }} />
                    Только кошелёк ({formatPrice(walletBalance)})
                  </Radio.Button>
                )}
                <Radio.Button value="card" style={paymentMethodStyle(paymentMethod === 'card')}>
                  <CreditCardOutlined style={{ marginRight: 8 }} />
                  Банковская карта
                </Radio.Button>
                <Radio.Button value="sbp" style={paymentMethodStyle(paymentMethod === 'sbp')}>
                  <BankOutlined style={{ marginRight: 8 }} />
                  СБП
                </Radio.Button>
                <Radio.Button value="apple_pay" style={paymentMethodStyle(paymentMethod === 'apple_pay')}>
                  <AppleOutlined style={{ marginRight: 8 }} />
                  Apple Pay
                </Radio.Button>
                <Radio.Button value="google_pay" style={paymentMethodStyle(paymentMethod === 'google_pay')}>
                  <GoogleOutlined style={{ marginRight: 8 }} />
                  Google Pay
                </Radio.Button>
              </Space>
            </Radio.Group>

            {paymentMethod === 'card' && cardOnlyPath && savedCards.length > 0 && (
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
                      <CreditCardOutlined style={{ marginRight: 8 }} /> Новая карта
                    </Radio>
                  </Space>
                </Radio.Group>
              </div>
            )}
          </Card>

          <Alert
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
            title="Политика отмены"
            description={(() => {
              const policy = (bathhouse as Record<string, unknown>)?.cancellation_policy as string
              switch (policy) {
                case 'moderate':
                  return 'Умеренная: >72ч — 100% возврат, 24-72ч — 50%, <24ч — без возврата.'
                case 'strict':
                  return 'Строгая: >7 дней — 100% возврат, 3-7 дней — 50%, <3 дней — без возврата.'
                default:
                  return 'Гибкая: >24ч до начала — 100% возврат, <24ч — 50%.'
              }
            })()}
          />
        </>
      )}

      <Divider />

      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
        <div>
          {currentStep > 0 ? (
            <Button size="large" onClick={() => setCurrentStep(0)}>Назад</Button>
          ) : (
            <Button size="large" onClick={() => navigate(`/client/bathhouse/${bathhouse?.slug ?? bathhouseId}`)}>
              Отмена
            </Button>
          )}
        </div>
        <div>
          {currentStep === 0 ? (
            <Button
              type="primary"
              size="large"
              onClick={() => setCurrentStep(1)}
              disabled={!canProceedToPayment}
            >
              К оплате
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
