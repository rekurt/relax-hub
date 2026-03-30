import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import OnboardingTour from '@/components/OnboardingTour'

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    post: vi.fn().mockResolvedValue({ data: { success: true } }),
  },
}))

vi.mock('@/lib/format', () => ({
  formatPrice: (v: number) => `${v / 100} ₽`,
}))

import { axiosInstance } from '@/api/axios-instance'

describe('OnboardingTour', () => {
  const mockOnComplete = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders welcome bonus step first when open', () => {
    render(<OnboardingTour open={true} onComplete={mockOnComplete} />)
    expect(screen.getByText('Добро пожаловать!')).toBeDefined()
    expect(screen.getByText(/Приветственный бонус/)).toBeDefined()
    expect(screen.getByText('Далее')).toBeDefined()
    expect(screen.getByText('Пропустить')).toBeDefined()
  })

  it('shows BY region bonus amount', () => {
    render(<OnboardingTour open={true} onComplete={mockOnComplete} region="BY" />)
    // BY bonus = 1500 kopecks = 15 BYN - appears in description and alert
    expect(screen.getAllByText(/15 ₽/).length).toBeGreaterThanOrEqual(1)
  })

  it('shows RU region bonus amount by default', () => {
    render(<OnboardingTour open={true} onComplete={mockOnComplete} />)
    // RU bonus = 50000 kopecks = 500 RUB - appears in description and alert
    expect(screen.getAllByText(/500 ₽/).length).toBeGreaterThanOrEqual(1)
  })

  it('does not render when closed', () => {
    render(<OnboardingTour open={false} onComplete={mockOnComplete} />)
    expect(screen.queryByText('Добро пожаловать!')).toBeNull()
  })

  it('navigates through all 6 steps', () => {
    render(<OnboardingTour open={true} onComplete={mockOnComplete} />)

    // Step 1: Welcome bonus
    expect(screen.getByText('Добро пожаловать!')).toBeDefined()
    fireEvent.click(screen.getByText('Далее'))

    // Step 2: Search
    expect(screen.getByText('Поиск бань')).toBeDefined()
    fireEvent.click(screen.getByText('Далее'))

    // Step 3: Booking
    expect(screen.getByText('Бронирование')).toBeDefined()
    fireEvent.click(screen.getByText('Далее'))

    // Step 4: Wallet
    expect(screen.getByText('Кошелёк и бонусы')).toBeDefined()
    fireEvent.click(screen.getByText('Далее'))

    // Step 5: Nearby recommendations
    expect(screen.getByText('Рекомендации рядом')).toBeDefined()
    fireEvent.click(screen.getByText('Далее'))

    // Step 6: Reviews (last step)
    expect(screen.getByText('Отзывы и рейтинг')).toBeDefined()
    expect(screen.getByText('Начать')).toBeDefined()
  })

  it('navigates back', () => {
    render(<OnboardingTour open={true} onComplete={mockOnComplete} />)
    fireEvent.click(screen.getByText('Далее'))
    expect(screen.getByText('Поиск бань')).toBeDefined()
    fireEvent.click(screen.getByText('Назад'))
    expect(screen.getByText('Добро пожаловать!')).toBeDefined()
  })

  it('shows Начать button on last step', () => {
    render(<OnboardingTour open={true} onComplete={mockOnComplete} />)
    // Navigate to the last step (6 steps total)
    for (let i = 0; i < 5; i++) {
      fireEvent.click(screen.getByText('Далее'))
    }
    expect(screen.getByText('Отзывы и рейтинг')).toBeDefined()
    expect(screen.getByText('Начать')).toBeDefined()
  })

  it('calls onComplete and posts to API on skip', async () => {
    render(<OnboardingTour open={true} onComplete={mockOnComplete} />)
    fireEvent.click(screen.getByText('Пропустить'))
    await waitFor(() => {
      expect(axiosInstance.post).toHaveBeenCalledWith('/my/onboarding/complete')
      expect(mockOnComplete).toHaveBeenCalled()
    })
  })

  it('calls onComplete on finish', async () => {
    render(<OnboardingTour open={true} onComplete={mockOnComplete} />)
    for (let i = 0; i < 5; i++) {
      fireEvent.click(screen.getByText('Далее'))
    }
    fireEvent.click(screen.getByText('Начать'))
    await waitFor(() => {
      expect(mockOnComplete).toHaveBeenCalled()
    })
  })

  it('includes nearby recommendations step', () => {
    render(<OnboardingTour open={true} onComplete={mockOnComplete} />)
    // Navigate to step 5 (index 4)
    for (let i = 0; i < 4; i++) {
      fireEvent.click(screen.getByText('Далее'))
    }
    expect(screen.getByText('Рекомендации рядом')).toBeDefined()
    expect(screen.getByText(/местоположени/)).toBeDefined()
  })
})
