import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import GeoHeatmap from '@/pages/admin/GeoHeatmap'

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    get: vi.fn(),
  },
}))

import { axiosInstance } from '@/api/axios-instance'

const mockGet = vi.mocked(axiosInstance.get)

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/heatmap']}>
            <Routes>
              <Route path="/admin/heatmap" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockHeatmapResponse = {
  data: {
    success: true,
    data: {
      period: '30d',
      cell_size: 0.01,
      cells: [
        { latitude: 55.75, longitude: 37.62, listing_count: 5, booking_count: 12, search_count: 45 },
        { latitude: 55.76, longitude: 37.63, listing_count: 3, booking_count: 8, search_count: 30 },
        { latitude: 55.74, longitude: 37.61, listing_count: 7, booking_count: 20, search_count: 60 },
      ],
    },
  },
}

describe('GeoHeatmap', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGet.mockResolvedValue(mockHeatmapResponse)
  })

  it('renders page title', () => {
    renderWithProviders(<GeoHeatmap />)
    expect(screen.getByText('Тепловая карта спроса и предложения')).toBeTruthy()
  })

  it('renders period selector', () => {
    renderWithProviders(<GeoHeatmap />)
    expect(screen.getByText('День')).toBeTruthy()
    expect(screen.getByText('Неделя')).toBeTruthy()
    expect(screen.getByText('Месяц')).toBeTruthy()
    expect(screen.getByText('3 месяца')).toBeTruthy()
  })

  it('renders layer toggle', () => {
    renderWithProviders(<GeoHeatmap />)
    expect(screen.getByText('Спрос')).toBeTruthy()
    expect(screen.getByText('Предложение')).toBeTruthy()
  })

  it('renders summary statistics after data loads', async () => {
    renderWithProviders(<GeoHeatmap />)
    await waitFor(() => {
      expect(screen.getByText('15')).toBeTruthy() // total listings: 5+3+7
    })
    expect(screen.getByText('135')).toBeTruthy() // total searches: 45+30+60
    expect(screen.getByText('40')).toBeTruthy() // total bookings: 12+8+20
  })

  it('renders map container', () => {
    renderWithProviders(<GeoHeatmap />)
    expect(screen.getByTestId('heatmap-map')).toBeTruthy()
  })

  it('calls API with default params', async () => {
    renderWithProviders(<GeoHeatmap />)
    await waitFor(() => {
      expect(mockGet).toHaveBeenCalledWith('/admin/analytics/heatmap', {
        params: { period: '30d', cell_size: 0.01 },
      })
    })
  })

  it('shows error alert on API failure', async () => {
    mockGet.mockRejectedValueOnce(new Error('Network error'))
    renderWithProviders(<GeoHeatmap />)
    await waitFor(() => {
      expect(screen.getByText('Ошибка загрузки данных')).toBeTruthy()
    })
  })

  it('shows empty state when no cells returned', async () => {
    mockGet.mockResolvedValueOnce({
      data: { success: true, data: { period: '30d', cell_size: 0.01, cells: [] } },
    })
    renderWithProviders(<GeoHeatmap />)
    await waitFor(() => {
      expect(screen.getByText(/Нет данных для отображения за выбранный период/)).toBeTruthy()
    })
  })
})
