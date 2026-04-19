package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

func newRFMTestService() (service.RFMService, *mock.GuestCardRepo, *mock.CustomSegmentRepo) {
	gcRepo := mock.NewGuestCardRepo().(*mock.GuestCardRepo)
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	csRepo := mock.NewCustomSegmentRepo(gcRepo).(*mock.CustomSegmentRepo)
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	return service.NewRFMService(gcRepo, csRepo, access, log), gcRepo, csRepo
}

func seedGuests(t *testing.T, gcRepo *mock.GuestCardRepo, ownerID, bathhouseID uuid.UUID, count int) []uuid.UUID {
	t.Helper()
	ctx := context.Background()
	clientIDs := make([]uuid.UUID, count)
	for i := 0; i < count; i++ {
		clientIDs[i] = uuid.New()
		card := &domain.GuestCard{
			ID:           uuid.New(),
			OwnerID:      ownerID,
			ClientID:     clientIDs[i],
			BathhouseID:  bathhouseID,
			FirstVisitAt: time.Now().Add(-time.Duration(count-i) * 24 * time.Hour),
			LastVisitAt:  time.Now().Add(-time.Duration(count-i) * 24 * time.Hour),
			VisitCount:   i + 1,
			TotalSpent:   int64((i + 1) * 100000),
			AvgCheck:     100000,
			Tags:         []string{},
		}
		if err := gcRepo.Upsert(ctx, card); err != nil {
			t.Fatalf("seed guest %d: %v", i, err)
		}
	}
	return clientIDs
}

func TestRFMService_GetRFMAnalysis(t *testing.T) {
	svc, gcRepo, _ := newRFMTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	bathhouseID := uuid.New()

	// Seed 10 guests with varying visit counts and spend
	seedGuests(t, gcRepo, ownerID, bathhouseID, 10)

	result, err := svc.GetRFMAnalysis(ctx, ownerID, domain.RoleOwner)
	if err != nil {
		t.Fatalf("GetRFMAnalysis: %v", err)
	}

	if len(result.Guests) != 10 {
		t.Errorf("expected 10 guests, got %d", len(result.Guests))
	}

	// Verify all guests have RFM scores 1-5
	for _, g := range result.Guests {
		if g.RFM.Recency < 1 || g.RFM.Recency > 5 {
			t.Errorf("recency score out of range: %d", g.RFM.Recency)
		}
		if g.RFM.Frequency < 1 || g.RFM.Frequency > 5 {
			t.Errorf("frequency score out of range: %d", g.RFM.Frequency)
		}
		if g.RFM.Monetary < 1 || g.RFM.Monetary > 5 {
			t.Errorf("monetary score out of range: %d", g.RFM.Monetary)
		}
	}

	// Verify matrix is populated
	if len(result.Matrix) == 0 {
		t.Error("expected non-empty RFM matrix")
	}

	// Verify matrix cell counts sum to total guests
	var matrixTotal int64
	for _, cell := range result.Matrix {
		matrixTotal += cell.Count
	}
	if matrixTotal != 10 {
		t.Errorf("matrix total %d != 10 guests", matrixTotal)
	}
}

func TestRFMService_GetRFMAnalysis_Empty(t *testing.T) {
	svc, _, _ := newRFMTestService()
	ctx := context.Background()

	result, err := svc.GetRFMAnalysis(ctx, uuid.New(), domain.RoleOwner)
	if err != nil {
		t.Fatalf("GetRFMAnalysis: %v", err)
	}

	if len(result.Guests) != 0 {
		t.Errorf("expected 0 guests, got %d", len(result.Guests))
	}
}

