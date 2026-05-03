import { useState, type ReactNode } from 'react'
import {
  Typography,
  Card,
  Form,
  Input,
  Button,
  Select,
  Tag,
  Alert,
  Space,
  Skeleton,
  App,
  Descriptions,
  Upload,
} from '@/components/design/system'
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  UploadOutlined,
  SendOutlined,
} from '@/components/design/icons'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetMyKyc,
  usePostMyKyc,
  getGetMyKycQueryKey,
} from '@/api/generated/kyc/kyc'
import type { InternalHandlerKycResponse } from '@/api/generated/model'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

const STATUS_MAP: Record<string, { color: string; label: string; icon: ReactNode }> = {
  pending: { color: 'processing', label: 'На рассмотрении', icon: <ClockCircleOutlined /> },
  approved: { color: 'success', label: 'Одобрено', icon: <CheckCircleOutlined /> },
  rejected: { color: 'error', label: 'Отклонено', icon: <CloseCircleOutlined /> },
}

const ENTITY_TYPE_OPTIONS = [
  { value: 'individual', label: 'Физическое лицо' },
  { value: 'sole_proprietor', label: 'Индивидуальный предприниматель' },
  { value: 'self_employed', label: 'Самозанятый' },
  { value: 'legal_entity', label: 'Юридическое лицо' },
]

const ENTITY_TYPE_LABELS: Record<string, string> = {
  individual: 'Физическое лицо',
  sole_proprietor: 'ИП',
  self_employed: 'Самозанятый',
  legal_entity: 'Юридическое лицо',
}

