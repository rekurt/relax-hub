package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditAction string

const (
	AuditActionCreate AuditAction = "create"
	AuditActionUpdate AuditAction = "update"
	AuditActionDelete AuditAction = "delete"
)

func (a AuditAction) IsValid() bool {
	switch a {
	case AuditActionCreate, AuditActionUpdate, AuditActionDelete:
		return true
	}
	return false
}

type AuditLog struct {
	ID            uuid.UUID
	EntityType    string
	EntityID      uuid.UUID
	UserID        uuid.UUID
	Action        AuditAction
	ChangedFields json.RawMessage
	CreatedAt     time.Time
}

type AuditLogFilter struct {
	EntityType *string
	EntityID   *uuid.UUID
	UserID     *uuid.UUID
	Action     *AuditAction
	FromDate   *time.Time
	ToDate     *time.Time
	Page       int
	PageSize   int
}
