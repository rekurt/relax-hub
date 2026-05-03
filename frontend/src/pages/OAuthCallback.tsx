import { useEffect, useMemo, useState } from 'react'
import { useParams, useSearchParams, useNavigate } from 'react-router-dom'
import { Spin, Result, Typography } from 'antd'
import { getAuthOauthProviderCallback } from '@/api/generated/oauth/oauth'
import { useAuthStore } from '@/stores/auth'
import { getRoleHomePath } from '@/stores/auth'

const { Text } = Typography

const PROVIDER_NAMES: Record<string, string> = {
  vk: 'VK',
  yandex: 'Яндекс',
  google: 'Google',
}

export default function OAuthCallback() {
  const { provider } = useParams<{ provider: string }>()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const setAuth = useAuthStore((s) => s.setAuth)
  const [apiError, setApiError] = useState<string | null>(null)

  const code = searchParams.get('code')
  const state = searchParams.get('state')
  const errorParam = searchParams.get('error')

  const immediateError = useMemo(() => {
    if (errorParam) {
      return searchParams.get('error_description') || errorParam
    }
    if (!code || !state || !provider) {
      return 'Отсутствуют необходимые параметры авторизации'
    }
    return null
  }, [errorParam, code, state, provider, searchParams])

  useEffect(() => {
    if (immediateError || !code || !state || !provider) return

    let cancelled = false

    getAuthOauthProviderCallback(provider, { code, state })
      .then((response) => {
        if (cancelled) return
        if (response.success && response.data?.token && response.data.user) {
          setAuth(response.data.token, response.data.user)
          const homePath = getRoleHomePath(response.data.user.role)
          navigate(homePath, { replace: true })
        } else {
          setApiError(response.error?.message || 'Ошибка авторизации через провайдер')
        }
      })
      .catch(() => {
        if (!cancelled) {
          setApiError('Не удалось завершить авторизацию. Попробуйте снова.')
        }
      })

    return () => {
      cancelled = true
    }
  }, [provider, code, state, immediateError, setAuth, navigate])

  const error = immediateError || apiError
  const providerName = PROVIDER_NAMES[provider || ''] || provider

  if (error) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: 'rgba(248, 244, 236, 0.78)' }}>
        <Result
          status="error"
          title="Ошибка авторизации"
          subTitle={error}
          extra={
            <a href="/login">Вернуться на страницу входа</a>
          }
        />
      </div>
    )
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: 'rgba(248, 244, 236, 0.78)', gap: 16 }}>
      <Spin size="large" />
      <Text type="secondary">Авторизация через {providerName}...</Text>
    </div>
  )
}
