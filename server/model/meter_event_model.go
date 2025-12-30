package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MeterEvent struct {
	Base             `json:",inline" bson:",inline"`
	EventName        string        `json:"event_name" bson:"event_name"`
	EventValue       int           `json:"event_value" bson:"event_value"`
	StripeCustomerID string        `json:"stripe_customer_id" bson:"stripe_customer_id"`
	OrganizationID   bson.ObjectID `json:"org_id" bson:"org_id"`
}
