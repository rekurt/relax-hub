import Axios, { AxiosRequestConfig } from 'axios'

const AUTH_TOKEN_KEY = 'bani_token'

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
      localStorage.removeItem(AUTH_TOKEN_KEY)
      window.location.href = '/login'
    }
    return Promise.reject(error)
  },
)

export const customInstance = <T>(config: AxiosRequestConfig): Promise<T> => {
  const source = Axios.CancelToken.source()
  const promise = axiosInstance({
    ...config,
    cancelToken: source.token,
  }).then(({ data }) => data)

  // @ts-expect-error -- orval cancel pattern
  promise.cancel = () => {
    source.cancel('Query was cancelled')
  }

  return promise
}

export default customInstance
