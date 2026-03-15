import { Routes, Route, Navigate } from 'react-router-dom'
import AppLayout from '@/components/AppLayout'
import ProtectedRoute from '@/components/ProtectedRoute'
import Login from '@/pages/Login'
import Register from '@/pages/Register'
import Dashboard from '@/pages/Dashboard'
import BathhouseList from '@/pages/bathhouses/BathhouseList'
import BathhouseForm from '@/pages/bathhouses/BathhouseForm'
import BookingList from '@/pages/bookings/BookingList'
import CalendarPage from '@/pages/calendar/CalendarPage'
import ReviewList from '@/pages/reviews/ReviewList'

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
        <Route path="bathhouses" element={<BathhouseList />} />
        <Route path="bathhouses/new" element={<BathhouseForm />} />
        <Route path="bathhouses/:id/edit" element={<BathhouseForm />} />
        <Route path="bookings" element={<BookingList />} />
        <Route path="reviews" element={<ReviewList />} />
        <Route path="calendar" element={<CalendarPage />} />
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
