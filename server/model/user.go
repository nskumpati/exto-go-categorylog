package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type UserRole string

// enum for user roles
const (
	RoleSuperAdmin        UserRole = "super_admin"
	RoleBillingAdmin      UserRole = "billing_admin"
	RoleOrganizationAdmin UserRole = "organization_admin"
	RoleMember            UserRole = "member"
	RoleGuest             UserRole = "guest"
)

type User struct {
	Base           `json:",inline" bson:",inline"`
	IdentityID     bson.ObjectID `json:"identity_id" bson:"identity_id"`
	Email          string        `json:"email" bson:"email"`
	FirstName      string        `json:"first_name" bson:"first_name"`
	LastName       string        `json:"last_name" bson:"last_name"`
	OrganizationID bson.ObjectID `json:"org_id" bson:"org_id"`
	Role           UserRole      `json:"role" bson:"role"`
	IsActive       bool          `json:"is_active" bson:"is_active"`
}

type CreateUser struct {
	IdentityID     bson.ObjectID
	Email          string
	FirstName      string
	LastName       string
	Role           UserRole
	OrganizationID bson.ObjectID
}

type UpdateUser struct {
	FirstName string   `json:"first_name" bson:"first_name"`
	LastName  string   `json:"last_name" bson:"last_name"`
	Role      UserRole `json:"role" bson:"role"`
	IsActive  bool     `json:"is_active" bson:"is_active"`
}
