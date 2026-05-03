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

function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(' ')
}

const baseClasses =
  'rh-btn inline-flex min-h-11 min-w-0 items-center justify-center gap-2 whitespace-nowrap rounded-rh-pill border border-transparent px-[22px] font-sans text-sm font-semibold leading-none tracking-normal transition duration-200 ease-in-out focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-[rgba(15,118,110,0.16)] disabled:pointer-events-none disabled:opacity-55'

const variantClasses = {
  primary:
    'rh-btn--primary bg-[linear-gradient(135deg,#0f766e,#0a5f59)] text-white shadow-rh-primary hover:-translate-y-px hover:shadow-[0_18px_30px_rgba(15,118,110,0.24)]',
  default:
    'rh-btn--default border-[rgba(15,23,42,0.12)] bg-white/75 text-rh-text shadow-none hover:-translate-y-px hover:border-[rgba(15,118,110,0.28)] hover:bg-white/95 hover:shadow-rh-soft',
  ghost:
    'rh-btn--ghost bg-transparent text-rh-text shadow-none hover:bg-[rgba(15,118,110,0.08)] hover:text-rh-primary-strong',
} as const

const sizeClasses = {
  sm: 'rh-btn--sm min-h-9 px-4 text-[13px]',
  md: '',
  lg: 'rh-btn--lg min-h-[52px] px-7 text-[15px]',
} as const

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
  const classes = cx(baseClasses, variantClasses[variant], sizeClasses[size], block && 'rh-btn--block w-full', className)
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
