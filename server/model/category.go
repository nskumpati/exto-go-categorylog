package model

type Category struct {
	Base         `json:",inline" bson:",inline"`
	Name         string  `json:"name" bson:"name"`
	PrimaryField string  `json:"primary_field" bson:"primary_field"`
	Slug         string  `json:"slug" bson:"slug"`
	Version      string  `json:"version" bson:"version"`
	Fields       []Field `json:"fields,omitempty" bson:"fields,omitempty"`
	Sys          bool    `json:"is_active" bson:"is_active"`
}

type Field struct {
	Name     string        `json:"name" bson:"name"`
	Label    string        `json:"label" bson:"label"`
	Type     FieldType     `json:"type" bson:"type"`
	Options  []FieldOption `json:"options,omitempty" bson:"options,omitempty"`
	Required bool          `json:"required" bson:"required"`
	Unique   bool          `json:"unique" bson:"unique"`

	Children []Field `json:"children,omitempty" bson:"children,omitempty"`
}

type FieldOption struct {
	Name             string   `json:"name" bson:"name"`
	AlternativeNames []string `json:"alternative_name,omitempty" bson:"alternative_name,omitempty"`
}

type FieldType string

const (
	FieldTypeText        FieldType = "text"
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
