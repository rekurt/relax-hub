package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

type addonTestEnv struct {
	svc       service.AddOnService
	addonRepo *mock.AddOnRepo
	bhRepo    *mock.BathhouseRepo
	repRepo   *mock.RepresentativeRepo
}

func newAddonTestEnv() *addonTestEnv {
	addonRepo := mock.NewAddOnRepo()
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	accessCheck := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	svc := service.NewAddOnService(addonRepo, accessCheck, log)
	return &addonTestEnv{
		svc:       svc,
		addonRepo: addonRepo,
		bhRepo:    bhRepo,
		repRepo:   repRepo,
	}
}

func createAddonBathhouse(env *addonTestEnv, ownerID uuid.UUID) *domain.Bathhouse {
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

func validAddOn(bathhouseID uuid.UUID) *domain.AddOn {
	return &domain.AddOn{
		BathhouseID: bathhouseID,
		Name:        "Веники берёзовые",
		Description: "Классические берёзовые веники",
		Price:       50000, // 500 rub
		Unit:        domain.AddOnUnitPerItem,
	}
}

// --- Create ---

func TestAddOnService_Create_Success(t *testing.T) {
	env := newAddonTestEnv()
	ownerID := uuid.New()
	bh := createAddonBathhouse(env, ownerID)

	addon := validAddOn(bh.ID)
	result, err := env.svc.CreateAddOn(context.Background(), ownerID, domain.RoleOwner, addon)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID == uuid.Nil {
		t.Error("expected addon ID to be set")
	}
	if !result.IsActive {
		t.Error("expected addon to be active")
	}
}

func TestAddOnService_Create_Forbidden(t *testing.T) {
	env := newAddonTestEnv()
	ownerID := uuid.New()
	otherID := uuid.New()
	bh := createAddonBathhouse(env, ownerID)

	addon := validAddOn(bh.ID)
	_, err := env.svc.CreateAddOn(context.Background(), otherID, domain.RoleOwner, addon)
	if err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got: %v", err)
	}
}

func TestAddOnService_Create_InvalidInput(t *testing.T) {
	env := newAddonTestEnv()
	ownerID := uuid.New()
	bh := createAddonBathhouse(env, ownerID)

	addon := &domain.AddOn{
		BathhouseID: bh.ID,
		Name:        "", // empty name
		Price:       50000,
		Unit:        domain.AddOnUnitPerItem,
	}
	_, err := env.svc.CreateAddOn(context.Background(), ownerID, domain.RoleOwner, addon)
	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestAddOnService_Create_LimitReached(t *testing.T) {
	env := newAddonTestEnv()
	ownerID := uuid.New()
	bh := createAddonBathhouse(env, ownerID)

	// Create 20 addons
	for i := 0; i < 20; i++ {
		addon := validAddOn(bh.ID)
		addon.Name = "addon " + uuid.New().String()[:8]
		if _, err := env.svc.CreateAddOn(context.Background(), ownerID, domain.RoleOwner, addon); err != nil {
			t.Fatalf("setup: unexpected error creating addon %d: %v", i, err)
		}
	}

	// 21st should fail
	addon := validAddOn(bh.ID)
	addon.Name = "one too many"
	_, err := env.svc.CreateAddOn(context.Background(), ownerID, domain.RoleOwner, addon)
	if err != domain.ErrAddOnLimitReached {
		t.Fatalf("expected ErrAddOnLimitReached, got: %v", err)
	}
}

// --- Update ---

func TestAddOnService_Update_Success(t *testing.T) {
	env := newAddonTestEnv()
	ownerID := uuid.New()
	bh := createAddonBathhouse(env, ownerID)

	addon := validAddOn(bh.ID)
	created, err := env.svc.CreateAddOn(context.Background(), ownerID, domain.RoleOwner, addon)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	update := &domain.AddOn{
		ID:       created.ID,
		Name:     "Updated name",
		Price:    60000,
		Unit:     domain.AddOnUnitPerHour,
		IsActive: true,
	}
	result, err := env.svc.UpdateAddOn(context.Background(), ownerID, domain.RoleOwner, update)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Updated name" {
		t.Errorf("expected updated name, got: %s", result.Name)
	}
	if result.Price != 60000 {
		t.Errorf("expected price 60000, got: %d", result.Price)
	}
}

func TestAddOnService_Update_NotFound(t *testing.T) {
	env := newAddonTestEnv()
	ownerID := uuid.New()
	createAddonBathhouse(env, ownerID)

	update := &domain.AddOn{
		ID:   uuid.New(),
		Name: "nonexistent",
		Unit: domain.AddOnUnitPerItem,
	}
	_, err := env.svc.UpdateAddOn(context.Background(), ownerID, domain.RoleOwner, update)
	if err != domain.ErrAddOnNotFound {
		t.Fatalf("expected ErrAddOnNotFound, got: %v", err)
	}
}

// --- Delete ---

func TestAddOnService_Delete_Success(t *testing.T) {
	env := newAddonTestEnv()
	ownerID := uuid.New()
	bh := createAddonBathhouse(env, ownerID)

	addon := validAddOn(bh.ID)
	created, err := env.svc.CreateAddOn(context.Background(), ownerID, domain.RoleOwner, addon)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	err = env.svc.DeleteAddOn(context.Background(), ownerID, domain.RoleOwner, created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = env.svc.GetAddOn(context.Background(), created.ID)
	if err != domain.ErrAddOnNotFound {
		t.Fatalf("expected ErrAddOnNotFound after delete, got: %v", err)
	}
}

func TestAddOnService_Delete_Forbidden(t *testing.T) {
	env := newAddonTestEnv()
	ownerID := uuid.New()
	otherID := uuid.New()
	bh := createAddonBathhouse(env, ownerID)

	addon := validAddOn(bh.ID)
	created, err := env.svc.CreateAddOn(context.Background(), ownerID, domain.RoleOwner, addon)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	err = env.svc.DeleteAddOn(context.Background(), otherID, domain.RoleOwner, created.ID)
	if err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got: %v", err)
	}
}

