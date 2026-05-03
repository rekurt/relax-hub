import { useMemo, useState } from 'react'
import Axios from 'axios'
import { useNavigate, useSearchParams } from 'react-router-dom'
import {
  Alert,
  App,
  Button,
  Card,
  Checkbox,
  Col,
  DatePicker,
  Input,
  InputNumber,
  Result,
  Row,
  Space,
  Spin,
  Steps,
} from 'antd'
import dayjs from 'dayjs'
import { CalendarOutlined, CheckCircleOutlined, ClockCircleOutlined, LockOutlined, UserOutlined } from '@ant-design/icons'
import { useGetBathhousesId, useGetBathhousesIdAvailableSlots } from '@/api/generated/bathhouses/bathhouses'
import { useGetBathhousesIdPriceCalculator } from '@/api/generated/pricing/pricing'
import { axiosInstance } from '@/api/axios-instance'
import ContiguousSlotSelector from '@/components/ContiguousSlotSelector'
import PageHeader from '@/components/PageHeader'
import PublicState from '@/components/PublicState'
import {
  getBookingModeTrustCopy,
  getCancellationPolicyLabel,
  getDepositSummary,
} from '@/lib/booking-flow'
import { formatPrice } from '@/lib/format'
import { useAuthStore } from '@/stores/auth'
import { getPublicErrorMessage } from '@/lib/public-route'
import { formatSlotTimeLabel, getRangeHours, resolveSlotRangeSelection, type SlotRangeSelection } from '@/lib/slot-selection'

const publicHttp = Axios.create({
  baseURL: '/api/v1',
})

interface AuthVerifyData {
  token?: string
  requires_2fa?: boolean
  user?: {
    id?: string
    name?: string
    email?: string
    phone?: string
    role?: string
    region?: string
  }
}

interface BookingData {
  id?: string
}

