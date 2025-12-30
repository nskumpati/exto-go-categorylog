package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CategoryData struct {
	Base           `json:",inline" bson:",inline"`
	FormatID       bson.ObjectID  `json:"format_id" bson:"format_id"`
	CategoryID     bson.ObjectID  `json:"category_id" bson:"category_id"`
	MetaData       map[string]any `json:"metadata,omitempty" bson:"metadata,omitempty"`
	RawData        map[string]any `json:"rawdata,omitempty" bson:"rawdata,omitempty"`
	DocumentPaths  []string       `json:"document_paths" bson:"document_paths"`
	OrganizationID bson.ObjectID  `json:"org_id" bson:"org_id"`
}

type CreateCategoryDataRequest struct {
	CategoryID    bson.ObjectID  `json:"category_id" bson:"category_id"`
	MetaData      map[string]any `json:"metadata,omitempty" bson:"metadata,omitempty"`
	RawData       map[string]any `json:"rawdata,omitempty" bson:"rawdata,omitempty"`
	DocumentPaths []string       `json:"document_paths" bson:"document_paths"`
}

type UpdateCategoryDataRequest struct {
	MetaData map[string]any `json:"metadata,omitempty" bson:"metadata,omitempty"`
}
