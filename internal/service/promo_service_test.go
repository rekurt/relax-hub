package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

type promoTestEnv struct {
	svc       service.PromoService
	promoRepo *mock.PromoCodeRepo
	addonRepo *mock.AddOnRepo
	bhRepo    *mock.BathhouseRepo
	repRepo   *mock.RepresentativeRepo
}

func newPromoTestEnv() *promoTestEnv {
	promoRepo := mock.NewPromoCodeRepo().(*mock.PromoCodeRepo)
	addonRepo := mock.NewAddOnRepo()
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	accessCheck := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	svc := service.NewPromoService(promoRepo, addonRepo, accessCheck, log)
	return &promoTestEnv{
		svc:       svc,
		promoRepo: promoRepo,
		addonRepo: addonRepo,
		bhRepo:    bhRepo,
		repRepo:   repRepo,
	}
}

func createPromoBathhouse(env *promoTestEnv, ownerID uuid.UUID) *domain.Bathhouse {
	bh := &domain.Bathhouse{
		ID:           uuid.New(),
		OwnerID:      ownerID,
		Name:         "Test Bathhouse",
		Address:      "Test Address",
		CityID:       1,
		PricePerHour: 500000,
		MinDuration:  1,
		MaxGuests:    10,
		Status:       domain.BathhouseStatusActive,
	}
	if err := env.bhRepo.Create(context.Background(), bh); err != nil {
		panic("setup: " + err.Error())
	}
	return bh
}

func validPromoCode(bathhouseID *uuid.UUID) *domain.PromoCode {
	return &domain.PromoCode{
		Code:        "SUMMER20",
		Type:        domain.PromoTypePercentage,
		Value:       20,
		BathhouseID: bathhouseID,
		MaxUses:     100,
		MinAmount:   100000,
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidUntil:  time.Now().Add(24 * time.Hour),
	}
}

// --- Create ---

func TestPromoService_Create_OwnerSuccess(t *testing.T) {
	env := newPromoTestEnv()
	ownerID := uuid.New()
	bh := createPromoBathhouse(env, ownerID)

	promo := validPromoCode(&bh.ID)
	result, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, promo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if result.Code != "SUMMER20" {
		t.Errorf("code = %q, want SUMMER20", result.Code)
	}
	if !result.IsActive {
		t.Error("expected IsActive = true")
	}
	if result.CreatorID != ownerID {
		t.Error("creator ID mismatch")
	}
}

