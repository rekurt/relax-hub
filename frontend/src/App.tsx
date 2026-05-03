import { useEffect } from 'react'
import { App as DesignApp } from '@/components/design/system'
import { useAuthStore } from '@/stores/auth'
import AppRouter from '@/router'

function AppWithAuth() {
  const loadProfile = useAuthStore((s) => s.loadProfile)

  useEffect(() => {
    loadProfile()
  }, [loadProfile])

  useEffect(() => {
    document.body.classList.add('rh-app')
    return () => {
      document.body.classList.remove('rh-app')
    }
  }, [])

  return <AppRouter />
}

export default function App() {
  return (
    <DesignApp>
      <AppWithAuth />
    </DesignApp>
  )
}