func TestRFMService_GetRFMAnalysis_ForbiddenForClient(t *testing.T) {
	svc, _, _ := newRFMTestService()
	ctx := context.Background()

	_, err := svc.GetRFMAnalysis(ctx, uuid.New(), domain.RoleClient)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestRFMService_CreateCustomSegment(t *testing.T) {
	svc, _, _ := newRFMTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	minVisits := 3
	segment := &domain.CustomSegment{
		Name: "Active guests",
		Conditions: domain.CustomSegmentCondition{
			VisitCountMin: &minVisits,
		},
	}

	err := svc.CreateCustomSegment(ctx, ownerID, domain.RoleOwner, segment)
	if err != nil {
		t.Fatalf("CreateCustomSegment: %v", err)
	}

	if segment.ID == uuid.Nil {
		t.Error("expected segment ID to be set")
	}
	if segment.OwnerID != ownerID {
		t.Errorf("expected owner_id=%s, got %s", ownerID, segment.OwnerID)
	}
}

func TestRFMService_CreateCustomSegment_Validation(t *testing.T) {
	svc, _, _ := newRFMTestService()
	ctx := context.Background()

	// Empty name
	segment := &domain.CustomSegment{
		Name: "",
	}
	err := svc.CreateCustomSegment(ctx, uuid.New(), domain.RoleOwner, segment)
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput for empty name, got %v", err)
	}

	// Invalid RFM score
	badScore := 6
	segment = &domain.CustomSegment{
		Name: "Bad RFM",
		Conditions: domain.CustomSegmentCondition{
			RFMRecencyMin: &badScore,
		},
	}
	err = svc.CreateCustomSegment(ctx, uuid.New(), domain.RoleOwner, segment)
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput for bad RFM score, got %v", err)
	}
}

func TestRFMService_CRUD_CustomSegment(t *testing.T) {
	svc, _, _ := newRFMTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	// Create
	minVisits := 2
	segment := &domain.CustomSegment{
		Name:       "Test segment",
		Conditions: domain.CustomSegmentCondition{VisitCountMin: &minVisits},
	}
	if err := svc.CreateCustomSegment(ctx, ownerID, domain.RoleOwner, segment); err != nil {
		t.Fatalf("Create: %v", err)
	}
	segmentID := segment.ID

	// List
	segments, err := svc.ListCustomSegments(ctx, ownerID, domain.RoleOwner)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}
	if segments[0].Name != "Test segment" {
		t.Errorf("expected name='Test segment', got '%s'", segments[0].Name)
	}

	// Update
	newMinVisits := 5
	updated := &domain.CustomSegment{
		ID:         segmentID,
		Name:       "Updated segment",
		Conditions: domain.CustomSegmentCondition{VisitCountMin: &newMinVisits},
	}
	if err := svc.UpdateCustomSegment(ctx, ownerID, domain.RoleOwner, updated); err != nil {
		t.Fatalf("Update: %v", err)
	}

	segments, _ = svc.ListCustomSegments(ctx, ownerID, domain.RoleOwner)
	if len(segments) != 1 || segments[0].Name != "Updated segment" {
		t.Error("segment not updated")
	}

	// Delete
	if err := svc.DeleteCustomSegment(ctx, ownerID, domain.RoleOwner, segmentID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	segments, _ = svc.ListCustomSegments(ctx, ownerID, domain.RoleOwner)
	if len(segments) != 0 {
		t.Errorf("expected 0 segments after delete, got %d", len(segments))
	}
}

func TestRFMService_CustomSegment_OwnerIsolation(t *testing.T) {
	svc, _, _ := newRFMTestService()
	ctx := context.Background()
	owner1 := uuid.New()
	owner2 := uuid.New()

	minVisits := 1
	seg := &domain.CustomSegment{
		Name:       "Owner1 segment",
		Conditions: domain.CustomSegmentCondition{VisitCountMin: &minVisits},
	}
	_ = svc.CreateCustomSegment(ctx, owner1, domain.RoleOwner, seg)

	// Owner2 cannot update or delete owner1's segment
	err := svc.UpdateCustomSegment(ctx, owner2, domain.RoleOwner, &domain.CustomSegment{
		ID:   seg.ID,
		Name: "Hacked",
	})
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden for cross-owner update, got %v", err)
	}

	err = svc.DeleteCustomSegment(ctx, owner2, domain.RoleOwner, seg.ID)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden for cross-owner delete, got %v", err)
	}
}

