import { useState } from 'react'
import { Button, App } from 'antd'
import { ShareAltOutlined, CopyOutlined, CheckOutlined } from '@ant-design/icons'

interface ShareButtonProps {
  url: string
  title?: string
  text?: string
  onBeforeShare?: () => Promise<string | undefined>
  size?: 'small' | 'middle' | 'large'
}

export default function ShareButton({ url, title, text, onBeforeShare, size = 'middle' }: ShareButtonProps) {
  const { message } = App.useApp()
  const [loading, setLoading] = useState(false)
  const [copied, setCopied] = useState(false)

  const handleShare = async () => {
    setLoading(true)
    try {
      let shareUrl = url
      if (onBeforeShare) {
        const result = await onBeforeShare()
        if (result) shareUrl = result
      }

      if (navigator.share) {
        await navigator.share({
          title: title ?? 'Bani',
          text: text ?? '',
          url: shareUrl,
        })
      } else {
        await navigator.clipboard.writeText(shareUrl)
        setCopied(true)
        message.success('Ссылка скопирована')
        setTimeout(() => setCopied(false), 2000)
      }
    } catch (err) {
      if ((err as DOMException)?.name === 'AbortError') return
      try {
        const shareUrl = url
        await navigator.clipboard.writeText(shareUrl)
        setCopied(true)
        message.success('Ссылка скопирована')
        setTimeout(() => setCopied(false), 2000)
      } catch {
        message.error('Не удалось поделиться')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <Button
      icon={copied ? <CheckOutlined /> : navigator.share ? <ShareAltOutlined /> : <CopyOutlined />}
      onClick={handleShare}
      loading={loading}
      size={size}
    >
      {copied ? 'Скопировано' : 'Поделиться'}
    </Button>
  )
}
