import { Navigate, useLocation } from 'react-router-dom'
import { Button, Result, Spin } from 'antd'
import { useAuthStore } from '@/stores/auth'

interface ProtectedRouteProps {
  children: React.ReactNode
  allowedRoles?: string[]
}

export default function ProtectedRoute({ children, allowedRoles = ['owner', 'representative'] }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, user, logout } = useAuthStore()
  const location = useLocation()

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
        <Spin size="large" />
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  if (user?.role && !allowedRoles.includes(user.role)) {
    return (
      <Result
        status="403"
        title="Доступ запрещён"
        subTitle="Панель управления доступна только для владельцев и представителей бань."
        extra={
          <Button type="primary" onClick={logout}>
            Выйти
          </Button>
        }
      />
    )
  }

  return <>{children}</>
}
