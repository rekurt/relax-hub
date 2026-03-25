package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type TicketCategory string

const (
	TicketCategoryQuestion     TicketCategory = "question"
	TicketCategoryProblem      TicketCategory = "problem"
	TicketCategoryComplaint    TicketCategory = "complaint"
	TicketCategoryRefundReq    TicketCategory = "refund_request"
	TicketCategoryAccountIssue TicketCategory = "account_issue"
)

func (c TicketCategory) IsValid() bool {
	switch c {
	case TicketCategoryQuestion, TicketCategoryProblem, TicketCategoryComplaint,
		TicketCategoryRefundReq, TicketCategoryAccountIssue:
		return true
	}
	return false
}

type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "open"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusEscalated  TicketStatus = "escalated"
	TicketStatusResolved   TicketStatus = "resolved"
	TicketStatusClosed     TicketStatus = "closed"
)

func (s TicketStatus) IsValid() bool {
	switch s {
	case TicketStatusOpen, TicketStatusInProgress, TicketStatusEscalated,
		TicketStatusResolved, TicketStatusClosed:
		return true
	}
	return false
}

type TicketPriority string

const (
	TicketPriorityLow      TicketPriority = "low"
	TicketPriorityMedium   TicketPriority = "medium"
	TicketPriorityHigh     TicketPriority = "high"
	TicketPriorityCritical TicketPriority = "critical"
)

func (p TicketPriority) IsValid() bool {
	switch p {
	case TicketPriorityLow, TicketPriorityMedium, TicketPriorityHigh, TicketPriorityCritical:
		return true
	}
	return false
}

type TicketLevel string

const (
	TicketLevelL1 TicketLevel = "L1"
	TicketLevelL2 TicketLevel = "L2"
	TicketLevelL3 TicketLevel = "L3"
)

func (l TicketLevel) IsValid() bool {
	switch l {
	case TicketLevelL1, TicketLevelL2, TicketLevelL3:
		return true
	}
	return false
}

type TicketSenderType string

const (
	TicketSenderUser  TicketSenderType = "user"
	TicketSenderAdmin TicketSenderType = "admin"
)

type Ticket struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	BookingID  *uuid.UUID
	Category   TicketCategory
	Status     TicketStatus
	Priority   TicketPriority
	Level      TicketLevel
	Subject    string
	AssignedTo *uuid.UUID
	CSATScore  *int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ResolvedAt *time.Time
}

func (t *Ticket) Validate() error {
	if t.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if !t.Category.IsValid() {
		return ErrInvalidInput
	}
	if t.Subject == "" || len(t.Subject) > 500 {
		return ErrInvalidInput
	}
	if t.Status != "" && !t.Status.IsValid() {
		return ErrInvalidInput
	}
	if t.Priority != "" && !t.Priority.IsValid() {
		return ErrInvalidInput
	}
	if t.Level != "" && !t.Level.IsValid() {
		return ErrInvalidInput
	}
	return nil
}

// AutoPriority returns priority based on category.
func AutoPriority(category TicketCategory) TicketPriority {
	switch category {
	case TicketCategoryRefundReq, TicketCategoryComplaint:
		return TicketPriorityHigh
	case TicketCategoryProblem, TicketCategoryAccountIssue:
		return TicketPriorityMedium
	default:
		return TicketPriorityLow
	}
}

type TicketMessage struct {
	ID          uuid.UUID
	TicketID    uuid.UUID
	SenderID    uuid.UUID
	SenderType  TicketSenderType
	Body        string
	Attachments []string
	CreatedAt   time.Time
}

func (m *TicketMessage) Validate() error {
	if m.TicketID == uuid.Nil || m.SenderID == uuid.Nil {
		return ErrInvalidInput
	}
	if m.Body == "" || len(m.Body) > 5000 {
		return ErrInvalidInput
	}
	if len(m.Attachments) > 10 {
		return ErrInvalidInput
	}
	for _, a := range m.Attachments {
		if len(a) > 2048 || (!strings.HasPrefix(a, "https://") && !strings.HasPrefix(a, "http://")) {
			return ErrInvalidInput
		}
	}
	return nil
}

type TicketFilter struct {
	UserID   *uuid.UUID
	Status   *TicketStatus
	Priority *TicketPriority
	Level    *TicketLevel
	Category *TicketCategory
	Page     int
	PageSize int
}

type TicketStatusCounts struct {
	Open       int64
	InProgress int64
	Escalated  int64
	Resolved   int64
	Closed     int64
}
