import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { Link } from 'react-router-dom'

interface DesignButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode
  variant?: 'primary' | 'default' | 'ghost'
  size?: 'sm' | 'md' | 'lg'
  block?: boolean
  to?: string
  href?: string
  icon?: ReactNode
}

export default function DesignButton({
  children,
  variant = 'default',
  size = 'md',
  block = false,
  to,
  href,
  icon,
  className,
  type = 'button',
  ...props
}: DesignButtonProps) {
  const classes = [
    'rh-btn',
    `rh-btn--${variant}`,
    size !== 'md' ? `rh-btn--${size}` : '',
    block ? 'rh-btn--block' : '',
    className,
  ].filter(Boolean).join(' ')
  const content = <>{icon}{children}</>

  if (to) {
    return (
      <Link className={classes} to={to}>
        {content}
      </Link>
    )
  }

  if (href) {
    return (
      <a className={classes} href={href}>
        {content}
      </a>
    )
  }

  return (
    <button className={classes} type={type} {...props}>
      {content}
    </button>
  )
}
