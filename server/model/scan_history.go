package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ScanHistory struct {
	Base            `json:",inline" bson:",inline"`
	FormatID        bson.ObjectID `json:"format_id" bson:"format_id"`
	CategoryID      bson.ObjectID `json:"category_id" bson:"category_id"`
	ScanCode        string        `json:"scan_code" bson:"scan_code"`
	CategoryDataCol string        `json:"category_data_col" bson:"category_data_col"`
	CategoryDataID  bson.ObjectID `json:"category_data_id" bson:"category_data_id"`
	BatchID         bson.ObjectID `json:"batch_id" bson:"batch_id"`
	Thumbnails      []string      `json:"thumbnails" bson:"thumbnails"`
}

type CreateScanHistoryRequest struct {
	ScanCode        string        `json:"scan_code" bson:"scan_code"`
	CategoryID      bson.ObjectID `json:"category_id" bson:"category_id"`
	CategoryDataID  bson.ObjectID `json:"category_data_id" bson:"category_data_id"`
	BatchID         bson.ObjectID `json:"batch_id" bson:"batch_id"`
	CategoryDataCol string        `json:"category_data_col" bson:"category_data_col"`
	Thumbnails      []string      `json:"thumbnails" bson:"thumbnails"`
}
