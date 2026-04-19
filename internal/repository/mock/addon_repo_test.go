package mock_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository/mock"
)

func TestAddOnRepo_CreateAndGet(t *testing.T) {
	repo := mock.NewAddOnRepo()
	ctx := context.Background()
	bathhouseID := uuid.New()

	addon := &domain.AddOn{
		BathhouseID: bathhouseID,
		Name:        "Birch broom",
		Description: "Classic birch broom",
		Price:       50000,
		Unit:        domain.AddOnUnitPerItem,
		IsActive:    true,
		SortOrder:   1,
	}

	if err := repo.Create(ctx, addon); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addon.ID == uuid.Nil {
		t.Error("expected non-nil ID after create")
	}

	got, err := repo.GetByID(ctx, addon.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Birch broom" {
		t.Errorf("expected name 'Birch broom', got %q", got.Name)
	}
	if got.Price != 50000 {
		t.Errorf("expected price 50000, got %d", got.Price)
	}
	if got.Unit != domain.AddOnUnitPerItem {
		t.Errorf("expected unit per_item, got %s", got.Unit)
	}
}

func TestAddOnRepo_GetByID_NotFound(t *testing.T) {
	repo := mock.NewAddOnRepo()

	_, err := repo.GetByID(context.Background(), uuid.New())
	if err != domain.ErrAddOnNotFound {
		t.Errorf("expected ErrAddOnNotFound, got %v", err)
	}
}

func TestAddOnRepo_Update(t *testing.T) {
	repo := mock.NewAddOnRepo()
	ctx := context.Background()

	addon := &domain.AddOn{
		BathhouseID: uuid.New(),
		Name:        "Towel",
		Price:       30000,
		Unit:        domain.AddOnUnitPerPerson,
		IsActive:    true,
	}
	_ = repo.Create(ctx, addon)

	addon.Name = "Large towel"
	addon.Price = 40000
	if err := repo.Update(ctx, addon); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := repo.GetByID(ctx, addon.ID)
	if got.Name != "Large towel" {
		t.Errorf("expected name 'Large towel', got %q", got.Name)
	}
	if got.Price != 40000 {
		t.Errorf("expected price 40000, got %d", got.Price)
	}
}

func TestAddOnRepo_Update_NotFound(t *testing.T) {
	repo := mock.NewAddOnRepo()

	err := repo.Update(context.Background(), &domain.AddOn{ID: uuid.New()})
	if err != domain.ErrAddOnNotFound {
		t.Errorf("expected ErrAddOnNotFound, got %v", err)
	}
}

