import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import PhotoVerification from '@/pages/admin/PhotoVerification'

vi.mock('@/api/generated/admin-photos/admin-photos', () => ({
  useGetAdminPhotosPending: vi.fn(),
  getGetAdminPhotosPendingQueryKey: vi.fn(() => ['/admin/photos/pending']),
  usePatchAdminPhotosIdVerify: vi.fn(),
  usePatchAdminPhotosIdReject: vi.fn(),
}))

import {
  useGetAdminPhotosPending,
  usePatchAdminPhotosIdVerify,
  usePatchAdminPhotosIdReject,
} from '@/api/generated/admin-photos/admin-photos'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter initialEntries={['/admin/photos']}>
            <Routes>
              <Route path="/admin/photos" element={ui} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockPhotos = [
  {
    id: 'p-1',
    bathhouse_id: 'b-aaaa-bbbb-cccc-dddd',
    url: 'https://example.com/photo1.jpg',
    thumbnail_url: 'https://example.com/thumb1.jpg',
    status: 'pending',
    position: 0,
    uploaded_at: '2026-03-15T10:00:00Z',
  },
  {
    id: 'p-2',
    bathhouse_id: 'b-eeee-ffff-gggg-hhhh',
    url: 'https://example.com/photo2.jpg',
    thumbnail_url: 'https://example.com/thumb2.jpg',
    status: 'pending',
    position: 2,
    uploaded_at: '2026-03-14T14:00:00Z',
  },
  {
    id: 'p-3',
    bathhouse_id: 'b-aaaa-bbbb-cccc-dddd',
    url: 'https://example.com/photo3.jpg',
    thumbnail_url: 'https://example.com/thumb3.jpg',
    status: 'pending',
    position: 1,
    uploaded_at: '2026-03-13T09:00:00Z',
  },
]

const mutationDefault = { mutateAsync: vi.fn(), isPending: false }

beforeEach(() => {
  vi.mocked(useGetAdminPhotosPending).mockReturnValue({
    data: {
      data: mockPhotos,
      success: true,
      meta: { page: 1, page_size: 12, total_count: 3, total_pages: 1 },
    },
    isLoading: false,
  } as unknown as ReturnType<typeof useGetAdminPhotosPending>)

  vi.mocked(usePatchAdminPhotosIdVerify).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePatchAdminPhotosIdVerify>,
  )
  vi.mocked(usePatchAdminPhotosIdReject).mockReturnValue(
    mutationDefault as unknown as ReturnType<typeof usePatchAdminPhotosIdReject>,
  )
})

describe('PhotoVerification', () => {
  it('renders page title and pending count', () => {
    renderWithProviders(<PhotoVerification />)

    expect(screen.getByText('Верификация фото')).toBeInTheDocument()
    expect(screen.getByText(/Ожидают проверки: 3/)).toBeInTheDocument()
  })

  it('renders photo cards with bathhouse info', () => {
    renderWithProviders(<PhotoVerification />)

    // Each photo card shows bathhouse ID prefix
    const bathhouseTexts = screen.getAllByText(/Баня:/)
    expect(bathhouseTexts).toHaveLength(3)
  })

  it('renders pending status tags', () => {
    renderWithProviders(<PhotoVerification />)

    const pendingTags = screen.getAllByText('Ожидает')
    expect(pendingTags).toHaveLength(3)
  })

  it('renders verify and reject buttons for each photo', () => {
    renderWithProviders(<PhotoVerification />)

    const verifyButtons = screen.getAllByText('Подтвердить')
    const rejectButtons = screen.getAllByText('Отклонить')
    expect(verifyButtons).toHaveLength(3)
    expect(rejectButtons).toHaveLength(3)
  })

  it('shows confirm modal when clicking verify', async () => {
    renderWithProviders(<PhotoVerification />)

    const verifyButtons = screen.getAllByText('Подтвердить')
    fireEvent.click(verifyButtons[0]!)

    await waitFor(() => {
      const matches = screen.getAllByText('Подтвердить фото?')
      expect(matches.length).toBeGreaterThanOrEqual(1)
    })
  })

  it('opens reject modal when clicking reject', async () => {
    renderWithProviders(<PhotoVerification />)

    const rejectButtons = screen.getAllByText('Отклонить')
    fireEvent.click(rejectButtons[0]!)

    await waitFor(() => {
      expect(screen.getByText('Отклонить фото')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Причина отклонения')).toBeInTheDocument()
    })
  })

  it('renders empty state when no photos', () => {
    vi.mocked(useGetAdminPhotosPending).mockReturnValue({
      data: {
        data: [],
        success: true,
        meta: { page: 1, page_size: 12, total_count: 0, total_pages: 0 },
      },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetAdminPhotosPending>)

    renderWithProviders(<PhotoVerification />)

    expect(screen.getByText('Нет фото на рассмотрении')).toBeInTheDocument()
  })

  it('renders loading state', () => {
    vi.mocked(useGetAdminPhotosPending).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetAdminPhotosPending>)

    renderWithProviders(<PhotoVerification />)

    expect(screen.getByText('Верификация фото')).toBeInTheDocument()
    // Spin renders a dot animation
    expect(document.querySelector('.ant-spin')).toBeInTheDocument()
  })

  it('renders photo position numbers', () => {
    renderWithProviders(<PhotoVerification />)

    expect(screen.getByText('#1')).toBeInTheDocument()
    expect(screen.getByText('#3')).toBeInTheDocument()
    expect(screen.getByText('#2')).toBeInTheDocument()
  })

  it('renders upload dates', () => {
    renderWithProviders(<PhotoVerification />)

    // formatDateTime should produce date strings for all 3 photos
    const dateElements = screen.getAllByText(/2026/)
    expect(dateElements.length).toBeGreaterThanOrEqual(3)
  })
})
