import { App, Card, Select, Spin, Switch, Table, Tag, Typography } from '@/components/design/system'
import type { ColumnsType } from '@/components/design/types'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetAdminFeatureFlags,
  getGetAdminFeatureFlagsQueryKey,
  usePutAdminFeatureFlagsKey,
} from '@/api/generated/admin/admin'
import type { InternalHandlerFeatureFlagResponse } from '@/api/generated/model'
import { formatDateTime } from '@/lib/format'
import { useState } from 'react'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

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
  const enabledCount = allFlags.filter((flag) => flag.enabled).length
  const globalCount = allFlags.filter((flag) => !flag.region).length
  const regionalCount = allFlags.length - globalCount

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
    <div className="rh-stack rh-admin-reference-page">
      <PageHeader
        eyebrow="Администрирование"
        title="Функциональные флаги"
        description="Контроль включения функций по регионам без визуального расхождения с рабочими справочниками платформы."
        extra={
          <Select
            className="rh-admin-filter-select"
            value={regionFilter}
            onChange={setRegionFilter}
            options={REGION_OPTIONS}
          />
        }
      />

      <div className="rh-stat-grid">
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Всего флагов</span>
          <span className="rh-stat-tile__value">{allFlags.length}</span>
          <span className="rh-stat-tile__hint">Полный реестр доступных feature-toggle параметров.</span>
        </div>
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Включены</span>
          <span className="rh-stat-tile__value">{enabledCount}</span>
          <span className="rh-stat-tile__hint">Активные флаги в текущей конфигурации.</span>
        </div>
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Региональные</span>
          <span className="rh-stat-tile__value">{regionalCount}</span>
          <span className="rh-stat-tile__hint">Переопределения для отдельных стран и рынков.</span>
        </div>
      </div>

      <Card className="rh-admin-reference-card" title="Реестр флагов">
        {isLoading ? (
          <div className="rh-admin-state-card">
            <Spin size="large" />
            <span>Загружаем функциональные флаги</span>
          </div>
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
      </Card>
    </div>
  )
}
