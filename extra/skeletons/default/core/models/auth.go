package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type (
	User struct {
		ModelID `bson:",inline"`

		Name              string  `json:"name" bson:"name"` // Default size for string is 255, you could reset it with this tag
		Password          string  `json:"-" bson:"password"`
		IsRoot            bool    `json:"isRoot" bson:"is_root"`
		Emails            []Email `json:"emails" bson:"emails"`
		Phones            []Phone `json:"phones" bson:"phones"`
		RealNameValidated bool

		Profile Profile `json:"profile" bson:"profile"`
		Tokens  []Token `json:"tokens" bson:"tokens"`

		Timestamps `bson:",inline"`
	}
	SMSCode struct {
		ModelID    `bson:",inline"`
		User       User   `bson:"user"`
		Code       string `bson:"code"`
		ExpireTime bool   `bson:"expireTime"`
		UsedTime   bool   `bson:"usedTime"`
	}
	Token struct {
		ModelID `bson:",inline"`

		UserID   bson.ObjectId `json:"userId" bson:"user_id,omitempty"`
		Token    string        `json:"token" bson:"token"`
		Consumer string        `json:"consumer" bson:"consumer"`
		Device   string        `json:"device" bson:"device"`
		Limited  bool          `json:"limited" bson:"limited"`
		Expire   time.Time     `json:"expire" bson:"expire"`

		Timestamps `bson:",inline"`
	}
	Profile struct {
		ModelID `bson:",inline"`

		UserID    bson.ObjectId `json:"userId" bson:"user_id,omitempty"`
		Addresses []Address     `json:"addresses" bson:"addresses"`

		Timestamps `bson:",inline"`
	}
	Email struct {
		ModelID `bson:",inline"`

		UserID      bson.ObjectId `json:"userId" bson:"user_id,omitempty"`
		Email       string        `json:"email" bson:"email"`
		IsPrimary   bool          `json:"isPrimary" bson:"is_primary"`
		IsValidated bool          `json:"isValidated" bson:"is_validated"`
		Subscribed  bool          `json:"subscribed" bson:"subscribed"`

		Timestamps `bson:",inline"`
	}
	Phone struct {
		ModelID     `bson:",inline"`
		UserID      bson.ObjectId `json:"userId" bson:"user_id,omitempty"`
		Number      string        `json:"number" bson:"number"`
		IsPrimary   bool          `json:"isPrimary" bson:"is_primary"`
		isValidated bool
		Subscribed  bool `json:"subscribed" bson:"subscribed"`

		Timestamps `bson:",inline"`
	}
	Address struct {
		ModelID `bson:",inline"`

		UserID bson.ObjectId `json:"userId" bson:"user_id,omitempty"`

		Lng      int    `json:"age" bson:"age"`
		Lat      int    `json:"age" bson:"age"`
		Country  int    `json:"sexy" bson:"sexy"`
		Province int    `json:"birthday" bson:"birthday"`
		City     int    `json:"birthday" bson:"birthday"`
		Town     int    `json:"birthday" bson:"birthday"`
		Street   int    `json:"birthday" bson:"birthday"`
		Post     string `json:"post" bson:"post"`

		Address1 string `json:"address1" bson:"address1"` // Set field as not nullable and unique
		Address2 string `json:"address2" bson:"address2"`

		IsPrimary bool `json:"isPrimary" bson:"is_primary"`

		Timestamps `bson:",inline"`
	}
	Role struct {
		ModelID `bson:",inline"`

		Code string `json:"code" bson:"code"`
		Name string `json:"name" bson:"name"`
		Note string `json:"note" bson:"note"`

		Timestamps `bson:",inline"`
	}
	Operation struct {
		ModelID `bson:",inline"`

		Code  string `json:"code" bson:"code"`
		Zone  string `json:"zone" bson:"zone"`
		Title string `json:"title" bson:"title"`
		Note  string `json:"note" bson:"note"`

		Timestamps `bson:",inline"`
	}
	Permission struct {
		ModelID `bson:",inline"`

		RecordID    bson.ObjectId `json:"recordId" bson:"record_id,omitempty"`
		Allowed     bool          `json:"allowed" bson:"allowed"`
		UserID      bson.ObjectId `json:"userId" bson:"user_id"`
		RoleID      bson.ObjectId `json:"roleId" bson:"role_id"`
		TokenID     bson.ObjectId `json:"tokenId" bson:"token_id"`
		OperationID bson.ObjectId `json:"operationId" bson:"operation_id"`

		Timestamps `bson:",inline"`
	}
)
