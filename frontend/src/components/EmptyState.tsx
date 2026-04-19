import type { ReactNode } from 'react'
import { Button, Empty } from 'antd'
import { useNavigate } from 'react-router-dom'

interface EmptyStateProps {
  description: ReactNode
  actionText?: string
  actionLink?: string
  onAction?: () => void
  icon?: ReactNode
  image?: ReactNode
}

export default function EmptyState({
  description,
  actionText,
  actionLink,
  onAction,
  icon,
  image,
}: EmptyStateProps) {
  const navigate = useNavigate()

  const handleAction = () => {
    if (onAction) {
      onAction()
    } else if (actionLink) {
      navigate(actionLink)
    }
  }

  return (
    <Empty
      image={image ?? Empty.PRESENTED_IMAGE_SIMPLE}
      description={description}
      style={{ padding: '48px 0' }}
    >
      {actionText && (
        <Button type="primary" icon={icon} onClick={handleAction}>
          {actionText}
        </Button>
      )}
    </Empty>
  )
}
