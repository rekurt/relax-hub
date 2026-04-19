import { useEffect, useMemo } from 'react'
import { Select } from 'antd'
import { useLocation, useNavigate } from 'react-router-dom'
import { useGetCities } from '@/api/generated/cities/cities'
import { useAuthStore } from '@/stores/auth'
import TopNavigationLayout from '@/components/TopNavigationLayout'
import {
  CLIENT_OVERFLOW_NAV_ITEMS,
  CLIENT_PRIMARY_NAV_ITEMS,
  PUBLIC_OVERFLOW_NAV_ITEMS,
  PUBLIC_PRIMARY_NAV_ITEMS,
  getProfilePath,
} from '@/navigation/menu'

export default function PublicLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const user = useAuthStore((s) => s.user)
  const { data: citiesData } = useGetCities()
  const cities = citiesData?.data ?? []
  const searchParams = new URLSearchParams(location.search)
  const selectedCitySlug = searchParams.get('city_slug') ?? undefined
  const preferredCity = useMemo(
    () => cities.find((city) => city.slug === 'moscow' || city.slug === 'moskva' || city.name === 'Москва') ?? cities[0],
    [cities],
  )

  useEffect(() => {
    const titles: Array<[string, string]> = [
      ['/catalog', 'BANI — Каталог бань'],
      ['/bathhouses/', 'BANI — Баня и доступные слоты'],
      ['/checkout', 'BANI — Бронирование'],
      ['/certificates', 'BANI — Подарочные сертификаты'],
      ['/faq', 'BANI — FAQ'],
      ['/contacts', 'BANI — Контакты'],
    ]
    const matched = titles.find(([path]) => location.pathname === path || location.pathname.startsWith(path))
    document.title = matched?.[1] ?? 'BANI — Бронирование бань'
  }, [location.pathname])

  const cityAccessory = cities.length > 0 ? (
    <Select
      value={selectedCitySlug}
      allowClear
      placeholder={preferredCity?.name ?? 'Город'}
      style={{ minWidth: 140, maxWidth: 176 }}
      size="middle"
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
  ) : null

  return (
    <TopNavigationLayout
      brandTitle="BANI"
      brandSubtitle="Публичный каталог и бронирование"
      homeTo="/"
      primaryItems={user?.role === 'client' ? CLIENT_PRIMARY_NAV_ITEMS : PUBLIC_PRIMARY_NAV_ITEMS}
      overflowItems={user?.role === 'client' ? CLIENT_OVERFLOW_NAV_ITEMS : PUBLIC_OVERFLOW_NAV_ITEMS}
      profilePath={user ? getProfilePath(user.role) : undefined}
      headerAccessory={cityAccessory}
    />
  )
}
