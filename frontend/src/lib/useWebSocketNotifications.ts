import { useEffect, useRef } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import {
  getGetMyConversationsQueryKey,
  getGetMyUnreadMessagesCountQueryKey,
} from '@/api/generated/chat/chat'

const AUTH_TOKEN_KEY = 'bani_token'

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

    function connect() {
      const token = localStorage.getItem(AUTH_TOKEN_KEY)
      if (!token) return

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/api/v1/ws/notifications?token=${token}`

      const ws = new WebSocket(wsUrl)
      wsRef.current = ws

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
        reconnectTimeoutRef.current = setTimeout(connect, 5000)
      }

      ws.onerror = () => {
        ws.close()
      }
    }

    connect()
    return () => {
      clearTimeout(reconnectTimeoutRef.current)
      wsRef.current?.close()
      wsRef.current = null
    }
  }, [enabled, queryClient])

  return wsRef
}
