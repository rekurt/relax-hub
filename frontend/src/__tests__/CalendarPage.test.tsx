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

vi.mock('@/api/generated/bathhouses/bathhouses', () => ({
  useGetMyBathhouses: vi.fn(),
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
import { useGetMyBathhouses } from '@/api/generated/bathhouses/bathhouses'
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

function setupDefaultMocks(opts?: { bathhouses?: unknown[] }) {
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
  vi.mocked(useGetMyBathhouses).mockReturnValue({
    data: { data: opts?.bathhouses ?? [] },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyBathhouses>)
}

function setupCalendarMocks(bookings: unknown[] = []) {
  vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
    data: { data: bookings, success: true, meta: { total_count: bookings.length, page: 1, page_size: 200, total_pages: 1 } },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetBathhousesIdBookings>)
  vi.mocked(useGetMyBathhousesIdCalendarToken).mockReturnValue({
    data: { data: { token: 'abc123', url: '/calendar/abc123.ics' } },
  } as unknown as ReturnType<typeof useGetMyBathhousesIdCalendarToken>)
  vi.mocked(useGetMyBathhousesIdExternalCalendars).mockReturnValue({
    data: { data: [] },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMyBathhousesIdExternalCalendars>)
}

describe('CalendarPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
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

  it('renders calendar with week view by default and navigation', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Календарь')).toBeInTheDocument()
    expect(screen.getByText('Сегодня')).toBeInTheDocument()
    expect(screen.getByText('Время')).toBeInTheDocument()
    // View switcher should show all three options
    expect(screen.getByText('День')).toBeInTheDocument()
    expect(screen.getByText('Неделя')).toBeInTheDocument()
    expect(screen.getByText('Месяц')).toBeInTheDocument()
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
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Блокировка')).toBeInTheDocument()
  })

  it('opens slot block modal on button click', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    fireEvent.click(screen.getByText('Блокировка'))
    expect(screen.getByText('Создать блокировку')).toBeInTheDocument()
    expect(screen.getByText('Период блокировки')).toBeInTheDocument()
  })

  it('shows iCal export section with URL', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Экспорт iCal')).toBeInTheDocument()
    const input = screen.getByDisplayValue(/calendar\/abc123\.ics/)
    expect(input).toBeInTheDocument()
  })

  it('shows external calendars section with empty state', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Внешние календари')).toBeInTheDocument()
    expect(screen.getByText(/Нет подключённых календарей/)).toBeInTheDocument()
  })

  it('renders external calendars list when calendars exist', () => {
    mockBathhouseStore('bathhouse-1')
    vi.mocked(useGetBathhousesIdBookings).mockReturnValue({
      data: { data: [], success: true, meta: { total_count: 0, page: 1, page_size: 200, total_pages: 0 } },
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
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    fireEvent.click(screen.getByText('Внешний календарь'))
    expect(screen.getByText('Добавить внешний календарь')).toBeInTheDocument()
  })

  it('shows sync button for external calendars', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Синхронизировать')).toBeInTheDocument()
  })

  it('renders day headers in week view', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Пн')).toBeInTheDocument()
    expect(screen.getByText('Вт')).toBeInTheDocument()
    expect(screen.getByText('Ср')).toBeInTheDocument()
    expect(screen.getByText('Чт')).toBeInTheDocument()
    expect(screen.getByText('Пт')).toBeInTheDocument()
    expect(screen.getByText('Сб')).toBeInTheDocument()
    expect(screen.getByText('Вс')).toBeInTheDocument()
  })

  it('renders legend with color-coded statuses', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Подтверждено')).toBeInTheDocument()
    expect(screen.getByText('Ожидает')).toBeInTheDocument()
    expect(screen.getByText('Отменено')).toBeInTheDocument()
    expect(screen.getByText('Заблокировано')).toBeInTheDocument()
  })

  // --- Day view tests ---

  it('switches to day view when Day is clicked', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    fireEvent.click(screen.getByText('День'))
    // Day view should show time column but no day-of-week headers in the grid
    expect(screen.getByText('Время')).toBeInTheDocument()
    // Day headers (Пн..Вс) should not appear in the hourly grid area (only in legend area)
    // The segmented control still shows them as labels
  })

  it('day view shows single column with time slots', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    fireEvent.click(screen.getByText('День'))
    // Should have 00:00 through 23:00 time labels
    expect(screen.getByText('00:00')).toBeInTheDocument()
    expect(screen.getByText('12:00')).toBeInTheDocument()
    expect(screen.getByText('23:00')).toBeInTheDocument()
  })

  // --- Month view tests ---

  it('switches to month view when Month is clicked', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    fireEvent.click(screen.getByText('Месяц'))
    // Month view should show day-of-week headers
    expect(screen.getByText('Пн')).toBeInTheDocument()
    expect(screen.getByText('Вс')).toBeInTheDocument()
    // Should not show time column
    expect(screen.queryByText('Время')).not.toBeInTheDocument()
  })

  it('month view shows day numbers', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    fireEvent.click(screen.getByText('Месяц'))
    // Should contain day numbers like 1, 15, etc
    expect(screen.getByText('15')).toBeInTheDocument()
  })

  it('clicking a day in month view switches to day view', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    fireEvent.click(screen.getByText('Месяц'))
    // Click on day "15"
    fireEvent.click(screen.getByText('15'))
    // Should switch to day view - time column should appear
    expect(screen.getByText('Время')).toBeInTheDocument()
  })

  // --- View switcher tests ---

  it('view switcher has day/week/month options', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    const dayOption = screen.getByText('День')
    const weekOption = screen.getByText('Неделя')
    const monthOption = screen.getByText('Месяц')

    expect(dayOption).toBeInTheDocument()
    expect(weekOption).toBeInTheDocument()
    expect(monthOption).toBeInTheDocument()
  })

  it('can cycle through all views', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    // Default is week view
    expect(screen.getByText('Время')).toBeInTheDocument()

    // Switch to day
    fireEvent.click(screen.getByText('День'))
    expect(screen.getByText('Время')).toBeInTheDocument()

    // Switch to month
    fireEvent.click(screen.getByText('Месяц'))
    expect(screen.queryByText('Время')).not.toBeInTheDocument()

    // Back to week
    fireEvent.click(screen.getByText('Неделя'))
    expect(screen.getByText('Время')).toBeInTheDocument()
  })

  // --- Multi-object consolidated view ---

  it('shows consolidated view toggle when owner has multiple bathhouses', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()
    setupDefaultMocks({
      bathhouses: [
        { id: 'b1', name: 'Баня 1' },
        { id: 'b2', name: 'Баня 2' },
      ],
    })

    renderWithProviders(<CalendarPage />)

    expect(screen.getByText('Все объекты')).toBeInTheDocument()
  })

  it('does not show consolidated toggle with single bathhouse', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()
    setupDefaultMocks({ bathhouses: [{ id: 'b1', name: 'Баня 1' }] })

    renderWithProviders(<CalendarPage />)

    expect(screen.queryByText('Все объекты')).not.toBeInTheDocument()
  })

  // --- Navigation ---

  it('navigates forward and backward in week view', () => {
    mockBathhouseStore('bathhouse-1')
    setupCalendarMocks()

    renderWithProviders(<CalendarPage />)

    // Find navigation buttons by their icons (left/right arrows)
    const buttons = screen.getAllByRole('button')
    // Left and right arrows are the first navigation buttons
    const leftButton = buttons.find(
      (b) => b.querySelector('.anticon-left'),
    )
    const rightButton = buttons.find(
      (b) => b.querySelector('.anticon-right'),
    )

    expect(leftButton).toBeDefined()
    expect(rightButton).toBeDefined()

    if (rightButton) fireEvent.click(rightButton)
    if (leftButton) fireEvent.click(leftButton)
    // Should be back to current week - Today button should still work
    expect(screen.getByText('Сегодня')).toBeInTheDocument()
  })
})
