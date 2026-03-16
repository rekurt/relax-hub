import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import PhotoManager from '@/pages/photos/PhotoManager'

vi.mock('@/api/generated/photos/photos', () => ({
  useGetMyBathhousesIdPhotos: vi.fn(),
  usePostMyBathhousesIdPhotos: vi.fn(),
  usePutMyBathhousesIdPhotosReorder: vi.fn(),
  useDeletePhotosId: vi.fn(),
}))

vi.mock('@/stores/bathhouse', () => ({
  useBathhouseStore: vi.fn(),
}))

import {
  useGetMyBathhousesIdPhotos,
  usePostMyBathhousesIdPhotos,
  usePutMyBathhousesIdPhotosReorder,
  useDeletePhotosId,
} from '@/api/generated/photos/photos'
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

const mockPhotos = [
  {
    id: 'photo-1',
    bathhouse_id: 'bath-1',
    url: 'https://example.com/photo1.jpg',
    thumbnail_url: 'https://example.com/photo1-thumb.jpg',
    status: 'verified',
    position: 0,
    uploaded_at: '2026-03-01T10:00:00Z',
  },
  {
    id: 'photo-2',
    bathhouse_id: 'bath-1',
    url: 'https://example.com/photo2.jpg',
    thumbnail_url: 'https://example.com/photo2-thumb.jpg',
    status: 'pending',
    position: 1,
    uploaded_at: '2026-03-05T12:00:00Z',
  },
  {
    id: 'photo-3',
    bathhouse_id: 'bath-1',
    url: 'https://example.com/photo3.jpg',
    thumbnail_url: null,
    status: 'rejected',
    position: 2,
    uploaded_at: '2026-03-10T08:00:00Z',
    rejection_reason: 'Низкое качество изображения',
  },
]

const mockUploadMutation = { mutate: vi.fn(), isPending: false }
const mockReorderMutation = { mutate: vi.fn(), isPending: false }
const mockDeleteMutation = { mutate: vi.fn(), isPending: false }

function mockBathhouseStore(id: string | null) {
  vi.mocked(useBathhouseStore).mockImplementation((selector) =>
    (selector as (state: { selectedBathhouseId: string | null }) => unknown)({
      selectedBathhouseId: id,
    }),
  )
}

function setupDefaultMocks() {
  vi.mocked(usePostMyBathhousesIdPhotos).mockReturnValue(
    mockUploadMutation as unknown as ReturnType<typeof usePostMyBathhousesIdPhotos>,
  )
  vi.mocked(usePutMyBathhousesIdPhotosReorder).mockReturnValue(
    mockReorderMutation as unknown as ReturnType<typeof usePutMyBathhousesIdPhotosReorder>,
  )
  vi.mocked(useDeletePhotosId).mockReturnValue(
    mockDeleteMutation as unknown as ReturnType<typeof useDeletePhotosId>,
  )
}

describe('PhotoManager', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setupDefaultMocks()
  })

  it('shows prompt when no bathhouse selected', () => {
    mockBathhouseStore(null)
    vi.mocked(useGetMyBathhousesIdPhotos).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPhotos>)

    renderWithProviders(<PhotoManager />)

    expect(screen.getByText('Фотографии')).toBeInTheDocument()
    expect(screen.getByText('Выберите баню для управления фотографиями')).toBeInTheDocument()
  })

  it('shows loading spinner while fetching', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPhotos).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPhotos>)

    renderWithProviders(<PhotoManager />)

    expect(document.querySelector('.ant-spin')).toBeInTheDocument()
  })

  it('shows empty state when no photos', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPhotos).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPhotos>)

    renderWithProviders(<PhotoManager />)

    expect(screen.getByText('Нет фотографий')).toBeInTheDocument()
    expect(screen.getByText('Загрузить первое фото')).toBeInTheDocument()
  })

  it('renders photo cards with data', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPhotos).mockReturnValue({
      data: { data: mockPhotos, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPhotos>)

    renderWithProviders(<PhotoManager />)

    expect(screen.getByText('Подтверждено')).toBeInTheDocument()
    expect(screen.getByText('На модерации')).toBeInTheDocument()
    expect(screen.getByText('Отклонено')).toBeInTheDocument()
  })

  it('shows rejection reason for rejected photos', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPhotos).mockReturnValue({
      data: { data: mockPhotos, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPhotos>)

    renderWithProviders(<PhotoManager />)

    expect(screen.getByText(/Низкое качество изображения/)).toBeInTheDocument()
  })

  it('shows upload dates', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPhotos).mockReturnValue({
      data: { data: mockPhotos, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPhotos>)

    renderWithProviders(<PhotoManager />)

    expect(screen.getByText('01.03.2026')).toBeInTheDocument()
    expect(screen.getByText('05.03.2026')).toBeInTheDocument()
    expect(screen.getByText('10.03.2026')).toBeInTheDocument()
  })

  it('opens upload modal when add button clicked', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPhotos).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPhotos>)

    renderWithProviders(<PhotoManager />)

    fireEvent.click(screen.getByText('Добавить фото'))

    expect(screen.getByText('URL фотографии')).toBeInTheDocument()
    expect(screen.getByText('URL превью (необязательно)')).toBeInTheDocument()
  })

  it('opens upload modal from empty state button', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPhotos).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPhotos>)

    renderWithProviders(<PhotoManager />)

    fireEvent.click(screen.getByText('Загрузить первое фото'))

    expect(screen.getByText('Добавить фото', { selector: '.ant-modal-title' })).toBeInTheDocument()
  })

  it('calls delete mutation when confirmed', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPhotos).mockReturnValue({
      data: { data: [mockPhotos[0]], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPhotos>)

    renderWithProviders(<PhotoManager />)

    const deleteIcons = document.querySelectorAll('.anticon-delete')
    expect(deleteIcons.length).toBeGreaterThan(0)
    fireEvent.click(deleteIcons[0]!)

    expect(screen.getByText('Удалить фото?')).toBeInTheDocument()

    const confirmBtn = screen.getByRole('button', { name: 'Удалить' })
    fireEvent.click(confirmBtn)

    expect(mockDeleteMutation.mutate).toHaveBeenCalledWith({ id: 'photo-1' })
  })

  it('shows drag icons for reordering', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPhotos).mockReturnValue({
      data: { data: mockPhotos, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPhotos>)

    renderWithProviders(<PhotoManager />)

    const dragIcons = document.querySelectorAll('.anticon-drag')
    expect(dragIcons.length).toBe(3)
  })

  it('renders photo images with thumbnails', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPhotos).mockReturnValue({
      data: { data: [mockPhotos[0]], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPhotos>)

    renderWithProviders(<PhotoManager />)

    const images = document.querySelectorAll('img')
    const photoImg = Array.from(images).find(
      (img) => img.src === 'https://example.com/photo1-thumb.jpg',
    )
    expect(photoImg).toBeTruthy()
  })

  it('shows add and refresh buttons', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPhotos).mockReturnValue({
      data: { data: [], success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPhotos>)

    renderWithProviders(<PhotoManager />)

    expect(screen.getByText('Добавить фото')).toBeInTheDocument()
    expect(screen.getByText('Обновить')).toBeInTheDocument()
  })

  it('has draggable cards', () => {
    mockBathhouseStore('bath-1')
    vi.mocked(useGetMyBathhousesIdPhotos).mockReturnValue({
      data: { data: mockPhotos, success: true },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyBathhousesIdPhotos>)

    renderWithProviders(<PhotoManager />)

    const cards = document.querySelectorAll('.ant-card[draggable="true"]')
    expect(cards.length).toBe(3)
  })
})
