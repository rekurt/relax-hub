import BrandLockup from '@/components/BrandLockup'
import TopNavigationLayout from '@/components/TopNavigationLayout'
import { PLATFORM_NAME } from '@/content/support'
import BathhouseSelector from '@/components/BathhouseSelector'
import { OWNER_OVERFLOW_NAV_ITEMS, OWNER_PRIMARY_NAV_ITEMS } from '@/navigation/menu'

export default function AppLayout() {
  return (
    <TopNavigationLayout
      brandTitle={(
        <BrandLockup
          size="header"
          subtitle="Управление объектами и продажами"
          className="bani-topnav__brand-lockup"
        />
      )}
      brandSubtitle={null}
      brandAriaLabel={PLATFORM_NAME}
      homeTo="/dashboard"
      primaryItems={OWNER_PRIMARY_NAV_ITEMS}
      overflowItems={OWNER_OVERFLOW_NAV_ITEMS}
      navigationMode="dropdown"
      profilePath="/settings"
      headerAccessory={<BathhouseSelector />}
    />
  )
}
