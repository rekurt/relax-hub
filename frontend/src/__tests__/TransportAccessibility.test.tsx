import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import TransportAccessibility, { type TransportItem } from '@/components/TransportAccessibility'

const mockItems: TransportItem[] = [
  { type: 'metro', name: 'Парк культуры', distance_meters: 450, lat: 55.7, lng: 37.5 },
  { type: 'bus_stop', name: 'ул. Мира', distance_meters: 200, lat: 55.7, lng: 37.5 },
  { type: 'parking', name: 'Парковка ТЦ', distance_meters: 1200, lat: 55.7, lng: 37.5 },
]

describe('TransportAccessibility', () => {
  it('renders nothing when items is empty', () => {
    const { container } = render(<TransportAccessibility items={[]} />)
    expect(container.innerHTML).toBe('')
  })

  it('renders transport items with names and distances', () => {
    render(<TransportAccessibility items={mockItems} />)

    expect(screen.getByText('Транспорт рядом:')).toBeInTheDocument()
    expect(screen.getByText('Парк культуры')).toBeInTheDocument()
    expect(screen.getByText('ул. Мира')).toBeInTheDocument()
    expect(screen.getByText('Парковка ТЦ')).toBeInTheDocument()
  })

  it('formats distance in meters for <1km', () => {
    render(<TransportAccessibility items={mockItems} />)
    expect(screen.getByText('— 450 м')).toBeInTheDocument()
    expect(screen.getByText('— 200 м')).toBeInTheDocument()
  })

  it('formats distance in km for >=1km', () => {
    render(<TransportAccessibility items={mockItems} />)
    expect(screen.getByText('— 1.2 км')).toBeInTheDocument()
  })

  it('renders type labels', () => {
    render(<TransportAccessibility items={mockItems} />)
    expect(screen.getByText('Метро')).toBeInTheDocument()
    expect(screen.getByText('Остановка')).toBeInTheDocument()
    expect(screen.getByText('Парковка')).toBeInTheDocument()
  })
})
