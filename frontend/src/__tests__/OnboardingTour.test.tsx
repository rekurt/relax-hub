import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import OnboardingTour from '@/components/OnboardingTour'

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    post: vi.fn().mockResolvedValue({ data: { success: true } }),
  },
}))

import { axiosInstance } from '@/api/axios-instance'

describe('OnboardingTour', () => {
  const mockOnComplete = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders first step when open', () => {
    render(<OnboardingTour open={true} onComplete={mockOnComplete} />)
    expect(screen.getByText('Поиск бань')).toBeDefined()
    expect(screen.getByText('Далее')).toBeDefined()
    expect(screen.getByText('Пропустить')).toBeDefined()
  })

  it('does not render when closed', () => {
    render(<OnboardingTour open={false} onComplete={mockOnComplete} />)
    expect(screen.queryByText('Поиск бань')).toBeNull()
  })

  it('navigates to next step', () => {
    render(<OnboardingTour open={true} onComplete={mockOnComplete} />)
    fireEvent.click(screen.getByText('Далее'))
    expect(screen.getByText('Бронирование')).toBeDefined()
  })

  it('navigates back', () => {
    render(<OnboardingTour open={true} onComplete={mockOnComplete} />)
    fireEvent.click(screen.getByText('Далее'))
    expect(screen.getByText('Бронирование')).toBeDefined()
    fireEvent.click(screen.getByText('Назад'))
    expect(screen.getByText('Поиск бань')).toBeDefined()
  })

  it('shows Начать button on last step', () => {
    render(<OnboardingTour open={true} onComplete={mockOnComplete} />)
    // Navigate to the last step (5 steps total)
    for (let i = 0; i < 4; i++) {
      fireEvent.click(screen.getByText('Далее'))
    }
    expect(screen.getByText('Ваш профиль')).toBeDefined()
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
    for (let i = 0; i < 4; i++) {
      fireEvent.click(screen.getByText('Далее'))
    }
    fireEvent.click(screen.getByText('Начать'))
    await waitFor(() => {
      expect(mockOnComplete).toHaveBeenCalled()
    })
  })
})