func TestAddOnRepo_Delete(t *testing.T) {
	repo := mock.NewAddOnRepo()
	ctx := context.Background()

	addon := &domain.AddOn{
		BathhouseID: uuid.New(),
		Name:        "Robe",
		Price:       60000,
		Unit:        domain.AddOnUnitPerPerson,
		IsActive:    true,
	}
	_ = repo.Create(ctx, addon)

	if err := repo.Delete(ctx, addon.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.GetByID(ctx, addon.ID)
	if err != nil {
		t.Fatalf("unexpected error after soft delete: %v", err)
	}
	if got.IsActive {
		t.Errorf("expected IsActive=false after soft delete, got true")
	}
}

func TestAddOnRepo_Delete_NotFound(t *testing.T) {
	repo := mock.NewAddOnRepo()

	err := repo.Delete(context.Background(), uuid.New())
	if err != domain.ErrAddOnNotFound {
		t.Errorf("expected ErrAddOnNotFound, got %v", err)
	}
}

func TestAddOnRepo_ListByBathhouse(t *testing.T) {
	repo := mock.NewAddOnRepo()
	ctx := context.Background()
	bhID := uuid.New()
	otherBhID := uuid.New()

	_ = repo.Create(ctx, &domain.AddOn{BathhouseID: bhID, Name: "A1", Unit: domain.AddOnUnitPerItem, IsActive: true})
	_ = repo.Create(ctx, &domain.AddOn{BathhouseID: bhID, Name: "A2", Unit: domain.AddOnUnitPerHour, IsActive: true})
	_ = repo.Create(ctx, &domain.AddOn{BathhouseID: otherBhID, Name: "A3", Unit: domain.AddOnUnitPerItem, IsActive: true})

	list, err := repo.ListByBathhouse(ctx, bhID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 addons, got %d", len(list))
	}
}

func TestAddOnRepo_CountByBathhouse(t *testing.T) {
	repo := mock.NewAddOnRepo()
	ctx := context.Background()
	bhID := uuid.New()

	_ = repo.Create(ctx, &domain.AddOn{BathhouseID: bhID, Name: "A1", Unit: domain.AddOnUnitPerItem})
	_ = repo.Create(ctx, &domain.AddOn{BathhouseID: bhID, Name: "A2", Unit: domain.AddOnUnitPerItem})
	_ = repo.Create(ctx, &domain.AddOn{BathhouseID: uuid.New(), Name: "A3", Unit: domain.AddOnUnitPerItem})

	count, err := repo.CountByBathhouse(ctx, bhID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestAddOnRepo_BookingAddOns(t *testing.T) {
	repo := mock.NewAddOnRepo()
	ctx := context.Background()
	bookingID := uuid.New()
	addonID := uuid.New()

	ba := &domain.BookingAddOn{
		BookingID:  bookingID,
		AddOnID:    addonID,
		Name:       "Birch broom",
		Quantity:   2,
		UnitPrice:  50000,
		TotalPrice: 100000,
	}

	if err := repo.CreateBookingAddOn(ctx, ba); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ba.ID == uuid.Nil {
		t.Error("expected non-nil ID after create")
	}

	list, err := repo.ListByBooking(ctx, bookingID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 booking addon, got %d", len(list))
	}
	if list[0].Name != "Birch broom" {
		t.Errorf("expected name 'Birch broom', got %q", list[0].Name)
	}
	if list[0].TotalPrice != 100000 {
		t.Errorf("expected total 100000, got %d", list[0].TotalPrice)
	}

	// Different booking should return empty
	other, err := repo.ListByBooking(ctx, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(other) != 0 {
		t.Errorf("expected 0 booking addons for other booking, got %d", len(other))
	}
}

func TestAddOn_Validate(t *testing.T) {
	tests := []struct {
		name    string
		addon   domain.AddOn
		wantErr bool
	}{
		{
			name: "valid",
			addon: domain.AddOn{
				BathhouseID: uuid.New(),
				Name:        "Towel",
				Price:       30000,
				Unit:        domain.AddOnUnitPerPerson,
			},
			wantErr: false,
		},
		{
			name: "zero price is valid",
			addon: domain.AddOn{
				BathhouseID: uuid.New(),
				Name:        "Free water",
				Price:       0,
				Unit:        domain.AddOnUnitPerItem,
			},
			wantErr: false,
		},
		{
			name: "missing bathhouse ID",
			addon: domain.AddOn{
				Name:  "Towel",
				Price: 30000,
				Unit:  domain.AddOnUnitPerPerson,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			addon: domain.AddOn{
				BathhouseID: uuid.New(),
				Price:       30000,
				Unit:        domain.AddOnUnitPerPerson,
			},
			wantErr: true,
		},
		{
			name: "negative price",
			addon: domain.AddOn{
				BathhouseID: uuid.New(),
				Name:        "Towel",
				Price:       -100,
				Unit:        domain.AddOnUnitPerPerson,
			},
			wantErr: true,
		},
		{
			name: "invalid unit",
			addon: domain.AddOn{
				BathhouseID: uuid.New(),
				Name:        "Towel",
				Price:       30000,
				Unit:        "per_kg",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.addon.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddOnUnit_IsValid(t *testing.T) {
	valid := []domain.AddOnUnit{domain.AddOnUnitPerItem, domain.AddOnUnitPerHour, domain.AddOnUnitPerPerson}
	for _, u := range valid {
		if !u.IsValid() {
			t.Errorf("expected %s to be valid", u)
		}
	}

	invalid := domain.AddOnUnit("per_kg")
	if invalid.IsValid() {
		t.Error("expected per_kg to be invalid")
	}
}
