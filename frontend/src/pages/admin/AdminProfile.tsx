import { useEffect, useState } from 'react'
import {
  Alert,
  Card,
  Form,
  Input,
  Button,
  Upload,
  Avatar,
  Space,
  Popconfirm,
  App,
  Tag,
  Typography,
  QRCode,
} from '@/components/design/system'
import {
  UserOutlined,
  UploadOutlined,
  DeleteOutlined,
  KeyOutlined,
  MobileOutlined,
} from '@/components/design/icons'
import { useAuthStore } from '@/stores/auth'
import {
  usePutAuthMe,
  usePostAuthMeAvatar,
  useDeleteAuthMeAvatar,
} from '@/api/generated/auth/auth'
import {
  usePostAuth2faTotpEnable,
  usePostAuth2faTotpVerify,
  useDeleteAuth2faTotp,
  usePostAuth2faSmsEnable,
} from '@/api/generated/2fa/2fa'
import PageHeader from '@/components/PageHeader'
import { resolveAssetUrl } from '@/lib/asset-url'

const { Text } = Typography

export default function AdminProfile() {
  const { user, loadProfile, setUser } = useAuthStore()
  const { message } = App.useApp()
  const [profileForm] = Form.useForm()
  const [avatarUploading, setAvatarUploading] = useState(false)

  const totpEnabled = user?.two_fa_method === 'totp'
  const smsEnabled = user?.two_fa_method === 'sms'

  const [qrUrl, setQrUrl] = useState('')
  const [totpSecret, setTotpSecret] = useState('')
  const [verifyCode, setVerifyCode] = useState('')
  const [disableCode, setDisableCode] = useState('')

  // Derived: 'done' when backend says TOTP is on; 'qr' when we have a fresh
  // unscanned QR; otherwise 'idle'. No effect needed to sync.
  const totpStep: 'idle' | 'qr' | 'done' = totpEnabled
    ? 'done'
    : qrUrl
      ? 'qr'
      : 'idle'

  const enableTotp = usePostAuth2faTotpEnable({
    mutation: {
      onSuccess: (res) => {
        const data = res?.data as { qr_url?: string; secret?: string } | undefined
        setQrUrl(data?.qr_url ?? '')
        setTotpSecret(data?.secret ?? '')
      },
      onError: () => message.error('Не удалось сгенерировать QR-код'),
    },
  })

  const verifyTotp = usePostAuth2faTotpVerify({
    mutation: {
      onSuccess: () => {
        message.success('TOTP 2FA успешно активирована')
        setQrUrl('')
        setTotpSecret('')
        setVerifyCode('')
        setUser({ two_fa_method: 'totp' })
      },
      onError: () => message.error('Неверный код подтверждения'),
    },
  })

  const disableTotp = useDeleteAuth2faTotp({
    mutation: {
      onSuccess: () => {
        message.success('TOTP 2FA отключена')
        setDisableCode('')
        setUser({ two_fa_method: 'none' })
      },
      onError: () => message.error('Неверный код'),
    },
  })

  const enableSms2fa = usePostAuth2faSmsEnable({
    mutation: {
      onSuccess: () => {
        message.success('SMS 2FA включена')
        setUser({ two_fa_method: 'sms' })
      },
      onError: () => message.error('Не удалось включить SMS 2FA. Убедитесь, что телефон подтверждён.'),
    },
  })

  useEffect(() => {
    if (user) {
      profileForm.setFieldsValue({
        name: user.name ?? '',
        phone: user.phone ?? '',
      })
    }
  }, [user, profileForm])

  useEffect(() => {
    if (typeof window === 'undefined') return
    if (window.location.hash !== '#two-factor') return
    const target = document.getElementById('two-factor')
    if (target) target.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }, [])

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

  const handleProfileSubmit = (values: { name?: string; phone?: string }) => {
    updateProfile.mutate({ data: values })
  }

  const handleAvatarUpload = (file: File) => {
    setAvatarUploading(true)
    uploadAvatar.mutate({ data: { avatar: file } })
    return false
  }

  return (
    <div className="rh-stack rh-admin-reference-page">
      <PageHeader
        eyebrow="Админка"
        title="Профиль администратора"
        description="Базовые данные администратора и аватар для служебных сценариев."
      />

      <div className="rh-stat-grid">
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Администратор</span>
          <span className="rh-stat-tile__value">{user?.name ?? 'Без имени'}</span>
          <span className="rh-stat-tile__hint">{user?.email ?? 'Email не указан'}</span>
        </div>
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Телефон</span>
          <span className="rh-stat-tile__value">{user?.phone ?? 'Не указан'}</span>
          <span className="rh-stat-tile__hint">Используется для служебных контактов и восстановления доступа.</span>
        </div>
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Статус фото</span>
          <span className="rh-stat-tile__value">{user?.avatar_url ? 'Загружен' : 'Не загружен'}</span>
          <span className="rh-stat-tile__hint">Отображается в служебных сценариях и внутренних списках.</span>
        </div>
      </div>

      <div className="rh-grid rh-grid--content-aside">
        <Card title="Основная информация">
          <Form
            form={profileForm}
            layout="vertical"
            onFinish={handleProfileSubmit}
            className="rh-admin-profile-form"
          >
            <Form.Item label="Email">
              <Input value={user?.email ?? ''} disabled />
            </Form.Item>
            <Form.Item label="Имя" name="name">
              <Input placeholder="Ваше имя" />
            </Form.Item>
            <Form.Item label="Телефон" name="phone">
              <Input placeholder="+7 999 123-45-67" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" loading={updateProfile.isPending}>
                Сохранить
              </Button>
            </Form.Item>
          </Form>
        </Card>

        <Card title="Аватар">
          <Space size={16} align="center">
            <Avatar
              size={88}
              src={resolveAssetUrl(user?.avatar_url)}
              icon={!user?.avatar_url && <UserOutlined />}
            />
            <Space orientation="vertical">
              <Text type="secondary">
                Используется в карточках профиля и служебных действиях администратора.
              </Text>
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
      </div>

      <Card
        id="two-factor"
        title={
          <Space>
            <span>Двухфакторная аутентификация</span>
            <Tag color={totpEnabled ? 'green' : smsEnabled ? 'green' : 'gold'}>
              {totpEnabled ? 'TOTP активна' : smsEnabled ? 'SMS активна' : 'Не настроена'}
            </Tag>
          </Space>
        }
      >
        {!totpEnabled && !smsEnabled && (
          <Alert
            type="warning"
            showIcon
            title="Доступ к админ-эндпоинтам требует 2FA"
            description="Все запросы к /api/v1/admin/* возвращают 403 admin_2fa_required, пока вы не включите второй фактор."
            className="rh-admin-inline-alert"
          />
        )}

        <div className="rh-admin-2fa-stack">
          <section>
            <Space align="center" className="rh-admin-inline-heading">
              <Text strong>TOTP (Google Authenticator, Authy)</Text>
              <Tag color={totpEnabled ? 'green' : 'default'}>
                {totpEnabled ? 'Включена' : 'Отключена'}
              </Tag>
            </Space>

            {totpStep === 'idle' && (
              <Space className="rh-admin-full-width" orientation="vertical">
                <Text type="secondary">
                  Сгенерируем QR-код, который вы отсканируете в приложении-аутентификаторе. Коды обновляются каждые 30 секунд.
                </Text>
                <Button
                  type="primary"
                  icon={<KeyOutlined />}
                  onClick={() => enableTotp.mutate()}
                  loading={enableTotp.isPending}
                >
                  Настроить TOTP
                </Button>
              </Space>
            )}

            {totpStep === 'qr' && (
              <Space align="start" size={24} className="rh-admin-full-width">
                <Space orientation="vertical" className="rh-admin-flex-fill">
                  <Alert
                    type="info"
                    showIcon
                    title="Отсканируйте QR"
                    description="Откройте приложение-аутентификатор и добавьте новый аккаунт по коду справа."
                  />
                  <div>
                    <Text type="secondary">Секретный ключ (если QR недоступен)</Text>
                    <div>
                      <Text copyable code>
                        {totpSecret}
                      </Text>
                    </div>
                  </div>
                  <Form
                    layout="inline"
                    onFinish={() => verifyTotp.mutate({ data: { code: verifyCode } })}
                  >
                    <Form.Item>
                      <Input
                        placeholder="6-значный код"
                        value={verifyCode}
                        onChange={(event) => setVerifyCode(event.target.value)}
                        maxLength={6}
                        className="rh-admin-code-input"
                      />
                    </Form.Item>
                    <Form.Item>
                      <Button
                        type="primary"
                        htmlType="submit"
                        loading={verifyTotp.isPending}
                        disabled={verifyCode.length !== 6}
                      >
                        Подтвердить
                      </Button>
                    </Form.Item>
                  </Form>
                </Space>
                {qrUrl && <QRCode value={qrUrl} size={176} />}
              </Space>
            )}

            {totpStep === 'done' && (
              <Space className="rh-admin-full-width" orientation="vertical">
                <Alert
                  type="success"
                  showIcon
                  title="TOTP активирован"
                  description="При следующем входе понадобится одноразовый код из приложения."
                />
                <Space>
                  <Input
                    placeholder="6-значный код для отключения"
                    value={disableCode}
                    onChange={(event) => setDisableCode(event.target.value)}
                    maxLength={6}
                    className="rh-admin-code-input rh-admin-code-input--wide"
                  />
                  <Popconfirm
                    title="Отключить TOTP 2FA?"
                    description="Вы потеряете доступ к админ-эндпоинтам, пока не включите 2FA снова."
                    onConfirm={() => disableTotp.mutate({ data: { code: disableCode } })}
                    okText="Отключить"
                    cancelText="Отмена"
                  >
                    <Button
                      danger
                      loading={disableTotp.isPending}
                      disabled={disableCode.length !== 6}
                    >
                      Отключить TOTP
                    </Button>
                  </Popconfirm>
                </Space>
              </Space>
            )}
          </section>

          <section>
            <Space align="center" className="rh-admin-inline-heading">
              <Text strong>SMS 2FA (резерв)</Text>
              <Tag color={smsEnabled ? 'green' : user?.phone ? 'gold' : 'default'}>
                {smsEnabled ? 'Включена' : user?.phone ? 'Доступна' : 'Нужен телефон'}
              </Tag>
            </Space>
            {smsEnabled ? (
              <Alert
                type="success"
                showIcon
                title="SMS 2FA включена"
                description="Коды будут приходить на подтверждённый номер телефона."
              />
            ) : user?.phone ? (
              <Space className="rh-admin-full-width" orientation="vertical">
                <Text type="secondary">Коды подтверждения будут отправляться на {user.phone}.</Text>
                <Button
                  icon={<MobileOutlined />}
                  onClick={() => enableSms2fa.mutate()}
                  loading={enableSms2fa.isPending}
                >
                  Включить SMS 2FA
                </Button>
              </Space>
            ) : (
              <Alert
                type="warning"
                showIcon
                title="Сначала добавьте номер телефона"
                description="Подтверждённый телефон нужен, чтобы включить SMS-подтверждение."
              />
            )}
          </section>
        </div>
      </Card>
    </div>
  )
}
