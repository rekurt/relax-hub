import { useState } from 'react'
import {
  Typography,
  Card,
  Table,
  Button,
  Tag,
  Space,
  App,
  Input,
  Alert,
  Popconfirm,
  Divider,
  Descriptions,
  Empty,
  Spin,
  Result,
  Form,
  QRCode,
} from 'antd'
import {
  LaptopOutlined,
  MobileOutlined,
  DeleteOutlined,
  KeyOutlined,
  MailOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  LockOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
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
import PageHeader from '@/components/PageHeader'

dayjs.extend(relativeTime)
dayjs.locale('ru')

const { Text, Paragraph } = Typography

interface SessionRow {
  id?: string
  device_info?: string
  browser?: string
  ip?: string
  last_active_at?: string
  is_current?: boolean
  created_at?: string
}

export default function SecuritySettings() {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const user = useAuthStore((s) => s.user)

  // --- Sessions ---
  const { data: sessionsData, isLoading: sessionsLoading } = useGetMySessions()
  const sessions: SessionRow[] = sessionsData?.data ?? []

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

  // --- 2FA TOTP ---
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

  // --- 2FA SMS ---
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

  // --- Password reset ---
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

  // --- Session table columns ---
  const sessionColumns: ColumnsType<SessionRow> = [
    {
      title: 'Устройство',
      dataIndex: 'device_info',
      key: 'device_info',
      render: (device: string | undefined, record: SessionRow) => (
        <Space>
          {device?.toLowerCase().includes('mobile') ? <MobileOutlined /> : <LaptopOutlined />}
          <div>
            <div>{device ?? 'Неизвестно'}</div>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {record.browser ?? ''}
            </Text>
          </div>
          {record.is_current && <Tag color="green">Текущая</Tag>}
        </Space>
      ),
    },
    {
      title: 'IP-адрес',
      dataIndex: 'ip',
      key: 'ip',
      responsive: ['md'],
    },
    {
      title: 'Последняя активность',
      dataIndex: 'last_active_at',
      key: 'last_active_at',
      render: (val: string | undefined) =>
        val ? dayjs(val).fromNow() : '—',
      responsive: ['sm'],
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_: unknown, record: SessionRow) =>
        record.is_current ? null : (
          <Popconfirm
            title="Завершить сессию?"
            description="Устройство будет отключено от аккаунта."
            onConfirm={() => terminateSession.mutate({ id: record.id! })}
            okText="Завершить"
            cancelText="Отмена"
          >
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
              loading={terminateSession.isPending}
            >
              Завершить
            </Button>
          </Popconfirm>
        ),
    },
  ]

  return (
    <div>
      <PageHeader
        eyebrow="Защита"
        title="Безопасность"
        description="Сессии, 2FA и сброс пароля собраны в отдельный безопасный контур."
      />

      {/* Section 1: Active Sessions */}
      <Card
        title="Активные сессии"
        style={{ marginBottom: 24 }}
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
        {sessionsLoading ? (
          <Spin />
        ) : sessions.length === 0 ? (
          <Empty description="Нет активных сессий" />
        ) : (
          <Table
            dataSource={sessions}
            columns={sessionColumns}
            rowKey="id"
            pagination={false}
            size="small"
          />
        )}
      </Card>

      {/* Section 2: Two-Factor Authentication */}
      <Card title="Двухфакторная аутентификация" style={{ marginBottom: 24 }}>
        <Paragraph type="secondary">
          Дополнительная защита аккаунта. После включения при входе потребуется ввести одноразовый код.
        </Paragraph>

        <Divider titlePlacement="left">TOTP (приложение-аутентификатор)</Divider>

        {totpStep === 'idle' && (
          <Space direction="vertical">
            <Text>
              Используйте Google Authenticator, Authy или аналогичное приложение для генерации одноразовых кодов.
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
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            <Alert
              message="Отсканируйте QR-код"
              description="Откройте приложение-аутентификатор и отсканируйте код ниже."
              type="info"
              showIcon
            />
            {qrUrl && (
              <div style={{ display: 'flex', justifyContent: 'center' }}>
                <QRCode value={qrUrl} size={200} />
              </div>
            )}
            <Descriptions column={1} size="small">
              <Descriptions.Item label="Секретный ключ">
                <Text copyable code>
                  {totpSecret}
                </Text>
              </Descriptions.Item>
            </Descriptions>
            <Form
              layout="inline"
              onFinish={() => verifyTotp.mutate({ data: { code: verifyCode } })}
            >
              <Form.Item>
                <Input
                  placeholder="Введите 6-значный код"
                  value={verifyCode}
                  onChange={(e) => setVerifyCode(e.target.value)}
                  maxLength={6}
                  style={{ width: 200 }}
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
        )}

        {totpStep === 'verify' && null}

        {totpStep === 'done' && (
          <Space direction="vertical" size="middle">
            <Result
              status="success"
              title="TOTP 2FA активирована"
              subTitle="При следующем входе потребуется одноразовый код."
              icon={<CheckCircleOutlined />}
            />
            <Divider />
            <Text>Для отключения введите текущий TOTP-код:</Text>
            <Space>
              <Input
                placeholder="6-значный код"
                value={disableCode}
                onChange={(e) => setDisableCode(e.target.value)}
                maxLength={6}
                style={{ width: 200 }}
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
            </Space>
          </Space>
        )}

        <Divider titlePlacement="left">SMS 2FA</Divider>

        {smsEnabled ? (
          <Alert
            message="SMS 2FA включена"
            description="Коды подтверждения будут отправляться на ваш номер телефона."
            type="success"
            showIcon
            icon={<CheckCircleOutlined />}
          />
        ) : (
          <Space direction="vertical">
            <Text>
              Получайте одноразовые коды по SMS. Требуется подтверждённый номер телефона.
            </Text>
            {user?.phone ? (
              <Button
                icon={<MobileOutlined />}
                onClick={() => enableSms2fa.mutate()}
                loading={enableSms2fa.isPending}
              >
                Включить SMS 2FA
              </Button>
            ) : (
              <Alert
                message="Сначала добавьте номер телефона в профиле"
                type="warning"
                showIcon
                icon={<ExclamationCircleOutlined />}
              />
            )}
          </Space>
        )}
      </Card>

      {/* Section 3: Password Reset */}
      <Card title="Пароль">
        <Paragraph type="secondary">
          Для смены пароля мы отправим ссылку на ваш email. Перейдите по ней, чтобы установить новый пароль.
        </Paragraph>

        {resetSent ? (
          <Alert
            message="Ссылка отправлена"
            description={`Проверьте почту ${user?.email ?? ''} и перейдите по ссылке для сброса пароля.`}
            type="success"
            showIcon
            icon={<MailOutlined />}
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
      </Card>
    </div>
  )
}
