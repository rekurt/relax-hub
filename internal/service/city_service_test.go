package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func TestCityService_Create_Success(t *testing.T) {
	cityRepo := mock.NewCityRepo()
	svc := service.NewCityService(cityRepo)

	city, err := svc.Create(context.Background(), service.CreateCityInput{
		Name: "Moscow", Slug: "moscow", Region: "central", Latitude: 55.75, Longitude: 37.62,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if city.Name != "Moscow" {
		t.Errorf("name = %q, want %q", city.Name, "Moscow")
	}
	if city.Region != "central" {
		t.Errorf("region = %q, want %q", city.Region, "central")
	}
	if city.ID == 0 {
		t.Error("city ID should be assigned")
	}
}

func TestCityService_Create_InvalidInput(t *testing.T) {
	cityRepo := mock.NewCityRepo()
	svc := service.NewCityService(cityRepo)

	_, err := svc.Create(context.Background(), service.CreateCityInput{
		Name: "", Slug: "empty",
	})

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail for empty name, got: %v", err)
	}
}

func TestCityService_GetAll(t *testing.T) {
	cityRepo := mock.NewCityRepo()
	svc := service.NewCityService(cityRepo)

	_, _ = svc.Create(context.Background(), service.CreateCityInput{
		Name: "Moscow", Slug: "moscow", Region: "central", Latitude: 55.75, Longitude: 37.62,
	})
	_, _ = svc.Create(context.Background(), service.CreateCityInput{
		Name: "Saint Petersburg", Slug: "spb", Region: "northwest", Latitude: 59.93, Longitude: 30.32,
	})

	cities, err := svc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cities) != 2 {
		t.Errorf("len = %d, want 2", len(cities))
	}
}

func TestCityService_Update(t *testing.T) {
	cityRepo := mock.NewCityRepo()
	svc := service.NewCityService(cityRepo)

	city, _ := svc.Create(context.Background(), service.CreateCityInput{
		Name: "Moscow", Slug: "moscow", Latitude: 55.75, Longitude: 37.62,
	})

	newName := "Москва"
	updated, err := svc.Update(context.Background(), city.ID, service.UpdateCityInput{
		Name: &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Москва" {
		t.Errorf("name = %q, want %q", updated.Name, "Москва")
	}
}

func TestCityService_Update_Region(t *testing.T) {
	cityRepo := mock.NewCityRepo()
	svc := service.NewCityService(cityRepo)

	city, _ := svc.Create(context.Background(), service.CreateCityInput{
		Name: "Moscow", Slug: "moscow", Region: "central", Latitude: 55.75, Longitude: 37.62,
	})

	newRegion := "moscow-region"
	updated, err := svc.Update(context.Background(), city.ID, service.UpdateCityInput{
		Region: &newRegion,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Region != "moscow-region" {
		t.Errorf("region = %q, want %q", updated.Region, "moscow-region")
	}
	if updated.Name != "Moscow" {
		t.Errorf("name should be preserved, got %q", updated.Name)
	}
}

func TestCityService_Create_EmptyRegion(t *testing.T) {
	cityRepo := mock.NewCityRepo()
	svc := service.NewCityService(cityRepo)

	city, err := svc.Create(context.Background(), service.CreateCityInput{
		Name: "Moscow", Slug: "moscow", Latitude: 55.75, Longitude: 37.62,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if city.Region != "" {
		t.Errorf("region = %q, want empty string", city.Region)
	}
}

func TestCityService_Delete(t *testing.T) {
	cityRepo := mock.NewCityRepo()
	svc := service.NewCityService(cityRepo)

	city, _ := svc.Create(context.Background(), service.CreateCityInput{
		Name: "Moscow", Slug: "moscow", Latitude: 55.75, Longitude: 37.62,
	})

	err := svc.Delete(context.Background(), city.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.GetBySlug(context.Background(), "moscow")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("city should be deleted, got: %v", err)
	}
}

func TestCityService_GetBySlug(t *testing.T) {
	cityRepo := mock.NewCityRepo()
	svc := service.NewCityService(cityRepo)

	_, _ = svc.Create(context.Background(), service.CreateCityInput{
		Name: "Moscow", Slug: "moscow", Region: "central", Latitude: 55.75, Longitude: 37.62,
	})

	city, err := svc.GetBySlug(context.Background(), "moscow")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if city.Name != "Moscow" {
		t.Errorf("name = %q, want %q", city.Name, "Moscow")
	}
	if city.Region != "central" {
		t.Errorf("region = %q, want %q", city.Region, "central")
	}
}
