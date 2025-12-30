package model

import "go.mongodb.org/mongo-driver/v2/bson"

type MonthlyUsageEvent struct {
	Base           `json:",inline" bson:",inline"`
	OrganizationID bson.ObjectID `json:"organization_id" bson:"organization_id"`
	EventName      string        `json:"event_name" bson:"event_name"`
	Year           int           `json:"year" bson:"year"`
	Month          int           `json:"month" bson:"month"`
	Quantity       int           `json:"quantity" bson:"quantity"`
}
