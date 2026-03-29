import { useEffect, useState } from 'react'
import { Alert, Progress } from 'antd'
import { useNavigate } from 'react-router-dom'
import { axiosInstance } from '@/api/axios-instance'
import BathhouseSearch from '@/pages/client/BathhouseSearch'

interface CompletenessData {
  percentage: number
  items: Array<{ field: string; label: string; complete: boolean }>
}

function ProfileNudge() {
  const [data, setData] = useState<CompletenessData | null>(null)
  const navigate = useNavigate()

  useEffect(() => {
    axiosInstance
      .get<{ success: boolean; data: CompletenessData }>('/my/profile-completeness')
      .then((res) => setData(res.data.data))
      .catch(() => {})
  }, [])

  if (!data || data.percentage === 100) return null

  const missing = data.items.filter((i) => !i.complete).map((i) => i.label)
  const strokeColor = data.percentage >= 80 ? '#52c41a' : data.percentage >= 50 ? '#faad14' : '#ff4d4f'

  return (
    <Alert
      type="info"
      showIcon
      closable
      style={{ marginBottom: 16 }}
      message={
        <span style={{ cursor: 'pointer' }} onClick={() => navigate('/client/profile')}>
          Заполните профиль ({data.percentage}%) — не хватает: {missing.join(', ')}
        </span>
      }
      description={
        <Progress percent={data.percentage} strokeColor={strokeColor} size="small" showInfo={false} />
      }
    />
  )
}

export default function ClientHome() {
  return (
    <>
      <ProfileNudge />
      <BathhouseSearch />
    </>
  )
}