// --- List ---

func TestAddOnService_List_Success(t *testing.T) {
	env := newAddonTestEnv()
	ownerID := uuid.New()
	bh := createAddonBathhouse(env, ownerID)

	for i := 0; i < 3; i++ {
		addon := validAddOn(bh.ID)
		addon.Name = "addon " + uuid.New().String()[:8]
		if _, err := env.svc.CreateAddOn(context.Background(), ownerID, domain.RoleOwner, addon); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	addons, err := env.svc.ListAddOns(context.Background(), bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(addons) != 3 {
		t.Errorf("expected 3 addons, got: %d", len(addons))
	}
}

// --- CalculateAddOnTotal ---

func TestAddOnService_CalculateAddOnTotal_PerItem(t *testing.T) {
	env := newAddonTestEnv()
	ownerID := uuid.New()
	bh := createAddonBathhouse(env, ownerID)

	addon := validAddOn(bh.ID)
	addon.Price = 50000 // 500 rub
	addon.Unit = domain.AddOnUnitPerItem
	created, err := env.svc.CreateAddOn(context.Background(), ownerID, domain.RoleOwner, addon)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	total, items, err := env.svc.CalculateAddOnTotal(context.Background(), []service.AddOnSelection{
		{AddOnID: created.ID, Quantity: 3},
	}, bh.ID, 2, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 150000 { // 3 * 50000
		t.Errorf("expected total 150000, got: %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got: %d", len(items))
	}
}

func TestAddOnService_CalculateAddOnTotal_PerHour(t *testing.T) {
	env := newAddonTestEnv()
	ownerID := uuid.New()
	bh := createAddonBathhouse(env, ownerID)

	addon := &domain.AddOn{
		BathhouseID: bh.ID,
		Name:        "Музыка",
		Price:       20000, // 200 rub per hour
		Unit:        domain.AddOnUnitPerHour,
	}
	created, err := env.svc.CreateAddOn(context.Background(), ownerID, domain.RoleOwner, addon)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	total, _, err := env.svc.CalculateAddOnTotal(context.Background(), []service.AddOnSelection{
		{AddOnID: created.ID, Quantity: 1},
	}, bh.ID, 3, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 60000 { // 1 * 20000 * 3 hours
		t.Errorf("expected total 60000, got: %d", total)
	}
}

func TestAddOnService_CalculateAddOnTotal_PerPerson(t *testing.T) {
	env := newAddonTestEnv()
	ownerID := uuid.New()
	bh := createAddonBathhouse(env, ownerID)

	addon := &domain.AddOn{
		BathhouseID: bh.ID,
		Name:        "Тапочки",
		Price:       10000, // 100 rub per person
		Unit:        domain.AddOnUnitPerPerson,
	}
	created, err := env.svc.CreateAddOn(context.Background(), ownerID, domain.RoleOwner, addon)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	total, _, err := env.svc.CalculateAddOnTotal(context.Background(), []service.AddOnSelection{
		{AddOnID: created.ID, Quantity: 1},
	}, bh.ID, 2, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 40000 { // 1 * 10000 * 4 persons
		t.Errorf("expected total 40000, got: %d", total)
	}
}

func TestAddOnService_CalculateAddOnTotal_WrongBathhouse(t *testing.T) {
	env := newAddonTestEnv()
	ownerID := uuid.New()
	bh := createAddonBathhouse(env, ownerID)

	addon := validAddOn(bh.ID)
	created, err := env.svc.CreateAddOn(context.Background(), ownerID, domain.RoleOwner, addon)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	otherBhID := uuid.New()
	_, _, err = env.svc.CalculateAddOnTotal(context.Background(), []service.AddOnSelection{
		{AddOnID: created.ID, Quantity: 1},
	}, otherBhID, 2, 5)
	if err != domain.ErrAddOnNotFound {
		t.Fatalf("expected ErrAddOnNotFound, got: %v", err)
	}
}

func TestAddOnService_CalculateAddOnTotal_InvalidQuantity(t *testing.T) {
	env := newAddonTestEnv()

	_, _, err := env.svc.CalculateAddOnTotal(context.Background(), []service.AddOnSelection{
		{AddOnID: uuid.New(), Quantity: 0},
	}, uuid.New(), 2, 5)
	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestAddOnService_CalculateAddOnTotal_InactiveAddOn(t *testing.T) {
	env := newAddonTestEnv()
	ownerID := uuid.New()
	bh := createAddonBathhouse(env, ownerID)

	addon := validAddOn(bh.ID)
	created, err := env.svc.CreateAddOn(context.Background(), ownerID, domain.RoleOwner, addon)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Deactivate via update
	update := &domain.AddOn{
		ID:       created.ID,
		Name:     created.Name,
		Price:    created.Price,
		Unit:     created.Unit,
		IsActive: false,
	}
	_, err = env.svc.UpdateAddOn(context.Background(), ownerID, domain.RoleOwner, update)
	if err != nil {
		t.Fatalf("setup update: %v", err)
	}

	_, _, err = env.svc.CalculateAddOnTotal(context.Background(), []service.AddOnSelection{
		{AddOnID: created.ID, Quantity: 1},
	}, bh.ID, 2, 5)
	if err != domain.ErrAddOnNotFound {
		t.Fatalf("expected ErrAddOnNotFound for inactive addon, got: %v", err)
	}
}
