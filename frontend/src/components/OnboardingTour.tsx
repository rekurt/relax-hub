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

const { Paragraph } = Typography

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
      icon: <GiftOutlined className="rh-onboarding__step-icon rh-onboarding__step-icon--success" />,
      title: 'Добро пожаловать!',
      description:
        `Вам начислен приветственный бонус ${formatPrice(bonusAmount)}! Бонус действует 30 дней и может быть использован для оплаты первого бронирования.`,
      bonusAmount,
      currencySymbol,
    },
    {
      icon: <SearchOutlined className="rh-onboarding__step-icon rh-onboarding__step-icon--accent" />,
      title: 'Поиск бань',
      description:
        'Используйте поиск, чтобы найти идеальную баню. Фильтруйте по городу, цене, удобствам и расположению на карте.',
    },
    {
      icon: <CalendarOutlined className="rh-onboarding__step-icon rh-onboarding__step-icon--warning" />,
      title: 'Бронирование',
      description:
        'Выберите дату и время, укажите количество гостей и дополнительные услуги. Оплатите онлайн картой, через СБП или из кошелька.',
    },
    {
      icon: <WalletOutlined className="rh-onboarding__step-icon rh-onboarding__step-icon--warning" />,
      title: 'Кошелёк и бонусы',
      description:
        'Пополняйте кошелёк для быстрой оплаты. Получайте кешбэк за бронирования, бонусы за приглашение друзей и повышайте уровень лояльности.',
    },
    {
      icon: <EnvironmentOutlined className="rh-onboarding__step-icon rh-onboarding__step-icon--accent" />,
      title: 'Рекомендации рядом',
      description:
        'Разрешите определение местоположения, и мы покажем лучшие бани поблизости. Персональные рекомендации учитывают ваши предпочтения и историю посещений.',
    },
    {
      icon: <StarOutlined className="rh-onboarding__step-icon rh-onboarding__step-icon--warning" />,
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
          <h2 className="rh-onboarding__title">
            {step.title}
          </h2>
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
