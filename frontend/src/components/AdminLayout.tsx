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
  DashboardOutlined,
  TeamOutlined,
  ShopOutlined,
  StarOutlined,
  CameraOutlined,
  WarningOutlined,
  EnvironmentOutlined,
  GiftOutlined,
  BellOutlined,
  UserOutlined,
  LogoutOutlined,
  MenuOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  CustomerServiceOutlined,
  SafetyOutlined,
  AppstoreOutlined,
  TagsOutlined,
  CalendarOutlined,
  WalletOutlined,
  ScheduleOutlined,
  SettingOutlined,
  ControlOutlined,
  DollarOutlined,
  BankOutlined,
  FileTextOutlined,
  FunnelPlotOutlined,
  LineChartOutlined,
  BarChartOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons'
import type { MenuProps } from 'antd'
import { useAuthStore } from '@/stores/auth'
import NotificationBell from '@/components/NotificationBell'

const { Header, Sider, Content } = Layout
const { useBreakpoint } = Grid

const adminMenuItems: MenuProps['items'] = [
  { key: '/admin', icon: <DashboardOutlined />, label: 'Дашборд' },
  { key: '/admin/users', icon: <TeamOutlined />, label: 'Пользователи' },
  { key: '/admin/bathhouses', icon: <ShopOutlined />, label: 'Бани' },
  { key: '/admin/reviews', icon: <StarOutlined />, label: 'Отзывы' },
  { key: '/admin/photos', icon: <CameraOutlined />, label: 'Фото' },
  { key: '/admin/complaints', icon: <WarningOutlined />, label: 'Жалобы' },
  { key: '/admin/cities', icon: <EnvironmentOutlined />, label: 'Города' },
  { key: '/admin/promos', icon: <GiftOutlined />, label: 'Промокоды' },
  { key: '/admin/tickets', icon: <CustomerServiceOutlined />, label: 'Обращения' },
  { key: '/admin/antifraud', icon: <SafetyOutlined />, label: 'Антифрод' },
  { key: '/admin/amenities', icon: <AppstoreOutlined />, label: 'Удобства' },
  { key: '/admin/object-types', icon: <TagsOutlined />, label: 'Типы объектов' },
  { key: '/admin/holidays', icon: <CalendarOutlined />, label: 'Праздники' },
  { key: '/admin/bookings', icon: <ScheduleOutlined />, label: 'Бронирования' },
  { key: '/admin/wallets', icon: <WalletOutlined />, label: 'Кошельки' },
  { key: '/admin/roles', icon: <SafetyOutlined />, label: 'Роли' },
  { key: '/admin/notifications', icon: <BellOutlined />, label: 'Уведомления' },
  { key: '/admin/notification-center', icon: <BellOutlined />, label: 'Центр оповещений' },
  { key: '/admin/settings', icon: <SettingOutlined />, label: 'Настройки' },
  { key: '/admin/feature-flags', icon: <ControlOutlined />, label: 'Флаги' },
  { key: '/admin/service-fees', icon: <DollarOutlined />, label: 'Комиссии' },
  { key: '/admin/finance', icon: <BankOutlined />, label: 'Финансы' },
  { key: '/admin/finance/reconciliation', icon: <BankOutlined />, label: 'Банк. сверка' },
  { key: '/admin/audit-log', icon: <FileTextOutlined />, label: 'Журнал аудита' },
  { key: '/admin/analytics/funnels', icon: <FunnelPlotOutlined />, label: 'Воронка' },
  { key: '/admin/analytics/cohorts', icon: <LineChartOutlined />, label: 'Когорты' },
  { key: '/admin/analytics/supply-demand', icon: <BarChartOutlined />, label: 'Спрос/Предл.' },
  { key: '/admin/force-majeure', icon: <ThunderboltOutlined />, label: 'Форс-мажор' },
  { key: '/admin/profile', icon: <UserOutlined />, label: 'Профиль' },
]

export default function AdminLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuthStore()
  const screens = useBreakpoint()
  const { token: themeToken } = theme.useToken()

  const isMobile = !screens.md

  const selectedKey = location.pathname
  const matchedKey = adminMenuItems?.find(item => {
    if (!item || !('key' in item)) return false
    const key = item.key as string
    if (key === '/admin') return selectedKey === '/admin'
    return selectedKey.startsWith(key)
  })?.key as string
  const selectedKeys = [matchedKey || '/admin']

  const handleMenuClick: MenuProps['onClick'] = ({ key }) => {
    navigate(key)
    if (isMobile) setDrawerOpen(false)
  }

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: 'Профиль',
      onClick: () => navigate('/admin/profile'),
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
      items={adminMenuItems}
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
            <span style={{ fontSize: 18, fontWeight: 600 }}>Админ</span>
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
              {collapsed ? 'А' : 'Админ'}
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
