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
} from 'antd'
import {
  UserOutlined,
  UploadOutlined,
  DeleteOutlined,
} from '@ant-design/icons'
import { useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '@/stores/auth'
import {
  usePutAuthMe,
  usePostAuthMeAvatar,
  useDeleteAuthMeAvatar,
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
import { PROVIDER_LABELS, PROVIDER_COLORS } from '@/lib/constants'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

const NOTIFICATION_CARDS = [
  {
    name: 'in_app',
    title: 'В приложении',
    description: 'События внутри кабинета и в рабочей ленте.',
  },
  {
    name: 'email',
    title: 'Email',
    description: 'Подтверждения, итоги и важные изменения по объектам.',
  },
  {
    name: 'push',
    title: 'Push-уведомления',
    description: 'Быстрые сигналы по бронированиям и действиям гостей.',
  },
  {
    name: 'booking_events',
    title: 'Бронирования',
    description: 'Системные уведомления по созданию и изменению брони.',
  },
  {
    name: 'review_events',
    title: 'Отзывы',
    description: 'Новые отзывы и ответы по объектам.',
  },
  {
    name: 'promo_events',
    title: 'Промокоды',
    description: 'Маркетинговые кампании и промо-сценарии.',
  },
  {
    name: 'reminders',
    title: 'Напоминания',
    description: 'Напоминания по ближайшим действиям и срокам.',
  },
] as const

export default function ProfileSettings() {
  const { user, loadProfile } = useAuthStore()
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [profileForm] = Form.useForm()
  const [prefsForm] = Form.useForm()
  const [avatarUploading, setAvatarUploading] = useState(false)

  const { data: citiesData } = useGetCities()
  const cities = citiesData?.data ?? []

  const { data: prefsData, isLoading: prefsLoading } = useGetMyNotificationPreferences()
  const prefs = prefsData?.data

  const { data: socialData, isLoading: socialLoading } = useGetAuthMeSocialAccounts()
  const socialAccounts = socialData?.data ?? []

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
  const currentCity = cities.find((city) => city.id === user?.city_id)?.name

  return (
    <div className="bani-stack">
      <PageHeader
        eyebrow="Профиль"
        title="Настройки профиля"
        description="Контактные данные, уведомления и привязанные аккаунты в одном месте."
      />

      <div className="bani-stat-grid">
        <div className="bani-stat-tile">
          <span className="bani-stat-tile__eyebrow">Профиль</span>
          <span className="bani-stat-tile__value">{user?.name ?? 'Без имени'}</span>
          <span className="bani-stat-tile__hint">{user?.email ?? 'Email не указан'}</span>
        </div>
        <div className="bani-stat-tile">
          <span className="bani-stat-tile__eyebrow">Город</span>
          <span className="bani-stat-tile__value">{currentCity ?? 'Не выбран'}</span>
          <span className="bani-stat-tile__hint">Используется в локальных сценариях и подстановках.</span>
        </div>
        <div className="bani-stat-tile">
          <span className="bani-stat-tile__eyebrow">Социальные связи</span>
          <span className="bani-stat-tile__value">{socialAccounts.length}</span>
          <span className="bani-stat-tile__hint">Дополнительные способы входа и восстановления доступа.</span>
        </div>
      </div>

      <div className="bani-grid bani-grid--content-aside">
        <div className="bani-stack">
          <Card title="Основная информация">
            <Form
              form={profileForm}
              layout="vertical"
              onFinish={handleProfileSubmit}
              style={{ maxWidth: 560 }}
            >
              <Form.Item label="Имя" name="name">
                <Input placeholder="Ваше имя" />
              </Form.Item>
              <Form.Item label="Телефон" name="phone">
                <Input placeholder="+7 999 123-45-67" />
              </Form.Item>
              <Form.Item label="О себе" name="bio">
                <Input.TextArea rows={4} placeholder="Расскажите о себе" />
              </Form.Item>
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
              >
                <div className="bani-toggle-grid" style={{ marginBottom: 20 }}>
                  {NOTIFICATION_CARDS.map((item) => (
                    <div key={item.name} className="bani-toggle-card">
                      <div className="bani-toggle-card__copy">
                        <Text className="bani-toggle-card__title">{item.title}</Text>
                        <Text className="bani-toggle-card__description">{item.description}</Text>
                      </div>
                      <Form.Item name={item.name} valuePropName="checked" style={{ marginBottom: 0 }}>
                        <Switch />
                      </Form.Item>
                    </div>
                  ))}
                </div>

                <Form.Item>
                  <Button type="primary" htmlType="submit" loading={updatePrefs.isPending}>
                    Сохранить настройки
                  </Button>
                </Form.Item>
              </Form>
            )}
          </Card>
        </div>

        <div className="bani-stack">
          <Card title="Аватар">
            <Space size={16} align="center">
              <Avatar
                size={88}
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
                        title={PROVIDER_LABELS[account.provider ?? ''] ?? account.provider}
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
        </div>
      </div>
    </div>
  )
}
