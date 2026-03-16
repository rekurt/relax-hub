import { Button, Divider, Space } from 'antd'

const OAUTH_PROVIDERS = [
  { key: 'vk', label: 'VK', color: '#0077FF' },
  { key: 'yandex', label: 'Яндекс', color: '#FC3F1D' },
  { key: 'google', label: 'Google', color: '#4285F4' },
] as const

interface OAuthButtonsProps {
  referralCode?: string
}

export default function OAuthButtons({ referralCode }: OAuthButtonsProps) {
  const handleOAuth = (provider: string) => {
    const params = new URLSearchParams()
    if (referralCode) {
      params.set('referral_code', referralCode)
    }
    const query = params.toString()
    const url = `/api/v1/auth/oauth/${provider}${query ? `?${query}` : ''}`
    window.location.assign(url)
  }

  return (
    <div>
      <Divider plain style={{ margin: '8px 0 16px' }}>
        или войдите через
      </Divider>
      <Space direction="vertical" style={{ width: '100%' }} size="small">
        {OAUTH_PROVIDERS.map((provider) => (
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
