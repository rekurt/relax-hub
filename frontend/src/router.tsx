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
import PricingRules from '@/pages/pricing/PricingRules'
import PromoList from '@/pages/promo/PromoList'
import RepresentativeList from '@/pages/representatives/RepresentativeList'
import ChatPage from '@/pages/chat/ChatPage'
import NotificationList from '@/pages/notifications/NotificationList'
import ProfileSettings from '@/pages/settings/ProfileSettings'
import SubscriptionPage from '@/pages/subscriptions/SubscriptionPage'
import WidgetSettings from '@/pages/widget/WidgetSettings'
import PhotoManager from '@/pages/photos/PhotoManager'

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
        <Route path="pricing" element={<PricingRules />} />
        <Route path="promo" element={<PromoList />} />
        <Route path="chat" element={<ChatPage />} />
        <Route path="representatives" element={<RepresentativeList />} />
        <Route path="subscriptions" element={<SubscriptionPage />} />
        <Route path="widget" element={<WidgetSettings />} />
        <Route path="photos" element={<PhotoManager />} />
        <Route path="settings" element={<ProfileSettings />} />
        <Route path="notifications" element={<NotificationList />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
