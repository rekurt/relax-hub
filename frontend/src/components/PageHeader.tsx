import type { ReactNode } from 'react'
import { SectionHeader } from '@/components/design'

interface PageHeaderProps {
  eyebrow?: ReactNode
  title: ReactNode
  description?: ReactNode
  extra?: ReactNode
  size?: 'default' | 'compact'
}

export default function PageHeader({ eyebrow, title, description, extra, size = 'default' }: PageHeaderProps) {
  return (
    <SectionHeader
      className={`rh-page-header rh-page-header--${size}`}
      eyebrow={eyebrow}
      title={title}
      subtitle={description}
      extra={extra}
    />
  )
}
