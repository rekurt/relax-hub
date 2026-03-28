import { useEffect } from 'react'
import {
  App,
  Button,
  Card,
  Checkbox,
  Col,
  Form,
  Input,
  InputNumber,
  Row,
  Select,
  Spin,
  TimePicker,
  Typography,
} from 'antd'
import { useNavigate, useParams } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import dayjs from 'dayjs'
import {
  useGetBathhousesId,
  usePostBathhouses,
  usePutBathhousesId,
} from '@/api/generated/bathhouses/bathhouses'
import { useGetCities } from '@/api/generated/cities/cities'
import type {
  InternalHandlerCreateBathhouseRequest,
  InternalHandlerWorkingHoursRequest,
} from '@/api/generated/model'
import { formatDayOfWeek } from '@/lib/format'

const { Title } = Typography
const { TextArea } = Input

const AMENITY_FIELDS = [
  { key: 'has_sauna', label: 'Сауна' },
  { key: 'has_steam_room', label: 'Парная' },
  { key: 'has_pool', label: 'Бассейн' },
  { key: 'has_hot_tub', label: 'Купель' },
  { key: 'has_bbq', label: 'Мангал' },
  { key: 'has_karaoke', label: 'Караоке' },
] as const

interface WorkingHoursFormItem {
  enabled: boolean
  open_time: dayjs.Dayjs | null
  close_time: dayjs.Dayjs | null
}

interface BathhouseFormValues {
  name: string
  description: string
  address: string
  city_id: number
  latitude: number
  longitude: number
  price_per_hour: number
  min_duration: number
  max_guests: number
  cancellation_policy: string
  security_deposit_percent: number
  images: string
  amenities: string[]
  working_hours: WorkingHoursFormItem[]
}

const CANCELLATION_POLICIES = [
  { value: 'flexible', label: 'Гибкая — 100% за 24ч+, 50% менее 24ч' },
  { value: 'moderate', label: 'Умеренная — 100% за 72ч+, 50% за 24-72ч, 0% менее 24ч' },
  { value: 'strict', label: 'Строгая — 100% за 7д+, 50% за 3-7д, 0% менее 3д' },
]

function buildWorkingHours(values: BathhouseFormValues): InternalHandlerWorkingHoursRequest[] {
  return values.working_hours
    .map((wh, index) => ({
      day_of_week: index,
      open_time: wh.open_time?.format('HH:mm') ?? '09:00',
      close_time: wh.close_time?.format('HH:mm') ?? '21:00',
      enabled: wh.enabled,
    }))
    .filter((wh) => wh.enabled)
    .map(({ enabled: _, ...rest }) => rest)
}

function buildRequest(values: BathhouseFormValues): InternalHandlerCreateBathhouseRequest {
  const amenitySet = new Set(values.amenities ?? [])
  return {
    name: values.name,
    description: values.description,
    address: values.address,
    city_id: values.city_id,
    latitude: values.latitude,
    longitude: values.longitude,
    price_per_hour: Math.round(values.price_per_hour * 100),
    min_duration: values.min_duration,
    max_guests: values.max_guests,
    images: values.images
      ? values.images
          .split('\n')
          .map((s) => s.trim())
          .filter(Boolean)
      : [],
    has_sauna: amenitySet.has('has_sauna'),
    has_steam_room: amenitySet.has('has_steam_room'),
    has_pool: amenitySet.has('has_pool'),
    has_hot_tub: amenitySet.has('has_hot_tub'),
    has_bbq: amenitySet.has('has_bbq'),
    has_karaoke: amenitySet.has('has_karaoke'),
    working_hours: buildWorkingHours(values),
    ...(values.cancellation_policy ? { cancellation_policy: values.cancellation_policy } : {}),
    ...(values.security_deposit_percent != null ? { security_deposit_percent: values.security_deposit_percent } : {}),
  } as InternalHandlerCreateBathhouseRequest
}

const DEFAULT_WORKING_HOURS: WorkingHoursFormItem[] = Array.from({ length: 7 }, () => ({
  enabled: true,
  open_time: dayjs('09:00', 'HH:mm'),
  close_time: dayjs('21:00', 'HH:mm'),
}))

