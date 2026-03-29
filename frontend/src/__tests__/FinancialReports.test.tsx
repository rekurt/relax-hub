import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi } from 'vitest'
import FinancialReports from '@/pages/finance/FinancialReports'

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetMyBathhouses: vi.fn(),
}))

import { useGetMyBathhouses } from '@/api/generated/bathhouses/bathhouses'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/finance/reports']}>
            <Routes>
              <Route path="/finance/reports" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockBathhouses = [
  { id: 'bath-1', name: 'Баня на Речной' },
  { id: 'bath-2', name: 'Парная Люкс' },
]

function setupMocks(overrides?: { bathhouses?: typeof mockBathhouses }) {
  vi.mocked(useGetMyBathhouses).mockReturnValue({
    data: {
      data: overrides?.bathhouses ?? mockBathhouses,
      success: true,
    },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyBathhouses>)
}

describe('FinancialReports', () => {
  it('renders page title', () => {
    setupMocks()
    renderWithProviders(<FinancialReports />)

    expect(screen.getByText('Отчёты')).toBeInTheDocument()
  })

  it('renders wallet history export card', () => {
    setupMocks()
    renderWithProviders(<FinancialReports />)

    expect(screen.getByText('История кошелька')).toBeInTheDocument()
    expect(screen.getByText(/Выгрузка всех операций/)).toBeInTheDocument()
  })

  it('renders act generation card', () => {
    setupMocks()
    renderWithProviders(<FinancialReports />)

    expect(screen.getByText('Акт оказанных услуг')).toBeInTheDocument()
    expect(screen.getByText('Скачать акт')).toBeInTheDocument()
  })

  it('renders 1C XML export card', () => {
    setupMocks()
    renderWithProviders(<FinancialReports />)

    expect(screen.getByText('Экспорт 1С')).toBeInTheDocument()
    expect(screen.getByText('Скачать XML')).toBeInTheDocument()
  })

  it('renders date range picker', () => {
    setupMocks()
    renderWithProviders(<FinancialReports />)

    expect(screen.getByPlaceholderText('С')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('По')).toBeInTheDocument()
  })

  it('disables act download when no bathhouse selected', () => {
    setupMocks()
    renderWithProviders(<FinancialReports />)

    const actButton = screen.getByText('Скачать акт').closest('button')
    expect(actButton).toBeDisabled()
  })

  it('renders bathhouse selector for acts', () => {
    setupMocks()
    renderWithProviders(<FinancialReports />)

    expect(screen.getByText('Выберите баню')).toBeInTheDocument()
  })
})
