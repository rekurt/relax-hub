import { useState } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { Layout, Drawer, Dropdown, Grid, Space, Typography } from '@/components/design/system'
import type { MenuProps } from '@/components/design/types'
import { useAuthStore } from '@/stores/auth'
import type { NavigationItem, NavigationSection } from '@/navigation/menu'
import NotificationBell from '@/components/NotificationBell'
import { DesignAvatar, DesignButton, DesignIcon } from '@/components/design'

const { Header, Content, Footer } = Layout
const { useBreakpoint } = Grid
const { Text } = Typography

interface TopNavigationLayoutProps {
  brandTitle: React.ReactNode
  brandSubtitle: React.ReactNode
  brandAriaLabel?: string
  surface?: 'public' | 'client' | 'owner' | 'admin'
  homeTo: string
  primaryItems: NavigationItem[]
  overflowItems?: NavigationItem[]
  overflowLabel?: string
  navigationMode?: 'pills' | 'dropdown'
  profilePath?: string
  profileMenuItems?: NavigationItem[]
  headerAccessory?: React.ReactNode
  headerAccessoryVariant?: 'default' | 'city'
  showNotifications?: boolean
  contentWidth?: number
  drawerSections?: NavigationSection[]
  footer?: React.ReactNode
  topBanner?: React.ReactNode
}

type DropdownItem = NonNullable<MenuProps['items']>[number]

function groupNavigationItems(items: NavigationItem[]) {
  const groups: Array<{ section?: string; items: NavigationItem[] }> = []
  const groupBySection = new Map<string, { section?: string; items: NavigationItem[] }>()

  for (const item of items) {
    const key = item.section ?? ''
    let group = groupBySection.get(key)
    if (!group) {
      group = { section: item.section, items: [] }
      groupBySection.set(key, group)
      groups.push(group)
    }

    group.items.push(item)
  }

  return groups
}

function buildGroupedMenuItems(items: NavigationItem[], navigate: (to: string) => void): NonNullable<MenuProps['items']> {
  return groupNavigationItems(items).reduce<NonNullable<MenuProps['items']>>((acc, group) => {
    const children = group.items.map((item) => ({
      key: item.key,
      label: item.label,
      onClick: () => navigate(item.to),
    } satisfies DropdownItem))

    if (!group.section) {
      acc.push(...children)
      return acc
    }

    acc.push({
      key: `group-${group.section}`,
      type: 'group',
      label: group.section,
      children,
    } satisfies DropdownItem)
    return acc
  }, [])
}

function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(' ')
}

function NavButton({
  item,
  active,
  onClick,
}: {
  item: NavigationItem
  active: boolean
  onClick: (to: string) => void
}) {
  return (
    <button
      type="button"
      className={cx(
        'rh-nav__link rh-topnav__nav-button rounded-rh-pill px-3.5 py-2 text-sm font-semibold text-rh-text transition hover:bg-[rgba(15,118,110,0.08)] hover:text-rh-primary-strong',
        active && 'rh-nav__link--active rh-topnav__nav-button--active bg-[rgba(15,118,110,0.10)] text-rh-primary-strong',
      )}
      onClick={() => onClick(item.to)}
    >
      {item.label}
    </button>
  )
}

