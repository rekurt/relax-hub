import { useState } from 'react'
import { Modal, Button, Typography } from '@/components/design/system'
import {
  SearchOutlined,
  CalendarOutlined,
  WalletOutlined,
  StarOutlined,
  EnvironmentOutlined,
  GiftOutlined,
} from '@/components/design/icons'
import { axiosInstance } from '@/api/axios-instance'
import { formatPrice } from '@/lib/format'

const { Title, Paragraph } = Typography

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
      icon: <GiftOutlined style={{ fontSize: 48, color: '#15803d' }} />,
      title: 'Добро пожаловать!',
      description:
        `Вам начислен приветственный бонус ${formatPrice(bonusAmount)}! Бонус действует 30 дней и может быть использован для оплаты первого бронирования.`,
      bonusAmount,
      currencySymbol,
    },
    {
      icon: <SearchOutlined style={{ fontSize: 48, color: '#0f766e' }} />,
      title: 'Поиск бань',
      description:
        'Используйте поиск, чтобы найти идеальную баню. Фильтруйте по городу, цене, удобствам и расположению на карте.',
    },
    {
      icon: <CalendarOutlined style={{ fontSize: 48, color: '#d97706' }} />,
      title: 'Бронирование',
      description:
        'Выберите дату и время, укажите количество гостей и дополнительные услуги. Оплатите онлайн картой, через СБП или из кошелька.',
    },
    {
      icon: <WalletOutlined style={{ fontSize: 48, color: '#d97706' }} />,
      title: 'Кошелёк и бонусы',
      description:
        'Пополняйте кошелёк для быстрой оплаты. Получайте кешбэк за бронирования, бонусы за приглашение друзей и повышайте уровень лояльности.',
    },
    {
      icon: <EnvironmentOutlined style={{ fontSize: 48, color: '#0f766e' }} />,
      title: 'Рекомендации рядом',
      description:
        'Разрешите определение местоположения, и мы покажем лучшие бани поблизости. Персональные рекомендации учитывают ваши предпочтения и историю посещений.',
    },
    {
      icon: <StarOutlined style={{ fontSize: 48, color: '#d97706' }} />,
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
      width="min(560px, calc(100vw - 32px))"
      centered
      className="rh-onboarding-modal"
    >
      <div className="rh-onboarding">
        <div className="rh-onboarding__hero">
          <div className="rh-onboarding__icon" aria-hidden="true">
            {step.icon}
          </div>
          <Title level={4} className="rh-onboarding__title">
            {step.title}
          </Title>
          <Paragraph type="secondary" className="rh-onboarding__description">
            {step.description}
          </Paragraph>
        </div>
        {current === 0 && (
          <div className="rh-onboarding__bonus" role="status">
            <GiftOutlined />
            <span>
              Приветственный бонус: <strong>{formatPrice(region === 'BY' ? WELCOME_BONUS_BY : WELCOME_BONUS_RU)}</strong>
            </span>
          </div>
        )}

        <div className="rh-onboarding__progress" role="tablist" aria-label="Шаги приветствия">
          {steps.map((item, index) => (
            <button
              key={item.title}
              type="button"
              className={`rh-onboarding__step${index === current ? ' rh-onboarding__step--active' : ''}${index < current ? ' rh-onboarding__step--done' : ''}`}
              aria-current={index === current ? 'step' : undefined}
              aria-label={`Шаг ${index + 1}: ${item.title}`}
              onClick={() => setCurrent(index)}
            >
              <span>{index + 1}</span>
            </button>
          ))}
        </div>

        <div className="rh-onboarding__footer">
          <Button onClick={handleComplete} type="link" className="rh-onboarding__skip">
            Пропустить
          </Button>
          <div className="rh-onboarding__actions">
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
          </div>
        </div>
      </div>
    </Modal>
  )
}
