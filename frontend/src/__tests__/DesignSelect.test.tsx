import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { Select } from '@/components/design/system'

describe('Design Select', () => {
  it('opens a custom readable dropdown instead of the native browser menu', () => {
    const onChange = vi.fn()

    render(
      <Select
        value="all"
        options={[
          { value: 'all', label: 'Все регионы' },
          { value: 'moscow', label: 'Москва' },
        ]}
        onChange={onChange}
      />,
    )

    const trigger = document.querySelector('.ant-select-selector')
    expect(trigger).toBeTruthy()

    fireEvent.click(trigger as Element)

    expect(screen.getByRole('listbox')).toBeInTheDocument()
    expect(screen.getByRole('option', { name: 'Все регионы' })).toBeInTheDocument()
    expect(screen.getByRole('option', { name: 'Москва' })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('option', { name: 'Москва' }))

    expect(onChange).toHaveBeenCalledWith('moscow', expect.objectContaining({ value: 'moscow' }))
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  it('supports keyboard navigation: ArrowDown opens, ArrowDown moves, Enter commits', () => {
    const onChange = vi.fn()

    render(
      <Select
        value="all"
        options={[
          { value: 'all', label: 'Все регионы' },
          { value: 'moscow', label: 'Москва' },
          { value: 'spb', label: 'Санкт-Петербург' },
        ]}
        onChange={onChange}
      />,
    )

    const trigger = document.querySelector('.ant-select-selector') as HTMLElement
    expect(trigger).toBeTruthy()
    trigger.focus()

    fireEvent.keyDown(trigger, { key: 'ArrowDown' })
    expect(screen.getByRole('listbox')).toBeInTheDocument()

    fireEvent.keyDown(trigger, { key: 'ArrowDown' })
    fireEvent.keyDown(trigger, { key: 'Enter' })

    expect(onChange).toHaveBeenCalledWith('moscow', expect.objectContaining({ value: 'moscow' }))
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })
})
