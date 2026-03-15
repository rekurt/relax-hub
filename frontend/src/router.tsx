import { Routes, Route, Navigate } from 'react-router-dom'
import AppLayout from '@/components/AppLayout'
import ProtectedRoute from '@/components/ProtectedRoute'
import Login from '@/pages/Login'
import Register from '@/pages/Register'
import Dashboard from '@/pages/Dashboard'

function Placeholder({ title }: { title: string }) {
  return <div>{title}</div>
}

export default function AppRouter() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <AppLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Dashboard />} />
        <Route path="bathhouses" element={<Placeholder title="Бани" />} />
        <Route path="bathhouses/new" element={<Placeholder title="Новая баня" />} />
        <Route path="bathhouses/:id/edit" element={<Placeholder title="Редактирование бани" />} />
        <Route path="bookings" element={<Placeholder title="Бронирования" />} />
        <Route path="reviews" element={<Placeholder title="Отзывы" />} />
        <Route path="calendar" element={<Placeholder title="Календарь" />} />
        <Route path="pricing" element={<Placeholder title="Цены" />} />
        <Route path="promo" element={<Placeholder title="Промокоды" />} />
        <Route path="chat" element={<Placeholder title="Чат" />} />
        <Route path="representatives" element={<Placeholder title="Представители" />} />
        <Route path="subscriptions" element={<Placeholder title="Подписки" />} />
        <Route path="widget" element={<Placeholder title="Виджет" />} />
        <Route path="photos" element={<Placeholder title="Фото" />} />
        <Route path="settings" element={<Placeholder title="Настройки" />} />
        <Route path="notifications" element={<Placeholder title="Уведомления" />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
