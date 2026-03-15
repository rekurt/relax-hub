import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ComplaintManagement from '@/pages/admin/ComplaintManagement'

vi.mock('@/api/generated/admin-complaints/admin-complaints', () => ({
  useGetAdminComplaints: vi.fn(),
  getGetAdminComplaintsQueryKey: vi.fn(() => ['/admin/complaints']),
  usePatchAdminComplaintsIdResolve: vi.fn(),
  usePatchAdminComplaintsIdDismiss: vi.fn(),
}))

import {
  useGetAdminComplaints,
  usePatchAdminComplaintsIdResolve,
  usePatchAdminComplaintsIdDismiss,
} from '@/api/generated/admin-complaints/admin-complaints'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/complaints']}>
            <Routes>
              <Route path="/admin/complaints" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockComplaints = [
  {
    id: 'c-1',
    reporter_id: 'u-100',
    target_id: 'r-200',
    target_type: 'review',
    reason: 'spam',
    description: 'Спам в отзыве',
    status: 'pending',
    created_at: '2026-03-15T10:00:00Z',
  },
  {
    id: 'c-2',
    reporter_id: 'u-101',
    target_id: 'b-300',
    target_type: 'bathhouse',
    reason: 'fraud',
    description: 'Мошенническое объявление',
    status: 'resolved',
    resolution: 'Объявление заблокировано',
    resolved_at: '2026-03-15T12:00:00Z',
    resolved_by_id: 'admin-1',
    created_at: '2026-03-14T08:00:00Z',
  },
  {
    id: 'c-3',
    reporter_id: 'u-102',
    target_id: 'u-400',
    target_type: 'user',
    reason: 'offensive',
    description: 'Оскорбительное поведение',
    status: 'dismissed',
    created_at: '2026-03-13T16:00:00Z',
  },
]

const mutationDefault = { mutateAsync: vi.fn(), isPending: false }

beforeEach(() => {
  vi.mocked(useGetAdminComplaints).mockReturnValue({
    data: {
      data: mockComplaints,
      success: true,
      meta: { page: 1, page_size: 20, total_count: 3, total_pages: 1 },
    },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetAdminComplaints>)

  vi.mocked(usePatchAdminComplaintsIdResolve).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePatchAdminComplaintsIdResolve>,
  )
  vi.mocked(usePatchAdminComplaintsIdDismiss).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePatchAdminComplaintsIdDismiss>,
  )
})

describe('ComplaintManagement', () => {
  it('renders page title and complaint list', () => {
    renderWithProviders(<ComplaintManagement />)

    expect(screen.getByText('Управление жалобами')).toBeInTheDocument()
    expect(screen.getByText('Спам в отзыве')).toBeInTheDocument()
    expect(screen.getByText('Мошенническое объявление')).toBeInTheDocument()
    expect(screen.getByText('Оскорбительное поведение')).toBeInTheDocument()
  })

  it('renders status filter options', () => {
    renderWithProviders(<ComplaintManagement />)

    expect(screen.getByText('Все')).toBeInTheDocument()
    // "На рассмотрении" appears both in segmented control and status tag
    expect(screen.getAllByText('На рассмотрении').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Решённые')).toBeInTheDocument()
    expect(screen.getByText('Отклонённые')).toBeInTheDocument()
  })

  it('renders target type filter options', () => {
    renderWithProviders(<ComplaintManagement />)

    expect(screen.getByText('Все типы')).toBeInTheDocument()
    // "Отзыв" also appears in the table
    expect(screen.getAllByText('Отзыв').length).toBeGreaterThanOrEqual(1)
  })

  it('renders reason filter options', () => {
    renderWithProviders(<ComplaintManagement />)

    expect(screen.getByText('Все причины')).toBeInTheDocument()
    // "Спам" appears in filter and in reason tag column
    expect(screen.getAllByText('Спам').length).toBeGreaterThanOrEqual(1)
  })

  it('renders status tags for complaints', () => {
    renderWithProviders(<ComplaintManagement />)

    expect(screen.getAllByText('На рассмотрении').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Решена')).toBeInTheDocument()
    expect(screen.getByText('Отклонена')).toBeInTheDocument()
  })

  it('renders reason tags in table', () => {
    renderWithProviders(<ComplaintManagement />)

    expect(screen.getAllByText('Спам').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Мошенничество').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Оскорбление').length).toBeGreaterThanOrEqual(1)
  })

  it('shows resolve/dismiss actions only for pending complaints', () => {
    renderWithProviders(<ComplaintManagement />)

    // Only c-1 is pending, so only 1 "Решить" and 1 "Отклонить" in the action column
    const resolveButtons = screen.getAllByText('Решить')
    expect(resolveButtons.length).toBe(1)
    const dismissButtons = screen.getAllByText('Отклонить')
    // "Отклонить" appears once in action column + once in Segmented filter as "Отклонённые"
    expect(dismissButtons.length).toBe(1)
  })

  it('opens detail drawer when clicking description', async () => {
    renderWithProviders(<ComplaintManagement />)

    fireEvent.click(screen.getByText('Спам в отзыве'))

    await waitFor(() => {
      expect(screen.getByText('Детали жалобы')).toBeInTheDocument()
    })
  })

  it('shows resolve and dismiss buttons in drawer for pending complaint', async () => {
    renderWithProviders(<ComplaintManagement />)

    fireEvent.click(screen.getByText('Спам в отзыве')) // c-1 is pending

    await waitFor(() => {
      expect(screen.getByText('Детали жалобы')).toBeInTheDocument()
    })

    // Drawer extra buttons
    const resolveButtons = screen.getAllByText('Решить')
    expect(resolveButtons.length).toBeGreaterThanOrEqual(2) // table + drawer
  })

  it('opens resolve modal with note input', async () => {
    renderWithProviders(<ComplaintManagement />)

    const resolveButtons = screen.getAllByText('Решить')
    fireEvent.click(resolveButtons[0])

    await waitFor(() => {
      expect(screen.getByText('Решить жалобу')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Примечание к решению')).toBeInTheDocument()
    })
  })

  it('renders empty state when no complaints', () => {
    vi.mocked(useGetAdminComplaints).mockReturnValue({
      data: {
        data: [],
        success: true,
        meta: { page: 1, page_size: 20, total_count: 0, total_pages: 0 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminComplaints>)

    renderWithProviders(<ComplaintManagement />)

    expect(screen.getByText('Нет жалоб')).toBeInTheDocument()
  })

  it('renders loading state', () => {
    vi.mocked(useGetAdminComplaints).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetAdminComplaints>)

    renderWithProviders(<ComplaintManagement />)

    expect(screen.getByText('Управление жалобами')).toBeInTheDocument()
    // Spin is shown
    expect(document.querySelector('.ant-spin')).toBeTruthy()
  })

  it('shows resolution info in drawer for resolved complaints', async () => {
    renderWithProviders(<ComplaintManagement />)

    fireEvent.click(screen.getByText('Мошенническое объявление')) // c-2 is resolved

    await waitFor(() => {
      expect(screen.getByText('Детали жалобы')).toBeInTheDocument()
      expect(screen.getByText('Объявление заблокировано')).toBeInTheDocument()
    })
  })
})
