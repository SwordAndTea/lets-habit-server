package dal

import (
	"github.com/pkg/errors"
	"github.com/swordandtea/fhwh/biz/response"
	"gorm.io/gorm"
	"time"
)

// UserEmailActivate the user who need to activate by verifying their email
type UserEmailActivate struct {
	ID        uint64    `json:"id"`
	UID       UID       `json:"uid"`
	Email     string    `json:"email"`
	Password  *Password `json:"-"`
	SendAt    time.Time `json:"send_at"`
	Activated bool      `json:"activated"`
}

type userEmailActivateDBHD struct {
}

var UserEmailActivateDBHD = &userEmailActivateDBHD{}

func (hd *userEmailActivateDBHD) Add(db *gorm.DB, uea *UserEmailActivate) response.SError {
	err := db.Create(&uea).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add inactivated user fail")
	}
	return nil
}

func (hd *userEmailActivateDBHD) GetByUID(db *gorm.DB, uid UID) (*UserEmailActivate, response.SError) {
	var uea *UserEmailActivate
	err := db.Where("uid=?", uid).First(&uea).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "get inactivated user fail")
	}
	return uea, nil
}

func (hd *userEmailActivateDBHD) GetByEmail(db *gorm.DB, email string) (*UserEmailActivate, response.SError) {
	var uea *UserEmailActivate
	err := db.Where("email=?", email).First(&uea).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "get inactivated user fail")
	}
	return uea, nil
}

func (hd *userEmailActivateDBHD) UpdateSendTime(db *gorm.DB, id uint64, sendAt time.Time) response.SError {
	err := db.Model(&UserEmailActivate{}).Where("id=?", id).Update("send_at", sendAt).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "update send time fail")
	}
	return nil
}

func (hd *userEmailActivateDBHD) SetActivated(db *gorm.DB, id uint64) response.SError {
	err := db.Model(&UserEmailActivate{}).Where("id=?", id).Update("activated", true).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "update send time fail")
	}
	return nil
}

func (hd *userEmailActivateDBHD) DeleteByUID(db *gorm.DB, uid UID) response.SError {
	err := db.Where("uid=?", uid).Delete(&UserEmailActivate{}).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "delete inactivated user fail")
	}
	return nil
}
