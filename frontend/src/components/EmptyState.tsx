import type { ReactNode } from 'react'
import { Empty } from 'antd'
import { useNavigate } from 'react-router-dom'
import { DesignButton } from '@/components/design'

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
      className="bani-empty-state"
      image={image ?? Empty.PRESENTED_IMAGE_SIMPLE}
      description={description}
    >
      {actionText && (
        <DesignButton variant="primary" icon={icon} onClick={handleAction}>
          {actionText}
        </DesignButton>
      )}
    </Empty>
  )
}
