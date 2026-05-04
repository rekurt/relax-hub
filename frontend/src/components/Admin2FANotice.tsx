import { useEffect } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { Alert, App } from '@/components/design/system'
import { SafetyOutlined } from '@/components/design/icons'
import { ADMIN_2FA_DEEP_LINK, ADMIN_2FA_START_EVENT, isAdmin2FAEnabled } from '@/lib/admin2faNotice'
import { useAuthStore } from '@/stores/auth'

const DESCRIPTION =
  'Для доступа к админ-панели включите двухфакторную аутентификацию.'

export function Admin2FABanner() {
  const user = useAuthStore((s) => s.user)
  const location = useLocation()
  if (user?.role !== 'admin') return null
  const method = user.two_fa_method
  if (isAdmin2FAEnabled(method)) return null
  const isAdminProfile = location.pathname === '/admin/profile' || location.pathname.startsWith('/admin/profile/')

  return (
    <div className="rh-admin-2fa-banner">
      <Alert
        type="warning"
        showIcon
        icon={<SafetyOutlined />}
        title="Включите двухфакторную аутентификацию"
        description={DESCRIPTION}
        action={
          <Link
            to={ADMIN_2FA_DEEP_LINK}
            className="ant-btn ant-btn-primary ant-btn-sm rh-inline-flex-action"
            onClick={(event) => {
              if (!isAdminProfile || typeof window === 'undefined') return
              event.preventDefault()
              window.dispatchEvent(new CustomEvent(ADMIN_2FA_START_EVENT))
            }}
          >
            Включить 2FA
          </Link>
        }
      />
    </div>
  )
}

export function Admin2FALockedState() {
  return (
    <section className="rh-admin-2fa-lock" aria-labelledby="admin-2fa-lock-title">
      <div className="rh-admin-2fa-lock__icon" aria-hidden="true">
        <SafetyOutlined />
      </div>
      <div className="rh-admin-2fa-lock__copy">
        <span className="rh-admin-2fa-lock__eyebrow">Защищённая админка</span>
        <h1 id="admin-2fa-lock-title" className="rh-admin-2fa-lock__title">
          Включите 2FA для доступа к админ-панели
        </h1>
        <p className="rh-admin-2fa-lock__text">
          Админские разделы и API-запросы будут доступны после настройки TOTP или SMS-кода в профиле.
        </p>
      </div>
      <div className="rh-admin-2fa-lock__actions">
        <Link to={ADMIN_2FA_DEEP_LINK} className="ant-btn ant-btn-primary rh-admin-2fa-lock__button">
          Настроить 2FA
        </Link>
      </div>
    </section>
  )
}

export function Admin2FAToastBridge() {
  const { message } = App.useApp()
  const noticeAt = useAuthStore((s) => s.notice2faRequiredAt)
  const role = useAuthStore((s) => s.user?.role)

  useEffect(() => {
    if (!noticeAt) return
    if (role && role !== 'admin') return
    message.warning(DESCRIPTION)
  }, [noticeAt, role, message])

  return null
}