export default function BathhouseForm() {
  const { id } = useParams<{ id: string }>()
  const isEdit = !!id
  const navigate = useNavigate()
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [form] = Form.useForm<BathhouseFormValues>()

  const { data: bathhouseData, isLoading: bathhouseLoading } = useGetBathhousesId(id ?? '', {
    query: { enabled: isEdit },
  })

  const { data: citiesData } = useGetCities()
  const cities = citiesData?.data ?? []

  const createMutation = usePostBathhouses({
    mutation: {
      onSuccess: () => {
        message.success('Баня создана')
        queryClient.invalidateQueries({ queryKey: ['/my/bathhouses'] })
        navigate('/bathhouses')
      },
      onError: () => {
        message.error('Не удалось создать баню')
      },
    },
  })

  const updateMutation = usePutBathhousesId({
    mutation: {
      onSuccess: () => {
        message.success('Баня обновлена')
        queryClient.invalidateQueries({ queryKey: ['/my/bathhouses'] })
        navigate('/bathhouses')
      },
      onError: () => {
        message.error('Не удалось обновить баню')
      },
    },
  })

  useEffect(() => {
    if (!isEdit || !bathhouseData?.data) return
    const b = bathhouseData.data
    const amenities: string[] = []
    for (const { key } of AMENITY_FIELDS) {
      if (b[key]) amenities.push(key)
    }

    const workingHours: WorkingHoursFormItem[] = DEFAULT_WORKING_HOURS.map((def, index) => {
      const existing = b.working_hours?.find((wh) => wh.day_of_week === index)
      if (existing) {
        return {
          enabled: true,
          open_time: dayjs(existing.open_time, 'HH:mm'),
          close_time: dayjs(existing.close_time, 'HH:mm'),
        }
      }
      return { ...def, enabled: false }
    })

    form.setFieldsValue({
      name: b.name,
      description: b.description,
      address: b.address,
      city_id: b.city_id,
      latitude: b.latitude,
      longitude: b.longitude,
      price_per_hour: (b.price_per_hour ?? 0) / 100,
      min_duration: b.min_duration,
      max_guests: b.max_guests,
      images: b.images?.join('\n') ?? '',
      cancellation_policy: (b as Record<string, unknown>).cancellation_policy as string ?? 'flexible',
      security_deposit_percent: (b as Record<string, unknown>).security_deposit_percent as number ?? 0,
      amenities,
      working_hours: workingHours,
    })
  }, [bathhouseData, isEdit, form])

  const onFinish = (values: BathhouseFormValues) => {
    const payload = buildRequest(values)
    if (isEdit && id) {
      updateMutation.mutate({ id, data: payload })
    } else {
      createMutation.mutate({ data: payload })
    }
  }

  const isSaving = createMutation.isPending || updateMutation.isPending

  if (isEdit && bathhouseLoading) {
    return (
      <div style={{ textAlign: 'center', padding: 48 }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div>
      <Title level={3}>{isEdit ? 'Редактирование бани' : 'Новая баня'}</Title>
      <Form
        form={form}
        layout="vertical"
        onFinish={onFinish}
        initialValues={{
          min_duration: 1,
          max_guests: 10,
          cancellation_policy: 'flexible',
          working_hours: DEFAULT_WORKING_HOURS,
          amenities: [],
        }}
        style={{ maxWidth: 800 }}
      >
        <Card title="Основная информация" style={{ marginBottom: 24 }}>
          <Form.Item
            name="name"
            label="Название"
            rules={[{ required: true, message: 'Введите название' }]}
          >
            <Input placeholder="Например: Баня на Пушкина" />
          </Form.Item>
          <Form.Item name="description" label="Описание">
            <TextArea rows={4} placeholder="Описание бани" />
          </Form.Item>
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item
                name="address"
                label="Адрес"
                rules={[{ required: true, message: 'Введите адрес' }]}
              >
                <Input placeholder="ул. Пушкина, д. 10" />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item
                name="city_id"
                label="Город"
                rules={[{ required: true, message: 'Выберите город' }]}
              >
                <Select placeholder="Выберите город">
                  {(cities as { id?: number; name?: string }[]).map((city) => (
                    <Select.Option key={city.id} value={city.id}>
                      {city.name}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item name="latitude" label="Широта">
                <InputNumber style={{ width: '100%' }} step={0.0001} placeholder="55.7558" />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="longitude" label="Долгота">
                <InputNumber style={{ width: '100%' }} step={0.0001} placeholder="37.6173" />
              </Form.Item>
            </Col>
          </Row>
        </Card>

        <Card title="Параметры" style={{ marginBottom: 24 }}>
          <Row gutter={16}>
            <Col xs={24} sm={8}>
              <Form.Item
                name="price_per_hour"
                label="Цена за час (руб.)"
                rules={[{ required: true, message: 'Укажите цену' }]}
              >
                <InputNumber min={0} style={{ width: '100%' }} placeholder="1500" />
              </Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item
                name="min_duration"
                label="Мин. длительность (ч)"
                rules={[{ required: true, message: 'Укажите длительность' }]}
              >
                <InputNumber min={1} max={24} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item
                name="max_guests"
                label="Макс. гостей"
                rules={[{ required: true, message: 'Укажите кол-во' }]}
              >
                <InputNumber min={1} max={100} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
        </Card>

        <Card title="Политика отмены" style={{ marginBottom: 24 }}>
          <Form.Item
            name="cancellation_policy"
            label="Условия возврата при отмене бронирования"
          >
            <Select>
              {CANCELLATION_POLICIES.map((p) => (
                <Select.Option key={p.value} value={p.value}>
                  {p.label}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item
            name="security_deposit_percent"
            label="Залог (% от базовой цены)"
            help="0 — залог не требуется. Макс. 50%. Удерживается при бронировании и возвращается через 48ч после визита."
          >
            <InputNumber min={0} max={50} style={{ width: '100%' }} placeholder="0" />
          </Form.Item>
        </Card>

        <Card title="Удобства" style={{ marginBottom: 24 }}>
          <Form.Item name="amenities">
            <Checkbox.Group>
              <Row gutter={[16, 8]}>
                {AMENITY_FIELDS.map(({ key, label }) => (
                  <Col xs={12} sm={8} key={key}>
                    <Checkbox value={key}>{label}</Checkbox>
                  </Col>
                ))}
              </Row>
            </Checkbox.Group>
          </Form.Item>
        </Card>

        <Card title="Рабочие часы" style={{ marginBottom: 24 }}>
          {Array.from({ length: 7 }, (_, day) => (
            <Row gutter={16} key={day} align="middle" style={{ marginBottom: 8 }}>
              <Col xs={6} sm={4}>
                <Form.Item
                  name={['working_hours', day, 'enabled']}
                  valuePropName="checked"
                  noStyle
                >
                  <Checkbox>{formatDayOfWeek(day, true)}</Checkbox>
                </Form.Item>
              </Col>
              <Col xs={9} sm={4}>
                <Form.Item name={['working_hours', day, 'open_time']} noStyle>
                  <TimePicker format="HH:mm" minuteStep={30} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
              <Col xs={9} sm={4}>
                <Form.Item name={['working_hours', day, 'close_time']} noStyle>
                  <TimePicker format="HH:mm" minuteStep={30} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
            </Row>
          ))}
        </Card>

        <Card title="Изображения" style={{ marginBottom: 24 }}>
          <Form.Item
            name="images"
            help="По одному URL на строку"
          >
            <TextArea rows={4} placeholder="https://example.com/photo1.jpg" />
          </Form.Item>
        </Card>

        <Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            loading={isSaving}
            style={{ marginRight: 12 }}
          >
            {isEdit ? 'Сохранить' : 'Создать'}
          </Button>
          <Button onClick={() => navigate('/bathhouses')}>Отмена</Button>
        </Form.Item>
      </Form>
    </div>
  )
}
