package model

type BatchStatus string

const (
	BatchStatusOpen   BatchStatus = "Open"
	BatchStatusClosed BatchStatus = "Closed"
)

type Batch struct {
	Base   `json:",inline" bson:",inline"`
	Name   string      `json:"name" bson:"name"`
	Status BatchStatus `json:"status" bson:"status"` // e.g., "Open", "Closed"
}

type CreateBatchRequest struct {
	Name   string      `json:"name" bson:"name"`
	Status BatchStatus `json:"status" bson:"status"`
}
