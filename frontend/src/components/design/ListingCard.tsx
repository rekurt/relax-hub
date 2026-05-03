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
      className={`rh-listing-card ${className ?? ''}`.trim()}
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
      <div className="rh-listing-card__media">
        {imageUrl ? (
          <img className="rh-listing-card__image" src={imageUrl} alt={imageAlt ?? ''} />
        ) : (
          <PhotoPlaceholder className="rh-listing-card__placeholder" />
        )}
        {badge && (
          <DesignTag tone={badgeTone} className="rh-listing-card__badge">
            {badge}
          </DesignTag>
        )}
      </div>
      <div className="rh-listing-card__body">
        <div className="rh-listing-card__head">
          <h3 className="rh-listing-card__title">
            {name}
            {verified && <DesignIcon name="checkc" size={14} className="rh-listing-card__verified" />}
          </h3>
          {price && <div className="rh-listing-card__price">{price}</div>}
        </div>
        {address && (
          <div className="rh-listing-card__address">
            <DesignIcon name="pin" size={13} />
            <span>{address}</span>
          </div>
        )}
        <RatingBadge value={rating} count={reviewCount} />
        {tags.length > 0 && (
          <div className="rh-listing-card__tags">
            {tags.map((tag, index) => (
              <DesignTag key={String(index)}>{tag}</DesignTag>
            ))}
          </div>
        )}
        {actions && <div className="rh-listing-card__actions">{actions}</div>}
      </div>
    </article>
  )
}
