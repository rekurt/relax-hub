package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleClient         UserRole = "client"
	RoleOwner          UserRole = "owner"
	RoleRepresentative UserRole = "representative"
	RoleAdmin          UserRole = "admin"
)

func (r UserRole) IsValid() bool {
	switch r {
	case RoleClient, RoleOwner, RoleRepresentative, RoleAdmin:
		return true
	}
	return false
}

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string `json:"-"`
	Name         string
	Phone        string
	Role         UserRole
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) Validate() error {
	if u.Email == "" {
		return ErrInvalidInput
	}
	if u.Name == "" {
		return ErrInvalidInput
	}
	if !u.Role.IsValid() {
		return ErrInvalidInput
	}
	return nil
}
