import { useEffect, useState } from 'react'
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
  Divider,
  Popconfirm,
  Select,
  List,
  App,
  Skeleton,
  Statistic,
  Row,
  Col,
  Modal,
  Alert,
  Tag,
  Result,
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
} from '@ant-design/icons'
import { useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
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

const { Text } = Typography

export default function ClientProfile() {
  const { user, loadProfile } = useAuthStore()
  const { message, modal } = App.useApp()
  const queryClient = useQueryClient()
  const [profileForm] = Form.useForm()
  const [prefsForm] = Form.useForm()
  const [avatarUploading, setAvatarUploading] = useState(false)
  const [deletionPending, setDeletionPending] = useState(false)
  const [regionSwitchModalOpen, setRegionSwitchModalOpen] = useState(false)
  const [regionError, setRegionError] = useState<string | null>(null)

  const { data: citiesData } = useGetCities()
  const cities = citiesData?.data ?? []

  const { data: statsData, isLoading: statsLoading } = useGetMyStats()
  const stats = statsData?.data

  const { data: prefsData, isLoading: prefsLoading } = useGetMyNotificationPreferences()
  const prefs = prefsData?.data

  const { data: socialData, isLoading: socialLoading } = useGetAuthMeSocialAccounts()
  const socialAccounts = socialData?.data ?? []

  const { data: regionData, isLoading: regionLoading } = useGetMyRegion()
  const currentRegion = regionData?.data?.region ?? user?.region ?? 'RU'

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

  const linkedProviders = new Set(socialAccounts.map((a) => a.provider))

  return (
    <div>
      <PageHeader
        eyebrow="Личный кабинет"
        title="Мой профиль"
        description="Профиль, безопасность, уведомления и платёжные привязки собраны в единую рабочую панель."
      />

      <ProfileCompleteness />

      <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
        <Card title="Статистика">
          {statsLoading ? (
            <Skeleton active />
          ) : (
            <Row gutter={[16, 16]}>
              <Col xs={12} sm={6}>
                <Statistic
                  title="Визиты"
                  value={stats?.total_visits ?? 0}
                  prefix={<CalendarOutlined />}
                />
              </Col>
              <Col xs={12} sm={6}>
                <Statistic
                  title="Отзывы"
                  value={stats?.review_count ?? 0}
                  prefix={<StarOutlined />}
                />
              </Col>
              <Col xs={12} sm={6}>
                <Statistic
                  title="Потрачено"
                  value={formatPrice(stats?.total_spent ?? 0)}
                  prefix={<WalletOutlined />}
                />
              </Col>
              <Col xs={12} sm={6}>
                <Statistic
                  title="Средний рейтинг"
                  value={stats?.avg_rating ? stats.avg_rating.toFixed(1) : '—'}
                  prefix={<TrophyOutlined />}
                />
              </Col>
            </Row>
          )}
        </Card>

        <Card title="Аватар">
          <Space size={16} align="center">
            <Avatar
              size={80}
              src={user?.avatar_url}
              icon={!user?.avatar_url && <UserOutlined />}
            />
            <Space direction="vertical">
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
          </Space>
        </Card>

        <Card title="Основная информация">
          <Form
            form={profileForm}
            layout="vertical"
            onFinish={handleProfileSubmit}
            style={{ maxWidth: 500 }}
          >
            <Form.Item label="Имя" name="name">
              <Input placeholder="Ваше имя" />
            </Form.Item>
            <Form.Item label="Телефон" name="phone">
              <Input placeholder="+7 999 123-45-67" />
            </Form.Item>
            <Form.Item label="О себе" name="bio">
              <Input.TextArea rows={3} placeholder="Расскажите о себе" />
            </Form.Item>
            <Form.Item label="Город" name="city_id">
              <Select
                placeholder="Выберите город"
                allowClear
                options={(cities as Array<{ id?: number; name?: string }>).map((c) => ({
                  value: c.id,
                  label: c.name,
                }))}
              />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" loading={updateProfile.isPending}>
                Сохранить
              </Button>
            </Form.Item>
          </Form>
        </Card>

        <Card title="Настройки уведомлений">
          {prefsLoading ? (
            <Skeleton active />
          ) : (
            <Form
              form={prefsForm}
              layout="vertical"
              onFinish={handlePrefsSubmit}
              style={{ maxWidth: 500 }}
            >
              <Divider plain>Каналы доставки</Divider>
              <Form.Item label="В приложении" name="in_app" valuePropName="checked">
                <Switch />
              </Form.Item>
              <Form.Item label="Email" name="email" valuePropName="checked">
                <Switch />
              </Form.Item>
              <Form.Item label="Push-уведомления" name="push" valuePropName="checked">
                <Switch />
              </Form.Item>

              <Divider plain>События</Divider>
              <Form.Item label="Бронирования" name="booking_events" valuePropName="checked">
                <Switch />
              </Form.Item>
              <Form.Item label="Отзывы" name="review_events" valuePropName="checked">
                <Switch />
              </Form.Item>
              <Form.Item label="Промокоды" name="promo_events" valuePropName="checked">
                <Switch />
              </Form.Item>
              <Form.Item label="Напоминания" name="reminders" valuePropName="checked">
                <Switch />
              </Form.Item>

              <Form.Item>
                <Button type="primary" htmlType="submit" loading={updatePrefs.isPending}>
                  Сохранить настройки
                </Button>
              </Form.Item>
            </Form>
          )}
        </Card>

        <Card
          title="Сохранённые карты"
          extra={
            <Link to="/client/cards">
              <Button type="link" icon={<CreditCardOutlined />}>Управление картами</Button>
            </Link>
          }
        >
          <Text type="secondary">Управляйте сохранёнными способами оплаты для быстрого бронирования.</Text>
        </Card>

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

              <Divider />
              <Space wrap>
                {(['vk', 'yandex', 'google'] as const).map((provider) =>
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

        <Card title={<Space><GlobalOutlined /> Регион</Space>}>
          {regionLoading ? (
            <Skeleton active paragraph={{ rows: 2 }} />
          ) : (
            <div style={{ maxWidth: 500 }}>
              <Space direction="vertical" size={16} style={{ width: '100%' }}>
                <div>
                  <Text type="secondary">Текущий регион: </Text>
                  <Tag color="blue" style={{ fontSize: 14 }}>
                    {REGION_LABELS[currentRegion] ?? currentRegion} ({REGION_CURRENCIES[currentRegion] ?? currentRegion})
                  </Tag>
                </div>

                <Alert
                  type="info"
                  showIcon
                  message="Смена региона"
                  description="При смене региона будет создан новый кошелёк в валюте нового региона, а старый — архивирован. Уровень лояльности будет сброшен. Смена невозможна при ненулевом балансе кошелька, активных бронированиях, открытых спорах или неактивированных сертификатах."
                />

                <Button
                  icon={<SwapOutlined />}
                  onClick={() => {
                    setRegionError(null)
                    setRegionSwitchModalOpen(true)
                  }}
                >
                  Сменить на {REGION_LABELS[targetRegion] ?? targetRegion}
                </Button>
              </Space>
            </div>
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
                Вы хотите сменить регион с <strong>{REGION_LABELS[currentRegion]}</strong> на <strong>{REGION_LABELS[targetRegion]}</strong>?
              </Text>
              <Alert
                type="warning"
                showIcon
                message="Последствия смены региона"
                description={
                  <ul style={{ paddingLeft: 20, margin: 0 }}>
                    <li>Текущий кошелёк ({REGION_CURRENCIES[currentRegion]}) будет архивирован</li>
                    <li>Создан новый кошелёк в {REGION_CURRENCIES[targetRegion]}</li>
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

        <Card
          title={
            <Space>
              <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />
              <Text>Удаление аккаунта</Text>
            </Space>
          }
        >
          <div style={{ maxWidth: 500 }}>
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
  )
}
