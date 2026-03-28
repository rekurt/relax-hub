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
} from 'antd'
import {
  UserOutlined,
  UploadOutlined,
  DeleteOutlined,
  CalendarOutlined,
  StarOutlined,
  WalletOutlined,
  TrophyOutlined,
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
import { useGetMyStats } from '@/api/generated/users/users'
import { formatPrice } from '@/lib/format'
import { PROVIDER_LABELS, PROVIDER_COLORS } from '@/lib/constants'
import ProfileCompleteness from '@/components/ProfileCompleteness'

const { Title, Text } = Typography

export default function ClientProfile() {
  const { user, loadProfile } = useAuthStore()
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [profileForm] = Form.useForm()
  const [prefsForm] = Form.useForm()
  const [avatarUploading, setAvatarUploading] = useState(false)

  const { data: citiesData } = useGetCities()
  const cities = citiesData?.data ?? []

  const { data: statsData, isLoading: statsLoading } = useGetMyStats()
  const stats = statsData?.data

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

  const linkedProviders = new Set(socialAccounts.map((a) => a.provider))

  return (
    <div>
      <Title level={4} style={{ marginBottom: 24 }}>Мой профиль</Title>

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
      </div>
    </div>
  )
}
