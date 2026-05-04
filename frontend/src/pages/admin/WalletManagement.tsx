import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Descriptions,
  Form,
  Input,
  InputNumber,
  Modal,
  Space,
  Tag,
} from '@/components/design/system'
import {
  WalletOutlined,
  PlusOutlined,
  MinusOutlined,
  LockOutlined,
  UnlockOutlined,
  SearchOutlined,
} from '@/components/design/icons'
import { axiosInstance } from '@/api/axios-instance'
import { formatPrice } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

interface WalletData {
  id: string
  user_id: string
  balance: number
  held_amount: number
  currency: string
  status: string
}

type WalletAction = 'credit' | 'debit' | 'freeze' | 'unfreeze'

const STATUS_TAGS: Record<string, { color: string; text: string }> = {
  active: { color: 'green', text: 'Активен' },
  frozen: { color: 'red', text: 'Заморожен' },
  archived: { color: 'default', text: 'Архив' },
}

export default function WalletManagement() {
  const { message, modal } = App.useApp()
  const [walletId, setWalletId] = useState('')
  const [wallet, setWallet] = useState<WalletData | null>(null)
  const [loading, setLoading] = useState(false)
  const [actionModal, setActionModal] = useState<WalletAction | null>(null)
  const [form] = Form.useForm()

  const fetchWallet = async () => {
    if (!walletId.trim()) {
      message.warning('Введите ID кошелька')
      return
    }
    setLoading(true)
    try {
      const res = await axiosInstance.get(`/admin/wallets/${walletId}`)
      setWallet(res.data?.data ?? null)
    } catch {
      message.error('Кошелёк не найден')
      setWallet(null)
    } finally {
      setLoading(false)
    }
  }

  const handleAction = async (action: WalletAction, values: { amount?: number; reason: string }) => {
    if (!wallet) return
    try {
      await axiosInstance.post(`/admin/wallets/${wallet.id}/${action}`, {
        amount: values.amount ? Math.round(values.amount * 100) : undefined,
        reason: values.reason,
      })
      message.success(
        action === 'credit'
          ? 'Средства зачислены'
          : action === 'debit'
            ? 'Средства списаны'
            : action === 'freeze'
              ? 'Кошелёк заморожен'
              : 'Кошелёк разморожен',
      )
      setActionModal(null)
      form.resetFields()
      await fetchWallet()
    } catch (err: unknown) {
      const errMsg =
        (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error
          ?.message ?? 'Ошибка выполнения операции'
      message.error(errMsg)
    }
  }

  const confirmAction = (action: WalletAction) => {
    form.validateFields().then((values) => {
      const actionLabels: Record<WalletAction, string> = {
        credit: `Зачислить ${formatPrice((values.amount ?? 0) * 100)}?`,
        debit: `Списать ${formatPrice((values.amount ?? 0) * 100)}?`,
        freeze: 'Заморозить кошелёк?',
        unfreeze: 'Разморозить кошелёк?',
      }
      modal.confirm({
        title: actionLabels[action],
        content: `Причина: ${values.reason}`,
        okText: 'Подтвердить',
        cancelText: 'Отмена',
        onOk: () => handleAction(action, values),
      })
    })
  }

  const needsAmount = actionModal === 'credit' || actionModal === 'debit'

  const actionTitles: Record<WalletAction, string> = {
    credit: 'Зачисление средств',
    debit: 'Списание средств',
    freeze: 'Заморозка кошелька',
    unfreeze: 'Разморозка кошелька',
  }

  const walletStats = wallet
    ? [
        { label: 'Баланс', value: formatPrice(wallet.balance), tone: 'default' },
        { label: 'Заморожено', value: formatPrice(wallet.held_amount), tone: 'warning' },
        { label: 'Доступно', value: formatPrice(wallet.balance - wallet.held_amount), tone: 'success' },
      ]
    : []

  return (
    <div className="rh-admin-wallet-page">
      <PageHeader
        size="compact"
        eyebrow="Финансовый контроль"
        title="Управление кошельками"
        description="Поиск кошелька по UUID, ручные корректировки баланса и блокировка операций с обязательной причиной для аудита."
      />

      <Card className="rh-admin-wallet-search">
        <div className="rh-admin-wallet-search__icon" aria-hidden>
          <WalletOutlined />
        </div>
        <div className="rh-admin-wallet-search__copy">
          <div className="rh-admin-wallet-search__title">Найти кошелёк</div>
          <div className="rh-admin-wallet-search__hint">Введите точный ID кошелька, чтобы открыть баланс и доступные операции.</div>
        </div>
        <Space.Compact className="rh-admin-wallet-search__control">
          <Input
            placeholder="ID кошелька (UUID)"
            value={walletId}
            onChange={(e) => setWalletId(e.target.value)}
            onPressEnter={fetchWallet}
          />
          <Button type="primary" icon={<SearchOutlined />} onClick={fetchWallet} loading={loading}>
            Найти
          </Button>
        </Space.Compact>
      </Card>

      {!wallet && !loading && (
        <Card className="rh-admin-wallet-empty">
          <div className="rh-admin-wallet-empty__icon" aria-hidden>
            <WalletOutlined />
          </div>
          <div className="rh-admin-wallet-empty__title">Кошелёк не выбран</div>
          <p className="rh-admin-wallet-empty__text">
            Введите ID кошелька для поиска. Вы сможете зачислить, списать средства или заморозить кошелёк.
          </p>
        </Card>
      )}

      {wallet && (
        <Card className="rh-admin-wallet-detail">
          <div className="rh-admin-wallet-detail__head">
            <div>
              <div className="rh-admin-wallet-detail__eyebrow">Карточка кошелька</div>
              <h2 className="rh-admin-wallet-detail__title">Информация о кошельке</h2>
            </div>
            <Tag color={STATUS_TAGS[wallet.status]?.color ?? 'default'}>
              {STATUS_TAGS[wallet.status]?.text ?? wallet.status}
            </Tag>
          </div>

          <div className="rh-admin-wallet-stats">
            {walletStats.map((stat) => (
              <div className={`rh-admin-wallet-stat rh-admin-wallet-stat--${stat.tone}`} key={stat.label}>
                <span className="rh-admin-wallet-stat__label">{stat.label}</span>
                <strong className="rh-admin-wallet-stat__value">{stat.value}</strong>
              </div>
            ))}
          </div>

          <Descriptions className="rh-admin-wallet-descriptions" column={2} bordered size="small">
            <Descriptions.Item label="ID кошелька">{wallet.id}</Descriptions.Item>
            <Descriptions.Item label="ID пользователя">{wallet.user_id}</Descriptions.Item>
            <Descriptions.Item label="Валюта">{wallet.currency}</Descriptions.Item>
          </Descriptions>

          <Space className="rh-admin-wallet-actions" wrap>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setActionModal('credit')}
              disabled={wallet.status === 'archived'}
            >
              Зачислить
            </Button>
            <Button
              danger
              icon={<MinusOutlined />}
              onClick={() => setActionModal('debit')}
              disabled={wallet.status !== 'active'}
            >
              Списать
            </Button>
            {wallet.status === 'active' ? (
              <Button icon={<LockOutlined />} onClick={() => setActionModal('freeze')}>
                Заморозить
              </Button>
            ) : wallet.status === 'frozen' ? (
              <Button icon={<UnlockOutlined />} onClick={() => setActionModal('unfreeze')}>
                Разморозить
              </Button>
            ) : null}
          </Space>
        </Card>
      )}

      <Modal
        open={!!actionModal}
        title={actionModal ? actionTitles[actionModal] : ''}
        onCancel={() => {
          setActionModal(null)
          form.resetFields()
        }}
        onOk={() => actionModal && confirmAction(actionModal)}
        okText="Выполнить"
        cancelText="Отмена"
      >
        <Form form={form} layout="vertical">
          {needsAmount && (
            <Form.Item
              label="Сумма (в рублях)"
            >
              <Space.Compact className="rh-compact-control">
                <Form.Item
                  name="amount"
                  noStyle
                  rules={[
                    { required: true, message: 'Введите сумму' },
                    { type: 'number', min: 0.01, message: 'Сумма должна быть положительной' },
                  ]}
                >
                  <InputNumber className="rh-admin-form-control" min={0.01} step={1} precision={2} />
                </Form.Item>
                <span className="rh-input-addon">₽</span>
              </Space.Compact>
            </Form.Item>
          )}
          <Form.Item
            name="reason"
            label="Причина"
            rules={[{ required: true, message: 'Укажите причину операции' }]}
          >
            <Input.TextArea rows={3} placeholder="Укажите причину для аудита..." />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