func TestPromoService_Create_AdminGlobal(t *testing.T) {
	env := newPromoTestEnv()
	adminID := uuid.New()

	promo := validPromoCode(nil)
	result, err := env.svc.Create(context.Background(), adminID, domain.RoleAdmin, promo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.BathhouseID != nil {
		t.Error("expected nil BathhouseID for global promo")
	}
}

func TestPromoService_Create_NonAdminGlobalForbidden(t *testing.T) {
	env := newPromoTestEnv()
	ownerID := uuid.New()

	promo := validPromoCode(nil)
	_, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, promo)
	if err != domain.ErrForbidden {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestPromoService_Create_OwnerOtherBathhouseForbidden(t *testing.T) {
	env := newPromoTestEnv()
	ownerID := uuid.New()
	otherOwnerID := uuid.New()
	bh := createPromoBathhouse(env, otherOwnerID)

	promo := validPromoCode(&bh.ID)
	_, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, promo)
	if err != domain.ErrForbidden {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestPromoService_Create_InvalidInput(t *testing.T) {
	env := newPromoTestEnv()
	adminID := uuid.New()

	promo := &domain.PromoCode{
		Code:       "",
		Type:       domain.PromoTypePercentage,
		Value:      20,
		ValidFrom:  time.Now(),
		ValidUntil: time.Now().Add(time.Hour),
	}
	_, err := env.svc.Create(context.Background(), adminID, domain.RoleAdmin, promo)
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestPromoService_Create_CodeUppercased(t *testing.T) {
	env := newPromoTestEnv()
	adminID := uuid.New()

	promo := validPromoCode(nil)
	promo.Code = "  summer20  "
	result, err := env.svc.Create(context.Background(), adminID, domain.RoleAdmin, promo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Code != "SUMMER20" {
		t.Errorf("code = %q, want SUMMER20", result.Code)
	}
}

// --- Validate ---

func TestPromoService_Validate_Success(t *testing.T) {
	env := newPromoTestEnv()
	adminID := uuid.New()

	promo := validPromoCode(nil)
	promo.Value = 10 // 10%
	_, err := env.svc.Create(context.Background(), adminID, domain.RoleAdmin, promo)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	result, discount, err := env.svc.Validate(context.Background(), "SUMMER20", uuid.New(), 200000)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if result.Code != "SUMMER20" {
		t.Errorf("code = %q, want SUMMER20", result.Code)
	}
	if discount != 20000 { // 10% of 200000
		t.Errorf("discount = %d, want 20000", discount)
	}
}

func TestPromoService_Validate_FixedAmount(t *testing.T) {
	env := newPromoTestEnv()
	adminID := uuid.New()

	promo := validPromoCode(nil)
	promo.Code = "FIXED500"
	promo.Type = domain.PromoTypeFixedAmount
	promo.Value = 50000 // 500 rubles
	promo.MinAmount = 0
	_, err := env.svc.Create(context.Background(), adminID, domain.RoleAdmin, promo)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, discount, err := env.svc.Validate(context.Background(), "FIXED500", uuid.New(), 200000)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if discount != 50000 {
		t.Errorf("discount = %d, want 50000", discount)
	}
}

func TestPromoService_Validate_FixedAmountCappedAtTotal(t *testing.T) {
	env := newPromoTestEnv()
	adminID := uuid.New()

	promo := validPromoCode(nil)
	promo.Code = "BIGFIX"
	promo.Type = domain.PromoTypeFixedAmount
	promo.Value = 500000
	promo.MinAmount = 0
	_, err := env.svc.Create(context.Background(), adminID, domain.RoleAdmin, promo)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, discount, err := env.svc.Validate(context.Background(), "BIGFIX", uuid.New(), 200000)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if discount != 200000 {
		t.Errorf("discount = %d, want 200000 (capped at amount)", discount)
	}
}

func TestPromoService_Validate_FreeHour(t *testing.T) {
	env := newPromoTestEnv()
	adminID := uuid.New()

	promo := validPromoCode(nil)
	promo.Code = "FREEHOUR"
	promo.Type = domain.PromoTypeFreeHour
	promo.Value = 300000 // hourly rate
	promo.MinAmount = 0
	_, err := env.svc.Create(context.Background(), adminID, domain.RoleAdmin, promo)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, discount, err := env.svc.Validate(context.Background(), "FREEHOUR", uuid.New(), 900000)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if discount != 300000 {
		t.Errorf("discount = %d, want 300000", discount)
	}
}

func TestPromoService_Validate_Expired(t *testing.T) {
	env := newPromoTestEnv()

	// Directly insert an expired promo
	promo := &domain.PromoCode{
		ID:          uuid.New(),
		Code:        "EXPIRED",
		Type:        domain.PromoTypePercentage,
		Value:       10,
		CreatorID:   uuid.New(),
		IsActive:    true,
		ValidFrom:   time.Now().Add(-48 * time.Hour),
		ValidUntil:  time.Now().Add(-24 * time.Hour),
		CreatedAt:   time.Now(),
	}
	if err := env.promoRepo.Create(context.Background(), promo); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, _, err := env.svc.Validate(context.Background(), "EXPIRED", uuid.New(), 200000)
	if err != domain.ErrPromoExpired {
		t.Errorf("err = %v, want ErrPromoExpired", err)
	}
}

func TestPromoService_Validate_MaxUsesReached(t *testing.T) {
	env := newPromoTestEnv()

	promo := &domain.PromoCode{
		ID:          uuid.New(),
		Code:        "MAXED",
		Type:        domain.PromoTypePercentage,
		Value:       10,
		CreatorID:   uuid.New(),
		MaxUses:     1,
		CurrentUses: 1,
		IsActive:    true,
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidUntil:  time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
	}
	if err := env.promoRepo.Create(context.Background(), promo); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, _, err := env.svc.Validate(context.Background(), "MAXED", uuid.New(), 200000)
	if err != domain.ErrPromoMaxUses {
		t.Errorf("err = %v, want ErrPromoMaxUses", err)
	}
}

func TestPromoService_Validate_MinAmountNotMet(t *testing.T) {
	env := newPromoTestEnv()
	adminID := uuid.New()

	promo := validPromoCode(nil)
	promo.MinAmount = 500000 // min 5000 rubles
	_, err := env.svc.Create(context.Background(), adminID, domain.RoleAdmin, promo)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, _, err = env.svc.Validate(context.Background(), "SUMMER20", uuid.New(), 200000)
	if err != domain.ErrPromoMinAmount {
		t.Errorf("err = %v, want ErrPromoMinAmount", err)
	}
}

func TestPromoService_Validate_Inactive(t *testing.T) {
	env := newPromoTestEnv()

	promo := &domain.PromoCode{
		ID:          uuid.New(),
		Code:        "INACTIVE",
		Type:        domain.PromoTypePercentage,
		Value:       10,
		CreatorID:   uuid.New(),
		IsActive:    false,
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidUntil:  time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
	}
	if err := env.promoRepo.Create(context.Background(), promo); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, _, err := env.svc.Validate(context.Background(), "INACTIVE", uuid.New(), 200000)
	if err != domain.ErrPromoInvalid {
		t.Errorf("err = %v, want ErrPromoInvalid", err)
	}
}

func TestPromoService_Validate_NotFound(t *testing.T) {
	env := newPromoTestEnv()

	_, _, err := env.svc.Validate(context.Background(), "NOEXIST", uuid.New(), 200000)
	if err != domain.ErrPromoNotFound {
		t.Errorf("err = %v, want ErrPromoNotFound", err)
	}
}

func TestPromoService_Validate_WrongBathhouse(t *testing.T) {
	env := newPromoTestEnv()

	bhID := uuid.New()
	promo := &domain.PromoCode{
		ID:          uuid.New(),
		Code:        "BHSPECIFIC",
		Type:        domain.PromoTypePercentage,
		Value:       10,
		BathhouseID: &bhID,
		CreatorID:   uuid.New(),
		IsActive:    true,
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidUntil:  time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
	}
	if err := env.promoRepo.Create(context.Background(), promo); err != nil {
		t.Fatalf("setup: %v", err)
	}

	otherBhID := uuid.New()
	_, _, err := env.svc.Validate(context.Background(), "BHSPECIFIC", otherBhID, 200000)
	if err != domain.ErrPromoInvalid {
		t.Errorf("err = %v, want ErrPromoInvalid", err)
	}
}

// --- Apply ---

func TestPromoService_Apply_Success(t *testing.T) {
	env := newPromoTestEnv()
	adminID := uuid.New()

	promo := validPromoCode(nil)
	promo.Value = 15 // 15%
	_, err := env.svc.Create(context.Background(), adminID, domain.RoleAdmin, promo)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	userID := uuid.New()
	bookingID := uuid.New()
	discount, err := env.svc.Apply(context.Background(), userID, "SUMMER20", bookingID, uuid.Nil, 200000)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if discount != 30000 { // 15% of 200000
		t.Errorf("discount = %d, want 30000", discount)
	}
}

func TestPromoService_Apply_EmptyCode(t *testing.T) {
	env := newPromoTestEnv()

	_, err := env.svc.Apply(context.Background(), uuid.New(), "", uuid.New(), uuid.Nil, 200000)
	if err != domain.ErrPromoInvalid {
		t.Errorf("err = %v, want ErrPromoInvalid", err)
	}
}

// --- Deactivate ---

func TestPromoService_Deactivate_OwnerSuccess(t *testing.T) {
	env := newPromoTestEnv()
	ownerID := uuid.New()
	bh := createPromoBathhouse(env, ownerID)

	promo := validPromoCode(&bh.ID)
	created, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, promo)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	err = env.svc.Deactivate(context.Background(), ownerID, domain.RoleOwner, created.ID)
	if err != nil {
		t.Fatalf("deactivate: %v", err)
	}

	// Verify it's inactive via validate
	_, _, err = env.svc.Validate(context.Background(), "SUMMER20", bh.ID, 200000)
	if err != domain.ErrPromoInvalid {
		t.Errorf("expected ErrPromoInvalid after deactivation, got %v", err)
	}
}

func TestPromoService_Deactivate_Forbidden(t *testing.T) {
	env := newPromoTestEnv()
	ownerID := uuid.New()
	otherUserID := uuid.New()
	bh := createPromoBathhouse(env, ownerID)

	promo := validPromoCode(&bh.ID)
	created, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, promo)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	err = env.svc.Deactivate(context.Background(), otherUserID, domain.RoleOwner, created.ID)
	if err != domain.ErrForbidden {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestPromoService_Deactivate_GlobalNonAdminForbidden(t *testing.T) {
	env := newPromoTestEnv()
	adminID := uuid.New()

	promo := validPromoCode(nil)
	created, err := env.svc.Create(context.Background(), adminID, domain.RoleAdmin, promo)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	err = env.svc.Deactivate(context.Background(), uuid.New(), domain.RoleOwner, created.ID)
	if err != domain.ErrForbidden {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

// --- ListByBathhouse ---

func TestPromoService_ListByBathhouse_Success(t *testing.T) {
	env := newPromoTestEnv()
	ownerID := uuid.New()
	bh := createPromoBathhouse(env, ownerID)

	for i := 0; i < 3; i++ {
		p := validPromoCode(&bh.ID)
		p.Code = "CODE" + string(rune('A'+i))
		_, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, p)
		if err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	result, err := env.svc.ListByBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID, 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("total = %d, want 3", result.TotalCount)
	}
}

func TestPromoService_ListByBathhouse_Forbidden(t *testing.T) {
	env := newPromoTestEnv()
	ownerID := uuid.New()
	bh := createPromoBathhouse(env, ownerID)

	otherUserID := uuid.New()
	_, err := env.svc.ListByBathhouse(context.Background(), otherUserID, domain.RoleOwner, bh.ID, 1, 10)
	if err != domain.ErrForbidden {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

// --- Free Addon ---

func TestPromoService_Create_FreeAddon_Success(t *testing.T) {
	env := newPromoTestEnv()
	ownerID := uuid.New()
	bh := createPromoBathhouse(env, ownerID)

	addon := &domain.AddOn{
		ID:          uuid.New(),
		BathhouseID: bh.ID,
		Name:        "Веник дубовый",
		Price:       50000,
		Unit:        domain.AddOnUnitPerItem,
		IsActive:    true,
	}
	if err := env.addonRepo.Create(context.Background(), addon); err != nil {
		t.Fatalf("setup addon: %v", err)
	}

	promo := &domain.PromoCode{
		Code:          "FREEOAK",
		Type:          domain.PromoTypeFreeAddon,
		Value:         1,
		BathhouseID:   &bh.ID,
		TargetAddOnID: &addon.ID,
		MaxUses:       100,
		MinAmount:     0,
		ValidFrom:     time.Now().Add(-time.Hour),
		ValidUntil:    time.Now().Add(24 * time.Hour),
	}

	result, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, promo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != domain.PromoTypeFreeAddon {
		t.Errorf("type = %q, want free_addon", result.Type)
	}
	if result.TargetAddOnID == nil || *result.TargetAddOnID != addon.ID {
		t.Error("target_addon_id mismatch")
	}
}

func TestPromoService_Create_FreeAddon_MissingTargetAddon(t *testing.T) {
	env := newPromoTestEnv()
	ownerID := uuid.New()
	bh := createPromoBathhouse(env, ownerID)

	promo := &domain.PromoCode{
		Code:        "FREENOTHING",
		Type:        domain.PromoTypeFreeAddon,
		Value:       1,
		BathhouseID: &bh.ID,
		MaxUses:     100,
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidUntil:  time.Now().Add(24 * time.Hour),
	}

	_, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, promo)
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestPromoService_Create_FreeAddon_NoBathhouse(t *testing.T) {
	env := newPromoTestEnv()
	adminID := uuid.New()
	addonID := uuid.New()

	promo := &domain.PromoCode{
		Code:          "FREEGLOBAL",
		Type:          domain.PromoTypeFreeAddon,
		Value:         1,
		TargetAddOnID: &addonID,
		MaxUses:       100,
		ValidFrom:     time.Now().Add(-time.Hour),
		ValidUntil:    time.Now().Add(24 * time.Hour),
	}

	_, err := env.svc.Create(context.Background(), adminID, domain.RoleAdmin, promo)
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestPromoService_Create_FreeAddon_WrongBathhouse(t *testing.T) {
	env := newPromoTestEnv()
	ownerID := uuid.New()
	bh := createPromoBathhouse(env, ownerID)

	otherBhID := uuid.New()
	addon := &domain.AddOn{
		ID:          uuid.New(),
		BathhouseID: otherBhID,
		Name:        "Веник берёзовый",
		Price:       30000,
		Unit:        domain.AddOnUnitPerItem,
		IsActive:    true,
	}
	if err := env.addonRepo.Create(context.Background(), addon); err != nil {
		t.Fatalf("setup addon: %v", err)
	}

	promo := &domain.PromoCode{
		Code:          "FREEWRONG",
		Type:          domain.PromoTypeFreeAddon,
		Value:         1,
		BathhouseID:   &bh.ID,
		TargetAddOnID: &addon.ID,
		MaxUses:       100,
		ValidFrom:     time.Now().Add(-time.Hour),
		ValidUntil:    time.Now().Add(24 * time.Hour),
	}

	_, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, promo)
	if err != domain.ErrInvalidInput {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestPromoService_Validate_FreeAddon_Success(t *testing.T) {
	env := newPromoTestEnv()
	ownerID := uuid.New()
	bh := createPromoBathhouse(env, ownerID)

	addon := &domain.AddOn{
		ID:          uuid.New(),
		BathhouseID: bh.ID,
		Name:        "Веник дубовый",
		Price:       50000,
		Unit:        domain.AddOnUnitPerItem,
		IsActive:    true,
	}
	if err := env.addonRepo.Create(context.Background(), addon); err != nil {
		t.Fatalf("setup addon: %v", err)
	}

	promo := &domain.PromoCode{
		Code:          "FREEOAK",
		Type:          domain.PromoTypeFreeAddon,
		Value:         1,
		BathhouseID:   &bh.ID,
		TargetAddOnID: &addon.ID,
		MaxUses:       100,
		ValidFrom:     time.Now().Add(-time.Hour),
		ValidUntil:    time.Now().Add(24 * time.Hour),
	}
	if _, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, promo); err != nil {
		t.Fatalf("create: %v", err)
	}

	result, discount, err := env.svc.Validate(context.Background(), "FREEOAK", bh.ID, 300000)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if result.Type != domain.PromoTypeFreeAddon {
		t.Errorf("type = %q, want free_addon", result.Type)
	}
	if discount != 50000 {
		t.Errorf("discount = %d, want 50000 (addon price)", discount)
	}
}

func TestPromoService_Apply_FreeAddon_Success(t *testing.T) {
	env := newPromoTestEnv()
	ownerID := uuid.New()
	bh := createPromoBathhouse(env, ownerID)

	addon := &domain.AddOn{
		ID:          uuid.New(),
		BathhouseID: bh.ID,
		Name:        "Веник дубовый",
		Price:       50000,
		Unit:        domain.AddOnUnitPerItem,
		IsActive:    true,
	}
	if err := env.addonRepo.Create(context.Background(), addon); err != nil {
		t.Fatalf("setup addon: %v", err)
	}

	promo := &domain.PromoCode{
		Code:          "FREEOAK",
		Type:          domain.PromoTypeFreeAddon,
		Value:         1,
		BathhouseID:   &bh.ID,
		TargetAddOnID: &addon.ID,
		MaxUses:       100,
		ValidFrom:     time.Now().Add(-time.Hour),
		ValidUntil:    time.Now().Add(24 * time.Hour),
	}
	if _, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, promo); err != nil {
		t.Fatalf("create: %v", err)
	}

	userID := uuid.New()
	bookingID := uuid.New()
	discount, err := env.svc.Apply(context.Background(), userID, "FREEOAK", bookingID, bh.ID, 300000)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if discount != 50000 {
		t.Errorf("discount = %d, want 50000", discount)
	}
}
