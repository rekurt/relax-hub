import { useEffect, useState } from 'react'
import { Card, Progress, List, Typography, Tag } from 'antd'
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
  preferences: 'Настройте предпочтения',
  notification_settings: 'Настройте уведомления в разделе "Настройки уведомлений"',
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

  const strokeColor = data.percentage >= 80 ? '#52c41a' : data.percentage >= 50 ? '#faad14' : '#ff4d4f'

  return (
    <Card
      title="Заполненность профиля"
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
                  <CheckCircleOutlined style={{ color: '#52c41a' }} />
                ) : (
                  <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
                )
              }
              title={<Text>{item.label}</Text>}
              description={
                <Text type="secondary" style={{ fontSize: 12 }}>
                  {FIELD_ACTIONS[item.field]}
                </Text>
              }
            />
            {!item.complete && (
              <Tag color="orange" style={{ fontSize: 11 }}>
                Не заполнено
              </Tag>
            )}
          </List.Item>
        )}
      />
    </Card>
  )
}
