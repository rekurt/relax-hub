import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ListingImport from '@/pages/bathhouses/ListingImport'

vi.mock('@/api/generated/listings/listings', () => ({
  usePostMyListingsImport: vi.fn(),
  useGetMyListingsImportTemplate: vi.fn(),
}))

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: { get: vi.fn() },
}))

import { usePostMyListingsImport } from '@/api/generated/listings/listings'

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

describe('ListingImport', () => {
  let mutateFn: ReturnType<typeof vi.fn>

  beforeEach(() => {
    mutateFn = vi.fn()
    vi.mocked(usePostMyListingsImport).mockReturnValue({
      mutate: mutateFn,
      isPending: false,
    } as unknown as ReturnType<typeof usePostMyListingsImport>)
  })

  it('renders import page with title and template download buttons', () => {
    renderWithProviders(<ListingImport />)

    expect(screen.getByText('Импорт объектов')).toBeInTheDocument()
    expect(screen.getByText('Скачать XLSX')).toBeInTheDocument()
    expect(screen.getByText('Скачать CSV')).toBeInTheDocument()
  })

  it('renders upload area', () => {
    renderWithProviders(<ListingImport />)

    expect(screen.getByText('Нажмите или перетащите файл для загрузки')).toBeInTheDocument()
    expect(screen.getByText('Поддерживаются форматы CSV и XLSX')).toBeInTheDocument()
  })

  it('shows back button', () => {
    renderWithProviders(<ListingImport />)

    expect(screen.getByText('Назад к списку')).toBeInTheDocument()
  })

  it('shows processing state when upload is in progress', () => {
    vi.mocked(usePostMyListingsImport).mockReturnValue({
      mutate: mutateFn,
      isPending: true,
    } as unknown as ReturnType<typeof usePostMyListingsImport>)

    renderWithProviders(<ListingImport />)

    expect(screen.getByText('Обработка файла...')).toBeInTheDocument()
  })

  it('calls import mutation when file is uploaded', async () => {
    renderWithProviders(<ListingImport />)

    const file = new File(['test'], 'test.csv', { type: 'text/csv' })
    const input = document.querySelector('input[type="file"]')!
    fireEvent.change(input, { target: { files: [file] } })

    await waitFor(() => {
      expect(mutateFn).toHaveBeenCalled()
    })
  })
})
