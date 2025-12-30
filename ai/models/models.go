package models

// import (
// 	"time"

// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )

// type Category struct {
// 	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
// 	Name        string             `json:"name" bson:"name"`
// 	FormatCount int                `json:"formatCount" bson:"formatCount"`
// 	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
// 	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
// }

// type Document struct {
// 	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
// 	CategoryID primitive.ObjectID `json:"categoryId" bson:"categoryId"`
// 	FileName   string             `json:"fileName" bson:"fileName"`
// 	FilePath   string             `json:"filePath" bson:"filePath"`
// 	FileSize   int64              `json:"fileSize" bson:"fileSize"`
// 	UploadedAt time.Time          `json:"uploadedAt" bson:"uploadedAt"`
// }

// type UploadResponse struct {
// 	Success  bool
// 	Message  string
// 	Category Category // Contains: ID, Name, FormatCount
// 	FileName string
// }

// type KeyValue struct {
// 	Key         string `json:"key"`
// 	Value       string `json:"value"`
// 	Description string `json:"description"`
// }

// type ExtractedFieldsWrapper struct {
// 	KeyValues []KeyValue `json:"key_values"`
// }

// type DocumentUploadResponse struct {
// 	Success         bool                   `json:"success"`
// 	Message         string                 `json:"message"`
// 	CategoryName    string                 `json:"categoryName"`
// 	FileName        string                 `json:"fileName"`
// 	FileSize        int64                  `json:"fileSize"`
// 	PageCount       int                    `json:"pageCount,omitempty"`
// 	ExtractedFields ExtractedFieldsWrapper `json:"extractedFields,omitempty"` // NOT json.RawMessage
// 	Formats         []string               `json:"formats,omitempty"`
// 	Confidence      string                 `json:"confidence,omitempty"`
// 	IsNewCategory   bool                   `json:"isNewCategory"`
// }
// type CategoryAnalysis struct {
// 	Category        string   `json:"category"`
// 	PageCount       int      `json:"pageCount,omitempty"`
// 	ExtractedFields []string `json:"extractedFields,omitempty"`
// 	Formats         []string `json:"formats,omitempty"`
// 	Confidence      float64  `json:"confidence,omitempty"`
// }

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Format represents a document format/schema version for a category
type Format struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	CategoryID      primitive.ObjectID     `bson:"categoryId" json:"categoryId"`
	CategoryName    string                 `bson:"categoryName" json:"categoryName"`
	FormatNumber    int                    `bson:"formatNumber" json:"formatNumber"`
	ExtractedFields map[string]interface{} `bson:"extractedfields" json:"extractedfields"`
	CreatedAt       time.Time              `bson:"createdAt" json:"createdAt"`
	SampleCount     int                    `bson:"sampleCount" json:"sampleCount"`
}

// Category represents a document category (updated to work with separate formats)
type Category struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	FormatCount int                `bson:"formatCount" json:"formatCount"`
	Schema      []Field            `bson:"schema,omitempty" json:"schema,omitempty"`
	Summary     string             `bson:"summary" json:"summary"`
	TotalDocs   int                `bson:"totalDocs" json:"totalDocs"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Field struct {
	Name     string    `json:"name" bson:"name"`
	Label    string    `json:"label" bson:"label"`
	Type     FieldType `json:"type" bson:"type"`
	Required bool      `json:"required" bson:"required"`
	Unique   bool      `json:"unique" bson:"unique"`
	Order    int       `bson:"order" json:"order"`
}

type FieldType string

const (
	FieldTypeText        FieldType = "text"
	FieldTypeTextarea    FieldType = "textarea"
	FieldTypeNumber      FieldType = "number"
	FieldTypeCurrency    FieldType = "currency"
	FieldTypeDate        FieldType = "date"
	FieldTypeDateTime    FieldType = "datetime"
	FieldTypeBoolean     FieldType = "boolean"
	FieldTypeSelect      FieldType = "select"
	FieldTypeMultiSelect FieldType = "multi_select"
	FieldTypeImage       FieldType = "image"
	FieldTypeURL         FieldType = "url"
	FieldTypeEmail       FieldType = "email"
	FieldTypePhone       FieldType = "phone"
	FieldTypeAddress     FieldType = "address"
	FieldTypeTable       FieldType = "table"
)

// Document represents an uploaded document
type Document struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CategoryID primitive.ObjectID `bson:"categoryId" json:"categoryId"`
	FormatID   primitive.ObjectID `bson:"formatId,omitempty" json:"formatId,omitempty"`
	FileName   string             `bson:"fileName" json:"fileName"`
	FilePath   string             `bson:"filePath" json:"filePath"`
	FileSize   int64              `bson:"fileSize" json:"fileSize"`
	PageCount  int                `bson:"pageCount,omitempty" json:"pageCount,omitempty"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type KeyValue struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

type ExtractedFieldsWrapper struct {
	KeyValues []KeyValue `json:"key_values"`
}

// DocumentUploadResponse is the response structure for document upload
type DocumentUploadResponse struct {
	Success         bool                   `json:"success"`
	Message         string                 `json:"message"`
	CategoryName    string                 `json:"categoryName"`
	CategoryID      primitive.ObjectID     `json:"categoryId,omitempty"`
	FileName        string                 `json:"fileName"`
	FileSize        int64                  `json:"fileSize"`
	PageCount       int                    `json:"pageCount"`
	ExtractedFields ExtractedFieldsWrapper `json:"extractedFields"`
	Formats         []Format               `json:"formats,omitempty"`
	Confidence      string                 `json:"confidence"`
	IsNewCategory   bool                   `json:"isNewCategory"`
}

// CategoryWithFormats represents a category with all its formats
type CategoryWithFormats struct {
	Category
	Formats []Format `json:"formats"`
}
