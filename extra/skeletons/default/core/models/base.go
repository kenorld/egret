package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type ModelID struct {
	ID bson.ObjectId `json:"id" bson:"_id,omitempty"`
}

type Timestamps struct {
	CreatedAt *time.Time `json:"createdAt,omitempty" bson:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty" bson:"updated_at,omitempty"`
	TrashedAt *time.Time `json:"trashedAt,omitempty" bson:"trashed_at,omitempty"`
}
