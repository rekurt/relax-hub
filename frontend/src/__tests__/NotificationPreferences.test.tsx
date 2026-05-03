import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import NotificationPreferences from '@/pages/client/NotificationPreferences'

vi.mock('@/api/generated/notifications/notifications', () => ({
  useGetMyNotificationPreferences: vi.fn(),
  usePutMyNotificationPreferences: vi.fn(),
}))

vi.mock('@/api/axios-instance', () => ({
  axiosInstance: {
    get: vi.fn(),
    put: vi.fn(),
  },
}))

import {
  useGetMyNotificationPreferences,
  usePutMyNotificationPreferences,
} from '@/api/generated/notifications/notifications'
import { axiosInstance } from '@/api/axios-instance'

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

describe('NotificationPreferences', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(usePutMyNotificationPreferences).mockReturnValue({
      mutateAsync: vi.fn(),
    } as never)
  })

  it('renders loading spinner while data is loading', () => {
    vi.mocked(useGetMyNotificationPreferences).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as never)
    vi.mocked(axiosInstance.get).mockReturnValue(new Promise(() => {}))

    renderWithProviders(<NotificationPreferences />)

    expect(document.querySelector('.ant-spin')).not.toBeNull()
  })

  it('renders channel toggles when loaded', async () => {
    vi.mocked(useGetMyNotificationPreferences).mockReturnValue({
      data: {
        data: {
          push: true,
          email: true,
          sms: false,
          telegram: false,
          in_app: true,
        },
      },
      isLoading: false,
    } as never)

    vi.mocked(axiosInstance.get).mockResolvedValueOnce({
      data: {
        data: [
          { event_type: 'booking_confirmed', push_enabled: true, email_enabled: true, sms_enabled: false, is_mandatory: true },
          { event_type: 'promo', push_enabled: true, email_enabled: false, sms_enabled: false, is_mandatory: false },
        ],
      },
    })

    renderWithProviders(<NotificationPreferences />)

    await waitFor(() => {
      expect(screen.getByText('Каналы доставки')).toBeDefined()
    })
  })

  it('renders event preference categories', async () => {
    vi.mocked(useGetMyNotificationPreferences).mockReturnValue({
      data: {
        data: { push: true, email: true, sms: false, telegram: false, in_app: true },
      },
      isLoading: false,
    } as never)

    vi.mocked(axiosInstance.get).mockResolvedValueOnce({
      data: {
        data: [
          { event_type: 'booking_confirmed', push_enabled: true, email_enabled: true, sms_enabled: false, is_mandatory: true },
        ],
      },
    })

    renderWithProviders(<NotificationPreferences />)

    await waitFor(() => {
      expect(screen.getByText('Настройки по типам событий')).toBeDefined()
    })
  })
})
