package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type (
	Parent struct {
		ModelID `bson:",inline"`

		User User `json:"name" bson:"name"`

		ChildrenIds []bson.ObjectId `json:"-" bson:"children_ids"`

		Children      []Child `json:"children,omitempty" bson:"-"`
		ChildrenCount int     `json:"childrenCount,omitempty" bson:"-"`

		Timestamps `bson:",inline"`
	}
	Child struct {
		ModelID `bson:",inline"`

		AuthCode string `json:"authCode" bson:"auth_code"`

		Age      int `json:"age" bson:"age"`
		Sexy     int `json:"sexy" bson:"sexy"`
		Birthday int `json:"birthday" bson:"birthday"`

		Timestamps `bson:",inline"`
	}
	Hospital struct {
		ModelID `bson:",inline"`

		User    User `json:"name" bson:"name"`
		Address Address

		Phone   string
		Fax     string
		WebSite string

		Timestamps `bson:",inline"`
	}
	Doctor struct {
		ModelID   `bson:",inline"`
		User      User `json:"name" bson:"name"`
		Hospitals []Hospital

		Name string
		Age  int
		Sexy bool

		Timestamps `bson:",inline"`
	}
	Evaluation struct {
		ModelID    `bson:",inline"`
		Score      int
		Remark     int
		Kind       int
		Timestamps `bson:",inline"`
	}
	Disease struct {
		ModelID `bson:",inline"`

		PreferHospital Hospital
		PreferDoctor   Doctor
		Parent         Parent
		Child          Child
		Reward         float64

		HealthNotes []HealthNote

		Description string

		Timestamps `bson:",inline"`
	}
	Treatment struct {
		ModelID `bson:",inline"`

		Hospital Hospital
		Doctor   Doctor
		Parent   Parent
		Child    Child
		Disease  Disease

		Adopted bool

		Description string

		Timestamps `bson:",inline"`
	}
	Reward struct {
		ModelID `bson:",inline"`

		Parent  Parent
		Child   Child
		Disease Disease

		Kind int

		CurrencyAmount int
		CurrencyUnit   int

		ViaPayPlatform string

		Description string

		Timestamps `bson:",inline"`
	}
	HealthNote struct {
		ModelID `bson:",inline"`

		CheckTime *time.Time `json:"trashedAt,omitempty" bson:"trashed_at,omitempty"`
		Height    string     `json:"height" bson:"height"`
		Weight    string     `json:"weight" bson:"weight"`

		Timestamps `bson:",inline"`
	}
)
