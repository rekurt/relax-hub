import { Typography } from 'antd'

const { Title } = Typography

export default function ClientHome() {
  return (
    <div>
      <Title level={3}>Поиск бань</Title>
      <Typography.Text type="secondary">Найдите идеальную баню для отдыха</Typography.Text>
    </div>
  )
}
