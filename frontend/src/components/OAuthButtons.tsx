import { useEffect, useMemo, useState } from 'react'
import { Button, Divider, Space } from 'antd'
import { axiosInstance } from '@/api/axios-instance'

const OAUTH_PROVIDERS = [
  { key: 'vk', label: 'VK', color: '#0077FF' },
  { key: 'yandex', label: 'Яндекс', color: '#FC3F1D' },
  { key: 'google', label: 'Google', color: '#4285F4' },
] as const

interface OAuthButtonsProps {
  referralCode?: string
}

export default function OAuthButtons({ referralCode }: OAuthButtonsProps) {
  const [availableProviders, setAvailableProviders] = useState<string[] | null>(null)

  useEffect(() => {
    let active = true

    axiosInstance
      .get<{ success?: boolean; data?: { providers?: string[] } }>('/auth/oauth/providers')
      .then((response) => {
        if (!active) return
        setAvailableProviders(response.data.data?.providers ?? [])
      })
      .catch(() => {
        if (active) {
          setAvailableProviders([])
        }
      })

    return () => {
      active = false
    }
  }, [])

  const visibleProviders = useMemo(() => {
    if (availableProviders == null) return []
    return OAUTH_PROVIDERS.filter((provider) => availableProviders.includes(provider.key))
  }, [availableProviders])

  const handleOAuth = (provider: string) => {
    const params = new URLSearchParams()
    if (referralCode) {
      params.set('referral_code', referralCode)
    }
    const query = params.toString()
    const url = `/api/v1/auth/oauth/${provider}${query ? `?${query}` : ''}`
    window.location.assign(url)
  }

  if (!visibleProviders.length) {
    return null
  }

  return (
    <div>
      <Divider plain style={{ margin: '8px 0 16px' }}>
        или войдите через
      </Divider>
      <Space orientation="vertical" style={{ width: '100%' }} size="small">
        {visibleProviders.map((provider) => (
          <Button
            key={provider.key}
            block
            size="large"
            onClick={() => handleOAuth(provider.key)}
            style={{
              borderColor: provider.color,
              color: provider.color,
            }}
          >
            Войти через {provider.label}
          </Button>
        ))}
      </Space>
    </div>
  )
}
