import { Routes, Route, Navigate } from 'react-router-dom'
import AppLayout from '@/components/AppLayout'
import ClientLayout from '@/components/ClientLayout'
import AdminLayout from '@/components/AdminLayout'
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
import ClientHome from '@/pages/client/ClientHome'
import BathhouseSearch from '@/pages/client/BathhouseSearch'
import BathhouseDetail from '@/pages/client/BathhouseDetail'
import BookingCreate from '@/pages/client/BookingCreate'
import ClientBookingList from '@/pages/client/BookingList'
import ClientBookingDetail from '@/pages/client/BookingDetail'
import ReviewForm from '@/pages/client/ReviewForm'
import Favorites from '@/pages/client/Favorites'
import Recommendations from '@/pages/client/Recommendations'
import Preferences from '@/pages/client/Preferences'
import LoyaltyDashboard from '@/pages/client/LoyaltyDashboard'
import ReferralProgram from '@/pages/client/ReferralProgram'
import CertificateList from '@/pages/client/CertificateList'
import CertificatePurchase from '@/pages/client/CertificatePurchase'
import PaymentHistory from '@/pages/client/PaymentHistory'
import ClientProfile from '@/pages/client/ClientProfile'
import ClientChat from '@/pages/client/ClientChat'
import ClientNotifications from '@/pages/client/ClientNotifications'
import AdminDashboard from '@/pages/admin/AdminDashboard'
import UserManagement from '@/pages/admin/UserManagement'
import BathhouseModeration from '@/pages/admin/BathhouseModeration'
import ReviewModeration from '@/pages/admin/ReviewModeration'
import PhotoVerification from '@/pages/admin/PhotoVerification'
import ComplaintManagement from '@/pages/admin/ComplaintManagement'
import CityManagement from '@/pages/admin/CityManagement'
import GlobalPromoCodes from '@/pages/admin/GlobalPromoCodes'
import AdminNotifications from '@/pages/admin/AdminNotifications'
import AdminProfile from '@/pages/admin/AdminProfile'
import OAuthCallback from '@/pages/OAuthCallback'

export default function AppRouter() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route path="/auth/oauth/callback/:provider" element={<OAuthCallback />} />
      <Route path="/certificates/purchase" element={<CertificatePurchase />} />

      {/* Owner/Representative routes */}
      <Route
        path="/"
        element={
          <ProtectedRoute allowedRoles={['owner', 'representative']}>
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

      {/* Client routes */}
      <Route
        path="/client"
        element={
          <ProtectedRoute allowedRoles={['client']}>
            <ClientLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<ClientHome />} />
        <Route path="search" element={<BathhouseSearch />} />
        <Route path="bathhouse/:id" element={<BathhouseDetail />} />
        <Route path="booking/new" element={<BookingCreate />} />
        <Route path="bookings" element={<ClientBookingList />} />
        <Route path="bookings/:id" element={<ClientBookingDetail />} />
        <Route path="review" element={<ReviewForm />} />
        <Route path="favorites" element={<Favorites />} />
        <Route path="recommendations" element={<Recommendations />} />
        <Route path="preferences" element={<Preferences />} />
        <Route path="loyalty" element={<LoyaltyDashboard />} />
        <Route path="referral" element={<ReferralProgram />} />
        <Route path="certificates" element={<CertificateList />} />
        <Route path="certificates/purchase" element={<CertificatePurchase />} />
        <Route path="payments" element={<PaymentHistory />} />
        <Route path="chat" element={<ClientChat />} />
        <Route path="notifications" element={<ClientNotifications />} />
        <Route path="profile" element={<ClientProfile />} />
      </Route>

      {/* Admin routes */}
      <Route
        path="/admin"
        element={
          <ProtectedRoute allowedRoles={['admin']}>
            <AdminLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<AdminDashboard />} />
        <Route path="users" element={<UserManagement />} />
        <Route path="bathhouses" element={<BathhouseModeration />} />
        <Route path="reviews" element={<ReviewModeration />} />
        <Route path="photos" element={<PhotoVerification />} />
        <Route path="complaints" element={<ComplaintManagement />} />
        <Route path="cities" element={<CityManagement />} />
        <Route path="promos" element={<GlobalPromoCodes />} />
        <Route path="notifications" element={<AdminNotifications />} />
        <Route path="profile" element={<AdminProfile />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
