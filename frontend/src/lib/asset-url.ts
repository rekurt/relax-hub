const DEFAULT_STORAGE_PUBLIC_BASE_URL = 'http://localhost:9102'
const STORAGE_HOSTS = new Set([
  'minio:9000',
  'banya-minio:9000',
  'relax-hub-minio:9000',
  'localhost:9000',
  '127.0.0.1:9000',
  '[::1]:9000',
  'localhost:9102',
  '127.0.0.1:9102',
  '[::1]:9102',
])

function storagePublicBaseURL() {
  const appOrigin = typeof window !== 'undefined' ? window.location.origin : ''

  return (
    import.meta.env.VITE_STORAGE_PUBLIC_BASE_URL
    || appOrigin
    || DEFAULT_STORAGE_PUBLIC_BASE_URL
  ).replace(/\/+$/, '')
}

function rewriteStorageUrl(value: string) {
  let candidate = value
  if (value.startsWith('//')) {
    // Protocol-relative — needs a protocol to parse via URL. In SSR/Node
    // execution there is no window, so bail to the absolute-URL passthrough
    // in resolveAssetUrl rather than throwing ReferenceError.
    if (typeof window === 'undefined') return null
    candidate = `${window.location.protocol}${value}`
  }

  try {
    const url = new URL(candidate)
    if (!STORAGE_HOSTS.has(url.host.toLowerCase())) {
      return null
    }

    return new URL(`${url.pathname}${url.search}${url.hash}`, `${storagePublicBaseURL()}/`).toString()
  } catch {
    return null
  }
}

export function resolveAssetUrl(value?: string | null) {
  if (!value) return ''

  const trimmed = value.trim()
  if (!trimmed) return ''

  const rewrittenStorageUrl = rewriteStorageUrl(trimmed)
  if (rewrittenStorageUrl) return rewrittenStorageUrl

  if (
    trimmed.startsWith('http://')
    || trimmed.startsWith('https://')
    || trimmed.startsWith('//')
    || trimmed.startsWith('data:')
    || trimmed.startsWith('blob:')
  ) {
    return trimmed
  }

  const normalizedPath = trimmed.startsWith('/') ? trimmed : `/${trimmed}`
  if (typeof window === 'undefined') return normalizedPath
  return new URL(normalizedPath, window.location.origin).toString()
}
