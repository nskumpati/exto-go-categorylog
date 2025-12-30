package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Base struct {
	ID        bson.ObjectID `json:"id,omitzero" bson:"_id,omitzero"`
	CreatedAt time.Time     `json:"created_at,omitzero" bson:"created_at"`
	UpdatedAt time.Time     `json:"updated_at,omitzero" bson:"updated_at,omitempty"`
	DeletedAt time.Time     `json:"deleted_at,omitzero" bson:"deleted_at,omitempty"`
	CreatedBy bson.ObjectID `json:"created_by" bson:"created_by"`
	UpdatedBy bson.ObjectID `json:"updated_by" bson:"updated_by"`
	DeletedBy bson.ObjectID `json:"deleted_by,omitzero" bson:"deleted_by,omitempty"`
}

func (b *Base) IsZero() bool {
	return b == nil || b.ID.IsZero()
}

func (b *Base) IsDeleted() bool {
	return !b.DeletedAt.IsZero()
}

type BaseUpdate struct {
	UpdatedAt time.Time     `json:"updated_at,omitzero" bson:"updated_at,omitempty"`
	DeletedAt time.Time     `json:"deleted_at,omitzero" bson:"deleted_at,omitempty"`
	UpdatedBy bson.ObjectID `json:"updated_by" bson:"updated_by"`
	DeletedBy bson.ObjectID `json:"deleted_by,omitzero" bson:"deleted_by,omitempty"`
}
