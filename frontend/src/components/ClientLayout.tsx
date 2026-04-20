import { useEffect, useState } from 'react'
import BrandLockup from '@/components/BrandLockup'
import OnboardingTour from '@/components/OnboardingTour'
import ShellFooter from '@/components/ShellFooter'
import TopNavigationLayout from '@/components/TopNavigationLayout'
import { PLATFORM_NAME } from '@/content/support'
import { useAuthStore } from '@/stores/auth'
import { CLIENT_DRAWER_SECTIONS, CLIENT_PRIMARY_NAV_ITEMS, CLIENT_PROFILE_MENU_ITEMS } from '@/navigation/menu'

export default function ClientLayout() {
  const [showTour, setShowTour] = useState(false)
  const { user, loadProfile } = useAuthStore()

  const shouldShowTour = !!user && !user.onboarding_completed

  useEffect(() => {
    if (!shouldShowTour) return
    const timer = setTimeout(() => setShowTour(true), 0)
    return () => clearTimeout(timer)
  }, [shouldShowTour])

  return (
    <>
      <TopNavigationLayout
        brandTitle={(
          <BrandLockup
            size="header"
            subtitle="Каталог, бронирование и личные поездки"
            className="bani-topnav__brand-lockup"
          />
        )}
        brandSubtitle={null}
        brandAriaLabel={PLATFORM_NAME}
        homeTo="/"
        primaryItems={CLIENT_PRIMARY_NAV_ITEMS}
        drawerSections={CLIENT_DRAWER_SECTIONS}
        profileMenuItems={CLIENT_PROFILE_MENU_ITEMS}
        profilePath="/client/profile"
        footer={<ShellFooter showClientSection />}
      />
      {showTour && (
        <OnboardingTour
          open={showTour}
          onComplete={async () => {
            setShowTour(false)
            await loadProfile()
          }}
          region={user?.region}
        />
      )}
    </>
  )
}
