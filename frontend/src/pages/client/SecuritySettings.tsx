import { useMemo, useState } from 'react'
import {
  Typography,
  Card,
  Button,
  Tag,
  App,
  Input,
  Alert,
  Popconfirm,
  Empty,
  Spin,
  Form,
  QRCode,
} from 'antd'
import {
  LaptopOutlined,
  MobileOutlined,
  DeleteOutlined,
  KeyOutlined,
  MailOutlined,
  ExclamationCircleOutlined,
  LockOutlined,
  SafetyCertificateOutlined,
  ClockCircleOutlined,
  MessageOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/ru'
import {
  useGetMySessions,
  useDeleteMySessionsId,
  useDeleteMySessions,
  getGetMySessionsQueryKey,
} from '@/api/generated/sessions/sessions'
import {
  usePostAuth2faTotpEnable,
  usePostAuth2faTotpVerify,
  useDeleteAuth2faTotp,
  usePostAuth2faSmsEnable,
} from '@/api/generated/2fa/2fa'
import { usePostAuthForgotPassword } from '@/api/generated/auth/auth'
import { useAuthStore } from '@/stores/auth'
import { useQueryClient } from '@tanstack/react-query'

dayjs.extend(relativeTime)
dayjs.locale('ru')

const { Title, Text, Paragraph } = Typography

interface SessionRow {
  id?: string
  device_info?: string
  browser?: string
  ip?: string
  last_active_at?: string
  is_current?: boolean
  created_at?: string
}

interface ProtectionItem {
  key: string
  title: string
  status: string
  description: string
  tone: 'success' | 'warning' | 'neutral'
}

function getSessionIcon(device?: string) {
  return device?.toLowerCase().includes('mobile') ? <MobileOutlined /> : <LaptopOutlined />
}

function getProtectionToneClass(tone: ProtectionItem['tone']) {
  switch (tone) {
    case 'success':
      return 'bani-security-rail-item--success'
    case 'warning':
      return 'bani-security-rail-item--warning'
    default:
      return 'bani-security-rail-item--neutral'
  }
}

export default function SecuritySettings() {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const user = useAuthStore((s) => s.user)

  const { data: sessionsData, isLoading: sessionsLoading } = useGetMySessions()
  const sessions: SessionRow[] = sessionsData?.data ?? []
  const currentSession = sessions.find((session) => session.is_current) ?? sessions[0] ?? null
  const otherSessionsCount = sessions.filter((session) => !session.is_current).length

  const terminateSession = useDeleteMySessionsId({
    mutation: {
      onSuccess: () => {
        message.success('Сессия завершена')
        queryClient.invalidateQueries({ queryKey: getGetMySessionsQueryKey() })
      },
      onError: () => message.error('Не удалось завершить сессию'),
    },
  })

  const terminateAllSessions = useDeleteMySessions({
    mutation: {
      onSuccess: () => {
        message.success('Все другие сессии завершены')
        queryClient.invalidateQueries({ queryKey: getGetMySessionsQueryKey() })
      },
      onError: () => message.error('Не удалось завершить сессии'),
    },
  })

  const [totpStep, setTotpStep] = useState<'idle' | 'qr' | 'verify' | 'done'>(
    (user as Record<string, unknown>)?.totp_enabled ? 'done' : 'idle',
  )
  const [qrUrl, setQrUrl] = useState('')
  const [totpSecret, setTotpSecret] = useState('')
  const [verifyCode, setVerifyCode] = useState('')
  const [disableCode, setDisableCode] = useState('')

  const enableTotp = usePostAuth2faTotpEnable({
    mutation: {
      onSuccess: (res) => {
        const data = res?.data as { qr_url?: string; secret?: string } | undefined
        setQrUrl(data?.qr_url ?? '')
        setTotpSecret(data?.secret ?? '')
        setTotpStep('qr')
      },
      onError: () => message.error('Не удалось сгенерировать QR-код'),
    },
  })

  const verifyTotp = usePostAuth2faTotpVerify({
    mutation: {
      onSuccess: () => {
        message.success('TOTP 2FA успешно активирована')
        setTotpStep('done')
        setVerifyCode('')
      },
      onError: () => message.error('Неверный код подтверждения'),
    },
  })

  const disableTotp = useDeleteAuth2faTotp({
    mutation: {
      onSuccess: () => {
        message.success('TOTP 2FA отключена')
        setTotpStep('idle')
        setDisableCode('')
      },
      onError: () => message.error('Неверный код'),
    },
  })

  const [smsEnabled, setSmsEnabled] = useState(
    !!(user as Record<string, unknown>)?.sms_2fa_enabled,
  )

  const enableSms2fa = usePostAuth2faSmsEnable({
    mutation: {
      onSuccess: () => {
        message.success('SMS 2FA включена')
        setSmsEnabled(true)
      },
      onError: () => message.error('Не удалось включить SMS 2FA. Убедитесь, что телефон подтверждён.'),
    },
  })

  const [resetSent, setResetSent] = useState(false)

  const forgotPassword = usePostAuthForgotPassword({
    mutation: {
      onSuccess: () => {
        message.success('Ссылка для сброса пароля отправлена на email')
        setResetSent(true)
      },
      onError: () => message.error('Не удалось отправить ссылку'),
    },
  })

  const totpEnabled = totpStep === 'done'
  const hasRecoveryEmail = Boolean(user?.email)
  const securityScore = useMemo(() => {
    const rawScore =
      42
      + (totpEnabled ? 32 : 0)
      + (smsEnabled ? 14 : 0)
      + (hasRecoveryEmail ? 12 : 0)
      + (otherSessionsCount === 0 ? 6 : 0)
      - Math.max(0, otherSessionsCount - 1) * 4

    return Math.max(22, Math.min(100, rawScore))
  }, [hasRecoveryEmail, otherSessionsCount, smsEnabled, totpEnabled])

  const securityStatus = securityScore >= 90
    ? 'Контур усилен'
    : securityScore >= 70
      ? 'Хороший уровень'
      : 'Нужна настройка'

  const securitySummary = securityScore >= 90
    ? 'Вход, резервный фактор и восстановление уже собраны в сильный контур.'
    : securityScore >= 70
      ? 'Аккаунт защищён выше базового уровня, но ещё есть запас для усиления.'
      : 'Сейчас аккаунт защищён только частично. Лучше усилить вход и резервное восстановление.'

  const protectionItems: ProtectionItem[] = [
    {
      key: 'totp',
      title: 'TOTP',
      status: totpEnabled ? 'Активен' : 'Не включён',
      description: totpEnabled
        ? 'Основной защитный фактор уже настроен через приложение-аутентификатор.'
        : 'Самый надёжный способ усилить вход без зависимости от SMS.',
      tone: totpEnabled ? 'success' : 'warning',
    },
    {
      key: 'sms',
      title: 'SMS 2FA',
      status: smsEnabled ? 'Резерв включён' : user?.phone ? 'Можно включить' : 'Нужен телефон',
      description: smsEnabled
        ? 'Одноразовые коды могут приходить на подтверждённый номер.'
        : user?.phone
          ? 'Резервный фактор можно включить за один шаг.'
          : 'Сначала добавьте подтверждённый номер телефона в профиль.',
      tone: smsEnabled ? 'success' : user?.phone ? 'warning' : 'neutral',
    },
    {
      key: 'recovery',
      title: 'Восстановление',
      status: hasRecoveryEmail ? 'Готово' : 'Нужен email',
      description: hasRecoveryEmail
        ? 'Ссылка на сброс пароля сможет прийти на основной email аккаунта.'
        : 'Без email восстановление доступа будет слишком хрупким.',
      tone: hasRecoveryEmail ? 'success' : 'warning',
    },
    {
      key: 'sessions',
      title: 'Сессии',
      status: otherSessionsCount > 0 ? `${otherSessionsCount} доп. устройств` : 'Под контролем',
      description: otherSessionsCount > 0
        ? 'Есть другие активные устройства. Если они не ваши, завершите их сразу.'
        : 'Сейчас вход удерживается только на текущем устройстве.',
      tone: otherSessionsCount > 0 ? 'warning' : 'success',
    },
  ]

  const recommendations = [
    !totpEnabled
      ? {
          key: 'rec-totp',
          title: 'Подключить приложение-аутентификатор',
          description: 'Это главный шаг, который сильнее всего поднимает защиту аккаунта.',
          action: 'Настроить TOTP',
          target: 'security-2fa',
        }
      : null,
    !smsEnabled && user?.phone
      ? {
          key: 'rec-sms',
          title: 'Добавить резерв через SMS',
          description: 'Удобно как запасной канал подтверждения, если нет доступа к TOTP.',
          action: 'Включить SMS 2FA',
          target: 'security-2fa',
        }
      : null,
    otherSessionsCount > 0
      ? {
          key: 'rec-sessions',
          title: 'Проверить активные устройства',
          description: 'Завершите лишние сессии, если больше не используете эти устройства.',
          action: 'Открыть сессии',
          target: 'security-sessions',
        }
      : null,
  ].filter(Boolean) as Array<{
    key: string
    title: string
    description: string
    action: string
    target: string
  }>

  const scrollToSection = (sectionId: string) => {
    const section = document.getElementById(sectionId)
    if (!section) return
    if (typeof section.scrollIntoView === 'function') {
      section.scrollIntoView({ behavior: 'smooth', block: 'start' })
    }
  }

  return (
    <div className="bani-stack bani-security-page">
      <section className="bani-security-hero">
        <div className="bani-security-hero__copy">
          <div className="bani-security-hero__eyebrow">Контур безопасности</div>
          <Title level={1} className="bani-security-hero__title">Безопасность</Title>
          <Paragraph className="bani-security-hero__description">
            Закрытый контур для входа, устройств и восстановления доступа. Здесь сразу видно,
            насколько хорошо собрана защита аккаунта и что ещё стоит усилить.
          </Paragraph>

          <div className="bani-security-hero__pill-row">
            <div className="bani-security-hero__pill">
              <span className="bani-security-hero__pill-label">Активных устройств</span>
              <strong>{sessionsLoading ? '...' : sessions.length}</strong>
            </div>
            <div className="bani-security-hero__pill">
              <span className="bani-security-hero__pill-label">Основной вход</span>
              <strong>{totpEnabled ? 'TOTP' : smsEnabled ? 'SMS 2FA' : 'Пароль'}</strong>
            </div>
            <div className="bani-security-hero__pill">
              <span className="bani-security-hero__pill-label">Восстановление</span>
              <strong>{hasRecoveryEmail ? 'Готово' : 'Нужно добавить'}</strong>
            </div>
          </div>

          <div className="bani-security-hero__actions">
            <Button type="primary" onClick={() => scrollToSection('security-2fa')}>
              Усилить вход
            </Button>
            <Button className="bani-security-hero__secondary" onClick={() => scrollToSection('security-sessions')}>
              Проверить сессии
            </Button>
          </div>
        </div>

        <div className="bani-security-hero__panel">
          <span className="bani-security-hero__panel-eyebrow">Индекс защиты</span>
          <div className="bani-security-hero__score-row">
            <div className="bani-security-hero__score">{securityScore}</div>
            <div className="bani-security-hero__score-copy">
              <div className="bani-security-hero__score-label">{securityStatus}</div>
              <Text className="bani-security-hero__score-description">
                {securitySummary}
              </Text>
            </div>
          </div>
          <div className="bani-security-hero__facts">
            <div className="bani-security-hero__fact">
              <span>Текущее устройство</span>
              <strong>{currentSession?.device_info ?? 'Не определено'}</strong>
            </div>
            <div className="bani-security-hero__fact">
              <span>Последняя активность</span>
              <strong>{currentSession?.last_active_at ? dayjs(currentSession.last_active_at).fromNow() : '—'}</strong>
            </div>
            <div className="bani-security-hero__fact">
              <span>Восстановление по email</span>
              <strong>{user?.email ?? 'Не настроено'}</strong>
            </div>
          </div>
        </div>
      </section>

      <div className="bani-security-layout">
        <div className="bani-security-main">
          <Card
            id="security-sessions"
            className="bani-security-card"
            title="Активные сессии"
            extra={
              sessions.length > 1 ? (
                <Popconfirm
                  title="Завершить все другие сессии?"
                  description="Все устройства, кроме текущего, будут отключены."
                  onConfirm={() => terminateAllSessions.mutate()}
                  okText="Завершить все"
                  cancelText="Отмена"
                >
                  <Button danger loading={terminateAllSessions.isPending}>
                    Завершить все другие
                  </Button>
                </Popconfirm>
              ) : null
            }
          >
            <div className="bani-card-toolbar bani-security-card__toolbar">
              <div className="bani-card-toolbar__copy">
                <Text strong>Устройства и входы</Text>
                <Text type="secondary">
                  Текущая сессия отмечена отдельно. Любую лишнюю сессию можно завершить сразу.
                </Text>
              </div>
            </div>

            {sessionsLoading ? (
              <Spin />
            ) : sessions.length === 0 ? (
              <Empty description="Нет активных сессий" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            ) : (
              <div className="bani-security-session-list">
                {sessions.map((session) => (
                  <div
                    key={session.id}
                    className={session.is_current ? 'bani-security-session bani-security-session--current' : 'bani-security-session'}
                  >
                    <div className="bani-security-session__icon">
                      {getSessionIcon(session.device_info)}
                    </div>

                    <div className="bani-security-session__copy">
                      <div className="bani-security-session__head">
                        <Text strong>{session.device_info ?? 'Неизвестное устройство'}</Text>
                        {session.is_current && <Tag color="green">Текущая</Tag>}
                      </div>
                      <div className="bani-security-session__meta">
                        <span>{session.browser ?? 'Браузер не определён'}</span>
                        {session.ip && <span>{session.ip}</span>}
                        <span>{session.last_active_at ? dayjs(session.last_active_at).fromNow() : '—'}</span>
                      </div>
                      {session.created_at && (
                        <Text type="secondary">
                          Вход открыт {dayjs(session.created_at).format('DD MMM YYYY, HH:mm')}
                        </Text>
                      )}
                    </div>

                    <div className="bani-security-session__actions">
                      {session.is_current ? (
                        <div className="bani-security-session__status">
                          Сейчас на этом устройстве
                        </div>
                      ) : (
                        <Popconfirm
                          title="Завершить сессию?"
                          description="Устройство будет отключено от аккаунта."
                          onConfirm={() => terminateSession.mutate({ id: session.id! })}
                          okText="Завершить"
                          cancelText="Отмена"
                        >
                          <Button
                            danger
                            icon={<DeleteOutlined />}
                            loading={terminateSession.isPending}
                          >
                            Завершить
                          </Button>
                        </Popconfirm>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </Card>

          <Card id="security-2fa" className="bani-security-card" title="Двухфакторная аутентификация">
            <div className="bani-security-method-grid">
              <section className={totpEnabled ? 'bani-security-method-card bani-security-method-card--accent' : 'bani-security-method-card'}>
                <div className="bani-security-method-card__header">
                  <div>
                    <div className="bani-security-method-card__eyebrow">Основной фактор</div>
                    <Title level={4} className="bani-security-method-card__title">
                      TOTP
                    </Title>
                  </div>
                  <Tag color={totpEnabled ? 'green' : 'gold'}>
                    {totpEnabled ? 'Активно' : 'Не настроено'}
                  </Tag>
                </div>

                <Paragraph className="bani-security-method-card__description">
                  Приложение-аутентификатор даёт самый стабильный и защищённый второй фактор без зависимости от SMS.
                </Paragraph>

                {totpStep === 'idle' && (
                  <div className="bani-security-method-card__body">
                    <div className="bani-security-checklist">
                      <div className="bani-security-checklist__item">
                        <SafetyCertificateOutlined />
                        <span>Работает даже без мобильной сети.</span>
                      </div>
                      <div className="bani-security-checklist__item">
                        <ClockCircleOutlined />
                        <span>Коды обновляются автоматически каждые 30 секунд.</span>
                      </div>
                    </div>
                    <Button
                      type="primary"
                      icon={<KeyOutlined />}
                      onClick={() => enableTotp.mutate()}
                      loading={enableTotp.isPending}
                    >
                      Настроить TOTP
                    </Button>
                  </div>
                )}

                {totpStep === 'qr' && (
                  <div className="bani-security-qr-layout">
                    <div className="bani-security-qr-copy">
                      <Alert
                        type="info"
                        showIcon
                        message="Отсканируйте QR-код"
                        description="Откройте приложение-аутентификатор и добавьте новый аккаунт по коду справа."
                      />
                      <div className="bani-security-secret">
                        <span className="bani-security-secret__label">Секретный ключ</span>
                        <Text copyable code>
                          {totpSecret}
                        </Text>
                      </div>
                      <Form
                        layout="inline"
                        className="bani-inline-form"
                        onFinish={() => verifyTotp.mutate({ data: { code: verifyCode } })}
                      >
                        <Form.Item>
                          <Input
                            placeholder="Введите 6-значный код"
                            value={verifyCode}
                            onChange={(event) => setVerifyCode(event.target.value)}
                            maxLength={6}
                            style={{ width: 220 }}
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

                    <div className="bani-security-qr-frame">
                      {qrUrl && <QRCode value={qrUrl} size={176} />}
                    </div>
                  </div>
                )}

                {totpStep === 'done' && (
                  <div className="bani-security-method-card__body">
                    <Alert
                      type="success"
                      showIcon
                      message="TOTP уже активирован"
                      description="При следующем входе потребуется одноразовый код из приложения."
                    />
                    <div className="bani-inline-form">
                      <Input
                        placeholder="6-значный код"
                        value={disableCode}
                        onChange={(event) => setDisableCode(event.target.value)}
                        maxLength={6}
                        style={{ width: 220 }}
                      />
                      <Popconfirm
                        title="Отключить TOTP 2FA?"
                        description="Вы потеряете дополнительную защиту аккаунта."
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

              <section className={smsEnabled ? 'bani-security-method-card bani-security-method-card--accent' : 'bani-security-method-card'}>
                <div className="bani-security-method-card__header">
                  <div>
                    <div className="bani-security-method-card__eyebrow">Резервный фактор</div>
                    <Title level={4} className="bani-security-method-card__title">
                      SMS 2FA
                    </Title>
                  </div>
                  <Tag color={smsEnabled ? 'green' : user?.phone ? 'gold' : 'default'}>
                    {smsEnabled ? 'Включено' : user?.phone ? 'Доступно' : 'Нужен телефон'}
                  </Tag>
                </div>

                <Paragraph className="bani-security-method-card__description">
                  Запасной канал для одноразовых кодов. Полезен, если нужно быстро подтвердить вход без приложения.
                </Paragraph>

                {smsEnabled ? (
                  <Alert
                    type="success"
                    showIcon
                    message="SMS 2FA включена"
                    description="Коды подтверждения будут приходить на подтверждённый номер телефона."
                  />
                ) : user?.phone ? (
                  <div className="bani-security-method-card__body">
                    <div className="bani-security-checklist">
                      <div className="bani-security-checklist__item">
                        <MessageOutlined />
                        <span>Коды будут отправляться на номер {user.phone}.</span>
                      </div>
                      <div className="bani-security-checklist__item">
                        <ExclamationCircleOutlined />
                        <span>Используйте как резерв, а не вместо TOTP.</span>
                      </div>
                    </div>
                    <Button
                      icon={<MobileOutlined />}
                      onClick={() => enableSms2fa.mutate()}
                      loading={enableSms2fa.isPending}
                    >
                      Включить SMS 2FA
                    </Button>
                  </div>
                ) : (
                  <Alert
                    type="warning"
                    showIcon
                    message="Сначала добавьте номер телефона"
                    description="Подтверждённый номер нужен, чтобы включить SMS-подтверждение как резервный фактор."
                  />
                )}
              </section>
            </div>
          </Card>

          <Card id="security-password" className="bani-security-card" title="Пароль">
            <div className="bani-security-recovery">
              <div className="bani-security-recovery__copy">
                <Text strong>Восстановление доступа</Text>
                <Paragraph className="bani-security-recovery__description">
                  Для смены пароля отправим ссылку на основной email. Это отдельный безопасный путь восстановления без ручной поддержки.
                </Paragraph>
                <div className="bani-security-recovery__email">
                  <MailOutlined />
                  <span>{user?.email ?? 'Добавьте email в профиль для восстановления доступа.'}</span>
                </div>
              </div>

              <div className="bani-security-recovery__action">
                {resetSent ? (
                  <Alert
                    type="success"
                    showIcon
                    message="Ссылка отправлена"
                    description={`Проверьте почту ${user?.email ?? ''} и перейдите по ссылке для сброса пароля.`}
                  />
                ) : (
                  <Button
                    icon={<LockOutlined />}
                    onClick={() => {
                      if (user?.email) {
                        forgotPassword.mutate({ data: { email: user.email } })
                      } else {
                        message.warning('Email не указан в профиле')
                      }
                    }}
                    loading={forgotPassword.isPending}
                  >
                    Отправить ссылку для сброса пароля
                  </Button>
                )}
              </div>
            </div>
          </Card>
        </div>

        <aside className="bani-security-rail">
          <Card className="bani-security-card" title="Способы защиты">
            <div className="bani-security-rail-list">
              {protectionItems.map((item) => (
                <div
                  key={item.key}
                  className={`bani-security-rail-item ${getProtectionToneClass(item.tone)}`}
                >
                  <div className="bani-security-rail-item__head">
                    <Text strong>{item.title}</Text>
                    <span className="bani-security-rail-item__status">{item.status}</span>
                  </div>
                  <Text type="secondary">{item.description}</Text>
                </div>
              ))}
            </div>
          </Card>

          <Card className="bani-security-card" title="Что усилить сейчас">
            {recommendations.length > 0 ? (
              <div className="bani-security-recommendations">
                {recommendations.map((item) => (
                  <div key={item.key} className="bani-security-recommendation">
                    <div>
                      <Text strong>{item.title}</Text>
                      <Paragraph>{item.description}</Paragraph>
                    </div>
                    <Button type="link" onClick={() => scrollToSection(item.target)}>
                      {item.action}
                    </Button>
                  </div>
                ))}
              </div>
            ) : (
              <Alert
                type="success"
                showIcon
                message="Контур защиты уже собран"
                description="Сейчас не видно срочных слабых мест. Остаётся только периодически проверять список сессий."
              />
            )}
          </Card>
        </aside>
      </div>
    </div>
  )
}
