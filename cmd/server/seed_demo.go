package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var seedDemoPassword string

var seedDemoCmd = &cobra.Command{
	Use:   "seed-demo",
	Short: "Seed a connected demo world for public and role-based frontend flows",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		pool, err := pgxpool.New(ctx, cfg.Database.DSN)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer pool.Close()

		hash, err := bcrypt.GenerateFromPassword([]byte(seedDemoPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash demo password: %w", err)
		}

		if err := seedDemoWorld(ctx, pool, string(hash), cfg.FrontendURL); err != nil {
			return err
		}
		if err := clearDemoCaches(ctx, cfg); err != nil {
			log.Printf("warning: failed to clear demo caches: %v\n", err)
		}

		log.Printf("Demo world seeded successfully.\n")
		log.Printf("Demo password for all seeded accounts: %s\n", seedDemoPassword)
		log.Printf("Client demo account: demo.client1@relax-hub.ru / +79990000031\n")
		log.Printf("Owner demo account: demo.owner1@relax-hub.ru / +79990000021\n")
		log.Printf("Admin demo account: demo.admin@relax-hub.ru / +79990000011\n")
		return nil
	},
}

func init() {
	seedDemoCmd.Flags().StringVar(&seedDemoPassword, "password", "DemoPass123!", "password for all demo accounts")
	rootCmd.AddCommand(seedDemoCmd)
}

type demoCity struct {
	ID        int64
	Name      string
	Slug      string
	Latitude  float64
	Longitude float64
	Region    string
}

type demoUser struct {
	ID                  uuid.UUID
	Email               string
	Name                string
	Phone               string
	Role                domain.UserRole
	AdminSubRole        domain.AdminSubRole
	CityID              *int64
	Region              domain.UserRegion
	AvatarURL           string
	Bio                 string
	ReferralCode        string
	OnboardingCompleted bool
}

type demoWallet struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Balance    int64
	HeldAmount int64
	Currency   domain.WalletCurrency
	Status     domain.WalletStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type demoWalletTransaction struct {
	ID            uuid.UUID
	WalletID      uuid.UUID
	Type          domain.WalletTransactionType
	Amount        int64
	BalanceAfter  int64
	Status        domain.WalletTransactionStatus
	Description   string
	ReferenceType string
	ReferenceID   *uuid.UUID
	IsBonus       bool
	ExpiresAt     *time.Time
	CreatedAt     time.Time
}

type demoBathhouse struct {
	ID                         uuid.UUID
	OwnerID                    uuid.UUID
	Name                       string
	Slug                       string
	Description                string
	Address                    string
	CityID                     int64
	Latitude                   float64
	Longitude                  float64
	PricePerHour               int64
	MinDuration                int
	MaxGuests                  int
	HasPool                    bool
	HasSauna                   bool
	HasSteamRoom               bool
	HasHotTub                  bool
	HasBBQ                     bool
	HasKaraoke                 bool
	Rating                     float64
	BayesianRating             float64
	ReviewCount                int
	ConversionRate             float64
	OccupancyRate              float64
	ViewCount                  int64
	Images                     []string
	WorkingHours               []domain.WorkingHours
	Status                     domain.BathhouseStatus
	LongSessionThresholdHours  int
	LongSessionDiscountPercent int
	BaseCapacity               int
	ExtraGuestSurcharge        int64
	LastMinuteEnabled          bool
	LastMinuteDiscountPercent  int
	LastMinuteHoursThreshold   int
	BufferMinutes              int
	LeadTimeHours              int
	MaxAdvanceDays             int
	BookingMode                string
	RequestTimeout             int
	ResponseRate               float64
	AvgResponseTimeMinutes     int
	IsPhotoVerified            bool
	ApiKey                     string
	CancellationPolicy         domain.CancellationPolicy
	SecurityDepositPercent     int
	CalendarToken              string
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

type demoRepresentative struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	BathhouseID uuid.UUID
	OwnerID     uuid.UUID
	Role        domain.RepresentativeRole
	CreatedAt   time.Time
}

