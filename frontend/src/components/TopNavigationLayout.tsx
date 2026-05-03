import { useState } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { Layout, Button, Drawer, Dropdown, Grid, Space, Typography, theme } from 'antd'
import type { MenuProps } from 'antd'
import { LogoutOutlined, MenuOutlined, UserOutlined } from '@ant-design/icons'
import { useAuthStore } from '@/stores/auth'
import type { NavigationItem, NavigationSection } from '@/navigation/menu'
import NotificationBell from '@/components/NotificationBell'

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
    <Button
      className={`bani-topnav__nav-button${active ? ' bani-topnav__nav-button--active' : ''}`}
      type="text"
      onClick={() => onClick(item.to)}
    >
      {item.label}
    </Button>
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
}: TopNavigationLayoutProps) {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const screens = useBreakpoint()
  const { token } = theme.useToken()
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
        icon: <UserOutlined />,
        label: 'Профиль',
        onClick: () => navigate(profilePath),
      } satisfies NonNullable<MenuProps['items']>[number]]
    : []
  const userMenuItems: MenuProps['items'] = [
    ...(customProfileMenuItems.length > 0 ? customProfileMenuItems : fallbackProfileMenuItems),
    { type: 'divider' as const },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
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
      className={`bani-shell bani-shell--${surface}`}
      style={{
        minHeight: '100vh',
        background: 'transparent',
      }}
    >
      <Header
        className="bani-topnav"
        style={{
          height: 'auto',
          borderBottomColor: token.colorBorderSecondary,
        }}
      >
        <div className="bani-topnav__inner" style={{ maxWidth: contentWidth }}>
          <button
            type="button"
            onClick={() => navigate(homeTo)}
            className="bani-topnav__brand"
            aria-label={brandAriaLabel}
          >
            <span className="bani-topnav__brand-title">
              {brandTitle}
            </span>
            {showBrandSubtitle && brandSubtitle && (
              <Text type="secondary" className="bani-topnav__brand-subtitle">
                {brandSubtitle}
              </Text>
            )}
          </button>

          {!isMobile && (
            <div className={`bani-topnav__nav${navigationMode === 'dropdown' ? ' bani-topnav__nav--dropdown' : ''}`}>
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
                  <Button
                    className="bani-topnav__workspace-button bani-topnav__profile-button"
                    type="default"
                    icon={<MenuOutlined />}
                    aria-label={`Раздел: ${activeNavigationItem?.label ?? 'Разделы'}`}
                  >
                    <span className="bani-topnav__workspace-label">
                      {activeNavigationItem?.label ?? 'Разделы'}
                    </span>
                  </Button>
                </Dropdown>
              ) : (
                <Space size={8} wrap={false}>
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
                      <Button
                        className={`bani-topnav__nav-button${overflowActive ? ' bani-topnav__nav-button--active' : ''}`}
                        type="text"
                        icon={<MenuOutlined />}
                      >
                        Разделы
                      </Button>
                    </Dropdown>
                  )}
                </Space>
              )}
            </div>
          )}

          <div className="bani-topnav__actions">
            {!isMobile && headerAccessory && (
              <div className={`bani-topnav__accessory bani-topnav__accessory--${headerAccessoryVariant}`}>
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
                  classNames={{ root: 'bani-topnav__profile-dropdown' }}
                >
                  <Button type="text" icon={<UserOutlined />} className="bani-topnav__profile-button">
                    {!isMobile && (
                      <span className="bani-topnav__profile-label">
                        {user.name ?? user.email ?? 'Профиль'}
                      </span>
                    )}
                  </Button>
                </Dropdown>
              </>
            ) : (
              <Button
                type="default"
                icon={<UserOutlined />}
                onClick={() => navigate('/login')}
                className="bani-topnav__profile-button bani-topnav__login-button"
              >
                Войти
              </Button>
            )}

            {isMobile && (
              <Button
                type="text"
                icon={<MenuOutlined />}
                aria-label="Открыть меню"
                onClick={() => setDrawerOpen(true)}
              />
            )}
          </div>
        </div>
      </Header>

      <Content className="bani-topnav__content" style={{ padding: isMobile ? '20px 16px 36px' : '28px 24px 52px', flex: '1 0 auto' }}>
        <div className="bani-topnav__content-inner" style={{ maxWidth: contentWidth }}>
          <Outlet />
        </div>
      </Content>

      {footer && (
        <Footer className="bani-topnav__footer">
          <div className="bani-topnav__footer-inner" style={{ maxWidth: contentWidth }}>
            {footer}
          </div>
        </Footer>
      )}

      <Drawer
        title={<div className="bani-topnav__drawer-brand">{brandTitle}</div>}
        placement="right"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
      >
        <Space orientation="vertical" style={{ width: '100%' }} size={12}>
          {headerAccessory && (
            <div className={`bani-topnav__drawer-accessory bani-topnav__drawer-accessory--${headerAccessoryVariant}`}>
              {headerAccessory}
            </div>
          )}
          {user && (
            <div className="bani-topnav__drawer-user">
              <Text strong>{user.name ?? user.email ?? 'Профиль'}</Text>
              <Text type="secondary">{user.role === 'client' ? 'Личный кабинет клиента' : 'Панель управления'}</Text>
            </div>
          )}
          {resolvedDrawerSections.map((section) => (
            <div key={section.key} className="bani-topnav__drawer-group">
              {section.title && (
                <Text type="secondary" className="bani-topnav__drawer-section">
                  {section.title}
                </Text>
              )}
              <Space orientation="vertical" style={{ width: '100%' }} size={8}>
                {section.items.map((item) => (
                  <Button
                    key={item.key}
                    className={`bani-topnav__drawer-button${isActive(item) ? ' bani-topnav__drawer-button--active' : ''}`}
                    type="text"
                    block
                    onClick={() => {
                      navigate(item.to)
                      setDrawerOpen(false)
                    }}
                  >
                    {item.label}
                  </Button>
                ))}
              </Space>
            </div>
          ))}
        </Space>
      </Drawer>
    </Layout>
  )
}
