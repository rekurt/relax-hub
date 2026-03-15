import { Typography } from 'antd'

const { Title } = Typography

export default function AdminDashboard() {
  return (
    <div>
      <Title level={3}>Панель администратора</Title>
      <Typography.Text type="secondary">Управление платформой</Typography.Text>
    </div>
  )
}
