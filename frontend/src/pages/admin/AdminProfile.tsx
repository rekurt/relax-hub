import { useCallback, useEffect, useState } from 'react'
import { useLocation } from 'react-router-dom'
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
  Typography,
  QRCode,
} from '@/components/design/system'
import {
  UserOutlined,
  UploadOutlined,
  DeleteOutlined,
  KeyOutlined,
  MobileOutlined,
  QrcodeOutlined,
  SafetyOutlined,
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
import { ADMIN_2FA_START_EVENT } from '@/lib/admin2faNotice'

const { Text } = Typography

export default function AdminProfile() {
  const location = useLocation()
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

  const focusTwoFactorSetup = useCallback(() => {
    if (typeof window === 'undefined') return
    const target = document.getElementById('two-factor')
    if (!target) return
    target.scrollIntoView({ behavior: 'smooth', block: 'start' })
    window.setTimeout(() => {
      const action = target.querySelector<HTMLElement>('[data-admin-2fa-primary-action]')
      action?.focus({ preventScroll: true })
    }, 220)
  }, [])

  useEffect(() => {
    const params = new URLSearchParams(location.search)
    if (location.hash !== '#two-factor' && params.get('focus2fa') !== '1') return
    focusTwoFactorSetup()
  }, [focusTwoFactorSetup, location.hash, location.search])

  useEffect(() => {
    if (typeof window === 'undefined') return
    const startTwoFactorSetup = () => {
      focusTwoFactorSetup()
      if (totpStep !== 'idle' || smsEnabled || enableTotp.isPending) return
      enableTotp.mutate()
    }
    window.addEventListener(ADMIN_2FA_START_EVENT, startTwoFactorSetup)
    return () => window.removeEventListener(ADMIN_2FA_START_EVENT, startTwoFactorSetup)
  }, [enableTotp, focusTwoFactorSetup, smsEnabled, totpStep])

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
        className="rh-admin-2fa-card"
        title={
          <div className="rh-admin-2fa-card__title">
            <span>Двухфакторная аутентификация</span>
            <span
              className={[
                'rh-admin-2fa-pill',
                totpEnabled || smsEnabled ? 'rh-admin-2fa-pill--success' : 'rh-admin-2fa-pill--warning',
              ].join(' ')}
            >
              {totpEnabled ? 'TOTP активна' : smsEnabled ? 'SMS активна' : 'Не настроена'}
            </span>
          </div>
        }
      >
        <div className="rh-admin-2fa-summary">
          <div className="rh-admin-2fa-summary__icon" aria-hidden="true">
            <SafetyOutlined />
          </div>
          <div className="rh-admin-2fa-summary__copy">
            <span className="rh-admin-2fa-summary__label">Админ-доступ</span>
            <strong>
              {totpEnabled || smsEnabled
                ? 'Второй фактор включён, админ-эндпоинты доступны.'
                : 'Включите второй фактор, чтобы открыть админ-эндпоинты.'}
            </strong>
            <span>
              Защищает все запросы к <code>/api/v1/admin/*</code> и критичные операции панели.
            </span>
          </div>
        </div>

        <div className="rh-admin-2fa-methods">
          <section className="rh-admin-2fa-method rh-admin-2fa-method--primary">
            <div className="rh-admin-2fa-method__head">
              <span className="rh-admin-2fa-method__icon" aria-hidden="true">
                <QrcodeOutlined />
              </span>
              <div>
                <h3>TOTP</h3>
                <p>Google Authenticator, Authy или любой совместимый генератор кодов.</p>
              </div>
              <span className={totpEnabled ? 'rh-admin-2fa-pill rh-admin-2fa-pill--success' : 'rh-admin-2fa-pill'}>
                {totpEnabled ? 'Включена' : 'Отключена'}
              </span>
            </div>

            {totpStep === 'idle' && (
              <div className="rh-admin-2fa-method__body">
                <p className="rh-admin-2fa-method__text">
                  Сначала сгенерируем QR-код. После сканирования подтвердите настройку 6-значным кодом.
                </p>
                <Button
                  type="primary"
                  icon={<KeyOutlined />}
                  onClick={() => enableTotp.mutate()}
                  loading={enableTotp.isPending}
                  className="rh-admin-2fa-method__button"
                  data-admin-2fa-primary-action
                >
                  Настроить TOTP
                </Button>
              </div>
            )}

            {totpStep === 'qr' && (
              <div className="rh-admin-2fa-qr-layout">
                <div className="rh-admin-2fa-qr-layout__copy">
                  <Alert
                    type="info"
                    showIcon
                    title="Отсканируйте QR"
                    description="Добавьте RelaxHUB в приложение-аутентификатор и введите свежий код ниже."
                  />
                  <div className="rh-admin-2fa-secret">
                    <span>Секретный ключ</span>
                    <Text copyable code>
                      {totpSecret}
                    </Text>
                  </div>
                  <Form
                    layout="inline"
                    className="rh-admin-2fa-confirm-form"
                    onFinish={() => verifyTotp.mutate({ data: { code: verifyCode } })}
                  >
                    <Form.Item>
                      <Input
                        placeholder="6-значный код"
                        value={verifyCode}
                        onChange={(event) => setVerifyCode(event.target.value)}
                        maxLength={6}
                        className="rh-admin-code-input"
                        data-admin-2fa-primary-action
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
                </div>
                {qrUrl && (
                  <div className="rh-admin-2fa-qr">
                    <QRCode value={qrUrl} size={176} />
                    <span>Сканировать в приложении</span>
                  </div>
                )}
              </div>
            )}

            {totpStep === 'done' && (
              <div className="rh-admin-2fa-method__body">
                <Alert
                  type="success"
                  showIcon
                  title="TOTP активирован"
                  description="При следующем входе понадобится одноразовый код из приложения."
                />
                <div className="rh-admin-2fa-disable-row">
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
                </div>
              </div>
            )}
          </section>

          <section className="rh-admin-2fa-method">
            <div className="rh-admin-2fa-method__head">
              <span className="rh-admin-2fa-method__icon rh-admin-2fa-method__icon--muted" aria-hidden="true">
                <MobileOutlined />
              </span>
              <div>
                <h3>SMS 2FA</h3>
                <p>Резервный способ для администраторов с подтверждённым телефоном.</p>
              </div>
              <span
                className={[
                  'rh-admin-2fa-pill',
                  smsEnabled ? 'rh-admin-2fa-pill--success' : user?.phone ? 'rh-admin-2fa-pill--warning' : '',
                ].join(' ')}
              >
                {smsEnabled ? 'Включена' : user?.phone ? 'Доступна' : 'Нужен телефон'}
              </span>
            </div>

            {smsEnabled ? (
              <Alert
                type="success"
                showIcon
                title="SMS 2FA включена"
                description="Коды будут приходить на подтверждённый номер телефона."
              />
            ) : user?.phone ? (
              <div className="rh-admin-2fa-method__body">
                <p className="rh-admin-2fa-method__text">Коды подтверждения будут отправляться на {user.phone}.</p>
                <Button
                  icon={<MobileOutlined />}
                  onClick={() => enableSms2fa.mutate()}
                  loading={enableSms2fa.isPending}
                  className="rh-admin-2fa-method__button"
                >
                  Включить SMS 2FA
                </Button>
              </div>
            ) : (
              <div className="rh-admin-2fa-phone-required">
                <strong>Сначала добавьте номер телефона</strong>
                <span>Подтверждённый телефон нужен, чтобы включить SMS-подтверждение.</span>
              </div>
            )}
          </section>
        </div>
      </Card>
    </div>
  )
}
