import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Col,
  Descriptions,
  Empty,
  Input,
  Row,
  Space,
  Statistic,
  Tag,
  Typography,
} from 'antd'
import {
  ArrowLeftOutlined,
  SaveOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import { useParams, useNavigate } from 'react-router-dom'
import {
  useGetMyCrmGuests,
  usePutMyCrmGuestsId,
} from '@/api/generated/crm/crm'
import { formatPrice } from '@/lib/format'
import { useQueryClient } from '@tanstack/react-query'

const { Title, Text } = Typography

export default function GuestCardDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { message } = App.useApp()

  const { data } = useGetMyCrmGuests({ page: 1, page_size: 50 })
  const guest = (data?.data ?? []).find((g) => g.id === id)

  const [notes, setNotes] = useState(guest?.notes ?? '')
  const [tagsInput, setTagsInput] = useState(guest?.tags?.join(', ') ?? '')
  const [initialized, setInitialized] = useState(false)

  if (guest && !initialized) {
    setNotes(guest.notes ?? '')
    setTagsInput(guest.tags?.join(', ') ?? '')
    setInitialized(true)
  }

  const updateMutation = usePutMyCrmGuestsId({
    mutation: {
      onSuccess: () => {
        message.success('Карточка обновлена')
        queryClient.invalidateQueries({ queryKey: ['/my/crm/guests'] })
      },
      onError: () => message.error('Не удалось обновить'),
    },
  })

  const handleSave = () => {
    if (!id) return
    const tags = tagsInput
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean)
    updateMutation.mutate({ id, data: { notes, tags } })
  }

  if (!guest) {
    return (
      <div>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/crm/guests')}>
          Назад
        </Button>
        <Empty
          description="Карточка гостя не найдена. Возможно, она была удалена или ещё не создана."
          style={{ padding: '48px 0' }}
        >
          <Button type="primary" onClick={() => navigate('/crm/guests')}>
            К списку гостей
          </Button>
        </Empty>
      </div>
    )
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/crm/guests')}>
          Назад
        </Button>
        <Title level={3} style={{ margin: 0 }}>Карточка гостя</Title>
      </Space>

      <Row gutter={[16, 16]}>
        <Col xs={24} md={8}>
          <Card size="small">
            <Statistic title="Визиты" value={guest.visit_count ?? 0} />
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card size="small">
            <Statistic title="Общая сумма" value={formatPrice(guest.total_spent ?? 0)} />
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card size="small">
            <Statistic title="Средний чек" value={formatPrice(guest.avg_check ?? 0)} />
          </Card>
        </Col>
      </Row>

      <Card style={{ marginTop: 16 }}>
        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label="ID клиента">
            <Text copyable style={{ fontFamily: 'monospace' }}>{guest.client_id}</Text>
          </Descriptions.Item>
          <Descriptions.Item label="Первый визит">
            {guest.first_visit_at ? dayjs(guest.first_visit_at).format('DD.MM.YYYY HH:mm') : '—'}
          </Descriptions.Item>
          <Descriptions.Item label="Последний визит">
            {guest.last_visit_at ? dayjs(guest.last_visit_at).format('DD.MM.YYYY HH:mm') : '—'}
          </Descriptions.Item>
          <Descriptions.Item label="Теги">
            {guest.tags?.length ? guest.tags.map((t) => <Tag key={t}>{t}</Tag>) : '—'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="Заметки и теги" style={{ marginTop: 16 }}>
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <div>
            <Text strong>Заметки</Text>
            <Input.TextArea
              rows={4}
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Добавьте заметки о госте..."
            />
          </div>
          <div>
            <Text strong>Теги (через запятую)</Text>
            <Input
              value={tagsInput}
              onChange={(e) => setTagsInput(e.target.value)}
              placeholder="vip, постоянный, корпоратив"
            />
          </div>
          <Button
            type="primary"
            icon={<SaveOutlined />}
            onClick={handleSave}
            loading={updateMutation.isPending}
          >
            Сохранить
          </Button>
        </Space>
      </Card>
    </div>
  )
}
