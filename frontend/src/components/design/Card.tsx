import type { HTMLAttributes, ReactNode } from 'react'

interface DesignCardProps extends HTMLAttributes<HTMLElement> {
  children: ReactNode
  as?: 'article' | 'div' | 'section'
  tone?: 'default' | 'flat' | 'dark'
}

export default function DesignCard({
  children,
  as: Component = 'div',
  tone = 'default',
  className,
  ...props
}: DesignCardProps) {
  const toneClass = tone === 'default' ? '' : `rh-card--${tone}`
  return (
    <Component className={`rh-card ${toneClass} ${className ?? ''}`.trim()} {...props}>
      {children}
    </Component>
  )
}
