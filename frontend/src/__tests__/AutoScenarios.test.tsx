import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import AutoScenarios from '@/pages/crm/AutoScenarios'

vi.mock('@/api/generated/crm/crm', () => ({
  useGetMyCrmAutoScenarios: vi.fn(),
  usePutMyCrmAutoScenariosType: vi.fn(),
}))

import {
  useGetMyCrmAutoScenarios,
  usePutMyCrmAutoScenariosType,
} from '@/api/generated/crm/crm'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU} theme={{ token: { motion: false } }}>
        <AntApp>
          <MemoryRouter>{ui}</MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockScenarios = [
  {
    id: 'sc-1',
    owner_id: 'owner-1',
    type: 'thank_after_visit',
    name: 'Благодарность за визит',
    description: 'Отправить благодарность после визита',
    enabled: true,
    channel: 'push',
    delay_hours: 1,
    custom_text: '',
  },
  {
    id: 'sc-2',
    owner_id: 'owner-1',
    type: 'request_review',
    name: 'Запрос отзыва',
    description: 'Попросить оставить отзыв',
    enabled: false,
    channel: 'email',
    delay_hours: 2,
    custom_text: '',
  },
  {
    id: 'sc-3',
    owner_id: 'owner-1',
    type: 'birthday_greeting',
    name: 'Поздравление с днём рождения',
    description: 'Отправить поздравление',
    enabled: true,
    channel: 'push',
    delay_hours: 0,
    custom_text: 'С днём рождения! Дарим скидку 20%',
  },
]

const mockUpdateMutation = { mutate: vi.fn(), isPending: false }

describe('AutoScenarios', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(usePutMyCrmAutoScenariosType).mockReturnValue(
      mockUpdateMutation as unknown as ReturnType<typeof usePutMyCrmAutoScenariosType>,
    )
  })

  it('renders page title', () => {
    vi.mocked(useGetMyCrmAutoScenarios).mockReturnValue({
      data: { data: mockScenarios },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmAutoScenarios>)

    renderWithProviders(<AutoScenarios />)

    expect(screen.getByText('Автоматические сценарии')).toBeInTheDocument()
  })

  it('renders scenario names and descriptions', () => {
    vi.mocked(useGetMyCrmAutoScenarios).mockReturnValue({
      data: { data: mockScenarios },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmAutoScenarios>)

    renderWithProviders(<AutoScenarios />)

    expect(screen.getByText('Благодарность за визит')).toBeInTheDocument()
    expect(screen.getByText('Запрос отзыва')).toBeInTheDocument()
    expect(screen.getByText('Поздравление с днём рождения')).toBeInTheDocument()
  })

  it('renders toggle switches', () => {
    vi.mocked(useGetMyCrmAutoScenarios).mockReturnValue({
      data: { data: mockScenarios },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmAutoScenarios>)

    renderWithProviders(<AutoScenarios />)

    const switches = screen.getAllByRole('switch')
    expect(switches).toHaveLength(3)
  })

  it('shows config fields for enabled scenarios', () => {
    vi.mocked(useGetMyCrmAutoScenarios).mockReturnValue({
      data: { data: mockScenarios },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmAutoScenarios>)

    renderWithProviders(<AutoScenarios />)

    // Enabled scenarios show channel selector and delay input
    const channelLabels = screen.getAllByText('Канал')
    expect(channelLabels.length).toBe(2) // 2 enabled scenarios
  })
})
