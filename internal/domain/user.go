package domain

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
