import type { ReactNode } from 'react'
import { Alert, Button, Card, Result, Space, Spin, Typography } from 'antd'
import EmptyState from '@/components/EmptyState'

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
      <Card style={{ borderRadius: 24, textAlign: 'center', padding: 24 }}>
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <Spin size="large" />
          <Text strong>{resolvedTitle}</Text>
          <Paragraph style={{ marginBottom: 0 }}>{resolvedDescription}</Paragraph>
        </Space>
      </Card>
    )
  }

  if (kind === 'degraded' || compact) {
    return (
      <Alert
        type={kind === 'error' ? 'error' : 'warning'}
        showIcon
        message={resolvedTitle}
        description={resolvedDescription}
        action={actionText ? <Button size="small" onClick={onAction}>{actionText}</Button> : undefined}
      />
    )
  }

  if (kind === 'empty') {
    return (
      <EmptyState
        description={(
          <Space direction="vertical" size={8}>
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
          <Button key="primary" type="primary" onClick={onAction}>
            {actionText}
          </Button>
        ) : null,
        secondaryActionText ? (
          <Button key="secondary" href={secondaryActionLink} onClick={onSecondaryAction}>
            {secondaryActionText}
          </Button>
        ) : null,
      ].filter(Boolean)}
    >
      {typeof resolvedDescription === 'string' ? null : (
        <Paragraph style={{ marginBottom: 0 }}>{resolvedDescription}</Paragraph>
      )}
    </Result>
  )
}
