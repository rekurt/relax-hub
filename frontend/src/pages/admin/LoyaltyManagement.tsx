import { useState } from 'react'
import {
  Button,
  Card,
  Col,
  Form,
  InputNumber,
  Modal,
  Row,
  Space,
  Spin,
  Table,
  Tag,
  Typography,
} from '@/components/design/system'
import { EyeOutlined, TrophyOutlined } from '@/components/design/icons'
import type { ColumnsType } from '@/components/design/types'
import {
  useGetMyLoyaltyLevels,
} from '@/api/generated/loyalty/loyalty'
import type { InternalHandlerLoyaltyLevelResponse } from '@/api/generated/model'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

const LEVEL_COLORS: Record<string, string> = {
  bronze: '#cd7f32',
  silver: '#c0c0c0',
  gold: '#ffd700',
  platinum: '#e5e4e2',
}

const LEVEL_NAMES: Record<string, string> = {
  bronze: 'Бронза',
  silver: 'Серебро',
  gold: 'Золото',
  platinum: 'Платина',
}

interface TierFormValues {
  min_visits: number
  discount_percent: number
  point_multiplier: number
}

export default function LoyaltyManagement() {
  const [form] = Form.useForm<TierFormValues>()

  const [editModalOpen, setEditModalOpen] = useState(false)
  const [editingLevel, setEditingLevel] = useState<InternalHandlerLoyaltyLevelResponse | null>(null)

  const { data, isLoading } = useGetMyLoyaltyLevels()

  const levels: InternalHandlerLoyaltyLevelResponse[] = data?.data ?? []

  const openViewModal = (level: InternalHandlerLoyaltyLevelResponse) => {
    setEditingLevel(level)
    form.setFieldsValue({
      min_visits: level.min_visits ?? 0,
      discount_percent: level.discount_percent ?? 0,
      point_multiplier: level.point_multiplier ?? 1,
    })
    setEditModalOpen(true)
  }

  const columns: ColumnsType<InternalHandlerLoyaltyLevelResponse> = [
    {
      title: 'Уровень',
      dataIndex: 'level',
      key: 'level',
      width: 160,
      render: (level: string) => (
        <Tag
          color={LEVEL_COLORS[level] ?? 'default'}
          icon={<TrophyOutlined />}
          className="rh-admin-loyalty-tag"
        >
          {LEVEL_NAMES[level] ?? level}
        </Tag>
      ),
    },
    {
      title: 'Мин. визитов',
      dataIndex: 'min_visits',
      key: 'min_visits',
      width: 140,
      render: (val: number) => val ?? '—',
    },
    {
      title: 'Кэшбэк, %',
      dataIndex: 'discount_percent',
      key: 'discount_percent',
      width: 130,
      render: (val: number) => (val !== undefined ? `${val}%` : '—'),
    },
    {
      title: 'Множитель баллов',
      dataIndex: 'point_multiplier',
      key: 'point_multiplier',
      width: 160,
      render: (val: number) => (val !== undefined ? `×${val}` : '—'),
    },
    {
      title: '',
      key: 'actions',
      width: 100,
      render: (_: unknown, record: InternalHandlerLoyaltyLevelResponse) => (
        <Button
          type="link"
          size="small"
          icon={<EyeOutlined />}
          onClick={() => openViewModal(record)}
        >
          Просмотр
        </Button>
      ),
    },
  ]

  const sortedLevels = [...levels].sort((a, b) => (a.min_visits ?? 0) - (b.min_visits ?? 0))

  return (
    <div className="rh-admin-loyalty-page">
      <PageHeader
        eyebrow="Клиентский опыт"
        title="Управление программой лояльности"
        description="Уровни, кэшбэк и множители баллов для повторных визитов."
      />

      <Row gutter={[16, 16]} className="rh-admin-loyalty-grid">
        {sortedLevels.map((level) => (
          <Col xs={12} sm={6} key={level.level}>
            <Card
              size="small"
              className={`rh-admin-loyalty-tier rh-admin-loyalty-tier--${level.level ?? 'default'}`}
            >
              <Text type="secondary">{LEVEL_NAMES[level.level ?? ''] ?? level.level}</Text>
              <div className="rh-admin-loyalty-tier__visits">от {level.min_visits ?? 0} визитов</div>
              <div className="rh-admin-loyalty-tier__meta">
                <Text type="secondary">Кэшбэк: {level.discount_percent ?? 0}%</Text>
              </div>
              <div className="rh-admin-loyalty-tier__meta">
                <Text type="secondary">Множитель: ×{level.point_multiplier ?? 1}</Text>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      <div className="rh-admin-toolbar rh-admin-toolbar--spaced">
        <div className="rh-admin-toolbar__copy">
          <h2 className="rh-admin-toolbar__title">Настройка уровней</h2>
          <span className="rh-admin-toolbar__hint">Текущие правила начислений доступны для просмотра.</span>
        </div>
      </div>

      {isLoading ? (
        <Card className="rh-admin-state-card"><Spin size="large" /></Card>
      ) : levels.length === 0 ? (
        <Card>
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Нет уровней лояльности</div>
            <p className="rh-admin-empty-state__text">
              Уровни появятся здесь после настройки правил программы.
            </p>
          </div>
        </Card>
      ) : (
        <Table
          dataSource={sortedLevels}
          columns={columns}
          rowKey="level"
          pagination={false}
          size="middle"
          locale={{ emptyText: 'Нет уровней лояльности' }}
        />
      )}

      <Modal
        title={`Уровень: ${LEVEL_NAMES[editingLevel?.level ?? ''] ?? editingLevel?.level}`}
        open={editModalOpen}
        onCancel={() => setEditModalOpen(false)}
        footer={<Button onClick={() => setEditModalOpen(false)}>Закрыть</Button>}
      >
        <Form form={form} layout="vertical" className="rh-admin-modal-form">
          <Form.Item name="min_visits" label="Минимальное количество визитов">
            <InputNumber className="rh-admin-form-control" disabled />
          </Form.Item>
          <Form.Item label="Процент кэшбэка">
            <Space.Compact className="rh-compact-control">
              <Form.Item name="discount_percent" noStyle>
                <InputNumber className="rh-admin-form-control" disabled />
              </Form.Item>
              <span className="rh-input-addon">%</span>
            </Space.Compact>
          </Form.Item>
          <Form.Item label="Множитель баллов">
            <Space.Compact className="rh-compact-control">
              <Form.Item name="point_multiplier" noStyle>
                <InputNumber className="rh-admin-form-control" disabled />
              </Form.Item>
              <span className="rh-input-addon">×</span>
            </Space.Compact>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
