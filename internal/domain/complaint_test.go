package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestComplaintTargetType_IsValid(t *testing.T) {
	tests := []struct {
		targetType ComplaintTargetType
		valid      bool
	}{
		{ComplaintTargetReview, true},
		{ComplaintTargetBathhouse, true},
		{ComplaintTargetUser, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.targetType.IsValid(); got != tt.valid {
			t.Errorf("ComplaintTargetType(%q).IsValid() = %v, want %v", tt.targetType, got, tt.valid)
		}
	}
}

func TestComplaintReason_IsValid(t *testing.T) {
	tests := []struct {
		reason ComplaintReason
		valid  bool
	}{
		{ComplaintReasonSpam, true},
		{ComplaintReasonOffensive, true},
		{ComplaintReasonFake, true},
		{ComplaintReasonFraud, true},
		{ComplaintReasonOther, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.reason.IsValid(); got != tt.valid {
			t.Errorf("ComplaintReason(%q).IsValid() = %v, want %v", tt.reason, got, tt.valid)
		}
	}
}

func TestComplaintStatus_IsValid(t *testing.T) {
	tests := []struct {
		status ComplaintStatus
		valid  bool
	}{
		{ComplaintStatusPending, true},
		{ComplaintStatusResolved, true},
		{ComplaintStatusDismissed, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.valid {
			t.Errorf("ComplaintStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.valid)
		}
	}
}

func TestComplaint_Validate(t *testing.T) {
	valid := &Complaint{
		ReporterID: uuid.New(),
		TargetType: ComplaintTargetReview,
		TargetID:   uuid.New(),
		Reason:     ComplaintReasonSpam,
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid complaint returned error: %v", err)
	}

	validWithStatus := &Complaint{
		ReporterID: uuid.New(),
		TargetType: ComplaintTargetBathhouse,
		TargetID:   uuid.New(),
		Reason:     ComplaintReasonFraud,
		Status:     ComplaintStatusPending,
	}
	if err := validWithStatus.Validate(); err != nil {
		t.Errorf("valid complaint with status returned error: %v", err)
	}

	tests := []struct {
		name      string
		complaint Complaint
	}{
		{"nil reporter", Complaint{ReporterID: uuid.Nil, TargetType: ComplaintTargetReview, TargetID: uuid.New(), Reason: ComplaintReasonSpam}},
		{"invalid target type", Complaint{ReporterID: uuid.New(), TargetType: "bad", TargetID: uuid.New(), Reason: ComplaintReasonSpam}},
		{"nil target", Complaint{ReporterID: uuid.New(), TargetType: ComplaintTargetReview, TargetID: uuid.Nil, Reason: ComplaintReasonSpam}},
		{"invalid reason", Complaint{ReporterID: uuid.New(), TargetType: ComplaintTargetReview, TargetID: uuid.New(), Reason: "bad"}},
		{"invalid status", Complaint{ReporterID: uuid.New(), TargetType: ComplaintTargetReview, TargetID: uuid.New(), Reason: ComplaintReasonSpam, Status: "bad"}},
		{"description too long", Complaint{ReporterID: uuid.New(), TargetType: ComplaintTargetReview, TargetID: uuid.New(), Reason: ComplaintReasonSpam, Description: string(make([]byte, 2001))}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.complaint.Validate(); err == nil {
				t.Error("expected error for invalid complaint")
			}
		})
	}
}
