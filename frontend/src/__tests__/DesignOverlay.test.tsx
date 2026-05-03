import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { Drawer, Modal } from '@/components/design/system'

describe('Design overlays', () => {
  it('renders modal as a body-level dialog with header, body, footer and close control', () => {
    const onCancel = vi.fn()
    const host = document.createElement('section')
    document.body.appendChild(host)

    render(
      <Modal open title="Добавить FAQ" width={640} onCancel={onCancel} onOk={vi.fn()} okText="Создать">
        <label htmlFor="faq-question">Вопрос</label>
        <textarea id="faq-question" />
      </Modal>,
      { container: host },
    )

    const dialog = screen.getByRole('dialog', { name: 'Добавить FAQ' })

    expect(dialog).toHaveClass('rh-modal')
    expect(dialog).toHaveStyle({ width: '640px' })
    expect(dialog.parentElement).toHaveClass('rh-modal-root')
    expect(dialog.parentElement?.parentElement).toBe(document.body)
    expect(host.querySelector('.rh-modal-root')).toBeNull()
    expect(screen.getByText('Вопрос')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Создать' })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Закрыть' }))

    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('renders drawer as a body-level dialog', () => {
    const onClose = vi.fn()
    const host = document.createElement('section')
    document.body.appendChild(host)

    render(
      <Drawer open title="Фильтры" onClose={onClose}>
        Настройки
      </Drawer>,
      { container: host },
    )

    const dialog = screen.getByRole('dialog', { name: 'Фильтры' })

    expect(dialog).toHaveClass('rh-drawer')
    expect(dialog.parentElement).toHaveClass('rh-drawer-root')
    expect(dialog.parentElement?.parentElement).toBe(document.body)
    expect(host.querySelector('.rh-drawer-root')).toBeNull()

    fireEvent.click(screen.getByRole('button', { name: 'Закрыть' }))

    expect(onClose).toHaveBeenCalledTimes(1)
  })
})
