import Axios, { AxiosRequestConfig } from 'axios'
import { AUTH_TOKEN_KEY } from '@/lib/constants'
import { useAuthStore } from '@/stores/auth'

export const axiosInstance = Axios.create({
  baseURL: '/api/v1',
})

axiosInstance.interceptors.request.use((config) => {
  const token = localStorage.getItem(AUTH_TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

axiosInstance.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout()
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
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
