import type { ReactNode } from 'react'
import { Alert, Result, Space, Spin, Typography } from '@/components/design/system'
import EmptyState from '@/components/EmptyState'
import { DesignButton, DesignCard } from '@/components/design'

const { Paragraph, Text } = Typography

type PublicStateKind = 'loading' | 'empty' | 'degraded' | 'error'

interface PublicStateProps {
  kind: PublicStateKind
  title?: ReactNode
  description?: ReactNode
  actionText?: string
  actionLink?: string
  secondaryActionText?: string
  secondaryActionLink?: string
  onAction?: () => void
  onSecondaryAction?: () => void
  compact?: boolean
}

export default function PublicState({
  kind,
  title,
  description,
  actionText,
  actionLink,
  secondaryActionText,
  secondaryActionLink,
  onAction,
  onSecondaryAction,
  compact = false,
}: PublicStateProps) {
  const resolvedTitle = title ?? (
    kind === 'loading'
      ? 'Загружаем данные'
      : kind === 'empty'
        ? 'Ничего не найдено'
        : kind === 'degraded'
          ? 'Блок временно недоступен'
          : 'Не удалось загрузить данные'
  )

  const resolvedDescription = description ?? (
    kind === 'loading'
      ? 'Пожалуйста, подождите.'
      : kind === 'empty'
        ? 'Попробуйте изменить параметры или вернитесь к поиску.'
        : kind === 'degraded'
          ? 'Попробуйте обновить этот блок чуть позже.'
          : 'Повторите попытку или вернитесь к предыдущему шагу.'
  )

  if (kind === 'loading') {
    return (
      <DesignCard className="rh-public-state-card">
        <Space orientation="vertical" size={16} style={{ width: '100%' }}>
          <Spin size="large" />
          <Text strong>{resolvedTitle}</Text>
          <Paragraph style={{ marginBottom: 0 }}>{resolvedDescription}</Paragraph>
        </Space>
      </DesignCard>
    )
  }

  if (kind === 'degraded' || compact) {
    return (
      <Alert
        type={kind === 'error' ? 'error' : 'warning'}
        showIcon
        title={resolvedTitle}
        description={resolvedDescription}
        action={actionText ? <DesignButton size="sm" onClick={onAction}>{actionText}</DesignButton> : undefined}
      />
    )
  }

  if (kind === 'empty') {
    return (
      <EmptyState
        description={(
          <Space orientation="vertical" size={8}>
            <Text strong>{resolvedTitle}</Text>
            <Text type="secondary">{resolvedDescription}</Text>
          </Space>
        )}
        actionText={actionText}
        actionLink={actionLink}
        onAction={onAction}
      />
    )
  }

  return (
    <Result
      status="error"
      title={resolvedTitle}
      subTitle={typeof resolvedDescription === 'string' ? resolvedDescription : undefined}
      extra={[
        actionText ? (
          <DesignButton key="primary" variant="primary" onClick={onAction}>
            {actionText}
          </DesignButton>
        ) : null,
        secondaryActionText ? (
          <DesignButton key="secondary" href={secondaryActionLink} onClick={onSecondaryAction}>
            {secondaryActionText}
          </DesignButton>
        ) : null,
      ].filter(Boolean)}
    >
      {typeof resolvedDescription === 'string' ? null : (
        <Paragraph style={{ marginBottom: 0 }}>{resolvedDescription}</Paragraph>
      )}
    </Result>
  )
}
