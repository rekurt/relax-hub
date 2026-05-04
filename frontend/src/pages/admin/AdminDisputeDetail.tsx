import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  App,
  Button,
  Card,
  Descriptions,
  Divider,
  Form,
  Image,
  Input,
  InputNumber,
  List,
  Modal,
  Select,
  Space,
  Spin,
  Tag,
  Typography,
} from '@/components/design/system'
import {
  ArrowLeftOutlined,
  CheckOutlined,
  CloseOutlined,
  UserSwitchOutlined,
} from '@/components/design/icons'
import dayjs from 'dayjs'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetAdminDisputesId,
  getGetAdminDisputesIdQueryKey,
  getGetAdminDisputesQueryKey,
  usePatchAdminDisputesIdAssign,
  usePatchAdminDisputesIdResolve,
  usePatchAdminDisputesIdClose,
} from '@/api/generated/disputes-admin/disputes-admin'
import {
  useGetMyDisputesIdEvidence,
  getGetMyDisputesIdEvidenceQueryKey,
} from '@/api/generated/disputes/disputes'
import type { InternalHandlerDisputeEvidenceResponse } from '@/api/generated/model'
import { formatDateTime, formatPrice } from '@/lib/format'
import { resolveAssetUrl } from '@/lib/asset-url'

const { Text, Paragraph } = Typography

const statusLabel: Record<string, string> = {
  open: 'Открыт',
  evidence_collection: 'Сбор доказательств',
  under_review: 'На рассмотрении',
  resolved: 'Решён',
  appealed: 'Апелляция',
  closed: 'Закрыт',
}

const statusColor: Record<string, string> = {
  open: 'blue',
  evidence_collection: 'processing',
  under_review: 'orange',
  resolved: 'green',
  appealed: 'volcano',
  closed: 'default',
}

const reasonLabel: Record<string, string> = {
  service_not_provided: 'Услуга не оказана',
  poor_quality: 'Низкое качество',
  damage: 'Повреждение имущества',
  safety_issue: 'Проблема безопасности',
  billing_error: 'Ошибка в счёте',
  other: 'Другое',
}

const resolutionLabel: Record<string, string> = {
  full_refund: 'Полный возврат',
  partial_refund: 'Частичный возврат',
  no_refund: 'Без возврата',
}

const evidenceTypeLabel: Record<string, string> = {
  photo: 'Фото',
  screenshot: 'Скриншот',
  gps: 'GPS-данные',
  message: 'Сообщение',
  receipt: 'Чек',
}

const RESOLUTION_OPTIONS = [
  { label: 'Полный возврат', value: 'full_refund' },
  { label: 'Частичный возврат', value: 'partial_refund' },
  { label: 'Без возврата', value: 'no_refund' },
]

