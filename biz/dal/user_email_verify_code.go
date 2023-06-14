package dal

import (
	"github.com/pkg/errors"
	"github.com/swordandtea/lets-habit-server/biz/response"
	"gorm.io/gorm"
	"time"
)

type UserEmailVerifyCode struct {
	Email      string `json:"-"`
	VerifyCode string `json:"-"`
	SendAt     time.Time
}

type userEmailVerifyCodeDBHD struct{}

var UserEmailVerifyCodeDBHD = &userEmailVerifyCodeDBHD{}

func (hd *userEmailVerifyCodeDBHD) Add(db *gorm.DB, verifyCode *UserEmailVerifyCode) response.SError {
	err := db.Create(verifyCode).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add user email verify code info fail")
	}
	return nil
}

func (hd *userEmailVerifyCodeDBHD) GetByEmail(db *gorm.DB, email string) (*UserEmailVerifyCode, response.SError) {
	var code *UserEmailVerifyCode
	err := db.Where("email=?", email).First(&code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "get user email verify code fail")
	}
	return code, nil
}

func (hd *userEmailVerifyCodeDBHD) DeleteByEmail(db *gorm.DB, email string) response.SError {
	err := db.Where("email=?", email).Delete(&UserEmailVerifyCode{}).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "delete user email verify code fail")
	}
	return nil
}
