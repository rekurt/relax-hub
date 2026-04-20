import { useEffect, useRef, useState, type ReactNode } from 'react'
import {
  Typography,
  Card,
  Form,
  Input,
  Button,
  Upload,
  Avatar,
  Space,
  Switch,
  Popconfirm,
  Select,
  List,
  App,
  Skeleton,
  Row,
  Col,
  Modal,
  Alert,
  Tag,
  Result,
  Divider,
} from 'antd'
import {
  UserOutlined,
  UploadOutlined,
  DeleteOutlined,
  CalendarOutlined,
  StarOutlined,
  WalletOutlined,
  TrophyOutlined,
  CreditCardOutlined,
  ExclamationCircleOutlined,
  SwapOutlined,
  GlobalOutlined,
  BellOutlined,
  MailOutlined,
  MobileOutlined,
} from '@ant-design/icons'
import { useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth'
import {
  usePutAuthMe,
  usePostAuthMeAvatar,
  useDeleteAuthMeAvatar,
  usePostAuthDeleteAccount,
  usePostAuthRestoreAccount,
} from '@/api/generated/auth/auth'
import {
  useGetMyNotificationPreferences,
  usePutMyNotificationPreferences,
} from '@/api/generated/notifications/notifications'
import {
  useGetAuthMeSocialAccounts,
  useDeleteAuthLinkProvider,
} from '@/api/generated/oauth/oauth'
import { useGetCities } from '@/api/generated/cities/cities'
import { useGetMyStats } from '@/api/generated/users/users'
import { useGetMyRegion, usePutMyRegion } from '@/api/generated/region/region'
import { PLATFORM_NAME } from '@/content/support'
import { formatPrice } from '@/lib/format'
import { PROVIDER_LABELS, PROVIDER_COLORS } from '@/lib/constants'
import ProfileCompleteness from '@/components/ProfileCompleteness'
import PageHeader from '@/components/PageHeader'

const REGION_LABELS: Record<string, string> = {
  RU: 'Россия',
  BY: 'Беларусь',
}

const REGION_CURRENCIES: Record<string, string> = {
  RU: 'RUB',
  BY: 'BYN',
}

const DELIVERY_CHANNELS = [
  {
    name: 'in_app',
    label: 'В приложении',
    description: 'Служебные события прямо в личном кабинете.',
    eyebrow: 'Основной канал',
    icon: BellOutlined,
  },
  {
    name: 'email',
    label: 'Email',
    description: 'Краткие письма с важными обновлениями.',
    eyebrow: 'Спокойный контур',
    icon: MailOutlined,
  },
  {
    name: 'push',
    label: 'Push-уведомления',
    description: 'Мгновенные напоминания на устройстве.',
    eyebrow: 'Срочный сигнал',
    icon: MobileOutlined,
  },
]

const EVENT_SETTINGS = [
  {
    name: 'booking_events',
    label: 'Бронирования',
    description: 'Подтверждения, изменения и напоминания.',
    eyebrow: 'Операционный контур',
    icon: CalendarOutlined,
  },
  {
    name: 'review_events',
    label: 'Отзывы',
    description: 'Новые отзывы и ответы на них.',
    eyebrow: 'Обратная связь',
    icon: StarOutlined,
  },
  {
    name: 'promo_events',
    label: 'Промокоды',
    description: 'Скидки, акции и персональные предложения.',
    eyebrow: 'Маркетинг',
    icon: TrophyOutlined,
  },
  {
    name: 'reminders',
    label: 'Напоминания',
    description: 'Заблаговременные напоминания о визите.',
    eyebrow: 'Контроль визита',
    icon: BellOutlined,
  },
]

const SOCIAL_PROVIDERS = ['vk', 'yandex', 'google'] as const

const { Text, Paragraph, Title } = Typography

interface ToggleRowProps {
  name: string
  label: string
  description: string
  eyebrow?: string
  icon?: ReactNode
}

function ToggleRow({ name, label, description, eyebrow, icon }: ToggleRowProps) {
  return (
    <div className="bani-notification-toggle">
      <div className="bani-notification-toggle__copy">
        <div className="bani-notification-toggle__meta">
          {icon ? <span className="bani-notification-toggle__icon">{icon}</span> : null}
          {eyebrow ? <span className="bani-notification-toggle__eyebrow">{eyebrow}</span> : null}
        </div>
        <Text strong className="bani-notification-toggle__title">{label}</Text>
        <Text type="secondary" className="bani-notification-toggle__description">{description}</Text>
      </div>
      <Form.Item name={name} valuePropName="checked" noStyle>
        <Switch />
      </Form.Item>
    </div>
  )
}

interface MetricTileProps {
  icon: ReactNode
  label: string
  value: ReactNode
}

function MetricTile({ icon, label, value }: MetricTileProps) {
  return (
    <div className="bani-profile-metric">
      <div className="bani-profile-metric__label">
        {icon}
        <span>{label}</span>
      </div>
      <div className="bani-profile-metric__value">{value}</div>
    </div>
  )
}

export default function ClientProfile() {
  const navigate = useNavigate()
  const { user, loadProfile } = useAuthStore()
  const { message, modal } = App.useApp()
  const queryClient = useQueryClient()
  const [profileForm] = Form.useForm()
  const [prefsForm] = Form.useForm()
  const [avatarUploading, setAvatarUploading] = useState(false)
  const [deletionPending, setDeletionPending] = useState(false)
  const [regionSwitchModalOpen, setRegionSwitchModalOpen] = useState(false)
  const [regionError, setRegionError] = useState<string | null>(null)

  const heroSectionRef = useRef<HTMLDivElement>(null)
  const profileSectionRef = useRef<HTMLDivElement>(null)
  const notificationsSectionRef = useRef<HTMLDivElement>(null)

  const { data: citiesData } = useGetCities()
  const cities = citiesData?.data ?? []

  const { data: statsData, isLoading: statsLoading } = useGetMyStats()
  const stats = statsData?.data

  const { data: prefsData, isLoading: prefsLoading } = useGetMyNotificationPreferences()
  const prefs = prefsData?.data
  const watchedPrefs = Form.useWatch([], prefsForm) as Record<string, boolean | undefined> | undefined

  const { data: socialData, isLoading: socialLoading } = useGetAuthMeSocialAccounts()
  const socialAccounts = socialData?.data ?? []

  const { data: regionData, isLoading: regionLoading } = useGetMyRegion()
  const currentRegion = regionData?.data?.region ?? user?.region ?? 'RU'
  const currentCityName = cities.find((city) => city.id === user?.city_id)?.name
  const activeDeliveryChannels = DELIVERY_CHANNELS.filter((channel) => watchedPrefs?.[channel.name]).length
  const activeEventScenarios = EVENT_SETTINGS.filter((event) => watchedPrefs?.[event.name]).length
  const promoSignal = watchedPrefs?.promo_events ? 'Включен' : 'Приглушён'

  useEffect(() => {
    if (user) {
      profileForm.setFieldsValue({
        name: user.name ?? '',
        phone: user.phone ?? '',
        bio: user.bio ?? '',
        city_id: user.city_id ?? undefined,
      })
    }
  }, [user, profileForm])

  useEffect(() => {
    if (prefs) {
      prefsForm.setFieldsValue({
        in_app: prefs.in_app ?? true,
        email: prefs.email ?? false,
        push: prefs.push ?? false,
        booking_events: prefs.booking_events ?? true,
        review_events: prefs.review_events ?? true,
        promo_events: prefs.promo_events ?? true,
        reminders: prefs.reminders ?? true,
      })
    }
  }, [prefs, prefsForm])

  const updateProfile = usePutAuthMe({
    mutation: {
      onSuccess: () => {
        message.success('Профиль обновлён')
        loadProfile()
      },
      onError: () => message.error('Ошибка при обновлении профиля'),
    },
  })

  const uploadAvatar = usePostAuthMeAvatar({
    mutation: {
      onSuccess: () => {
        message.success('Аватар обновлён')
        setAvatarUploading(false)
        loadProfile()
      },
      onError: () => {
        message.error('Ошибка при загрузке аватара')
        setAvatarUploading(false)
      },
    },
  })

  const deleteAvatar = useDeleteAuthMeAvatar({
    mutation: {
      onSuccess: () => {
        message.success('Аватар удалён')
        loadProfile()
      },
      onError: () => message.error('Ошибка при удалении аватара'),
    },
  })

  const updatePrefs = usePutMyNotificationPreferences({
    mutation: {
      onSuccess: () => {
        message.success('Настройки уведомлений сохранены')
        queryClient.invalidateQueries({ queryKey: ['/my/notification-preferences'] })
      },
      onError: () => message.error('Ошибка при сохранении настроек'),
    },
  })

  const unlinkSocial = useDeleteAuthLinkProvider({
    mutation: {
      onSuccess: () => {
        message.success('Аккаунт отвязан')
        queryClient.invalidateQueries({ queryKey: ['/auth/me/social-accounts'] })
      },
      onError: () => message.error('Ошибка при отвязке аккаунта'),
    },
  })

  const deleteAccount = usePostAuthDeleteAccount({
    mutation: {
      onSuccess: () => {
        message.success('Запрос на удаление аккаунта отправлен. У вас есть 30 дней, чтобы отменить.')
        setDeletionPending(true)
      },
      onError: (error: unknown) => {
        const status = (error as { response?: { status?: number } })?.response?.status
        if (status === 409) {
          message.warning('Удаление аккаунта уже запрошено')
          setDeletionPending(true)
        } else {
          message.error('Ошибка при запросе удаления аккаунта')
        }
      },
    },
  })

  const restoreAccount = usePostAuthRestoreAccount({
    mutation: {
      onSuccess: () => {
        message.success('Удаление аккаунта отменено')
        setDeletionPending(false)
      },
      onError: (error: unknown) => {
        const status = (error as { response?: { status?: number } })?.response?.status
        if (status === 400) {
          message.info('Удаление аккаунта не было запрошено')
          setDeletionPending(false)
        } else {
          message.error('Ошибка при восстановлении аккаунта')
        }
      },
    },
  })

  const switchRegion = usePutMyRegion({
    mutation: {
      onSuccess: () => {
        message.success('Регион успешно изменён')
        setRegionSwitchModalOpen(false)
        setRegionError(null)
        queryClient.invalidateQueries({ queryKey: ['/my/region'] })
        loadProfile()
      },
      onError: (error: unknown) => {
        const status = (error as { response?: { status?: number } })?.response?.status
        const errorMessage = (error as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
        if (status === 409) {
          setRegionError(errorMessage ?? 'Невозможно сменить регион: проверьте баланс кошелька, активные бронирования, открытые споры и неактивированные сертификаты')
        } else if (status === 400) {
          setRegionError(errorMessage ?? 'Некорректный регион или вы уже в этом регионе')
        } else {
          setRegionError('Ошибка при смене региона')
        }
      },
    },
  })

  const targetRegion = currentRegion === 'RU' ? 'BY' : 'RU'
  const currentRegionLabel = REGION_LABELS[currentRegion] ?? currentRegion
  const currentCurrency = REGION_CURRENCIES[currentRegion] ?? currentRegion
  const targetRegionLabel = REGION_LABELS[targetRegion] ?? targetRegion

  const scrollToSection = (sectionRef: { current: HTMLDivElement | null }) => {
    sectionRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  const handleCompletenessNavigate = (field: string) => {
    if (field === 'avatar') {
      scrollToSection(heroSectionRef)
      return
    }

    if (field === 'preferences') {
      navigate('/client/preferences')
      return
    }

    if (field === 'notification_settings') {
      navigate('/client/notification-preferences')
      return
    }

    scrollToSection(profileSectionRef)
  }

  const handleDeleteAccount = () => {
    modal.confirm({
      title: 'Удаление аккаунта',
      icon: <ExclamationCircleOutlined />,
      content: (
        <div>
          <p>Вы уверены, что хотите удалить аккаунт?</p>
          <Alert
            type="warning"
            showIcon
            style={{ marginBottom: 12 }}
            message="Период восстановления — 30 дней"
            description="После подтверждения ваш аккаунт будет деактивирован. В течение 30 дней вы можете отменить удаление."
          />
          <Alert
            type="info"
            showIcon
            message="Что произойдёт с вашими средствами"
            description="Средства, внесённые пополнением, будут возвращены. Бонусы и промо-баланс будут утеряны."
          />
        </div>
      ),
      okText: 'Удалить аккаунт',
      okType: 'danger',
      cancelText: 'Отмена',
      onOk: () => deleteAccount.mutate(),
    })
  }

  const handleRegionSwitch = () => {
    setRegionError(null)
    switchRegion.mutate({ data: { region: targetRegion } })
  }

  const handleProfileSubmit = (values: { name?: string; phone?: string; bio?: string; city_id?: number }) => {
    updateProfile.mutate({ data: values })
  }

  const handlePrefsSubmit = (values: Record<string, boolean>) => {
    updatePrefs.mutate({ data: values })
  }

  const handleAvatarUpload = (file: File) => {
    setAvatarUploading(true)
    uploadAvatar.mutate({ data: { avatar: file } })
    return false
  }

  const handleLinkProvider = (provider: string) => {
    window.location.assign(`/api/v1/auth/oauth/${provider}?action=link`)
  }

  const linkedProviders = new Set(socialAccounts.map((account) => account.provider))

  return (
    <div>
      <PageHeader
        eyebrow="Личный кабинет"
        title="Мой профиль"
        description="Профиль, безопасность, уведомления и платёжные привязки собраны в единую рабочую панель."
        extra={(
          <Space wrap>
            <Button icon={<StarOutlined />} onClick={() => navigate('/client/preferences')}>
              Предпочтения
            </Button>
          </Space>
        )}
      />

      <div className="bani-profile-layout">
        <div className="bani-profile-layout__main">
          <div ref={heroSectionRef}>
            <Card
              className="bani-profile-avatar-card"
              title="Аватар"
              extra={
                <Space wrap>
                  <Upload
                    showUploadList={false}
                    accept="image/jpeg,image/png"
                    beforeUpload={handleAvatarUpload}
                  >
                    <Button icon={<UploadOutlined />} loading={avatarUploading}>
                      Загрузить
                    </Button>
                  </Upload>
                  {user?.avatar_url && (
                    <Popconfirm
                      title="Удалить аватар?"
                      onConfirm={() => deleteAvatar.mutate()}
                      okText="Да"
                      cancelText="Нет"
                    >
                      <Button danger icon={<DeleteOutlined />} loading={deleteAvatar.isPending}>
                        Удалить
                      </Button>
                    </Popconfirm>
                  )}
                </Space>
              }
            >
              <div className="bani-profile-hero">
                <div className="bani-profile-hero__identity">
                  <Avatar
                    size={96}
                    src={user?.avatar_url}
                    icon={!user?.avatar_url && <UserOutlined />}
                  />
                  <div className="bani-profile-hero__copy">
                    <Space wrap size={8}>
                      <Tag color="geekblue">Клиент</Tag>
                      <Tag color="blue">Регион {currentCurrency}</Tag>
                      {currentCityName && <Tag color="gold">{currentCityName}</Tag>}
                    </Space>

                    <Title level={3} style={{ marginBottom: 0, marginTop: 0 }}>
                      {user?.name ?? 'Профиль клиента'}
                    </Title>

                    <Space direction="vertical" size={4} style={{ width: '100%' }}>
                      <Text type="secondary">{user?.email ?? 'Email не указан'}</Text>
                      <Text type="secondary">{user?.phone ?? 'Телефон не указан'}</Text>
                      {user?.bio ? (
                        <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                          {user.bio}
                        </Paragraph>
                      ) : (
                        <Text type="secondary">
                          Добавьте короткое описание о себе, чтобы профиль выглядел аккуратнее.
                        </Text>
                      )}
                    </Space>
                  </div>
                </div>

                <div className="bani-profile-hero__side">
                  <Text type="secondary">
                    Данные ниже можно менять без лишних переходов. Кликай по заполненности профиля, чтобы сразу перейти к нужному разделу.
                  </Text>
                </div>
              </div>
            </Card>
          </div>

          <ProfileCompleteness onNavigate={handleCompletenessNavigate} />

          <div className="bani-profile-stack">
            <div ref={profileSectionRef}>
              <Card title="Основная информация">
                <Form
                  form={profileForm}
                  layout="vertical"
                  onFinish={handleProfileSubmit}
                >
                  <Row gutter={[16, 0]}>
                    <Col xs={24} md={12}>
                      <Form.Item label="Email">
                        <Input value={user?.email ?? ''} disabled />
                      </Form.Item>
                    </Col>
                    <Col xs={24} md={12}>
                      <Form.Item label="Имя" name="name">
                        <Input placeholder="Ваше имя" />
                      </Form.Item>
                    </Col>
                    <Col xs={24} md={12}>
                      <Form.Item label="Телефон" name="phone">
                        <Input placeholder="+7 999 123-45-67" />
                      </Form.Item>
                    </Col>
                    <Col xs={24} md={12}>
                      <Form.Item label="Город" name="city_id">
                        <Select
                          placeholder="Выберите город"
                          allowClear
                          options={(cities as Array<{ id?: number; name?: string }>).map((city) => ({
                            value: city.id,
                            label: city.name,
                          }))}
                        />
                      </Form.Item>
                    </Col>
                    <Col xs={24}>
                      <Form.Item label="О себе" name="bio">
                        <Input.TextArea rows={4} placeholder="Расскажите о себе" />
                      </Form.Item>
                    </Col>
                  </Row>

                  <div className="bani-profile-form-actions">
                    <Button type="primary" htmlType="submit" loading={updateProfile.isPending}>
                      Сохранить
                    </Button>
                  </div>
                </Form>
              </Card>
            </div>

            <div ref={notificationsSectionRef}>
              <Card title="Настройки уведомлений">
                {prefsLoading ? (
                  <Skeleton active />
                ) : (
                  <Form
                    form={prefsForm}
                    layout="vertical"
                    onFinish={handlePrefsSubmit}
                  >
                    <div className="bani-notification-prefs">
                      <section className="bani-notification-prefs__hero">
                        <div className="bani-notification-prefs__hero-copy">
                          <div className="bani-notification-prefs__eyebrow">Контур уведомлений</div>
                          <Title level={3} className="bani-notification-prefs__title">
                            Только нужные сигналы
                          </Title>
                          <Paragraph className="bani-notification-prefs__description">
                            Выберите, через какие каналы {PLATFORM_NAME} может связываться с вами, и оставьте
                            включёнными только те сценарии, которые действительно требуют внимания.
                          </Paragraph>
                        </div>

                        <div className="bani-notification-prefs__stats">
                          <div className="bani-notification-prefs__stat">
                            <span>Активных каналов</span>
                            <strong>{activeDeliveryChannels}/3</strong>
                          </div>
                          <div className="bani-notification-prefs__stat">
                            <span>Активных сценариев</span>
                            <strong>{activeEventScenarios}/4</strong>
                          </div>
                          <div className="bani-notification-prefs__stat">
                            <span>Промо-поток</span>
                            <strong>{promoSignal}</strong>
                          </div>
                        </div>
                      </section>

                      <div className="bani-notification-prefs__section">
                        <div className="bani-notification-prefs__section-header">
                          <div className="bani-notification-prefs__section-eyebrow">Каналы связи</div>
                          <Text strong className="bani-notification-prefs__section-title">Куда отправлять общие уведомления</Text>
                          <Text type="secondary" className="bani-notification-prefs__section-description">
                            Сначала настройте базовые каналы. После этого ниже можно выбрать,
                            какие типы событий будут в них попадать.
                          </Text>
                        </div>

                        <div className="bani-notification-prefs__grid">
                          {DELIVERY_CHANNELS.map((channel) => (
                            <ToggleRow
                              key={channel.name}
                              name={channel.name}
                              label={channel.label}
                              description={channel.description}
                              eyebrow={channel.eyebrow}
                              icon={<channel.icon />}
                            />
                          ))}
                        </div>
                      </div>

                      <div className="bani-notification-prefs__section">
                        <div className="bani-notification-prefs__section-header">
                          <div className="bani-notification-prefs__section-eyebrow">Сценарии</div>
                          <Text strong className="bani-notification-prefs__section-title">Какие события должны доходить до вас</Text>
                          <Text type="secondary" className="bani-notification-prefs__section-description">
                            Операционные и сервисные события держите включёнными, а маркетинговые
                            сигналы можно приглушить без потери важной информации.
                          </Text>
                        </div>

                        <div className="bani-notification-prefs__grid bani-notification-prefs__grid--events">
                          {EVENT_SETTINGS.map((event) => (
                            <ToggleRow
                              key={event.name}
                              name={event.name}
                              label={event.label}
                              description={event.description}
                              eyebrow={event.eyebrow}
                              icon={<event.icon />}
                            />
                          ))}
                        </div>
                      </div>

                      <div className="bani-notification-prefs__footer">
                        <Text type="secondary">
                          Изменения применяются только к вашему профилю и не затрагивают системные уведомления.
                        </Text>
                        <div className="bani-profile-form-actions">
                          <Button type="primary" htmlType="submit" loading={updatePrefs.isPending}>
                            Сохранить настройки
                          </Button>
                        </div>
                      </div>
                    </div>
                  </Form>
                )}
              </Card>
            </div>

            <Card title="Привязанные аккаунты">
              {socialLoading ? (
                <Skeleton active />
              ) : (
                <>
                  <List
                    dataSource={socialAccounts}
                    locale={{ emptyText: 'Нет привязанных аккаунтов' }}
                    renderItem={(account) => (
                      <List.Item
                        actions={[
                          <Popconfirm
                            key="unlink"
                            title={`Отвязать ${PROVIDER_LABELS[account.provider ?? ''] ?? account.provider}?`}
                            onConfirm={() => {
                              if (account.provider) unlinkSocial.mutate({ provider: account.provider })
                            }}
                            okText="Да"
                            cancelText="Нет"
                          >
                            <Button danger size="small">
                              Отвязать
                            </Button>
                          </Popconfirm>,
                        ]}
                      >
                        <List.Item.Meta
                          avatar={
                            <Avatar
                              src={account.avatar_url}
                              style={{
                                backgroundColor: PROVIDER_COLORS[account.provider ?? ''] ?? '#999',
                              }}
                            >
                              {(account.provider ?? '')[0]?.toUpperCase()}
                            </Avatar>
                          }
                          title={
                            <Text>
                              {PROVIDER_LABELS[account.provider ?? ''] ?? account.provider}
                            </Text>
                          }
                          description={account.email ?? account.name ?? ''}
                        />
                      </List.Item>
                    )}
                  />

                  <Divider style={{ margin: '20px 0 16px' }} />
                  <Space wrap>
                    {SOCIAL_PROVIDERS.map((provider) =>
                      !linkedProviders.has(provider) ? (
                        <Button
                          key={provider}
                          onClick={() => handleLinkProvider(provider)}
                          style={{
                            borderColor: PROVIDER_COLORS[provider],
                            color: PROVIDER_COLORS[provider],
                          }}
                        >
                          Привязать {PROVIDER_LABELS[provider]}
                        </Button>
                      ) : null,
                    )}
                  </Space>
                </>
              )}
            </Card>
          </div>
        </div>

        <div className="bani-profile-layout__aside">
          <Card title="Статистика" loading={statsLoading}>
            <Row gutter={[12, 12]}>
              <Col xs={12}>
                <MetricTile
                  icon={<CalendarOutlined />}
                  label="Визиты"
                  value={stats?.total_visits ?? 0}
                />
              </Col>
              <Col xs={12}>
                <MetricTile
                  icon={<StarOutlined />}
                  label="Отзывы"
                  value={stats?.review_count ?? 0}
                />
              </Col>
              <Col xs={12}>
                <MetricTile
                  icon={<WalletOutlined />}
                  label="Потрачено"
                  value={formatPrice(stats?.total_spent ?? 0)}
                />
              </Col>
              <Col xs={12}>
                <MetricTile
                  icon={<TrophyOutlined />}
                  label="Средний рейтинг"
                  value={stats?.avg_rating ? stats.avg_rating.toFixed(1) : '—'}
                />
              </Col>
            </Row>
          </Card>

          <Card
            title="Карты"
            extra={
              <Button type="link" icon={<CreditCardOutlined />} onClick={() => navigate('/client/cards')}>
                Управление картами
              </Button>
            }
          >
            <Text type="secondary">
              Удобно для быстрых оплат и повторных бронирований без повторного ввода данных карты.
            </Text>
          </Card>

          <div>
            <Card title={<Space><GlobalOutlined /> Регион</Space>}>
              {regionLoading ? (
                <Skeleton active paragraph={{ rows: 2 }} />
              ) : (
                <Space direction="vertical" size={16} style={{ width: '100%' }}>
                  <div>
                    <Text type="secondary">Текущий регион: </Text>
                    <Tag color="blue" style={{ fontSize: 14 }}>
                      {currentRegionLabel} ({currentCurrency})
                    </Tag>
                  </div>

                  <Alert
                    type="info"
                    showIcon
                    message="Смена региона"
                    description="При смене создаётся новый кошелёк в другой валюте, а старый архивируется. Смена недоступна при ненулевом балансе, активных бронированиях, спорах или неактивированных сертификатах."
                  />

                  <Button
                    icon={<SwapOutlined />}
                    onClick={() => {
                      setRegionError(null)
                      setRegionSwitchModalOpen(true)
                    }}
                  >
                    Сменить на {targetRegionLabel}
                  </Button>
                </Space>
              )}

              <Modal
                title="Подтверждение смены региона"
                open={regionSwitchModalOpen}
                onCancel={() => {
                  setRegionSwitchModalOpen(false)
                  setRegionError(null)
                }}
                onOk={handleRegionSwitch}
                okText="Подтвердить"
                cancelText="Отмена"
                confirmLoading={switchRegion.isPending}
              >
                <Space direction="vertical" size={12} style={{ width: '100%' }}>
                  <Text>
                    Вы хотите сменить регион с <strong>{currentRegionLabel}</strong> на <strong>{targetRegionLabel}</strong>?
                  </Text>
                  <Alert
                    type="warning"
                    showIcon
                    message="Последствия смены региона"
                    description={
                      <ul style={{ paddingLeft: 20, margin: 0 }}>
                        <li>Текущий кошелёк ({currentCurrency}) будет архивирован</li>
                        <li>Создан новый кошелёк в {REGION_CURRENCIES[targetRegion] ?? targetRegion}</li>
                        <li>Уровень лояльности будет сброшен</li>
                      </ul>
                    }
                  />
                  {regionError && (
                    <Alert type="error" showIcon message={regionError} />
                  )}
                </Space>
              </Modal>
            </Card>
          </div>

          <Card
            className="bani-profile-danger-card"
            title={
              <Space>
                <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />
                <Text>Удаление аккаунта</Text>
              </Space>
            }
          >
            <div>
              {deletionPending ? (
                <Result
                  status="warning"
                  title="Удаление аккаунта запрошено"
                  subTitle="Ваш аккаунт будет удалён через 30 дней. Вы можете отменить удаление в любой момент до этого срока."
                  extra={
                    <Button
                      type="primary"
                      onClick={() => restoreAccount.mutate()}
                      loading={restoreAccount.isPending}
                    >
                      Отменить удаление
                    </Button>
                  }
                />
              ) : (
                <Space direction="vertical" size={16} style={{ width: '100%' }}>
                  <Text type="secondary">
                    После подтверждения ваш аккаунт будет деактивирован. В течение 30 дней вы можете отменить удаление.
                    Средства, внесённые пополнением, будут возвращены. Бонусы и промо-баланс будут утеряны.
                  </Text>
                  <Button
                    danger
                    icon={<DeleteOutlined />}
                    onClick={handleDeleteAccount}
                    loading={deleteAccount.isPending}
                  >
                    Удалить аккаунт
                  </Button>
                </Space>
              )}
            </div>
          </Card>
        </div>
      </div>
    </div>
  )
}