func TestRFMService_GetCustomSegmentGuests(t *testing.T) {
	svc, gcRepo, _ := newRFMTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	bathhouseID := uuid.New()

	// Seed guests: 5 with 1 visit, 5 with 5 visits
	for i := 0; i < 5; i++ {
		card := &domain.GuestCard{
			ID:           uuid.New(),
			OwnerID:      ownerID,
			ClientID:     uuid.New(),
			BathhouseID:  bathhouseID,
			FirstVisitAt: time.Now(),
			LastVisitAt:  time.Now(),
			VisitCount:   1,
			TotalSpent:   100000,
			AvgCheck:     100000,
			Tags:         []string{},
		}
		_ = gcRepo.Upsert(ctx, card)
	}
	for i := 0; i < 5; i++ {
		card := &domain.GuestCard{
			ID:           uuid.New(),
			OwnerID:      ownerID,
			ClientID:     uuid.New(),
			BathhouseID:  bathhouseID,
			FirstVisitAt: time.Now(),
			LastVisitAt:  time.Now(),
			VisitCount:   5,
			TotalSpent:   500000,
			AvgCheck:     100000,
			Tags:         []string{},
		}
		_ = gcRepo.Upsert(ctx, card)
	}

	// Create segment: visit_count >= 3
	minVisits := 3
	seg := &domain.CustomSegment{
		Name:       "Frequent visitors",
		Conditions: domain.CustomSegmentCondition{VisitCountMin: &minVisits},
	}
	_ = svc.CreateCustomSegment(ctx, ownerID, domain.RoleOwner, seg)

	// Get guests
	result, err := svc.GetCustomSegmentGuests(ctx, ownerID, domain.RoleOwner, seg.ID, 1, 20)
	if err != nil {
		t.Fatalf("GetCustomSegmentGuests: %v", err)
	}

	if result.TotalCount != 5 {
		t.Errorf("expected 5 frequent visitors, got %d", result.TotalCount)
	}
	for _, g := range result.Items {
		if g.VisitCount < 3 {
			t.Errorf("expected visit_count >= 3, got %d", g.VisitCount)
		}
	}
}

func TestRFMService_GetCustomSegmentGuests_WithTags(t *testing.T) {
	svc, gcRepo, _ := newRFMTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	bathhouseID := uuid.New()

	// Guest with tag "vip"
	card1 := &domain.GuestCard{
		ID: uuid.New(), OwnerID: ownerID, ClientID: uuid.New(), BathhouseID: bathhouseID,
		FirstVisitAt: time.Now(), LastVisitAt: time.Now(),
		VisitCount: 3, TotalSpent: 300000, AvgCheck: 100000,
		Tags: []string{"vip", "regular"},
	}
	_ = gcRepo.Upsert(ctx, card1)

	// Guest without tag
	card2 := &domain.GuestCard{
		ID: uuid.New(), OwnerID: ownerID, ClientID: uuid.New(), BathhouseID: bathhouseID,
		FirstVisitAt: time.Now(), LastVisitAt: time.Now(),
		VisitCount: 2, TotalSpent: 200000, AvgCheck: 100000,
		Tags: []string{},
	}
	_ = gcRepo.Upsert(ctx, card2)

	// Segment: must have tag "vip"
	seg := &domain.CustomSegment{
		Name:       "VIP tagged",
		Conditions: domain.CustomSegmentCondition{TagsInclude: []string{"vip"}},
	}
	_ = svc.CreateCustomSegment(ctx, ownerID, domain.RoleOwner, seg)

	result, err := svc.GetCustomSegmentGuests(ctx, ownerID, domain.RoleOwner, seg.ID, 1, 20)
	if err != nil {
		t.Fatalf("GetCustomSegmentGuests: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("expected 1 guest with tag 'vip', got %d", result.TotalCount)
	}
}

func TestCustomSegment_Validate(t *testing.T) {
	tests := []struct {
		name    string
		segment domain.CustomSegment
		wantErr bool
	}{
		{
			name:    "valid simple segment",
			segment: domain.CustomSegment{Name: "Test", OwnerID: uuid.New()},
			wantErr: false,
		},
		{
			name:    "empty name",
			segment: domain.CustomSegment{Name: "", OwnerID: uuid.New()},
			wantErr: true,
		},
		{
			name:    "nil owner",
			segment: domain.CustomSegment{Name: "Test"},
			wantErr: true,
		},
		{
			name: "inverted visit range",
			segment: domain.CustomSegment{
				Name: "Bad", OwnerID: uuid.New(),
				Conditions: domain.CustomSegmentCondition{
					VisitCountMin: intPtr(10), VisitCountMax: intPtr(5),
				},
			},
			wantErr: true,
		},
		{
			name: "RFM score out of range",
			segment: domain.CustomSegment{
				Name: "Bad RFM", OwnerID: uuid.New(),
				Conditions: domain.CustomSegmentCondition{RFMRecencyMin: intPtr(0)},
			},
			wantErr: true,
		},
		{
			name: "valid RFM range",
			segment: domain.CustomSegment{
				Name: "Good RFM", OwnerID: uuid.New(),
				Conditions: domain.CustomSegmentCondition{
					RFMRecencyMin:   intPtr(3),
					RFMRecencyMax:   intPtr(5),
					RFMFrequencyMin: intPtr(4),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.segment.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func intPtr(v int) *int {
	return &v
}
