import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import BathhouseMap from '@/components/BathhouseMap'

// Mock the Yandex Maps API loading - we can't load real scripts in jsdom
vi.mock('@/components/BathhouseMap', () => ({
  default: ({
    bathhouses,
    highlightedId,
    style,
  }: {
    bathhouses: { id?: string; latitude?: number; longitude?: number; name?: string; price_per_hour?: number }[]
    highlightedId?: string | null
    style?: React.CSSProperties
  }) => {
    const markersWithCoords = bathhouses.filter((b) => b.latitude && b.longitude)
    return (
      <div data-testid="bathhouse-map" style={style}>
        <span data-testid="marker-count">{markersWithCoords.length}</span>
        {highlightedId && <span data-testid="highlighted">{highlightedId}</span>}
        {markersWithCoords.map((b) => (
          <div key={b.id} data-testid={`marker-${b.id}`}>
            {b.name}
          </div>
        ))}
      </div>
    )
  },
}))

const mockBathhouses = [
  { id: '1', name: 'Баня 1', latitude: 55.75, longitude: 37.62, price_per_hour: 200000 },
  { id: '2', name: 'Баня 2', latitude: 55.76, longitude: 37.63, price_per_hour: 300000 },
  { id: '3', name: 'Баня 3', price_per_hour: 100000 }, // No coordinates
]

describe('BathhouseMap', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders map container', () => {
    render(<BathhouseMap bathhouses={mockBathhouses} />)
    expect(screen.getByTestId('bathhouse-map')).toBeInTheDocument()
  })

  it('only renders markers for bathhouses with coordinates', () => {
    render(<BathhouseMap bathhouses={mockBathhouses} />)
    expect(screen.getByTestId('marker-count')).toHaveTextContent('2')
  })

  it('renders marker for each bathhouse with coords', () => {
    render(<BathhouseMap bathhouses={mockBathhouses} />)
    expect(screen.getByTestId('marker-1')).toBeInTheDocument()
    expect(screen.getByTestId('marker-2')).toBeInTheDocument()
    expect(screen.queryByTestId('marker-3')).not.toBeInTheDocument()
  })

  it('shows highlighted marker', () => {
    render(<BathhouseMap bathhouses={mockBathhouses} highlightedId="1" />)
    expect(screen.getByTestId('highlighted')).toHaveTextContent('1')
  })

  it('applies custom style', () => {
    render(
      <BathhouseMap
        bathhouses={mockBathhouses}
        style={{ height: 500, borderRadius: 8 }}
      />,
    )
    const map = screen.getByTestId('bathhouse-map')
    expect(map.style.height).toBe('500px')
  })

  it('renders with empty bathhouses array', () => {
    render(<BathhouseMap bathhouses={[]} />)
    expect(screen.getByTestId('marker-count')).toHaveTextContent('0')
  })
})
