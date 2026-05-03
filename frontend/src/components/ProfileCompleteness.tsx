import { useEffect, useState } from 'react'
import { Button, Card, Progress, List, Typography, Tag } from '@/components/design/system'
import { CheckCircleOutlined, CloseCircleOutlined } from '@/components/design/icons'
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
      className="rh-profile-completeness"
    >
      <Progress
        percent={data.percentage}
        strokeColor={strokeColor}
        className="rh-profile-completeness__progress"
      />
      <List
        size="small"
        dataSource={data.items.filter((i) => !i.complete)}
        renderItem={(item) => (
          <List.Item
            className={onNavigate ? 'rh-profile-completeness__item rh-profile-completeness__item--clickable' : 'rh-profile-completeness__item'}
            onClick={() => onNavigate?.(item.field)}
          >
            <List.Item.Meta
              avatar={
                item.complete ? (
                  <CheckCircleOutlined className="rh-status-icon rh-status-icon--success" />
                ) : (
                  <CloseCircleOutlined className="rh-status-icon rh-status-icon--danger" />
                )
              }
              title={<Text>{item.label}</Text>}
              description={
                <Text type="secondary" className="rh-table-meta-text">
                  {FIELD_ACTIONS[item.field]}
                </Text>
              }
            />
            <div className="rh-profile-completeness__actions">
              {!item.complete && (
                <Tag color="orange" className="rh-compact-tag">
                  Не заполнено
                </Tag>
              )}
              {onNavigate && !item.complete && (
                <Button
                  type="link"
                  size="small"
                  className="rh-link-action"
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
