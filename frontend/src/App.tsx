import { useEffect } from 'react'
import { App as AntApp } from 'antd'
import { useAuthStore } from '@/stores/auth'
import AppRouter from '@/router'

function AppWithAuth() {
  const loadProfile = useAuthStore((s) => s.loadProfile)

  useEffect(() => {
    loadProfile()
  }, [loadProfile])

  useEffect(() => {
    document.body.classList.add('bani-app', 'rh-app')
    return () => {
      document.body.classList.remove('bani-app', 'rh-app')
    }
  }, [])

  return <AppRouter />
}

export default function App() {
  return (
    <AntApp>
      <AppWithAuth />
    </AntApp>
  )
}
