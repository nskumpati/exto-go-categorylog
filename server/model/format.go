package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Format struct {
	Base       `json:",inline" bson:",inline"`
	Name       string        `json:"name" bson:"name"`
	CategoryID bson.ObjectID `json:"category_id" bson:"category_id"`
	Sys        bool          `json:"is_active" bson:"is_active"`

	ExtractionFields []ExtractionField `json:"extraction_fields,omitempty" bson:"extraction_fields,omitempty"`
	Documents        []FormatDoc       `json:"documents" bson:"documents"`
	ExtractedSample  any               `json:"extracted_sample,omitempty" bson:"extracted_sample,omitempty"`
}

type FormatDoc struct {
	DocumentPath string   `json:"document_path" bson:"document_path"`
	Vectors      []string `json:"vectors,omitempty" bson:"vectors,omitempty"`
}

type ExtractionField struct {
	Name              string           `json:"name" bson:"name"`
	CategoryFieldName string           `json:"category_field_name" bson:"category_field_name"`
	Prompt            ExtractionPrompt `json:"extraction_prompt" bson:"extraction_prompt"`
}

type ExtractionPrompt struct {
	Text         string   `json:"text" bson:"text"`
	SampleValues []string `json:"sample_values,omitempty" bson:"sample_values,omitempty"`
	Images       []string `json:"images,omitempty" bson:"images,omitempty"`
}
