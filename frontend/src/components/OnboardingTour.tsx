import { useState } from 'react'
import { Modal, Button, Steps, Typography, Space } from 'antd'
import {
  SearchOutlined,
  CalendarOutlined,
  WalletOutlined,
  StarOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { axiosInstance } from '@/api/axios-instance'

const { Title, Paragraph } = Typography

interface OnboardingTourProps {
  open: boolean
  onComplete: () => void
}

const STEPS = [
  {
    icon: <SearchOutlined style={{ fontSize: 48, color: '#1677ff' }} />,
    title: 'Поиск бань',
    description:
      'Используйте поиск, чтобы найти идеальную баню. Фильтруйте по городу, цене, удобствам и расположению на карте.',
  },
  {
    icon: <CalendarOutlined style={{ fontSize: 48, color: '#52c41a' }} />,
    title: 'Бронирование',
    description:
      'Выберите дату и время, укажите количество гостей и дополнительные услуги. Оплатите онлайн картой или через кошелёк.',
  },
  {
    icon: <WalletOutlined style={{ fontSize: 48, color: '#faad14' }} />,
    title: 'Кошелёк и бонусы',
    description:
      'Пополняйте кошелёк для быстрой оплаты. Получайте бонусы за бронирования и приглашение друзей.',
  },
  {
    icon: <StarOutlined style={{ fontSize: 48, color: '#eb2f96' }} />,
    title: 'Отзывы и рейтинг',
    description:
      'Оставляйте отзывы после посещения. Оценивайте чистоту, точность описания, общение и цену. Ваши отзывы помогут другим.',
  },
  {
    icon: <UserOutlined style={{ fontSize: 48, color: '#722ed1' }} />,
    title: 'Ваш профиль',
    description:
      'Заполните профиль для персональных рекомендаций. Укажите предпочтения, город и загрузите фото.',
  },
]

export default function OnboardingTour({ open, onComplete }: OnboardingTourProps) {
  const [current, setCurrent] = useState(0)

  const handleComplete = async () => {
    try {
      await axiosInstance.post('/my/onboarding/complete')
    } catch {
      // ignore - non-critical
    }
    onComplete()
  }

  const isLast = current === STEPS.length - 1

  return (
    <Modal
      open={open}
      closable={false}
      footer={null}
      width={520}
      centered
    >
      <div style={{ textAlign: 'center', padding: '24px 0 8px' }}>
        {STEPS[current]!.icon}
        <Title level={4} style={{ marginTop: 16 }}>
          {STEPS[current]!.title}
        </Title>
        <Paragraph type="secondary" style={{ fontSize: 15, minHeight: 66 }}>
          {STEPS[current]!.description}
        </Paragraph>
      </div>

      <Steps
        current={current}
        size="small"
        items={STEPS.map((_, i) => ({ title: '', key: i }))}
        style={{ marginBottom: 24 }}
      />

      <Space style={{ width: '100%', justifyContent: 'space-between' }}>
        <Button onClick={handleComplete} type="link">
          Пропустить
        </Button>
        <Space>
          {current > 0 && (
            <Button onClick={() => setCurrent(current - 1)}>Назад</Button>
          )}
          {isLast ? (
            <Button type="primary" onClick={handleComplete}>
              Начать
            </Button>
          ) : (
            <Button type="primary" onClick={() => setCurrent(current + 1)}>
              Далее
            </Button>
          )}
        </Space>
      </Space>
    </Modal>
  )
}
