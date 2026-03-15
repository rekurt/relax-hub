import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import WidgetSettings from '@/pages/widget/WidgetSettings'

vi.mock('@/api/generated/widgets/widgets', () => ({
  useGetMyBathhousesIdWidgetCode: vi.fn(),
  useGetMyBathhousesIdWidgetKey: vi.fn(),
  usePostMyBathhousesIdWidgetKeyRegenerate: vi.fn(),
}))

vi.mock('@/stores/bathhouse', () => ({
  useBathhouseStore: vi.fn(),
}))

import {
  useGetMyBathhousesIdWidgetCode,
  useGetMyBathhousesIdWidgetKey,
  usePostMyBathhousesIdWidgetKeyRegenerate,
} from '@/api/generated/widgets/widgets'
import { useBathhouseStore } from '@/stores/bathhouse'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter>{ui}</MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

function mockBathhouseStore(id: string | null) {
  vi.mocked(useBathhouseStore).mockImplementation((selector) =>
    (selector as (state: { selectedBathhouseId: string | null }) => unknown)({
      selectedBathhouseId: id,
    }),
  )
}

const mockRegenerateMutation = { mutate: vi.fn(), isPending: false }

function setupDefaultMocks() {
  vi.mocked(usePostMyBathhousesIdWidgetKeyRegenerate).mockReturnValue(
    mockRegenerateMutation as unknown as ReturnType<typeof usePostMyBathhousesIdWidgetKeyRegenerate>,
  )
}

const mockWidgetCode = {
  code: '<div id="bani-widget" data-key="abc123"></div><script src="https://bani.ru/widget.js"></script>',
  api_key: 'abc123',
  script_url: 'https://bani.ru/widget.js',
  style_url: 'https://bani.ru/widget.css',
}

const mockWidgetKey = {
  api_key: 'abc123-def456-ghi789',
}

describe('WidgetSettings', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setupDefaultMocks()
  })

  it('shows prompt when no bathhouse selected', () => {
    mockBathhouseStore(null)
    vi.mocked(useGetMyBathhousesIdWidgetCode).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetCode>)
    vi.mocked(useGetMyBathhousesIdWidgetKey).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetKey>)

    renderWithProviders(<WidgetSettings />)

    expect(screen.getByText('Виджет бронирования')).toBeInTheDocument()
    expect(screen.getByText('Выберите баню для настройки виджета')).toBeInTheDocument()
  })

  it('renders widget settings form', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdWidgetCode).mockReturnValue({
      data: { data: mockWidgetCode, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetCode>)
    vi.mocked(useGetMyBathhousesIdWidgetKey).mockReturnValue({
      data: { data: mockWidgetKey, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetKey>)

    renderWithProviders(<WidgetSettings />)

    expect(screen.getByText('Настройки виджета')).toBeInTheDocument()
    expect(screen.getByText('Акцентный цвет')).toBeInTheDocument()
    expect(screen.getByText('Шрифт')).toBeInTheDocument()
    expect(screen.getByText('Показывать цену')).toBeInTheDocument()
    expect(screen.getByText('Показывать рейтинг')).toBeInTheDocument()
  })

  it('shows embed code', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdWidgetCode).mockReturnValue({
      data: { data: mockWidgetCode, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetCode>)
    vi.mocked(useGetMyBathhousesIdWidgetKey).mockReturnValue({
      data: { data: mockWidgetKey, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetKey>)

    renderWithProviders(<WidgetSettings />)

    expect(screen.getByText('Код для вставки')).toBeInTheDocument()
    expect(screen.getByText(mockWidgetCode.code)).toBeInTheDocument()
  })

  it('shows API key section', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdWidgetCode).mockReturnValue({
      data: { data: mockWidgetCode, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetCode>)
    vi.mocked(useGetMyBathhousesIdWidgetKey).mockReturnValue({
      data: { data: mockWidgetKey, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetKey>)

    renderWithProviders(<WidgetSettings />)

    expect(screen.getByText('API-ключ')).toBeInTheDocument()
    expect(screen.getByDisplayValue('abc123-def456-ghi789')).toBeInTheDocument()
    expect(screen.getByText('Перегенерировать ключ')).toBeInTheDocument()
  })

  it('shows regenerate confirmation', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdWidgetCode).mockReturnValue({
      data: { data: mockWidgetCode, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetCode>)
    vi.mocked(useGetMyBathhousesIdWidgetKey).mockReturnValue({
      data: { data: mockWidgetKey, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetKey>)

    renderWithProviders(<WidgetSettings />)

    fireEvent.click(screen.getByText('Перегенерировать ключ'))

    expect(screen.getByText('Перегенерировать API-ключ?')).toBeInTheDocument()
  })

  it('calls regenerate mutation on confirm', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdWidgetCode).mockReturnValue({
      data: { data: mockWidgetCode, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetCode>)
    vi.mocked(useGetMyBathhousesIdWidgetKey).mockReturnValue({
      data: { data: mockWidgetKey, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetKey>)

    renderWithProviders(<WidgetSettings />)

    fireEvent.click(screen.getByText('Перегенерировать ключ'))
    const confirmBtn = screen.getByRole('button', { name: 'Перегенерировать' })
    fireEvent.click(confirmBtn)

    expect(mockRegenerateMutation.mutate).toHaveBeenCalledWith({ id: 'bath-1' })
  })

  it('shows preview section with button', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdWidgetCode).mockReturnValue({
      data: { data: mockWidgetCode, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetCode>)
    vi.mocked(useGetMyBathhousesIdWidgetKey).mockReturnValue({
      data: { data: mockWidgetKey, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetKey>)

    renderWithProviders(<WidgetSettings />)

    expect(screen.getByText('Предпросмотр')).toBeInTheDocument()
    expect(screen.getByText('Забронировать')).toBeInTheDocument()
  })

  it('shows price and rating in preview by default', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdWidgetCode).mockReturnValue({
      data: { data: mockWidgetCode, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetCode>)
    vi.mocked(useGetMyBathhousesIdWidgetKey).mockReturnValue({
      data: { data: mockWidgetKey, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetKey>)

    renderWithProviders(<WidgetSettings />)

    expect(screen.getByText(/Цена от 2 000 ₽\/ч/)).toBeInTheDocument()
    expect(screen.getByText(/★ 4.8/)).toBeInTheDocument()
  })

  it('shows copy buttons', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdWidgetCode).mockReturnValue({
      data: { data: mockWidgetCode, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetCode>)
    vi.mocked(useGetMyBathhousesIdWidgetKey).mockReturnValue({
      data: { data: mockWidgetKey, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetKey>)

    renderWithProviders(<WidgetSettings />)

    expect(screen.getByText('Копировать')).toBeInTheDocument()
  })

  it('shows helper text for embed code', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdWidgetCode).mockReturnValue({
      data: { data: mockWidgetCode, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetCode>)
    vi.mocked(useGetMyBathhousesIdWidgetKey).mockReturnValue({
      data: { data: mockWidgetKey, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdWidgetKey>)

    renderWithProviders(<WidgetSettings />)

    expect(screen.getByText(/Вставьте этот код на ваш сайт/)).toBeInTheDocument()
  })
})
