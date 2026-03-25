package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestDisputeReason_IsValid(t *testing.T) {
	tests := []struct {
		reason DisputeReason
		valid  bool
	}{
		{DisputeReasonServiceNotProvided, true},
		{DisputeReasonPoorQuality, true},
		{DisputeReasonDamage, true},
		{DisputeReasonSafetyIssue, true},
		{DisputeReasonBillingError, true},
		{DisputeReasonOther, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.reason.IsValid(); got != tt.valid {
			t.Errorf("DisputeReason(%q).IsValid() = %v, want %v", tt.reason, got, tt.valid)
		}
	}
}

func TestDisputeStatus_IsValid(t *testing.T) {
	tests := []struct {
		status DisputeStatus
		valid  bool
	}{
		{DisputeStatusOpen, true},
		{DisputeStatusEvidenceCollection, true},
		{DisputeStatusUnderReview, true},
		{DisputeStatusResolved, true},
		{DisputeStatusAppealed, true},
		{DisputeStatusClosed, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.valid {
			t.Errorf("DisputeStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.valid)
		}
	}
}

func TestDisputeResolution_IsValid(t *testing.T) {
	tests := []struct {
		resolution DisputeResolution
		valid      bool
	}{
		{DisputeResolutionFullRefund, true},
		{DisputeResolutionPartialRefund, true},
		{DisputeResolutionNoRefund, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.resolution.IsValid(); got != tt.valid {
			t.Errorf("DisputeResolution(%q).IsValid() = %v, want %v", tt.resolution, got, tt.valid)
		}
	}
}

func TestDisputeEvidenceType_IsValid(t *testing.T) {
	tests := []struct {
		evType DisputeEvidenceType
		valid  bool
	}{
		{DisputeEvidencePhoto, true},
		{DisputeEvidenceScreenshot, true},
		{DisputeEvidenceGPS, true},
		{DisputeEvidenceMessage, true},
		{DisputeEvidenceReceipt, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.evType.IsValid(); got != tt.valid {
			t.Errorf("DisputeEvidenceType(%q).IsValid() = %v, want %v", tt.evType, got, tt.valid)
		}
	}
}

func TestDispute_Validate(t *testing.T) {
	valid := Dispute{
		BookingID:    uuid.New(),
		InitiatorID:  uuid.New(),
		RespondentID: uuid.New(),
		Reason:       DisputeReasonPoorQuality,
		Description:  "The service was below standard",
	}

	if err := valid.Validate(); err != nil {
		t.Errorf("valid dispute: unexpected error: %v", err)
	}

	tests := []struct {
		name    string
		modify  func(*Dispute)
		wantErr bool
	}{
		{"missing booking_id", func(d *Dispute) { d.BookingID = uuid.Nil }, true},
		{"missing initiator_id", func(d *Dispute) { d.InitiatorID = uuid.Nil }, true},
		{"missing respondent_id", func(d *Dispute) { d.RespondentID = uuid.Nil }, true},
		{"invalid reason", func(d *Dispute) { d.Reason = "invalid" }, true},
		{"empty description", func(d *Dispute) { d.Description = "" }, true},
		{"description too long", func(d *Dispute) {
			d.Description = string(make([]byte, 5001))
		}, true},
		{"invalid status", func(d *Dispute) { d.Status = "bad" }, true},
		{"empty status ok", func(d *Dispute) { d.Status = "" }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := valid
			tt.modify(&d)
			err := d.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDisputeEvidence_Validate(t *testing.T) {
	valid := DisputeEvidence{
		DisputeID: uuid.New(),
		UserID:    uuid.New(),
		Type:      DisputeEvidencePhoto,
		URL:       "https://example.com/photo.jpg",
	}

	if err := valid.Validate(); err != nil {
		t.Errorf("valid evidence: unexpected error: %v", err)
	}

	tests := []struct {
		name    string
		modify  func(*DisputeEvidence)
		wantErr bool
	}{
		{"missing dispute_id", func(e *DisputeEvidence) { e.DisputeID = uuid.Nil }, true},
		{"missing user_id", func(e *DisputeEvidence) { e.UserID = uuid.Nil }, true},
		{"invalid type", func(e *DisputeEvidence) { e.Type = "invalid" }, true},
		{"empty url", func(e *DisputeEvidence) { e.URL = "" }, true},
		{"url too long", func(e *DisputeEvidence) {
			e.URL = "https://example.com/" + string(make([]byte, 2000))
		}, true},
		{"url without protocol", func(e *DisputeEvidence) { e.URL = "example.com/photo.jpg" }, true},
		{"description too long", func(e *DisputeEvidence) {
			e.Description = string(make([]byte, 2001))
		}, true},
		{"http url ok", func(e *DisputeEvidence) { e.URL = "http://example.com/photo.jpg" }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := valid
			tt.modify(&e)
			err := e.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
