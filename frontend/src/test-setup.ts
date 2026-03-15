import '@testing-library/jest-dom'

;(globalThis as Record<string, unknown>).ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}

;(globalThis as Record<string, unknown>).IntersectionObserver = class IntersectionObserver {
  readonly root = null
  readonly rootMargin = ''
  readonly thresholds: ReadonlyArray<number> = []
  constructor(private callback: IntersectionObserverCallback) {}
  observe(target: Element) {
    this.callback(
      [{ isIntersecting: true, target, intersectionRatio: 1 } as IntersectionObserverEntry],
      this as unknown as IntersectionObserver,
    )
  }
  unobserve() {}
  disconnect() {}
  takeRecords(): IntersectionObserverEntry[] { return [] }
}

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})