type demoSubscription struct {
	ID           uuid.UUID
	BathhouseID  uuid.UUID
	OwnerID      uuid.UUID
	Plan         domain.SubscriptionPlan
	Status       domain.SubscriptionStatus
	StartDate    time.Time
	EndDate      *time.Time
	AutoRenew    bool
	PriceKopecks int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type demoPromotion struct {
	ID              uuid.UUID
	BathhouseID     uuid.UUID
	DailyBidKopecks int64
	BudgetKopecks   int64
	SpentKopecks    int64
	StartDate       time.Time
	EndDate         time.Time
	TargetCityID    *int64
	Status          domain.PromotionStatus
	ImpressionCount int64
	ClickCount      int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type demoFavorite struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	BathhouseID uuid.UUID
	CreatedAt   time.Time
}

type demoBooking struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	BathhouseID         uuid.UUID
	StartTime           time.Time
	EndTime             time.Time
	GuestCount          int
	TotalPrice          int64
	AddOnTotal          int64
	PointsSpent         int64
	ReferralBonusUsed   int64
	BasePrice           int64
	LongSessionDiscount int64
	ExtraGuestSurcharge int64
	LastMinuteDiscount  int64
	ServiceFeeAmount    int64
	ModificationCount   int
	DepositAmount       int64
	DepositStatus       domain.DepositStatus
	DepositExternalID   string
	DepositReleasedAt   *time.Time
	CheckedInAt         *time.Time
	CheckedOutAt        *time.Time
	HoldID              *uuid.UUID
	RejectionReason     string
	CancelledByOwner    bool
	Status              domain.BookingStatus
	Comment             string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type demoPayment struct {
	ID            uuid.UUID
	BookingID     uuid.UUID
	UserID        uuid.UUID
	Amount        int64
	Currency      string
	Status        domain.PaymentStatus
	Provider      string
	ExternalID    string
	PaymentMethod domain.PaymentMethod
	WalletAmount  int64
	CardAmount    int64
	IsHold        bool
	CapturedAt    *time.Time
	RefundAmount  int64
	RefundedAt    *time.Time
	Metadata      map[string]string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type demoReview struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	BathhouseID      uuid.UUID
	BookingID        uuid.UUID
	Rating           int
	Cleanliness      float64
	Accuracy         float64
	Communication    float64
	ValueForMoney    float64
	Text             string
	Status           domain.ReviewStatus
	RejectionReasons []string
	OwnerResponse    string
	OwnerResponseAt  *time.Time
	ModerationScore  *float64
	ModerationFlags  []string
	Images           []string
	RevealAt         *time.Time
	IsRevealed       bool
	ModeratedBy      *uuid.UUID
	ModeratedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type demoClientReview struct {
	ID              uuid.UUID
	OwnerID         uuid.UUID
	ClientID        uuid.UUID
	BookingID       uuid.UUID
	BathhouseID     uuid.UUID
	Punctuality     float64
	Cleanliness     float64
	RuleCompliance  float64
	Rating          int
	Text            string
	RevealAt        time.Time
	IsRevealed      bool
	ModerationScore *float64
	ModerationFlags []string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type demoCertificate struct {
	ID             uuid.UUID
	Code           string
	PurchaserID    *uuid.UUID
	PurchaserEmail string
	RecipientEmail string
	RecipientName  string
	Amount         int64
	Balance        int64
	Message        string
	Status         string
	ValidUntil     time.Time
	RedeemedByID   *uuid.UUID
	CreatedAt      time.Time
}

type demoCertificateUsage struct {
	ID            uuid.UUID
	CertificateID uuid.UUID
	BookingID     uuid.UUID
	Amount        int64
	UsedAt        time.Time
}

type demoNotification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      domain.NotificationType
	Title     string
	Body      string
	Data      map[string]string
	IsRead    bool
	ReadAt    *time.Time
	CreatedAt time.Time
}

type demoAdminNotification struct {
	ID        uuid.UUID
	Role      domain.AdminSubRole
	Severity  domain.AdminNotifSeverity
	Type      domain.AdminNotificationType
	Title     string
	Body      string
	Data      map[string]string
	IsRead    bool
	ReadAt    *time.Time
	ReadBy    *uuid.UUID
	CreatedAt time.Time
}

type demoTicket struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	BookingID  *uuid.UUID
	Category   domain.TicketCategory
	Status     domain.TicketStatus
	Priority   domain.TicketPriority
	Level      domain.TicketLevel
	Subject    string
	AssignedTo *uuid.UUID
	CSATScore  *int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ResolvedAt *time.Time
}

type demoTicketMessage struct {
	ID          uuid.UUID
	TicketID    uuid.UUID
	SenderID    uuid.UUID
	SenderType  domain.TicketSenderType
	Body        string
	Attachments []string
	CreatedAt   time.Time
}

type demoDispute struct {
	ID                 uuid.UUID
	BookingID          uuid.UUID
	InitiatorID        uuid.UUID
	RespondentID       uuid.UUID
	Reason             domain.DisputeReason
	Description        string
	Status             domain.DisputeStatus
	Resolution         *domain.DisputeResolution
	RefundAmount       int64
	CompensationAmount int64
	MediatorID         *uuid.UUID
	MediatorNotes      string
	AppealDeadline     *time.Time
	EvidenceDeadline   *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	ResolvedAt         *time.Time
}

type demoDisputeEvidence struct {
	ID          uuid.UUID
	DisputeID   uuid.UUID
	UserID      uuid.UUID
	Type        domain.DisputeEvidenceType
	URL         string
	Description string
	CreatedAt   time.Time
}

type demoSavedCard struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	ProviderToken string
	Last4         string
	Brand         string
	ExpiryMonth   int
	ExpiryYear    int
	IsDefault     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type demoShare struct {
	ID          uuid.UUID
	Token       string
	CreatedBy   uuid.UUID
	BathhouseID uuid.UUID
	StartTime   time.Time
	EndTime     time.Time
	GuestCount  int
	BookingID   *uuid.UUID
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

type demoFAQ struct {
	ID        uuid.UUID
	Category  domain.FAQCategory
	Question  string
	Answer    string
	Keywords  []string
	SortOrder int
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func seedDemoWorld(ctx context.Context, pool *pgxpool.Pool, passwordHash, frontendURL string) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	tz := loadMoscowLocation()
	now := time.Now().In(tz)
	baseDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, tz)
	hours := defaultWorkingHours()

	cities := []demoCity{
		{ID: 101, Name: "Москва", Slug: "moskva", Latitude: 55.7558, Longitude: 37.6176, Region: "Москва"},
		{ID: 102, Name: "Одинцово", Slug: "odintsovo", Latitude: 55.6780, Longitude: 37.2777, Region: "Московская область"},
		{ID: 103, Name: "Красногорск", Slug: "krasnogorsk", Latitude: 55.8310, Longitude: 37.3302, Region: "Московская область"},
	}

	cityIDs := map[string]int64{
		"moskva":      101,
		"odintsovo":   102,
		"krasnogorsk": 103,
	}

	adminID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	supportID := uuid.MustParse("10000000-0000-0000-0000-000000000002")
	// Extra admin sub-roles so the /admin UI has at least one user in every
	// permission bucket (moderator, support L1, support L3, finance).
	modID := uuid.MustParse("10000000-0000-0000-0000-000000000003")
	support1ID := uuid.MustParse("10000000-0000-0000-0000-000000000004")
	support3ID := uuid.MustParse("10000000-0000-0000-0000-000000000005")
	financeID := uuid.MustParse("10000000-0000-0000-0000-000000000006")
	owner1ID := uuid.MustParse("20000000-0000-0000-0000-000000000001")
	owner2ID := uuid.MustParse("20000000-0000-0000-0000-000000000002")
	rep1ID := uuid.MustParse("20000000-0000-0000-0000-000000000003")
	client1ID := uuid.MustParse("30000000-0000-0000-0000-000000000001")
	client2ID := uuid.MustParse("30000000-0000-0000-0000-000000000002")
	client3ID := uuid.MustParse("30000000-0000-0000-0000-000000000003")
	client4ID := uuid.MustParse("30000000-0000-0000-0000-000000000004")

	users := []demoUser{
		{
			ID: adminID, Email: "demo.admin@relax-hub.ru", Name: "Марина Админ", Phone: "+79990000011",
			Role: domain.RoleAdmin, AdminSubRole: domain.AdminSubRoleSuperAdmin, CityID: ptrInt64(cityIDs["moskva"]),
			Region: domain.RegionRU, AvatarURL: "https://i.pravatar.cc/240?img=48",
			Bio: "Курирует публичную витрину, модерацию и демонстрационный контур.", ReferralCode: "ADMINM11", OnboardingCompleted: true,
		},
		{
			ID: supportID, Email: "demo.support@relax-hub.ru", Name: "Егор Поддержка", Phone: "+79990000012",
			Role: domain.RoleAdmin, AdminSubRole: domain.AdminSubRoleSupportL2, CityID: ptrInt64(cityIDs["moskva"]),
			Region: domain.RegionRU, AvatarURL: "https://i.pravatar.cc/240?img=14",
			Bio: "Ведет клиентские обращения, возвраты и споры в demo-мире.", ReferralCode: "SUPPORT12", OnboardingCompleted: true,
		},
		{
			ID: modID, Email: "demo.admin.mod@relax-hub.ru", Name: "Вера Модератор", Phone: "+79990000013",
			Role: domain.RoleAdmin, AdminSubRole: domain.AdminSubRoleModerator, CityID: ptrInt64(cityIDs["moskva"]),
			Region: domain.RegionRU, AvatarURL: "https://i.pravatar.cc/240?img=47",
			Bio: "Модерирует заявки на публикацию объектов, фото и отзывы.", ReferralCode: "MOD00013", OnboardingCompleted: true,
		},
		{
			ID: support1ID, Email: "demo.admin.sup1@relax-hub.ru", Name: "Паша Первая линия", Phone: "+79990000014",
			Role: domain.RoleAdmin, AdminSubRole: domain.AdminSubRoleSupportL1, CityID: ptrInt64(cityIDs["moskva"]),
			Region: domain.RegionRU, AvatarURL: "https://i.pravatar.cc/240?img=52",
			Bio: "Первая линия поддержки — FAQ, быстрые ответы, маршрутизация L2.", ReferralCode: "SUP1L014", OnboardingCompleted: true,
		},
		{
			ID: support3ID, Email: "demo.admin.sup3@relax-hub.ru", Name: "Арсений Третья линия", Phone: "+79990000015",
			Role: domain.RoleAdmin, AdminSubRole: domain.AdminSubRoleSupportL3, CityID: ptrInt64(cityIDs["moskva"]),
			Region: domain.RegionRU, AvatarURL: "https://i.pravatar.cc/240?img=11",
			Bio: "Ведёт эскалированные споры, координирует с финансовой командой.", ReferralCode: "SUP3L015", OnboardingCompleted: true,
		},
		{
			ID: financeID, Email: "demo.admin.fin@relax-hub.ru", Name: "Ольга Финансы", Phone: "+79990000016",
			Role: domain.RoleAdmin, AdminSubRole: domain.AdminSubRoleFinance, CityID: ptrInt64(cityIDs["moskva"]),
			Region: domain.RegionRU, AvatarURL: "https://i.pravatar.cc/240?img=45",
			Bio: "Сверка платежей, подтверждение выплат, реконсиляция банковских выписок.", ReferralCode: "FIN00016", OnboardingCompleted: true,
		},
		{
			ID: modID, Email: "demo.admin.mod@relax-hub.ru", Name: "Вера Модератор", Phone: "+79990000013",
			Role: domain.RoleAdmin, AdminSubRole: domain.AdminSubRoleModerator, CityID: ptrInt64(cityIDs["moskva"]),
			Region: domain.RegionRU, AvatarURL: "https://i.pravatar.cc/240?img=47",
			Bio: "Модерирует заявки на публикацию объектов, фото и отзывы.", ReferralCode: "MOD00013", OnboardingCompleted: true,
		},
		{
			ID: support1ID, Email: "demo.admin.sup1@relax-hub.ru", Name: "Паша Первая линия", Phone: "+79990000014",
			Role: domain.RoleAdmin, AdminSubRole: domain.AdminSubRoleSupportL1, CityID: ptrInt64(cityIDs["moskva"]),
			Region: domain.RegionRU, AvatarURL: "https://i.pravatar.cc/240?img=52",
			Bio: "Первая линия поддержки — FAQ, быстрые ответы, маршрутизация L2.", ReferralCode: "SUP1L014", OnboardingCompleted: true,
		},
		{
			ID: support3ID, Email: "demo.admin.sup3@relax-hub.ru", Name: "Арсений Третья линия", Phone: "+79990000015",
			Role: domain.RoleAdmin, AdminSubRole: domain.AdminSubRoleSupportL3, CityID: ptrInt64(cityIDs["moskva"]),
			Region: domain.RegionRU, AvatarURL: "https://i.pravatar.cc/240?img=11",
			Bio: "Ведёт эскалированные споры, координирует с финансовой командой.", ReferralCode: "SUP3L015", OnboardingCompleted: true,
		},
		{
			ID: financeID, Email: "demo.admin.fin@relax-hub.ru", Name: "Ольга Финансы", Phone: "+79990000016",
			Role: domain.RoleAdmin, AdminSubRole: domain.AdminSubRoleFinance, CityID: ptrInt64(cityIDs["moskva"]),
			Region: domain.RegionRU, AvatarURL: "https://i.pravatar.cc/240?img=45",
			Bio: "Сверка платежей, подтверждение выплат, реконсиляция банковских выписок.", ReferralCode: "FIN00016", OnboardingCompleted: true,
		},
		{
			ID: owner1ID, Email: "demo.owner1@relax-hub.ru", Name: "Сергей Хозяев", Phone: "+79990000021",
			Role: domain.RoleOwner, CityID: ptrInt64(cityIDs["moskva"]), Region: domain.RegionRU,
			AvatarURL: "https://i.pravatar.cc/240?img=12",
			Bio:       "Управляет премиальными объектами BANI PRO в Москве и пригороде.", ReferralCode: "OWNER021", OnboardingCompleted: true,
		},
		{
			ID: owner2ID, Email: "demo.owner2@relax-hub.ru", Name: "Анна Управляющая", Phone: "+79990000022",
			Role: domain.RoleOwner, CityID: ptrInt64(cityIDs["krasnogorsk"]), Region: domain.RegionRU,
			AvatarURL: "https://i.pravatar.cc/240?img=32",
			Bio:       "Ведет семейные и загородные объекты, тестирует request-mode и частные сценарии.", ReferralCode: "OWNER022", OnboardingCompleted: true,
		},
		{
			ID: rep1ID, Email: "demo.rep1@relax-hub.ru", Name: "Илья Представитель", Phone: "+79990000023",
			Role: domain.RoleRepresentative, CityID: ptrInt64(cityIDs["krasnogorsk"]), Region: domain.RegionRU,
			AvatarURL: "https://i.pravatar.cc/240?img=15",
			Bio:       "Помогает владельцу вести календарь, заявки и коммуникацию по объектам.", ReferralCode: "REP00023", OnboardingCompleted: true,
		},
		{
			ID: client1ID, Email: "demo.client1@relax-hub.ru", Name: "Алина Воронцова", Phone: "+79990000031",
			Role: domain.RoleClient, CityID: ptrInt64(cityIDs["moskva"]), Region: domain.RegionRU,
			AvatarURL: "https://i.pravatar.cc/240?img=23",
			Bio:       "Любит камерные SPA-форматы на двоих, часто платит из кошелька и оставляет отзывы.", ReferralCode: "ALINA031", OnboardingCompleted: true,
		},
		{
			ID: client2ID, Email: "demo.client2@relax-hub.ru", Name: "Максим Серов", Phone: "+79990000032",
			Role: domain.RoleClient, CityID: ptrInt64(cityIDs["odintsovo"]), Region: domain.RegionRU,
			AvatarURL: "https://i.pravatar.cc/240?img=54",
			Bio:       "Бронирует большие дома для компаний и корпоративных выездов.", ReferralCode: "MAXIM032", OnboardingCompleted: true,
		},
		{
			ID: client3ID, Email: "demo.client3@relax-hub.ru", Name: "Екатерина Левина", Phone: "+79990000033",
			Role: domain.RoleClient, CityID: ptrInt64(cityIDs["krasnogorsk"]), Region: domain.RegionRU,
			AvatarURL: "https://i.pravatar.cc/240?img=36",
			Bio:       "Ищет семейные выезды на выходные, внимательно смотрит правила отмены и сертификаты.", ReferralCode: "KATYA033", OnboardingCompleted: true,
		},
		{
			ID: client4ID, Email: "demo.client4@relax-hub.ru", Name: "Новый клиент", Phone: "+79990000034",
			Role: domain.RoleClient, CityID: ptrInt64(cityIDs["moskva"]), Region: domain.RegionRU,
			AvatarURL: "https://i.pravatar.cc/240?img=41",
			Bio:       "Свежий пользователь для onboarding, публичного checkout и нового клиентского UX.", ReferralCode: "FRESH034", OnboardingCompleted: false,
		},
	}
	reviewerUsers := buildSyntheticReviewUsers(cityIDs)
	users = append(users, reviewerUsers...)

	bh1 := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	bh2 := uuid.MustParse("40000000-0000-0000-0000-000000000002")
	bh3 := uuid.MustParse("40000000-0000-0000-0000-000000000003")
	bh4 := uuid.MustParse("40000000-0000-0000-0000-000000000004")
	bh5 := uuid.MustParse("40000000-0000-0000-0000-000000000005")
	bh6 := uuid.MustParse("40000000-0000-0000-0000-000000000006")
	bh7 := uuid.MustParse("40000000-0000-0000-0000-000000000007")
	bh8 := uuid.MustParse("40000000-0000-0000-0000-000000000008")
	bh9 := uuid.MustParse("40000000-0000-0000-0000-000000000009")
	bh10 := uuid.MustParse("40000000-0000-0000-0000-000000000010")
	bh11 := uuid.MustParse("40000000-0000-0000-0000-000000000011")
	bh12 := uuid.MustParse("40000000-0000-0000-0000-000000000012")

	bathhouses := []demoBathhouse{
		{
			ID: bh1, OwnerID: owner1ID, Name: "Тихий Берег", Slug: "tihiy-bereg", CityID: cityIDs["moskva"],
			Description: "Камерная баня на двоих с отдельным двориком, чаном и тихой SPA-подсветкой. Подходит для вечернего приватного сценария без компании.",
			Address:     "Москва, Серебряническая наб., 24", Latitude: 55.7489, Longitude: 37.6452, PricePerHour: rub(6900),
			MinDuration: 2, MaxGuests: 4, BaseCapacity: 2, HasSauna: true, HasSteamRoom: true, HasHotTub: true,
			Rating: 4.9, BayesianRating: 4.82, ReviewCount: 18, ConversionRate: 0.29, OccupancyRate: 0.74, ViewCount: 928,
			Images: demoBathhouseImages(frontendURL, "banya-couple"), WorkingHours: hours, Status: domain.BathhouseStatusActive,
			LongSessionThresholdHours: 4, LongSessionDiscountPercent: 10, ExtraGuestSurcharge: rub(1200), LastMinuteEnabled: true,
			LastMinuteDiscountPercent: 15, LastMinuteHoursThreshold: 6, BufferMinutes: 30, LeadTimeHours: 1, MaxAdvanceDays: 90,
			BookingMode: domain.BookingModeInstant, RequestTimeout: 24, ResponseRate: 0.98, AvgResponseTimeMinutes: 6,
			IsPhotoVerified: true, ApiKey: "demo_widget_tihiy_bereg", CancellationPolicy: domain.CancellationPolicyFlexible,
			SecurityDepositPercent: 10, CalendarToken: "demo_calendar_tihiy_bereg", CreatedAt: baseDay.AddDate(0, -5, 0), UpdatedAt: now,
		},
		{
			ID: bh2, OwnerID: owner1ID, Name: "Сосновый Пар", Slug: "sosnoviy-par", CityID: cityIDs["moskva"],
			Description: "Городской комплекс с бассейном и горячей купелью, куда удобно приехать после работы. Внутри приватная гостиная и зона для чаепития.",
			Address:     "Москва, ул. Доватора, 8", Latitude: 55.7243, Longitude: 37.5645, PricePerHour: rub(8400),
			MinDuration: 2, MaxGuests: 6, BaseCapacity: 4, HasPool: true, HasSauna: true, HasSteamRoom: true, HasHotTub: true,
			Rating: 4.8, BayesianRating: 4.74, ReviewCount: 24, ConversionRate: 0.27, OccupancyRate: 0.69, ViewCount: 1210,
			Images: demoBathhouseImages(frontendURL, "banya-pool"), WorkingHours: hours, Status: domain.BathhouseStatusActive,
			LongSessionThresholdHours: 5, LongSessionDiscountPercent: 12, ExtraGuestSurcharge: rub(900), LastMinuteEnabled: true,
			LastMinuteDiscountPercent: 20, LastMinuteHoursThreshold: 8, BufferMinutes: 45, LeadTimeHours: 2, MaxAdvanceDays: 120,
			BookingMode: domain.BookingModeInstant, RequestTimeout: 24, ResponseRate: 0.97, AvgResponseTimeMinutes: 10,
			IsPhotoVerified: true, ApiKey: "demo_widget_sosnoviy_par", CancellationPolicy: domain.CancellationPolicyModerate,
			SecurityDepositPercent: 15, CalendarToken: "demo_calendar_sosnoviy_par", CreatedAt: baseDay.AddDate(0, -4, 0), UpdatedAt: now,
		},
		{
			ID: bh3, OwnerID: owner1ID, Name: "Лесной Причал", Slug: "lesnoy-prichal", CityID: cityIDs["odintsovo"],
			Description: "Большой дом для компании до 10 гостей с BBQ-зоной, террасой и отдельным залом для долгих посиделок.",
			Address:     "Одинцово, Подушкинское ш., 17", Latitude: 55.6707, Longitude: 37.2394, PricePerHour: rub(9800),
			MinDuration: 3, MaxGuests: 10, BaseCapacity: 6, HasSauna: true, HasSteamRoom: true, HasBBQ: true, HasKaraoke: true,
			Rating: 4.7, BayesianRating: 4.66, ReviewCount: 16, ConversionRate: 0.22, OccupancyRate: 0.61, ViewCount: 874,
			Images: demoBathhouseImages(frontendURL, "banya-company"), WorkingHours: hours, Status: domain.BathhouseStatusActive,
			LongSessionThresholdHours: 5, LongSessionDiscountPercent: 15, ExtraGuestSurcharge: rub(700), LastMinuteEnabled: false,
			LastMinuteDiscountPercent: 20, LastMinuteHoursThreshold: 6, BufferMinutes: 60, LeadTimeHours: 3, MaxAdvanceDays: 120,
			BookingMode: domain.BookingModeInstant, RequestTimeout: 24, ResponseRate: 0.95, AvgResponseTimeMinutes: 18,
			IsPhotoVerified: true, ApiKey: "demo_widget_lesnoy_prichal", CancellationPolicy: domain.CancellationPolicyModerate,
			SecurityDepositPercent: 20, CalendarToken: "demo_calendar_lesnoy_prichal", CreatedAt: baseDay.AddDate(0, -6, 0), UpdatedAt: now,
		},
		{
			ID: bh4, OwnerID: owner1ID, Name: "Белый Чан", Slug: "beliy-chan", CityID: cityIDs["krasnogorsk"],
			Description: "Загородная баня с большим чаном и панорамной купелью. Сценарий для длинных вечеров и выездов на выходные.",
			Address:     "Красногорск, Ильинское ш., 5", Latitude: 55.8215, Longitude: 37.3168, PricePerHour: rub(7600),
			MinDuration: 2, MaxGuests: 8, BaseCapacity: 4, HasSauna: true, HasSteamRoom: true, HasHotTub: true, HasBBQ: true,
			Rating: 4.9, BayesianRating: 4.81, ReviewCount: 27, ConversionRate: 0.31, OccupancyRate: 0.77, ViewCount: 1324,
			Images: demoBathhouseImages(frontendURL, "banya-hottub"), WorkingHours: hours, Status: domain.BathhouseStatusActive,
			LongSessionThresholdHours: 4, LongSessionDiscountPercent: 12, ExtraGuestSurcharge: rub(800), LastMinuteEnabled: true,
			LastMinuteDiscountPercent: 18, LastMinuteHoursThreshold: 10, BufferMinutes: 45, LeadTimeHours: 2, MaxAdvanceDays: 150,
			BookingMode: domain.BookingModeInstant, RequestTimeout: 24, ResponseRate: 0.99, AvgResponseTimeMinutes: 5,
			IsPhotoVerified: true, ApiKey: "demo_widget_beliy_chan", CancellationPolicy: domain.CancellationPolicyFlexible,
			SecurityDepositPercent: 12, CalendarToken: "demo_calendar_beliy_chan", CreatedAt: baseDay.AddDate(0, -7, 0), UpdatedAt: now,
		},
		{
			ID: bh5, OwnerID: owner1ID, Name: "Пар и Бассейн", Slug: "par-i-basseyn", CityID: cityIDs["moskva"],
			Description: "Публичный фаворит с бассейном, мягким сервисом и понятным checkout. Хорошо подходит для первого бронирования нового клиента.",
			Address:     "Москва, Ленинградский проспект, 39", Latitude: 55.7933, Longitude: 37.5454, PricePerHour: rub(9200),
			MinDuration: 2, MaxGuests: 6, BaseCapacity: 4, HasPool: true, HasSauna: true, HasSteamRoom: true, HasHotTub: false,
			Rating: 4.8, BayesianRating: 4.78, ReviewCount: 33, ConversionRate: 0.34, OccupancyRate: 0.81, ViewCount: 1670,
			Images: demoBathhouseImages(frontendURL, "banya-premium-pool"), WorkingHours: hours, Status: domain.BathhouseStatusActive,
			LongSessionThresholdHours: 4, LongSessionDiscountPercent: 10, ExtraGuestSurcharge: rub(1000), LastMinuteEnabled: true,
			LastMinuteDiscountPercent: 15, LastMinuteHoursThreshold: 6, BufferMinutes: 30, LeadTimeHours: 1, MaxAdvanceDays: 100,
			BookingMode: domain.BookingModeInstant, RequestTimeout: 24, ResponseRate: 0.99, AvgResponseTimeMinutes: 7,
			IsPhotoVerified: true, ApiKey: "demo_widget_par_i_basseyn", CancellationPolicy: domain.CancellationPolicyFlexible,
			SecurityDepositPercent: 10, CalendarToken: "demo_calendar_par_i_basseyn", CreatedAt: baseDay.AddDate(0, -8, 0), UpdatedAt: now,
		},
		{
			ID: bh6, OwnerID: owner1ID, Name: "Выходной Хутор", Slug: "vihodnoy-hutor", CityID: cityIDs["odintsovo"],
			Description: "Дом выходного дня с баней, верандой и местом под длинные посиделки. Здесь часто бронируют семейные субботы.",
			Address:     "Одинцово, Дачная ул., 44", Latitude: 55.6591, Longitude: 37.2601, PricePerHour: rub(8700),
			MinDuration: 3, MaxGuests: 9, BaseCapacity: 5, HasSauna: true, HasSteamRoom: true, HasBBQ: true, HasPool: false,
			Rating: 4.6, BayesianRating: 4.58, ReviewCount: 11, ConversionRate: 0.19, OccupancyRate: 0.56, ViewCount: 614,
			Images: demoBathhouseImages(frontendURL, "banya-weekend"), WorkingHours: hours, Status: domain.BathhouseStatusActive,
			LongSessionThresholdHours: 5, LongSessionDiscountPercent: 18, ExtraGuestSurcharge: rub(700), LastMinuteEnabled: false,
			LastMinuteDiscountPercent: 20, LastMinuteHoursThreshold: 6, BufferMinutes: 45, LeadTimeHours: 5, MaxAdvanceDays: 180,
			BookingMode: domain.BookingModeInstant, RequestTimeout: 24, ResponseRate: 0.94, AvgResponseTimeMinutes: 22,
			IsPhotoVerified: true, ApiKey: "demo_widget_vihodnoy_hutor", CancellationPolicy: domain.CancellationPolicyStrict,
			SecurityDepositPercent: 20, CalendarToken: "demo_calendar_vihodnoy_hutor", CreatedAt: baseDay.AddDate(0, -4, -10), UpdatedAt: now,
		},
		{
			ID: bh7, OwnerID: owner2ID, Name: "Купель 24", Slug: "kupel-24", CityID: cityIDs["krasnogorsk"],
			Description: "Объект с ручным подтверждением заявок. Хорошо показывает request-mode, отклики владельца и коммуникацию через карточку брони.",
			Address:     "Красногорск, Речная ул., 14", Latitude: 55.8204, Longitude: 37.3384, PricePerHour: rub(7100),
			MinDuration: 2, MaxGuests: 5, BaseCapacity: 3, HasSauna: true, HasSteamRoom: true, HasHotTub: true,
			Rating: 4.5, BayesianRating: 4.41, ReviewCount: 8, ConversionRate: 0.16, OccupancyRate: 0.49, ViewCount: 502,
			Images: demoBathhouseImages(frontendURL, "banya-request"), WorkingHours: hours, Status: domain.BathhouseStatusActive,
			LongSessionThresholdHours: 4, LongSessionDiscountPercent: 8, ExtraGuestSurcharge: rub(900), LastMinuteEnabled: false,
			LastMinuteDiscountPercent: 20, LastMinuteHoursThreshold: 6, BufferMinutes: 30, LeadTimeHours: 2, MaxAdvanceDays: 60,
			BookingMode: domain.BookingModeRequest, RequestTimeout: 12, ResponseRate: 0.83, AvgResponseTimeMinutes: 38,
			IsPhotoVerified: true, ApiKey: "demo_widget_kupel_24", CancellationPolicy: domain.CancellationPolicyModerate,
			SecurityDepositPercent: 10, CalendarToken: "demo_calendar_kupel_24", CreatedAt: baseDay.AddDate(0, -3, -5), UpdatedAt: now,
		},
		{
			ID: bh8, OwnerID: owner2ID, Name: "Премиум Двор", Slug: "premium-dvor", CityID: cityIDs["moskva"],
			Description: "Флагманский сценарий для дорогих бронирований, сертификатов и подарков. Много приватности, высокий сервис и самый сильный social proof.",
			Address:     "Москва, Рублевское ш., 91", Latitude: 55.7583, Longitude: 37.4031, PricePerHour: rub(12900),
			MinDuration: 3, MaxGuests: 8, BaseCapacity: 4, HasPool: true, HasSauna: true, HasSteamRoom: true, HasHotTub: true, HasBBQ: true,
			Rating: 5.0, BayesianRating: 4.92, ReviewCount: 41, ConversionRate: 0.37, OccupancyRate: 0.84, ViewCount: 2110,
			Images: demoBathhouseImages(frontendURL, "banya-premium"), WorkingHours: hours, Status: domain.BathhouseStatusActive,
			LongSessionThresholdHours: 5, LongSessionDiscountPercent: 12, ExtraGuestSurcharge: rub(1500), LastMinuteEnabled: true,
			LastMinuteDiscountPercent: 10, LastMinuteHoursThreshold: 4, BufferMinutes: 60, LeadTimeHours: 4, MaxAdvanceDays: 180,
			BookingMode: domain.BookingModeInstant, RequestTimeout: 24, ResponseRate: 0.99, AvgResponseTimeMinutes: 4,
			IsPhotoVerified: true, ApiKey: "demo_widget_premium_dvor", CancellationPolicy: domain.CancellationPolicyModerate,
			SecurityDepositPercent: 25, CalendarToken: "demo_calendar_premium_dvor", CreatedAt: baseDay.AddDate(0, -6, -3), UpdatedAt: now,
		},
		{
			ID: bh9, OwnerID: owner2ID, Name: "Семейная Мыза", Slug: "semeynaya-miza", CityID: cityIDs["odintsovo"],
			Description: "Мягкий семейный формат с безопасным двором, детской комнатой и простым выбором слотов для выходных.",
			Address:     "Одинцово, Лесная ул., 7", Latitude: 55.6860, Longitude: 37.2842, PricePerHour: rub(7900),
			MinDuration: 3, MaxGuests: 7, BaseCapacity: 4, HasSauna: true, HasSteamRoom: true, HasBBQ: true,
			Rating: 4.7, BayesianRating: 4.63, ReviewCount: 13, ConversionRate: 0.21, OccupancyRate: 0.57, ViewCount: 720,
			Images: demoBathhouseImages(frontendURL, "banya-family"), WorkingHours: hours, Status: domain.BathhouseStatusActive,
			LongSessionThresholdHours: 4, LongSessionDiscountPercent: 15, ExtraGuestSurcharge: rub(750), LastMinuteEnabled: false,
			LastMinuteDiscountPercent: 20, LastMinuteHoursThreshold: 6, BufferMinutes: 45, LeadTimeHours: 4, MaxAdvanceDays: 150,
			BookingMode: domain.BookingModeInstant, RequestTimeout: 24, ResponseRate: 0.93, AvgResponseTimeMinutes: 19,
			IsPhotoVerified: true, ApiKey: "demo_widget_semeynaya_miza", CancellationPolicy: domain.CancellationPolicyFlexible,
			SecurityDepositPercent: 12, CalendarToken: "demo_calendar_semeynaya_miza", CreatedAt: baseDay.AddDate(0, -5, -7), UpdatedAt: now,
		},
		{
			ID: bh10, OwnerID: owner2ID, Name: "Речной Клуб", Slug: "rechnoy-klub", CityID: cityIDs["krasnogorsk"],
			Description: "Универсальная баня с удобным подъездом, быстрым чекаутом и хорошей доступностью для спонтанных бронирований.",
			Address:     "Красногорск, Набережная ул., 18", Latitude: 55.8222, Longitude: 37.3514, PricePerHour: rub(6800),
			MinDuration: 2, MaxGuests: 6, BaseCapacity: 3, HasSauna: true, HasSteamRoom: true, HasPool: false, HasHotTub: false,
			Rating: 4.6, BayesianRating: 4.55, ReviewCount: 9, ConversionRate: 0.18, OccupancyRate: 0.54, ViewCount: 488,
			Images: demoBathhouseImages(frontendURL, "banya-river"), WorkingHours: hours, Status: domain.BathhouseStatusActive,
			LongSessionThresholdHours: 4, LongSessionDiscountPercent: 8, ExtraGuestSurcharge: rub(650), LastMinuteEnabled: true,
			LastMinuteDiscountPercent: 20, LastMinuteHoursThreshold: 5, BufferMinutes: 30, LeadTimeHours: 1, MaxAdvanceDays: 90,
			BookingMode: domain.BookingModeInstant, RequestTimeout: 24, ResponseRate: 0.96, AvgResponseTimeMinutes: 12,
			IsPhotoVerified: true, ApiKey: "demo_widget_rechnoy_klub", CancellationPolicy: domain.CancellationPolicyFlexible,
			SecurityDepositPercent: 8, CalendarToken: "demo_calendar_rechnoy_klub", CreatedAt: baseDay.AddDate(0, -3, -12), UpdatedAt: now,
		},
		{
			ID: bh11, OwnerID: owner2ID, Name: "Дубовый Черновик", Slug: "duboviy-chernovik", CityID: cityIDs["moskva"],
			Description: "Объект в процессе модерации. Нужен для owner/admin сценариев, но не должен торчать в публичной витрине.",
			Address:     "Москва, ул. Академика Королева, 11", Latitude: 55.8230, Longitude: 37.6260, PricePerHour: rub(7500),
			MinDuration: 2, MaxGuests: 6, BaseCapacity: 4, HasSauna: true, HasSteamRoom: true,
			Rating: 0, BayesianRating: 0, ReviewCount: 0, ConversionRate: 0, OccupancyRate: 0, ViewCount: 31,
			Images: demoBathhouseImages(frontendURL, "banya-pending"), WorkingHours: hours, Status: domain.BathhouseStatusPending,
			LongSessionThresholdHours: 4, LongSessionDiscountPercent: 10, ExtraGuestSurcharge: rub(500), LastMinuteEnabled: false,
			LastMinuteDiscountPercent: 20, LastMinuteHoursThreshold: 6, BufferMinutes: 30, LeadTimeHours: 2, MaxAdvanceDays: 90,
			BookingMode: domain.BookingModeInstant, RequestTimeout: 24, ResponseRate: 1.0, AvgResponseTimeMinutes: 0,
			IsPhotoVerified: false, ApiKey: "demo_widget_duboviy_chernovik", CancellationPolicy: domain.CancellationPolicyFlexible,
			SecurityDepositPercent: 10, CalendarToken: "demo_calendar_duboviy_chernovik", CreatedAt: baseDay.AddDate(0, -1, -4), UpdatedAt: now,
		},
		{
			ID: bh12, OwnerID: owner2ID, Name: "Архивный Пар", Slug: "arhivniy-par", CityID: cityIDs["krasnogorsk"],
			Description: "Отклоненный объект для модерационного кабинета и скрытых разделов. В публичную часть не попадает.",
			Address:     "Красногорск, Центральная ул., 3", Latitude: 55.8295, Longitude: 37.3241, PricePerHour: rub(6200),
			MinDuration: 2, MaxGuests: 5, BaseCapacity: 3, HasSauna: true, HasSteamRoom: false,
			Rating: 0, BayesianRating: 0, ReviewCount: 0, ConversionRate: 0, OccupancyRate: 0, ViewCount: 17,
			Images: demoBathhouseImages(frontendURL, "banya-rejected"), WorkingHours: hours, Status: domain.BathhouseStatusRejected,
			LongSessionThresholdHours: 4, LongSessionDiscountPercent: 8, ExtraGuestSurcharge: rub(400), LastMinuteEnabled: false,
			LastMinuteDiscountPercent: 20, LastMinuteHoursThreshold: 6, BufferMinutes: 30, LeadTimeHours: 2, MaxAdvanceDays: 90,
			BookingMode: domain.BookingModeInstant, RequestTimeout: 24, ResponseRate: 0.72, AvgResponseTimeMinutes: 0,
			IsPhotoVerified: false, ApiKey: "demo_widget_arhivniy_par", CancellationPolicy: domain.CancellationPolicyFlexible,
			SecurityDepositPercent: 10, CalendarToken: "demo_calendar_arhivniy_par", CreatedAt: baseDay.AddDate(0, -1, -9), UpdatedAt: now,
		},
	}

	repAssignments := []demoRepresentative{
		{ID: uuid.MustParse("41000000-0000-0000-0000-000000000001"), UserID: rep1ID, BathhouseID: bh7, OwnerID: owner2ID, Role: domain.RepRoleManager, CreatedAt: now.AddDate(0, -1, 0)},
		{ID: uuid.MustParse("41000000-0000-0000-0000-000000000002"), UserID: rep1ID, BathhouseID: bh10, OwnerID: owner2ID, Role: domain.RepRoleObserver, CreatedAt: now.AddDate(0, -1, -5)},
	}

	subscriptions := []demoSubscription{
		newSubscription("42000000-0000-0000-0000-000000000001", bh1, owner1ID, domain.PlanPremium, now.AddDate(0, -1, 0), now.AddDate(0, 1, 0), true, 500000),
		newSubscription("42000000-0000-0000-0000-000000000002", bh2, owner1ID, domain.PlanPremium, now.AddDate(0, -1, -5), now.AddDate(0, 1, -5), true, 500000),
		newSubscription("42000000-0000-0000-0000-000000000003", bh3, owner1ID, domain.PlanPromoted, now.AddDate(0, -1, -10), now.AddDate(0, 1, -10), true, 1000000),
		newSubscription("42000000-0000-0000-0000-000000000004", bh4, owner1ID, domain.PlanPromoted, now.AddDate(0, -1, -3), now.AddDate(0, 1, -3), true, 1000000),
		newSubscription("42000000-0000-0000-0000-000000000005", bh5, owner1ID, domain.PlanPromoted, now.AddDate(0, -1, -7), now.AddDate(0, 1, -7), true, 1000000),
		newSubscription("42000000-0000-0000-0000-000000000006", bh6, owner1ID, domain.PlanPremium, now.AddDate(0, -1, -9), now.AddDate(0, 1, -9), true, 500000),
		newSubscription("42000000-0000-0000-0000-000000000007", bh7, owner2ID, domain.PlanPremium, now.AddDate(0, -1, -4), now.AddDate(0, 1, -4), true, 500000),
		newSubscription("42000000-0000-0000-0000-000000000008", bh8, owner2ID, domain.PlanPromoted, now.AddDate(0, -1, -15), now.AddDate(0, 1, -15), true, 1000000),
		newSubscription("42000000-0000-0000-0000-000000000009", bh9, owner2ID, domain.PlanPremium, now.AddDate(0, -1, -11), now.AddDate(0, 1, -11), true, 500000),
		newSubscription("42000000-0000-0000-0000-000000000010", bh10, owner2ID, domain.PlanPremium, now.AddDate(0, -1, -6), now.AddDate(0, 1, -6), true, 500000),
	}

	promotions := []demoPromotion{
		{
			ID: uuid.MustParse("43000000-0000-0000-0000-000000000001"), BathhouseID: bh1, DailyBidKopecks: 9000,
			BudgetKopecks: 240000, SpentKopecks: 72000, StartDate: now.AddDate(0, 0, -8), EndDate: now.AddDate(0, 0, 20),
			TargetCityID: ptrInt64(cityIDs["moskva"]), Status: domain.PromotionActive, ImpressionCount: 18200, ClickCount: 941,
			CreatedAt: now.AddDate(0, 0, -8), UpdatedAt: now.Add(-3 * time.Hour),
		},
		{
			ID: uuid.MustParse("43000000-0000-0000-0000-000000000002"), BathhouseID: bh5, DailyBidKopecks: 12000,
			BudgetKopecks: 300000, SpentKopecks: 108000, StartDate: now.AddDate(0, 0, -10), EndDate: now.AddDate(0, 0, 14),
			TargetCityID: ptrInt64(cityIDs["moskva"]), Status: domain.PromotionActive, ImpressionCount: 25100, ClickCount: 1330,
			CreatedAt: now.AddDate(0, 0, -10), UpdatedAt: now.Add(-2 * time.Hour),
		},
		{
			ID: uuid.MustParse("43000000-0000-0000-0000-000000000003"), BathhouseID: bh8, DailyBidKopecks: 15000,
			BudgetKopecks: 420000, SpentKopecks: 165000, StartDate: now.AddDate(0, 0, -12), EndDate: now.AddDate(0, 0, 16),
			TargetCityID: ptrInt64(cityIDs["moskva"]), Status: domain.PromotionActive, ImpressionCount: 31200, ClickCount: 1842,
			CreatedAt: now.AddDate(0, 0, -12), UpdatedAt: now.Add(-1 * time.Hour),
		},
	}

	wallets := []demoWallet{
		{ID: uuid.MustParse("44000000-0000-0000-0000-000000000001"), UserID: owner1ID, Balance: 960000, HeldAmount: 0, Currency: domain.WalletCurrencyRUB, Status: domain.WalletStatusActive, CreatedAt: now.AddDate(0, -3, 0), UpdatedAt: now},
		{ID: uuid.MustParse("44000000-0000-0000-0000-000000000002"), UserID: owner2ID, Balance: 720000, HeldAmount: 0, Currency: domain.WalletCurrencyRUB, Status: domain.WalletStatusActive, CreatedAt: now.AddDate(0, -2, 0), UpdatedAt: now},
		{ID: uuid.MustParse("44000000-0000-0000-0000-000000000003"), UserID: client1ID, Balance: 145000, HeldAmount: 0, Currency: domain.WalletCurrencyRUB, Status: domain.WalletStatusActive, CreatedAt: now.AddDate(0, -2, -4), UpdatedAt: now},
		{ID: uuid.MustParse("44000000-0000-0000-0000-000000000004"), UserID: client2ID, Balance: 49000, HeldAmount: 0, Currency: domain.WalletCurrencyRUB, Status: domain.WalletStatusActive, CreatedAt: now.AddDate(0, -1, -16), UpdatedAt: now},
		{ID: uuid.MustParse("44000000-0000-0000-0000-000000000005"), UserID: client3ID, Balance: 52000, HeldAmount: 0, Currency: domain.WalletCurrencyRUB, Status: domain.WalletStatusActive, CreatedAt: now.AddDate(0, -1, -20), UpdatedAt: now},
		{ID: uuid.MustParse("44000000-0000-0000-0000-000000000006"), UserID: client4ID, Balance: 50000, HeldAmount: 0, Currency: domain.WalletCurrencyRUB, Status: domain.WalletStatusActive, CreatedAt: now.AddDate(0, 0, -2), UpdatedAt: now},
	}

	booking1Start, booking1End := slotAt(baseDay, 3, 19, 0, 3*time.Hour)
	booking2Start, booking2End := slotAt(baseDay, -9, 18, 0, 4*time.Hour)
	booking3Start, booking3End := slotAt(baseDay, -6, 17, 0, 5*time.Hour)
	booking4Start, booking4End := slotAt(baseDay, -13, 16, 0, 4*time.Hour)
	booking5Start, booking5End := slotAt(baseDay, 4, 15, 0, 3*time.Hour)
	booking6Start, booking6End := slotAt(baseDay, 6, 13, 0, 4*time.Hour)
	booking7Start, booking7End := slotAt(baseDay, -2, 19, 0, 3*time.Hour)
	booking8Start, booking8End := slotAt(baseDay, 8, 18, 0, 4*time.Hour)

	depositReleased := booking7End.Add(3 * time.Hour)
	review1ResponseAt := booking2End.Add(36 * time.Hour)
	review2ResponseAt := booking4End.Add(30 * time.Hour)
	review3ResponseAt := booking7End.Add(18 * time.Hour)
	reviewRevealAt := booking2End.Add(24 * time.Hour)
	disputeEvidenceDeadline := booking7End.Add(48 * time.Hour)
	disputeAppealDeadline := booking7End.Add(7 * 24 * time.Hour)
	ticketResolvedAt := now.AddDate(0, 0, -4)
	certificateValidUntil := now.AddDate(0, 6, 0)
	certificateUsedAt := booking4Start.Add(45 * time.Minute)

	booking1ID := uuid.MustParse("45000000-0000-0000-0000-000000000001")
	booking2ID := uuid.MustParse("45000000-0000-0000-0000-000000000002")
	booking3ID := uuid.MustParse("45000000-0000-0000-0000-000000000003")
	booking4ID := uuid.MustParse("45000000-0000-0000-0000-000000000004")
	booking5ID := uuid.MustParse("45000000-0000-0000-0000-000000000005")
	booking6ID := uuid.MustParse("45000000-0000-0000-0000-000000000006")
	booking7ID := uuid.MustParse("45000000-0000-0000-0000-000000000007")
	booking8ID := uuid.MustParse("45000000-0000-0000-0000-000000000008")

	bookings := []demoBooking{
		{
			ID: booking1ID, UserID: client1ID, BathhouseID: bh1, StartTime: booking1Start, EndTime: booking1End,
			GuestCount: 2, TotalPrice: rub(22470), BasePrice: rub(20700), LongSessionDiscount: 0, ExtraGuestSurcharge: 0, LastMinuteDiscount: 0,
			ServiceFeeAmount: rub(1770), Status: domain.BookingConfirmed, Comment: "Нужен приватный заезд и чай с травами.",
			CreatedAt: now.AddDate(0, 0, -2), UpdatedAt: now.Add(-90 * time.Minute),
		},
		{
			ID: booking2ID, UserID: client1ID, BathhouseID: bh2, StartTime: booking2Start, EndTime: booking2End,
			GuestCount: 4, TotalPrice: rub(36360), BasePrice: rub(33600), LongSessionDiscount: 0, ExtraGuestSurcharge: 0, LastMinuteDiscount: 0,
			ServiceFeeAmount: rub(2760), Status: domain.BookingCompleted, Comment: "Праздновали годовщину, все прошло спокойно.",
			CheckedInAt: ptrTime(booking2Start.Add(5 * time.Minute)), CheckedOutAt: ptrTime(booking2End.Add(-5 * time.Minute)),
			CreatedAt: now.AddDate(0, 0, -16), UpdatedAt: booking2End,
		},
		{
			ID: booking3ID, UserID: client2ID, BathhouseID: bh3, StartTime: booking3Start, EndTime: booking3End,
			GuestCount: 8, TotalPrice: rub(48300), BasePrice: rub(49000), LongSessionDiscount: rub(4900), ExtraGuestSurcharge: 0, LastMinuteDiscount: 0,
			ServiceFeeAmount: rub(4200), Status: domain.BookingCancelled, Comment: "Перенесли корпоратив на другую дату.",
			CreatedAt: now.AddDate(0, 0, -11), UpdatedAt: now.AddDate(0, 0, -5),
		},
		{
			ID: booking4ID, UserID: client3ID, BathhouseID: bh4, StartTime: booking4Start, EndTime: booking4End,
			GuestCount: 4, TotalPrice: rub(33440), BasePrice: rub(30400), LongSessionDiscount: 0, ExtraGuestSurcharge: 0, LastMinuteDiscount: 0,
			ServiceFeeAmount: rub(3040), Status: domain.BookingCompleted, Comment: "Брали как семейный подарок, понравился чан.",
			CheckedInAt: ptrTime(booking4Start.Add(2 * time.Minute)), CheckedOutAt: ptrTime(booking4End.Add(-8 * time.Minute)),
			CreatedAt: now.AddDate(0, 0, -18), UpdatedAt: booking4End,
		},
		{
			ID: booking5ID, UserID: client2ID, BathhouseID: bh7, StartTime: booking5Start, EndTime: booking5End,
			GuestCount: 3, TotalPrice: rub(23430), BasePrice: rub(21300), LongSessionDiscount: 0, ExtraGuestSurcharge: 0, LastMinuteDiscount: 0,
			ServiceFeeAmount: rub(2130), Status: domain.BookingPendingOwner, Comment: "Нужна ранняя подготовка купели ко времени заезда.",
			CreatedAt: now.AddDate(0, 0, -1), UpdatedAt: now.Add(-2 * time.Hour),
		},
		{
			ID: booking6ID, UserID: client3ID, BathhouseID: bh5, StartTime: booking6Start, EndTime: booking6End,
			GuestCount: 5, TotalPrice: rub(44080), BasePrice: rub(36800), LongSessionDiscount: 0, ExtraGuestSurcharge: rub(4000), LastMinuteDiscount: 0,
			ServiceFeeAmount: rub(3280), Status: domain.BookingConfirmed, Comment: "Нужен детский стул и халаты.",
			CreatedAt: now.AddDate(0, 0, -1), UpdatedAt: now.Add(-40 * time.Minute),
		},
		{
			ID: booking7ID, UserID: client2ID, BathhouseID: bh8, StartTime: booking7Start, EndTime: booking7End,
			GuestCount: 4, TotalPrice: rub(56760), BasePrice: rub(51600), LongSessionDiscount: 0, ExtraGuestSurcharge: 0, LastMinuteDiscount: 0,
			ServiceFeeAmount: rub(5160), DepositAmount: rub(12900), DepositStatus: domain.DepositReleased, DepositExternalID: "dep_demo_premium_dvor_01",
			DepositReleasedAt: &depositReleased, Status: domain.BookingCompleted, Comment: "Премиальный сценарий для подарочного визита.",
			CheckedInAt: ptrTime(booking7Start.Add(1 * time.Minute)), CheckedOutAt: ptrTime(booking7End.Add(-10 * time.Minute)),
			CreatedAt: now.AddDate(0, 0, -5), UpdatedAt: depositReleased,
		},
		{
			ID: booking8ID, UserID: client4ID, BathhouseID: bh10, StartTime: booking8Start, EndTime: booking8End,
			GuestCount: 2, TotalPrice: rub(26656), BasePrice: rub(27200), LongSessionDiscount: 0, ExtraGuestSurcharge: 0, LastMinuteDiscount: rub(2720),
			ServiceFeeAmount: rub(2176), Status: domain.BookingConfirmed, Comment: "Сделана через публичный checkout для smoke-сценария.",
			CreatedAt: now.AddDate(0, 0, -1), UpdatedAt: now.Add(-20 * time.Minute),
		},
	}

	payments := []demoPayment{
		{
			ID: uuid.MustParse("46000000-0000-0000-0000-000000000001"), BookingID: booking1ID, UserID: client1ID,
			Amount: rub(22470), Currency: "RUB", Status: domain.PaymentPending, Provider: "yookassa", ExternalID: "demo_pay_pending_01",
			PaymentMethod: domain.PaymentMethodCard, WalletAmount: 0, CardAmount: rub(22470), Metadata: map[string]string{"flow": "upcoming"},
			CreatedAt: now.AddDate(0, 0, -2), UpdatedAt: now.Add(-80 * time.Minute),
		},
		{
			ID: uuid.MustParse("46000000-0000-0000-0000-000000000002"), BookingID: booking2ID, UserID: client1ID,
			Amount: rub(36360), Currency: "RUB", Status: domain.PaymentSucceeded, Provider: "yookassa", ExternalID: "demo_pay_success_02",
			PaymentMethod: domain.PaymentMethodCombo, WalletAmount: 25000, CardAmount: rub(36110), Metadata: map[string]string{"flow": "combo"},
			CreatedAt: booking2Start.Add(-2 * time.Hour), UpdatedAt: booking2Start.Add(-90 * time.Minute),
		},
		{
			ID: uuid.MustParse("46000000-0000-0000-0000-000000000003"), BookingID: booking3ID, UserID: client2ID,
			Amount: rub(48300), Currency: "RUB", Status: domain.PaymentRefunded, Provider: "yookassa", ExternalID: "demo_pay_refund_03",
			PaymentMethod: domain.PaymentMethodCard, WalletAmount: 0, CardAmount: rub(48300), RefundAmount: rub(48300), RefundedAt: ptrTime(now.AddDate(0, 0, -5)),
			Metadata:  map[string]string{"flow": "refund"},
			CreatedAt: booking3Start.Add(-3 * time.Hour), UpdatedAt: now.AddDate(0, 0, -5),
		},
		{
			ID: uuid.MustParse("46000000-0000-0000-0000-000000000004"), BookingID: booking4ID, UserID: client3ID,
			Amount: rub(33440), Currency: "RUB", Status: domain.PaymentSucceeded, Provider: "yookassa", ExternalID: "demo_pay_success_04",
			PaymentMethod: domain.PaymentMethodCard, WalletAmount: 0, CardAmount: rub(33440), Metadata: map[string]string{"flow": "gift_certificate"},
			CreatedAt: booking4Start.Add(-90 * time.Minute), UpdatedAt: booking4Start.Add(-60 * time.Minute),
		},
		{
			ID: uuid.MustParse("46000000-0000-0000-0000-000000000005"), BookingID: booking6ID, UserID: client3ID,
			Amount: rub(44080), Currency: "RUB", Status: domain.PaymentPending, Provider: "yookassa", ExternalID: "demo_pay_pending_05",
			PaymentMethod: domain.PaymentMethodSBP, WalletAmount: 0, CardAmount: rub(44080), Metadata: map[string]string{"flow": "awaiting_payment"},
			CreatedAt: now.AddDate(0, 0, -1), UpdatedAt: now.Add(-30 * time.Minute),
		},
		{
			ID: uuid.MustParse("46000000-0000-0000-0000-000000000006"), BookingID: booking7ID, UserID: client2ID,
			Amount: rub(56760), Currency: "RUB", Status: domain.PaymentSucceeded, Provider: "yookassa", ExternalID: "demo_pay_success_06",
			PaymentMethod: domain.PaymentMethodCard, WalletAmount: 0, CardAmount: rub(56760), Metadata: map[string]string{"flow": "premium"},
			CreatedAt: booking7Start.Add(-120 * time.Minute), UpdatedAt: booking7Start.Add(-100 * time.Minute),
		},
		{
			ID: uuid.MustParse("46000000-0000-0000-0000-000000000007"), BookingID: booking8ID, UserID: client4ID,
			Amount: rub(26656), Currency: "RUB", Status: domain.PaymentPending, Provider: "yookassa", ExternalID: "demo_pay_pending_07",
			PaymentMethod: domain.PaymentMethodGooglePay, WalletAmount: 0, CardAmount: rub(26656), Metadata: map[string]string{"flow": "public_checkout"},
			CreatedAt: now.AddDate(0, 0, -1), UpdatedAt: now.Add(-15 * time.Minute),
		},
	}

	moderationScore1 := 0.97
	moderationScore2 := 0.94
	moderationScore3 := 0.99
	reviews := []demoReview{
		{
			ID: uuid.MustParse("47000000-0000-0000-0000-000000000001"), UserID: client1ID, BathhouseID: bh2, BookingID: booking2ID,
			Rating: 5, Cleanliness: 5, Accuracy: 5, Communication: 4.5, ValueForMoney: 4.5,
			Text:   "Чисто, спокойно и без лишней суеты. Бассейн прогрет, хозяева быстро отвечают, повторила бы такой визит без сомнений.",
			Status: domain.ReviewStatusApproved, OwnerResponse: "Спасибо, уже сохранили ваши пожелания по чаю к следующему визиту.",
			OwnerResponseAt: &review1ResponseAt, ModerationScore: &moderationScore1, ModerationFlags: []string{},
			Images: demoReviewImages(frontendURL, "review-couple"), RevealAt: &reviewRevealAt, IsRevealed: true,
			ModeratedBy: &adminID, ModeratedAt: &review1ResponseAt, CreatedAt: booking2End.Add(10 * time.Hour), UpdatedAt: review1ResponseAt,
		},
		{
			ID: uuid.MustParse("47000000-0000-0000-0000-000000000002"), UserID: client3ID, BathhouseID: bh4, BookingID: booking4ID,
			Rating: 5, Cleanliness: 4.5, Accuracy: 5, Communication: 5, ValueForMoney: 4.5,
			Text:   "Чан и вид из окна реально работают как на фото. Для подарочного семейного выезда место идеальное, детям тоже было удобно.",
			Status: domain.ReviewStatusApproved, OwnerResponse: "Благодарим, добавили ваш отзыв в подборку для семейных сценариев.",
			OwnerResponseAt: &review2ResponseAt, ModerationScore: &moderationScore2, ModerationFlags: []string{},
			Images: demoReviewImages(frontendURL, "review-family"), RevealAt: ptrTime(booking4End.Add(24 * time.Hour)), IsRevealed: true,
			ModeratedBy: &adminID, ModeratedAt: &review2ResponseAt, CreatedAt: booking4End.Add(12 * time.Hour), UpdatedAt: review2ResponseAt,
		},
		{
			ID: uuid.MustParse("47000000-0000-0000-0000-000000000003"), UserID: client2ID, BathhouseID: bh8, BookingID: booking7ID,
			Rating: 5, Cleanliness: 5, Accuracy: 5, Communication: 5, ValueForMoney: 5,
			Text:   "Премиальный сервис без показухи. Заезд прошел вовремя, пространство приватное, подарок с сертификатом сработал идеально.",
			Status: domain.ReviewStatusApproved, OwnerResponse: "Спасибо за точную обратную связь, рады что сценарий с сертификатом зашел.",
			OwnerResponseAt: &review3ResponseAt, ModerationScore: &moderationScore3, ModerationFlags: []string{},
			Images: demoReviewImages(frontendURL, "review-premium"), RevealAt: ptrTime(booking7End.Add(24 * time.Hour)), IsRevealed: true,
			ModeratedBy: &adminID, ModeratedAt: &review3ResponseAt, CreatedAt: booking7End.Add(8 * time.Hour), UpdatedAt: review3ResponseAt,
		},
	}
	syntheticBookings, syntheticReviews := buildSyntheticReviewBackfill(frontendURL, now, bathhouses, reviews, reviewerUsers, adminID)
	bookings = append(bookings, syntheticBookings...)
	reviews = append(reviews, syntheticReviews...)

	clientReviewScore := 0.95
	clientReviews := []demoClientReview{
		{
			ID: uuid.MustParse("47100000-0000-0000-0000-000000000001"), OwnerID: owner1ID, ClientID: client1ID, BookingID: booking2ID, BathhouseID: bh2,
			Punctuality: 5, Cleanliness: 5, RuleCompliance: 5, Rating: 5,
			Text:     "Приехала вовремя, соблюдала правила и оставила объект в идеальном состоянии.",
			RevealAt: booking2End.Add(24 * time.Hour), IsRevealed: true, ModerationScore: &clientReviewScore, ModerationFlags: []string{},
			Status: "approved", CreatedAt: booking2End.Add(6 * time.Hour), UpdatedAt: booking2End.Add(8 * time.Hour),
		},
	}

	certificate1ID := uuid.MustParse("48000000-0000-0000-0000-000000000001")
	certificate2ID := uuid.MustParse("48000000-0000-0000-0000-000000000002")
	certificates := []demoCertificate{
		{
			ID: certificate1ID, Code: "BANI-GIFT-001", PurchaserID: &client1ID, PurchaserEmail: "demo.client1@relax-hub.ru",
			RecipientEmail: "friend@example.com", RecipientName: "Ольга", Amount: rub(15000), Balance: rub(15000),
			Message: "Для спокойного выезда вдвоем без лишней суеты.", Status: "active", ValidUntil: certificateValidUntil, CreatedAt: now.AddDate(0, 0, -3),
		},
		{
			ID: certificate2ID, Code: "BANI-GIFT-002", PurchaserID: &client3ID, PurchaserEmail: "demo.client3@relax-hub.ru",
			RecipientEmail: "demo.client3@relax-hub.ru", RecipientName: "Екатерина", Amount: rub(10000), Balance: 0,
			Message: "Подарок на семейный уикенд.", Status: "used", ValidUntil: certificateValidUntil, RedeemedByID: &client3ID, CreatedAt: now.AddDate(0, 0, -20),
		},
	}

	certificateUsages := []demoCertificateUsage{
		{ID: uuid.MustParse("48100000-0000-0000-0000-000000000001"), CertificateID: certificate2ID, BookingID: booking4ID, Amount: rub(10000), UsedAt: certificateUsedAt},
	}

	favorites := []demoFavorite{
		{ID: uuid.MustParse("49000000-0000-0000-0000-000000000001"), UserID: client1ID, BathhouseID: bh1, CreatedAt: now.AddDate(0, 0, -7)},
		{ID: uuid.MustParse("49000000-0000-0000-0000-000000000002"), UserID: client1ID, BathhouseID: bh8, CreatedAt: now.AddDate(0, 0, -2)},
		{ID: uuid.MustParse("49000000-0000-0000-0000-000000000003"), UserID: client2ID, BathhouseID: bh3, CreatedAt: now.AddDate(0, 0, -9)},
		{ID: uuid.MustParse("49000000-0000-0000-0000-000000000004"), UserID: client2ID, BathhouseID: bh5, CreatedAt: now.AddDate(0, 0, -1)},
	}

	csati := 5
	tickets := []demoTicket{
		{
			ID: uuid.MustParse("50000000-0000-0000-0000-000000000001"), UserID: client1ID, BookingID: &booking1ID,
			Category: domain.TicketCategoryQuestion, Status: domain.TicketStatusOpen, Priority: domain.TicketPriorityLow, Level: domain.TicketLevelL1,
			Subject: "Подтвердите ранний заезд и чайную подачу", AssignedTo: &supportID, CreatedAt: now.Add(-5 * time.Hour), UpdatedAt: now.Add(-90 * time.Minute),
		},
		{
			ID: uuid.MustParse("50000000-0000-0000-0000-000000000002"), UserID: client2ID, BookingID: &booking3ID,
			Category: domain.TicketCategoryRefundReq, Status: domain.TicketStatusResolved, Priority: domain.TicketPriorityHigh, Level: domain.TicketLevelL2,
			Subject: "Возврат по отмененной корпоративной броне", AssignedTo: &supportID, CSATScore: &csati, CreatedAt: now.AddDate(0, 0, -6), UpdatedAt: ticketResolvedAt, ResolvedAt: &ticketResolvedAt,
		},
		{
			ID: uuid.MustParse("50000000-0000-0000-0000-000000000003"), UserID: client2ID, BookingID: &booking7ID,
			Category: domain.TicketCategoryComplaint, Status: domain.TicketStatusEscalated, Priority: domain.TicketPriorityHigh, Level: domain.TicketLevelL2,
			Subject: "Нужна проверка премиальной брони и компенсации", AssignedTo: &supportID, CreatedAt: now.AddDate(0, 0, -1), UpdatedAt: now.Add(-3 * time.Hour),
		},
	}

	ticketMessages := []demoTicketMessage{
		{
			ID: uuid.MustParse("50100000-0000-0000-0000-000000000001"), TicketID: uuid.MustParse("50000000-0000-0000-0000-000000000001"),
			SenderID: client1ID, SenderType: domain.TicketSenderUser, Body: "Привет. Заезжаем к 19:00, можно ли заранее подготовить чай и банные простыни?",
			CreatedAt: now.Add(-5 * time.Hour),
		},
		{
			ID: uuid.MustParse("50100000-0000-0000-0000-000000000002"), TicketID: uuid.MustParse("50000000-0000-0000-0000-000000000001"),
			SenderID: supportID, SenderType: domain.TicketSenderAdmin, Body: "Да, запрос уже передан владельцу. Подтвердим окончательно в течение 10 минут.",
			CreatedAt: now.Add(-3 * time.Hour),
		},
		{
			ID: uuid.MustParse("50100000-0000-0000-0000-000000000003"), TicketID: uuid.MustParse("50000000-0000-0000-0000-000000000002"),
			SenderID: client2ID, SenderType: domain.TicketSenderUser, Body: "Бронь отменили, нужен полный возврат по корпоративной оплате.",
			CreatedAt: now.AddDate(0, 0, -6),
		},
		{
			ID: uuid.MustParse("50100000-0000-0000-0000-000000000004"), TicketID: uuid.MustParse("50000000-0000-0000-0000-000000000002"),
			SenderID: supportID, SenderType: domain.TicketSenderAdmin, Body: "Возврат проведен, чек и детали уже в истории платежей.",
			CreatedAt: now.AddDate(0, 0, -4),
		},
		{
			ID: uuid.MustParse("50100000-0000-0000-0000-000000000005"), TicketID: uuid.MustParse("50000000-0000-0000-0000-000000000003"),
			SenderID: client2ID, SenderType: domain.TicketSenderUser, Body: "Хочу разобраться с депозитом и возможной компенсацией по премиальной броне.",
			CreatedAt: now.AddDate(0, 0, -1),
		},
	}

	resolution := domain.DisputeResolutionPartialRefund
	disputes := []demoDispute{
		{
			ID: uuid.MustParse("51000000-0000-0000-0000-000000000001"), BookingID: booking7ID, InitiatorID: client2ID, RespondentID: owner2ID,
			Reason: domain.DisputeReasonBillingError, Description: "Клиент считает, что часть депозита удержали без достаточного описания причин.",
			Status: domain.DisputeStatusUnderReview, MediatorID: &supportID, MediatorNotes: "Собираем подтверждение по депозиту и журнал заезда.",
			EvidenceDeadline: &disputeEvidenceDeadline, AppealDeadline: &disputeAppealDeadline, CreatedAt: now.AddDate(0, 0, -1), UpdatedAt: now.Add(-2 * time.Hour),
		},
		{
			ID: uuid.MustParse("51000000-0000-0000-0000-000000000002"), BookingID: booking3ID, InitiatorID: client2ID, RespondentID: owner1ID,
			Reason: domain.DisputeReasonOther, Description: "Исторический спор по старой отмене, закрыт частичным возвратом.",
			Status: domain.DisputeStatusResolved, Resolution: &resolution, RefundAmount: 20000, CompensationAmount: 0,
			MediatorID: &supportID, MediatorNotes: "Команда поддержки согласовала частичный возврат по правилам отмены.",
			EvidenceDeadline: ptrTime(now.AddDate(0, 0, -5)), AppealDeadline: ptrTime(now.AddDate(0, 0, 2)), CreatedAt: now.AddDate(0, 0, -5), UpdatedAt: now.AddDate(0, 0, -4),
			ResolvedAt: ptrTime(now.AddDate(0, 0, -4)),
		},
	}

	disputeEvidence := []demoDisputeEvidence{
		{
			ID: uuid.MustParse("51100000-0000-0000-0000-000000000001"), DisputeID: uuid.MustParse("51000000-0000-0000-0000-000000000001"),
			UserID: client2ID, Type: domain.DisputeEvidenceScreenshot, URL: "https://images.unsplash.com/photo-1517248135467-4c7edcad34c4?auto=format&fit=crop&w=900&q=80",
			Description: "Скрин коммуникации по удержанию депозита.", CreatedAt: now.AddDate(0, 0, -1),
		},
		{
			ID: uuid.MustParse("51100000-0000-0000-0000-000000000002"), DisputeID: uuid.MustParse("51000000-0000-0000-0000-000000000001"),
			UserID: owner2ID, Type: domain.DisputeEvidencePhoto, URL: "https://images.unsplash.com/photo-1506744038136-46273834b3fb?auto=format&fit=crop&w=900&q=80",
			Description: "Фото состояния объекта после выезда.", CreatedAt: now.Add(-8 * time.Hour),
		},
	}

	savedCards := []demoSavedCard{
		{
			ID: uuid.MustParse("52000000-0000-0000-0000-000000000001"), UserID: client1ID, ProviderToken: "card_demo_client1",
			Last4: "4242", Brand: "Visa", ExpiryMonth: 12, ExpiryYear: now.Year() + 2, IsDefault: true, CreatedAt: now.AddDate(0, -2, 0), UpdatedAt: now,
		},
		{
			ID: uuid.MustParse("52000000-0000-0000-0000-000000000002"), UserID: client2ID, ProviderToken: "card_demo_client2",
			Last4: "2200", Brand: "Mir", ExpiryMonth: 9, ExpiryYear: now.Year() + 3, IsDefault: true, CreatedAt: now.AddDate(0, -1, -12), UpdatedAt: now,
		},
	}

	shares := []demoShare{
		{
			ID: uuid.MustParse("53000000-0000-0000-0000-000000000001"), Token: "BANIWEEKENDMOSCOWSHARE0001", CreatedBy: owner1ID, BathhouseID: bh6,
			StartTime: booking6Start, EndTime: booking6End, GuestCount: 5, BookingID: &booking6ID, CreatedAt: now.AddDate(0, 0, -1), ExpiresAt: now.AddDate(0, 0, 14),
		},
	}

	faqItems := []demoFAQ{
		newFAQ("54000000-0000-0000-0000-000000000001", domain.FAQCategoryBooking, "Можно ли забронировать без регистрации?", "Да. На публичном checkout мы подтверждаем телефон по SMS и сразу создаем клиентский кабинет без отдельной формы регистрации.", []string{"checkout", "sms", "регистрация", "гость"}, 1, now),
		newFAQ("54000000-0000-0000-0000-000000000002", domain.FAQCategoryBooking, "Как выбрать баню на выходные?", "Откройте публичное меню и выберите пункт «На выходные» — каталог сразу подставит ближайшую субботу и покажет объекты с доступными слотами.", []string{"выходные", "суббота", "каталог"}, 2, now),
		newFAQ("54000000-0000-0000-0000-000000000003", domain.FAQCategoryPayment, "Когда оплачивается бронь?", "После подтверждения телефона бронь появляется в кабинете клиента. Там можно завершить оплату картой, СБП, Apple Pay, Google Pay или кошельком.", []string{"оплата", "сбп", "кошелек", "карта"}, 1, now),
		newFAQ("54000000-0000-0000-0000-000000000004", domain.FAQCategoryCancellation, "Как работает отмена?", "Политика отмены указана в карточке объекта. Для demo-мира мы используем гибкую, умеренную и строгую модели, чтобы можно было проверить разные сценарии возврата.", []string{"отмена", "возврат", "policy"}, 1, now),
		newFAQ("54000000-0000-0000-0000-000000000005", domain.FAQCategoryWallet, "Зачем нужен кошелек?", "Кошелек ускоряет повторные оплаты, хранит бонусы и возвраты. В demo-данных он уже пополнен у нескольких клиентов, чтобы проверить комбинированные платежи.", []string{"кошелек", "бонусы", "комбо"}, 1, now),
		newFAQ("54000000-0000-0000-0000-000000000006", domain.FAQCategoryGeneral, "Есть ли подарочные сертификаты?", "Да. Публичная страница сертификатов доступна без входа, а в demo-мире уже созданы активный и использованный сертификаты для проверки клиентских разделов.", []string{"сертификат", "подарок"}, 1, now),
	}

	walletTxs := []demoWalletTransaction{
		{
			ID: uuid.MustParse("55000000-0000-0000-0000-000000000001"), WalletID: wallets[2].ID, Type: domain.WalletTxWelcomeBonus, Amount: 50000,
			BalanceAfter: 50000, Status: domain.WalletTxStatusCompleted, Description: "Приветственный бонус нового клиента", IsBonus: true,
			ExpiresAt: ptrTime(now.AddDate(0, 1, 0)), CreatedAt: now.AddDate(0, -1, 0),
		},
		{
			ID: uuid.MustParse("55000000-0000-0000-0000-000000000002"), WalletID: wallets[2].ID, Type: domain.WalletTxTopUp, Amount: 120000,
			BalanceAfter: 170000, Status: domain.WalletTxStatusCompleted, Description: "Пополнение для быстрых повторных оплат", CreatedAt: now.AddDate(0, 0, -20),
		},
		{
			ID: uuid.MustParse("55000000-0000-0000-0000-000000000003"), WalletID: wallets[2].ID, Type: domain.WalletTxSpend, Amount: 25000,
			BalanceAfter: 145000, Status: domain.WalletTxStatusCompleted, Description: "Часть оплаты по комбинированному платежу", ReferenceType: "payment", ReferenceID: &payments[1].ID, CreatedAt: booking2Start.Add(-90 * time.Minute),
		},
		{
			ID: uuid.MustParse("55000000-0000-0000-0000-000000000004"), WalletID: wallets[3].ID, Type: domain.WalletTxRefund, Amount: 49000,
			BalanceAfter: 49000, Status: domain.WalletTxStatusCompleted, Description: "Возврат по отмененной корпоративной броне", ReferenceType: "payment", ReferenceID: &payments[2].ID, CreatedAt: now.AddDate(0, 0, -5),
		},
		{
			ID: uuid.MustParse("55000000-0000-0000-0000-000000000005"), WalletID: wallets[0].ID, Type: domain.WalletTxAdminCredit, Amount: 1200000,
			BalanceAfter: 1200000, Status: domain.WalletTxStatusCompleted, Description: "Демо-пополнение кошелька владельца", CreatedAt: now.AddDate(0, -2, 0),
		},
		{
			ID: uuid.MustParse("55000000-0000-0000-0000-000000000006"), WalletID: wallets[0].ID, Type: domain.WalletTxSpend, Amount: 240000,
			BalanceAfter: 960000, Status: domain.WalletTxStatusCompleted, Description: "Списания на продвижение и сервисные расходы", ReferenceType: "promotion", ReferenceID: &promotions[0].ID, CreatedAt: now.AddDate(0, 0, -2),
		},
	}

	notifications := []demoNotification{
		{
			ID: uuid.MustParse("56000000-0000-0000-0000-000000000001"), UserID: client1ID, Type: domain.NotifBookingConfirmed,
			Title: "Бронь подтверждена", Body: "Тихий Берег подтвердил ваш вечерний слот на пятницу.", Data: map[string]string{"booking_id": booking1ID.String()}, IsRead: false, CreatedAt: now.Add(-80 * time.Minute),
		},
		{
			ID: uuid.MustParse("56000000-0000-0000-0000-000000000002"), UserID: client1ID, Type: domain.NotifPromo,
			Title: "Сценарий для двоих", Body: "Для вас собрана подборка с чаном и спокойным чек-аутом в один экран.", Data: map[string]string{"bathhouse_id": bh1.String()}, IsRead: true, ReadAt: ptrTime(now.AddDate(0, 0, -1)), CreatedAt: now.AddDate(0, 0, -1),
		},
		{
			ID: uuid.MustParse("56000000-0000-0000-0000-000000000003"), UserID: owner1ID, Type: domain.NotifPromo,
			Title: "Кампания дает лиды", Body: "Пар и Бассейн показывает лучший CTR среди объектов owner-кабинета.", Data: map[string]string{"promotion_id": promotions[1].ID.String()}, IsRead: false, CreatedAt: now.Add(-3 * time.Hour),
		},
		{
			ID: uuid.MustParse("56000000-0000-0000-0000-000000000004"), UserID: client2ID, Type: domain.NotifSystem,
			Title: "Открыт спор по депозиту", Body: "Поддержка приняла обращение и собирает подтверждение по брони Премиум Двор.", Data: map[string]string{"dispute_id": disputes[0].ID.String()}, IsRead: false, CreatedAt: now.Add(-7 * time.Hour),
		},
	}

	adminNotifications := []demoAdminNotification{
		{
			ID: uuid.MustParse("57000000-0000-0000-0000-000000000001"), Role: domain.AdminSubRoleSuperAdmin, Severity: domain.AdminNotifSeverityWarning,
			Type: domain.AdminNotifDisputeOpened, Title: "Открыт спор по депозиту", Body: "Клиент оспаривает удержание депозита по премиальной броне.", Data: map[string]string{"dispute_id": disputes[0].ID.String()}, IsRead: false, CreatedAt: now.Add(-8 * time.Hour),
		},
		{
			ID: uuid.MustParse("57000000-0000-0000-0000-000000000002"), Role: domain.AdminSubRoleSupportL2, Severity: domain.AdminNotifSeverityInfo,
			Type: domain.AdminNotifTicketEscalation, Title: "Эскалация тикета о возврате", Body: "Открыт L2 тикет с жалобой на премиальную бронь.", Data: map[string]string{"ticket_id": tickets[2].ID.String()}, IsRead: false, CreatedAt: now.Add(-6 * time.Hour),
		},
		{
			ID: uuid.MustParse("57000000-0000-0000-0000-000000000003"), Role: domain.AdminSubRoleFinance, Severity: domain.AdminNotifSeverityError,
			Type: domain.AdminNotifReconciliationMismatch, Title: "Проверьте тестовую сверку платежей", Body: "В demo-мире оставлен кейс с возвратом и депозитом для проверки финансовых экранов.", Data: map[string]string{"payment_id": payments[2].ID.String()}, IsRead: false, CreatedAt: now.AddDate(0, 0, -1),
		},
	}

	for _, city := range cities {
		if err := upsertCity(ctx, tx, city); err != nil {
			return err
		}
	}

	for _, user := range users {
		if err := upsertUser(ctx, tx, user, passwordHash, now); err != nil {
			return err
		}
		if err := upsertNotificationPreferences(ctx, tx, user.ID); err != nil {
			return err
		}
		if err := upsertNotificationEventPreferences(ctx, tx, user.ID); err != nil {
			return err
		}
	}

	for _, wallet := range wallets {
		if err := upsertWallet(ctx, tx, wallet); err != nil {
			return err
		}
	}
	for _, walletTx := range walletTxs {
		if err := upsertWalletTransaction(ctx, tx, walletTx); err != nil {
			return err
		}
	}

	for _, bathhouse := range bathhouses {
		if err := upsertBathhouse(ctx, tx, bathhouse); err != nil {
			return err
		}
	}
	for _, rep := range repAssignments {
		if err := upsertRepresentative(ctx, tx, rep); err != nil {
			return err
		}
	}
	for _, subscription := range subscriptions {
		if err := upsertSubscription(ctx, tx, subscription); err != nil {
			return err
		}
	}
	for _, promotion := range promotions {
		if err := upsertPromotion(ctx, tx, promotion); err != nil {
			return err
		}
	}
	for _, favorite := range favorites {
		if err := upsertFavorite(ctx, tx, favorite); err != nil {
			return err
		}
	}
	for _, booking := range bookings {
		if err := upsertBooking(ctx, tx, booking); err != nil {
			return err
		}
	}
	for _, payment := range payments {
		if err := upsertPayment(ctx, tx, payment); err != nil {
			return err
		}
	}
	for _, review := range reviews {
		if err := upsertReview(ctx, tx, review); err != nil {
			return err
		}
	}
	for _, review := range clientReviews {
		if err := upsertClientReview(ctx, tx, review); err != nil {
			return err
		}
	}
	for _, certificate := range certificates {
		if err := upsertCertificate(ctx, tx, certificate); err != nil {
			return err
		}
	}
	for _, usage := range certificateUsages {
		if err := upsertCertificateUsage(ctx, tx, usage); err != nil {
			return err
		}
	}
	for _, notification := range notifications {
		if err := upsertNotification(ctx, tx, notification); err != nil {
			return err
		}
	}
	for _, notification := range adminNotifications {
		if err := upsertAdminNotification(ctx, tx, notification); err != nil {
			return err
		}
	}
	for _, ticket := range tickets {
		if err := upsertTicket(ctx, tx, ticket); err != nil {
			return err
		}
	}
	for _, message := range ticketMessages {
		if err := upsertTicketMessage(ctx, tx, message); err != nil {
			return err
		}
	}
	for _, dispute := range disputes {
		if err := upsertDispute(ctx, tx, dispute); err != nil {
			return err
		}
	}
	for _, evidence := range disputeEvidence {
		if err := upsertDisputeEvidence(ctx, tx, evidence); err != nil {
			return err
		}
	}
	for _, card := range savedCards {
		if err := upsertSavedCard(ctx, tx, card); err != nil {
			return err
		}
	}
	for _, share := range shares {
		if err := upsertBookingShare(ctx, tx, share); err != nil {
			return err
		}
	}
	for _, faq := range faqItems {
		if err := upsertFAQ(ctx, tx, faq); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit demo seed tx: %w", err)
	}
	return nil
}

func upsertCity(ctx context.Context, tx pgx.Tx, city demoCity) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO cities (id, name, slug, latitude, longitude, region)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			slug = EXCLUDED.slug,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			region = EXCLUDED.region
	`, city.ID, city.Name, city.Slug, city.Latitude, city.Longitude, city.Region)
	if err != nil {
		return fmt.Errorf("upsert city %s: %w", city.Slug, err)
	}
	return nil
}

func upsertUser(ctx context.Context, tx pgx.Tx, user demoUser, passwordHash string, now time.Time) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO users (
			id, email, password_hash, name, phone, phone_verified, role, admin_sub_role, is_active,
			avatar_url, bio, city_id, region, referral_code, totp_secret, two_fa_method,
			onboarding_completed, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, true, $6, $7, true,
			$8, $9, $10, $11, $12, '', $13,
			$14, $15, $16
		)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			password_hash = EXCLUDED.password_hash,
			name = EXCLUDED.name,
			phone = EXCLUDED.phone,
			phone_verified = EXCLUDED.phone_verified,
			role = EXCLUDED.role,
			admin_sub_role = EXCLUDED.admin_sub_role,
			is_active = EXCLUDED.is_active,
			avatar_url = EXCLUDED.avatar_url,
			bio = EXCLUDED.bio,
			city_id = EXCLUDED.city_id,
			region = EXCLUDED.region,
			referral_code = EXCLUDED.referral_code,
			two_fa_method = EXCLUDED.two_fa_method,
			onboarding_completed = EXCLUDED.onboarding_completed,
			updated_at = EXCLUDED.updated_at
	`,
		user.ID, user.Email, passwordHash, user.Name, user.Phone, string(user.Role), string(user.AdminSubRole),
		user.AvatarURL, user.Bio, user.CityID, string(user.Region), user.ReferralCode, string(domain.TwoFANone),
		user.OnboardingCompleted, now.AddDate(0, -2, 0), now,
	)
	if err != nil {
		return fmt.Errorf("upsert user %s: %w", user.Email, err)
	}
	return nil
}

