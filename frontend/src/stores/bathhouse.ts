import { create } from 'zustand'

const SELECTED_BATHHOUSE_KEY = 'rh_selected_bathhouse'

export interface BathhouseState {
  selectedBathhouseId: string | null
  setSelectedBathhouseId: (id: string | null) => void
}

export const useBathhouseStore = create<BathhouseState>((set) => ({
  selectedBathhouseId: localStorage.getItem(SELECTED_BATHHOUSE_KEY),

  setSelectedBathhouseId: (id) => {
    if (id) {
      localStorage.setItem(SELECTED_BATHHOUSE_KEY, id)
    } else {
      localStorage.removeItem(SELECTED_BATHHOUSE_KEY)
    }
    set({ selectedBathhouseId: id })
  },
}))
