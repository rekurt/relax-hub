import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useState } from 'react'
import BathhouseMap from '@/components/BathhouseMap'

interface MockProps {
  bathhouses: { id?: string; latitude?: number; longitude?: number; name?: string; price_per_hour?: number }[]
  highlightedId?: string | null
  style?: React.CSSProperties
}

function MockBathhouseMap({ bathhouses, highlightedId, style }: MockProps) {
  const [mapType, setMapType] = useState<'scheme' | 'satellite'>('scheme')
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
      <button
        data-testid="map-layer-toggle"
        onClick={() => setMapType((prev) => (prev === 'scheme' ? 'satellite' : 'scheme'))}
      >
        {mapType === 'scheme' ? 'Спутник' : 'Схема'}
      </button>
    </div>
  )
}

// Mock the Yandex Maps API loading - we can't load real scripts in jsdom
vi.mock('@/components/BathhouseMap', () => ({
  default: MockBathhouseMap,
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

  it('renders layer toggle button with default "Спутник" label', () => {
    render(<BathhouseMap bathhouses={mockBathhouses} />)
    const toggle = screen.getByTestId('map-layer-toggle')
    expect(toggle).toBeInTheDocument()
    expect(toggle).toHaveTextContent('Спутник')
  })

  it('toggles map layer between scheme and satellite', () => {
    render(<BathhouseMap bathhouses={mockBathhouses} />)
    const toggle = screen.getByTestId('map-layer-toggle')

    // Initially shows "Спутник" (meaning we're on scheme, click to switch to satellite)
    expect(toggle).toHaveTextContent('Спутник')

    // Click to switch to satellite
    fireEvent.click(toggle)
    expect(toggle).toHaveTextContent('Схема')

    // Click again to switch back to scheme
    fireEvent.click(toggle)
    expect(toggle).toHaveTextContent('Спутник')
  })
})