export default function PublicCheckout() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { message } = App.useApp()
  const setAuth = useAuthStore((s) => s.setAuth)
  const currentUser = useAuthStore((s) => s.user)

  const bathhouseId = searchParams.get('bathhouse') ?? ''
  const [selectedDate, setSelectedDate] = useState(
    searchParams.get('date')
      ?? searchParams.get('start')?.slice(0, 10)
      ?? dayjs().format('YYYY-MM-DD'),
  )
  const [selectedSlotRange, setSelectedSlotRange] = useState<SlotRangeSelection | null>(() => {
    const from = searchParams.get('from') ?? searchParams.get('start')
    const to = searchParams.get('to') ?? searchParams.get('end')
    return from && to ? { from, to } : null
  })
  const [guestCount, setGuestCount] = useState(Number(searchParams.get('guests') ?? '2'))
  const [name, setName] = useState(currentUser?.name ?? '')
  const [phone, setPhone] = useState(currentUser?.phone ?? '')
  const [comment, setComment] = useState('')
  const [ageConfirmed, setAgeConfirmed] = useState(false)
  const [otpSent, setOtpSent] = useState(false)
  const [otpCode, setOtpCode] = useState('')
  const [sendingOtp, setSendingOtp] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [createdBookingId, setCreatedBookingId] = useState<string | null>(null)
  const [otpError, setOtpError] = useState<string | null>(null)
  const [checkoutError, setCheckoutError] = useState<string | null>(null)
  const [slotConflictError, setSlotConflictError] = useState<string | null>(null)

  const {
    data: bathhouseData,
    isLoading: bathhouseLoading,
    isError: bathhouseIsError,
    refetch: refetchBathhouse,
  } = useGetBathhousesId(bathhouseId, {
    query: { enabled: !!bathhouseId },
  })
  const bathhouse = bathhouseData?.data

  const {
    data: slotsData,
    isLoading: slotsLoading,
    isError: slotsIsError,
    refetch: refetchSlots,
  } = useGetBathhousesIdAvailableSlots(
    bathhouseId,
    { date: selectedDate },
    { query: { enabled: !!bathhouseId && !!selectedDate } },
  )
  const slots = useMemo(() => slotsData?.data ?? [], [slotsData?.data])
  const minDurationHours = Math.max(1, bathhouse?.min_duration ?? 1)
  const resolvedSlotRange = resolveSlotRangeSelection(slots, selectedSlotRange)
  const selectedRangeHours = resolvedSlotRange ? getRangeHours(resolvedSlotRange.from, resolvedSlotRange.to) : 0
  const selectedSlot = resolvedSlotRange && selectedRangeHours >= minDurationHours ? resolvedSlotRange : null
  const selectedRangeNeedsHours = resolvedSlotRange ? Math.max(0, minDurationHours - selectedRangeHours) : 0

  const { data: priceData } = useGetBathhousesIdPriceCalculator(
    bathhouseId,
    { start: selectedSlot?.from ?? '', end: selectedSlot?.to ?? '' },
    { query: { enabled: !!bathhouseId && !!selectedSlot } },
  )
  const price = priceData?.data

  const currentStep = useMemo(() => {
    if (createdBookingId) return 2
    if (currentUser?.role === 'client' || otpSent) return 1
    return 0
  }, [createdBookingId, currentUser?.role, otpSent])

  const bookingModeLabel = getBookingModeTrustCopy((bathhouse as Record<string, unknown> | undefined)?.booking_mode as string | undefined)
  const cancellationPolicyLabel = getCancellationPolicyLabel((bathhouse as Record<string, unknown> | undefined)?.cancellation_policy as string | undefined)
  const minimumDurationLabel = (bathhouse as Record<string, unknown> | undefined)?.min_duration
    ? `Минимум ${(bathhouse as Record<string, unknown>).min_duration} ч`
    : 'Минимум без ограничений'
  const depositSummary = getDepositSummary((bathhouse as Record<string, unknown> | undefined)?.security_deposit_percent as number | undefined)

  const clearCheckoutErrors = () => {
    setOtpError(null)
    setCheckoutError(null)
    setSlotConflictError(null)
  }

  const createBooking = async () => {
    if (!selectedSlot || !bathhouseId) return
    const response = await axiosInstance.post<{ data?: BookingData }>('/bookings', {
      bathhouse_id: bathhouseId,
      start_time: selectedSlot.from,
      end_time: selectedSlot.to,
      guest_count: guestCount,
      comment: comment || undefined,
    })
    const bookingId = response.data.data?.id
    if (!bookingId) {
      throw new Error('booking_id_missing')
    }
    setCreatedBookingId(bookingId)
    message.success('Бронирование создано')
    navigate(`/client/bookings/${bookingId}`, { replace: true })
  }

  const handleStartOTP = async () => {
    clearCheckoutErrors()
    if (!phone.trim()) {
      message.warning('Введите телефон')
      return
    }
    if (!selectedSlot) {
      message.warning('Сначала выберите слот')
      return
    }
    if (!name.trim()) {
      message.warning('Введите имя')
      return
    }
    if (!ageConfirmed) {
      message.warning('Подтвердите возраст 18+')
      return
    }

    setSendingOtp(true)
    try {
      await publicHttp.post('/auth/phone/start', { phone })
      setOtpSent(true)
      message.success('Код отправлен')
    } catch {
      setOtpError('Не удалось отправить код')
    } finally {
      setSendingOtp(false)
    }
  }

  const handleVerifyAndBook = async () => {
    clearCheckoutErrors()
    if (!selectedSlot) {
      message.warning('Слот не выбран')
      return
    }
    if (!currentUser?.role && (!otpCode.trim() || otpCode.trim().length !== 6)) {
      message.warning('Введите 6-значный код')
      return
    }

    setSubmitting(true)
    try {
      if (!currentUser?.role) {
        const authResponse = await publicHttp.post<{ data?: AuthVerifyData }>('/auth/verify-phone', {
          phone,
          code: otpCode.trim(),
          name,
          age_confirmed: ageConfirmed,
        })

        const authData = authResponse.data.data
        if (authData?.requires_2fa) {
          message.warning('Для этого номера включена дополнительная защита. Завершите вход через страницу логина.')
          navigate('/login')
          return
        }

        if (authData?.token && authData.user) {
          setAuth(authData.token, authData.user)
        } else {
          throw new Error('auth_incomplete')
        }
      }

      await createBooking()
    } catch (error) {
      if (Axios.isAxiosError(error) && error.response?.status === 409) {
        setSlotConflictError('Выбранный слот уже недоступен')
        return
      }

      const fallbackMessage = currentUser?.role === 'client'
        ? 'Не удалось завершить бронирование. Попробуйте еще раз.'
        : 'Не удалось подтвердить код. Проверьте его и попробуйте еще раз.'
      const errorMessage = getPublicErrorMessage(error, fallbackMessage)

      if (currentUser?.role === 'client') {
        setCheckoutError(errorMessage)
      } else {
        setOtpError(errorMessage)
      }
    } finally {
      setSubmitting(false)
    }
  }

  if (bathhouseLoading) {
    return (
      <PublicState
        kind="loading"
        title="Загружаем бронирование"
        description="Проверяем баню, слоты и стоимость."
      />
    )
  }

  if (bathhouseIsError) {
    return (
      <PublicState
        kind="error"
        title="Не удалось загрузить страницу бронирования"
        description="Повторите попытку или вернитесь в каталог."
        actionText="Повторить"
        onAction={() => void refetchBathhouse()}
        secondaryActionText="В каталог"
        secondaryActionLink="/catalog"
      />
    )
  }

  if (!bathhouseId || !bathhouse) {
    return (
      <PublicState
        kind="empty"
        title="Сначала выберите баню и слот"
        description="Для публичного бронирования вернитесь в каталог и откройте свободное время."
        actionText="В каталог"
        actionLink="/catalog"
      />
    )
  }

  if (createdBookingId) {
    return (
      <Result
        status="success"
        icon={<CheckCircleOutlined />}
        title="Бронь создана"
        subTitle="Переводим вас в личный кабинет клиента для оплаты и дальнейшего управления поездкой."
      />
    )
  }

  return (
    <div className="bani-stack">
      <PageHeader
        eyebrow="Публичное бронирование"
        title={bathhouse.name}
        description="Вы выбираете слот, подтверждаете телефон и сразу попадаете в свою бронь без отдельной регистрации и лишних ответвлений."
      />

      <section className="bani-hero-panel">
        <div className="bani-hero-panel__eyebrow">Оформление</div>
        <h2 className="bani-hero-panel__title">Остался один короткий шаг до брони</h2>
        <div className="bani-hero-panel__description">
          На этом экране остаются только реальные действия: слот, контакты, SMS и создание брони. Ограничения и правила видны рядом со сводкой.
        </div>
        <div className="bani-hero-panel__meta">
          <div className="bani-hero-panel__meta-item">
            <span className="bani-hero-panel__meta-label">Подтверждение</span>
            <div className="bani-hero-panel__meta-value">{bookingModeLabel}</div>
          </div>
          <div className="bani-hero-panel__meta-item">
            <span className="bani-hero-panel__meta-label">Минимум</span>
            <div className="bani-hero-panel__meta-value">
              {(bathhouse as Record<string, unknown>).min_duration ? `${(bathhouse as Record<string, unknown>).min_duration} ч` : 'Без ограничения'}
            </div>
          </div>
          <div className="bani-hero-panel__meta-item">
            <span className="bani-hero-panel__meta-label">Оплата</span>
            <div className="bani-hero-panel__meta-value">После создания брони в кабинете клиента</div>
          </div>
        </div>
      </section>

      <div className="bani-grid bani-grid--content-aside">
        <Card>
          <div className="bani-section-card">
            <Steps
              current={currentStep}
              items={[
                { title: 'Слот' },
                { title: currentUser?.role === 'client' ? 'Подтверждение' : 'Контакты и SMS' },
                { title: 'Готово' },
              ]}
            />

            <div className="bani-info-grid">
              <div className="bani-info-card">
                <span className="bani-info-card__label">Дата посещения</span>
                <DatePicker
                  value={dayjs(selectedDate)}
                  onChange={(value) => {
                    const nextDate = value ? value.format('YYYY-MM-DD') : dayjs().format('YYYY-MM-DD')
                    setSelectedDate(nextDate)
                    setSelectedSlotRange(null)
                    setSlotConflictError(null)
                  }}
                  style={{ width: '100%' }}
                />
              </div>
              <div className="bani-info-card">
                <span className="bani-info-card__label">Количество гостей</span>
                <InputNumber
                  min={1}
                  max={bathhouse.max_guests ?? 20}
                  value={guestCount}
                  onChange={(value) => setGuestCount(value ?? 1)}
                  style={{ width: '100%' }}
                />
              </div>
            </div>

            <div className="bani-section-card">
              <h2 className="bani-section-card__title">Доступные слоты</h2>
              <div className="bani-section-card__description">
                Сначала показываем только доступные интервалы. После выбора слот закрепляется в правой сводке вместе с ценой.
              </div>
              <Spin spinning={slotsLoading && !slotsIsError}>
                {slotsIsError ? (
                  <PublicState
                    kind="error"
                    compact
                    title="Не удалось загрузить доступные слоты"
                    description="Обновите список слотов или выберите другую дату."
                    actionText="Обновить слоты"
                    onAction={() => void refetchSlots()}
                  />
                ) : slots.filter((slot) => slot.available).length === 0 ? (
                  <Alert
                    type="info"
                    showIcon
                    title="На выбранную дату нет доступных слотов"
                    description="Попробуйте другую дату или вернитесь в каталог, чтобы посмотреть похожие варианты."
                  />
                ) : (
                  <Space orientation="vertical" style={{ width: '100%' }} size="middle">
                    <ContiguousSlotSelector
                      slots={slots}
                      value={selectedSlotRange}
                      onChange={(nextRange) => {
                        clearCheckoutErrors()
                        setSelectedSlotRange(nextRange)
                      }}
                      minDurationHours={minDurationHours}
                      label={`Выберите непрерывный интервал${minDurationHours > 1 ? ` (от ${minDurationHours} ч)` : ''}`}
                      description="В одном блоке выбираются соседние слоты подряд. Ненужный край диапазона можно снять повторным кликом."
                    />

                    {resolvedSlotRange && !selectedSlot && (
                      <Alert
                        type="info"
                        showIcon={false}
                        title={`${formatSlotTimeLabel(resolvedSlotRange.from)} - ${formatSlotTimeLabel(resolvedSlotRange.to)} · ${selectedRangeHours} ч`}
                        description={`Добавьте ещё ${selectedRangeNeedsHours} ч, чтобы продолжить checkout.`}
                      />
                    )}
                  </Space>
                )}
              </Spin>
            </div>

            <div className="bani-section-card">
              <h2 className="bani-section-card__title">Контактные данные</h2>
              <div className="bani-section-card__description">
                Поля оставлены только для того, что реально нужно для создания брони и входа в кабинет.
              </div>
              <Row gutter={[16, 16]}>
                <Col xs={24} lg={12}>
                  <Input
                    size="large"
                    prefix={<UserOutlined />}
                    placeholder="Имя"
                    value={name}
                    onChange={(event) => setName(event.target.value)}
                    disabled={currentUser?.role === 'client'}
                  />
                </Col>
                <Col xs={24} lg={12}>
                  <Input
                    size="large"
                    placeholder="Телефон"
                    value={phone}
                    onChange={(event) => setPhone(event.target.value)}
                    disabled={currentUser?.role === 'client'}
                  />
                </Col>
                <Col xs={24}>
                  <Input.TextArea
                    rows={3}
                    placeholder="Комментарий к бронированию"
                    value={comment}
                    onChange={(event) => setComment(event.target.value)}
                  />
                </Col>
                {!currentUser && (
                  <Col xs={24}>
                    <Checkbox checked={ageConfirmed} onChange={(event) => setAgeConfirmed(event.target.checked)}>
                      Мне исполнилось 18 лет
                    </Checkbox>
                  </Col>
                )}
              </Row>
            </div>

            {slotConflictError && (
              <PublicState
                kind="error"
                compact
                title={slotConflictError}
                description="Обновите список слотов и выберите другое время. Имя, телефон и комментарий сохраняются."
                actionText="Обновить слоты"
                onAction={() => void refetchSlots()}
              />
            )}

            {checkoutError && (
              <Alert
                type="error"
                showIcon
                title={checkoutError}
              />
            )}

            {!currentUser && (
              <div className="bani-section-card">
                <Alert
                  type="info"
                  showIcon
                  icon={<LockOutlined />}
                  title="Авторизация встроена в checkout"
                  description="Мы подтверждаем телефон по SMS, создаем клиентский аккаунт и сохраняем бронь в ваш личный кабинет без отдельной регистрации."
                />
                {otpError && (
                  <Alert
                    type="error"
                    showIcon
                    title={otpError}
                  />
                )}
                {!otpSent ? (
                  <Button type="primary" size="large" loading={sendingOtp} onClick={handleStartOTP}>
                    Получить SMS-код
                  </Button>
                ) : (
                  <Space orientation="vertical" style={{ width: '100%' }} size="middle">
                    <Input
                      size="large"
                      placeholder="Код из SMS"
                      maxLength={6}
                      value={otpCode}
                      onChange={(event) => setOtpCode(event.target.value.replace(/\D/g, ''))}
                    />
                    <Button type="primary" size="large" loading={submitting} onClick={handleVerifyAndBook}>
                      Подтвердить и создать бронь
                    </Button>
                  </Space>
                )}
              </div>
            )}

            {currentUser?.role === 'client' && (
              <Button type="primary" size="large" loading={submitting} onClick={handleVerifyAndBook}>
                Создать бронь
              </Button>
            )}
          </div>
        </Card>

        <div className="bani-stack">
          <Card>
            <div className="bani-section-card">
              <h2 className="bani-section-card__title">Сводка бронирования</h2>
              <div className="bani-kv">
                <div className="bani-kv__row">
                  <span className="bani-kv__label">Баня</span>
                  <span className="bani-kv__value">{bathhouse.name}</span>
                </div>
                <div className="bani-kv__row">
                  <span className="bani-kv__label">Адрес</span>
                  <span className="bani-kv__value">{bathhouse.address || '—'}</span>
                </div>
                <div className="bani-kv__row">
                  <span className="bani-kv__label">Дата и время</span>
                  <span className="bani-kv__value">
                    {resolvedSlotRange
                      ? `${dayjs(resolvedSlotRange.from).format('D MMMM')} · ${dayjs(resolvedSlotRange.from).format('HH:mm')} - ${dayjs(resolvedSlotRange.to).format('HH:mm')} · ${selectedRangeHours} ч${selectedSlot ? '' : ` · нужно ещё ${selectedRangeNeedsHours} ч`}`
                      : 'Выберите слот'}
                  </span>
                </div>
                <div className="bani-kv__row">
                  <span className="bani-kv__label">Гостей</span>
                  <span className="bani-kv__value">{guestCount}</span>
                </div>
                <div className="bani-kv__row">
                  <span className="bani-kv__label">Цена</span>
                  <span className="bani-kv__value">
                    {price?.final_price ? formatPrice(price.final_price) : 'Будет рассчитана после выбора слота'}
                  </span>
                </div>
              </div>
            </div>
          </Card>

          <Card>
            <div className="bani-section-card">
              <h2 className="bani-section-card__title">Что важно до подтверждения</h2>
              <div className="bani-feature-list">
                <div className="bani-feature-item">
                  <div className="bani-feature-item__icon"><ClockCircleOutlined /></div>
                  <div className="bani-feature-item__copy">
                    <div className="bani-feature-item__title">{minimumDurationLabel}</div>
                    <div className="bani-feature-item__description">Минимальная длительность фиксирована заранее и не меняется после отправки SMS.</div>
                  </div>
                </div>
                <div className="bani-feature-item">
                  <div className="bani-feature-item__icon"><CheckCircleOutlined /></div>
                  <div className="bani-feature-item__copy">
                    <div className="bani-feature-item__title">{cancellationPolicyLabel} отмена</div>
                    <div className="bani-feature-item__description">Правила отмены видны до создания брони, а не после handoff в кабинет.</div>
                  </div>
                </div>
                <div className="bani-feature-item">
                  <div className="bani-feature-item__icon"><LockOutlined /></div>
                  <div className="bani-feature-item__copy">
                    <div className="bani-feature-item__title">{bookingModeLabel}</div>
                    <div className="bani-feature-item__description">Сценарий один для гостя и клиента, отличается только шаг с OTP.</div>
                  </div>
                </div>
                <div className="bani-feature-item">
                  <div className="bani-feature-item__icon"><CalendarOutlined /></div>
                  <div className="bani-feature-item__copy">
                    <div className="bani-feature-item__title">{depositSummary}</div>
                    <div className="bani-feature-item__description">Если по объекту нужен залог, вы видите это до перехода к созданной брони.</div>
                  </div>
                </div>
              </div>
            </div>
          </Card>

          <Card>
            <Alert
              type="warning"
              showIcon
              icon={<CalendarOutlined />}
              title="Оплата после создания брони"
              description="После подтверждения телефона бронь появляется в кабинете клиента, где вы сможете завершить оплату и дальнейшие действия."
            />
          </Card>
        </div>
      </div>
    </div>
  )
}
