import { useEffect, useState } from 'react'
import OnboardingTour from '@/components/OnboardingTour'
import TopNavigationLayout from '@/components/TopNavigationLayout'
import { useAuthStore } from '@/stores/auth'
import { CLIENT_OVERFLOW_NAV_ITEMS, CLIENT_PRIMARY_NAV_ITEMS } from '@/navigation/menu'

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
        brandTitle="BANI"
        brandSubtitle="Каталог, бронирование и личные поездки"
        homeTo="/"
        primaryItems={CLIENT_PRIMARY_NAV_ITEMS}
        overflowItems={CLIENT_OVERFLOW_NAV_ITEMS}
        profilePath="/client/profile"
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
