package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Identity struct {
	Base         `json:",inline" bson:",inline"`
	Email        string        `json:"email" bson:"email"`
	FirstName    string        `json:"first_name" bson:"first_name"`
	LastName     string        `json:"last_name" bson:"last_name"`
	IsActive     bool          `json:"is_active" bson:"is_active"`
	CurrentOrgID bson.ObjectID `json:"current_org_id" bson:"current_org_id"`
}

type CreateIdentity struct {
	Email        string
	FirstName    string
	LastName     string
	CurrentOrgID bson.ObjectID
}

type UpdateIdentity struct {
	FirstName    string
	LastName     string
	IsActive     bool
	CurrentOrgID bson.ObjectID
}
