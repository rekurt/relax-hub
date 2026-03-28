import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ObjectTypeManagement from '@/pages/admin/ObjectTypeManagement'

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
          <MemoryRouter initialEntries={['/admin/object-types']}>
            <Routes>
              <Route path="/admin/object-types" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockObjectTypes = [
  { id: 'o1', name: 'Русская баня', description: 'Традиционная русская баня', sort_order: 1, is_active: true, created_at: '2026-01-01', updated_at: '2026-01-01' },
  { id: 'o2', name: 'Финская сауна', description: 'Классическая финская сауна', sort_order: 2, is_active: true, created_at: '2026-01-01', updated_at: '2026-01-01' },
  { id: 'o3', name: 'Хаммам', description: 'Турецкая баня', sort_order: 3, is_active: false, created_at: '2026-01-01', updated_at: '2026-01-01' },
]

beforeEach(() => {
  vi.clearAllMocks()
  mockGet.mockResolvedValue({ data: { data: mockObjectTypes } })
  mockPost.mockResolvedValue({ data: { data: { id: 'new' } } })
  mockPut.mockResolvedValue({ data: { data: {} } })
  mockDelete.mockResolvedValue({ data: { data: {} } })
})

describe('ObjectTypeManagement', () => {
  it('renders page title and object type list', async () => {
    renderWithProviders(<ObjectTypeManagement />)

    await waitFor(() => {
      expect(screen.getByText('Типы объектов')).toBeInTheDocument()
      expect(screen.getByText('Русская баня')).toBeInTheDocument()
      expect(screen.getByText('Финская сауна')).toBeInTheDocument()
      expect(screen.getByText('Хаммам')).toBeInTheDocument()
    })
  })

  it('renders description column', async () => {
    renderWithProviders(<ObjectTypeManagement />)

    await waitFor(() => {
      expect(screen.getByText('Традиционная русская баня')).toBeInTheDocument()
      expect(screen.getByText('Классическая финская сауна')).toBeInTheDocument()
    })
  })

  it('renders active/inactive status tags', async () => {
    renderWithProviders(<ObjectTypeManagement />)

    await waitFor(() => {
      const activeTags = screen.getAllByText('Активно')
      expect(activeTags.length).toBe(2)
      expect(screen.getByText('Неактивно')).toBeInTheDocument()
    })
  })

  it('renders add button', async () => {
    renderWithProviders(<ObjectTypeManagement />)

    await waitFor(() => {
      expect(screen.getByText('Добавить тип')).toBeInTheDocument()
    })
  })

  it('opens create modal when clicking add button', async () => {
    renderWithProviders(<ObjectTypeManagement />)

    await waitFor(() => {
      expect(screen.getByText('Добавить тип')).toBeInTheDocument()
    })
    fireEvent.click(screen.getByText('Добавить тип'))

    await waitFor(() => {
      expect(screen.getByText('Добавить тип объекта')).toBeInTheDocument()
    })
  })

  it('shows edit and delete action buttons', async () => {
    renderWithProviders(<ObjectTypeManagement />)

    await waitFor(() => {
      const editButtons = screen.getAllByText('Изменить')
      expect(editButtons.length).toBe(3)
      const deleteButtons = screen.getAllByText('Удалить')
      expect(deleteButtons.length).toBe(3)
    })
  })

  it('shows empty state when no object types', async () => {
    mockGet.mockResolvedValue({ data: { data: [] } })

    renderWithProviders(<ObjectTypeManagement />)

    await waitFor(() => {
      expect(screen.getByText('Нет типов объектов')).toBeInTheDocument()
    })
  })
})
