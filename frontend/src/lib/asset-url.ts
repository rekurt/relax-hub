export function resolveAssetUrl(value?: string | null) {
  if (!value) return ''

  const trimmed = value.trim()
  if (!trimmed) return ''

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
  return new URL(normalizedPath, window.location.origin).toString()
}
