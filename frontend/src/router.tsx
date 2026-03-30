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
import ListingImport from '@/pages/bathhouses/ListingImport'
import AuditLog from '@/pages/bathhouses/AuditLog'
import BookingList from '@/pages/bookings/BookingList'
import CalendarPage from '@/pages/calendar/CalendarPage'
import ReviewList from '@/pages/reviews/ReviewList'
import PricingRules from '@/pages/pricing/PricingRules'
import PromoList from '@/pages/promo/PromoList'
import RepresentativeList from '@/pages/representatives/RepresentativeList'
import ChatPage from '@/pages/chat/ChatPage'
import NotificationList from '@/pages/notifications/NotificationList'
import ProfileSettings from '@/pages/settings/ProfileSettings'
import WebhookSettings from '@/pages/settings/WebhookSettings'
import PMSIntegration from '@/pages/settings/PMSIntegration'
import SubscriptionPage from '@/pages/subscriptions/SubscriptionPage'
import PromotionCampaign from '@/pages/promotion/PromotionCampaign'
import WidgetSettings from '@/pages/widget/WidgetSettings'
import PhotoManager from '@/pages/photos/PhotoManager'
import PhotoOrderPage from '@/pages/photos/PhotoOrderPage'
import GuestCardList from '@/pages/crm/GuestCardList'
import GuestCardDetail from '@/pages/crm/GuestCardDetail'
import SegmentList from '@/pages/crm/SegmentList'
import BroadcastList from '@/pages/crm/BroadcastList'
import BroadcastCreate from '@/pages/crm/BroadcastCreate'
import AutoScenarios from '@/pages/crm/AutoScenarios'
import ResponseTemplates from '@/pages/crm/ResponseTemplates'
import RFMAnalysis from '@/pages/crm/RFMAnalysis'
import SegmentBuilder from '@/pages/crm/SegmentBuilder'
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
import WalletDashboard from '@/pages/client/WalletDashboard'
import ClientProfile from '@/pages/client/ClientProfile'
import SavedCards from '@/pages/client/SavedCards'
import SecuritySettings from '@/pages/client/SecuritySettings'
import ClientChat from '@/pages/client/ClientChat'
import ClientNotifications from '@/pages/client/ClientNotifications'
import NotificationPreferences from '@/pages/client/NotificationPreferences'
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
import TicketManagement from '@/pages/admin/TicketManagement'
import AdminTicketDetail from '@/pages/admin/AdminTicketDetail'
import SupportTickets from '@/pages/client/SupportTickets'
import ClientTicketDetail from '@/pages/client/TicketDetail'
import DisputeList from '@/pages/client/DisputeList'
import ClientDisputeDetail from '@/pages/client/DisputeDetail'
import DisputeCreate from '@/pages/client/DisputeCreate'
import DisputeManagement from '@/pages/admin/DisputeManagement'
import AdminDisputeDetail from '@/pages/admin/AdminDisputeDetail'
import ComparisonPage from '@/pages/client/ComparisonPage'
import SavedSearches from '@/pages/client/SavedSearches'
import AntiFraudDashboard from '@/pages/admin/AntiFraudDashboard'
import AmenityManagement from '@/pages/admin/AmenityManagement'
import ObjectTypeManagement from '@/pages/admin/ObjectTypeManagement'
import HolidayManagement from '@/pages/admin/HolidayManagement'
import WalletManagement from '@/pages/admin/WalletManagement'
import BookingManagement from '@/pages/admin/BookingManagement'
import RoleManagement from '@/pages/admin/RoleManagement'
import AdminNotificationCenter from '@/pages/admin/AdminNotificationCenter'
import GeoHeatmap from '@/pages/admin/GeoHeatmap'
import PlatformSettings from '@/pages/admin/PlatformSettings'
import FeatureFlags from '@/pages/admin/FeatureFlags'
import ServiceFeeConfig from '@/pages/admin/ServiceFeeConfig'
import AdminFinanceDashboard from '@/pages/admin/AdminFinanceDashboard'
import BankReconciliation from '@/pages/admin/BankReconciliation'
import AdminAuditLog from '@/pages/admin/AdminAuditLog'
import ConversionFunnels from '@/pages/admin/ConversionFunnels'
import CohortAnalysis from '@/pages/admin/CohortAnalysis'
import SupplyDemandMetrics from '@/pages/admin/SupplyDemandMetrics'
import OAuthCallback from '@/pages/OAuthCallback'
import ShareRedirect from '@/pages/ShareRedirect'
import OwnerAnalytics from '@/pages/analytics/OwnerAnalytics'
import FinanceDashboard from '@/pages/finance/FinanceDashboard'
import PayoutPage from '@/pages/finance/PayoutPage'
import FinancialReports from '@/pages/finance/FinancialReports'

