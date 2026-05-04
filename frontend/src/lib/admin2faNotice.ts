import { useAuthStore } from '@/stores/auth'

export const ADMIN_2FA_REQUIRED_CODE = 'admin_2fa_required'
export const ADMIN_2FA_DEEP_LINK = '/admin/profile?focus2fa=1#two-factor'
export const ADMIN_2FA_START_EVENT = 'relaxhub:admin-2fa-start'

// Imperative trigger for the "admin needs 2FA" notice. Called from the axios
// response interceptor (which lives outside React), routed through zustand to
// a bridge component that owns the toast API.
export function notifyAdmin2FARequired(): void {
  useAuthStore.getState().bumpNotice2FARequired()
}
