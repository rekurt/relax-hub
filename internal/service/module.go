package service

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/notification"
	"github.com/nikitaaldaev/bani/internal/payment"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var Module = fx.Module("service",
	fx.Provide(
		NewAccessChecker,
		fx.Annotate(
			NewAuthService,
			fx.As(new(AuthService)),
			fx.As(new(middleware.AuthService)),
		),
		fx.Annotate(NewUserService, fx.As(new(UserService))),
		fx.Annotate(NewBathhouseService, fx.As(new(BathhouseService))),
		fx.Annotate(NewBookingService, fx.As(new(BookingService))),
		fx.Annotate(NewRepresentativeService, fx.As(new(RepresentativeService))),
		fx.Annotate(NewReviewService, fx.As(new(ReviewService))),
		fx.Annotate(NewFavoriteService, fx.As(new(FavoriteService))),
		fx.Annotate(NewCityService, fx.As(new(CityService))),
		fx.Annotate(NewNotificationService, fx.As(new(NotificationService))),
		fx.Annotate(NewOAuthService, fx.As(new(OAuthService))),
		fx.Annotate(NewRecommendationService, fx.As(new(RecommendationService))),
		fx.Annotate(NewSubscriptionService, fx.As(new(SubscriptionService))),
		fx.Annotate(NewPricingService, fx.As(new(PricingService))),
		fx.Annotate(NewLoyaltyService, fx.As(new(LoyaltyService))),
		fx.Annotate(NewChatService, fx.As(new(ChatService))),
		fx.Annotate(NewAnalyticsService, fx.As(new(AnalyticsService))),
		fx.Annotate(NewTelegramLinkService, fx.As(new(TelegramLinkService))),
		fx.Annotate(NewComplaintService, fx.As(new(ComplaintService))),
		fx.Annotate(NewReferralService, fx.As(new(ReferralService))),
		fx.Annotate(NewCertificateService, fx.As(new(CertificateService))),
		fx.Annotate(NewPhotoVerificationService, fx.As(new(PhotoVerificationService))),
		fx.Annotate(NewPromoService, fx.As(new(PromoService))),
		fx.Annotate(NewMediaService, fx.As(new(MediaService))),
		fx.Annotate(NewCalendarService, fx.As(new(CalendarService))),
		fx.Annotate(NewPromotionService, fx.As(new(PromotionService))),
		fx.Annotate(NewDeviceTokenService, fx.As(new(DeviceTokenService))),
		fx.Annotate(NewOTPService, fx.As(new(OTPService))),
		fx.Annotate(NewTwoFAService, fx.As(new(TwoFAService))),
		fx.Annotate(NewWalletService, fx.As(new(WalletService))),
		fx.Annotate(NewPayoutService, fx.As(new(PayoutService))),
		fx.Annotate(
			NewSessionService,
			fx.As(new(SessionService)),
			fx.As(new(middleware.SessionValidator)),
		),
		fx.Annotate(
			func(paymentRepo repository.PaymentRepository, bookingRepo repository.BookingRepository, provider payment.PaymentProvider, notifSvc NotificationService, cfg *config.Config, log *logger.Logger) PaymentService {
				return NewPaymentService(paymentRepo, bookingRepo, provider, notifSvc, cfg.Payment.ReturnURL, log)
			},
			fx.As(new(PaymentService)),
		),
		fx.Annotate(
			func(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, redisClient *redis.Client, emailSender notification.EmailSender, cfg *config.Config, log *logger.Logger) PasswordResetService {
				return NewPasswordResetService(userRepo, sessionRepo, redisClient, emailSender, log, cfg.FrontendURL)
			},
			fx.As(new(PasswordResetService)),
		),
		fx.Annotate(NewAccountDeletionService, fx.As(new(AccountDeletionService))),
		fx.Annotate(NewKYCService, fx.As(new(KYCService))),
		fx.Annotate(NewOfferService, fx.As(new(OfferService))),
		fx.Annotate(NewPaymentDetailsService, fx.As(new(PaymentDetailsService))),
	),
)