func upsertNotificationPreferences(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO notification_preferences (user_id, in_app, email, push, telegram, sms, booking_events, review_events, promo_events, reminders)
		VALUES ($1, true, true, true, false, false, true, true, true, true)
		ON CONFLICT (user_id) DO UPDATE SET
			in_app = EXCLUDED.in_app,
			email = EXCLUDED.email,
			push = EXCLUDED.push,
			telegram = EXCLUDED.telegram,
			sms = EXCLUDED.sms,
			booking_events = EXCLUDED.booking_events,
			review_events = EXCLUDED.review_events,
			promo_events = EXCLUDED.promo_events,
			reminders = EXCLUDED.reminders
	`, userID)
	if err != nil {
		return fmt.Errorf("upsert notification preferences for %s: %w", userID, err)
	}
	return nil
}

func upsertNotificationEventPreferences(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	eventTypes := []domain.NotificationType{
		domain.NotifBookingConfirmed,
		domain.NotifBookingReminder24h,
		domain.NotifPromo,
		domain.NotifSystem,
	}
	for _, eventType := range eventTypes {
		_, err := tx.Exec(ctx, `
			INSERT INTO notification_event_preferences (user_id, event_type, push_enabled, email_enabled, sms_enabled)
			VALUES ($1, $2, true, true, false)
			ON CONFLICT (user_id, event_type) DO UPDATE SET
				push_enabled = EXCLUDED.push_enabled,
				email_enabled = EXCLUDED.email_enabled,
				sms_enabled = EXCLUDED.sms_enabled
		`, userID, string(eventType))
		if err != nil {
			return fmt.Errorf("upsert notification event preference %s for %s: %w", eventType, userID, err)
		}
	}
	return nil
}

func upsertWallet(ctx context.Context, tx pgx.Tx, wallet demoWallet) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO wallets (id, user_id, balance, held_amount, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			balance = EXCLUDED.balance,
			held_amount = EXCLUDED.held_amount,
			currency = EXCLUDED.currency,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`, wallet.ID, wallet.UserID, wallet.Balance, wallet.HeldAmount, string(wallet.Currency), string(wallet.Status), wallet.CreatedAt, wallet.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert wallet %s: %w", wallet.ID, err)
	}
	return nil
}