export default function TopNavigationLayout({
  brandTitle,
  brandSubtitle,
  brandAriaLabel,
  surface = 'public',
  homeTo,
  primaryItems,
  overflowItems = [],
  overflowLabel = 'Разделы',
  navigationMode = 'pills',
  profilePath,
  profileMenuItems = [],
  headerAccessory,
  headerAccessoryVariant = 'default',
  showNotifications = true,
  contentWidth = 1480,
  drawerSections,
  footer,
  topBanner,
}: TopNavigationLayoutProps) {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const screens = useBreakpoint()
  const isMobile = !screens.lg
  const showBrandSubtitle = !isMobile && (navigationMode === 'dropdown' ? !!screens.xxl : !!screens.xl)
  const { user, logout } = useAuthStore()
  const searchParams = new URLSearchParams(location.search)

  const isActive = (item: NavigationItem) => item.isActive(location.pathname, searchParams)
  const allNavigationItems = [...primaryItems, ...overflowItems]
  const overflowActive = overflowItems.some(isActive)
  const activeNavigationItem = allNavigationItems.find(isActive) ?? primaryItems[0] ?? overflowItems[0] ?? null
  const overflowMenuItems = buildGroupedMenuItems(overflowItems, navigate)
  const navigationMenuItems = buildGroupedMenuItems(allNavigationItems, navigate)
  const customProfileMenuItems = profileMenuItems.length > 0 ? buildGroupedMenuItems(profileMenuItems, navigate) : []
  const fallbackProfileMenuItems = profilePath
    ? [{
        key: 'profile',
        icon: <DesignIcon name="user" size={16} />,
        label: 'Профиль',
        onClick: () => navigate(profilePath),
      } satisfies NonNullable<MenuProps['items']>[number]]
    : []
  const userMenuItems: MenuProps['items'] = [
    ...(customProfileMenuItems.length > 0 ? customProfileMenuItems : fallbackProfileMenuItems),
    { type: 'divider' as const },
    {
      key: 'logout',
      icon: <DesignIcon name="out" size={16} />,
      label: 'Выйти',
      danger: true,
      onClick: () => {
        logout()
        navigate('/login')
      },
    },
  ]
  const resolvedDrawerSections = drawerSections ?? groupNavigationItems(allNavigationItems).map((group, index) => ({
    key: group.section ?? `drawer-group-${index}`,
    title: group.section ?? '',
    items: group.items,
  }))

  return (
    <Layout
      className={`rh-shell rh-shell--${surface} min-h-screen bg-transparent font-sans text-rh-text`}
      style={{
        minHeight: '100vh',
        background: 'transparent',
      }}
    >
      <Header
        className="rh-nav rh-topnav sticky top-0 z-40 border-b border-[rgba(15,23,42,0.08)] bg-[rgba(255,252,247,0.80)] px-4 py-3 shadow-[0_4px_14px_rgba(15,23,42,0.04)] backdrop-blur-[20px] sm:px-6"
        style={{
          height: 'auto',
        }}
      >
        <div className="rh-topnav__inner mx-auto flex w-full min-w-0 items-center gap-4" style={{ maxWidth: contentWidth }}>
          <button
            type="button"
            onClick={() => navigate(homeTo)}
            className="rh-topnav__brand flex min-w-0 shrink-0 items-center text-left"
            aria-label={brandAriaLabel}
          >
            <span className="rh-topnav__brand-title">
              {brandTitle}
            </span>
            {showBrandSubtitle && brandSubtitle && (
              <Text type="secondary" className="rh-topnav__brand-subtitle">
                {brandSubtitle}
              </Text>
            )}
          </button>

          {!isMobile && (
            <div className={`rh-topnav__nav${navigationMode === 'dropdown' ? ' rh-topnav__nav--dropdown' : ''}`}>
              {navigationMode === 'dropdown' ? (
                <Dropdown
                  menu={{
                    items: navigationMenuItems,
                    selectable: true,
                    selectedKeys: activeNavigationItem ? [activeNavigationItem.key] : [],
                  }}
                  trigger={['click']}
                  placement="bottomRight"
                  classNames={{ root: 'rh-topnav__workspace-dropdown' }}
                >
                  <button
                    type="button"
                    className="rh-nav__user rh-topnav__workspace-button rh-topnav__profile-button"
                    aria-haspopup="menu"
                    aria-label={`Раздел: ${activeNavigationItem?.label ?? 'Разделы'}`}
                  >
                    <DesignIcon name="grid" size={16} />
                    <span className="rh-topnav__workspace-label">
                      {activeNavigationItem?.label ?? 'Разделы'}
                    </span>
                  </button>
                </Dropdown>
              ) : (
                <div className="rh-nav__links rh-topnav__nav-links flex min-w-0 flex-1 items-center gap-2">
                  {primaryItems.map((item) => (
                    <NavButton
                      key={item.key}
                      item={item}
                      active={isActive(item)}
                      onClick={navigate}
                    />
                  ))}
                  {overflowItems.length > 0 && (
                    <Dropdown
                      menu={{
                        items: overflowMenuItems,
                        selectable: true,
                        selectedKeys: overflowItems.filter(isActive).map((item) => item.key),
                      }}
                      trigger={['click']}
                      placement="bottomRight"
                      classNames={{ root: 'rh-topnav__workspace-dropdown' }}
                    >
                      <button
                        type="button"
                        className={cx(
                          'rh-nav__link rh-topnav__nav-button inline-flex items-center gap-1.5 rounded-rh-pill px-3.5 py-2 text-sm font-semibold text-rh-text transition hover:bg-[rgba(15,118,110,0.08)] hover:text-rh-primary-strong',
                          overflowActive && 'rh-nav__link--active rh-topnav__nav-button--active bg-[rgba(15,118,110,0.10)] text-rh-primary-strong',
                        )}
                        aria-haspopup="menu"
                      >
                        <DesignIcon name="more" size={16} />
                        {overflowLabel}
                      </button>
                    </Dropdown>
                  )}
                </div>
              )}
            </div>
          )}

          <div className="rh-nav__actions rh-topnav__actions ml-auto flex min-w-0 items-center gap-2">
            {!isMobile && headerAccessory && (
              <div className={`rh-topnav__accessory rh-topnav__accessory--${headerAccessoryVariant}`}>
                {headerAccessory}
              </div>
            )}

            {user ? (
              <>
                {showNotifications && <NotificationBell />}
                <Dropdown
                  menu={{ items: userMenuItems }}
                  trigger={['click']}
                  placement="bottomRight"
                  classNames={{ root: 'rh-topnav__profile-dropdown' }}
                >
                  <button type="button" className="rh-nav__user rh-topnav__profile-button inline-flex min-w-0 items-center gap-2 rounded-rh-pill border border-[rgba(15,23,42,0.10)] bg-white/70 px-2.5 py-1.5 text-sm font-semibold text-rh-text transition hover:bg-white/95">
                    <DesignAvatar name={user.name ?? user.email ?? 'Профиль'} size={32} />
                    {!isMobile && (
                      <span className="rh-topnav__profile-label">
                        {user.name ?? user.email ?? 'Профиль'}
                      </span>
                    )}
                  </button>
                </Dropdown>
              </>
            ) : (
              <DesignButton
                variant="primary"
                size="sm"
                onClick={() => navigate('/login')}
                className="rh-topnav__profile-button rh-topnav__login-button"
                icon={<DesignIcon name="user" size={16} />}
              >
                Войти
              </DesignButton>
            )}

            {isMobile && (
              <button
                type="button"
                className="rh-nav__icon-button rh-topnav__menu-button inline-flex h-10 w-10 items-center justify-center rounded-rh-pill border border-[rgba(15,23,42,0.10)] bg-white/75 text-rh-text shadow-rh-soft"
                aria-label="Открыть меню"
                onClick={() => setDrawerOpen(true)}
              >
                <DesignIcon name="grid" size={18} />
              </button>
            )}
          </div>
        </div>
      </Header>

      <Content className="rh-topnav__content flex-1 px-4 py-5 sm:px-6 sm:py-7 lg:pb-[52px]" style={{ flex: '1 0 auto' }}>
        <div className="rh-topnav__content-inner mx-auto w-full min-w-0" style={{ maxWidth: contentWidth }}>
          {topBanner}
          <Outlet />
        </div>
      </Content>

      {footer && (
        <Footer className="rh-topnav__footer px-4 pb-6 sm:px-6">
          <div className="rh-topnav__footer-inner mx-auto w-full" style={{ maxWidth: contentWidth }}>
            {footer}
          </div>
        </Footer>
      )}

      <Drawer
        title={<div className="rh-topnav__drawer-brand">{brandTitle}</div>}
        placement="right"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
      >
        <Space orientation="vertical" style={{ width: '100%' }} size={12}>
          {headerAccessory && (
            <div className={`rh-topnav__drawer-accessory rh-topnav__drawer-accessory--${headerAccessoryVariant}`}>
              {headerAccessory}
            </div>
          )}
          {user && (
            <div className="rh-topnav__drawer-user">
              <Text strong>{user.name ?? user.email ?? 'Профиль'}</Text>
              <Text type="secondary">{user.role === 'client' ? 'Личный кабинет клиента' : 'Панель управления'}</Text>
            </div>
          )}
          {resolvedDrawerSections.map((section) => (
            <div key={section.key} className="rh-topnav__drawer-group">
              {section.title && (
                <Text type="secondary" className="rh-topnav__drawer-section">
                  {section.title}
                </Text>
              )}
              <Space orientation="vertical" style={{ width: '100%' }} size={8}>
                {section.items.map((item) => (
                  <button
                    key={item.key}
                    type="button"
                    className={cx(
                      'rh-nav__link rh-topnav__drawer-button w-full rounded-rh-lg px-4 py-3 text-left text-sm font-semibold text-rh-text transition hover:bg-[rgba(15,118,110,0.08)]',
                      isActive(item) && 'rh-nav__link--active rh-topnav__drawer-button--active bg-[rgba(15,118,110,0.10)] text-rh-primary-strong',
                    )}
                    onClick={() => {
                      navigate(item.to)
                      setDrawerOpen(false)
                    }}
                  >
                    {item.label}
                  </button>
                ))}
              </Space>
            </div>
          ))}
        </Space>
      </Drawer>
    </Layout>
  )
}
