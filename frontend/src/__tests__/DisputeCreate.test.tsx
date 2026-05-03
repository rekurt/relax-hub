import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import DisputeCreate from '@/pages/client/DisputeCreate'

vi.mock('@/api/generated/disputes/disputes', () => ({
  usePostBookingsIdDispute: vi.fn(),
  getGetMyDisputesQueryKey: vi.fn().mockReturnValue(['/my/disputes']),
}))

import { usePostBookingsIdDispute } from '@/api/generated/disputes/disputes'

function renderWithProviders(route = '/client/disputes/create') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={[route]}>
            <DisputeCreate />
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

describe('DisputeCreate', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(usePostBookingsIdDispute).mockReturnValue({
      mutateAsync: vi.fn(),
      isError: false,
      isPending: false,
    } as never)
  })

  it('renders "booking not specified" when no booking param', () => {
    renderWithProviders('/client/disputes/create')

    expect(screen.getByText('Бронирование не указано')).toBeDefined()
    expect(screen.getByText('К бронированиям')).toBeDefined()
  })

  it('renders dispute form when booking param present', () => {
    renderWithProviders('/client/disputes/create?booking=b123')

    expect(screen.getAllByText('Открыть спор').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Причина спора')).toBeDefined()
    expect(screen.getByText('Описание')).toBeDefined()
  })

  it('displays booking ID in disabled input', () => {
    renderWithProviders('/client/disputes/create?booking=booking-xyz')

    const input = screen.getByDisplayValue('booking-xyz')
    expect(input).toBeDefined()
    expect((input as HTMLInputElement).disabled).toBe(true)
  })

  it('has submit button', () => {
    renderWithProviders('/client/disputes/create?booking=b123')

    expect(screen.getByRole('button', { name: /открыть спор/i })).toBeDefined()
  })
})
