import Axios from 'axios'

const EXACT_PUBLIC_PATHS = new Set([
  '/',
  '/catalog',
  '/checkout',
  '/certificates',
  '/faq',
  '/contacts',
])

const PUBLIC_PATH_PREFIXES = [
  '/bathhouses/',
  '/share/booking/',
]

const AUTH_PATHS = new Set([
  '/login',
  '/register',
  '/forgot-password',
])

const AUTH_PATH_PREFIXES = [
  '/auth/oauth/callback/',
]

function normalizePathname(pathname?: string) {
  if (!pathname) return '/'
  if (pathname.length > 1 && pathname.endsWith('/')) {
    return pathname.slice(0, -1)
  }
  return pathname
}

export function isPublicPathname(pathname?: string) {
  const normalized = normalizePathname(pathname)
  if (EXACT_PUBLIC_PATHS.has(normalized)) return true
  return PUBLIC_PATH_PREFIXES.some((prefix) => normalized.startsWith(prefix))
}

export function isAuthPathname(pathname?: string) {
  const normalized = normalizePathname(pathname)
  if (AUTH_PATHS.has(normalized)) return true
  return AUTH_PATH_PREFIXES.some((prefix) => normalized.startsWith(prefix))
}

export function shouldRedirectToLogin(pathname?: string) {
  const normalized = normalizePathname(pathname)
  if (isPublicPathname(normalized) || isAuthPathname(normalized)) {
    return false
  }
  return true
}

export function redirectToLogin() {
  if (typeof window === 'undefined') return
  if (window.location.pathname === '/login') return

  try {
    window.history.replaceState({}, '', '/login')
    window.dispatchEvent(new PopStateEvent('popstate'))
  } catch {
    window.location.assign('/login')
  }
}

export function getPublicErrorMessage(error: unknown, fallback = 'Не удалось выполнить запрос. Попробуйте еще раз.') {
  if (!Axios.isAxiosError(error)) {
    return fallback
  }

  const message = error.response?.data?.error?.message
  if (typeof message !== 'string') {
    return fallback
  }

  const trimmed = message.trim()
  if (!trimmed || trimmed.length > 180) {
    return fallback
  }

  return trimmed
}