func upsertWalletTransaction(ctx context.Context, tx pgx.Tx, transaction demoWalletTransaction) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO wallet_transactions (
			id, wallet_id, type, amount, balance_after, status, description, reference_type, reference_id, is_bonus, expires_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			wallet_id = EXCLUDED.wallet_id,
			type = EXCLUDED.type,
			amount = EXCLUDED.amount,
			balance_after = EXCLUDED.balance_after,
			status = EXCLUDED.status,
			description = EXCLUDED.description,
			reference_type = EXCLUDED.reference_type,
			reference_id = EXCLUDED.reference_id,
			is_bonus = EXCLUDED.is_bonus,
			expires_at = EXCLUDED.expires_at,
			created_at = EXCLUDED.created_at
	`, transaction.ID, transaction.WalletID, string(transaction.Type), transaction.Amount, transaction.BalanceAfter,
		string(transaction.Status), transaction.Description, transaction.ReferenceType, transaction.ReferenceID, transaction.IsBonus,
		transaction.ExpiresAt, transaction.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert wallet transaction %s: %w", transaction.ID, err)
	}
	return nil
}

func upsertBathhouse(ctx context.Context, tx pgx.Tx, bathhouse demoBathhouse) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO bathhouses (
			id, owner_id, name, slug, description, address, city_id, latitude, longitude,
			price_per_hour, min_duration, max_guests, has_pool, has_sauna, has_steam_room,
			has_hot_tub, has_bbq, has_karaoke, rating, bayesian_rating, review_count,
			conversion_rate, occupancy_rate, view_count, images, working_hours, status,
			long_session_threshold_hours, long_session_discount_percent, base_capacity, extra_guest_surcharge,
			last_minute_enabled, last_minute_discount_percent, last_minute_hours_threshold,
			buffer_minutes, lead_time_hours, max_advance_days, booking_mode, request_timeout,
			response_rate, avg_response_time_minutes, is_photo_verified, api_key, cancellation_policy,
			security_deposit_percent, calendar_token, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21,
			$22, $23, $24, $25::jsonb, $26::jsonb, $27,
			$28, $29, $30, $31,
			$32, $33, $34,
			$35, $36, $37, $38, $39,
			$40, $41, $42, $43, $44,
			$45, $46, $47, $48
		)
		ON CONFLICT (id) DO UPDATE SET
			owner_id = EXCLUDED.owner_id,
			name = EXCLUDED.name,
			slug = EXCLUDED.slug,
			description = EXCLUDED.description,
			address = EXCLUDED.address,
			city_id = EXCLUDED.city_id,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			price_per_hour = EXCLUDED.price_per_hour,
			min_duration = EXCLUDED.min_duration,
			max_guests = EXCLUDED.max_guests,
			has_pool = EXCLUDED.has_pool,
			has_sauna = EXCLUDED.has_sauna,
			has_steam_room = EXCLUDED.has_steam_room,
			has_hot_tub = EXCLUDED.has_hot_tub,
			has_bbq = EXCLUDED.has_bbq,
			has_karaoke = EXCLUDED.has_karaoke,
			rating = EXCLUDED.rating,
			bayesian_rating = EXCLUDED.bayesian_rating,
			review_count = EXCLUDED.review_count,
			conversion_rate = EXCLUDED.conversion_rate,
			occupancy_rate = EXCLUDED.occupancy_rate,
			view_count = EXCLUDED.view_count,
			images = EXCLUDED.images,
			working_hours = EXCLUDED.working_hours,
			status = EXCLUDED.status,
			long_session_threshold_hours = EXCLUDED.long_session_threshold_hours,
			long_session_discount_percent = EXCLUDED.long_session_discount_percent,
			base_capacity = EXCLUDED.base_capacity,
			extra_guest_surcharge = EXCLUDED.extra_guest_surcharge,
			last_minute_enabled = EXCLUDED.last_minute_enabled,
			last_minute_discount_percent = EXCLUDED.last_minute_discount_percent,
			last_minute_hours_threshold = EXCLUDED.last_minute_hours_threshold,
			buffer_minutes = EXCLUDED.buffer_minutes,
			lead_time_hours = EXCLUDED.lead_time_hours,
			max_advance_days = EXCLUDED.max_advance_days,
			booking_mode = EXCLUDED.booking_mode,
			request_timeout = EXCLUDED.request_timeout,
			response_rate = EXCLUDED.response_rate,
			avg_response_time_minutes = EXCLUDED.avg_response_time_minutes,
			is_photo_verified = EXCLUDED.is_photo_verified,
			api_key = EXCLUDED.api_key,
			cancellation_policy = EXCLUDED.cancellation_policy,
			security_deposit_percent = EXCLUDED.security_deposit_percent,
			calendar_token = EXCLUDED.calendar_token,
			updated_at = EXCLUDED.updated_at
	`,
		bathhouse.ID, bathhouse.OwnerID, bathhouse.Name, bathhouse.Slug, bathhouse.Description, bathhouse.Address, bathhouse.CityID,
		bathhouse.Latitude, bathhouse.Longitude, bathhouse.PricePerHour, bathhouse.MinDuration, bathhouse.MaxGuests,
		bathhouse.HasPool, bathhouse.HasSauna, bathhouse.HasSteamRoom, bathhouse.HasHotTub, bathhouse.HasBBQ, bathhouse.HasKaraoke,
		bathhouse.Rating, bathhouse.BayesianRating, bathhouse.ReviewCount, bathhouse.ConversionRate, bathhouse.OccupancyRate, bathhouse.ViewCount,
		mustJSON(bathhouse.Images), mustJSON(bathhouse.WorkingHours), string(bathhouse.Status), bathhouse.LongSessionThresholdHours,
		bathhouse.LongSessionDiscountPercent, bathhouse.BaseCapacity, bathhouse.ExtraGuestSurcharge, bathhouse.LastMinuteEnabled,
		bathhouse.LastMinuteDiscountPercent, bathhouse.LastMinuteHoursThreshold, bathhouse.BufferMinutes, bathhouse.LeadTimeHours,
		bathhouse.MaxAdvanceDays, bathhouse.BookingMode, bathhouse.RequestTimeout, bathhouse.ResponseRate, bathhouse.AvgResponseTimeMinutes,
		bathhouse.IsPhotoVerified, bathhouse.ApiKey, string(bathhouse.CancellationPolicy), bathhouse.SecurityDepositPercent,
		bathhouse.CalendarToken, bathhouse.CreatedAt, bathhouse.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert bathhouse %s: %w", bathhouse.Slug, err)
	}
	return nil
}