export default function KYCSettings() {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [form] = Form.useForm()
  const [entityType, setEntityType] = useState<string>('')
  const [documentUrls, setDocumentUrls] = useState<string[]>([])

  const { data: kycData, isLoading, error } = useGetMyKyc()
  const kyc: InternalHandlerKycResponse | undefined = kycData?.data
  const hasKyc = !!kyc && !error

  const submitKyc = usePostMyKyc({
    mutation: {
      onSuccess: () => {
        message.success('Заявка на KYC подана')
        queryClient.invalidateQueries({ queryKey: getGetMyKycQueryKey() })
      },
      onError: () => message.error('Ошибка при подаче заявки'),
    },
  })

  const handleSubmit = (values: {
    entity_type: string
    full_name: string
    inn: string
    ogrnip?: string
    company_name?: string
  }) => {
    submitKyc.mutate({
      data: {
        entity_type: values.entity_type,
        full_name: values.full_name,
        inn: values.inn,
        ogrnip: values.ogrnip,
        company_name: values.company_name,
        document_urls: documentUrls,
      },
    })
  }

  const handleDocumentUpload = (file: File) => {
    const url = URL.createObjectURL(file)
    setDocumentUrls((prev) => [...prev, url])
    message.info('Документ добавлен')
    return false
  }

  if (isLoading) {
    return (
      <div className="rh-stack">
        <PageHeader
          eyebrow="Комплаенс"
          title="Верификация KYC"
          description="Проверка данных владельца и юридической сущности для работы на платформе."
        />
        <Skeleton active />
      </div>
    )
  }

  const statusInfo = hasKyc ? STATUS_MAP[kyc.status ?? ''] : null

  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="Комплаенс"
        title="Верификация KYC"
        description="Проверка данных владельца и юридической сущности для работы на платформе."
      />

      <div className="rh-stat-grid">
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Статус</span>
          <span className="rh-stat-tile__value">{statusInfo?.label ?? 'Не начато'}</span>
          <span className="rh-stat-tile__hint">Текущая стадия проверки профиля и документов.</span>
        </div>
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Тип лица</span>
          <span className="rh-stat-tile__value">
            {hasKyc ? (ENTITY_TYPE_LABELS[kyc.entity_type ?? ''] ?? kyc.entity_type ?? '—') : 'Не выбран'}
          </span>
          <span className="rh-stat-tile__hint">Влияет на набор документов и реквизитов заявки.</span>
        </div>
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Документы</span>
          <span className="rh-stat-tile__value">{documentUrls.length || kyc?.document_urls?.length || 0}</span>
          <span className="rh-stat-tile__hint">Загруженные файлы для проверки и повторной подачи.</span>
        </div>
      </div>

      {hasKyc && (
        <Card title="Текущий статус" className="rh-admin-detail-card">
          <Space orientation="vertical" size="middle" className="rh-full-width">
            <Space>
              <Text strong>Статус:</Text>
              {statusInfo ? (
                <Tag color={statusInfo.color} icon={statusInfo.icon}>
                  {statusInfo.label}
                </Tag>
              ) : (
                <Tag>Неизвестно</Tag>
              )}
            </Space>

            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="Тип лица">
                {ENTITY_TYPE_LABELS[kyc.entity_type ?? ''] ?? kyc.entity_type}
              </Descriptions.Item>
              <Descriptions.Item label="ФИО / Название">
                {kyc.full_name}
              </Descriptions.Item>
              <Descriptions.Item label="ИНН">
                {kyc.inn}
              </Descriptions.Item>
              {kyc.ogrnip && (
                <Descriptions.Item label="ОГРНИП">
                  {kyc.ogrnip}
                </Descriptions.Item>
              )}
              {kyc.company_name && (
                <Descriptions.Item label="Название компании">
                  {kyc.company_name}
                </Descriptions.Item>
              )}
              <Descriptions.Item label="Дата подачи">
                {kyc.submitted_at
                  ? new Date(kyc.submitted_at).toLocaleString('ru-RU')
                  : '-'}
              </Descriptions.Item>
              {kyc.reviewed_at && (
                <Descriptions.Item label="Дата проверки">
                  {new Date(kyc.reviewed_at).toLocaleString('ru-RU')}
                </Descriptions.Item>
              )}
              {kyc.expires_at && (
                <Descriptions.Item label="Действует до">
                  {new Date(kyc.expires_at).toLocaleString('ru-RU')}
                </Descriptions.Item>
              )}
            </Descriptions>

            {kyc.status === 'rejected' && kyc.rejection_reason && (
              <Alert
                type="error"
                showIcon
                title="Причина отклонения"
                description={kyc.rejection_reason}
              />
            )}

            {kyc.status === 'approved' && (
              <Alert
                type="success"
                showIcon
                title="Верификация пройдена"
                description="Вы можете создавать объявления на платформе."
              />
            )}

            {kyc.status === 'pending' && (
              <Alert
                type="info"
                showIcon
                title="Заявка на рассмотрении"
                description="Обычно проверка занимает 1-3 рабочих дня."
              />
            )}
          </Space>
        </Card>
      )}

      {(!hasKyc || kyc?.status === 'rejected') && (
        <Card title={hasKyc ? 'Подать заявку повторно' : 'Подать заявку на KYC'}>
          <Form
            form={form}
            layout="vertical"
            onFinish={handleSubmit}
            className="rh-kyc-form"
          >
            <Form.Item
              name="entity_type"
              label="Тип лица"
              rules={[{ required: true, message: 'Выберите тип лица' }]}
            >
              <Select
                placeholder="Выберите тип"
                options={ENTITY_TYPE_OPTIONS}
                onChange={(value) => setEntityType(value)}
              />
            </Form.Item>

            <Form.Item
              name="full_name"
              label="ФИО"
              rules={[{ required: true, message: 'Введите ФИО' }]}
            >
              <Input placeholder="Иванов Иван Иванович" />
            </Form.Item>

            <Form.Item
              name="inn"
              label="ИНН"
              rules={[{ required: true, message: 'Введите ИНН' }]}
            >
              <Input placeholder="1234567890" maxLength={12} />
            </Form.Item>

            {entityType === 'sole_proprietor' && (
              <Form.Item
                name="ogrnip"
                label="ОГРНИП"
              >
                <Input placeholder="123456789012345" maxLength={15} />
              </Form.Item>
            )}

            {entityType === 'legal_entity' && (
              <Form.Item
                name="company_name"
                label="Название компании"
              >
                <Input placeholder='ООО "Название"' />
              </Form.Item>
            )}

            <Form.Item label="Документы">
              <Upload
                multiple
                showUploadList
                beforeUpload={handleDocumentUpload}
                accept=".pdf,.jpg,.jpeg,.png"
              >
                <Button icon={<UploadOutlined />}>Загрузить документ</Button>
              </Upload>
              <Text type="secondary" className="rh-field-help-text">
                Паспорт, свидетельство о регистрации и другие подтверждающие документы
              </Text>
            </Form.Item>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                icon={<SendOutlined />}
                loading={submitKyc.isPending}
              >
                Подать заявку
              </Button>
            </Form.Item>
          </Form>
        </Card>
      )}
    </div>
  )
}
