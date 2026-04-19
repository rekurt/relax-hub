import type { ReactNode } from 'react'
import { Typography } from 'antd'

const { Title, Text } = Typography

interface PageHeaderProps {
  eyebrow?: ReactNode
  title: ReactNode
  description?: ReactNode
  extra?: ReactNode
}

export default function PageHeader({ eyebrow, title, description, extra }: PageHeaderProps) {
  return (
    <header className="bani-page-header">
      <div className="bani-page-header__copy">
        {eyebrow && (
          <Text className="bani-page-header__eyebrow">
            {eyebrow}
          </Text>
        )}
        <Title level={3} className="bani-page-header__title">
          {title}
        </Title>
        {description && (
          <Text className="bani-page-header__description">
            {description}
          </Text>
        )}
      </div>

      {extra && <div className="bani-page-header__extra">{extra}</div>}
    </header>
  )
}
