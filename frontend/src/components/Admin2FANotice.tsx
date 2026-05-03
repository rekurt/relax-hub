import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Alert, App } from '@/components/design/system'
import { ADMIN_2FA_DEEP_LINK } from '@/lib/admin2faNotice'
import { useAuthStore } from '@/stores/auth'

const DESCRIPTION =
  'Для доступа к админ-панели включите двухфакторную аутентификацию.'

export function Admin2FABanner() {
  const user = useAuthStore((s) => s.user)
  if (user?.role !== 'admin') return null
  const method = user.two_fa_method
  if (method === 'totp' || method === 'sms') return null

  return (
    <div className="rh-admin-2fa-banner" style={{ padding: '12px 16px 0' }}>
      <Alert
        type="warning"
        showIcon
        title="Включите двухфакторную аутентификацию"
        description={DESCRIPTION}
        action={
          <Link
            to={ADMIN_2FA_DEEP_LINK}
            className="ant-btn ant-btn-primary ant-btn-sm"
            style={{ display: 'inline-flex', alignItems: 'center' }}
          >
            Включить 2FA
          </Link>
        }
      />
    </div>
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