export default function AppRouter() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route path="/auth/oauth/callback/:provider" element={<OAuthCallback />} />
      <Route path="/certificates/purchase" element={<CertificatePurchase />} />
      <Route path="/bathhouses/:slug" element={<BathhouseDetail />} />
      <Route path="/share/booking/:token" element={<ShareRedirect />} />

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
        <Route path="bathhouses/import" element={<ListingImport />} />
        <Route path="bathhouses/:id/edit" element={<BathhouseForm />} />
        <Route path="bathhouses/:id/audit" element={<AuditLog />} />
        <Route path="bookings" element={<BookingList />} />
        <Route path="reviews" element={<ReviewList />} />
        <Route path="calendar" element={<CalendarPage />} />
        <Route path="pricing" element={<PricingRules />} />
        <Route path="analytics" element={<OwnerAnalytics />} />
        <Route path="promo" element={<PromoList />} />
        <Route path="chat" element={<ChatPage />} />
        <Route path="representatives" element={<RepresentativeList />} />
        <Route path="subscriptions" element={<SubscriptionPage />} />
        <Route path="promotion" element={<PromotionCampaign />} />
        <Route path="widget" element={<WidgetSettings />} />
        <Route path="photos" element={<PhotoManager />} />
        <Route path="photo-order" element={<PhotoOrderPage />} />
        <Route path="crm/guests" element={<GuestCardList />} />
        <Route path="crm/guests/:id" element={<GuestCardDetail />} />
        <Route path="crm/segments" element={<SegmentList />} />
        <Route path="crm/rfm" element={<RFMAnalysis />} />
        <Route path="crm/segments/custom" element={<SegmentBuilder />} />
        <Route path="crm/broadcasts" element={<BroadcastList />} />
        <Route path="crm/broadcasts/new" element={<BroadcastCreate />} />
        <Route path="crm/scenarios" element={<AutoScenarios />} />
        <Route path="crm/templates" element={<ResponseTemplates />} />
        <Route path="finance" element={<FinanceDashboard />} />
        <Route path="finance/payouts" element={<PayoutPage />} />
        <Route path="finance/reports" element={<FinancialReports />} />
        <Route path="settings" element={<ProfileSettings />} />
        <Route path="settings/webhooks" element={<WebhookSettings />} />
        <Route path="settings/pms" element={<PMSIntegration />} />
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
        <Route path="bathhouse/:slug" element={<BathhouseDetail />} />
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
        <Route path="cards" element={<SavedCards />} />
        <Route path="wallet" element={<WalletDashboard />} />
        <Route path="chat" element={<ClientChat />} />
        <Route path="tickets" element={<SupportTickets />} />
        <Route path="tickets/:id" element={<ClientTicketDetail />} />
        <Route path="disputes" element={<DisputeList />} />
        <Route path="disputes/new" element={<DisputeCreate />} />
        <Route path="disputes/:id" element={<ClientDisputeDetail />} />
        <Route path="comparison" element={<ComparisonPage />} />
        <Route path="saved-searches" element={<SavedSearches />} />
        <Route path="notifications" element={<ClientNotifications />} />
        <Route path="notification-preferences" element={<NotificationPreferences />} />
        <Route path="security" element={<SecuritySettings />} />
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
        <Route path="tickets" element={<TicketManagement />} />
        <Route path="tickets/:id" element={<AdminTicketDetail />} />
        <Route path="disputes" element={<DisputeManagement />} />
        <Route path="disputes/:id" element={<AdminDisputeDetail />} />
        <Route path="antifraud" element={<AntiFraudDashboard />} />
        <Route path="amenities" element={<AmenityManagement />} />
        <Route path="object-types" element={<ObjectTypeManagement />} />
        <Route path="holidays" element={<HolidayManagement />} />
        <Route path="wallets" element={<WalletManagement />} />
        <Route path="bookings" element={<BookingManagement />} />
        <Route path="roles" element={<RoleManagement />} />
        <Route path="notifications" element={<AdminNotifications />} />
        <Route path="notification-center" element={<AdminNotificationCenter />} />
        <Route path="heatmap" element={<GeoHeatmap />} />
        <Route path="settings" element={<PlatformSettings />} />
        <Route path="feature-flags" element={<FeatureFlags />} />
        <Route path="service-fees" element={<ServiceFeeConfig />} />
        <Route path="finance" element={<AdminFinanceDashboard />} />
        <Route path="finance/reconciliation" element={<BankReconciliation />} />
        <Route path="audit-log" element={<AdminAuditLog />} />
        <Route path="analytics/funnels" element={<ConversionFunnels />} />
        <Route path="analytics/cohorts" element={<CohortAnalysis />} />
        <Route path="analytics/supply-demand" element={<SupplyDemandMetrics />} />
        <Route path="profile" element={<AdminProfile />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