func upsertRepresentative(ctx context.Context, tx pgx.Tx, representative demoRepresentative) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO representatives (id, user_id, bathhouse_id, owner_id, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			bathhouse_id = EXCLUDED.bathhouse_id,
			owner_id = EXCLUDED.owner_id,
			role = EXCLUDED.role,
			created_at = EXCLUDED.created_at
	`, representative.ID, representative.UserID, representative.BathhouseID, representative.OwnerID, string(representative.Role), representative.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert representative %s: %w", representative.ID, err)
	}
	return nil
}

func upsertSubscription(ctx context.Context, tx pgx.Tx, subscription demoSubscription) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO subscriptions (id, bathhouse_id, owner_id, plan, status, start_date, end_date, auto_renew, price_kopecks, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			bathhouse_id = EXCLUDED.bathhouse_id,
			owner_id = EXCLUDED.owner_id,
			plan = EXCLUDED.plan,
			status = EXCLUDED.status,
			start_date = EXCLUDED.start_date,
			end_date = EXCLUDED.end_date,
			auto_renew = EXCLUDED.auto_renew,
			price_kopecks = EXCLUDED.price_kopecks,
			updated_at = EXCLUDED.updated_at
	`, subscription.ID, subscription.BathhouseID, subscription.OwnerID, string(subscription.Plan), string(subscription.Status),
		subscription.StartDate, subscription.EndDate, subscription.AutoRenew, subscription.PriceKopecks, subscription.CreatedAt, subscription.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert subscription %s: %w", subscription.ID, err)
	}
	return nil
}

