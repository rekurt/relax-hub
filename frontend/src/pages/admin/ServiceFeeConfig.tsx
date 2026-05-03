import { useState } from 'react'
import {
  App,
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Spin,
  Table,
  Tag,
} from '@/components/design/system'
import { PlusOutlined, EditOutlined } from '@/components/design/icons'
import type { ColumnsType } from '@/components/design/types'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetAdminServiceFee,
  getGetAdminServiceFeeQueryKey,
  usePutAdminServiceFee,
} from '@/api/generated/admin/admin'
import type { InternalHandlerServiceFeeConfigResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

const REGION_OPTIONS = [
  { value: '*', label: 'Глобальный (по умолчанию)' },
  { value: 'RU', label: 'Россия' },
  { value: 'BY', label: 'Беларусь' },
]

interface FeeFormValues {
  region: string
  category: string
  fee_percent: number
}

export default function ServiceFeeConfig() {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [form] = Form.useForm<FeeFormValues>()

  const [formModalOpen, setFormModalOpen] = useState(false)
  const [editingConfig, setEditingConfig] = useState<InternalHandlerServiceFeeConfigResponse | null>(null)

  const { data, isLoading } = useGetAdminServiceFee()
  const upsertMutation = usePutAdminServiceFee()

  const configs: InternalHandlerServiceFeeConfigResponse[] = data?.data ?? []

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetAdminServiceFeeQueryKey() })
  }

  const openCreateModal = () => {
    setEditingConfig(null)
    form.resetFields()
    form.setFieldsValue({ region: '*', category: '', fee_percent: 10 })
    setFormModalOpen(true)
  }

  const openEditModal = (config: InternalHandlerServiceFeeConfigResponse) => {
    setEditingConfig(config)
    form.setFieldsValue({
      region: config.region ?? '*',
      category: config.category ?? '',
      fee_percent: config.fee_percent ?? 10,
    })
    setFormModalOpen(true)
  }

  const handleFormSubmit = async () => {
    const values = await form.validateFields()
    await upsertMutation.mutateAsync({
      data: {
        region: values.region,
        category: values.category || undefined,
        fee_percent: values.fee_percent,
      },
    })
    message.success(editingConfig ? 'Комиссия обновлена' : 'Комиссия создана')
    setFormModalOpen(false)
    invalidate()
  }

  const columns: ColumnsType<InternalHandlerServiceFeeConfigResponse> = [
    {
      title: 'Регион',
      dataIndex: 'region',
      key: 'region',
      width: 160,
      render: (region: string) => {
        if (!region || region === '*') return <Tag>Глобальный</Tag>
        return <Tag color={region === 'RU' ? 'blue' : 'green'}>{region}</Tag>
      },
    },
    {
      title: 'Категория',
      dataIndex: 'category',
      key: 'category',
      render: (category: string) => category || <Tag>Все категории</Tag>,
    },
    {
      title: 'Комиссия, %',
      dataIndex: 'fee_percent',
      key: 'fee_percent',
      width: 140,
      render: (val: number) => `${val}%`,
    },
    {
      title: 'Обновлено',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 160,
      responsive: ['md'] as const,
      render: (val: string) => val ? formatDateTime(val) : '—',
    },
    {
      title: '',
      key: 'actions',
      width: 100,
      render: (_: unknown, record: InternalHandlerServiceFeeConfigResponse) => (
        <Button
          type="link"
          size="small"
          icon={<EditOutlined />}
          onClick={() => openEditModal(record)}
        >
          Изменить
        </Button>
      ),
    },
  ]

  return (
    <div className="rh-stack rh-admin-reference-page">
      <PageHeader
        eyebrow="Финансовая политика"
        title="Комиссия платформы"
        description="Настройка комиссий по регионам и категориям теперь выглядит как управляемый справочник."
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
            Добавить
          </Button>
        }
      />

      <Card className="rh-admin-reference-card" title="Правила комиссий">
        {isLoading ? (
          <div className="rh-admin-state-card">
            <Spin size="large" />
            <span>Загружаем правила комиссий</span>
          </div>
        ) : configs.length === 0 ? (
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Нет настроек комиссии</div>
            <p className="rh-admin-empty-state__text">
              Добавьте глобальное правило или региональное переопределение, чтобы управлять комиссией из единого справочника.
            </p>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
              Добавить
            </Button>
          </div>
        ) : (
          <Table
            dataSource={configs}
            columns={columns}
            rowKey="id"
            pagination={false}
            size="middle"
            locale={{ emptyText: 'Нет настроек комиссии' }}
          />
        )}
      </Card>

      <Modal
        title={editingConfig ? 'Редактировать комиссию' : 'Добавить комиссию'}
        open={formModalOpen}
        onCancel={() => setFormModalOpen(false)}
        onOk={handleFormSubmit}
        okText={editingConfig ? 'Сохранить' : 'Создать'}
        cancelText="Отмена"
        confirmLoading={upsertMutation.isPending}
      >
        <Form form={form} layout="vertical" className="rh-admin-modal-form">
          <Form.Item
            name="region"
            label="Регион"
            rules={[{ required: true, message: 'Выберите регион' }]}
          >
            <Select options={REGION_OPTIONS} />
          </Form.Item>
          <Form.Item
            name="category"
            label="Категория (необязательно)"
          >
            <Input placeholder="Оставьте пустым для всех категорий" />
          </Form.Item>
          <Form.Item
            label="Процент комиссии"
          >
            <Space.Compact className="rh-compact-control">
              <Form.Item name="fee_percent" noStyle rules={[{ required: true, message: 'Укажите процент' }]}>
                <InputNumber
                  className="rh-admin-form-control"
                  min={0}
                  max={25}
                  step={0.5}
                />
              </Form.Item>
              <span className="rh-input-addon">%</span>
            </Space.Compact>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
