import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import AmenityManagement from '@/pages/admin/AmenityManagement'

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
          <MemoryRouter initialEntries={['/admin/amenities']}>
            <Routes>
              <Route path="/admin/amenities" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockAmenities = [
  { id: 'a1', name: 'Бассейн', icon: 'pool', sort_order: 1, is_active: true, created_at: '2026-01-01', updated_at: '2026-01-01' },
  { id: 'a2', name: 'Сауна', icon: 'sauna', sort_order: 2, is_active: true, created_at: '2026-01-01', updated_at: '2026-01-01' },
  { id: 'a3', name: 'Мангал', icon: 'bbq', sort_order: 3, is_active: false, created_at: '2026-01-01', updated_at: '2026-01-01' },
]

beforeEach(() => {
  vi.clearAllMocks()
  mockGet.mockResolvedValue({ data: { data: mockAmenities } })
  mockPost.mockResolvedValue({ data: { data: { id: 'new' } } })
  mockPut.mockResolvedValue({ data: { data: {} } })
  mockDelete.mockResolvedValue({ data: { data: {} } })
})

describe('AmenityManagement', () => {
  it('renders page title and amenity list', async () => {
    renderWithProviders(<AmenityManagement />)

    await waitFor(() => {
      expect(screen.getByText('Управление удобствами')).toBeInTheDocument()
      expect(screen.getByText('Бассейн')).toBeInTheDocument()
      expect(screen.getByText('Сауна')).toBeInTheDocument()
      expect(screen.getByText('Мангал')).toBeInTheDocument()
    })
  })

  it('renders icon column', async () => {
    renderWithProviders(<AmenityManagement />)

    await waitFor(() => {
      expect(screen.getByText('pool')).toBeInTheDocument()
      expect(screen.getByText('sauna')).toBeInTheDocument()
    })
  })

  it('renders active/inactive status tags', async () => {
    renderWithProviders(<AmenityManagement />)

    await waitFor(() => {
      const activeTags = screen.getAllByText('Активно')
      expect(activeTags.length).toBe(2)
      expect(screen.getByText('Неактивно')).toBeInTheDocument()
    })
  })

  it('renders add amenity button', async () => {
    renderWithProviders(<AmenityManagement />)

    await waitFor(() => {
      expect(screen.getByText('Добавить удобство')).toBeInTheDocument()
    })
  })

  it('opens create modal when clicking add button', async () => {
    renderWithProviders(<AmenityManagement />)

    await waitFor(() => {
      expect(screen.getByText('Добавить удобство')).toBeInTheDocument()
    })
    fireEvent.click(screen.getByText('Добавить удобство'))

    await waitFor(() => {
      expect(screen.getByText('Добавить удобство', { selector: '.ant-modal-title' })).toBeInTheDocument()
    })
  })

  it('shows edit and delete action buttons', async () => {
    renderWithProviders(<AmenityManagement />)

    await waitFor(() => {
      const editButtons = screen.getAllByText('Изменить')
      expect(editButtons.length).toBe(3)
      const deleteButtons = screen.getAllByText('Удалить')
      expect(deleteButtons.length).toBe(3)
    })
  })

  it('shows empty state when no amenities', async () => {
    mockGet.mockResolvedValue({ data: { data: [] } })

    renderWithProviders(<AmenityManagement />)

    await waitFor(() => {
      expect(screen.getByText('Нет удобств')).toBeInTheDocument()
    })
  })
})
