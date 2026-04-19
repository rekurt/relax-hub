import { useState } from 'react'
import { App, Button, Form, Input, InputNumber, Modal, Spin, Switch, Table, Tag, Typography } from 'antd'
import { EditOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetAdminSettings,
  getGetAdminSettingsQueryKey,
  usePutAdminSettingsKey,
} from '@/api/generated/admin/admin'
import type { InternalHandlerPlatformSettingResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

const { Title, Text } = Typography
const { TextArea } = Input

const TYPE_LABELS: Record<string, { color: string; text: string }> = {
  int: { color: 'blue', text: 'Целое число' },
  float: { color: 'cyan', text: 'Дробное число' },
  string: { color: 'green', text: 'Строка' },
  bool: { color: 'orange', text: 'Логическое' },
  json: { color: 'purple', text: 'JSON' },
}

export default function PlatformSettings() {
  const { message } = App.useApp()
  const queryClient = useQueryClient()

  const [editingKey, setEditingKey] = useState<string | null>(null)
  const [editingType, setEditingType] = useState<string>('string')
  const [editingDescription, setEditingDescription] = useState<string>('')
  const [formValue, setFormValue] = useState<string>('')
  const [boolValue, setBoolValue] = useState(false)

  const { data, isLoading } = useGetAdminSettings()
  const updateMutation = usePutAdminSettingsKey()

  const settings: InternalHandlerPlatformSettingResponse[] = data?.data ?? []

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetAdminSettingsQueryKey() })
  }

  const openEditModal = (setting: InternalHandlerPlatformSettingResponse) => {
    setEditingKey(setting.key ?? '')
    setEditingType(setting.type ?? 'string')
    setEditingDescription(setting.description ?? '')
    if (setting.type === 'bool') {
      setBoolValue(setting.value === 'true')
    } else {
      setFormValue(setting.value ?? '')
    }
  }

  const handleSave = async () => {
    if (!editingKey) return
    const value = editingType === 'bool' ? String(boolValue) : formValue
    await updateMutation.mutateAsync({ key: editingKey, data: { value } })
    message.success('Настройка обновлена')
    setEditingKey(null)
    invalidate()
  }

  const renderValueInput = () => {
    switch (editingType) {
      case 'int':
        return (
          <InputNumber
            value={formValue ? Number(formValue) : undefined}
            onChange={(v) => setFormValue(String(v ?? ''))}
            style={{ width: '100%' }}
            precision={0}
          />
        )
      case 'float':
        return (
          <InputNumber
            value={formValue ? Number(formValue) : undefined}
            onChange={(v) => setFormValue(String(v ?? ''))}
            style={{ width: '100%' }}
            step={0.01}
          />
        )
      case 'bool':
        return (
          <Switch
            checked={boolValue}
            onChange={setBoolValue}
            checkedChildren="Вкл"
            unCheckedChildren="Выкл"
          />
        )
      case 'json':
        return (
          <TextArea
            value={formValue}
            onChange={(e) => setFormValue(e.target.value)}
            rows={6}
            placeholder="{}"
          />
        )
      default:
        return (
          <Input
            value={formValue}
            onChange={(e) => setFormValue(e.target.value)}
          />
        )
    }
  }

  const columns: ColumnsType<InternalHandlerPlatformSettingResponse> = [
    {
      title: 'Ключ',
      dataIndex: 'key',
      key: 'key',
      render: (key: string) => <Text code>{key}</Text>,
    },
    {
      title: 'Значение',
      dataIndex: 'value',
      key: 'value',
      render: (value: string, record) => {
        if (record.type === 'bool') {
          return <Tag color={value === 'true' ? 'green' : 'red'}>{value === 'true' ? 'Да' : 'Нет'}</Tag>
        }
        return <Text>{value}</Text>
      },
    },
    {
      title: 'Тип',
      dataIndex: 'type',
      key: 'type',
      width: 140,
      render: (type: string) => {
        const config = TYPE_LABELS[type] ?? { color: 'default', text: type }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: 'Описание',
      dataIndex: 'description',
      key: 'description',
      responsive: ['lg'] as const,
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
      render: (_: unknown, record: InternalHandlerPlatformSettingResponse) => (
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
    <div>
      <PageHeader
        eyebrow="Системные параметры"
        title="Настройки платформы"
        description="Служебные параметры и их значения в более читаемом рабочем виде."
      />

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: 48 }}><Spin size="large" /></div>
      ) : (
        <Table
          dataSource={settings}
          columns={columns}
          rowKey="key"
          pagination={false}
          size="middle"
          locale={{ emptyText: 'Нет настроек' }}
        />
      )}

      <Modal
        title={`Изменить: ${editingKey}`}
        open={!!editingKey}
        onCancel={() => setEditingKey(null)}
        onOk={handleSave}
        okText="Сохранить"
        cancelText="Отмена"
        confirmLoading={updateMutation.isPending}
      >
        {editingDescription && (
          <Text type="secondary" style={{ display: 'block', marginBottom: 12 }}>
            {editingDescription}
          </Text>
        )}
        <Form layout="vertical" style={{ marginTop: 8 }}>
          <Form.Item label="Значение">
            {renderValueInput()}
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
