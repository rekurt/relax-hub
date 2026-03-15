import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import BathhouseForm from '@/pages/bathhouses/BathhouseForm'

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetBathhousesId: vi.fn(),
  usePostBathhouses: vi.fn(),
  usePutBathhousesId: vi.fn(),
}))

vi.mock('@/api/generated/cities/cities', () => ({
  useGetCities: vi.fn(),
}))

import {
  useGetBathhousesId,
  usePostBathhouses,
  usePutBathhousesId,
} from '@/api/generated/bathhouses/bathhouses'
import { useGetCities } from '@/api/generated/cities/cities'

function renderWithProviders(ui: React.ReactElement, route = '/bathhouses/new') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={[route]}>
            <Routes>
              <Route path="/bathhouses/new" element={ui} />
              <Route path="/bathhouses/:id/edit" element={ui} />
              <Route path="/bathhouses" element={<div>list</div>} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockCities = [
  { id: 1, name: 'Москва', slug: 'moscow' },
  { id: 2, name: 'Санкт-Петербург', slug: 'spb' },
]

describe('BathhouseForm', () => {
  beforeEach(() => {
    vi.mocked(useGetCities).mockReturnValue({
      data: { data: mockCities, success: true },
    } as unknown as ReturnType<typeof useGetCities>)

    vi.mocked(usePostBathhouses).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePostBathhouses>)

    vi.mocked(usePutBathhousesId).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as unknown as ReturnType<typeof usePutBathhousesId>)

    vi.mocked(useGetBathhousesId).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesId>)
  })

  it('renders create form with title', () => {
    renderWithProviders(<BathhouseForm />)

    expect(screen.getByText('Новая баня')).toBeInTheDocument()
    expect(screen.getByText('Создать')).toBeInTheDocument()
    expect(screen.getByText('Отмена')).toBeInTheDocument()
  })

  it('renders all form sections', () => {
    renderWithProviders(<BathhouseForm />)

    expect(screen.getByText('Основная информация')).toBeInTheDocument()
    expect(screen.getByText('Параметры')).toBeInTheDocument()
    expect(screen.getByText('Удобства')).toBeInTheDocument()
    expect(screen.getByText('Рабочие часы')).toBeInTheDocument()
    expect(screen.getByText('Изображения')).toBeInTheDocument()
  })

  it('renders amenity checkboxes', () => {
    renderWithProviders(<BathhouseForm />)

    expect(screen.getByText('Сауна')).toBeInTheDocument()
    expect(screen.getByText('Парная')).toBeInTheDocument()
    expect(screen.getByText('Бассейн')).toBeInTheDocument()
    expect(screen.getByText('Купель')).toBeInTheDocument()
    expect(screen.getByText('Мангал')).toBeInTheDocument()
    expect(screen.getByText('Караоке')).toBeInTheDocument()
  })

  it('renders working hours for all days', () => {
    renderWithProviders(<BathhouseForm />)

    expect(screen.getByText('Пн')).toBeInTheDocument()
    expect(screen.getByText('Вт')).toBeInTheDocument()
    expect(screen.getByText('Ср')).toBeInTheDocument()
    expect(screen.getByText('Чт')).toBeInTheDocument()
    expect(screen.getByText('Пт')).toBeInTheDocument()
    expect(screen.getByText('Сб')).toBeInTheDocument()
    expect(screen.getByText('Вс')).toBeInTheDocument()
  })

  it('renders edit form with correct title when editing', () => {
    vi.mocked(useGetBathhousesId).mockReturnValue({
      data: {
        data: {
          id: 'abc',
          name: 'Test Bathhouse',
          address: 'Test addr',
          city_id: 1,
          price_per_hour: 150000,
          min_duration: 2,
          max_guests: 8,
          has_sauna: true,
          working_hours: [
            { day_of_week: 0, open_time: '10:00', close_time: '22:00' },
          ],
        },
        success: true,
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesId>)

    renderWithProviders(<BathhouseForm />, '/bathhouses/abc/edit')

    expect(screen.getByText('Редактирование бани')).toBeInTheDocument()
    expect(screen.getByText('Сохранить')).toBeInTheDocument()
  })

  it('shows loading spinner when fetching bathhouse data', () => {
    vi.mocked(useGetBathhousesId).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetBathhousesId>)

    renderWithProviders(<BathhouseForm />, '/bathhouses/abc/edit')

    expect(screen.queryByText('Редактирование бани')).not.toBeInTheDocument()
  })
})
