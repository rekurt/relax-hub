import { useState } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { Layout, Button, Drawer, Dropdown, Grid, Space, Typography, theme } from 'antd'
import type { MenuProps } from 'antd'
import { DownOutlined, LogoutOutlined, MenuOutlined, UserOutlined } from '@ant-design/icons'
import { useAuthStore } from '@/stores/auth'
import type { NavigationItem } from '@/navigation/menu'
import NotificationBell from '@/components/NotificationBell'

const { Header, Content } = Layout
const { useBreakpoint } = Grid
const { Text } = Typography

interface TopNavigationLayoutProps {
  brandTitle: string
  brandSubtitle: string
  homeTo: string
  primaryItems: NavigationItem[]
  overflowItems?: NavigationItem[]
  profilePath?: string
  headerAccessory?: React.ReactNode
  showNotifications?: boolean
  contentWidth?: number
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
      type={active ? 'primary' : 'text'}
      onClick={() => onClick(item.to)}
      style={{
        borderRadius: 999,
        fontWeight: active ? 600 : 500,
        boxShadow: active ? '0 10px 24px rgba(19, 79, 92, 0.12)' : 'none',
      }}
    >
      {item.label}
    </Button>
  )
}

export default function TopNavigationLayout({
  brandTitle,
  brandSubtitle,
  homeTo,
  primaryItems,
  overflowItems = [],
  profilePath,
  headerAccessory,
  showNotifications = true,
  contentWidth = 1480,
}: TopNavigationLayoutProps) {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const screens = useBreakpoint()
  const { token } = theme.useToken()
  const isMobile = !screens.lg
  const { user, logout } = useAuthStore()
  const searchParams = new URLSearchParams(location.search)

  const isActive = (item: NavigationItem) => item.isActive(location.pathname, searchParams)
  const overflowActive = overflowItems.some(isActive)

  const overflowMenuItems: MenuProps['items'] = overflowItems.map((item) => ({
    key: item.key,
    label: item.label,
    onClick: () => navigate(item.to),
  }))

  const userMenuItems: MenuProps['items'] = [
    ...(profilePath
      ? [{
          key: 'profile',
          icon: <UserOutlined />,
          label: 'Профиль',
          onClick: () => navigate(profilePath),
        } satisfies NonNullable<MenuProps['items']>[number]]
      : []),
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

  const drawerItems = [...primaryItems, ...overflowItems]

  return (
    <Layout
      style={{
        minHeight: '100vh',
        background: 'transparent',
      }}
    >
      <Header
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 20,
          height: 'auto',
          padding: '18px 24px',
          background: 'rgba(255, 252, 247, 0.80)',
          backdropFilter: 'blur(20px)',
          borderBottom: `1px solid ${token.colorBorderSecondary}`,
          boxShadow: '0 12px 36px rgba(15, 23, 42, 0.05)',
        }}
      >
        <div
          style={{
            maxWidth: contentWidth,
            margin: '0 auto',
            display: 'flex',
            alignItems: 'center',
            gap: 20,
          }}
        >
          <button
            type="button"
            onClick={() => navigate(homeTo)}
            style={{
              border: 0,
              background: 'transparent',
              padding: 0,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'flex-start',
              cursor: 'pointer',
            }}
          >
            <Text strong style={{ fontSize: 20, color: '#16343d', letterSpacing: '0.04em' }}>
              {brandTitle}
            </Text>
            {!isMobile && (
              <Text type="secondary" style={{ fontSize: 12, letterSpacing: '0.02em' }}>
                {brandSubtitle}
              </Text>
            )}
          </button>

          {!isMobile && (
            <div style={{ flex: 1, minWidth: 0, overflow: 'hidden' }}>
              <Space size={8} wrap>
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
                      type={overflowActive ? 'primary' : 'text'}
                      style={{ borderRadius: 999, fontWeight: overflowActive ? 600 : 500 }}
                    >
                      Еще <DownOutlined />
                    </Button>
                  </Dropdown>
                )}
              </Space>
            </div>
          )}

          <div style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: 12 }}>
            {!isMobile && headerAccessory}

            {user ? (
              <>
                {showNotifications && <NotificationBell />}
                <Dropdown menu={{ items: userMenuItems }} trigger={['click']} placement="bottomRight">
                  <Button type="text" icon={<UserOutlined />} style={{ borderRadius: 999 }}>
                    {!isMobile && (user.name ?? user.email ?? 'Профиль')}
                  </Button>
                </Dropdown>
              </>
            ) : (
              <Button type="default" onClick={() => navigate('/login')} style={{ borderRadius: 999 }}>
                Войти
              </Button>
            )}

            {isMobile && (
              <Button
                type="text"
                icon={<MenuOutlined />}
                onClick={() => setDrawerOpen(true)}
              />
            )}
          </div>
        </div>
      </Header>

      <Content style={{ padding: isMobile ? '20px 16px 36px' : '28px 24px 52px' }}>
        <div style={{ maxWidth: contentWidth, margin: '0 auto' }}>
          <Outlet />
        </div>
      </Content>

      <Drawer
        title={brandTitle}
        placement="right"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
      >
        <Space orientation="vertical" style={{ width: '100%' }} size={8}>
          {headerAccessory}
          {drawerItems.map((item) => (
            <Button
              key={item.key}
              type={isActive(item) ? 'primary' : 'text'}
              block
              style={{ justifyContent: 'flex-start', borderRadius: 16, height: 44 }}
              onClick={() => {
                navigate(item.to)
                setDrawerOpen(false)
              }}
            >
              {item.label}
            </Button>
          ))}
        </Space>
      </Drawer>
    </Layout>
  )
}
