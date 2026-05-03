import { useParams, useNavigate } from 'react-router-dom'
import { useEffect } from 'react'
import { Spin, Result, Button } from '@/components/design/system'
import { useGetApiV1ShareBookingToken } from '@/api/generated/share/share'

export default function ShareRedirect() {
  const { token } = useParams<{ token: string }>()
  const navigate = useNavigate()

  const { data, isLoading, isError } = useGetApiV1ShareBookingToken(token ?? '', {
    query: { enabled: !!token },
  })

  const resolved = data?.data as { bathhouse_id?: string; booking_id?: string; start_time?: string; end_time?: string; guest_count?: number } | undefined

  useEffect(() => {
    if (!resolved) return

    if (resolved.booking_id) {
      navigate(`/client/bookings/${resolved.booking_id}`, { replace: true })
    } else if (resolved.bathhouse_id) {
      const params = new URLSearchParams()
      if (resolved.start_time) {
        params.set('from', resolved.start_time)
        params.set('date', resolved.start_time.slice(0, 10))
      }
      if (resolved.end_time) params.set('to', resolved.end_time)
      if (resolved.guest_count) params.set('guests', String(resolved.guest_count))
      const qs = params.toString()
      navigate(`/checkout?bathhouse=${resolved.bathhouse_id}${qs ? `&${qs}` : ''}`, { replace: true })
    }
  }, [resolved, navigate])

  if (isLoading) {
    return (
      <div className="rh-fullscreen-state rh-fullscreen-state--stack">
        <Spin size="large" />
      </div>
    )
  }

  if (isError || !token) {
    return (
      <div className="rh-fullscreen-state">
        <Result
          status="warning"
          title="Ссылка недействительна"
          subTitle="Срок действия ссылки истёк или она некорректна."
          extra={
            <Button type="primary" onClick={() => navigate('/')}>
              На главную
            </Button>
          }
        />
      </div>
    )
  }

  return (
    <div className="rh-fullscreen-state rh-fullscreen-state--stack">
      <Spin size="large" />
    </div>
  )
}
