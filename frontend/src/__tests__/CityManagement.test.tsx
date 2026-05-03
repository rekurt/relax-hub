import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import CityManagement from '@/pages/admin/CityManagement'

vi.mock('@/api/generated/cities/cities', () => ({
  useGetCities: vi.fn(),
  getGetCitiesQueryKey: vi.fn(() => ['/cities']),
}))

vi.mock('@/api/generated/admin-cities/admin-cities', () => ({
  usePostAdminCities: vi.fn(),
  usePutAdminCitiesId: vi.fn(),
  useDeleteAdminCitiesId: vi.fn(),
}))

import { useGetCities } from '@/api/generated/cities/cities'
import {
  usePostAdminCities,
  usePutAdminCitiesId,
  useDeleteAdminCitiesId,
} from '@/api/generated/admin-cities/admin-cities'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/cities']}>
            <Routes>
              <Route path="/admin/cities" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockCities = [
  { id: 1, name: 'Москва', slug: 'moscow', latitude: 55.7558, longitude: 37.6173 },
  { id: 2, name: 'Санкт-Петербург', slug: 'spb', latitude: 59.9343, longitude: 30.3351 },
  { id: 3, name: 'Казань', slug: 'kazan', latitude: 55.7887, longitude: 49.1221 },
]

const mutationDefault = { mutateAsync: vi.fn(), isPending: false }

beforeEach(() => {
  vi.mocked(useGetCities).mockReturnValue({
    data: { data: mockCities, success: true },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetCities>)

  vi.mocked(usePostAdminCities).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePostAdminCities>,
  )
  vi.mocked(usePutAdminCitiesId).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePutAdminCitiesId>,
  )
  vi.mocked(useDeleteAdminCitiesId).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof useDeleteAdminCitiesId>,
  )
})

describe('CityManagement', () => {
  it('renders page title and city list', () => {
    renderWithProviders(<CityManagement />)

    expect(screen.getByText('Управление городами')).toBeInTheDocument()
    expect(screen.getByText('Москва')).toBeInTheDocument()
    expect(screen.getByText('Санкт-Петербург')).toBeInTheDocument()
    expect(screen.getByText('Казань')).toBeInTheDocument()
  })

  it('renders slug column for cities', () => {
    renderWithProviders(<CityManagement />)

    expect(screen.getByText('moscow')).toBeInTheDocument()
    expect(screen.getByText('spb')).toBeInTheDocument()
    expect(screen.getByText('kazan')).toBeInTheDocument()
  })

  it('renders add city button', () => {
    renderWithProviders(<CityManagement />)

    expect(screen.getByText('Добавить город')).toBeInTheDocument()
  })

  it('opens create modal when clicking add button', async () => {
    renderWithProviders(<CityManagement />)

    fireEvent.click(screen.getByText('Добавить город'))

    await waitFor(() => {
      expect(screen.getByText('Добавить город', { selector: '.ant-modal-title' })).toBeInTheDocument()
    })
  })

  it('shows edit and delete action buttons for each city', () => {
    renderWithProviders(<CityManagement />)

    const editButtons = screen.getAllByText('Изменить')
    expect(editButtons.length).toBe(3)

    const deleteButtons = screen.getAllByText('Удалить')
    expect(deleteButtons.length).toBe(3)
  })

  it('opens edit modal with city data when clicking edit', async () => {
    renderWithProviders(<CityManagement />)

    const editButtons = screen.getAllByText('Изменить')
    fireEvent.click(editButtons[0]!)

    await waitFor(() => {
      expect(screen.getByText('Редактировать город')).toBeInTheDocument()
    })
  })

  it('renders empty state when no cities', () => {
    vi.mocked(useGetCities).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetCities>)

    renderWithProviders(<CityManagement />)

    expect(screen.getByText('Нет городов')).toBeInTheDocument()
  })

  it('renders loading state', () => {
    vi.mocked(useGetCities).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetCities>)

    renderWithProviders(<CityManagement />)

    expect(screen.getByText('Управление городами')).toBeInTheDocument()
    expect(document.querySelector('.ant-spin')).toBeTruthy()
  })

  it('shows delete confirmation when clicking delete', async () => {
    renderWithProviders(<CityManagement />)

    const deleteButtons = screen.getAllByText('Удалить')
    fireEvent.click(deleteButtons[0]!)

    await waitFor(() => {
      expect(screen.getAllByText('Удалить город?').length).toBeGreaterThanOrEqual(1)
    })
  })

  it('renders coordinate columns', () => {
    renderWithProviders(<CityManagement />)

    expect(screen.getByText('Широта')).toBeInTheDocument()
    expect(screen.getByText('Долгота')).toBeInTheDocument()
    expect(screen.getByText('55.7558')).toBeInTheDocument()
    expect(screen.getByText('37.6173')).toBeInTheDocument()
  })
})
