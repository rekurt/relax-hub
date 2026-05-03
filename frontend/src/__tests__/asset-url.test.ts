import { resolveAssetUrl } from '@/lib/asset-url'

describe('resolveAssetUrl', () => {
  it('rewrites Docker-internal MinIO URLs to the browser-accessible local endpoint', () => {
    expect(resolveAssetUrl('http://minio:9000/bani/avatars/avatar.png')).toBe(
      'http://localhost:9102/bani/avatars/avatar.png',
    )
  })

  it('keeps normal absolute URLs unchanged', () => {
    expect(resolveAssetUrl('https://cdn.example.com/photo.jpg')).toBe('https://cdn.example.com/photo.jpg')
  })
})
