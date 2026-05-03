import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  App,
  Button,
  Card,
  Checkbox,
  Col,
  Form,
  Input,
  InputNumber,
  List,
  Progress,
  Row,
  Select,
  Space,
  Spin,
  Steps,
  Tag,
  TimePicker,
  Typography,
} from '@/components/design/system'
import {
  ArrowLeftOutlined,
  ArrowRightOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  EyeOutlined,
  PlayCircleOutlined,
  SaveOutlined,
} from '@/components/design/icons'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import dayjs from 'dayjs'
import {
  useGetBathhousesId,
  useGetMyBathhousesIdCompleteness,
  usePostBathhouses,
  usePutBathhousesId,
} from '@/api/generated/bathhouses/bathhouses'
import { useGetCities } from '@/api/generated/cities/cities'
import {
  usePostMyListingDrafts,
  usePutMyListingDraftsIdStepStep,
  usePostMyListingDraftsIdSubmit,
} from '@/api/generated/listing-drafts/listing-drafts'
import type {
  GithubComRekurtRelaxHubInternalServiceCompletenessItem,
  InternalHandlerCreateBathhouseRequest,
  InternalHandlerWorkingHoursRequest,
} from '@/api/generated/model'
import { formatDayOfWeek } from '@/lib/format'
import { axiosInstance } from '@/api/axios-instance'
import PageHeader from '@/components/PageHeader'

const { Paragraph, Text } = Typography
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
  {
    value: 'flexible',
    label: 'Гибкая',
    description: '100% возврат за 24ч+, 50% менее 24ч до начала',
    tagColor: 'success',
    tone: 'success',
  },
  {
    value: 'moderate',
    label: 'Умеренная',
    description: '100% за 72ч+, 50% за 24-72ч, 0% менее 24ч',
    tagColor: 'warning',
    tone: 'warning',
  },
  {
    value: 'strict',
    label: 'Строгая',
    description: '100% за 7д+, 50% за 3-7д, 0% менее 3д',
    tagColor: 'error',
    tone: 'danger',
  },
] as const

const WIZARD_STEPS = [
  { title: 'Начало', icon: '👋' },
  { title: 'Информация', icon: '📝' },
  { title: 'Фото', icon: '📷' },
  { title: 'Цены', icon: '💰' },
  { title: 'Расписание', icon: '🕐' },
  { title: 'Условия', icon: '📋' },
  { title: 'Предпросмотр', icon: '👁' },
]

