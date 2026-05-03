import { useState } from 'react'
import {
  Button,
  Card,
  Col,
  Empty,
  Form,
  InputNumber,
  Modal,
  Row,
  Space,
  Spin,
  Table,
  Tag,
  Typography,
} from 'antd'
import { EyeOutlined, TrophyOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  useGetMyLoyaltyLevels,
} from '@/api/generated/loyalty/loyalty'
import type { InternalHandlerLoyaltyLevelResponse } from '@/api/generated/model'

const { Title, Text } = Typography

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
          style={{ fontSize: 14 }}
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
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>
        Управление программой лояльности
      </Title>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        {sortedLevels.map((level) => (
          <Col xs={12} sm={6} key={level.level}>
            <Card
              size="small"
              style={{ borderTop: `3px solid ${LEVEL_COLORS[level.level ?? ''] ?? '#c9c1b5'}` }}
            >
              <Text type="secondary">{LEVEL_NAMES[level.level ?? ''] ?? level.level}</Text>
              <div style={{ fontSize: 20, fontWeight: 600 }}>от {level.min_visits ?? 0} визитов</div>
              <div style={{ marginTop: 4 }}>
                <Text type="secondary">Кэшбэк: {level.discount_percent ?? 0}%</Text>
              </div>
              <div>
                <Text type="secondary">Множитель: ×{level.point_multiplier ?? 1}</Text>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      <Title level={4} style={{ marginBottom: 12 }}>
        Настройка уровней
      </Title>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: 48 }}><Spin size="large" /></div>
      ) : levels.length === 0 ? (
        <Empty description="Нет уровней лояльности" />
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
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="min_visits" label="Минимальное количество визитов">
            <InputNumber style={{ width: '100%' }} disabled />
          </Form.Item>
          <Form.Item label="Процент кэшбэка">
            <Space.Compact className="bani-compact-control">
              <Form.Item name="discount_percent" noStyle>
                <InputNumber style={{ width: '100%' }} disabled />
              </Form.Item>
              <span className="bani-input-addon">%</span>
            </Space.Compact>
          </Form.Item>
          <Form.Item label="Множитель баллов">
            <Space.Compact className="bani-compact-control">
              <Form.Item name="point_multiplier" noStyle>
                <InputNumber style={{ width: '100%' }} disabled />
              </Form.Item>
              <span className="bani-input-addon">×</span>
            </Space.Compact>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