export default function AdminDisputeDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { message, modal } = App.useApp()
  const queryClient = useQueryClient()

  const [assignModalOpen, setAssignModalOpen] = useState(false)
  const [assignInput, setAssignInput] = useState('')
  const [resolveModalOpen, setResolveModalOpen] = useState(false)
  const [resolveForm] = Form.useForm()

  const { data: disputeData, isLoading: disputeLoading } = useGetAdminDisputesId(id!)
  const { data: evidenceData, isLoading: evidenceLoading } = useGetMyDisputesIdEvidence(id!)

  const assignMutation = usePatchAdminDisputesIdAssign()
  const resolveMutation = usePatchAdminDisputesIdResolve()
  const closeMutation = usePatchAdminDisputesIdClose()

  const dispute = disputeData?.data
  const evidence: InternalHandlerDisputeEvidenceResponse[] = evidenceData?.data ?? []

  const invalidateAll = () => {
    queryClient.invalidateQueries({ queryKey: getGetAdminDisputesIdQueryKey(id) })
    queryClient.invalidateQueries({ queryKey: getGetMyDisputesIdEvidenceQueryKey(id) })
    queryClient.invalidateQueries({ queryKey: getGetAdminDisputesQueryKey() })
  }

  const handleAssign = async () => {
    if (!assignInput.trim()) {
      message.warning('Укажите ID медиатора')
      return
    }
    try {
      await assignMutation.mutateAsync({
        id: id!,
        data: { mediator_id: assignInput.trim() },
      })
      message.success('Медиатор назначен')
      setAssignModalOpen(false)
      setAssignInput('')
      invalidateAll()
    } catch {
      message.error('Не удалось назначить медиатора')
    }
  }

  const handleResolve = async () => {
    try {
      const values = await resolveForm.validateFields()
      await resolveMutation.mutateAsync({
        id: id!,
        data: {
          resolution: values.resolution,
          refund_amount: Math.round((values.refund_amount ?? 0) * 100),
          compensation_amount: Math.round((values.compensation_amount ?? 0) * 100),
          mediator_notes: values.mediator_notes || undefined,
        },
      })
      message.success('Спор решён')
      setResolveModalOpen(false)
      resolveForm.resetFields()
      invalidateAll()
    } catch {
      if (resolveMutation.isError) {
        message.error('Не удалось решить спор')
      }
    }
  }

  const handleClose = () => {
    modal.confirm({
      title: 'Закрыть спор?',
      content: 'Спор будет закрыт окончательно.',
      okText: 'Закрыть',
      cancelText: 'Отмена',
      onOk: () =>
        closeMutation
          .mutateAsync({ id: id! })
          .then(() => {
            message.success('Спор закрыт')
            invalidateAll()
          })
          .catch(() => {
            message.error('Не удалось закрыть спор')
          }),
    })
  }

  if (disputeLoading || evidenceLoading) {
    return (
      <Card className="rh-admin-state-card">
        <Spin size="large" />
      </Card>
    )
  }

  if (!dispute) {
    return (
      <Card>
        <div className="rh-admin-empty-state">
          <div className="rh-admin-empty-state__title">Спор не найден</div>
          <p className="rh-admin-empty-state__text">
            Проверьте идентификатор спора или вернитесь к списку обращений.
          </p>
        </div>
      </Card>
    )
  }

  const canAssign = dispute.status !== 'closed' && dispute.status !== 'resolved'
  const canResolve =
    dispute.status === 'under_review' ||
    dispute.status === 'evidence_collection' ||
    dispute.status === 'open' ||
    dispute.status === 'appealed'
  const canClose = dispute.status === 'resolved' || dispute.status === 'appealed'

  const isImageUrl = (url: string) => /\.(jpg|jpeg|png|gif|webp)(\?|$)/i.test(url)

  const evidenceDeadlinePassed =
    dispute.evidence_deadline && dayjs(dispute.evidence_deadline).isBefore(dayjs())

  return (
    <div className="rh-admin-detail-page">
      <div className="rh-admin-detail-back">
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/admin/disputes')}
        >
          Назад
        </Button>
      </div>

      <Card className="rh-admin-detail-hero">
        <div className="rh-admin-toolbar">
          <div className="rh-admin-toolbar__copy">
            <span className="rh-admin-toolbar__hint">Медиация и компенсации</span>
            <h1 className="rh-admin-toolbar__title">
              Спор: {reasonLabel[dispute.reason!] ?? dispute.reason}
            </h1>
          </div>
          <Space className="rh-admin-toolbar__actions" wrap>
            {canAssign && (
              <Button
                icon={<UserSwitchOutlined />}
                onClick={() => setAssignModalOpen(true)}
              >
                Назначить медиатора
              </Button>
            )}
            {canResolve && (
              <Button
                type="primary"
                icon={<CheckOutlined />}
                onClick={() => setResolveModalOpen(true)}
              >
                Решить спор
              </Button>
            )}
            {canClose && (
              <Button
                danger
                icon={<CloseOutlined />}
                onClick={handleClose}
                loading={closeMutation.isPending}
              >
                Закрыть
              </Button>
            )}
          </Space>
        </div>
      </Card>

      <Card className="rh-admin-detail-card">
        <Descriptions column={{ xs: 1, sm: 2, md: 3 }} size="small">
          <Descriptions.Item label="Статус">
            <Tag color={statusColor[dispute.status!] ?? 'default'}>
              {statusLabel[dispute.status!] ?? dispute.status}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Причина">
            {reasonLabel[dispute.reason!] ?? dispute.reason}
          </Descriptions.Item>
          <Descriptions.Item label="Бронирование">
            {dispute.booking_id ? (
              <Text copyable={{ text: dispute.booking_id }}>
                {dispute.booking_id.slice(0, 8)}...
              </Text>
            ) : (
              '-'
            )}
          </Descriptions.Item>
          <Descriptions.Item label="Инициатор">
            {dispute.initiator_id ? (
              <Text copyable={{ text: dispute.initiator_id }}>
                {dispute.initiator_id.slice(0, 8)}...
              </Text>
            ) : (
              '-'
            )}
          </Descriptions.Item>
          <Descriptions.Item label="Ответчик">
            {dispute.respondent_id ? (
              <Text copyable={{ text: dispute.respondent_id }}>
                {dispute.respondent_id.slice(0, 8)}...
              </Text>
            ) : (
              '-'
            )}
          </Descriptions.Item>
          <Descriptions.Item label="Медиатор">
            {dispute.mediator_id ? (
              <Text copyable={{ text: dispute.mediator_id }}>
                {dispute.mediator_id.slice(0, 8)}...
              </Text>
            ) : (
              <Text type="secondary">Не назначен</Text>
            )}
          </Descriptions.Item>
          <Descriptions.Item label="Создан">
            {dispute.created_at ? formatDateTime(dispute.created_at) : '-'}
          </Descriptions.Item>
          {dispute.evidence_deadline && (
            <Descriptions.Item label="Срок доказательств">
              {formatDateTime(dispute.evidence_deadline)}
              {evidenceDeadlinePassed && (
                <Tag color="red" className="rh-admin-inline-tag">Истёк</Tag>
              )}
            </Descriptions.Item>
          )}
          {dispute.resolution && (
            <Descriptions.Item label="Решение">
              <Tag color={dispute.resolution === 'full_refund' ? 'green' : dispute.resolution === 'no_refund' ? 'red' : 'orange'}>
                {resolutionLabel[dispute.resolution] ?? dispute.resolution}
              </Tag>
            </Descriptions.Item>
          )}
          {(dispute.refund_amount ?? 0) > 0 && (
            <Descriptions.Item label="Сумма возврата">
              {formatPrice(dispute.refund_amount!)}
            </Descriptions.Item>
          )}
          {(dispute.compensation_amount ?? 0) > 0 && (
            <Descriptions.Item label="Компенсация">
              {formatPrice(dispute.compensation_amount!)}
            </Descriptions.Item>
          )}
          {dispute.resolved_at && (
            <Descriptions.Item label="Решён">
              {formatDateTime(dispute.resolved_at)}
            </Descriptions.Item>
          )}
          {dispute.appeal_deadline && (
            <Descriptions.Item label="Срок апелляции">
              {formatDateTime(dispute.appeal_deadline)}
            </Descriptions.Item>
          )}
        </Descriptions>

        {dispute.description && (
          <>
            <Divider className="rh-admin-detail-divider" />
            <Text strong>Описание от инициатора:</Text>
            <Paragraph className="rh-admin-detail-note">
              {dispute.description}
            </Paragraph>
          </>
        )}

        {dispute.mediator_notes && (
          <>
            <Divider className="rh-admin-detail-divider" />
            <Text strong>Заметки медиатора:</Text>
            <Paragraph className="rh-admin-detail-note">
              {dispute.mediator_notes}
            </Paragraph>
          </>
        )}
      </Card>

      {/* Evidence section */}
      <Card title="Доказательства" className="rh-admin-detail-card rh-admin-evidence-card">
        {evidence.length === 0 ? (
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Нет доказательств</div>
            <p className="rh-admin-empty-state__text">
              Пользователи пока не приложили материалы по этому спору.
            </p>
          </div>
        ) : (
          <List
            dataSource={evidence}
            renderItem={(item) => (
              <List.Item>
                <List.Item.Meta
                  title={
                    <Space>
                      <Tag>{evidenceTypeLabel[item.type!] ?? item.type}</Tag>
                      <Text type="secondary" className="rh-admin-evidence-meta">
                        от {item.user_id?.slice(0, 8)}...
                      </Text>
                      {item.created_at && (
                        <Text type="secondary" className="rh-admin-evidence-meta">
                          {formatDateTime(item.created_at)}
                        </Text>
                      )}
                    </Space>
                  }
                  description={
                    <div className="rh-admin-evidence-body">
                      {item.description && (
                        <Paragraph className="rh-admin-evidence-description">
                          {item.description}
                        </Paragraph>
                      )}
                      {item.url && isImageUrl(item.url) ? (
                        <Image
                          src={resolveAssetUrl(item.url)}
                          alt="Доказательство"
                          width={200}
                          className="rh-admin-evidence-image"
                        />
                      ) : item.url ? (
                        <a href={resolveAssetUrl(item.url)} target="_blank" rel="noopener noreferrer">
                          Открыть файл
                        </a>
                      ) : null}
                    </div>
                  }
                />
              </List.Item>
            )}
          />
        )}
      </Card>

      {/* Assign modal */}
      <Modal
        title="Назначить медиатора"
        open={assignModalOpen}
        onCancel={() => {
          setAssignModalOpen(false)
          setAssignInput('')
        }}
        onOk={handleAssign}
        okText="Назначить"
        cancelText="Отмена"
        confirmLoading={assignMutation.isPending}
      >
        <Text className="rh-admin-modal-label">ID администратора-медиатора:</Text>
        <Input
          placeholder="UUID администратора"
          value={assignInput}
          onChange={(e) => setAssignInput(e.target.value)}
        />
      </Modal>

      {/* Resolve modal */}
      <Modal
        title="Решить спор"
        open={resolveModalOpen}
        onCancel={() => {
          setResolveModalOpen(false)
          resolveForm.resetFields()
        }}
        onOk={handleResolve}
        okText="Решить"
        cancelText="Отмена"
        confirmLoading={resolveMutation.isPending}
        width={520}
      >
        <Form form={resolveForm} layout="vertical">
          <Form.Item
            name="resolution"
            label="Решение"
            rules={[{ required: true, message: 'Выберите решение' }]}
          >
            <Select
              placeholder="Выберите решение"
              options={RESOLUTION_OPTIONS}
            />
          </Form.Item>
          <Form.Item
            label="Сумма возврата (в рублях)"
          >
            <Space.Compact className="rh-compact-control">
              <Form.Item
                name="refund_amount"
                noStyle
                rules={[{ required: true, message: 'Укажите сумму возврата' }]}
              >
                <InputNumber
                  min={0}
                  precision={2}
                  className="rh-admin-form-control"
                  placeholder="0.00"
                />
              </Form.Item>
              <span className="rh-input-addon">₽</span>
            </Space.Compact>
          </Form.Item>
          <Form.Item
            label="Компенсация на кошелёк (в рублях)"
          >
            <Space.Compact className="rh-compact-control">
              <Form.Item name="compensation_amount" noStyle>
                <InputNumber
                  min={0}
                  precision={2}
                  className="rh-admin-form-control"
                  placeholder="0.00"
                />
              </Form.Item>
              <span className="rh-input-addon">₽</span>
            </Space.Compact>
          </Form.Item>
          <Form.Item name="mediator_notes" label="Заметки медиатора">
            <Input.TextArea
              rows={3}
              placeholder="Обоснование решения"
              maxLength={5000}
              showCount
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
