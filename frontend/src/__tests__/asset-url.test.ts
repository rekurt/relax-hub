import { resolveAssetUrl } from '@/lib/asset-url'

describe('resolveAssetUrl', () => {
  it('rewrites Docker-internal MinIO URLs to the current app origin', () => {
    expect(resolveAssetUrl('http://minio:9000/bani/avatars/avatar.png')).toBe(
      new URL('/bani/avatars/avatar.png', window.location.origin).toString(),
    )
  })

  it('rewrites local MinIO URLs to the current app origin', () => {
    expect(resolveAssetUrl('http://localhost:9102/bani/avatars/avatar.png')).toBe(
      new URL('/bani/avatars/avatar.png', window.location.origin).toString(),
    )
  })

  it('keeps normal absolute URLs unchanged', () => {
    expect(resolveAssetUrl('https://cdn.example.com/photo.jpg')).toBe('https://cdn.example.com/photo.jpg')
  })

  it('returns protocol-relative URLs unchanged when window is unavailable (SSR-safe)', () => {
    const originalWindow = globalThis.window
    // @ts-expect-error simulate SSR
    delete (globalThis as any).window
    try {
      expect(resolveAssetUrl('//cdn.example.com/a.png')).toBe('//cdn.example.com/a.png')
      // Relative paths fall back to a server-relative path string instead of
      // throwing on the missing window.location.
      expect(resolveAssetUrl('/local/path.png')).toBe('/local/path.png')
    } finally {
      globalThis.window = originalWindow
    }
  })
})
