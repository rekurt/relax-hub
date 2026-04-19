import TopNavigationLayout from '@/components/TopNavigationLayout'
import BathhouseSelector from '@/components/BathhouseSelector'
import { OWNER_OVERFLOW_NAV_ITEMS, OWNER_PRIMARY_NAV_ITEMS } from '@/navigation/menu'

export default function AppLayout() {
  return (
    <TopNavigationLayout
      brandTitle="BANI PRO"
      brandSubtitle="Управление объектами и продажами"
      homeTo="/dashboard"
      primaryItems={OWNER_PRIMARY_NAV_ITEMS}
      overflowItems={OWNER_OVERFLOW_NAV_ITEMS}
      profilePath="/settings"
      headerAccessory={<BathhouseSelector />}
    />
  )
}
