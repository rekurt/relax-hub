import { useBathhouseStore } from '@/stores/bathhouse'

const SELECTED_BATHHOUSE_KEY = 'bani_selected_bathhouse'

describe('bathhouse store', () => {
  beforeEach(() => {
    localStorage.clear()
    useBathhouseStore.setState({ selectedBathhouseId: null })
  })

  it('initializes with null when no localStorage value', () => {
    const { selectedBathhouseId } = useBathhouseStore.getState()
    expect(selectedBathhouseId).toBeNull()
  })

  it('sets selected bathhouse and persists to localStorage', () => {
    useBathhouseStore.getState().setSelectedBathhouseId('bath-123')

    expect(useBathhouseStore.getState().selectedBathhouseId).toBe('bath-123')
    expect(localStorage.getItem(SELECTED_BATHHOUSE_KEY)).toBe('bath-123')
  })

  it('clears selected bathhouse and removes from localStorage', () => {
    useBathhouseStore.getState().setSelectedBathhouseId('bath-123')
    useBathhouseStore.getState().setSelectedBathhouseId(null)

    expect(useBathhouseStore.getState().selectedBathhouseId).toBeNull()
    expect(localStorage.getItem(SELECTED_BATHHOUSE_KEY)).toBeNull()
  })

  it('can update selected bathhouse', () => {
    useBathhouseStore.getState().setSelectedBathhouseId('bath-1')
    useBathhouseStore.getState().setSelectedBathhouseId('bath-2')

    expect(useBathhouseStore.getState().selectedBathhouseId).toBe('bath-2')
    expect(localStorage.getItem(SELECTED_BATHHOUSE_KEY)).toBe('bath-2')
  })
})
