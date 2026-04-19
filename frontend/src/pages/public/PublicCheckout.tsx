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
  Empty,
  Input,
  InputNumber,
  Result,
  Row,
  Space,
  Spin,
  Steps,
  Tag,
  Typography,
} from 'antd'
import dayjs from 'dayjs'
import { CalendarOutlined, CheckCircleOutlined, ClockCircleOutlined, LockOutlined, UserOutlined } from '@ant-design/icons'
import { useGetBathhousesId, useGetBathhousesIdAvailableSlots } from '@/api/generated/bathhouses/bathhouses'
import { useGetBathhousesIdPriceCalculator } from '@/api/generated/pricing/pricing'
import { axiosInstance } from '@/api/axios-instance'
import { useAuthStore } from '@/stores/auth'
import { formatPrice } from '@/lib/format'

const { Title, Text, Paragraph } = Typography

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
  const [selectedSlot, setSelectedSlot] = useState<{ from: string; to: string } | null>(() => {
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

  const availableSlots = slots.filter((slot) => slot.available)

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
    } catch (error) {
      void error
      message.error('Не удалось отправить код')
    } finally {
      setSendingOtp(false)
    }
  }

  const handleVerifyAndBook = async () => {
    if (!otpCode.trim() || otpCode.trim().length !== 6) {
      message.warning('Введите 6-значный код')
      return
    }
    if (!selectedSlot) {
      message.warning('Слот не выбран')
      return
    }

    setSubmitting(true)
    try {
      if (currentUser?.role === 'client') {
        await createBooking()
        return
      }

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

      await createBooking()
    } catch (error) {
      const errorMessage = Axios.isAxiosError(error)
        ? error.response?.data?.error?.message
        : 'Не удалось завершить бронирование'
      message.error(errorMessage || 'Не удалось завершить бронирование')
    } finally {
      setSubmitting(false)
    }
  }

  if (bathhouseLoading) {
    return <Spin size="large" style={{ display: 'block', margin: '120px auto' }} />
  }

  if (!bathhouseId || !bathhouse) {
    return (
      <Empty
        description="Для публичного бронирования сначала выберите баню и слот"
        image={Empty.PRESENTED_IMAGE_SIMPLE}
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
    <Row gutter={[24, 24]} align="top">
      <Col xs={24} xl={16}>
        <Card style={{ borderRadius: 24, marginBottom: 16 }}>
          <Tag color="gold">Публичное бронирование</Tag>
          <Title level={1} style={{ marginTop: 12, marginBottom: 12 }}>
            {bathhouse.name}
          </Title>
          <Paragraph style={{ maxWidth: 720, fontSize: 16, marginBottom: 0 }}>
            Мы оставили только ключевой flow: выбрать слот, подтвердить телефон и сразу перейти к броне без отдельной регистрации.
          </Paragraph>
        </Card>

        <Card style={{ borderRadius: 24 }}>
          <Steps
            current={currentStep}
            items={[
              { title: 'Слот' },
              { title: currentUser?.role === 'client' ? 'Подтверждение' : 'Контакты и SMS' },
              { title: 'Готово' },
            ]}
            style={{ marginBottom: 24 }}
          />

          <Row gutter={[16, 16]}>
            <Col xs={24} lg={12}>
              <Text strong>Дата посещения</Text>
              <DatePicker
                value={dayjs(selectedDate)}
                onChange={(value) => {
                  const nextDate = value ? value.format('YYYY-MM-DD') : dayjs().format('YYYY-MM-DD')
                  setSelectedDate(nextDate)
                  setSelectedSlot(null)
                }}
                style={{ width: '100%', marginTop: 8 }}
              />
            </Col>
            <Col xs={24} lg={12}>
              <Text strong>Количество гостей</Text>
              <InputNumber
                min={1}
                max={bathhouse.max_guests ?? 20}
                value={guestCount}
                onChange={(value) => setGuestCount(value ?? 1)}
                style={{ width: '100%', marginTop: 8 }}
              />
            </Col>
          </Row>

          <div style={{ marginTop: 24 }}>
            <Text strong>Доступные слоты</Text>
            <div style={{ marginTop: 12 }}>
              <Spin spinning={slotsLoading}>
                {availableSlots.length === 0 ? (
                  <Alert
                    type="info"
                    showIcon
                    title="На выбранную дату нет доступных слотов"
                    description="Попробуйте другую дату или вернитесь в каталог, чтобы посмотреть похожие варианты."
                  />
                ) : (
                  <Space wrap size={12}>
                    {availableSlots.map((slot) => {
                      const isSelected = selectedSlot?.from === slot.startTime && selectedSlot?.to === slot.endTime
                      const fromTime = dayjs(slot.startTime).format('HH:mm')
                      const toTime = dayjs(slot.endTime).format('HH:mm')
                      return (
                        <Button
                          key={`${slot.startTime}-${slot.endTime}`}
                          type={isSelected ? 'primary' : 'default'}
                          size="large"
                          icon={<ClockCircleOutlined />}
                          onClick={() => {
                            if (slot.startTime && slot.endTime) {
                              setSelectedSlot({ from: slot.startTime, to: slot.endTime })
                            }
                          }}
                        >
                          {fromTime} - {toTime}
                        </Button>
                      )
                    })}
                  </Space>
                )}
              </Spin>
            </div>
          </div>

          <div style={{ marginTop: 24 }}>
            <Text strong>Контактные данные</Text>
            <Row gutter={[16, 16]} style={{ marginTop: 8 }}>
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

          {!currentUser && (
            <div style={{ marginTop: 24 }}>
              <Alert
                type="info"
                showIcon
                icon={<LockOutlined />}
                title="Авторизация встроена в checkout"
                description="Мы подтверждаем телефон по SMS, создаем клиентский аккаунт и сохраняем бронь в ваш личный кабинет без отдельной регистрации."
                style={{ marginBottom: 16 }}
              />
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
            <div style={{ marginTop: 24 }}>
              <Button type="primary" size="large" loading={submitting} onClick={handleVerifyAndBook}>
                Создать бронь
              </Button>
            </div>
          )}
        </Card>
      </Col>

      <Col xs={24} xl={8}>
        <Card style={{ borderRadius: 24, position: 'sticky', top: 112 }}>
          <Title level={4}>Сводка бронирования</Title>
          <Space orientation="vertical" style={{ width: '100%' }} size={12}>
            <div>
              <Text type="secondary">Баня</Text>
              <div>{bathhouse.name}</div>
            </div>
            <div>
              <Text type="secondary">Адрес</Text>
              <div>{bathhouse.address || '—'}</div>
            </div>
            <div>
              <Text type="secondary">Дата и время</Text>
              <div>
                {selectedSlot
                  ? `${dayjs(selectedSlot.from).format('D MMMM')} · ${dayjs(selectedSlot.from).format('HH:mm')} - ${dayjs(selectedSlot.to).format('HH:mm')}`
                  : 'Выберите слот'}
              </div>
            </div>
            <div>
              <Text type="secondary">Гостей</Text>
              <div>{guestCount}</div>
            </div>
            <div>
              <Text type="secondary">Цена</Text>
              <div style={{ fontSize: 28, fontWeight: 700 }}>
                {price?.final_price ? formatPrice(price.final_price) : 'Будет рассчитана после выбора слота'}
              </div>
            </div>
            <Alert
              type="warning"
              showIcon
              icon={<CalendarOutlined />}
              title="Оплата после создания брони"
              description="После подтверждения телефона бронь появляется в кабинете клиента, где вы сможете завершить оплату и дальнейшие действия."
            />
          </Space>
        </Card>
      </Col>
    </Row>
  )
}
