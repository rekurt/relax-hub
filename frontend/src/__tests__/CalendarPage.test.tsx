import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import CalendarPage from '@/pages/calendar/CalendarPage'

vi.mock('@/api/generated/calendar/calendar', () => ({
  useGetMyBathhousesIdCalendarToken: vi.fn(),
  usePostMyBathhousesIdSlotBlocks: vi.fn(),
  useGetMyBathhousesIdExternalCalendars: vi.fn(),
  usePostMyBathhousesIdExternalCalendars: vi.fn(),
  usePostMyBathhousesIdExternalCalendarsSync: vi.fn(),
  useDeleteMyExternalCalendarsId: vi.fn(),
}))

vi.mock('@/api/generated/bookings/bookings', () => ({
  useGetBathhousesIdBookings: vi.fn(),
}))

vi.mock('@/stores/bathhouse', () => ({
  useBathhouseStore: vi.fn(),
}))

import {
  useGetMyBathhousesIdCalendarToken,
  usePostMyBathhousesIdSlotBlocks,
  useGetMyBathhousesIdExternalCalendars,
  usePostMyBathhousesIdExternalCalendars,
  usePostMyBathhousesIdExternalCalendarsSync,
  useDeleteMyExternalCalendarsId,
} from '@/api/generated/calendar/calendar'
import { useGetBathhousesIdBookings } from '@/api/generated/bookings/bookings'
import { useBathhouseStore } from '@/stores/bathhouse'

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

const mockMutation = { mutate: vi.fn(), isPending: false }

function mockBathhouseStore(id: string | null) {
  vi.mocked(useBathhouseStore).mockImplementation((selector) =>
    (selector as (state: { selectedBathhouseId: string | null }) => unknown)({
      selectedBathhouseId: id,
    }),
  )
}

function setupDefaultMocks() {
  vi.mocked(usePostMyBathhousesIdSlotBlocks).mockReturnValue(
    mockMutation as unknown as ReturnType<typeof usePostMyBathhousesIdSlotBlocks>,
  )
  vi.mocked(usePostMyBathhousesIdExternalCalendars).mockReturnValue(
    mockMutation as unknown as ReturnType<typeof usePostMyBathhousesIdExternalCalendars>,
  )
  vi.mocked(usePostMyBathhousesIdExternalCalendarsSync).mockReturnValue(
    mockMutation as unknown as ReturnType<typeof usePostMyBathhousesIdExternalCalendarsSync>,
  )
  vi.mocked(useDeleteMyExternalCalendarsId).mockReturnValue(
    mockMutation as unknown as ReturnType<typeof useDeleteMyExternalCalendarsId>,
  )
}

