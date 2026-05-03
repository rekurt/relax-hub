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
})
