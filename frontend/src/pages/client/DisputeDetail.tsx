import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Alert,
  App,
  Button,
  Card,
  Descriptions,
  Divider,
  Empty,
  Form,
  Image,
  Input,
  List,
  Select,
  Space,
  Spin,
  Tag,
  Typography,
} from '@/components/design/system'
import {
  ArrowLeftOutlined,
  PlusOutlined,
  WarningOutlined,
} from '@/components/design/icons'
import dayjs from 'dayjs'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetMyDisputesId,
  getGetMyDisputesIdQueryKey,
  useGetMyDisputesIdEvidence,
  getGetMyDisputesIdEvidenceQueryKey,
  usePostMyDisputesIdEvidence,
  usePostMyDisputesIdAppeal,
  getGetMyDisputesQueryKey,
} from '@/api/generated/disputes/disputes'
import type { InternalHandlerDisputeEvidenceResponse } from '@/api/generated/model'
import { formatDateTime, formatPrice } from '@/lib/format'

const { Title, Text, Paragraph } = Typography

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

const EVIDENCE_TYPE_OPTIONS = [
  { label: 'Фото', value: 'photo' },
  { label: 'Скриншот', value: 'screenshot' },
  { label: 'GPS-данные', value: 'gps' },
  { label: 'Сообщение', value: 'message' },
  { label: 'Чек', value: 'receipt' },
]

