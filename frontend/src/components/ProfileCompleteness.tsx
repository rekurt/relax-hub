import { useEffect, useState } from 'react'
import { Button, Card, Progress, List, Typography, Tag } from 'antd'
import { CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons'
import { axiosInstance } from '@/api/axios-instance'

const { Text } = Typography

interface CompletenessItem {
  field: string
  label: string
  complete: boolean
}

interface CompletenessData {
  percentage: number
  items: CompletenessItem[]
}

interface ProfileCompletenessProps {
  onNavigate?: (field: string) => void
}

const FIELD_ACTIONS: Record<string, string> = {
  name: 'Укажите имя в форме ниже',
  avatar: 'Загрузите фото в разделе "Аватар"',
  phone: 'Добавьте номер телефона',
  email: 'Укажите адрес электронной почты',
  preferences: 'Откройте страницу предпочтений и задайте критерии подбора',
  notification_settings: 'Откройте страницу уведомлений и настройте нужные события',
}

const FIELD_LINK_LABELS: Record<string, string> = {
  name: 'К форме',
  avatar: 'К аватару',
  phone: 'К форме',
  email: 'К форме',
  preferences: 'Открыть',
  notification_settings: 'Открыть',
}

export default function ProfileCompleteness({ onNavigate }: ProfileCompletenessProps) {
  const [data, setData] = useState<CompletenessData | null>(null)

  useEffect(() => {
    axiosInstance
      .get<{ success: boolean; data: CompletenessData }>('/my/profile-completeness')
      .then((res) => setData(res.data.data))
      .catch(() => {})
  }, [])

  if (!data) return null
  if (data.percentage === 100) return null

  const strokeColor = data.percentage >= 80 ? '#15803d' : data.percentage >= 50 ? '#d97706' : '#b42318'

  return (
    <Card
      title="Что ещё заполнить"
      size="small"
      style={{ marginBottom: 0 }}
    >
      <Progress
        percent={data.percentage}
        strokeColor={strokeColor}
        style={{ marginBottom: 16 }}
      />
      <List
        size="small"
        dataSource={data.items.filter((i) => !i.complete)}
        renderItem={(item) => (
          <List.Item
            style={{ cursor: onNavigate ? 'pointer' : 'default', padding: '6px 0' }}
            onClick={() => onNavigate?.(item.field)}
          >
            <List.Item.Meta
              avatar={
                item.complete ? (
                  <CheckCircleOutlined style={{ color: '#15803d' }} />
                ) : (
                  <CloseCircleOutlined style={{ color: '#b42318' }} />
                )
              }
              title={<Text>{item.label}</Text>}
              description={
                <Text type="secondary" style={{ fontSize: 12 }}>
                  {FIELD_ACTIONS[item.field]}
                </Text>
              }
            />
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              {!item.complete && (
                <Tag color="orange" style={{ fontSize: 11, marginInlineEnd: 0 }}>
                  Не заполнено
                </Tag>
              )}
              {onNavigate && !item.complete && (
                <Button
                  type="link"
                  size="small"
                  style={{ paddingInline: 0 }}
                  onClick={(event) => {
                    event.stopPropagation()
                    onNavigate(item.field)
                  }}
                >
                  {FIELD_LINK_LABELS[item.field] ?? 'Открыть'}
                </Button>
              )}
            </div>
          </List.Item>
        )}
      />
    </Card>
  )
}
