package model

type UsageEvent struct {
	Base           `json:",inline" bson:",inline"`
	OrganizationID string `json:"organization_id" bson:"organization_id"`
	UserID         string `json:"user_id" bson:"user_id"`
	EventName      string `json:"event_name" bson:"event_name"`
	Quantity       int    `json:"quantity" bson:"quantity"`
	Timestamp      int64  `json:"timestamp" bson:"timestamp"`
}
