import { useState, useCallback, useEffect, useRef, useMemo } from 'react'
import {
  Typography,
  Input,
  Select,
  Slider,
  Checkbox,
  Row,
  Col,
  Pagination,
  Spin,
  Card,
  Space,
  Button,
  InputNumber,
  Collapse,
  Badge,
  Segmented,
  App,
  Tag,
  DatePicker,
  TimePicker,
} from 'antd'
import {
  SearchOutlined,
  EnvironmentOutlined,
  FilterOutlined,
  SortAscendingOutlined,
  SwapOutlined,
  UnorderedListOutlined,
  EnvironmentFilled,
  SplitCellsOutlined,
  CarOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons'
import { isAxiosError } from 'axios'
import { useNavigate, useSearchParams } from 'react-router-dom'
import dayjs from 'dayjs'
import { useGetBathhouses } from '@/api/generated/bathhouses/bathhouses'
import { useGetCities } from '@/api/generated/cities/cities'
import type { GetBathhousesParams } from '@/api/generated/model'
import BathhouseCard from '@/components/BathhouseCard'
import BathhouseMap from '@/components/BathhouseMap'
import SearchSuggestions from '@/components/SearchSuggestions'
import { formatPrice } from '@/lib/format'
import { axiosInstance } from '@/api/axios-instance'
import PublicState from '@/components/PublicState'
import { useAuthStore } from '@/stores/auth'
import { PUBLIC_SHORTCUT_CARDS } from '@/navigation/menu'

const { Title, Text } = Typography

type ViewMode = 'list' | 'map' | 'split'
type GeoMode = 'radius' | 'travel_time'
type TravelMode = 'car' | 'transit'
type SortOption = {
  value: string
  label: string
  sortBy: string
  sortOrder: string
}

const SORT_OPTIONS: SortOption[] = [
  { value: 'rating_desc', label: 'Сначала лучшие', sortBy: 'rating', sortOrder: 'desc' },
  { value: 'price_asc', label: 'Сначала дешёвые', sortBy: 'price', sortOrder: 'asc' },
  { value: 'price_desc', label: 'Сначала дорогие', sortBy: 'price', sortOrder: 'desc' },
  { value: 'newest_desc', label: 'Новые', sortBy: 'newest', sortOrder: 'desc' },
  { value: 'distance_asc', label: 'Ближайшие', sortBy: 'distance', sortOrder: 'asc' },
]

const DEFAULT_SORT_VALUE = SORT_OPTIONS[0]!.value
const DEFAULT_VIEW_MODE: ViewMode = 'list'

const VIEW_MODE_OPTIONS = [
  { value: 'list', icon: <UnorderedListOutlined />, label: 'Список' },
  { value: 'split', icon: <SplitCellsOutlined />, label: 'Сплит' },
  { value: 'map', icon: <EnvironmentFilled />, label: 'Карта' },
]

const TRAVEL_TIME_OPTIONS = [
  { value: 15, label: '15 мин' },
  { value: 30, label: '30 мин' },
  { value: 45, label: '45 мин' },
  { value: 60, label: '60 мин' },
]

const TRAVEL_MODE_OPTIONS = [
  { value: 'car', label: 'На машине', icon: <CarOutlined /> },
  { value: 'transit', label: 'Пешком/транспорт', icon: <ClockCircleOutlined /> },
]

const PRICE_FILTER_MAX = 2_500_000
const MAX_COMPARE = 3

interface IsochroneData {
  coordinates: number[][]
  wkt: string
}

interface CatalogUrlState {
  page: number
  search: string
  sortValue: string
  viewMode: ViewMode
  filters: Partial<GetBathhousesParams>
  availableDate: dayjs.Dayjs | null
  availableTimeFrom: dayjs.Dayjs | null
  availableTimeTo: dayjs.Dayjs | null
  openNow: boolean
  minRating?: number
  geoEnabled: boolean
  geoCoords: { lat: number; lng: number } | null
  geoMode: GeoMode
  travelMode: TravelMode
  travelMinutes: number
}

function parseSortOption(value: string): { sort_by: string; sort_order: string } {
  const option = SORT_OPTIONS.find((item) => item.value === value) ?? SORT_OPTIONS[0]!
  return {
    sort_by: option.sortBy,
    sort_order: option.sortOrder,
  }
}

function parseBooleanParam(value: string | null) {
  if (value !== 'true') return undefined
  return true
}

function parseNumberParam(value: string | null) {
  if (!value) return undefined
  const nextValue = Number(value)
  return Number.isFinite(nextValue) ? nextValue : undefined
}

function pluralizeGuestCount(count: number) {
  const mod10 = count % 10
  const mod100 = count % 100
  if (mod10 === 1 && mod100 !== 11) return `${count} гость`
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return `${count} гостя`
  return `${count} гостей`
}

function buildCatalogSearchParams({
  search,
  sortValue,
  viewMode,
  page,
  filters,
  availableDate,
  availableTimeFrom,
  availableTimeTo,
  openNow,
  geoEnabled,
  geoCoords,
  geoMode,
  travelMode,
  travelMinutes,
  minRating,
}: {
  search: string
  sortValue: string
  viewMode: ViewMode
  page: number
  filters: Partial<GetBathhousesParams>
  availableDate: dayjs.Dayjs | null
  availableTimeFrom: dayjs.Dayjs | null
  availableTimeTo: dayjs.Dayjs | null
  openNow: boolean
  geoEnabled: boolean
  geoCoords: { lat: number; lng: number } | null
  geoMode: GeoMode
  travelMode: TravelMode
  travelMinutes: number
  minRating?: number
}) {
  const params = new URLSearchParams()
  const normalizedSearch = search.trim()

  if (normalizedSearch) params.set('q', normalizedSearch)
  if (filters.city_slug) params.set('city_slug', String(filters.city_slug))
  if (filters.guest_count) params.set('guest_count', String(filters.guest_count))
  if (filters.price_min) params.set('price_min', String(filters.price_min))
  if (filters.price_max) params.set('price_max', String(filters.price_max))
  if (filters.has_sauna) params.set('has_sauna', 'true')
  if (filters.has_steam_room) params.set('has_steam_room', 'true')
  if (filters.has_pool) params.set('has_pool', 'true')
  if (filters.has_hot_tub) params.set('has_hot_tub', 'true')
  if (filters.has_bbq) params.set('has_bbq', 'true')
  if (filters.has_karaoke) params.set('has_karaoke', 'true')
  if (openNow) params.set('open_now', 'true')
  if (availableDate) params.set('available_date', availableDate.format('YYYY-MM-DD'))
  if (availableTimeFrom) params.set('available_time_from', availableTimeFrom.format('HH:mm'))
  if (availableTimeTo) params.set('available_time_to', availableTimeTo.format('HH:mm'))
  if (minRating != null) params.set('min_rating', String(minRating))

  if (geoEnabled && geoCoords) {
    params.set('lat', String(geoCoords.lat))
    params.set('lng', String(geoCoords.lng))
    params.set('geo_mode', geoMode)
    if (geoMode === 'radius') {
      params.set('radius_km', String(filters.radius_km ?? 10))
    } else {
      params.set('travel_mode', travelMode)
      params.set('travel_minutes', String(travelMinutes))
    }
  }

  if (sortValue !== DEFAULT_SORT_VALUE) params.set('sort', sortValue)
  if (viewMode !== DEFAULT_VIEW_MODE) params.set('view', viewMode)
  if (page > 1) params.set('page', String(page))

  return params
}

function getFilterButtonClass(isActive: boolean) {
  return isActive ? 'bani-catalog__quick-filter bani-catalog__quick-filter--active' : 'bani-catalog__quick-filter'
}

function parseCatalogUrlState(searchParamsString: string): CatalogUrlState {
  const nextSearchParams = new URLSearchParams(searchParamsString)
  const nextSearch = nextSearchParams.get('q') ?? ''
  const nextLat = parseNumberParam(nextSearchParams.get('lat'))
  const nextLng = parseNumberParam(nextSearchParams.get('lng'))
  const hasGeoCoords = nextLat != null && nextLng != null
  const nextView = nextSearchParams.get('view')
  const requestedSortValue = nextSearchParams.get('sort') ?? DEFAULT_SORT_VALUE

  return {
    page: parseNumberParam(nextSearchParams.get('page')) ?? 1,
    search: nextSearch,
    sortValue: requestedSortValue === 'distance_asc' && !hasGeoCoords ? DEFAULT_SORT_VALUE : requestedSortValue,
    viewMode: nextView === 'map' || nextView === 'split' ? nextView : DEFAULT_VIEW_MODE,
    availableDate: nextSearchParams.get('available_date') ? dayjs(nextSearchParams.get('available_date')) : null,
    availableTimeFrom: nextSearchParams.get('available_time_from')
      ? dayjs(nextSearchParams.get('available_time_from'), 'HH:mm')
      : null,
    availableTimeTo: nextSearchParams.get('available_time_to')
      ? dayjs(nextSearchParams.get('available_time_to'), 'HH:mm')
      : null,
    openNow: nextSearchParams.get('open_now') === 'true',
    minRating: parseNumberParam(nextSearchParams.get('min_rating')),
    geoEnabled: hasGeoCoords,
    geoCoords: hasGeoCoords ? { lat: nextLat ?? 0, lng: nextLng ?? 0 } : null,
    geoMode: nextSearchParams.get('geo_mode') === 'travel_time' ? 'travel_time' : 'radius',
    travelMode: nextSearchParams.get('travel_mode') === 'transit' ? 'transit' : 'car',
    travelMinutes: parseNumberParam(nextSearchParams.get('travel_minutes')) ?? 15,
    filters: {
      city_slug: nextSearchParams.get('city_slug') ?? undefined,
      guest_count: parseNumberParam(nextSearchParams.get('guest_count')),
      price_min: parseNumberParam(nextSearchParams.get('price_min')),
      price_max: parseNumberParam(nextSearchParams.get('price_max')),
      radius_km: parseNumberParam(nextSearchParams.get('radius_km')),
      has_sauna: parseBooleanParam(nextSearchParams.get('has_sauna')),
      has_steam_room: parseBooleanParam(nextSearchParams.get('has_steam_room')),
      has_hot_tub: parseBooleanParam(nextSearchParams.get('has_hot_tub')),
      has_pool: parseBooleanParam(nextSearchParams.get('has_pool')),
      has_bbq: parseBooleanParam(nextSearchParams.get('has_bbq')),
      has_karaoke: parseBooleanParam(nextSearchParams.get('has_karaoke')),
    },
  }
}

export default function BathhouseSearch() {
  const [searchParams, setSearchParams] = useSearchParams()
  const searchParamsString = searchParams.toString()
  const [initialCatalogState] = useState(() => parseCatalogUrlState(searchParamsString))
  const [page, setPage] = useState(initialCatalogState.page)
  const [pageSize] = useState(12)
  const [search, setSearch] = useState(initialCatalogState.search)
  const [sortValue, setSortValue] = useState(initialCatalogState.sortValue)
  const [filters, setFilters] = useState<Partial<GetBathhousesParams>>(initialCatalogState.filters)
  const [geoEnabled, setGeoEnabled] = useState(initialCatalogState.geoEnabled)
  const [geoCoords, setGeoCoords] = useState<{ lat: number; lng: number } | null>(initialCatalogState.geoCoords)
  const [viewMode, setViewMode] = useState<ViewMode>(initialCatalogState.viewMode)
  const [compareIds, setCompareIds] = useState<string[]>([])
  const [highlightedId, setHighlightedId] = useState<string | null>(null)
  const [debouncedSearch, setDebouncedSearch] = useState(initialCatalogState.search)
  const [showSuggestions, setShowSuggestions] = useState(false)
  const [availableDate, setAvailableDate] = useState<dayjs.Dayjs | null>(initialCatalogState.availableDate)
  const [availableTimeFrom, setAvailableTimeFrom] = useState<dayjs.Dayjs | null>(initialCatalogState.availableTimeFrom)
  const [availableTimeTo, setAvailableTimeTo] = useState<dayjs.Dayjs | null>(initialCatalogState.availableTimeTo)
  const [openNow, setOpenNow] = useState(initialCatalogState.openNow)
  const [minRating, setMinRating] = useState<number | undefined>(initialCatalogState.minRating)
  const searchWrapperRef = useRef<HTMLDivElement>(null)
  const debounceTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const hasHydratedFromUrlRef = useRef(false)
  const skipNextUrlWriteRef = useRef(false)

  useEffect(() => {
    debounceTimer.current = setTimeout(() => {
      setDebouncedSearch(search)
    }, 300)
    return () => {
      if (debounceTimer.current) clearTimeout(debounceTimer.current)
    }
  }, [search])

  const [geoMode, setGeoMode] = useState<GeoMode>(initialCatalogState.geoMode)
  const [travelMode, setTravelMode] = useState<TravelMode>(initialCatalogState.travelMode)
  const [travelMinutes, setTravelMinutes] = useState<number>(initialCatalogState.travelMinutes)
  const [isochroneData, setIsochroneData] = useState<IsochroneData | null>(null)
  const [isochroneLoading, setIsochroneLoading] = useState(false)
  const [geoError, setGeoError] = useState<string | null>(null)

  const applyCatalogState = useCallback((nextState: CatalogUrlState) => {
    setSearch(nextState.search)
    setDebouncedSearch(nextState.search)
    setSortValue(nextState.sortValue)
    setViewMode(nextState.viewMode)
    setPage(nextState.page)
    setAvailableDate(nextState.availableDate)
    setAvailableTimeFrom(nextState.availableTimeFrom)
    setAvailableTimeTo(nextState.availableTimeTo)
    setOpenNow(nextState.openNow)
    setMinRating(nextState.minRating)
    setGeoEnabled(nextState.geoEnabled)
    setGeoCoords(nextState.geoCoords)
    setGeoMode(nextState.geoMode)
    setTravelMode(nextState.travelMode)
    setTravelMinutes(nextState.travelMinutes)
    setFilters(nextState.filters)
  }, [])

  const navigate = useNavigate()
  const { message } = App.useApp()
  const currentUser = useAuthStore((s) => s.user)
  const {
    data: citiesData,
    isError: isCitiesError,
    error: citiesError,
    refetch: refetchCities,
  } = useGetCities()
  const cities = useMemo(() => citiesData?.data ?? [], [citiesData?.data])
  const canCompare = currentUser?.role === 'client'
  const cityLabelBySlug = useMemo(
    () => new Map(cities.map((city) => [city.slug ?? String(city.id), city.name ?? 'Город'])),
    [cities],
  )

  useEffect(() => {
    const nextState = parseCatalogUrlState(searchParamsString)
    if (hasHydratedFromUrlRef.current) {
      skipNextUrlWriteRef.current = true
    }
    hasHydratedFromUrlRef.current = true

    applyCatalogState(nextState)
  }, [applyCatalogState, searchParamsString])

  useEffect(() => {
    if (skipNextUrlWriteRef.current) {
      skipNextUrlWriteRef.current = false
      return
    }

    const nextParams = buildCatalogSearchParams({
      search,
      sortValue,
      viewMode,
      page,
      filters,
      availableDate,
      availableTimeFrom,
      availableTimeTo,
      openNow,
      geoEnabled,
      geoCoords,
      geoMode,
      travelMode,
      travelMinutes,
      minRating,
    })
    const nextSearchParamsString = nextParams.toString()
    if (nextSearchParamsString !== searchParamsString) {
      setSearchParams(nextParams, { replace: true })
    }
  }, [
    availableDate,
    availableTimeFrom,
    availableTimeTo,
    filters,
    geoCoords,
    geoEnabled,
    geoMode,
    minRating,
    openNow,
    page,
    search,
    searchParamsString,
    setSearchParams,
    sortValue,
    travelMinutes,
    travelMode,
    viewMode,
  ])

  useEffect(() => {
    if (!geoEnabled && sortValue === 'distance_asc') {
      setSortValue(DEFAULT_SORT_VALUE)
    }
  }, [geoEnabled, sortValue])

  const { sort_by, sort_order } = parseSortOption(sortValue)
  const selectedSortLabel = SORT_OPTIONS.find((option) => option.value === sortValue)?.label ?? SORT_OPTIONS[0]!.label
  const sortOptions = useMemo(
    () => SORT_OPTIONS.map((option) => (option.value === 'distance_asc' ? { ...option, disabled: !geoEnabled } : option)),
    [geoEnabled],
  )
  const { radius_km, ...apiFilters } = filters

  const queryParams: GetBathhousesParams = {
    page,
    page_size: pageSize,
    sort_by,
    sort_order,
    q: debouncedSearch || undefined,
    min_rating: minRating,
    open_now: openNow || undefined,
    available_date: availableDate ? availableDate.format('YYYY-MM-DD') : undefined,
    available_time_from: availableTimeFrom ? availableTimeFrom.format('HH:mm') : undefined,
    available_time_to: availableTimeTo ? availableTimeTo.format('HH:mm') : undefined,
    ...apiFilters,
    ...(geoEnabled && geoCoords && geoMode === 'radius'
      ? { lat: geoCoords.lat, lng: geoCoords.lng, radius_km: radius_km ?? 10 }
      : {}),
    ...(geoEnabled && geoCoords && geoMode === 'travel_time' && isochroneData
      ? { lat: geoCoords.lat, lng: geoCoords.lng, isochrone_wkt: isochroneData.wkt }
      : {}),
  }

  const {
    data,
    isLoading,
    isError,
    error,
    refetch,
  } = useGetBathhouses(queryParams as GetBathhousesParams)
  const bathhouses = data?.data ?? []
  const meta = data?.meta

  const fetchIsochrone = useCallback(async (lat: number, lng: number, mode: TravelMode, minutes: number) => {
    setIsochroneLoading(true)
    try {
      const resp = await axiosInstance.get('/isochrone', {
        params: { lat, lon: lng, mode, minutes },
      })
      const result = resp.data?.data ?? resp.data
      if (result?.coordinates && result?.wkt) {
        setIsochroneData({ coordinates: result.coordinates, wkt: result.wkt })
        setGeoError(null)
      }
    } catch {
      message.error('Не удалось построить зону доступности')
      setIsochroneData(null)
      setGeoError('Не удалось построить зону доступности. Показываем общую выдачу без зоны поездки.')
    } finally {
      setIsochroneLoading(false)
    }
  }, [message])

  useEffect(() => {
    if (!geoEnabled || !geoCoords || geoMode !== 'travel_time') {
      setIsochroneData(null)
      setGeoError(null)
      return
    }

    void fetchIsochrone(geoCoords.lat, geoCoords.lng, travelMode, travelMinutes)
  }, [fetchIsochrone, geoCoords, geoEnabled, geoMode, travelMinutes, travelMode])

  const handleGeoSearch = () => {
    if (!navigator.geolocation) return
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setGeoCoords({ lat: pos.coords.latitude, lng: pos.coords.longitude })
        setGeoEnabled(true)
        setPage(1)
      },
      () => {/* geolocation denied - ignore */},
    )
  }

  const handleGeoModeChange = (mode: GeoMode) => {
    setGeoMode(mode)
    setPage(1)
  }

  const handleTravelModeChange = (mode: TravelMode) => {
    setTravelMode(mode)
    setPage(1)
  }

  const handleTravelMinutesChange = (minutes: number) => {
    setTravelMinutes(minutes)
    setPage(1)
  }

  const updateFilter = (key: keyof GetBathhousesParams, value: unknown) => {
    setFilters((prev) => ({ ...prev, [key]: value || undefined }))
    setPage(1)
  }

  const handleCompareToggle = (id: string) => {
    setCompareIds((prev) => {
      if (prev.includes(id)) return prev.filter((x) => x !== id)
      if (prev.length >= MAX_COMPARE) {
        message.warning(`Можно сравнить не более ${MAX_COMPARE} бань`)
        return prev
      }
      return [...prev, id]
    })
  }

  const handleGoCompare = () => {
    if (!canCompare) return
    if (compareIds.length < 2) {
      message.warning('Выберите минимум 2 бани для сравнения')
      return
    }
    navigate(`/client/comparison?ids=${compareIds.join(',')}`)
  }

  const handleBoundsChange = (bounds: { north: number; south: number; east: number; west: number }) => {
    const nextRadius = Math.max(
      1,
      Math.round(
        haversineDistance(
          bounds.south,
          bounds.west,
          bounds.north,
          bounds.east,
        ) / 2,
      ),
    )

    setGeoCoords({
      lat: (bounds.north + bounds.south) / 2,
      lng: (bounds.east + bounds.west) / 2,
    })
    setGeoEnabled(true)
    setGeoMode('radius')
    setFilters((prev) => ({ ...prev, radius_km: nextRadius }))
    setPage(1)
  }

  const handleMarkerClick = (id: string) => {
    const found = bathhouses.find((bathhouse) => bathhouse.id === id)
    navigate(`/bathhouses/${found?.slug ?? id}`)
  }

  const isochronePolygon: number[][] | null = isochroneData?.coordinates
    ? isochroneData.coordinates.map((coord: number[]) => [coord[1] ?? 0, coord[0] ?? 0] as [number, number])
    : null

  const activeFilters = useMemo(() => {
    const items: Array<{ key: string; label: string; onRemove: () => void }> = []

    if (search.trim()) {
      items.push({
        key: 'q',
        label: `Поиск: ${search.trim()}`,
        onRemove: () => {
          setSearch('')
          setDebouncedSearch('')
          setPage(1)
        },
      })
    }

    if (filters.city_slug) {
      items.push({
        key: 'city_slug',
        label: cityLabelBySlug.get(String(filters.city_slug)) ?? 'Город',
        onRemove: () => updateFilter('city_slug', undefined),
      })
    }

    if (filters.guest_count) {
      items.push({
        key: 'guest_count',
        label: pluralizeGuestCount(filters.guest_count),
        onRemove: () => updateFilter('guest_count', undefined),
      })
    }

    if (openNow) {
      items.push({
        key: 'open_now',
        label: 'Открыто сейчас',
        onRemove: () => {
          setOpenNow(false)
          setPage(1)
        },
      })
    }

    if (availableDate) {
      items.push({
        key: 'available_date',
        label: `Дата: ${availableDate.format('DD.MM')}`,
        onRemove: () => {
          setAvailableDate(null)
          setPage(1)
        },
      })
    }

    if (availableTimeFrom || availableTimeTo) {
      items.push({
        key: 'time_range',
        label: `Время: ${availableTimeFrom?.format('HH:mm') ?? 'любое'}-${availableTimeTo?.format('HH:mm') ?? 'любое'}`,
        onRemove: () => {
          setAvailableTimeFrom(null)
          setAvailableTimeTo(null)
          setPage(1)
        },
      })
    }

    if (filters.price_min != null || filters.price_max != null) {
      const rangeParts = []
      if (filters.price_min != null) rangeParts.push(`от ${formatPrice(filters.price_min)}`)
      if (filters.price_max != null) rangeParts.push(`до ${formatPrice(filters.price_max)}`)
      items.push({
        key: 'price_range',
        label: `Цена ${rangeParts.join(' ')}`,
        onRemove: () => {
          setFilters((prev) => ({ ...prev, price_min: undefined, price_max: undefined }))
          setPage(1)
        },
      })
    }

    if (filters.has_hot_tub) {
      items.push({
        key: 'has_hot_tub',
        label: 'С чаном',
        onRemove: () => updateFilter('has_hot_tub', undefined),
      })
    }

    if (filters.has_pool) {
      items.push({
        key: 'has_pool',
        label: 'С бассейном',
        onRemove: () => updateFilter('has_pool', undefined),
      })
    }

    if (filters.has_sauna) {
      items.push({
        key: 'has_sauna',
        label: 'С сауной',
        onRemove: () => updateFilter('has_sauna', undefined),
      })
    }

    if (filters.has_steam_room) {
      items.push({
        key: 'has_steam_room',
        label: 'С парной',
        onRemove: () => updateFilter('has_steam_room', undefined),
      })
    }

    if (filters.has_bbq) {
      items.push({
        key: 'has_bbq',
        label: 'С мангалом',
        onRemove: () => updateFilter('has_bbq', undefined),
      })
    }

    if (filters.has_karaoke) {
      items.push({
        key: 'has_karaoke',
        label: 'С караоке',
        onRemove: () => updateFilter('has_karaoke', undefined),
      })
    }

    if (geoEnabled && geoCoords) {
      const geoLabel = geoMode === 'travel_time'
        ? `${travelMinutes} мин ${travelMode === 'car' ? 'на машине' : 'пешком/транспортом'}`
        : `Рядом: ${radius_km ?? 10} км`

      items.push({
        key: 'geo',
        label: geoLabel,
        onRemove: () => {
          setGeoEnabled(false)
          setGeoCoords(null)
          setGeoMode('radius')
          setIsochroneData(null)
          setGeoError(null)
          setPage(1)
        },
      })
    }

    if (minRating != null) {
      items.push({
        key: 'min_rating',
        label: `Рейтинг от ${minRating.toFixed(1)}`,
        onRemove: () => {
          setMinRating(undefined)
          setPage(1)
        },
      })
    }

    return items
  }, [
    availableDate,
    availableTimeFrom,
    availableTimeTo,
    cityLabelBySlug,
    filters,
    geoCoords,
    geoEnabled,
    geoMode,
    minRating,
    openNow,
    radius_km,
    search,
    travelMinutes,
    travelMode,
  ])

  const resetDiscoveryFilters = useCallback(() => {
    setSearch('')
    setDebouncedSearch('')
    setShowSuggestions(false)
    setSortValue(DEFAULT_SORT_VALUE)
    setViewMode(DEFAULT_VIEW_MODE)
    setFilters({})
    setAvailableDate(null)
    setAvailableTimeFrom(null)
    setAvailableTimeTo(null)
    setOpenNow(false)
    setMinRating(undefined)
    setGeoEnabled(false)
    setGeoCoords(null)
    setGeoMode('radius')
    setTravelMode('car')
    setTravelMinutes(15)
    setIsochroneData(null)
    setGeoError(null)
    setPage(1)
  }, [])

  const handleShortcutSelect = useCallback((params: Record<string, string>) => {
    const nextParams = new URLSearchParams(params)
    skipNextUrlWriteRef.current = true
    applyCatalogState(parseCatalogUrlState(nextParams.toString()))
    setSearchParams(nextParams)
  }, [applyCatalogState, setSearchParams])

  const activeShortcutKey = PUBLIC_SHORTCUT_CARDS.find((card) => {
    const [cardPath, cardQuery] = card.to.split('?')
    if (cardPath !== '/catalog') return false
    const cardParams = new URLSearchParams(cardQuery ?? '')
    return Array.from(cardParams.entries()).every(([name, value]) => searchParams.get(name) === value)
  })?.key

  const isSharedCatalogFailure = isError && isCitiesError
  const catalogErrorStatus = getQueryErrorStatus(error)
  const citiesErrorStatus = getQueryErrorStatus(citiesError)
  const shouldShowServiceUnavailableState = isSharedCatalogFailure && (
    isLikelyServiceUnavailable(error) ||
    isLikelyServiceUnavailable(citiesError) ||
    (
      (catalogErrorStatus == null || catalogErrorStatus >= 500) &&
      (citiesErrorStatus == null || citiesErrorStatus >= 500)
    )
  )
  const isInvalidFilterRequest = catalogErrorStatus === 400
  const errorTitle = shouldShowServiceUnavailableState
    ? 'Каталог временно недоступен'
    : isInvalidFilterRequest
      ? 'Часть параметров не удалось применить'
      : 'Не удалось загрузить варианты'
  const errorDescription = shouldShowServiceUnavailableState
    ? 'Сервис подбора не отвечает. Повторите попытку через несколько секунд.'
    : isInvalidFilterRequest
      ? 'Сбросьте часть фильтров и повторите поиск.'
      : 'Повторите попытку или скорректируйте параметры поиска.'

  const listContent = (
    <>
      {geoError && (
        <div style={{ marginBottom: 16 }}>
          <PublicState
            kind="degraded"
            compact
            title="Зона доступности временно недоступна"
            description={geoError}
          />
        </div>
      )}
      {isLoading ? (
        <PublicState
          kind="loading"
          title="Загружаем варианты"
          description="Подбираем бани по вашим параметрам."
        />
      ) : isError ? (
        <PublicState
          kind="error"
          title={errorTitle}
          description={errorDescription}
          actionText="Повторить"
          onAction={() => {
            void refetch()
            void refetchCities()
          }}
          secondaryActionText={activeFilters.length > 0 ? 'Сбросить фильтры' : undefined}
          onSecondaryAction={activeFilters.length > 0 ? resetDiscoveryFilters : undefined}
        />
      ) : bathhouses.length === 0 ? (
        <PublicState
          kind="empty"
          title="Ничего не найдено"
          description="Попробуйте изменить параметры поиска или сбросить фильтры."
          actionText="Сбросить фильтры"
          onAction={resetDiscoveryFilters}
        />
      ) : (
        <Spin spinning={isochroneLoading}>
          <>
            <Row gutter={[16, 16]}>
              {bathhouses.map((bathhouse) => (
                <Col
                  key={bathhouse.id}
                  xs={24}
                  sm={viewMode === 'split' ? 24 : 12}
                  md={viewMode === 'split' ? 24 : 8}
                  lg={viewMode === 'split' ? 12 : 6}
                  onMouseEnter={() => setHighlightedId(bathhouse.id ?? null)}
                  onMouseLeave={() => setHighlightedId(null)}
                >
                  <div
                    className={highlightedId === bathhouse.id ? 'bani-catalog__card-shell bani-catalog__card-shell--highlighted' : 'bani-catalog__card-shell'}
                    data-testid={`card-wrapper-${bathhouse.id}`}
                  >
                    <BathhouseCard
                      bathhouse={bathhouse}
                      showCompare={canCompare}
                      isCompareSelected={compareIds.includes(bathhouse.id ?? '')}
                      onCompareToggle={handleCompareToggle}
                    />
                  </div>
                </Col>
              ))}
            </Row>
            {meta && meta.total_pages && meta.total_pages > 1 && (
              <div style={{ textAlign: 'center', marginTop: 24 }}>
                <Pagination
                  current={page}
                  pageSize={pageSize}
                  total={meta.total_count}
                  onChange={(nextPage) => setPage(nextPage)}
                  showSizeChanger={false}
                />
              </div>
            )}
          </>
        </Spin>
      )}
    </>
  )

  return (
    <div className="bani-catalog">
      <section className="bani-catalog__hero">
        <div className="bani-catalog__hero-copy">
          <Text className="bani-catalog__eyebrow">Публичный каталог</Text>
          <Title level={2} className="bani-catalog__title">Поиск бань</Title>
          <Typography.Text className="bani-catalog__description">
            Каталог остаётся list-first: сначала понятная выдача, затем карта и сплит для уточнения. Основные фильтры вынесены наверх, активные параметры всегда видны в URL и на экране.
          </Typography.Text>
        </div>
        <Segmented
          options={VIEW_MODE_OPTIONS}
          value={viewMode}
          onChange={(value) => setViewMode(value as ViewMode)}
        />
      </section>

      <section className="bani-catalog__shortcut-row" aria-label="Сценарии подбора">
        {PUBLIC_SHORTCUT_CARDS.map((card) => (
          <button
            key={card.key}
            type="button"
            className={activeShortcutKey === card.key ? 'bani-catalog__shortcut bani-catalog__shortcut--active' : 'bani-catalog__shortcut'}
            onClick={() => handleShortcutSelect(card.params)}
          >
            <span className="bani-catalog__shortcut-title">{card.title}</span>
            <span className="bani-catalog__shortcut-description">{card.description}</span>
          </button>
        ))}
      </section>

      <Space orientation="vertical" size="middle" style={{ width: '100%', marginBottom: 24 }}>
        <Card variant="borderless" className="bani-catalog__filter-card">
          <Row gutter={[16, 16]} align="middle">
            <Col xs={24} lg={10}>
              <div ref={searchWrapperRef} style={{ position: 'relative' }}>
                <Input
                  placeholder="Поиск по названию..."
                  prefix={<SearchOutlined />}
                  value={search}
                  onChange={(event) => {
                    setSearch(event.target.value)
                    setPage(1)
                    setShowSuggestions(true)
                  }}
                  onFocus={() => {
                    if (search.length >= 2) setShowSuggestions(true)
                  }}
                  allowClear
                />
                <SearchSuggestions
                  query={search}
                  visible={showSuggestions}
                  onSelect={(text) => {
                    setSearch(text)
                    setDebouncedSearch(text)
                    setShowSuggestions(false)
                    setPage(1)
                  }}
                  onClose={() => setShowSuggestions(false)}
                />
              </div>
            </Col>
            <Col xs={12} lg={4}>
              <Select
                placeholder="Город"
                style={{ width: '100%' }}
                allowClear
                value={filters.city_slug as string | undefined}
                onChange={(slug) => updateFilter('city_slug', slug)}
                options={cities.map((city) => ({ value: city.slug ?? String(city.id), label: city.name }))}
              />
            </Col>
            <Col xs={12} lg={3}>
              <InputNumber
                min={1}
                max={50}
                placeholder="Гости"
                style={{ width: '100%' }}
                value={filters.guest_count}
                onChange={(value) => updateFilter('guest_count', value)}
              />
            </Col>
            <Col xs={12} lg={3}>
              <DatePicker
                style={{ width: '100%' }}
                placeholder="Дата"
                value={availableDate}
                onChange={(dateValue) => {
                  setAvailableDate(dateValue)
                  setPage(1)
                }}
                disabledDate={(current) => current && current.isBefore(dayjs(), 'day')}
                data-testid="date-filter"
              />
            </Col>
            <Col xs={12} lg={4}>
              <Select
                style={{ width: '100%' }}
                value={sortValue}
                onChange={setSortValue}
                options={sortOptions}
                suffixIcon={<SortAscendingOutlined />}
              />
            </Col>
          </Row>

          <div className="bani-catalog__quick-filters">
            <Text className="bani-catalog__quick-filters-label">Быстрые фильтры</Text>
            <Space wrap size={[8, 8]}>
              <Button
                size="small"
                type="text"
                className={getFilterButtonClass(openNow)}
                onClick={() => {
                  setOpenNow((prev) => !prev)
                  setPage(1)
                }}
              >
                Открыто сейчас
              </Button>
              <Button
                size="small"
                type="text"
                className={getFilterButtonClass(Boolean(filters.has_pool))}
                onClick={() => updateFilter('has_pool', !filters.has_pool)}
              >
                С бассейном
              </Button>
              <Button
                size="small"
                type="text"
                className={getFilterButtonClass(Boolean(filters.has_hot_tub))}
                onClick={() => updateFilter('has_hot_tub', !filters.has_hot_tub)}
              >
                С чаном
              </Button>
              <Button
                size="small"
                type="text"
                className={getFilterButtonClass(Boolean(filters.has_sauna))}
                onClick={() => updateFilter('has_sauna', !filters.has_sauna)}
              >
                С сауной
              </Button>
              <Button
                size="small"
                type="text"
                className={getFilterButtonClass(minRating === 4)}
                onClick={() => {
                  setMinRating((prev) => (prev === 4 ? undefined : 4))
                  setPage(1)
                }}
              >
                Рейтинг 4+
              </Button>
            </Space>
          </div>

          {isCitiesError && !isError && (
            <div style={{ marginTop: 16 }}>
              <PublicState
                kind="degraded"
                compact
                title="Список городов временно недоступен"
                description="Остальной поиск продолжает работать. Можно искать по названию и быстрым фильтрам."
                actionText="Повторить"
                onAction={() => void refetchCities()}
              />
            </div>
          )}

          <div className="bani-catalog__utility-row">
            <Button
              icon={<EnvironmentOutlined />}
              onClick={handleGeoSearch}
              type={geoEnabled ? 'primary' : 'default'}
            >
              {geoEnabled ? 'Рядом со мной' : 'Найти рядом'}
            </Button>
            <Collapse
              ghost
              className="bani-catalog__advanced-collapse"
              items={[
                {
                  key: 'filters',
                  label: (
                    <Space>
                      <FilterOutlined />
                      Фильтры
                    </Space>
                  ),
                  children: (
                    <Card size="small">
                      <Typography.Text type="secondary" className="bani-catalog__advanced-note">
                        Все параметры ниже применяются ко всей выдаче и сразу влияют на счётчик результатов.
                      </Typography.Text>
                      <Row gutter={[16, 16]}>
                        <Col xs={24} sm={12}>
                          <Typography.Text type="secondary">Цена за час</Typography.Text>
                          <Slider
                            range
                            min={0}
                            max={PRICE_FILTER_MAX}
                            step={25000}
                            value={[filters.price_min ?? 0, filters.price_max ?? PRICE_FILTER_MAX]}
                            onChange={(values: number[]) => {
                              const [min, max] = values
                              setFilters((prev) => ({
                                ...prev,
                                price_min: min !== 0 ? min : undefined,
                                price_max: max !== undefined && max < PRICE_FILTER_MAX ? max : undefined,
                              }))
                              setPage(1)
                            }}
                            tooltip={{
                              formatter: (value) => (value != null ? formatPrice(value) : ''),
                            }}
                          />
                        </Col>
                        {geoEnabled && (
                          <Col xs={24} sm={12}>
                            <Typography.Text type="secondary">Способ поиска</Typography.Text>
                            <Segmented
                              block
                              options={[
                                { value: 'radius', label: 'По радиусу' },
                                { value: 'travel_time', label: 'По времени пути' },
                              ]}
                              value={geoMode}
                              onChange={(value) => handleGeoModeChange(value as GeoMode)}
                              style={{ marginTop: 4 }}
                            />
                          </Col>
                        )}
                        {geoEnabled && geoMode === 'radius' && (
                          <Col xs={24} sm={12}>
                            <Typography.Text type="secondary">Радиус (км)</Typography.Text>
                            <InputNumber
                              min={1}
                              max={100}
                              value={radius_km ?? 10}
                              onChange={(value) => updateFilter('radius_km', value)}
                              style={{ width: '100%' }}
                            />
                          </Col>
                        )}
                        {geoEnabled && geoMode === 'travel_time' && (
                          <>
                            <Col xs={12} sm={6}>
                              <Typography.Text type="secondary">Способ</Typography.Text>
                              <Select
                                style={{ width: '100%' }}
                                value={travelMode}
                                onChange={handleTravelModeChange}
                                options={TRAVEL_MODE_OPTIONS}
                              />
                            </Col>
                            <Col xs={12} sm={6}>
                              <Typography.Text type="secondary">Время в пути</Typography.Text>
                              <Select
                                style={{ width: '100%' }}
                                value={travelMinutes}
                                onChange={handleTravelMinutesChange}
                                options={TRAVEL_TIME_OPTIONS}
                              />
                            </Col>
                          </>
                        )}
                        <Col xs={12} sm={6}>
                          <Typography.Text type="secondary">Время с</Typography.Text>
                          <TimePicker
                            style={{ width: '100%' }}
                            format="HH:mm"
                            minuteStep={30}
                            placeholder="С"
                            value={availableTimeFrom}
                            onChange={(timeValue) => {
                              setAvailableTimeFrom(timeValue)
                              setPage(1)
                            }}
                            data-testid="time-from-filter"
                          />
                        </Col>
                        <Col xs={12} sm={6}>
                          <Typography.Text type="secondary">Время до</Typography.Text>
                          <TimePicker
                            style={{ width: '100%' }}
                            format="HH:mm"
                            minuteStep={30}
                            placeholder="До"
                            value={availableTimeTo}
                            onChange={(timeValue) => {
                              setAvailableTimeTo(timeValue)
                              setPage(1)
                            }}
                            data-testid="time-to-filter"
                          />
                        </Col>
                        <Col xs={24} sm={12}>
                          <Typography.Text type="secondary">Мин. рейтинг</Typography.Text>
                          <Select
                            style={{ width: '100%' }}
                            placeholder="Любой"
                            allowClear
                            value={minRating}
                            onChange={(value) => {
                              setMinRating(value)
                              setPage(1)
                            }}
                            options={[
                              { value: 3, label: 'От 3.0' },
                              { value: 3.5, label: 'От 3.5' },
                              { value: 4, label: 'От 4.0' },
                              { value: 4.5, label: 'От 4.5' },
                            ]}
                            data-testid="min-rating-filter"
                          />
                        </Col>
                        <Col xs={24}>
                          <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
                            Удобства
                          </Typography.Text>
                          <Space wrap>
                            <Checkbox checked={filters.has_sauna} onChange={(event) => updateFilter('has_sauna', event.target.checked || undefined)}>Сауна</Checkbox>
                            <Checkbox checked={filters.has_steam_room} onChange={(event) => updateFilter('has_steam_room', event.target.checked || undefined)}>Парная</Checkbox>
                            <Checkbox checked={filters.has_pool} onChange={(event) => updateFilter('has_pool', event.target.checked || undefined)}>Бассейн</Checkbox>
                            <Checkbox checked={filters.has_hot_tub} onChange={(event) => updateFilter('has_hot_tub', event.target.checked || undefined)}>Джакузи</Checkbox>
                            <Checkbox checked={filters.has_bbq} onChange={(event) => updateFilter('has_bbq', event.target.checked || undefined)}>Мангал</Checkbox>
                            <Checkbox checked={filters.has_karaoke} onChange={(event) => updateFilter('has_karaoke', event.target.checked || undefined)}>Караоке</Checkbox>
                          </Space>
                        </Col>
                      </Row>
                    </Card>
                  ),
                },
              ]}
            />
          </div>
        </Card>
      </Space>

      {activeFilters.length > 0 && (
        <section className="bani-catalog__active-filters" data-testid="active-filter-summary">
          <div className="bani-catalog__active-filters-copy">
            <Text className="bani-catalog__eyebrow">Активные фильтры</Text>
            <Typography.Text>Эти параметры уже применены ко всей выдаче и учтены в счётчике результатов.</Typography.Text>
          </div>
          <div className="bani-catalog__active-filter-tags">
            {activeFilters.map((item) => (
              <Tag
                key={item.key}
                closable
                onClose={(event) => {
                  event.preventDefault()
                  item.onRemove()
                }}
              >
                {item.label}
              </Tag>
            ))}
            <Button type="link" onClick={resetDiscoveryFilters}>
              Сбросить всё
            </Button>
          </div>
        </section>
      )}

      {compareIds.length > 0 && (
        <div className="bani-catalog__compare-bar">
          <Space>
            <SwapOutlined />
            <span>
              Выбрано для сравнения: <Badge count={compareIds.length} style={{ backgroundColor: '#0f766e' }} />
            </span>
          </Space>
          <Space>
            <Button size="small" onClick={() => setCompareIds([])}>
              Сбросить
            </Button>
            <Button
              type="primary"
              size="small"
              disabled={compareIds.length < 2}
              onClick={handleGoCompare}
            >
              Сравнить
            </Button>
          </Space>
        </div>
      )}

      {meta?.total_count != null && (
        <div className="bani-catalog__result-toolbar">
          <div className="bani-catalog__result-summary">
            <Tag color="blue" style={{ fontSize: 14, padding: '2px 10px' }} data-testid="result-counter">
              {isLoading ? '...' : `Найдено: ${pluralizeBathhouse(meta.total_count)}`}
            </Tag>
            <Text type="secondary">Сортировка: {selectedSortLabel}</Text>
            <Text type="secondary">
              {activeFilters.length > 0 ? `Активно фильтров: ${activeFilters.length}` : 'Без дополнительных ограничений'}
            </Text>
          </div>
        </div>
      )}

      {viewMode === 'list' && listContent}

      {viewMode === 'map' && (
        <div className="bani-catalog__map">
          <BathhouseMap
            bathhouses={bathhouses}
            highlightedId={highlightedId}
            onBoundsChange={handleBoundsChange}
            onMarkerClick={handleMarkerClick}
            onMarkerHover={setHighlightedId}
            showMiniCard
            center={geoCoords ?? undefined}
            isochronePolygon={isochronePolygon}
            style={{ height: 600, borderRadius: 24, overflow: 'hidden' }}
          />
        </div>
      )}

      {viewMode === 'split' && (
        <Row gutter={16}>
          <Col xs={24} md={12}>
            <div className="bani-catalog__split-list">{listContent}</div>
          </Col>
          <Col xs={24} md={12}>
            <div className="bani-catalog__map">
              <BathhouseMap
                bathhouses={bathhouses}
                highlightedId={highlightedId}
                onBoundsChange={handleBoundsChange}
                onMarkerClick={handleMarkerClick}
                onMarkerHover={setHighlightedId}
                showMiniCard
                center={geoCoords ?? undefined}
                isochronePolygon={isochronePolygon}
                style={{ height: 700, borderRadius: 24, overflow: 'hidden' }}
              />
            </div>
          </Col>
        </Row>
      )}
    </div>
  )
}

