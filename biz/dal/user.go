package dal

import (
	"github.com/pkg/errors"
	"github.com/swordandtea/fhwh/biz/response"
	"github.com/swordandtea/fhwh/biz/service"
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
	Name             string           `json:"name"`
	Email            string           `json:"email"`
	Password         Password         `json:"-"`
	Portrait         string           `json:"-"` //portrait Tos Key
	PortraitURL      string           `json:"portrait" gorm:"-"`
	UserRegisterType UserRegisterType `json:"user_register_type"`
}

func postProcessUserField(users []*User) {
	for _, u := range users {
		u.PortraitURL = service.GetObjectStorageExecutor().ObjectKeyToURL(u.Portrait)
	}
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
	postProcessUserField([]*User{user})
	return nil
}

// GetByUID get a User by uid
func (hd *userDBHD) GetByUID(db *gorm.DB, uid UID) (*User, response.SError) {
	var user *User
	err := db.Where("uid=?", uid).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "get user by fail")
	}
	postProcessUserField([]*User{user})
	return user, nil
}

// ListByUIDs list Users by a list of uid
func (hd *userDBHD) ListByUIDs(db *gorm.DB, uids []UID) ([]*User, response.SError) {
	var users []*User
	err := db.Where("uid in (?)", uids).Find(&users).Error
	if err != nil {
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "list user fail")
	}
	postProcessUserField(users)
	return users, nil
}
