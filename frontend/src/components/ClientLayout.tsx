import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
  Layout,
  Menu,
  Button,
  Dropdown,
  Grid,
  Drawer,
  theme,
} from 'antd'
import {
  SearchOutlined,
  CalendarOutlined,
  HeartOutlined,
  TrophyOutlined,
  UsergroupAddOutlined,
  GiftOutlined,
  WalletOutlined,
  MessageOutlined,
  UserOutlined,
  LogoutOutlined,
  MenuOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  BulbOutlined,
  BellOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import type { MenuProps } from 'antd'
import { useAuthStore } from '@/stores/auth'
import NotificationBell from '@/components/NotificationBell'

const { Header, Sider, Content } = Layout
const { useBreakpoint } = Grid

const clientMenuItems: MenuProps['items'] = [
  { key: '/client', icon: <SearchOutlined />, label: 'Поиск бань' },
  { key: '/client/bookings', icon: <CalendarOutlined />, label: 'Мои бронирования' },
  { key: '/client/favorites', icon: <HeartOutlined />, label: 'Избранное' },
  { key: '/client/recommendations', icon: <BulbOutlined />, label: 'Рекомендации' },
  { key: '/client/preferences', icon: <SettingOutlined />, label: 'Предпочтения' },
  { key: '/client/loyalty', icon: <TrophyOutlined />, label: 'Лояльность' },
  { key: '/client/referral', icon: <UsergroupAddOutlined />, label: 'Рефералы' },
  { key: '/client/certificates', icon: <GiftOutlined />, label: 'Сертификаты' },
  { key: '/client/payments', icon: <WalletOutlined />, label: 'Платежи' },
  { key: '/client/chat', icon: <MessageOutlined />, label: 'Чат' },
  { key: '/client/notifications', icon: <BellOutlined />, label: 'Уведомления' },
  { key: '/client/profile', icon: <UserOutlined />, label: 'Профиль' },
]

export default function ClientLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuthStore()
  const screens = useBreakpoint()
  const { token: themeToken } = theme.useToken()

  const isMobile = !screens.md

  const selectedKey = location.pathname
  const matchedKey = clientMenuItems?.find(item => {
    if (!item || !('key' in item)) return false
    const key = item.key as string
    if (key === '/client') return selectedKey === '/client'
    return selectedKey.startsWith(key)
  })?.key as string
  const selectedKeys = [matchedKey || '/client']

  const handleMenuClick: MenuProps['onClick'] = ({ key }) => {
    navigate(key)
    if (isMobile) setDrawerOpen(false)
  }

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: 'Профиль',
      onClick: () => navigate('/client/profile'),
    },
    { type: 'divider' },
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

  const siderMenu = (
    <Menu
      mode="inline"
      selectedKeys={selectedKeys}
      items={clientMenuItems}
      onClick={handleMenuClick}
      style={{ borderRight: 0 }}
    />
  )

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {isMobile ? (
        <Drawer
          placement="left"
          open={drawerOpen}
          onClose={() => setDrawerOpen(false)}
          width={250}
          styles={{ body: { padding: 0 } }}
        >
          <div
            style={{
              height: 64,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              borderBottom: `1px solid ${themeToken.colorBorderSecondary}`,
            }}
          >
            <span style={{ fontSize: 18, fontWeight: 600 }}>Бани</span>
          </div>
          {siderMenu}
        </Drawer>
      ) : (
        <Sider
          collapsible
          collapsed={collapsed}
          onCollapse={setCollapsed}
          breakpoint="lg"
          theme="light"
          style={{
            overflow: 'auto',
            height: '100vh',
            position: 'sticky',
            top: 0,
            left: 0,
          }}
        >
          <div
            style={{
              height: 64,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              borderBottom: `1px solid ${themeToken.colorBorderSecondary}`,
            }}
          >
            <span style={{ fontSize: collapsed ? 16 : 18, fontWeight: 600 }}>
              {collapsed ? 'Б' : 'Бани'}
            </span>
          </div>
          {siderMenu}
        </Sider>
      )}

      <Layout>
        <Header
          style={{
            background: themeToken.colorBgContainer,
            padding: '0 24px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: `1px solid ${themeToken.colorBorderSecondary}`,
            position: 'sticky',
            top: 0,
            zIndex: 10,
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            {isMobile ? (
              <Button
                type="text"
                icon={<MenuOutlined />}
                onClick={() => setDrawerOpen(true)}
              />
            ) : (
              <Button
                type="text"
                icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                onClick={() => setCollapsed(!collapsed)}
              />
            )}
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <NotificationBell />
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <Button type="text" icon={<UserOutlined />}>
                {!isMobile && (user?.name ?? user?.email ?? 'Профиль')}
              </Button>
            </Dropdown>
          </div>
        </Header>

        <Content style={{ margin: 24 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
