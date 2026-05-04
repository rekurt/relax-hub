import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Table } from '@/components/design/system'
import type { ColumnsType } from '@/components/design/types'

interface Row {
  id: string
  name: string
  views: number
  bookings: number
}

const columns: ColumnsType<Row> = [
  { title: 'Название', dataIndex: 'name', key: 'name', width: 260 },
  { title: 'Просмотры', dataIndex: 'views', key: 'views', width: 130 },
  { title: 'Бронирования', dataIndex: 'bookings', key: 'bookings', width: 150 },
]

describe('Design Table', () => {
  it('derives a safe min width from column widths even when scroll is smaller', () => {
    const { container } = render(
      <Table
        columns={columns}
        dataSource={[{ id: '1', name: 'Баня Люкс', views: 120, bookings: 12 }]}
        rowKey="id"
        scroll={{ x: 320 }}
      />,
    )

    expect(container.querySelector('table')).toHaveStyle({ minWidth: '540px' })
    expect(screen.getByText('Бронирования')).toBeInTheDocument()
  })

  it('renders empty state as a full-width table row', () => {
    const { container } = render(
      <Table
        columns={columns}
        dataSource={[]}
        rowKey="id"
        locale={{ emptyText: 'Нет данных' }}
      />,
    )

    const emptyCell = container.querySelector('.rh-table__empty-cell')
    expect(emptyCell).toHaveAttribute('colspan', '3')
    expect(screen.getByText('Нет данных')).toHaveClass('rh-table__empty')
  })
})
