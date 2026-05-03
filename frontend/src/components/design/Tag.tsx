import type { ReactNode } from 'react'

export type DesignTagTone = 'default' | 'primary' | 'gold' | 'cyan' | 'red' | 'green' | 'ghost'

interface DesignTagProps {
  children: ReactNode
  tone?: DesignTagTone
  className?: string
}

export default function DesignTag({ children, tone = 'default', className }: DesignTagProps) {
  const toneClass = tone === 'default' ? '' : `rh-tag--${tone}`
  return <span className={`rh-tag ${toneClass} ${className ?? ''}`.trim()}>{children}</span>
}
