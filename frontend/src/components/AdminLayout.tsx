import BrandLockup from '@/components/BrandLockup'
import TopNavigationLayout from '@/components/TopNavigationLayout'
import { PLATFORM_NAME } from '@/content/support'
import { ADMIN_OVERFLOW_NAV_ITEMS, ADMIN_PRIMARY_NAV_ITEMS } from '@/navigation/menu'

export default function AdminLayout() {
  return (
    <TopNavigationLayout
      brandTitle={(
        <BrandLockup
          size="header"
          subtitle="Модерация, финансы и контроль платформы"
          className="bani-topnav__brand-lockup"
        />
      )}
      brandSubtitle={null}
      brandAriaLabel={PLATFORM_NAME}
      homeTo="/admin"
      primaryItems={ADMIN_PRIMARY_NAV_ITEMS}
      overflowItems={ADMIN_OVERFLOW_NAV_ITEMS}
      navigationMode="dropdown"
      profilePath="/admin/profile"
    />
  )
}
