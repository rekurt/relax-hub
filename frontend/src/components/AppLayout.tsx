import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
  Layout,
  Menu,
  Button,
  Dropdown,
  Breadcrumb,
  Grid,
  Drawer,
  theme,
} from 'antd'
import {
  DashboardOutlined,
  ShopOutlined,
  CalendarOutlined,
  ScheduleOutlined,
  StarOutlined,
  DollarOutlined,
  GiftOutlined,
  MessageOutlined,
  TeamOutlined,
  CrownOutlined,
  CodeOutlined,
  CameraOutlined,
  UserOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  MenuOutlined,
} from '@ant-design/icons'
import type { MenuProps } from 'antd'
import { useAuthStore } from '@/stores/auth'
import BathhouseSelector from '@/components/BathhouseSelector'

const { Header, Sider, Content } = Layout
const { useBreakpoint } = Grid

const menuItems: MenuProps['items'] = [
  { key: '/', icon: <DashboardOutlined />, label: 'Дашборд' },
  { key: '/bathhouses', icon: <ShopOutlined />, label: 'Бани' },
  { key: '/bookings', icon: <CalendarOutlined />, label: 'Бронирования' },
  { key: '/reviews', icon: <StarOutlined />, label: 'Отзывы' },
  { key: '/calendar', icon: <ScheduleOutlined />, label: 'Календарь' },
  { key: '/pricing', icon: <DollarOutlined />, label: 'Цены' },
  { key: '/promo', icon: <GiftOutlined />, label: 'Промокоды' },
  { key: '/chat', icon: <MessageOutlined />, label: 'Чат' },
  { key: '/representatives', icon: <TeamOutlined />, label: 'Представители' },
  { key: '/subscriptions', icon: <CrownOutlined />, label: 'Подписки' },
  { key: '/widget', icon: <CodeOutlined />, label: 'Виджет' },
  { key: '/photos', icon: <CameraOutlined />, label: 'Фото' },
  { key: '/settings', icon: <UserOutlined />, label: 'Настройки' },
]

const breadcrumbNameMap: Record<string, string> = {
  '/': 'Дашборд',
  '/bathhouses': 'Бани',
  '/bathhouses/new': 'Новая баня',
  '/bookings': 'Бронирования',
  '/reviews': 'Отзывы',
  '/calendar': 'Календарь',
  '/pricing': 'Цены',
  '/promo': 'Промокоды',
  '/chat': 'Чат',
  '/representatives': 'Представители',
  '/subscriptions': 'Подписки',
  '/widget': 'Виджет',
  '/photos': 'Фото',
  '/settings': 'Настройки',
  '/notifications': 'Уведомления',
}

function useBreadcrumbs() {
  const location = useLocation()
  const pathSnippets = location.pathname.split('/').filter((i) => i)

  const items = [{ title: 'Главная', href: '/' }]

  let currentPath = ''
  for (const snippet of pathSnippets) {
    currentPath += `/${snippet}`
    const name = breadcrumbNameMap[currentPath]
    if (name) {
      items.push({ title: name, href: currentPath })
    }
  }

  return items
}

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuthStore()
  const screens = useBreakpoint()
  const { token: themeToken } = theme.useToken()

  const isMobile = !screens.md

  const breadcrumbs = useBreadcrumbs()

  const selectedKey = '/' + (location.pathname.split('/')[1] ?? '')
  const selectedKeys = [selectedKey === '/' ? '/' : selectedKey]

  const handleMenuClick: MenuProps['onClick'] = ({ key }) => {
    navigate(key)
    if (isMobile) setDrawerOpen(false)
  }

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'settings',
      icon: <UserOutlined />,
      label: 'Настройки профиля',
      onClick: () => navigate('/settings'),
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
      items={menuItems}
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
            <BathhouseSelector />
          </div>

          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <Button type="text" icon={<UserOutlined />}>
              {!isMobile && (user?.name ?? user?.email ?? 'Профиль')}
            </Button>
          </Dropdown>
        </Header>

        <Content style={{ margin: 24 }}>
          <Breadcrumb
            style={{ marginBottom: 16 }}
            items={breadcrumbs.map((item) => ({
              title: item.href ? item.title : item.title,
            }))}
          />
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
