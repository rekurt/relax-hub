import { Navigate, useLocation } from 'react-router-dom'
import { Spin } from '@/components/design/system'
import { useAuthStore } from '@/stores/auth'
import { getRoleHomePath } from '@/stores/auth'

interface ProtectedRouteProps {
  children: React.ReactNode
  allowedRoles?: string[]
}

export default function ProtectedRoute({ children, allowedRoles }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, user } = useAuthStore()
  const location = useLocation()

  if (isLoading) {
    return (
      <div className="rh-fullscreen-state">
        <Spin size="large" />
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  if (allowedRoles && (!user?.role || !allowedRoles.includes(user.role))) {
    const homePath = getRoleHomePath(user?.role)
    return <Navigate to={homePath} replace />
  }

  return <>{children}</>
}