describe('CalendarPage', () => {
  beforeEach(() => {
    setupDefaultMocks()
  })

  it('shows prompt when no bathhouse selected', () => {
    mockBathhouseStore(null)
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)
    vi.mocked(useGetMyBathhousesIdCalendarToken).mockReturnValue({
      data: undefined,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdCalendarToken>)
    vi.mocked(useGetMyBathhousesIdExternalCalendars).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdExternalCalendars>)

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Календарь')).toBeInTheDocument()
    expect(screen.getByText('Выберите баню')).toBeInTheDocument()
  })

  it('renders calendar with week navigation', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0, page: 0, page_size: 100, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)
    vi.mocked(useGetMyBathhousesIdCalendarToken).mockReturnValue({
      data: { data: { token: 'abc123', url: '/calendar/abc123.ics' } },
    } as unknown as ReturnType<typeof useGetMyBathhousesIdCalendarToken>)
    vi.mocked(useGetMyBathhousesIdExternalCalendars).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdExternalCalendars>)

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Календарь')).toBeInTheDocument()
    expect(screen.getByText('Сегодня')).toBeInTheDocument()
    expect(screen.getByText('Время')).toBeInTheDocument()
  })

  it('shows loading state for bookings', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)
    vi.mocked(useGetMyBathhousesIdCalendarToken).mockReturnValue({
      data: undefined,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdCalendarToken>)
    vi.mocked(useGetMyBathhousesIdExternalCalendars).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdExternalCalendars>)

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Календарь')).toBeInTheDocument()
    expect(document.querySelector('.ant-spin')).toBeInTheDocument()
  })

  it('shows slot block creation button', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0, page: 0, page_size: 100, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)
    vi.mocked(useGetMyBathhousesIdCalendarToken).mockReturnValue({
      data: { data: { token: 'abc123', url: '/calendar/abc123.ics' } },
    } as unknown as ReturnType<typeof useGetMyBathhousesIdCalendarToken>)
    vi.mocked(useGetMyBathhousesIdExternalCalendars).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdExternalCalendars>)

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Блокировка')).toBeInTheDocument()
  })

  it('opens slot block modal on button click', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0, page: 0, page_size: 100, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)
    vi.mocked(useGetMyBathhousesIdCalendarToken).mockReturnValue({
      data: { data: { token: 'abc123', url: '/calendar/abc123.ics' } },
    } as unknown as ReturnType<typeof useGetMyBathhousesIdCalendarToken>)
    vi.mocked(useGetMyBathhousesIdExternalCalendars).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdExternalCalendars>)

    renderWithProviders(<CalendarPage />)

    fireEvent.click(screen.getByText('Блокировка'))
    expect(screen.getByText('Создать блокировку')).toBeInTheDocument()
    expect(screen.getByText('Период блокировки')).toBeInTheDocument()
  })

  it('shows iCal export section with URL', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0, page: 0, page_size: 100, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)
    vi.mocked(useGetMyBathhousesIdCalendarToken).mockReturnValue({
      data: { data: { token: 'abc123', url: '/calendar/abc123.ics' } },
    } as unknown as ReturnType<typeof useGetMyBathhousesIdCalendarToken>)
    vi.mocked(useGetMyBathhousesIdExternalCalendars).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdExternalCalendars>)

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Экспорт iCal')).toBeInTheDocument()
    const input = screen.getByDisplayValue(/calendar\/abc123\.ics/)
    expect(input).toBeInTheDocument()
  })

  it('shows external calendars section with empty state', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0, page: 0, page_size: 100, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)
    vi.mocked(useGetMyBathhousesIdCalendarToken).mockReturnValue({
      data: { data: { token: 'abc123', url: '/calendar/abc123.ics' } },
    } as unknown as ReturnType<typeof useGetMyBathhousesIdCalendarToken>)
    vi.mocked(useGetMyBathhousesIdExternalCalendars).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdExternalCalendars>)

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Внешние календари')).toBeInTheDocument()
    expect(screen.getByText(/Нет подключённых календарей/)).toBeInTheDocument()
  })

  it('renders external calendars list when calendars exist', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0, page: 0, page_size: 100, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)
    vi.mocked(useGetMyBathhousesIdCalendarToken).mockReturnValue({
      data: { data: { token: 'abc123', url: '/calendar/abc123.ics' } },
    } as unknown as ReturnType<typeof useGetMyBathhousesIdCalendarToken>)
    vi.mocked(useGetMyBathhousesIdExternalCalendars).mockReturnValue({
      data: {
        data: [
          {
            id: 'cal-1',
            bathhouse_id: 'bathhouse-1',
            url: 'https://calendar.google.com/calendar/ical/test/basic.ics',
            source: 'google_calendar',
            last_sync_at: '2026-03-15T10:00:00Z',
            last_error: '',
            created_at: '2026-03-10T08:00:00Z',
          },
        ],
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdExternalCalendars>)

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Google')).toBeInTheDocument()
    expect(screen.getByText(/calendar\.google\.com/)).toBeInTheDocument()
  })

  it('opens external calendar modal on button click', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0, page: 0, page_size: 100, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)
    vi.mocked(useGetMyBathhousesIdCalendarToken).mockReturnValue({
      data: { data: { token: 'abc123', url: '/calendar/abc123.ics' } },
    } as unknown as ReturnType<typeof useGetMyBathhousesIdCalendarToken>)
    vi.mocked(useGetMyBathhousesIdExternalCalendars).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdExternalCalendars>)

    renderWithProviders(<CalendarPage />)

    fireEvent.click(screen.getByText('Внешний календарь'))
    expect(screen.getByText('Добавить внешний календарь')).toBeInTheDocument()
  })

  it('shows sync button for external calendars', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0, page: 0, page_size: 100, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)
    vi.mocked(useGetMyBathhousesIdCalendarToken).mockReturnValue({
      data: { data: { token: 'abc123', url: '/calendar/abc123.ics' } },
    } as unknown as ReturnType<typeof useGetMyBathhousesIdCalendarToken>)
    vi.mocked(useGetMyBathhousesIdExternalCalendars).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdExternalCalendars>)

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Синхронизировать')).toBeInTheDocument()
  })

  it('renders day headers in calendar', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0, page: 0, page_size: 100, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)
    vi.mocked(useGetMyBathhousesIdCalendarToken).mockReturnValue({
      data: { data: { token: 'abc123', url: '/calendar/abc123.ics' } },
    } as unknown as ReturnType<typeof useGetMyBathhousesIdCalendarToken>)
    vi.mocked(useGetMyBathhousesIdExternalCalendars).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdExternalCalendars>)

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Пн')).toBeInTheDocument()
    expect(screen.getByText('Вт')).toBeInTheDocument()
    expect(screen.getByText('Ср')).toBeInTheDocument()
    expect(screen.getByText('Чт')).toBeInTheDocument()
    expect(screen.getByText('Пт')).toBeInTheDocument()
    expect(screen.getByText('Сб')).toBeInTheDocument()
    expect(screen.getByText('Вс')).toBeInTheDocument()
  })

  it('renders legend tags', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0, page: 0, page_size: 100, total_pages: 0 } },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)
    vi.mocked(useGetMyBathhousesIdCalendarToken).mockReturnValue({
      data: { data: { token: 'abc123', url: '/calendar/abc123.ics' } },
    } as unknown as ReturnType<typeof useGetMyBathhousesIdCalendarToken>)
    vi.mocked(useGetMyBathhousesIdExternalCalendars).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdExternalCalendars>)

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Подтверждено')).toBeInTheDocument()
    expect(screen.getByText('Ожидает')).toBeInTheDocument()
    expect(screen.getByText('Завершено')).toBeInTheDocument()
  })
})