func upsertPromotion(ctx context.Context, tx pgx.Tx, promotion demoPromotion) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO promotions (
			id, bathhouse_id, daily_bid_kopecks, budget_kopecks, spent_kopecks, start_date, end_date, target_city_id, status, impression_count, click_count, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			bathhouse_id = EXCLUDED.bathhouse_id,
			daily_bid_kopecks = EXCLUDED.daily_bid_kopecks,
			budget_kopecks = EXCLUDED.budget_kopecks,
			spent_kopecks = EXCLUDED.spent_kopecks,
			start_date = EXCLUDED.start_date,
			end_date = EXCLUDED.end_date,
			target_city_id = EXCLUDED.target_city_id,
			status = EXCLUDED.status,
			impression_count = EXCLUDED.impression_count,
			click_count = EXCLUDED.click_count,
			updated_at = EXCLUDED.updated_at
	`, promotion.ID, promotion.BathhouseID, promotion.DailyBidKopecks, promotion.BudgetKopecks, promotion.SpentKopecks,
		promotion.StartDate, promotion.EndDate, promotion.TargetCityID, string(promotion.Status), promotion.ImpressionCount, promotion.ClickCount,
		promotion.CreatedAt, promotion.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert promotion %s: %w", promotion.ID, err)
	}
	return nil
}

func upsertFavorite(ctx context.Context, tx pgx.Tx, favorite demoFavorite) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO favorites (id, user_id, bathhouse_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			bathhouse_id = EXCLUDED.bathhouse_id,
			created_at = EXCLUDED.created_at
	`, favorite.ID, favorite.UserID, favorite.BathhouseID, favorite.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert favorite %s: %w", favorite.ID, err)
	}
	return nil
}

