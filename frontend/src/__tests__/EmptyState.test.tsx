import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import EmptyState from '@/components/EmptyState'

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return { ...actual, useNavigate: () => mockNavigate }
})

function renderComponent(props: React.ComponentProps<typeof EmptyState>) {
  return render(
    <MemoryRouter>
      <EmptyState {...props} />
    </MemoryRouter>,
  )
}

describe('EmptyState', () => {
  beforeEach(() => {
    mockNavigate.mockClear()
  })

  it('renders description text', () => {
    renderComponent({ description: 'Нет данных для отображения' })
    expect(screen.getByText('Нет данных для отображения')).toBeInTheDocument()
  })

  it('renders without action button when no actionText', () => {
    renderComponent({ description: 'Пусто' })
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('renders action button when actionText provided', () => {
    renderComponent({
      description: 'Список пуст',
      actionText: 'Найти баню',
      actionLink: '/client/search',
    })
    expect(screen.getByRole('button', { name: /Найти баню/ })).toBeInTheDocument()
  })

  it('navigates on action button click with actionLink', () => {
    renderComponent({
      description: 'Пусто',
      actionText: 'Перейти',
      actionLink: '/some/path',
    })
    fireEvent.click(screen.getByRole('button', { name: /Перейти/ }))
    expect(mockNavigate).toHaveBeenCalledWith('/some/path')
  })

  it('calls onAction callback when provided', () => {
    const onAction = vi.fn()
    renderComponent({
      description: 'Пусто',
      actionText: 'Действие',
      onAction,
    })
    fireEvent.click(screen.getByRole('button', { name: /Действие/ }))
    expect(onAction).toHaveBeenCalledOnce()
  })

  it('prefers onAction over actionLink', () => {
    const onAction = vi.fn()
    renderComponent({
      description: 'Пусто',
      actionText: 'Действие',
      actionLink: '/path',
      onAction,
    })
    fireEvent.click(screen.getByRole('button', { name: /Действие/ }))
    expect(onAction).toHaveBeenCalledOnce()
    expect(mockNavigate).not.toHaveBeenCalled()
  })

  it('renders empty image', () => {
    const { container } = renderComponent({ description: 'Пусто' })
    expect(container.querySelector('.ant-empty')).toBeInTheDocument()
  })

  it('snapshot: basic empty state', () => {
    const { container } = renderComponent({ description: 'У вас пока нет бронирований' })
    expect(container.firstChild).toMatchSnapshot()
  })

  it('snapshot: empty state with action', () => {
    const { container } = renderComponent({
      description: 'Ваш список избранного пуст',
      actionText: 'Найти баню',
      actionLink: '/client/search',
    })
    expect(container.firstChild).toMatchSnapshot()
  })
})
