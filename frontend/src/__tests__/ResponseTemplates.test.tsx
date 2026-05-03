import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ResponseTemplates from '@/pages/crm/ResponseTemplates'

vi.mock('@/api/generated/crm/crm', () => ({
  useGetMyCrmTemplates: vi.fn(),
  usePostMyCrmTemplates: vi.fn(),
  usePutMyCrmTemplatesId: vi.fn(),
  useDeleteMyCrmTemplatesId: vi.fn(),
}))

import {
  useGetMyCrmTemplates,
  usePostMyCrmTemplates,
  usePutMyCrmTemplatesId,
  useDeleteMyCrmTemplatesId,
} from '@/api/generated/crm/crm'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU} theme={{ token: { motion: false } }}>
        <AntApp>
          <MemoryRouter>{ui}</MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

const mockTemplates = [
  {
    id: 'tpl-1',
    owner_id: 'owner-1',
    title: 'Благодарность за отзыв',
    body: 'Спасибо за ваш отзыв! Мы ценим ваше мнение.',
    is_default: true,
    sort_order: 0,
    created_at: '2026-01-01T10:00:00Z',
  },
  {
    id: 'tpl-2',
    owner_id: 'owner-1',
    title: 'Мой кастомный шаблон',
    body: 'Приходите к нам снова!',
    is_default: false,
    sort_order: 1,
    created_at: '2026-03-01T10:00:00Z',
  },
]

const mockCreateMutation = { mutate: vi.fn(), isPending: false }
const mockUpdateMutation = { mutate: vi.fn(), isPending: false }
const mockDeleteMutation = { mutate: vi.fn(), isPending: false }

describe('ResponseTemplates', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(usePostMyCrmTemplates).mockReturnValue(
      mockCreateMutation as unknown as ReturnType<typeof usePostMyCrmTemplates>,
    )
    vi.mocked(usePutMyCrmTemplatesId).mockReturnValue(
      mockUpdateMutation as unknown as ReturnType<typeof usePutMyCrmTemplatesId>,
    )
    vi.mocked(useDeleteMyCrmTemplatesId).mockReturnValue(
      mockDeleteMutation as unknown as ReturnType<typeof useDeleteMyCrmTemplatesId>,
    )
  })

  it('renders page title', () => {
    vi.mocked(useGetMyCrmTemplates).mockReturnValue({
      data: { data: mockTemplates },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmTemplates>)

    renderWithProviders(<ResponseTemplates />)

    expect(screen.getByText('Шаблоны ответов')).toBeInTheDocument()
  })

  it('renders template cards', () => {
    vi.mocked(useGetMyCrmTemplates).mockReturnValue({
      data: { data: mockTemplates },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmTemplates>)

    renderWithProviders(<ResponseTemplates />)

    expect(screen.getByText('Благодарность за отзыв')).toBeInTheDocument()
    expect(screen.getByText('Мой кастомный шаблон')).toBeInTheDocument()
  })

  it('shows default tag for system templates', () => {
    vi.mocked(useGetMyCrmTemplates).mockReturnValue({
      data: { data: mockTemplates },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmTemplates>)

    renderWithProviders(<ResponseTemplates />)

    expect(screen.getByText('По умолчанию')).toBeInTheDocument()
  })

  it('shows template body text', () => {
    vi.mocked(useGetMyCrmTemplates).mockReturnValue({
      data: { data: mockTemplates },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmTemplates>)

    renderWithProviders(<ResponseTemplates />)

    expect(screen.getByText('Спасибо за ваш отзыв! Мы ценим ваше мнение.')).toBeInTheDocument()
    expect(screen.getByText('Приходите к нам снова!')).toBeInTheDocument()
  })

  it('shows create button', () => {
    vi.mocked(useGetMyCrmTemplates).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmTemplates>)

    renderWithProviders(<ResponseTemplates />)

    expect(screen.getByText('Создать шаблон')).toBeInTheDocument()
  })

  it('opens create modal when button clicked', () => {
    vi.mocked(useGetMyCrmTemplates).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmTemplates>)

    renderWithProviders(<ResponseTemplates />)

    fireEvent.click(screen.getByText('Создать шаблон'))

    expect(screen.getByText('Новый шаблон')).toBeInTheDocument()
  })

  it('shows empty state when no templates', () => {
    vi.mocked(useGetMyCrmTemplates).mockReturnValue({
      data: { data: [] },
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMyCrmTemplates>)

    renderWithProviders(<ResponseTemplates />)

    expect(screen.getByText(/Нет шаблонов/)).toBeInTheDocument()
  })
})
