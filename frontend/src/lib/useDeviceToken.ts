import { useCallback, useRef, useEffect } from 'react'
import { usePostDeviceTokens, useDeleteDeviceTokensId } from '@/api/generated/device-tokens/device-tokens'
import { useAuthStore } from '@/stores/auth'

const DEVICE_TOKEN_KEY = 'rh_device_token_id'

function getPlatform(): string {
  const ua = navigator.userAgent.toLowerCase()
  if (/iphone|ipad|ipod/.test(ua)) return 'ios'
  if (/android/.test(ua)) return 'android'
  return 'web'
}

export function useDeviceToken() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const registerMutation = usePostDeviceTokens()
  const deleteMutation = useDeleteDeviceTokensId()
  const registeredRef = useRef(false)
  const registerMutationRef = useRef(registerMutation)
  const deleteMutationRef = useRef(deleteMutation)

  useEffect(() => {
    registerMutationRef.current = registerMutation
  }, [registerMutation])

  useEffect(() => {
    deleteMutationRef.current = deleteMutation
  }, [deleteMutation])

  const registerToken = useCallback(
    async (token: string) => {
      try {
        const response = await registerMutationRef.current.mutateAsync({
          data: { token, platform: getPlatform() },
        })
        if (response.data?.id) {
          localStorage.setItem(DEVICE_TOKEN_KEY, response.data.id)
        }
      } catch {
        // Registration failed silently - push notifications won't work but app continues
      }
    },
    [],
  )

  const unregisterToken = useCallback(async () => {
    const tokenId = localStorage.getItem(DEVICE_TOKEN_KEY)
    if (!tokenId) return
    try {
      await deleteMutationRef.current.mutateAsync({ id: tokenId })
    } catch {
      // Deletion failed silently
    } finally {
      localStorage.removeItem(DEVICE_TOKEN_KEY)
    }
  }, [])

  useEffect(() => {
    if (!isAuthenticated) {
      registeredRef.current = false
      return
    }

    if (registeredRef.current) return

    if ('serviceWorker' in navigator && 'PushManager' in window) {
      registeredRef.current = true
      navigator.serviceWorker.ready
        .then((registration) => registration.pushManager.getSubscription())
        .then((subscription) => {
          if (subscription) {
            const token = JSON.stringify(subscription.toJSON())
            registerToken(token)
          }
        })
        .catch(() => {
          // Push not available
        })
    }
  }, [isAuthenticated, registerToken])

  const requestPushPermission = useCallback(async () => {
    if (!('Notification' in window) || Notification.permission !== 'default') return
    const permission = await Notification.requestPermission()
    if (permission === 'granted' && 'serviceWorker' in navigator && 'PushManager' in window) {
      try {
        const registration = await navigator.serviceWorker.ready
        const subscription = await registration.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: undefined,
        })
        const token = JSON.stringify(subscription.toJSON())
        await registerToken(token)
      } catch {
        // Push subscription failed - non-critical
      }
    }
  }, [registerToken])

  return { registerToken, unregisterToken, requestPushPermission }
}
