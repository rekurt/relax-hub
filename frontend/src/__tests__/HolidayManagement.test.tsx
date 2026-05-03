import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import HolidayManagement from '@/pages/admin/HolidayManagement'

const mockGet = vi.fn()
const mockPost = vi.fn()
const mockPut = vi.fn()
const mockDelete = vi.fn()

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    get: (...args: unknown[]) => mockGet(...args),
    post: (...args: unknown[]) => mockPost(...args),
    put: (...args: unknown[]) => mockPut(...args),
    delete: (...args: unknown[]) => mockDelete(...args),
  },
}))

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/holidays']}>
            <Routes>
              <Route path="/admin/holidays" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockHolidays = [
  { id: 'h1', name: 'Новый год', date: '2026-01-01', region: 'RU', is_recurring: true, created_at: '2026-01-01', updated_at: '2026-01-01' },
  { id: 'h2', name: 'День Победы', date: '2026-05-09', region: 'RU', is_recurring: true, created_at: '2026-01-01', updated_at: '2026-01-01' },
  { id: 'h3', name: 'День Независимости', date: '2026-07-03', region: 'BY', is_recurring: true, created_at: '2026-01-01', updated_at: '2026-01-01' },
]

beforeEach(() => {
  vi.clearAllMocks()
  mockGet.mockResolvedValue({ data: { data: mockHolidays } })
  mockPost.mockResolvedValue({ data: { data: { id: 'new' } } })
  mockPut.mockResolvedValue({ data: { data: {} } })
  mockDelete.mockResolvedValue({ data: { data: {} } })
})

describe('HolidayManagement', () => {
  it('renders page title and holiday list', async () => {
    renderWithProviders(<HolidayManagement />)

    await waitFor(() => {
      expect(screen.getByText('Праздничные дни')).toBeInTheDocument()
      expect(screen.getByText('Новый год')).toBeInTheDocument()
      expect(screen.getByText('День Победы')).toBeInTheDocument()
      expect(screen.getByText('День Независимости')).toBeInTheDocument()
    })
  })

  it('renders date column in DD.MM.YYYY format', async () => {
    renderWithProviders(<HolidayManagement />)

    await waitFor(() => {
      expect(screen.getByText('01.01.2026')).toBeInTheDocument()
      expect(screen.getByText('09.05.2026')).toBeInTheDocument()
    })
  })

  it('renders region labels', async () => {
    renderWithProviders(<HolidayManagement />)

    await waitFor(() => {
      const russiaLabels = screen.getAllByText('Россия')
      expect(russiaLabels.length).toBe(2)
      expect(screen.getByText('Беларусь')).toBeInTheDocument()
    })
  })

  it('renders recurring/one-time tags', async () => {
    renderWithProviders(<HolidayManagement />)

    await waitFor(() => {
      const recurringTags = screen.getAllByText('Ежегодно')
      expect(recurringTags.length).toBe(3)
    })
  })

  it('renders add holiday button', async () => {
    renderWithProviders(<HolidayManagement />)

    await waitFor(() => {
      expect(screen.getByText('Добавить праздник')).toBeInTheDocument()
    })
  })

  it('opens create modal when clicking add button', async () => {
    renderWithProviders(<HolidayManagement />)

    await waitFor(() => {
      expect(screen.getByText('Добавить праздник')).toBeInTheDocument()
    })
    fireEvent.click(screen.getByText('Добавить праздник'))

    await waitFor(() => {
      expect(screen.getByText('Добавить праздник', { selector: '.ant-modal-title' })).toBeInTheDocument()
    })
  })

  it('shows edit and delete action buttons', async () => {
    renderWithProviders(<HolidayManagement />)

    await waitFor(() => {
      const editButtons = screen.getAllByText('Изменить')
      expect(editButtons.length).toBe(3)
      const deleteButtons = screen.getAllByText('Удалить')
      expect(deleteButtons.length).toBe(3)
    })
  })

  it('shows empty state when no holidays', async () => {
    mockGet.mockResolvedValue({ data: { data: [] } })

    renderWithProviders(<HolidayManagement />)

    await waitFor(() => {
      expect(screen.getByText('Нет праздников')).toBeInTheDocument()
    })
  })

  it('has region filter select', async () => {
    renderWithProviders(<HolidayManagement />)

    await waitFor(() => {
      expect(screen.getByText('Праздничные дни')).toBeInTheDocument()
    })
    // Region filter should be present (Select component with placeholder)
    expect(document.querySelector('.ant-select')).toBeTruthy()
  })
})
