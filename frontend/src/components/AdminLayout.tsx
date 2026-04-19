import TopNavigationLayout from '@/components/TopNavigationLayout'
import { ADMIN_OVERFLOW_NAV_ITEMS, ADMIN_PRIMARY_NAV_ITEMS } from '@/navigation/menu'

export default function AdminLayout() {
  return (
    <TopNavigationLayout
      brandTitle="BANI ADMIN"
      brandSubtitle="Модерация, финансы и контроль платформы"
      homeTo="/admin"
      primaryItems={ADMIN_PRIMARY_NAV_ITEMS}
      overflowItems={ADMIN_OVERFLOW_NAV_ITEMS}
      profilePath="/admin/profile"
    />
  )
}
