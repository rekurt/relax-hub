package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

type offerTestEnv struct {
	svc       service.OfferService
	offerRepo *mock.OfferRepo
}

func newOfferTestEnv() *offerTestEnv {
	offerRepo := mock.NewOfferRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewOfferService(offerRepo, log)
	return &offerTestEnv{
		svc:       svc,
		offerRepo: offerRepo,
	}
}

func TestOfferService_Accept_Success(t *testing.T) {
	env := newOfferTestEnv()
	userID := uuid.New()

	acceptance, err := env.svc.Accept(context.Background(), userID, service.AcceptOfferInput{
		IPAddress: "127.0.0.1",
		UserAgent: "TestBrowser/1.0",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acceptance.UserID != userID {
		t.Errorf("user_id = %v, want %v", acceptance.UserID, userID)
	}
	if acceptance.OfferVersion == "" {
		t.Error("offer_version should not be empty")
	}
	if acceptance.IPAddress != "127.0.0.1" {
		t.Errorf("ip_address = %v, want 127.0.0.1", acceptance.IPAddress)
	}
	if acceptance.UserAgent != "TestBrowser/1.0" {
		t.Errorf("user_agent = %v, want TestBrowser/1.0", acceptance.UserAgent)
	}
}

func TestOfferService_Accept_AlreadyAccepted(t *testing.T) {
	env := newOfferTestEnv()
	userID := uuid.New()

	_, err := env.svc.Accept(context.Background(), userID, service.AcceptOfferInput{
		IPAddress: "127.0.0.1",
		UserAgent: "TestBrowser/1.0",
	})
	if err != nil {
		t.Fatalf("first accept failed: %v", err)
	}

	_, err = env.svc.Accept(context.Background(), userID, service.AcceptOfferInput{
		IPAddress: "127.0.0.1",
		UserAgent: "TestBrowser/1.0",
	})
	if err != domain.ErrOfferAlreadyAccepted {
		t.Fatalf("expected ErrOfferAlreadyAccepted, got: %v", err)
	}
}

func TestOfferService_GetStatus_NotAccepted(t *testing.T) {
	env := newOfferTestEnv()
	userID := uuid.New()

	status, err := env.svc.GetStatus(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Accepted {
		t.Error("expected not accepted for new user")
	}
	if status.CurrentVersion == "" {
		t.Error("current_version should not be empty")
	}
	if status.Acceptance != nil {
		t.Error("acceptance should be nil when not accepted")
	}
}

func TestOfferService_GetStatus_Accepted(t *testing.T) {
	env := newOfferTestEnv()
	userID := uuid.New()

	_, err := env.svc.Accept(context.Background(), userID, service.AcceptOfferInput{
		IPAddress: "10.0.0.1",
		UserAgent: "Mozilla/5.0",
	})
	if err != nil {
		t.Fatalf("accept failed: %v", err)
	}

	status, err := env.svc.GetStatus(context.Background(), userID)
	if err != nil {
		t.Fatalf("get status failed: %v", err)
	}
	if !status.Accepted {
		t.Error("expected accepted")
	}
	if status.Acceptance == nil {
		t.Fatal("acceptance should not be nil")
	}
	if status.Acceptance.UserID != userID {
		t.Errorf("acceptance.user_id = %v, want %v", status.Acceptance.UserID, userID)
	}
}

func TestOfferService_IsAccepted(t *testing.T) {
	env := newOfferTestEnv()
	userID := uuid.New()

	// Not accepted
	accepted, err := env.svc.IsAccepted(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accepted {
		t.Error("expected not accepted")
	}

	// Accept
	_, err = env.svc.Accept(context.Background(), userID, service.AcceptOfferInput{
		IPAddress: "127.0.0.1",
		UserAgent: "Test",
	})
	if err != nil {
		t.Fatalf("accept failed: %v", err)
	}

	// Now accepted
	accepted, err = env.svc.IsAccepted(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !accepted {
		t.Error("expected accepted after accepting")
	}
}

func TestOfferService_DifferentUsersCanAccept(t *testing.T) {
	env := newOfferTestEnv()
	user1 := uuid.New()
	user2 := uuid.New()

	_, err := env.svc.Accept(context.Background(), user1, service.AcceptOfferInput{
		IPAddress: "127.0.0.1",
		UserAgent: "Test",
	})
	if err != nil {
		t.Fatalf("user1 accept failed: %v", err)
	}

	_, err = env.svc.Accept(context.Background(), user2, service.AcceptOfferInput{
		IPAddress: "127.0.0.2",
		UserAgent: "Test",
	})
	if err != nil {
		t.Fatalf("user2 accept failed: %v", err)
	}
}
