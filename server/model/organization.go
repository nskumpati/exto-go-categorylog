package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Organization struct {
	Base             `json:",inline" bson:",inline"`
	Name             string        `json:"name" bson:"name"`
	Slug             string        `json:"slug" bson:"slug"`
	OwnerID          bson.ObjectID `json:"owner_id" bson:"owner_id"`
	IsActive         bool          `json:"is_active" bson:"is_active"`
	ScanCounter      int           `json:"scan_counter" bson:"scan_counter"`
	LastActiveAt     time.Time     `json:"last_active_at" bson:"last_active_at"`
	StripeCustomerId string        `json:"stripe_customer_id" bson:"stripe_customer_id"`
	Billing          Billing       `json:"billing" bson:"billing"`
}
type Billing struct {
	FullName      string `json:"full_name" bson:"full_name"`
	Email         string `json:"email" bson:"email"`
	Phone         string `json:"phone" bson:"phone"`
	StreetAddress string `json:"street_address" bson:"street_address"`
	Country       string `json:"country" bson:"country"`
	State         string `json:"state" bson:"state"`
	City          string `json:"city" bson:"city"`
	ZipCode       string `json:"zip_code" bson:"zip_code"`
}

type CreateOrganization struct {
	ID   bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Name string        `json:"name" bson:"name"`
	Slug string        `json:"slug" bson:"slug"`
}

type UpdateOrganization struct {
	OwnerID          bson.ObjectID `json:"owner_id" bson:"owner_id,omitempty"`
	StripeCustomerId string        `json:"stripe_customer_id" bson:"stripe_customer_id,omitempty"`
	Billing          Billing       `json:"billing" bson:"billing,omitempty"`
	IsActive         bool          `json:"is_active" bson:"is_active,omitempty"`
}

type MongoUpdateOrganization struct {
	BaseUpdate       `json:",inline" bson:",inline"`
	OwnerID          bson.ObjectID `json:"owner_id" bson:"owner_id,omitempty"`
	StripeCustomerId string        `json:"stripe_customer_id" bson:"stripe_customer_id,omitempty"`
	Billing          Billing       `json:"billing" bson:"billing,omitempty"`
	IsActive         bool          `json:"is_active" bson:"is_active,omitempty"`
}
