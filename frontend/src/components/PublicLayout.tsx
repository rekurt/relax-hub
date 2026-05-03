import { useEffect, useMemo } from 'react'
import { Select } from '@/components/design/system'
import { DownOutlined, EnvironmentOutlined } from '@/components/design/icons'
import { useLocation, useNavigate } from 'react-router-dom'
import { useGetCities } from '@/api/generated/cities/cities'
import BrandLockup from '@/components/BrandLockup'
import { useAuthStore } from '@/stores/auth'
import ShellFooter from '@/components/ShellFooter'
import TopNavigationLayout from '@/components/TopNavigationLayout'
import { PLATFORM_NAME } from '@/content/support'
import {
  CLIENT_DRAWER_SECTIONS,
  CLIENT_PROFILE_MENU_ITEMS,
  CLIENT_PRIMARY_NAV_ITEMS,
  PUBLIC_DRAWER_SECTIONS,
  PUBLIC_PRIMARY_NAV_ITEMS,
  getProfilePath,
} from '@/navigation/menu'

export default function PublicLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const user = useAuthStore((s) => s.user)
  const { data: citiesData } = useGetCities()
  const cities = useMemo(() => citiesData?.data ?? [], [citiesData?.data])
  const isClientUser = user?.role === 'client'
  const searchParams = new URLSearchParams(location.search)
  const selectedCitySlug = searchParams.get('city_slug') ?? undefined
  const preferredCity = useMemo(
    () => cities.find((city) => city.slug === 'moscow' || city.slug === 'moskva' || city.name === 'Москва') ?? cities[0],
    [cities],
  )
  const selectedCityValue = selectedCitySlug

  useEffect(() => {
    const titles: Array<[string, string]> = [
      ['/catalog', `${PLATFORM_NAME} — Каталог бань`],
      ['/bathhouses/', `${PLATFORM_NAME} — Баня и доступные слоты`],
      ['/checkout', `${PLATFORM_NAME} — Бронирование`],
      ['/certificates', `${PLATFORM_NAME} — Подарочные сертификаты`],
      ['/faq', `${PLATFORM_NAME} — Помощь`],
      ['/contacts', `${PLATFORM_NAME} — Контакты`],
      ['/terms', `${PLATFORM_NAME} — Условия использования`],
    ]
    const matched = titles.find(([path]) => location.pathname === path || location.pathname.startsWith(path))
    document.title = matched?.[1] ?? `${PLATFORM_NAME} — Бронирование бань`
  }, [location.pathname])

  const cityAccessory = cities.length > 0 ? (
    <div className="rh-topnav__city-picker" aria-label="Выбор города">
      <EnvironmentOutlined className="rh-topnav__city-icon" aria-hidden />
      <Select
        value={selectedCityValue}
        className="rh-topnav__city-select"
        classNames={{ popup: { root: 'rh-topnav__city-dropdown' } }}
        placeholder={preferredCity?.name ?? 'Город'}
        suffixIcon={<DownOutlined />}
        variant="borderless"
        size="middle"
        allowClear
        options={cities.map((city) => ({
          label: city.name ?? 'Город',
          value: city.slug ?? '',
        }))}
        onChange={(slug) => {
          const nextParams = new URLSearchParams(location.search)
          if (slug) {
            nextParams.set('city_slug', slug)
          } else {
            nextParams.delete('city_slug')
          }
          const nextSearch = nextParams.toString()
          navigate(nextSearch ? `/catalog?${nextSearch}` : '/catalog')
        }}
      />
    </div>
  ) : null

  return (
    <TopNavigationLayout
      brandTitle={(
        <BrandLockup
          size="header"
          subtitle="Публичный каталог и бронирование"
          className="rh-topnav__brand-lockup"
        />
      )}
      brandSubtitle={null}
      brandAriaLabel={PLATFORM_NAME}
      surface={isClientUser ? 'client' : 'public'}
      homeTo="/"
      primaryItems={isClientUser ? CLIENT_PRIMARY_NAV_ITEMS : PUBLIC_PRIMARY_NAV_ITEMS}
      drawerSections={isClientUser ? CLIENT_DRAWER_SECTIONS : PUBLIC_DRAWER_SECTIONS}
      profileMenuItems={isClientUser ? CLIENT_PROFILE_MENU_ITEMS : undefined}
      profilePath={user ? getProfilePath(user.role) : undefined}
      headerAccessory={cityAccessory}
      headerAccessoryVariant="city"
      footer={<ShellFooter showClientSection={isClientUser} />}
    />
  )
}