export default function DisputeDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { message, modal } = App.useApp()
  const queryClient = useQueryClient()

  const [evidenceFormVisible, setEvidenceFormVisible] = useState(false)
  const [form] = Form.useForm()

  const { data: disputeData, isLoading: disputeLoading } = useGetMyDisputesId(id!)
  const { data: evidenceData, isLoading: evidenceLoading } = useGetMyDisputesIdEvidence(id!)

  const submitEvidenceMutation = usePostMyDisputesIdEvidence()
  const appealMutation = usePostMyDisputesIdAppeal()

  const dispute = disputeData?.data
  const evidence: InternalHandlerDisputeEvidenceResponse[] = evidenceData?.data ?? []

  const invalidateAll = () => {
    queryClient.invalidateQueries({ queryKey: getGetMyDisputesIdQueryKey(id) })
    queryClient.invalidateQueries({ queryKey: getGetMyDisputesIdEvidenceQueryKey(id) })
    queryClient.invalidateQueries({ queryKey: getGetMyDisputesQueryKey() })
  }

  const handleSubmitEvidence = async () => {
    try {
      const values = await form.validateFields()
      await submitEvidenceMutation.mutateAsync({
        id: id!,
        data: {
          type: values.type,
          url: values.url,
          description: values.description || undefined,
        },
      })
      message.success('Доказательство добавлено')
      form.resetFields()
      setEvidenceFormVisible(false)
      invalidateAll()
    } catch {
      if (submitEvidenceMutation.isError) {
        message.error('Не удалось добавить доказательство')
      }
    }
  }

  const handleAppeal = () => {
    modal.confirm({
      title: 'Подать апелляцию?',
      content: 'Вы можете подать апелляцию на решение по спору. Спор будет повторно рассмотрен.',
      okText: 'Подать апелляцию',
      cancelText: 'Отмена',
      icon: <WarningOutlined />,
      onOk: () =>
        appealMutation
          .mutateAsync({ id: id! })
          .then(() => {
            message.success('Апелляция подана')
            invalidateAll()
          })
          .catch(() => {
            message.error('Не удалось подать апелляцию')
          }),
    })
  }

  if (disputeLoading || evidenceLoading) {
    return (
      <div style={{ textAlign: 'center', padding: 48 }}>
        <Spin size="large" />
      </div>
    )
  }

  if (!dispute) {
    return <Empty description="Спор не найден" />
  }

  const isEvidenceOpen =
    (dispute.status === 'open' || dispute.status === 'evidence_collection') &&
    dispute.evidence_deadline &&
    dayjs(dispute.evidence_deadline).isAfter(dayjs())

  const canAppeal =
    dispute.status === 'resolved' &&
    dispute.appeal_deadline &&
    dayjs(dispute.appeal_deadline).isAfter(dayjs())

  const evidenceTimeLeft = dispute.evidence_deadline
    ? dayjs(dispute.evidence_deadline).diff(dayjs(), 'hour')
    : 0

  const appealTimeLeft = dispute.appeal_deadline
    ? dayjs(dispute.appeal_deadline).diff(dayjs(), 'hour')
    : 0

  const isImageUrl = (url: string) => /\.(jpg|jpeg|png|gif|webp)(\?|$)/i.test(url)

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/client/disputes')}
        >
          Назад
        </Button>
      </Space>

      <Title level={3} style={{ marginBottom: 16 }}>
        Спор: {reasonLabel[dispute.reason!] ?? dispute.reason}
      </Title>

      <Card style={{ marginBottom: 16 }}>
        <Descriptions column={{ xs: 1, sm: 2 }} size="small">
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
              <a onClick={() => navigate(`/client/bookings/${dispute.booking_id}`)}>
                {dispute.booking_id.slice(0, 8)}...
              </a>
            ) : (
              '-'
            )}
          </Descriptions.Item>
          <Descriptions.Item label="Создан">
            {dispute.created_at ? formatDateTime(dispute.created_at) : '-'}
          </Descriptions.Item>
          {dispute.evidence_deadline && (
            <Descriptions.Item label="Срок подачи доказательств">
              {formatDateTime(dispute.evidence_deadline)}
              {isEvidenceOpen && (
                <Text type="warning" style={{ marginLeft: 8 }}>
                  (осталось {evidenceTimeLeft}ч)
                </Text>
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
          {dispute.appeal_deadline && dispute.status === 'resolved' && (
            <Descriptions.Item label="Срок апелляции">
              {formatDateTime(dispute.appeal_deadline)}
              {canAppeal && (
                <Text type="warning" style={{ marginLeft: 8 }}>
                  (осталось {appealTimeLeft}ч)
                </Text>
              )}
            </Descriptions.Item>
          )}
        </Descriptions>

        {dispute.description && (
          <>
            <Divider style={{ margin: '12px 0' }} />
            <Text strong>Описание:</Text>
            <Paragraph style={{ marginTop: 4, whiteSpace: 'pre-wrap' }}>
              {dispute.description}
            </Paragraph>
          </>
        )}

        {dispute.mediator_notes && (
          <>
            <Divider style={{ margin: '12px 0' }} />
            <Text strong>Комментарий медиатора:</Text>
            <Paragraph style={{ marginTop: 4, whiteSpace: 'pre-wrap' }}>
              {dispute.mediator_notes}
            </Paragraph>
          </>
        )}
      </Card>

      {/* Appeal button */}
      {canAppeal && (
        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
          title="Вы можете подать апелляцию"
          description={`Срок подачи апелляции истекает ${formatDateTime(dispute.appeal_deadline!)}. После этого решение станет окончательным.`}
          action={
            <Button
              type="primary"
              danger
              onClick={handleAppeal}
              loading={appealMutation.isPending}
            >
              Подать апелляцию
            </Button>
          }
        />
      )}

      {/* Evidence section */}
      <Card
        title="Доказательства"
        style={{ marginBottom: 16 }}
        extra={
          isEvidenceOpen && (
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setEvidenceFormVisible(true)}
              size="small"
            >
              Добавить
            </Button>
          )
        }
      >
        {isEvidenceOpen && (
          <Alert
            type="warning"
            showIcon
            style={{ marginBottom: 16 }}
            title={`Окно подачи доказательств закрывается через ${evidenceTimeLeft}ч (${formatDateTime(dispute.evidence_deadline!)})`}
          />
        )}

        {evidence.length === 0 ? (
          <Empty description="Нет доказательств" />
        ) : (
          <List
            dataSource={evidence}
            renderItem={(item) => (
              <List.Item>
                <List.Item.Meta
                  title={
                    <Space>
                      <Tag>{evidenceTypeLabel[item.type!] ?? item.type}</Tag>
                      {item.created_at && (
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          {formatDateTime(item.created_at)}
                        </Text>
                      )}
                    </Space>
                  }
                  description={
                    <div>
                      {item.description && (
                        <Paragraph style={{ marginBottom: 8 }}>
                          {item.description}
                        </Paragraph>
                      )}
                      {item.url && isImageUrl(item.url) ? (
                        <Image
                          src={item.url}
                          alt="Доказательство"
                          width={200}
                          style={{ borderRadius: 12 }}
                        />
                      ) : item.url ? (
                        <a href={item.url} target="_blank" rel="noopener noreferrer">
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

        {/* Evidence submission form */}
        {evidenceFormVisible && (
          <>
            <Divider />
            <Form form={form} layout="vertical" onFinish={handleSubmitEvidence}>
              <Form.Item
                name="type"
                label="Тип доказательства"
                rules={[{ required: true, message: 'Выберите тип' }]}
              >
                <Select
                  placeholder="Выберите тип"
                  options={EVIDENCE_TYPE_OPTIONS}
                />
              </Form.Item>
              <Form.Item
                name="url"
                label="Ссылка на файл"
                rules={[
                  { required: true, message: 'Укажите ссылку' },
                  { type: 'url', message: 'Введите корректный URL' },
                ]}
              >
                <Input placeholder="https://..." maxLength={2000} />
              </Form.Item>
              <Form.Item name="description" label="Описание">
                <Input.TextArea
                  rows={3}
                  placeholder="Опишите доказательство"
                  maxLength={2000}
                  showCount
                />
              </Form.Item>
              <Form.Item>
                <Space>
                  <Button
                    type="primary"
                    htmlType="submit"
                    loading={submitEvidenceMutation.isPending}
                  >
                    Отправить
                  </Button>
                  <Button
                    onClick={() => {
                      setEvidenceFormVisible(false)
                      form.resetFields()
                    }}
                  >
                    Отмена
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </>
        )}
      </Card>
    </div>
  )
}
