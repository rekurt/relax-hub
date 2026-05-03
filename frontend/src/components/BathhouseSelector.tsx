import { Select, Space, Typography } from '@/components/design/system'
import { useGetMyBathhouses } from '@/api/generated/bathhouses/bathhouses'
import { useBathhouseStore } from '@/stores/bathhouse'
import { useEffect, useMemo } from 'react'

const { Text } = Typography

export default function BathhouseSelector() {
  const { selectedBathhouseId, setSelectedBathhouseId } = useBathhouseStore()
  const { data, isLoading } = useGetMyBathhouses({ page: 1, page_size: 100 })

  const bathhouses = useMemo(() => data?.data ?? [], [data?.data])

  useEffect(() => {
    if (bathhouses.length > 0) {
      if (!selectedBathhouseId || !bathhouses.some((b) => b.id === selectedBathhouseId)) {
        setSelectedBathhouseId(bathhouses[0]?.id ?? null)
      }
    }
  }, [bathhouses, selectedBathhouseId, setSelectedBathhouseId])

  if (bathhouses.length === 0 && !isLoading) {
    return <Text type="secondary">Нет бань</Text>
  }

  return (
    <Space>
      <Text type="secondary">Баня:</Text>
      <Select
        value={selectedBathhouseId}
        onChange={setSelectedBathhouseId}
        loading={isLoading}
        className="rh-bathhouse-selector__select"
        popupMatchSelectWidth={false}
        placeholder="Выберите баню"
        options={bathhouses.map((b) => ({
          label: b.name ?? 'Без названия',
          value: b.id,
        }))}
      />
    </Space>
  )
}