function pluralizeBathhouse(count: number): string {
  const mod10 = count % 10
  const mod100 = count % 100
  if (mod100 >= 11 && mod100 <= 19) return `${count} бань`
  if (mod10 === 1) return `${count} баня`
  if (mod10 >= 2 && mod10 <= 4) return `${count} бани`
  return `${count} бань`
}

function haversineDistance(lat1: number, lon1: number, lat2: number, lon2: number): number {
  const R = 6371
  const dLat = ((lat2 - lat1) * Math.PI) / 180
  const dLon = ((lon2 - lon1) * Math.PI) / 180
  const a =
    Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos((lat1 * Math.PI) / 180) *
      Math.cos((lat2 * Math.PI) / 180) *
      Math.sin(dLon / 2) *
      Math.sin(dLon / 2)
  return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
}

function getQueryErrorStatus(error: unknown): number | undefined {
  if (isAxiosError(error)) return error.response?.status
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { status?: number } }).response
    return typeof response?.status === 'number' ? response.status : undefined
  }
  return undefined
}

function isLikelyServiceUnavailable(error: unknown): boolean {
  if (!error) return false
  if (isAxiosError(error)) {
    if (!error.response) return true
    const body = typeof error.response.data === 'string' ? error.response.data.toLowerCase() : ''
    const message = `${error.message} ${body}`.toLowerCase()
    return message.includes('econnrefused') || message.includes('proxy') || message.includes('network')
  }
  return false
}
