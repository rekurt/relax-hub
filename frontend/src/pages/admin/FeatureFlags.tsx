import { App, Select, Space, Spin, Switch, Table, Tag, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetAdminFeatureFlags,
  getGetAdminFeatureFlagsQueryKey,
  usePutAdminFeatureFlagsKey,
} from '@/api/generated/admin/admin'
import type { InternalHandlerFeatureFlagResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'
import { useState } from 'react'

const { Title, Text } = Typography

const REGION_OPTIONS = [
  { value: '', label: 'Все регионы' },
  { value: 'RU', label: 'Россия' },
  { value: 'BY', label: 'Беларусь' },
]

export default function FeatureFlags() {
  const { message } = App.useApp()
  const queryClient = useQueryClient()
  const [regionFilter, setRegionFilter] = useState('')

  const { data, isLoading } = useGetAdminFeatureFlags()
  const updateMutation = usePutAdminFeatureFlagsKey()

  const allFlags: InternalHandlerFeatureFlagResponse[] = data?.data ?? []
  const flags = regionFilter
    ? allFlags.filter((f) => !f.region || f.region === regionFilter)
    : allFlags

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetAdminFeatureFlagsQueryKey() })
  }

  const handleToggle = async (flag: InternalHandlerFeatureFlagResponse, enabled: boolean) => {
    if (!flag.key) return
    await updateMutation.mutateAsync({
      key: flag.key,
      data: { enabled, region: flag.region },
    })
    message.success(`Флаг "${flag.key}" ${enabled ? 'включён' : 'выключен'}`)
    invalidate()
  }

  const columns: ColumnsType<InternalHandlerFeatureFlagResponse> = [
    {
      title: 'Ключ',
      dataIndex: 'key',
      key: 'key',
      render: (key: string) => <Text code>{key}</Text>,
    },
    {
      title: 'Статус',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 120,
      render: (enabled: boolean, record) => (
        <Switch
          checked={enabled}
          onChange={(checked) => handleToggle(record, checked)}
          loading={updateMutation.isPending && updateMutation.variables?.key === record.key}
          checkedChildren="Вкл"
          unCheckedChildren="Выкл"
        />
      ),
    },
    {
      title: 'Регион',
      dataIndex: 'region',
      key: 'region',
      width: 120,
      render: (region: string) => {
        if (!region) return <Tag>Глобальный</Tag>
        return <Tag color={region === 'RU' ? 'blue' : 'green'}>{region}</Tag>
      },
    },
    {
      title: 'Описание',
      dataIndex: 'description',
      key: 'description',
      responsive: ['md'] as const,
    },
    {
      title: 'Обновлено',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 160,
      responsive: ['lg'] as const,
      render: (val: string) => val ? formatDateTime(val) : '—',
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>
          Функциональные флаги
        </Title>
        <Space>
          <Select
            value={regionFilter}
            onChange={setRegionFilter}
            options={REGION_OPTIONS}
            style={{ width: 160 }}
          />
        </Space>
      </div>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: 48 }}><Spin size="large" /></div>
      ) : (
        <Table
          dataSource={flags}
          columns={columns}
          rowKey={(record) => `${record.key}-${record.region ?? ''}`}
          pagination={false}
          size="middle"
          locale={{ emptyText: 'Нет флагов' }}
        />
      )}
    </div>
  )
}
