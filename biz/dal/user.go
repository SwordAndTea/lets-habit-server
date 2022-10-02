package dal

import (
	"github.com/pkg/errors"
	"github.com/swordandtea/fhwh/biz/response"
	"gorm.io/gorm"
)

type UserRegisterType string

const (
	UserRegisterTypeEmail  UserRegisterType = "email"
	UserRegisterTypeWechat UserRegisterType = "wechat"
)

// User the user registered
type User struct {
	ID               uint64           `json:"id"`
	UID              UID              `json:"uid"`
	Email            string           `json:"email"`
	Portrait         string           `json:"portrait"`
	UserRegisterType UserRegisterType `json:"user_register_type"`
}

// userDBHD the handler to operate the user table
type userDBHD struct{}

// UserDBHD the default userDBHD
var UserDBHD = &userDBHD{}

// Add insert a user record
func (hd *userDBHD) Add(db *gorm.DB, user *User) response.SError {
	err := db.Create(user).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add user fail")
	}
	return nil
}

// GetByUID get a User by uid
func (hd *userDBHD) GetByUID(db *gorm.DB, uid UID) (*User, response.SError) {
	var u *User
	err := db.Where("uid=?", uid).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "get user by fail")
	}
	return u, nil
}

// ListByUIDs list Users by a list of uid
func (hd *userDBHD) ListByUIDs(db *gorm.DB, uids []UID) ([]*User, response.SError) {
	var us []*User
	err := db.Where("uid in (?)", uids).Find(&us).Error
	if err != nil {
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "list user fail")
	}
	return us, nil
}
