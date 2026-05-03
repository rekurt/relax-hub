import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { beforeEach, describe, it, expect, vi } from 'vitest'

vi.mock('@/api/generated/share/share', () => ({
  useGetApiV1ShareBookingToken: vi.fn(),
}))

import { useGetApiV1ShareBookingToken } from '@/api/generated/share/share'

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ token: 'test-token-123' }),
  }
})

import ShareRedirect from '@/pages/ShareRedirect'

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

describe('ShareRedirect', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading spinner while resolving', () => {
    vi.mocked(useGetApiV1ShareBookingToken).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
    } as unknown as ReturnType<typeof useGetApiV1ShareBookingToken>)

    const { container } = renderWithProviders(<ShareRedirect />)
    expect(container.querySelector('.ant-spin')).toBeTruthy()
  })

  it('shows error state on failed resolution', () => {
    vi.mocked(useGetApiV1ShareBookingToken).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    } as unknown as ReturnType<typeof useGetApiV1ShareBookingToken>)

    renderWithProviders(<ShareRedirect />)
    expect(screen.getByText('Ссылка недействительна')).toBeInTheDocument()
    expect(screen.getByText('На главную')).toBeInTheDocument()
  })

  it('redirects to booking detail when booking_id is resolved', () => {
    vi.mocked(useGetApiV1ShareBookingToken).mockReturnValue({
      data: { data: { booking_id: 'booking-123', bathhouse_id: 'bath-1' } },
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof useGetApiV1ShareBookingToken>)

    renderWithProviders(<ShareRedirect />)
    expect(mockNavigate).toHaveBeenCalledWith('/client/bookings/booking-123', { replace: true })
  })

  it('redirects to booking creation when only bathhouse_id is resolved', () => {
    vi.mocked(useGetApiV1ShareBookingToken).mockReturnValue({
      data: {
        data: {
          bathhouse_id: 'bath-1',
          start_time: '2026-04-01T10:00:00Z',
          guest_count: 4,
        },
      },
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof useGetApiV1ShareBookingToken>)

    renderWithProviders(<ShareRedirect />)
    expect(mockNavigate).toHaveBeenCalledWith(
      '/checkout?bathhouse=bath-1&from=2026-04-01T10%3A00%3A00Z&date=2026-04-01&guests=4',
      { replace: true },
    )
  })
})
