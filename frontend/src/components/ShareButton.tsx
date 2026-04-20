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

export async function copyToClipboard(value: string) {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(value)
    return
  }

  const textarea = document.createElement('textarea')
  textarea.value = value
  textarea.setAttribute('readonly', 'true')
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  textarea.style.pointerEvents = 'none'
  document.body.appendChild(textarea)
  textarea.select()
  textarea.setSelectionRange(0, textarea.value.length)
  const copied = document.execCommand('copy')
  document.body.removeChild(textarea)

  if (!copied) {
    throw new Error('clipboard_unavailable')
  }
}

function canNativeShare() {
  return typeof navigator.share === 'function'
}

export default function ShareButton({ url, title, text, onBeforeShare, size = 'middle' }: ShareButtonProps) {
  const { message } = App.useApp()
  const [loading, setLoading] = useState(false)
  const [copied, setCopied] = useState(false)

  const handleShare = async () => {
    setLoading(true)
    let shareUrl = url
    try {
      if (onBeforeShare) {
        const result = await onBeforeShare()
        if (result) shareUrl = result
      }

      if (canNativeShare()) {
        await navigator.share({
          title: title ?? 'Bani',
          text: text ?? '',
          url: shareUrl,
        })
      } else {
        await copyToClipboard(shareUrl)
        setCopied(true)
        message.success('Ссылка скопирована')
        setTimeout(() => setCopied(false), 2000)
      }
    } catch (err) {
      if ((err as DOMException)?.name === 'AbortError') return
      try {
        await copyToClipboard(shareUrl)
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
      icon={copied ? <CheckOutlined /> : canNativeShare() ? <ShareAltOutlined /> : <CopyOutlined />}
      onClick={handleShare}
      loading={loading}
      size={size}
    >
      {copied ? 'Скопировано' : 'Поделиться'}
    </Button>
  )
}