const SCHEDULE_PRESETS = [
  {
    label: 'Стандартная рабочая неделя (Пн-Пт 09:00–21:00)',
    days: [true, true, true, true, true, false, false],
    open: '09:00',
    close: '21:00',
  },
  {
    label: 'Каждый день (09:00–23:00)',
    days: [true, true, true, true, true, true, true],
    open: '09:00',
    close: '23:00',
  },
  {
    label: 'Выходные (Сб-Вс 10:00–22:00)',
    days: [false, false, false, false, false, true, true],
    open: '10:00',
    close: '22:00',
  },
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

function useWizardVideoUrl() {
  return useQuery({
    queryKey: ['settings', 'listing_wizard_video_url'],
    queryFn: async () => {
      const { data } = await axiosInstance.get<{ success: boolean; data: { key: string; value: string } }>('/settings/listing_wizard_video_url')
      return data.data?.value || ''
    },
    staleTime: 10 * 60 * 1000,
    retry: false,
  })
}

function useAreaAveragePrice(cityId: number | undefined) {
  return useQuery({
    queryKey: ['area-average-price', cityId],
    queryFn: async () => {
      const { data } = await axiosInstance.get<{ success: boolean; data: { average_price: number } }>(`/cities/${cityId}/average-price`)
      return data.data?.average_price ?? 0
    },
    enabled: !!cityId,
    staleTime: 5 * 60 * 1000,
    retry: false,
  })
}

function getEmbedUrl(url: string): string | null {
  if (!url) return null
  const ytMatch = url.match(/(?:youtube\.com\/(?:watch\?v=|embed\/)|youtu\.be\/)([a-zA-Z0-9_-]{11})/)
  if (ytMatch) return `https://www.youtube.com/embed/${ytMatch[1]}`
  const vimeoMatch = url.match(/(?:vimeo\.com\/(?:video\/)?|player\.vimeo\.com\/video\/)(\d+)/)
  if (vimeoMatch) return `https://player.vimeo.com/video/${vimeoMatch[1]}`
  if (url.includes('/embed/') || url.includes('player.vimeo.com')) return url
  return null
}

function CompletenessChecklist({ id }: { id: string }) {
  const { data } = useGetMyBathhousesIdCompleteness(id, {
    query: { staleTime: 30_000 },
  })

  const result = data?.data
  if (!result) return null

  const score = Math.round((result.score ?? 0) * 100)
  const items = result.items ?? []
  const requiredItems = items.filter((i) => i.required)
  const optionalItems = items.filter((i) => !i.required)

  return (
    <Card
      title="Полнота объявления"
      size="small"
      className="rh-completeness-card"
    >
      <div className="rh-completeness-summary">
        <Progress
          type="circle"
          percent={score}
          size={80}
          status={result.ready ? 'success' : 'normal'}
        />
        <div className="rh-completeness-badge">
          {result.ready ? (
            <Tag color="success">Готово к модерации</Tag>
          ) : (
            <Tag color="warning">Заполните обязательные поля</Tag>
          )}
        </div>
      </div>

      {requiredItems.length > 0 && (
        <>
          <Text strong className="rh-completeness-section-title">
            Обязательные ({result.done_required}/{result.total_required})
          </Text>
          <List
            size="small"
            dataSource={requiredItems}
            renderItem={(item: GithubComRekurtRelaxHubInternalServiceCompletenessItem) => (
              <List.Item className="rh-completeness-list-item">
                {item.complete ? (
                  <CheckCircleOutlined className="rh-completeness-icon rh-completeness-icon--done" />
                ) : (
                  <CloseCircleOutlined className="rh-completeness-icon rh-completeness-icon--required" />
                )}
                <span className={item.complete ? 'rh-completeness-label rh-completeness-label--done' : 'rh-completeness-label'}>
                  {item.label}
                </span>
              </List.Item>
            )}
          />
        </>
      )}

      {optionalItems.length > 0 && (
        <>
          <Text strong className="rh-completeness-section-title rh-completeness-section-title--optional">
            Дополнительные ({result.done_optional}/{result.total_optional})
          </Text>
          <List
            size="small"
            dataSource={optionalItems}
            renderItem={(item: GithubComRekurtRelaxHubInternalServiceCompletenessItem) => (
              <List.Item className="rh-completeness-list-item">
                {item.complete ? (
                  <CheckCircleOutlined className="rh-completeness-icon rh-completeness-icon--done" />
                ) : (
                  <CloseCircleOutlined className="rh-completeness-icon rh-completeness-icon--optional" />
                )}
                <span className={item.complete ? 'rh-completeness-label rh-completeness-label--done' : 'rh-completeness-label'}>
                  {item.label}
                </span>
              </List.Item>
            )}
          />
        </>
      )}
    </Card>
  )
}

function WelcomeStep({ videoUrl, onNext }: { videoUrl: string; onNext: () => void }) {
  const embedUrl = getEmbedUrl(videoUrl)
  return (
    <div className="rh-owner-narrow-page rh-owner-welcome">
      <Card className="rh-admin-detail-card rh-owner-welcome-card">
        <PlayCircleOutlined className="rh-owner-welcome-icon" />
        <h2 className="rh-section-heading">Как создать объявление</h2>
        <Paragraph type="secondary">
          Посмотрите короткое видео о том, как заполнить информацию о вашем объекте,
          чтобы он привлекал больше гостей.
        </Paragraph>
        {embedUrl && (
          <div className="rh-owner-video">
            <iframe
              src={embedUrl}
              title="Как создать объявление"
              className="rh-owner-video__frame"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowFullScreen
            />
          </div>
        )}
        <Paragraph className="rh-owner-welcome-copy">
          Процесс создания объявления включает 7 шагов:
        </Paragraph>
        <List
          className="rh-owner-welcome-list"
          size="small"
          dataSource={[
            'Информация об объекте — название, адрес, описание, удобства',
            'Фотографии — загрузите привлекательные фото',
            'Ценообразование — установите цену за час',
            'Расписание — рабочие часы по дням недели',
            'Политика отмены — условия возврата при отмене',
            'Предпросмотр — как ваше объявление увидят гости',
          ]}
          renderItem={(item, index) => (
            <List.Item className="rh-owner-welcome-list__item">
              <Text type="secondary">{index + 1}.</Text>{' '}
              <Text>{item}</Text>
            </List.Item>
          )}
        />
        <Button
          type="primary"
          size="large"
          onClick={onNext}
        >
          Начать заполнение
        </Button>
      </Card>
    </div>
  )
}

function ObjectInfoStep({ cities }: { cities: { id?: number; name?: string }[] }) {
  return (
    <>
      <Card title="Основная информация" className="rh-admin-detail-card">
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
                {cities.map((city) => (
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
              <InputNumber className="rh-full-width" step={0.0001} placeholder="55.7558" />
            </Form.Item>
          </Col>
          <Col xs={24} sm={12}>
            <Form.Item name="longitude" label="Долгота">
              <InputNumber className="rh-full-width" step={0.0001} placeholder="37.6173" />
            </Form.Item>
          </Col>
        </Row>
      </Card>

      <Card title="Параметры" className="rh-admin-detail-card">
        <Row gutter={16}>
          <Col xs={24} sm={8}>
            <Form.Item
              name="min_duration"
              label="Мин. длительность (ч)"
              rules={[{ required: true, message: 'Укажите длительность' }]}
            >
              <InputNumber min={1} max={24} className="rh-full-width" />
            </Form.Item>
          </Col>
          <Col xs={24} sm={8}>
            <Form.Item
              name="max_guests"
              label="Макс. гостей"
              rules={[{ required: true, message: 'Укажите кол-во' }]}
            >
              <InputNumber min={1} max={100} className="rh-full-width" />
            </Form.Item>
          </Col>
        </Row>
      </Card>

      <Card title="Удобства" className="rh-admin-detail-card">
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
    </>
  )
}

function PhotosStep() {
  return (
    <Card title="Фотографии" className="rh-admin-detail-card">
      <Paragraph type="secondary" className="rh-form-section-note">
        Добавьте URL фотографий вашего объекта. Качественные фото привлекают больше гостей.
        Первое фото станет обложкой объявления.
      </Paragraph>
      <Form.Item
        name="images"
        help="По одному URL на строку. Первое фото — обложка"
      >
        <TextArea rows={6} placeholder="https://example.com/photo1.jpg&#10;https://example.com/photo2.jpg" />
      </Form.Item>
    </Card>
  )
}

function PricingStep({ cityId }: { cityId: number | undefined }) {
  const { data: avgPrice } = useAreaAveragePrice(cityId)
  const avgPriceRub = avgPrice ? (avgPrice / 100).toLocaleString('ru-RU') : null

  return (
    <Card title="Ценообразование" className="rh-admin-detail-card">
      {avgPriceRub && (
        <div className="rh-pricing-benchmark">
          <Text type="secondary">
            Средняя цена в вашем городе: <Text strong>{avgPriceRub} ₽/час</Text>
          </Text>
        </div>
      )}
      <Form.Item
        name="price_per_hour"
        label="Цена за час (руб.)"
        rules={[{ required: true, message: 'Укажите цену' }]}
      >
        <InputNumber min={0} className="rh-form-control-medium" placeholder="1500" />
      </Form.Item>
      <Form.Item
        name="security_deposit_percent"
        label="Залог (% от базовой цены)"
        help="0 — залог не требуется. Макс. 50%. Удерживается при бронировании и возвращается через 48ч после визита."
      >
        <InputNumber min={0} max={50} className="rh-form-control-medium" placeholder="0" />
      </Form.Item>
    </Card>
  )
}

function ScheduleStep({ form }: { form: ReturnType<typeof Form.useForm<BathhouseFormValues>>[0] }) {
  const applyPreset = (preset: typeof SCHEDULE_PRESETS[number]) => {
    const hours = preset.days.map((enabled) => ({
      enabled,
      open_time: dayjs(preset.open, 'HH:mm'),
      close_time: dayjs(preset.close, 'HH:mm'),
    }))
    form.setFieldsValue({ working_hours: hours })
  }

  return (
    <Card title="Расписание" className="rh-admin-detail-card">
      <div className="rh-schedule-presets">
        <Text type="secondary" className="rh-schedule-presets__label">Шаблоны:</Text>
        <Space wrap>
          {SCHEDULE_PRESETS.map((preset) => (
            <Button key={preset.label} size="small" onClick={() => applyPreset(preset)}>
              {preset.label}
            </Button>
          ))}
        </Space>
      </div>
      {Array.from({ length: 7 }, (_, day) => (
        <Row gutter={16} key={day} align="middle" className="rh-schedule-day-row">
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
              <TimePicker format="HH:mm" minuteStep={30} className="rh-full-width" />
            </Form.Item>
          </Col>
          <Col xs={9} sm={4}>
            <Form.Item name={['working_hours', day, 'close_time']} noStyle>
              <TimePicker format="HH:mm" minuteStep={30} className="rh-full-width" />
            </Form.Item>
          </Col>
        </Row>
      ))}
    </Card>
  )
}

function CancellationPolicyStep() {
  return (
    <Card title="Политика отмены" className="rh-admin-detail-card">
      <Paragraph type="secondary" className="rh-form-section-note">
        Выберите политику отмены бронирований. Более гибкая политика привлекает больше гостей.
      </Paragraph>
      <Form.Item name="cancellation_policy">
        <div>
          {CANCELLATION_POLICIES.map((policy) => (
            <Form.Item key={policy.value} noStyle name="cancellation_policy">
              {/* Render will be handled by parent Select; this shows visual comparison */}
            </Form.Item>
          ))}
          <Select className="rh-full-width">
            {CANCELLATION_POLICIES.map((p) => (
              <Select.Option key={p.value} value={p.value}>
                <div>
                  <Tag color={p.tagColor}>{p.label}</Tag>
                  <Text type="secondary" className="rh-option-description">{p.description}</Text>
                </div>
              </Select.Option>
            ))}
          </Select>
        </div>
      </Form.Item>
      <div className="rh-policy-list">
        {CANCELLATION_POLICIES.map((policy) => (
          <Card
            key={policy.value}
            size="small"
            className={`rh-policy-card rh-policy-card--${policy.tone}`}
          >
            <Text strong>{policy.label}</Text>
            <br />
            <Text type="secondary">{policy.description}</Text>
          </Card>
        ))}
      </div>
    </Card>
  )
}

function PreviewStep({ form }: { form: ReturnType<typeof Form.useForm<BathhouseFormValues>>[0] }) {
  const values = Form.useWatch([], form) as BathhouseFormValues | undefined

  if (!values) return null

  const images = values.images
    ? values.images.split('\n').map((s) => s.trim()).filter(Boolean)
    : []
  const amenities = (values.amenities ?? []).map(
    (key) => AMENITY_FIELDS.find((a) => a.key === key)?.label ?? key,
  )
  const policyInfo = CANCELLATION_POLICIES.find((p) => p.value === values.cancellation_policy)

  return (
    <div className="rh-stack">
      <Card
        className="rh-admin-detail-card rh-preview-card"
        cover={
          images.length > 0 ? (
            <div className="rh-preview-cover">
              <img
                src={images[0]}
                alt="Обложка"
                className="rh-preview-cover__image"
                onError={(e) => {
                  (e.target as HTMLImageElement).style.display = 'none'
                }}
              />
            </div>
          ) : (
            <div className="rh-preview-cover rh-preview-cover--empty">
              <Text type="secondary">Нет фотографий</Text>
            </div>
          )
        }
      >
        <h2 className="rh-preview-title">{values.name || 'Без названия'}</h2>
        {values.address && <Text type="secondary">{values.address}</Text>}
        {values.price_per_hour > 0 && (
          <div className="rh-preview-price">
            <Text strong>
              {values.price_per_hour.toLocaleString('ru-RU')} ₽/час
            </Text>
          </div>
        )}
        {values.description && (
          <Paragraph className="rh-preview-description">{values.description}</Paragraph>
        )}
        {amenities.length > 0 && (
          <div className="rh-preview-tags">
            {amenities.map((a) => (
              <Tag key={a} className="rh-detail-tag">{a}</Tag>
            ))}
          </div>
        )}
        {values.max_guests > 0 && (
          <div className="rh-preview-meta">
            <Text type="secondary">До {values.max_guests} гостей</Text>
          </div>
        )}
        {policyInfo && (
          <div className="rh-preview-meta">
            <Tag color={policyInfo.tagColor}>Отмена: {policyInfo.label}</Tag>
          </div>
        )}
      </Card>
      <Card size="small" className="rh-admin-detail-card rh-preview-note-card">
        <EyeOutlined className="rh-inline-icon" />
        <Text type="secondary">
          Так ваше объявление будет выглядеть для гостей. Проверьте, что всё заполнено корректно, и нажмите «Создать».
        </Text>
      </Card>
    </div>
  )
}

export default function BathhouseForm() {
  const { id } = useParams<{ id: string }>()
  const isEdit = !!id
  const navigate = useNavigate()
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [form] = Form.useForm<BathhouseFormValues>()
  const [currentStep, setCurrentStep] = useState(isEdit ? 1 : 0)
  const [draftId, setDraftId] = useState<string | null>(null)
  const [draftSaving, setDraftSaving] = useState(false)
  const { data: videoUrl } = useWizardVideoUrl()

  const cityId = Form.useWatch('city_id', form) as number | undefined

  const { data: bathhouseData, isLoading: bathhouseLoading } = useGetBathhousesId(id ?? '', {
    query: { enabled: isEdit },
  })

  const { data: citiesData } = useGetCities()
  const cities = citiesData?.data ?? []

  const createDraftMutation = usePostMyListingDrafts({
    mutation: {
      onSuccess: (data) => {
        const draft = data?.data as { id?: string } | undefined
        if (draft?.id) {
          setDraftId(draft.id)
        }
      },
    },
  })

  const saveStepMutation = usePutMyListingDraftsIdStepStep()

  const submitDraftMutation = usePostMyListingDraftsIdSubmit({
    mutation: {
      onSuccess: () => {
        message.success('Объявление отправлено на модерацию')
        queryClient.invalidateQueries({ queryKey: ['/my/bathhouses'] })
        navigate('/bathhouses')
      },
      onError: () => {
        message.error('Не удалось отправить объявление')
      },
    },
  })

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

  const saveDraft = useCallback(async () => {
    if (isEdit || !draftId) return
    setDraftSaving(true)
    try {
      const values = form.getFieldsValue()
      await saveStepMutation.mutateAsync({
        id: draftId,
        step: currentStep,
        data: { data: JSON.stringify(values) as unknown as number[] },
      })
      message.success('Черновик сохранён')
    } catch {
      message.error('Не удалось сохранить черновик')
    } finally {
      setDraftSaving(false)
    }
  }, [isEdit, draftId, currentStep, form, saveStepMutation, message])

  const stepFieldsMap: Record<number, (keyof BathhouseFormValues)[]> = useMemo(() => ({
    1: ['name', 'address', 'city_id', 'min_duration', 'max_guests'],
    2: [],
    3: ['price_per_hour'],
    4: [],
    5: [],
    6: [],
  }), [])

  const goNext = useCallback(async () => {
    const fieldsToValidate = stepFieldsMap[currentStep] ?? []
    if (fieldsToValidate.length > 0) {
      try {
        await form.validateFields(fieldsToValidate)
      } catch {
        return
      }
    }
    if (!isEdit && draftId && currentStep >= 1) {
      await saveDraft()
    }
    setCurrentStep((s) => Math.min(s + 1, WIZARD_STEPS.length - 1))
  }, [currentStep, form, stepFieldsMap, isEdit, draftId, saveDraft])

  const goBack = useCallback(() => {
    setCurrentStep((s) => Math.max(s - 1, isEdit ? 1 : 0))
  }, [isEdit])

  const onFinish = (values: BathhouseFormValues) => {
    const payload = buildRequest(values)
    if (isEdit && id) {
      updateMutation.mutate({ id, data: payload })
    } else if (draftId) {
      submitDraftMutation.mutate({ id: draftId })
    } else {
      createMutation.mutate({ data: payload })
    }
  }

  const isSaving = createMutation.isPending || updateMutation.isPending || submitDraftMutation.isPending

  // Create draft on first mount for new listings.
  // useRef sentinel prevents React StrictMode double-mount from creating two phantom drafts (audit A2.2).
  const draftCreateFiredRef = useRef(false)
  useEffect(() => {
    if (isEdit || draftId || draftCreateFiredRef.current) return
    draftCreateFiredRef.current = true
    createDraftMutation.mutate()
    // Only run once on mount
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  if (isEdit && bathhouseLoading) {
    return (
      <div className="rh-loading-block">
        <Spin size="large" />
      </div>
    )
  }

  const stepContent = () => {
    switch (currentStep) {
      case 0:
        return <WelcomeStep videoUrl={videoUrl ?? ''} onNext={goNext} />
      case 1:
        return <ObjectInfoStep cities={cities as { id?: number; name?: string }[]} />
      case 2:
        return <PhotosStep />
      case 3:
        return <PricingStep cityId={cityId} />
      case 4:
        return <ScheduleStep form={form} />
      case 5:
        return <CancellationPolicyStep />
      case 6:
        return <PreviewStep form={form} />
      default:
        return null
    }
  }

  const isFirstStep = currentStep === 0
  const isLastStep = currentStep === WIZARD_STEPS.length - 1
  const showNavigation = currentStep > 0

  const progressPercent = Math.round((currentStep / (WIZARD_STEPS.length - 1)) * 100)

  const formContent = (
    <div className="rh-stack rh-owner-form-page">
      <PageHeader
        eyebrow="Объекты"
        title={isEdit ? 'Редактирование бани' : 'Создание объекта'}
        description={isEdit ? 'Обновите карточку объекта и проверьте готовность к модерации.' : 'Заполните карточку объекта пошагово и сохраните черновик перед публикацией.'}
        size="compact"
      />

      <Steps
        current={currentStep}
        size="small"
        className="rh-steps rh-owner-wizard-steps"
        items={WIZARD_STEPS.map((step) => ({
          title: step.title,
        }))}
      />

      {!isEdit && currentStep > 0 && (
        <div className="rh-draft-progress">
          <Progress percent={progressPercent} size="small" showInfo={false} />
          <Text type="secondary" className="rh-draft-progress__text">
            Шаг {currentStep} из {WIZARD_STEPS.length - 1}
            {draftSaving && ' — Сохранение...'}
            {draftId && !draftSaving && ' — Черновик сохранён'}
          </Text>
        </div>
      )}

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
          security_deposit_percent: 0,
        }}
      >
        {stepContent()}

        {showNavigation && (
          <div className="rh-form-navigation">
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={goBack}
              disabled={isFirstStep || (isEdit && currentStep <= 1)}
            >
              Назад
            </Button>
            <Space>
              {!isEdit && draftId && currentStep >= 1 && (
                <Button
                  icon={<SaveOutlined />}
                  onClick={saveDraft}
                  loading={draftSaving}
                >
                  Сохранить черновик
                </Button>
              )}
              {isLastStep ? (
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={isSaving}
                >
                  {isEdit ? 'Сохранить' : 'Создать'}
                </Button>
              ) : (
                <Button
                  type="primary"
                  icon={<ArrowRightOutlined />}
                  onClick={goNext}
                >
                  Далее
                </Button>
              )}
              <Button onClick={() => navigate('/bathhouses')}>Отмена</Button>
            </Space>
          </div>
        )}
      </Form>
    </div>
  )

  if (isEdit && id) {
    return (
      <Row gutter={[24, 24]} className="rh-owner-edit-layout">
        <Col xs={24} lg={16}>{formContent}</Col>
        <Col xs={24} lg={8}>
          <CompletenessChecklist id={id} />
        </Col>
      </Row>
    )
  }

  return formContent
}
