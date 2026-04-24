import Axios, { AxiosRequestConfig } from 'axios'
import { AUTH_TOKEN_KEY } from '@/lib/constants'
import { redirectToLogin, shouldRedirectToLogin } from '@/lib/public-route'
import { useAuthStore } from '@/stores/auth'

export const axiosInstance = Axios.create({
  baseURL: '/api/v1',
})

// Auth endpoints that must NEVER receive an Authorization header — a stale
// token from localStorage leaking into login/register/verify requests causes
// the backend to reject otherwise-valid credentials with 401.
const UNAUTHENTICATED_PATHS = [
  '/auth/login',
  '/auth/register',
  '/auth/login/phone',
  '/auth/register/phone',
  '/auth/verify/phone',
  '/auth/forgot-password',
  '/auth/reset-password',
  '/auth/oauth/',
]

const isUnauthenticatedRequest = (url: string | undefined): boolean => {
  if (!url) return false
  return UNAUTHENTICATED_PATHS.some((path) => url.includes(path))
}

axiosInstance.interceptors.request.use((config) => {
  if (isUnauthenticatedRequest(config.url)) {
    // Hard-strip any Authorization header axios defaults / other interceptors
    // may have attached. Login/register must be treated as anonymous calls.
    if (config.headers) {
      delete config.headers.Authorization
      delete (config.headers as Record<string, string>).authorization
    }
    return config
  }

  const token = localStorage.getItem(AUTH_TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

axiosInstance.interceptors.response.use(
  (response) => response,
  (error) => {
    // Never log out on 401 from public auth endpoints — those 401s mean
    // "wrong credentials for this call", not "your session died".
    const url: string | undefined = error?.config?.url
    if (error.response?.status === 401 && !isUnauthenticatedRequest(url)) {
      useAuthStore.getState().logout()
      if (shouldRedirectToLogin(window.location.pathname)) {
        redirectToLogin()
      }
    }
    return Promise.reject(error)
  },
)

export const customInstance = <T>(config: AxiosRequestConfig): Promise<T> => {
  const controller = new AbortController()
  const promise = axiosInstance({
    ...config,
    signal: controller.signal,
  }).then(({ data }) => data) as Promise<T> & { cancel: () => void }

  promise.cancel = () => {
    controller.abort()
  }

  return promise
}

export default customInstance
