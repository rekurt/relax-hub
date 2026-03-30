import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi } from 'vitest'
import LoyaltyManagement from '@/pages/admin/LoyaltyManagement'

vi.mock('@/api/generated/loyalty/loyalty', () => ({
  useGetMyLoyaltyLevels: vi.fn(),
}))

import { useGetMyLoyaltyLevels } from '@/api/generated/loyalty/loyalty'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/loyalty']}>
            <Routes>
              <Route path="/admin/loyalty" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockLevels = [
  { level: 'bronze', min_visits: 0, discount_percent: 0, point_multiplier: 1 },
  { level: 'silver', min_visits: 3, discount_percent: 3, point_multiplier: 1.2 },
  { level: 'gold', min_visits: 10, discount_percent: 5, point_multiplier: 1.5 },
  { level: 'platinum', min_visits: 25, discount_percent: 10, point_multiplier: 2 },
]

function setupMocks() {
  vi.mocked(useGetMyLoyaltyLevels).mockReturnValue({
    data: { data: mockLevels, success: true },
    isLoading: false,
  } as ReturnType<typeof useGetMyLoyaltyLevels>)
}

describe('LoyaltyManagement', () => {
  it('renders title', () => {
    setupMocks()
    renderWithProviders(<LoyaltyManagement />)

    expect(screen.getByText('Управление программой лояльности')).toBeInTheDocument()
  })

  it('renders all four loyalty levels', () => {
    setupMocks()
    renderWithProviders(<LoyaltyManagement />)

    expect(screen.getAllByText('Бронза').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Серебро').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Золото').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Платина').length).toBeGreaterThanOrEqual(1)
  })

  it('renders level cards with visit thresholds', () => {
    setupMocks()
    renderWithProviders(<LoyaltyManagement />)

    expect(screen.getByText('от 0 визитов')).toBeInTheDocument()
    expect(screen.getByText('от 3 визитов')).toBeInTheDocument()
    expect(screen.getByText('от 10 визитов')).toBeInTheDocument()
    expect(screen.getByText('от 25 визитов')).toBeInTheDocument()
  })

  it('renders view buttons for each level', () => {
    setupMocks()
    renderWithProviders(<LoyaltyManagement />)

    const viewButtons = screen.getAllByText('Просмотр')
    expect(viewButtons.length).toBe(4)
  })

  it('opens view modal on click', () => {
    setupMocks()
    renderWithProviders(<LoyaltyManagement />)

    const viewButtons = screen.getAllByText('Просмотр')
    fireEvent.click(viewButtons[0]!)

    expect(screen.getByText(/Уровень:/)).toBeInTheDocument()
    expect(screen.getByText('Минимальное количество визитов')).toBeInTheDocument()
    expect(screen.getByText('Процент кэшбэка')).toBeInTheDocument()
    // "Множитель баллов" appears in both table and modal
    expect(screen.getAllByText('Множитель баллов').length).toBeGreaterThanOrEqual(2)
  })

  it('renders settings table header', () => {
    setupMocks()
    renderWithProviders(<LoyaltyManagement />)

    expect(screen.getByText('Настройка уровней')).toBeInTheDocument()
  })

  it('renders cashback info in table', () => {
    setupMocks()
    renderWithProviders(<LoyaltyManagement />)

    expect(screen.getByText('0%')).toBeInTheDocument()
    expect(screen.getByText('3%')).toBeInTheDocument()
    expect(screen.getByText('5%')).toBeInTheDocument()
    expect(screen.getByText('10%')).toBeInTheDocument()
  })

  it('renders point multipliers in table', () => {
    setupMocks()
    renderWithProviders(<LoyaltyManagement />)

    expect(screen.getByText('×1')).toBeInTheDocument()
    expect(screen.getByText('×1.2')).toBeInTheDocument()
    expect(screen.getByText('×1.5')).toBeInTheDocument()
    expect(screen.getByText('×2')).toBeInTheDocument()
  })

  it('shows loading state', () => {
    vi.mocked(useGetMyLoyaltyLevels).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as ReturnType<typeof useGetMyLoyaltyLevels>)

    renderWithProviders(<LoyaltyManagement />)
    expect(screen.getByText('Управление программой лояльности')).toBeInTheDocument()
  })

  it('shows empty state when no levels', () => {
    vi.mocked(useGetMyLoyaltyLevels).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as ReturnType<typeof useGetMyLoyaltyLevels>)

    renderWithProviders(<LoyaltyManagement />)
    expect(screen.getByText('Нет уровней лояльности')).toBeInTheDocument()
  })
})
