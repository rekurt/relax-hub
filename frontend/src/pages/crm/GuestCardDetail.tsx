import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Col,
  Descriptions,
  Input,
  Row,
  Space,
  Statistic,
  Tag,
  Typography,
} from '@/components/design/system'
import {
  ArrowLeftOutlined,
  SaveOutlined,
} from '@/components/design/icons'
import dayjs from 'dayjs'
import { useParams, useNavigate } from 'react-router-dom'
import {
  useGetMyCrmGuests,
  usePutMyCrmGuestsId,
} from '@/api/generated/crm/crm'
import { formatPrice } from '@/lib/format'
import { useQueryClient } from '@tanstack/react-query'
import EmptyState from '@/components/EmptyState'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

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
      <div className="rh-stack">
        <Button type="link" className="rh-admin-detail-back" icon={<ArrowLeftOutlined />} onClick={() => navigate('/crm/guests')}>
          Назад
        </Button>
        <EmptyState
          description="Карточка гостя не найдена. Возможно, она была удалена или ещё не создана."
          actionText="К списку гостей"
          onAction={() => navigate('/crm/guests')}
        />
      </div>
    )
  }

  return (
    <div className="rh-stack">
      <div className="rh-admin-detail-back">
        <Button type="link" icon={<ArrowLeftOutlined />} onClick={() => navigate('/crm/guests')}>
          Назад
        </Button>
      </div>
      <PageHeader
        eyebrow="CRM"
        title="Карточка гостя"
        description="История визитов, суммы, заметки и теги гостя."
        size="compact"
      />

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

      <Card className="rh-admin-detail-card rh-section-offset">
        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label="ID клиента">
            <Text copyable className="rh-code-text">{guest.client_id}</Text>
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

      <Card title="Заметки и теги" className="rh-admin-detail-card rh-section-offset">
        <Space orientation="vertical" className="rh-full-width" size="middle">
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
