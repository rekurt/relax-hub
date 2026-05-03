import { useState } from 'react'
import { Modal, Button, Steps, Typography, Space, Alert } from 'antd'
import {
  SearchOutlined,
  CalendarOutlined,
  WalletOutlined,
  StarOutlined,
  EnvironmentOutlined,
  GiftOutlined,
} from '@ant-design/icons'
import { axiosInstance } from '@/api/axios-instance'
import { formatPrice } from '@/lib/format'

const { Title, Paragraph, Text } = Typography

interface OnboardingTourProps {
  open: boolean
  onComplete: () => void
  region?: string
}

const WELCOME_BONUS_RU = 50000 // 500 RUB in kopecks
const WELCOME_BONUS_BY = 1500 // 15 BYN in kopecks

function getSteps(region?: string) {
  const bonusAmount = region === 'BY' ? WELCOME_BONUS_BY : WELCOME_BONUS_RU
  const currencySymbol = region === 'BY' ? 'BYN' : '₽'

  return [
    {
      icon: <GiftOutlined style={{ fontSize: 48, color: '#52c41a' }} />,
      title: 'Добро пожаловать!',
      description:
        `Вам начислен приветственный бонус ${formatPrice(bonusAmount)}! Бонус действует 30 дней и может быть использован для оплаты первого бронирования.`,
      bonusAmount,
      currencySymbol,
    },
    {
      icon: <SearchOutlined style={{ fontSize: 48, color: '#1677ff' }} />,
      title: 'Поиск бань',
      description:
        'Используйте поиск, чтобы найти идеальную баню. Фильтруйте по городу, цене, удобствам и расположению на карте.',
    },
    {
      icon: <CalendarOutlined style={{ fontSize: 48, color: '#fa8c16' }} />,
      title: 'Бронирование',
      description:
        'Выберите дату и время, укажите количество гостей и дополнительные услуги. Оплатите онлайн картой, через СБП или из кошелька.',
    },
    {
      icon: <WalletOutlined style={{ fontSize: 48, color: '#faad14' }} />,
      title: 'Кошелёк и бонусы',
      description:
        'Пополняйте кошелёк для быстрой оплаты. Получайте кешбэк за бронирования, бонусы за приглашение друзей и повышайте уровень лояльности.',
    },
    {
      icon: <EnvironmentOutlined style={{ fontSize: 48, color: '#13c2c2' }} />,
      title: 'Рекомендации рядом',
      description:
        'Разрешите определение местоположения, и мы покажем лучшие бани поблизости. Персональные рекомендации учитывают ваши предпочтения и историю посещений.',
    },
    {
      icon: <StarOutlined style={{ fontSize: 48, color: '#eb2f96' }} />,
      title: 'Отзывы и рейтинг',
      description:
        'Оставляйте отзывы после посещения. Оценивайте чистоту, точность описания, общение и цену. Ваши отзывы помогут другим.',
    },
  ]
}

export default function OnboardingTour({ open, onComplete, region }: OnboardingTourProps) {
  const [current, setCurrent] = useState(0)
  const steps = getSteps(region)

  const handleComplete = async () => {
    try {
      await axiosInstance.post('/my/onboarding/complete')
    } catch {
      // ignore - non-critical
    }
    onComplete()
  }

  const isLast = current === steps.length - 1
  const step = steps[current]!

  return (
    <Modal
      open={open}
      closable={false}
      footer={null}
      width={520}
      centered
    >
      <div style={{ textAlign: 'center', padding: '24px 0 8px' }}>
        {step.icon}
        <Title level={4} style={{ marginTop: 16 }}>
          {step.title}
        </Title>
        <Paragraph type="secondary" style={{ fontSize: 15, minHeight: 66 }}>
          {step.description}
        </Paragraph>
        {current === 0 && (
          <Alert
            type="success"
            showIcon
            icon={<GiftOutlined />}
            title={
              <Text strong>
                Приветственный бонус: {formatPrice(region === 'BY' ? WELCOME_BONUS_BY : WELCOME_BONUS_RU)}
              </Text>
            }
            style={{ marginTop: 8, textAlign: 'left' }}
          />
        )}
      </div>

      <Steps
        current={current}
        size="small"
        items={steps.map((_, i) => ({ title: '', key: i }))}
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
