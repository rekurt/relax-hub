import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi } from 'vitest'
import SavedCards from '@/pages/client/SavedCards'

vi.mock('@/api/generated/saved-cards/saved-cards', () => ({
  useGetMySavedCards: vi.fn(),
  useDeleteMySavedCardsId: vi.fn(),
  usePostMySavedCardsIdDefault: vi.fn(),
  getGetMySavedCardsQueryKey: vi.fn(() => ['/my/saved-cards']),
}))

import {
  useGetMySavedCards,
  useDeleteMySavedCardsId,
  usePostMySavedCardsIdDefault,
} from '@/api/generated/saved-cards/saved-cards'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/client/cards']}>
            <Routes>
              <Route path="/client/cards" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockCards = [
  {
    id: 'card-1',
    brand: 'Visa',
    last4: '4242',
    expiry_month: 12,
    expiry_year: 2027,
    is_default: true,
    created_at: '2026-01-15T10:00:00Z',
  },
  {
    id: 'card-2',
    brand: 'Mastercard',
    last4: '5555',
    expiry_month: 6,
    expiry_year: 2028,
    is_default: false,
    created_at: '2026-02-20T14:00:00Z',
  },
  {
    id: 'card-3',
    brand: 'Mir',
    last4: '2200',
    expiry_month: 3,
    expiry_year: 2029,
    is_default: false,
    created_at: '2026-03-10T09:00:00Z',
  },
]

const mockDeleteMutate = vi.fn()
const mockSetDefaultMutate = vi.fn()

function setupMocks(overrides?: {
  cards?: typeof mockCards
}) {
  vi.mocked(useGetMySavedCards).mockReturnValue({
    data: {
      data: overrides?.cards ?? mockCards,
      success: true,
      meta: {
        page: 1,
        page_size: 20,
        total_count: (overrides?.cards ?? mockCards).length,
        total_pages: 1,
      },
    },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMySavedCards>)

  vi.mocked(useDeleteMySavedCardsId).mockReturnValue({
    mutate: mockDeleteMutate,
    isPending: false,
  } as unknown as ReturnType<typeof useDeleteMySavedCardsId>)

  vi.mocked(usePostMySavedCardsIdDefault).mockReturnValue({
    mutate: mockSetDefaultMutate,
    isPending: false,
  } as unknown as ReturnType<typeof usePostMySavedCardsIdDefault>)
}

describe('SavedCards', () => {
  beforeEach(() => {
    mockDeleteMutate.mockReset()
    mockSetDefaultMutate.mockReset()
  })

  it('renders page title and statistics', () => {
    setupMocks()
    renderWithProviders(<SavedCards />)

    expect(screen.getByText('Сохранённые карты')).toBeInTheDocument()
    expect(screen.getByText('Сохранённых карт')).toBeInTheDocument()
    expect(screen.getByText('Карта по умолчанию')).toBeInTheDocument()
  })

  it('renders card list with brand and last4', () => {
    setupMocks()
    renderWithProviders(<SavedCards />)

    expect(screen.getAllByText(/4242/).length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText(/5555/)).toBeInTheDocument()
    expect(screen.getByText(/2200/)).toBeInTheDocument()
    expect(screen.getByText('Mastercard')).toBeInTheDocument()
    expect(screen.getByText('Mir')).toBeInTheDocument()
  })

  it('shows default tag on default card', () => {
    setupMocks()
    renderWithProviders(<SavedCards />)

    // Tag "По умолчанию" + buttons "По умолчанию" for non-default cards
    const matches = screen.getAllByText('По умолчанию')
    expect(matches.length).toBeGreaterThanOrEqual(1)
  })

  it('shows expiry dates', () => {
    setupMocks()
    renderWithProviders(<SavedCards />)

    expect(screen.getByText('12/27')).toBeInTheDocument()
    expect(screen.getByText('06/28')).toBeInTheDocument()
    expect(screen.getByText('03/29')).toBeInTheDocument()
  })

  it('shows delete confirmation dialog on click', async () => {
    setupMocks()
    renderWithProviders(<SavedCards />)

    const deleteButtons = screen.getAllByText('Удалить')
    fireEvent.click(deleteButtons[0]!)

    await waitFor(() => {
      expect(screen.getByText('Удалить карту?')).toBeInTheDocument()
      expect(screen.getByText('Карта будет удалена из сохранённых способов оплаты.')).toBeInTheDocument()
    })
  })

  it('calls delete mutation on confirm', async () => {
    setupMocks()
    renderWithProviders(<SavedCards />)

    const deleteButtons = screen.getAllByText('Удалить')
    fireEvent.click(deleteButtons[0]!)

    await waitFor(() => {
      expect(screen.getByText('Удалить карту?')).toBeInTheDocument()
    })

    const confirmButtons = screen.getAllByRole('button', { name: /Удалить/i })
    const popconfirmButton = confirmButtons.find(
      (btn) => btn.closest('.ant-popconfirm-buttons'),
    )
    if (popconfirmButton) {
      fireEvent.click(popconfirmButton)
    }

    await waitFor(() => {
      expect(mockDeleteMutate).toHaveBeenCalled()
    })
  })

  it('shows set default button for non-default cards', () => {
    setupMocks()
    renderWithProviders(<SavedCards />)

    const defaultButtons = screen.getAllByRole('button', { name: /По умолчанию/i })
    // Only non-default cards get the button (2 out of 3)
    expect(defaultButtons.length).toBe(2)
  })

  it('calls set default mutation on click', async () => {
    setupMocks()
    renderWithProviders(<SavedCards />)

    const defaultButtons = screen.getAllByRole('button', { name: /По умолчанию/i })
    fireEvent.click(defaultButtons[0]!)

    expect(mockSetDefaultMutate).toHaveBeenCalledWith({ id: 'card-2' })
  })

  it('renders empty state when no cards', () => {
    setupMocks({ cards: [] })
    renderWithProviders(<SavedCards />)

    expect(screen.getByText('У вас нет сохранённых карт')).toBeInTheDocument()
  })

  it('shows card count in statistics', () => {
    setupMocks()
    renderWithProviders(<SavedCards />)

    expect(screen.getByText('3')).toBeInTheDocument()
    expect(screen.getByText('/ 10')).toBeInTheDocument()
  })

  it('shows default card info in statistics', () => {
    setupMocks()
    renderWithProviders(<SavedCards />)

    expect(screen.getByText('Visa •••• 4242')).toBeInTheDocument()
  })

  it('shows "not selected" when no default card', () => {
    const cardsNoDefault = mockCards.map((c) => ({ ...c, is_default: false }))
    setupMocks({ cards: cardsNoDefault })
    renderWithProviders(<SavedCards />)

    expect(screen.getByText('Не выбрана')).toBeInTheDocument()
  })
})