func upsertBooking(ctx context.Context, tx pgx.Tx, booking demoBooking) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO bookings (
			id, user_id, bathhouse_id, start_time, end_time, guest_count, total_price, addon_total, points_spent, referral_bonus_used,
			base_price, long_session_discount, extra_guest_surcharge, last_minute_discount, service_fee_amount, modification_count,
			deposit_amount, deposit_status, deposit_external_id, deposit_released_at, checked_in_at, checked_out_at, hold_id,
			rejection_reason, cancelled_by_owner, status, comment, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23,
			$24, $25, $26, $27, $28, $29
		)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			bathhouse_id = EXCLUDED.bathhouse_id,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			guest_count = EXCLUDED.guest_count,
			total_price = EXCLUDED.total_price,
			addon_total = EXCLUDED.addon_total,
			points_spent = EXCLUDED.points_spent,
			referral_bonus_used = EXCLUDED.referral_bonus_used,
			base_price = EXCLUDED.base_price,
			long_session_discount = EXCLUDED.long_session_discount,
			extra_guest_surcharge = EXCLUDED.extra_guest_surcharge,
			last_minute_discount = EXCLUDED.last_minute_discount,
			service_fee_amount = EXCLUDED.service_fee_amount,
			modification_count = EXCLUDED.modification_count,
			deposit_amount = EXCLUDED.deposit_amount,
			deposit_status = EXCLUDED.deposit_status,
			deposit_external_id = EXCLUDED.deposit_external_id,
			deposit_released_at = EXCLUDED.deposit_released_at,
			checked_in_at = EXCLUDED.checked_in_at,
			checked_out_at = EXCLUDED.checked_out_at,
			hold_id = EXCLUDED.hold_id,
			rejection_reason = EXCLUDED.rejection_reason,
			cancelled_by_owner = EXCLUDED.cancelled_by_owner,
			status = EXCLUDED.status,
			comment = EXCLUDED.comment,
			updated_at = EXCLUDED.updated_at
	`,
		booking.ID, booking.UserID, booking.BathhouseID, booking.StartTime, booking.EndTime, booking.GuestCount, booking.TotalPrice,
		booking.AddOnTotal, booking.PointsSpent, booking.ReferralBonusUsed, booking.BasePrice, booking.LongSessionDiscount,
		booking.ExtraGuestSurcharge, booking.LastMinuteDiscount, booking.ServiceFeeAmount, booking.ModificationCount, booking.DepositAmount,
		string(booking.DepositStatus), booking.DepositExternalID, booking.DepositReleasedAt, booking.CheckedInAt, booking.CheckedOutAt,
		booking.HoldID, booking.RejectionReason, booking.CancelledByOwner, string(booking.Status), booking.Comment, booking.CreatedAt, booking.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert booking %s: %w", booking.ID, err)
	}
	return nil
}

func upsertPayment(ctx context.Context, tx pgx.Tx, payment demoPayment) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO payments (
			id, booking_id, user_id, amount, currency, status, provider, external_id, refund_amount, refunded_at, metadata,
			payment_method, wallet_amount, card_amount, is_hold, captured_at, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9, $10, $11::jsonb,
			$12, $13, $14, $15, $16, $17, $18
		)
		ON CONFLICT (id) DO UPDATE SET
			booking_id = EXCLUDED.booking_id,
			user_id = EXCLUDED.user_id,
			amount = EXCLUDED.amount,
			currency = EXCLUDED.currency,
			status = EXCLUDED.status,
			provider = EXCLUDED.provider,
			external_id = EXCLUDED.external_id,
			refund_amount = EXCLUDED.refund_amount,
			refunded_at = EXCLUDED.refunded_at,
			metadata = EXCLUDED.metadata,
			payment_method = EXCLUDED.payment_method,
			wallet_amount = EXCLUDED.wallet_amount,
			card_amount = EXCLUDED.card_amount,
			is_hold = EXCLUDED.is_hold,
			captured_at = EXCLUDED.captured_at,
			updated_at = EXCLUDED.updated_at
	`,
		payment.ID, payment.BookingID, payment.UserID, payment.Amount, payment.Currency, string(payment.Status), payment.Provider, payment.ExternalID,
		payment.RefundAmount, payment.RefundedAt, mustJSON(payment.Metadata), string(payment.PaymentMethod), payment.WalletAmount, payment.CardAmount,
		payment.IsHold, payment.CapturedAt, payment.CreatedAt, payment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert payment %s: %w", payment.ID, err)
	}
	return nil
}

func upsertReview(ctx context.Context, tx pgx.Tx, review demoReview) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO reviews (
			id, user_id, bathhouse_id, booking_id, rating, cleanliness, accuracy, communication, value_for_money, text, status,
			rejection_reasons, owner_response, owner_response_at, moderation_score, moderation_flags, images, reveal_at, is_revealed,
			moderated_by, moderated_at, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17::jsonb, $18, $19,
			$20, $21, $22, $23
		)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			bathhouse_id = EXCLUDED.bathhouse_id,
			booking_id = EXCLUDED.booking_id,
			rating = EXCLUDED.rating,
			cleanliness = EXCLUDED.cleanliness,
			accuracy = EXCLUDED.accuracy,
			communication = EXCLUDED.communication,
			value_for_money = EXCLUDED.value_for_money,
			text = EXCLUDED.text,
			status = EXCLUDED.status,
			rejection_reasons = EXCLUDED.rejection_reasons,
			owner_response = EXCLUDED.owner_response,
			owner_response_at = EXCLUDED.owner_response_at,
			moderation_score = EXCLUDED.moderation_score,
			moderation_flags = EXCLUDED.moderation_flags,
			images = EXCLUDED.images,
			reveal_at = EXCLUDED.reveal_at,
			is_revealed = EXCLUDED.is_revealed,
			moderated_by = EXCLUDED.moderated_by,
			moderated_at = EXCLUDED.moderated_at,
			updated_at = EXCLUDED.updated_at
	`,
		review.ID, review.UserID, review.BathhouseID, review.BookingID, review.Rating, review.Cleanliness, review.Accuracy,
		review.Communication, review.ValueForMoney, review.Text, string(review.Status), stringSliceOrEmpty(review.RejectionReasons), review.OwnerResponse,
		review.OwnerResponseAt, review.ModerationScore, stringSliceOrEmpty(review.ModerationFlags), mustJSON(review.Images), review.RevealAt, review.IsRevealed,
		review.ModeratedBy, review.ModeratedAt, review.CreatedAt, review.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert review %s: %w", review.ID, err)
	}
	return nil
}

func upsertClientReview(ctx context.Context, tx pgx.Tx, review demoClientReview) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO client_reviews (
			id, owner_id, client_id, booking_id, bathhouse_id, punctuality, cleanliness, rule_compliance, rating, text,
			reveal_at, is_revealed, moderation_score, moderation_flags, status, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17
		)
		ON CONFLICT (id) DO UPDATE SET
			owner_id = EXCLUDED.owner_id,
			client_id = EXCLUDED.client_id,
			booking_id = EXCLUDED.booking_id,
			bathhouse_id = EXCLUDED.bathhouse_id,
			punctuality = EXCLUDED.punctuality,
			cleanliness = EXCLUDED.cleanliness,
			rule_compliance = EXCLUDED.rule_compliance,
			rating = EXCLUDED.rating,
			text = EXCLUDED.text,
			reveal_at = EXCLUDED.reveal_at,
			is_revealed = EXCLUDED.is_revealed,
			moderation_score = EXCLUDED.moderation_score,
			moderation_flags = EXCLUDED.moderation_flags,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`,
		review.ID, review.OwnerID, review.ClientID, review.BookingID, review.BathhouseID, review.Punctuality, review.Cleanliness,
		review.RuleCompliance, review.Rating, review.Text, review.RevealAt, review.IsRevealed, review.ModerationScore, stringSliceOrEmpty(review.ModerationFlags),
		review.Status, review.CreatedAt, review.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert client review %s: %w", review.ID, err)
	}
	return nil
}

func upsertCertificate(ctx context.Context, tx pgx.Tx, certificate demoCertificate) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO gift_certificates (
			id, code, purchaser_id, purchaser_email, recipient_email, recipient_name, amount, balance, message, status, valid_until, redeemed_by_id, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			code = EXCLUDED.code,
			purchaser_id = EXCLUDED.purchaser_id,
			purchaser_email = EXCLUDED.purchaser_email,
			recipient_email = EXCLUDED.recipient_email,
			recipient_name = EXCLUDED.recipient_name,
			amount = EXCLUDED.amount,
			balance = EXCLUDED.balance,
			message = EXCLUDED.message,
			status = EXCLUDED.status,
			valid_until = EXCLUDED.valid_until,
			redeemed_by_id = EXCLUDED.redeemed_by_id,
			created_at = EXCLUDED.created_at
	`, certificate.ID, certificate.Code, certificate.PurchaserID, certificate.PurchaserEmail, certificate.RecipientEmail,
		certificate.RecipientName, certificate.Amount, certificate.Balance, certificate.Message, certificate.Status, certificate.ValidUntil,
		certificate.RedeemedByID, certificate.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert certificate %s: %w", certificate.Code, err)
	}
	return nil
}

func upsertCertificateUsage(ctx context.Context, tx pgx.Tx, usage demoCertificateUsage) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO certificate_usages (id, certificate_id, booking_id, amount, used_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			certificate_id = EXCLUDED.certificate_id,
			booking_id = EXCLUDED.booking_id,
			amount = EXCLUDED.amount,
			used_at = EXCLUDED.used_at
	`, usage.ID, usage.CertificateID, usage.BookingID, usage.Amount, usage.UsedAt)
	if err != nil {
		return fmt.Errorf("upsert certificate usage %s: %w", usage.ID, err)
	}
	return nil
}

func upsertNotification(ctx context.Context, tx pgx.Tx, notification demoNotification) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO notifications (id, user_id, type, title, body, data, is_read, read_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			type = EXCLUDED.type,
			title = EXCLUDED.title,
			body = EXCLUDED.body,
			data = EXCLUDED.data,
			is_read = EXCLUDED.is_read,
			read_at = EXCLUDED.read_at,
			created_at = EXCLUDED.created_at
	`, notification.ID, notification.UserID, string(notification.Type), notification.Title, notification.Body,
		mustJSON(notification.Data), notification.IsRead, notification.ReadAt, notification.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert notification %s: %w", notification.ID, err)
	}
	return nil
}

func upsertAdminNotification(ctx context.Context, tx pgx.Tx, notification demoAdminNotification) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO admin_notifications (id, role, severity, type, title, body, data, is_read, read_at, read_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			role = EXCLUDED.role,
			severity = EXCLUDED.severity,
			type = EXCLUDED.type,
			title = EXCLUDED.title,
			body = EXCLUDED.body,
			data = EXCLUDED.data,
			is_read = EXCLUDED.is_read,
			read_at = EXCLUDED.read_at,
			read_by = EXCLUDED.read_by,
			created_at = EXCLUDED.created_at
	`, notification.ID, string(notification.Role), string(notification.Severity), string(notification.Type), notification.Title, notification.Body,
		mustJSON(notification.Data), notification.IsRead, notification.ReadAt, notification.ReadBy, notification.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert admin notification %s: %w", notification.ID, err)
	}
	return nil
}

func upsertTicket(ctx context.Context, tx pgx.Tx, ticket demoTicket) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO support_tickets (
			id, user_id, booking_id, category, status, priority, level, subject, assigned_to, csat_score, created_at, updated_at, resolved_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			booking_id = EXCLUDED.booking_id,
			category = EXCLUDED.category,
			status = EXCLUDED.status,
			priority = EXCLUDED.priority,
			level = EXCLUDED.level,
			subject = EXCLUDED.subject,
			assigned_to = EXCLUDED.assigned_to,
			csat_score = EXCLUDED.csat_score,
			updated_at = EXCLUDED.updated_at,
			resolved_at = EXCLUDED.resolved_at
	`, ticket.ID, ticket.UserID, ticket.BookingID, string(ticket.Category), string(ticket.Status), string(ticket.Priority),
		string(ticket.Level), ticket.Subject, ticket.AssignedTo, ticket.CSATScore, ticket.CreatedAt, ticket.UpdatedAt, ticket.ResolvedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert ticket %s: %w", ticket.ID, err)
	}
	return nil
}

func upsertTicketMessage(ctx context.Context, tx pgx.Tx, message demoTicketMessage) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO ticket_messages (id, ticket_id, sender_id, sender_type, body, attachments, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			ticket_id = EXCLUDED.ticket_id,
			sender_id = EXCLUDED.sender_id,
			sender_type = EXCLUDED.sender_type,
			body = EXCLUDED.body,
			attachments = EXCLUDED.attachments,
			created_at = EXCLUDED.created_at
	`, message.ID, message.TicketID, message.SenderID, string(message.SenderType), message.Body, stringSliceOrEmpty(message.Attachments), message.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert ticket message %s: %w", message.ID, err)
	}
	return nil
}

func upsertDispute(ctx context.Context, tx pgx.Tx, dispute demoDispute) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO disputes (
			id, booking_id, initiator_id, respondent_id, reason, description, status, resolution, refund_amount, compensation_amount,
			mediator_id, mediator_notes, appeal_deadline, evidence_deadline, created_at, updated_at, resolved_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17
		)
		ON CONFLICT (id) DO UPDATE SET
			booking_id = EXCLUDED.booking_id,
			initiator_id = EXCLUDED.initiator_id,
			respondent_id = EXCLUDED.respondent_id,
			reason = EXCLUDED.reason,
			description = EXCLUDED.description,
			status = EXCLUDED.status,
			resolution = EXCLUDED.resolution,
			refund_amount = EXCLUDED.refund_amount,
			compensation_amount = EXCLUDED.compensation_amount,
			mediator_id = EXCLUDED.mediator_id,
			mediator_notes = EXCLUDED.mediator_notes,
			appeal_deadline = EXCLUDED.appeal_deadline,
			evidence_deadline = EXCLUDED.evidence_deadline,
			updated_at = EXCLUDED.updated_at,
			resolved_at = EXCLUDED.resolved_at
	`, dispute.ID, dispute.BookingID, dispute.InitiatorID, dispute.RespondentID, string(dispute.Reason), dispute.Description,
		string(dispute.Status), dispute.Resolution, dispute.RefundAmount, dispute.CompensationAmount, dispute.MediatorID, dispute.MediatorNotes,
		dispute.AppealDeadline, dispute.EvidenceDeadline, dispute.CreatedAt, dispute.UpdatedAt, dispute.ResolvedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert dispute %s: %w", dispute.ID, err)
	}
	return nil
}

func upsertDisputeEvidence(ctx context.Context, tx pgx.Tx, evidence demoDisputeEvidence) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO dispute_evidence (id, dispute_id, user_id, type, url, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			dispute_id = EXCLUDED.dispute_id,
			user_id = EXCLUDED.user_id,
			type = EXCLUDED.type,
			url = EXCLUDED.url,
			description = EXCLUDED.description,
			created_at = EXCLUDED.created_at
	`, evidence.ID, evidence.DisputeID, evidence.UserID, string(evidence.Type), evidence.URL, evidence.Description, evidence.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert dispute evidence %s: %w", evidence.ID, err)
	}
	return nil
}

func upsertSavedCard(ctx context.Context, tx pgx.Tx, card demoSavedCard) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO saved_cards (id, user_id, provider_token, last4, brand, expiry_month, expiry_year, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			provider_token = EXCLUDED.provider_token,
			last4 = EXCLUDED.last4,
			brand = EXCLUDED.brand,
			expiry_month = EXCLUDED.expiry_month,
			expiry_year = EXCLUDED.expiry_year,
			is_default = EXCLUDED.is_default,
			updated_at = EXCLUDED.updated_at
	`, card.ID, card.UserID, card.ProviderToken, card.Last4, card.Brand, card.ExpiryMonth, card.ExpiryYear, card.IsDefault, card.CreatedAt, card.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert saved card %s: %w", card.ID, err)
	}
	return nil
}

func upsertBookingShare(ctx context.Context, tx pgx.Tx, share demoShare) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO booking_shares (id, token, created_by, bathhouse_id, start_time, end_time, guest_count, booking_id, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			token = EXCLUDED.token,
			created_by = EXCLUDED.created_by,
			bathhouse_id = EXCLUDED.bathhouse_id,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			guest_count = EXCLUDED.guest_count,
			booking_id = EXCLUDED.booking_id,
			created_at = EXCLUDED.created_at,
			expires_at = EXCLUDED.expires_at
	`, share.ID, share.Token, share.CreatedBy, share.BathhouseID, share.StartTime, share.EndTime, share.GuestCount, share.BookingID, share.CreatedAt, share.ExpiresAt)
	if err != nil {
		return fmt.Errorf("upsert booking share %s: %w", share.ID, err)
	}
	return nil
}

