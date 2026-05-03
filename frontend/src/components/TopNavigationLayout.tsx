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
  return items.reduce<Array<{ section?: string; items: NavigationItem[] }>>((acc, item) => {
    const lastGroup = acc[acc.length - 1]
    if (lastGroup && lastGroup.section === item.section) {
      lastGroup.items.push(item)
      return acc
    }
    acc.push({ section: item.section, items: [item] })
    return acc
  }, [])
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
      className={`rh-nav__link rh-topnav__nav-button${active ? ' rh-nav__link--active rh-topnav__nav-button--active' : ''}`}
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
      className={`rh-shell rh-shell--${surface}`}
      style={{
        minHeight: '100vh',
        background: 'transparent',
      }}
    >
      <Header
        className="rh-nav rh-topnav"
        style={{
          height: 'auto',
        }}
      >
        <div className="rh-topnav__inner" style={{ maxWidth: contentWidth }}>
          <button
            type="button"
            onClick={() => navigate(homeTo)}
            className="rh-topnav__brand"
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
                >
                  <button
                    type="button"
                    className="rh-nav__user rh-topnav__workspace-button rh-topnav__profile-button"
                    aria-label={`Раздел: ${activeNavigationItem?.label ?? 'Разделы'}`}
                  >
                    <DesignIcon name="grid" size={16} />
                    <span className="rh-topnav__workspace-label">
                      {activeNavigationItem?.label ?? 'Разделы'}
                    </span>
                  </button>
                </Dropdown>
              ) : (
                <div className="rh-nav__links rh-topnav__nav-links">
                  {primaryItems.map((item) => (
                    <NavButton
                      key={item.key}
                      item={item}
                      active={isActive(item)}
                      onClick={navigate}
                    />
                  ))}
                  {overflowItems.length > 0 && (
                    <Dropdown menu={{ items: overflowMenuItems }} trigger={['click']} placement="bottomRight">
                      <button
                        type="button"
                        className={`rh-nav__link rh-topnav__nav-button${overflowActive ? ' rh-nav__link--active rh-topnav__nav-button--active' : ''}`}
                      >
                        <DesignIcon name="more" size={16} />
                        Разделы
                      </button>
                    </Dropdown>
                  )}
                </div>
              )}
            </div>
          )}

          <div className="rh-nav__actions rh-topnav__actions">
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
                  <button type="button" className="rh-nav__user rh-topnav__profile-button">
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
                className="rh-nav__icon-button rh-topnav__menu-button"
                aria-label="Открыть меню"
                onClick={() => setDrawerOpen(true)}
              >
                <DesignIcon name="grid" size={18} />
              </button>
            )}
          </div>
        </div>
      </Header>

      <Content className="rh-topnav__content" style={{ padding: isMobile ? '20px 16px 36px' : '28px 24px 52px', flex: '1 0 auto' }}>
        <div className="rh-topnav__content-inner" style={{ maxWidth: contentWidth }}>
          {topBanner}
          <Outlet />
        </div>
      </Content>

      {footer && (
        <Footer className="rh-topnav__footer">
          <div className="rh-topnav__footer-inner" style={{ maxWidth: contentWidth }}>
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
                    className={`rh-nav__link rh-topnav__drawer-button${isActive(item) ? ' rh-nav__link--active rh-topnav__drawer-button--active' : ''}`}
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
