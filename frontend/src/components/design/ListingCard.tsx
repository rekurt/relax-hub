import type { ReactNode } from 'react'
import DesignIcon from '@/components/design/Icon'
import PhotoPlaceholder from '@/components/design/PhotoPlaceholder'
import RatingBadge from '@/components/design/RatingBadge'
import DesignTag, { type DesignTagTone } from '@/components/design/Tag'

interface DesignListingCardProps {
  name?: ReactNode
  address?: ReactNode
  price?: ReactNode
  rating?: number | null
  reviewCount?: number | null
  tags?: ReactNode[]
  badge?: ReactNode
  badgeTone?: DesignTagTone
  imageUrl?: string
  imageAlt?: string
  verified?: boolean
  actions?: ReactNode
  onClick?: () => void
  className?: string
}

function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(' ')
}

export default function DesignListingCard({
  name,
  address,
  price,
  rating,
  reviewCount,
  tags = [],
  badge,
  badgeTone = 'gold',
  imageUrl,
  imageAlt,
  verified,
  actions,
  onClick,
  className,
}: DesignListingCardProps) {
  return (
    <article
      className={cx(
        'rh-listing-card group min-w-0 overflow-hidden rounded-rh-2xl border border-[rgba(15,23,42,0.12)] bg-[linear-gradient(180deg,rgba(255,255,255,0.92),rgba(255,252,246,0.84))] shadow-rh-soft backdrop-blur-[18px] transition duration-200 ease-in-out hover:-translate-y-0.5 hover:shadow-rh',
        onClick && 'cursor-pointer focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-[rgba(15,118,110,0.16)]',
        className,
      )}
      onClick={onClick}
      role={onClick ? 'button' : undefined}
      tabIndex={onClick ? 0 : undefined}
      onKeyDown={(event) => {
        if (!onClick) return
        if (event.key === 'Enter' || event.key === ' ') {
          event.preventDefault()
          onClick()
        }
      }}
    >
      <div className="rh-listing-card__media relative aspect-[4/3] overflow-hidden rounded-t-rh-2xl">
        {imageUrl ? (
          <img className="rh-listing-card__image h-full w-full object-cover transition duration-300 group-hover:scale-[1.02]" src={imageUrl} alt={imageAlt ?? ''} />
        ) : (
          <PhotoPlaceholder className="rh-listing-card__placeholder h-full min-h-0 rounded-none" />
        )}
        {badge && (
          <DesignTag tone={badgeTone} className="rh-listing-card__badge absolute left-3 top-3">
            {badge}
          </DesignTag>
        )}
      </div>
      <div className="rh-listing-card__body grid gap-3 p-4">
        <div className="rh-listing-card__head flex items-start justify-between gap-3">
          <h3 className="rh-listing-card__title m-0 flex min-w-0 items-center gap-1.5 font-sans text-[18px] font-extrabold leading-tight tracking-normal text-rh-text">
            {name}
            {verified && <DesignIcon name="checkc" size={14} className="rh-listing-card__verified shrink-0 text-rh-primary" />}
          </h3>
          {price && <div className="rh-listing-card__price shrink-0 text-right font-sans text-sm font-extrabold tabular-nums text-rh-text">{price}</div>}
        </div>
        {address && (
          <div className="rh-listing-card__address flex min-w-0 items-center gap-1.5 text-[13px] font-medium leading-snug text-rh-text-soft">
            <DesignIcon name="pin" size={13} />
            <span className="min-w-0 truncate">{address}</span>
          </div>
        )}
        {typeof rating === 'number' && (
          <RatingBadge value={rating} count={reviewCount} />
        )}
        {tags.length > 0 && (
          <div className="rh-listing-card__tags flex flex-wrap gap-1.5">
            {tags.map((tag, index) => (
              <DesignTag key={String(index)}>{tag}</DesignTag>
            ))}
          </div>
        )}
        {actions && <div className="rh-listing-card__actions flex flex-wrap gap-2 pt-1">{actions}</div>}
      </div>
    </article>
  )
}