func upsertFAQ(ctx context.Context, tx pgx.Tx, faq demoFAQ) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO faq (id, category, question, answer, keywords, sort_order, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			category = EXCLUDED.category,
			question = EXCLUDED.question,
			answer = EXCLUDED.answer,
			keywords = EXCLUDED.keywords,
			sort_order = EXCLUDED.sort_order,
			active = EXCLUDED.active,
			updated_at = EXCLUDED.updated_at
	`, faq.ID, string(faq.Category), faq.Question, faq.Answer, stringSliceOrEmpty(faq.Keywords), faq.SortOrder, faq.Active, faq.CreatedAt, faq.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert faq %s: %w", faq.ID, err)
	}
	return nil
}

func loadMoscowLocation() *time.Location {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.FixedZone("MSK", 3*60*60)
	}
	return location
}

func defaultWorkingHours() []domain.WorkingHours {
	return []domain.WorkingHours{
		{DayOfWeek: 0, OpenTime: "10:00", CloseTime: "23:00"},
		{DayOfWeek: 1, OpenTime: "10:00", CloseTime: "23:00"},
		{DayOfWeek: 2, OpenTime: "10:00", CloseTime: "23:00"},
		{DayOfWeek: 3, OpenTime: "10:00", CloseTime: "23:00"},
		{DayOfWeek: 4, OpenTime: "10:00", CloseTime: "01:00"},
		{DayOfWeek: 5, OpenTime: "09:00", CloseTime: "01:00"},
		{DayOfWeek: 6, OpenTime: "09:00", CloseTime: "23:00"},
	}
}

func demoBathhouseImages(frontendURL, theme string) []string {
	assetSet := map[string][]string{
		"banya-couple": {
			"kupel-courtyard-evening.png",
			"steam-room-birch.png",
			"tea-lounge-samovar.png",
			"log-bathhouse-entrance.png",
		},
		"banya-pool": {
			"premium-indoor-pool.png",
			"steam-room-birch.png",
			"premium-courtyard-spa.png",
			"sauna-benches-steam.png",
		},
		"banya-company": {
			"group-getaway-log-house.png",
			"steam-room-birch.png",
			"tea-lounge-samovar.png",
			"log-bathhouse-entrance.png",
		},
		"banya-hottub": {
			"log-sauna-kupel-forest.png",
			"kupel-courtyard-evening.png",
			"steam-room-birch.png",
			"premium-courtyard-spa.png",
		},
		"banya-premium-pool": {
			"premium-indoor-pool.png",
			"premium-courtyard-spa.png",
			"tea-lounge-samovar.png",
			"sauna-benches-steam.png",
		},
		"banya-weekend": {
			"family-bathhouse-yard.png",
			"group-getaway-log-house.png",
			"tea-lounge-samovar.png",
			"log-bathhouse-entrance.png",
		},
		"banya-request": {
			"log-sauna-kupel-forest.png",
			"steam-room-birch.png",
			"kupel-courtyard-evening.png",
			"log-bathhouse-entrance.png",
		},
		"banya-premium": {
			"premium-courtyard-spa.png",
			"premium-indoor-pool.png",
			"kupel-courtyard-evening.png",
			"tea-lounge-samovar.png",
		},
		"banya-family": {
			"family-bathhouse-yard.png",
			"tea-lounge-samovar.png",
			"sauna-benches-steam.png",
			"log-bathhouse-entrance.png",
		},
		"banya-river": {
			"riverside-log-bathhouse.png",
			"steam-room-birch.png",
			"log-bathhouse-entrance.png",
			"tea-lounge-samovar.png",
		},
		"banya-pending": {
			"log-bathhouse-entrance.png",
			"steam-room-birch.png",
			"sauna-benches-steam.png",
		},
		"banya-rejected": {
			"group-getaway-log-house.png",
			"steam-room-birch.png",
			"tea-lounge-samovar.png",
		},
	}

	assets := assetSet[theme]
	if len(assets) == 0 {
		assets = []string{
			"kupel-courtyard-evening.png",
			"steam-room-birch.png",
			"tea-lounge-samovar.png",
			"log-bathhouse-entrance.png",
		}
	}
	return demoBathhouseAssetURLs(frontendURL, assets)
}

func demoReviewImages(frontendURL, seed string) []string {
	assets := []string{
		"steam-room-birch.png",
		"log-sauna-kupel-forest.png",
		"premium-indoor-pool.png",
		"tea-lounge-samovar.png",
		"sauna-benches-steam.png",
		"log-bathhouse-entrance.png",
	}

	offset := 0
	for _, char := range []byte(seed) {
		offset += int(char)
	}

	first := assets[offset%len(assets)]
	second := assets[(offset+3)%len(assets)]
	if second == first {
		second = assets[(offset+1)%len(assets)]
	}

	return demoBathhouseAssetURLs(frontendURL, []string{first, second})
}

func demoBathhouseAssetURLs(_ string, assets []string) []string {
	urls := make([]string, 0, len(assets))
	for _, asset := range assets {
		urls = append(urls, fmt.Sprintf("/demo/bathhouses/%s", asset))
	}
	return urls
}

func slotAt(base time.Time, dayOffset, hour, minute int, duration time.Duration) (time.Time, time.Time) {
	start := time.Date(base.Year(), base.Month(), base.Day()+dayOffset, hour, minute, 0, 0, base.Location())
	return start, start.Add(duration)
}

func newSubscription(id string, bathhouseID, ownerID uuid.UUID, plan domain.SubscriptionPlan, startDate, endDate time.Time, autoRenew bool, priceKopecks int64) demoSubscription {
	return demoSubscription{
		ID:           uuid.MustParse(id),
		BathhouseID:  bathhouseID,
		OwnerID:      ownerID,
		Plan:         plan,
		Status:       domain.SubscriptionActive,
		StartDate:    startDate,
		EndDate:      &endDate,
		AutoRenew:    autoRenew,
		PriceKopecks: priceKopecks,
		CreatedAt:    startDate,
		UpdatedAt:    time.Now(),
	}
}

func newFAQ(id string, category domain.FAQCategory, question, answer string, keywords []string, sortOrder int, now time.Time) demoFAQ {
	return demoFAQ{
		ID:        uuid.MustParse(id),
		Category:  category,
		Question:  question,
		Answer:    answer,
		Keywords:  keywords,
		SortOrder: sortOrder,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func clearDemoCaches(ctx context.Context, cfg *config.Config) error {
	if cfg.Redis.Addr == "" {
		return nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer client.Close()

	keys, err := client.Keys(ctx, "area_avg_price:*").Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return client.Del(ctx, keys...).Err()
}

func rub(value int64) int64 {
	return value * 100
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func ptrInt64(value int64) *int64 {
	return &value
}

func buildSyntheticReviewUsers(cityIDs map[string]int64) []demoUser {
	names := []string{
		"Никита Орлов", "Мария Белова", "Денис Климов", "Ольга Соколова",
		"Павел Нестеров", "Ирина Журавлева", "Глеб Смирнов", "Дарья Корнеева",
		"Виктор Лебедев", "Надежда Миронова", "Лев Громов", "Елена Давыдова",
	}
	cityOrder := []int64{cityIDs["moskva"], cityIDs["odintsovo"], cityIDs["krasnogorsk"]}

	users := make([]demoUser, 0, len(names))
	for index, name := range names {
		userID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(fmt.Sprintf("demo-reviewer:%02d", index+1)))
		cityID := cityOrder[index%len(cityOrder)]
		users = append(users, demoUser{
			ID:                  userID,
			Email:               fmt.Sprintf("demo.reviewer%02d@relax-hub.ru", index+1),
			Name:                name,
			Phone:               fmt.Sprintf("+799900001%02d", index+1),
			Role:                domain.RoleClient,
			CityID:              ptrInt64(cityID),
			Region:              domain.RegionRU,
			AvatarURL:           fmt.Sprintf("https://i.pravatar.cc/240?img=%d", 60+index),
			Bio:                 "Технический демо-аккаунт для исторических отзывов и social proof в публичной витрине.",
			ReferralCode:        fmt.Sprintf("REVIEW%02d", index+1),
			OnboardingCompleted: true,
		})
	}

	return users
}

func buildSyntheticReviewBackfill(
	frontendURL string,
	now time.Time,
	bathhouses []demoBathhouse,
	existingReviews []demoReview,
	reviewerUsers []demoUser,
	adminID uuid.UUID,
) ([]demoBooking, []demoReview) {
	if len(reviewerUsers) == 0 {
		return nil, nil
	}

	existingCount := make(map[uuid.UUID]int, len(bathhouses))
	for _, review := range existingReviews {
		if review.Status == domain.ReviewStatusApproved {
			existingCount[review.BathhouseID]++
		}
	}

	reviewTexts := []string{
		"Бронь прошла спокойно: внутри чисто, пар ровный, хозяин заранее прислал понятные инструкции по заезду.",
		"Выбирали без долгих переписок. Фото совпали с реальностью, место подготовили вовремя, отдых получился без лишнего шума.",
		"Хороший вариант для вечернего сценария: понятный вход, аккуратная зона отдыха и комфортная температура во всех помещениях.",
		"Брали слот на несколько часов подряд. Нигде не подгоняли, объект был готов к приезду, по сервису всё предсказуемо.",
		"Для компании формат оказался удобным: легко припарковались, внутри было чисто, коммуникация с владельцем без задержек.",
		"Выезд получился именно таким, как ожидали по карточке. Особенно понравились приватность, чистота и внятные правила посещения.",
	}
	ownerResponses := []string{
		"Спасибо, сохранили ваши замечания и уже внесли их в подготовку следующих визитов.",
		"Благодарим за отзыв. Команда площадки отметила ваши комментарии по температуре и сервису.",
		"Спасибо за визит. Приятно, что сценарий совпал с ожиданиями и карточка объекта не подвела.",
	}

	bookings := make([]demoBooking, 0)
	reviews := make([]demoReview, 0)

	for bathhouseIndex, bathhouse := range bathhouses {
		if bathhouse.Status != domain.BathhouseStatusActive || bathhouse.ReviewCount <= 0 {
			continue
		}

		missing := bathhouse.ReviewCount - existingCount[bathhouse.ID]
		if missing <= 0 {
			continue
		}

		for offset := 0; offset < missing; offset++ {
			ordinal := existingCount[bathhouse.ID] + offset + 1
			reviewer := reviewerUsers[(bathhouseIndex+offset)%len(reviewerUsers)]
			durationHours := maxInt(bathhouse.MinDuration, 2+(ordinal%3))
			if bathhouse.LongSessionThresholdHours > 0 && ordinal%5 == 0 {
				durationHours = maxInt(durationHours, bathhouse.LongSessionThresholdHours)
			}
			guestCount := minInt(
				bathhouse.MaxGuests,
				maxInt(2, bathhouse.BaseCapacity+((ordinal+bathhouseIndex)%3)-1),
			)

			daysAgo := 45 + bathhouseIndex*11 + offset*2
			startTime := time.Date(
				now.Year(),
				now.Month(),
				now.Day()-daysAgo,
				16+((ordinal+bathhouseIndex)%4),
				0,
				0,
				0,
				now.Location(),
			)
			endTime := startTime.Add(time.Duration(durationHours) * time.Hour)

			basePrice := bathhouse.PricePerHour * int64(durationHours)
			longSessionDiscount := int64(0)
			if bathhouse.LongSessionThresholdHours > 0 && durationHours >= bathhouse.LongSessionThresholdHours {
				longSessionDiscount = basePrice * int64(bathhouse.LongSessionDiscountPercent) / 100
			}
			extraGuests := maxInt(0, guestCount-bathhouse.BaseCapacity)
			extraGuestSurcharge := int64(extraGuests) * bathhouse.ExtraGuestSurcharge * int64(durationHours)
			netPrice := basePrice - longSessionDiscount + extraGuestSurcharge
			serviceFee := netPrice / 10
			totalPrice := netPrice + serviceFee
			depositAmount := int64(0)
			depositStatus := domain.DepositNone
			var depositReleasedAt *time.Time
			if bathhouse.SecurityDepositPercent > 0 {
				depositAmount = netPrice * int64(bathhouse.SecurityDepositPercent) / 100
				depositStatus = domain.DepositReleased
				depositReleasedAt = ptrTime(endTime.Add(48 * time.Hour))
			}

			bookingID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(fmt.Sprintf("demo-review-booking:%s:%03d", bathhouse.ID, ordinal)))
			reviewID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(fmt.Sprintf("demo-review:%s:%03d", bathhouse.ID, ordinal)))
			createdAt := endTime.Add(time.Duration(3+(ordinal%9)) * time.Hour)
			updatedAt := createdAt.Add(2 * time.Hour)
			var ownerResponse string
			var ownerResponseAt *time.Time
			if ordinal%2 == 0 {
				ownerResponse = ownerResponses[(bathhouseIndex+offset)%len(ownerResponses)]
				ownerResponseAt = ptrTime(createdAt.Add(12 * time.Hour))
				updatedAt = ownerResponseAt.Add(90 * time.Minute)
			}

			ratingBase := bathhouse.Rating
			if ratingBase <= 0 {
				ratingBase = 4.7
			}
			rating := int(ratingBase + 0.3)
			if rating < 4 {
				rating = 4
			}
			if rating > 5 {
				rating = 5
			}
			score := clampReviewCriterion(ratingBase)
			accuracyScore := clampReviewCriterion(score + 0.5)
			communicationScore := clampReviewCriterion(score + 0.5)
			valueScore := clampReviewCriterion(score)
			moderationScore := 0.93 + float64((ordinal+bathhouseIndex)%6)*0.01
			if moderationScore > 0.99 {
				moderationScore = 0.99
			}

			bookings = append(bookings, demoBooking{
				ID:                  bookingID,
				UserID:              reviewer.ID,
				BathhouseID:         bathhouse.ID,
				StartTime:           startTime,
				EndTime:             endTime,
				GuestCount:          guestCount,
				TotalPrice:          totalPrice,
				BasePrice:           basePrice,
				LongSessionDiscount: longSessionDiscount,
				ExtraGuestSurcharge: extraGuestSurcharge,
				LastMinuteDiscount:  0,
				ServiceFeeAmount:    serviceFee,
				DepositAmount:       depositAmount,
				DepositStatus:       depositStatus,
				DepositReleasedAt:   depositReleasedAt,
				CheckedInAt:         ptrTime(startTime.Add(5 * time.Minute)),
				CheckedOutAt:        ptrTime(endTime.Add(-10 * time.Minute)),
				Status:              domain.BookingCompleted,
				Comment:             "Историческая демо-бронь для наполнения карточки объекта и social proof.",
				CreatedAt:           startTime.Add(-72 * time.Hour),
				UpdatedAt:           updatedAt,
			})
			reviews = append(reviews, demoReview{
				ID:              reviewID,
				UserID:          reviewer.ID,
				BathhouseID:     bathhouse.ID,
				BookingID:       bookingID,
				Rating:          rating,
				Cleanliness:     score,
				Accuracy:        accuracyScore,
				Communication:   communicationScore,
				ValueForMoney:   valueScore,
				Text:            reviewTexts[(bathhouseIndex+offset)%len(reviewTexts)],
				Status:          domain.ReviewStatusApproved,
				OwnerResponse:   ownerResponse,
				OwnerResponseAt: ownerResponseAt,
				ModerationScore: &moderationScore,
				ModerationFlags: []string{},
				Images:          reviewImagesForOrdinal(frontendURL, bathhouse.Slug, ordinal),
				RevealAt:        ptrTime(endTime.Add(24 * time.Hour)),
				IsRevealed:      true,
				ModeratedBy:     &adminID,
				ModeratedAt:     ptrTime(createdAt.Add(90 * time.Minute)),
				CreatedAt:       createdAt,
				UpdatedAt:       updatedAt,
			})
		}
	}

	return bookings, reviews
}

func reviewImagesForOrdinal(frontendURL, seed string, ordinal int) []string {
	if ordinal%3 != 0 {
		return []string{}
	}
	return demoReviewImages(frontendURL, fmt.Sprintf("%s-%03d", seed, ordinal))
}

func mustJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func stringSliceOrEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func minFloat(left, right float64) float64 {
	if left < right {
		return left
	}
	return right
}

func clampReviewCriterion(value float64) float64 {
	clamped := minFloat(5, value)
	if clamped < 1 {
		clamped = 1
	}
	return float64(int(clamped*2+0.5)) / 2
}
