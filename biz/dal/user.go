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
	UID              string           `json:"uid"`
	Email            string           `json:"email"`
	Portrait         string           `json:"portrait"`
	UserRegisterType UserRegisterType `json:"user_register_type"`
}

type userDBHD struct{}

var UserDBHD = &userDBHD{}

func (hd *userDBHD) Add(db *gorm.DB, user *User) response.SError {
	err := db.Create(user).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add user fail")
	}
	return nil
}

func (hd *userDBHD) GetByUID(db *gorm.DB, uid string) (*User, response.SError) {
	var u *User
	err := db.Where("uid=?", uid).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "get user by uid %s fail", uid)
	}
	return u, nil
}
