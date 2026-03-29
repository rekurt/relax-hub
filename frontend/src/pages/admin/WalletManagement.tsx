import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Descriptions,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Space,
  Tag,
  Typography,
} from 'antd'
import {
  WalletOutlined,
  PlusOutlined,
  MinusOutlined,
  LockOutlined,
  UnlockOutlined,
} from '@ant-design/icons'
import { axiosInstance } from '@/api/axios-instance'
import { formatPrice } from '@/lib/format'

const { Title, Text } = Typography

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
        amount: values.amount ? values.amount * 100 : undefined,
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

  return (
    <div style={{ padding: 24 }}>
      <Title level={3}>
        <WalletOutlined /> Управление кошельками
      </Title>

      <Card style={{ marginBottom: 24 }}>
        <Space.Compact style={{ width: '100%', maxWidth: 600 }}>
          <Input
            placeholder="ID кошелька (UUID)"
            value={walletId}
            onChange={(e) => setWalletId(e.target.value)}
            onPressEnter={fetchWallet}
          />
          <Button type="primary" onClick={fetchWallet} loading={loading}>
            Найти
          </Button>
        </Space.Compact>
      </Card>

      {!wallet && !loading && (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="Введите ID кошелька для поиска. Вы сможете зачислить, списать средства или заморозить кошелёк."
          style={{ padding: '48px 0' }}
        />
      )}

      {wallet && (
        <Card
          title="Информация о кошельке"
          extra={
            <Tag color={STATUS_TAGS[wallet.status]?.color ?? 'default'}>
              {STATUS_TAGS[wallet.status]?.text ?? wallet.status}
            </Tag>
          }
        >
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="ID кошелька">{wallet.id}</Descriptions.Item>
            <Descriptions.Item label="ID пользователя">{wallet.user_id}</Descriptions.Item>
            <Descriptions.Item label="Баланс">
              <Text strong>{formatPrice(wallet.balance)}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="Заморожено">{formatPrice(wallet.held_amount)}</Descriptions.Item>
            <Descriptions.Item label="Доступно">
              <Text type="success">{formatPrice(wallet.balance - wallet.held_amount)}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="Валюта">{wallet.currency}</Descriptions.Item>
          </Descriptions>

          <Space style={{ marginTop: 16 }} wrap>
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
              name="amount"
              label="Сумма (в рублях)"
              rules={[
                { required: true, message: 'Введите сумму' },
                { type: 'number', min: 0.01, message: 'Сумма должна быть положительной' },
              ]}
            >
              <InputNumber style={{ width: '100%' }} min={0.01} step={1} precision={2} addonAfter="₽" />
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
