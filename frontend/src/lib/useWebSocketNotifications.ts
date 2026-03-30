import { useEffect, useRef } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import {
  getGetMyConversationsQueryKey,
  getGetMyUnreadMessagesCountQueryKey,
} from '@/api/generated/chat/chat'
import { AUTH_TOKEN_KEY } from '@/lib/constants'

interface UseWebSocketNotificationsOptions {
  enabled?: boolean
  onMessage?: (data: unknown) => void
  activeConversationId?: string | null
}

export function useWebSocketNotifications({
  enabled = true,
  onMessage,
  activeConversationId,
}: UseWebSocketNotificationsOptions = {}) {
  const wsRef = useRef<WebSocket | null>(null)
  const queryClient = useQueryClient()
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout>>()
  const reconnectDelayRef = useRef(1000)
  const onMessageRef = useRef(onMessage)
  const activeConversationIdRef = useRef(activeConversationId)

  useEffect(() => {
    onMessageRef.current = onMessage
  }, [onMessage])

  useEffect(() => {
    activeConversationIdRef.current = activeConversationId
  }, [activeConversationId])

  useEffect(() => {
    if (!enabled) return

    let mounted = true

    function connect() {
      if (!mounted) return
      const token = localStorage.getItem(AUTH_TOKEN_KEY)
      if (!token) return

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/api/v1/ws/notifications?token=${encodeURIComponent(token)}`

      const ws = new WebSocket(wsUrl)
      wsRef.current = ws

      ws.onopen = () => {
        reconnectDelayRef.current = 1000
      }

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          onMessageRef.current?.(data)

          if (data.type === 'new_message') {
            queryClient.invalidateQueries({
              queryKey: getGetMyConversationsQueryKey(),
            })
            queryClient.invalidateQueries({
              queryKey: getGetMyUnreadMessagesCountQueryKey(),
            })
            const convId = activeConversationIdRef.current
            if (convId && data.conversation_id === convId) {
              queryClient.invalidateQueries({
                queryKey: [`/conversations/${convId}/messages`],
              })
            }
          }
        } catch {
          // ignore non-JSON messages
        }
      }

      ws.onclose = () => {
        wsRef.current = null
        if (!mounted) return
        const delay = reconnectDelayRef.current
        reconnectDelayRef.current = Math.min(delay * 2, 60000)
        reconnectTimeoutRef.current = setTimeout(connect, delay)
      }

      ws.onerror = () => {
        ws.close()
      }
    }

    connect()
    return () => {
      mounted = false
      clearTimeout(reconnectTimeoutRef.current)
      wsRef.current?.close()
      wsRef.current = null
    }
  }, [enabled, queryClient])

  return wsRef
}
