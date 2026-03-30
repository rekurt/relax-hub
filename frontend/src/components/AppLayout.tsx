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
  Badge,
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
  ContactsOutlined,
  FundOutlined,
  UserOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  MenuOutlined,
  SettingOutlined,
  ApiOutlined,
  BarChartOutlined,
} from '@ant-design/icons'
import type { MenuProps } from 'antd'
import { useAuthStore } from '@/stores/auth'
import { useGetMyUnreadMessagesCount } from '@/api/generated/chat/chat'
import BathhouseSelector from '@/components/BathhouseSelector'
import NotificationBell from '@/components/NotificationBell'

const { Header, Sider, Content } = Layout
const { useBreakpoint } = Grid

function useMenuItems(unreadCount: number): MenuProps['items'] {
  return [
    { key: '/', icon: <DashboardOutlined />, label: 'Дашборд' },
    { key: '/bathhouses', icon: <ShopOutlined />, label: 'Бани' },
    { key: '/bookings', icon: <CalendarOutlined />, label: 'Бронирования' },
    { key: '/reviews', icon: <StarOutlined />, label: 'Отзывы' },
    { key: '/calendar', icon: <ScheduleOutlined />, label: 'Календарь' },
    { key: '/pricing', icon: <DollarOutlined />, label: 'Цены' },
    { key: '/analytics', icon: <BarChartOutlined />, label: 'Аналитика' },
    { key: '/promo', icon: <GiftOutlined />, label: 'Промокоды' },
    {
      key: '/chat',
      icon: (
        <Badge count={unreadCount} size="small" offset={[4, 0]}>
          <MessageOutlined />
        </Badge>
      ),
      label: 'Чат',
    },
    { key: '/representatives', icon: <TeamOutlined />, label: 'Представители' },
    { key: '/subscriptions', icon: <CrownOutlined />, label: 'Подписки' },
    { key: '/promotion', icon: <FundOutlined />, label: 'Продвижение' },
    { key: '/widget', icon: <CodeOutlined />, label: 'Виджет' },
    { key: '/photos', icon: <CameraOutlined />, label: 'Фото' },
    {
      key: '/finance-group',
      icon: <FundOutlined />,
      label: 'Финансы',
      children: [
        { key: '/finance', label: 'Обзор' },
        { key: '/finance/payouts', label: 'Выплаты' },
        { key: '/finance/reports', label: 'Отчёты' },
      ],
    },
    {
      key: '/crm',
      icon: <ContactsOutlined />,
      label: 'CRM',
      children: [
        { key: '/crm/guests', label: 'Гости' },
        { key: '/crm/segments', label: 'Сегменты' },
        { key: '/crm/rfm', label: 'RFM-анализ' },
        { key: '/crm/broadcasts', label: 'Рассылки' },
        { key: '/crm/scenarios', label: 'Сценарии' },
        { key: '/crm/templates', label: 'Шаблоны' },
      ],
    },
    {
      key: '/settings',
      icon: <SettingOutlined />,
      label: 'Настройки',
      children: [
        { key: '/settings', label: 'Профиль' },
        { key: '/settings/webhooks', label: 'Вебхуки' },
        { key: '/settings/pms', label: 'PMS интеграции', icon: <ApiOutlined /> },
      ],
    },
  ]
}

const breadcrumbNameMap: Record<string, string> = {
  '/': 'Дашборд',
  '/bathhouses': 'Бани',
  '/bathhouses/new': 'Новая баня',
  '/bookings': 'Бронирования',
  '/reviews': 'Отзывы',
  '/calendar': 'Календарь',
  '/pricing': 'Цены',
  '/analytics': 'Аналитика',
  '/promo': 'Промокоды',
  '/chat': 'Чат',
  '/representatives': 'Представители',
  '/subscriptions': 'Подписки',
  '/widget': 'Виджет',
  '/photos': 'Фото',
  '/crm': 'CRM',
  '/crm/guests': 'Гости',
  '/crm/segments': 'Сегменты',
  '/crm/rfm': 'RFM-анализ',
  '/crm/broadcasts': 'Рассылки',
  '/crm/broadcasts/new': 'Новая рассылка',
  '/crm/scenarios': 'Сценарии',
  '/crm/templates': 'Шаблоны',
  '/finance': 'Финансы',
  '/finance/payouts': 'Выплаты',
  '/finance/reports': 'Отчёты',
  '/settings': 'Настройки',
  '/settings/webhooks': 'Вебхуки',
  '/settings/pms': 'PMS интеграции',
  '/promotion': 'Продвижение',
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

  const { data: unreadData } = useGetMyUnreadMessagesCount({
    query: { refetchInterval: 30000 },
  })
  const unreadCount = unreadData?.data?.unread_count ?? 0
  const menuItems = useMenuItems(unreadCount)

  const isMobile = !screens.md

  const breadcrumbs = useBreadcrumbs()

  const pathParts = location.pathname.split('/').filter(Boolean)
  const subMenuPrefixes = ['crm', 'settings', 'finance']
  const selectedKey = (pathParts.length >= 2 && subMenuPrefixes.includes(pathParts[0] ?? ''))
    ? '/' + pathParts.slice(0, 2).join('/')
    : '/' + (pathParts[0] ?? '')
  const selectedKeys = [selectedKey === '/' ? '/' : selectedKey]
  const openKeys: string[] = []
  if (pathParts[0] === 'crm') openKeys.push('/crm')
  if (pathParts[0] === 'finance') openKeys.push('/finance')
  if (pathParts[0] === 'settings') openKeys.push('/settings')

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
      defaultOpenKeys={openKeys}
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
          <Breadcrumb
            style={{ marginBottom: 16 }}
            items={breadcrumbs.map((item) => ({
              title: item.title,
              href: item.href,
            }))}
          />
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
