package dal

import (
	"github.com/pkg/errors"
	"github.com/swordandtea/lets-habit-server/biz/response"
	"github.com/swordandtea/lets-habit-server/biz/service"
	"gorm.io/gorm"
	"time"
)

// UserRegisterType indentify how user is registered
type UserRegisterType string

const (
	UserRegisterTypeEmail  UserRegisterType = "email"  // user registered directly with email
	UserRegisterTypeWechat UserRegisterType = "wechat" // user registered with wechat oauth
)

// User the user registered
type User struct {
	ID               uint64           `json:"id"`
	UID              UID              `json:"uid"`
	Name             *string          `json:"name"`
	Email            *string          `json:"email"`
	EmailActive      bool             `json:"email_active"`
	Password         *Password        `json:"-"`
	Portrait         *string          `json:"-"` //portrait object storage Key
	PortraitURL      string           `json:"portrait" gorm:"-"`
	UserRegisterType UserRegisterType `json:"user_register_type"`
}

// postProcessUserField process some field after User data is fetched from db,
// basically is some field related with time and url
func postProcessUserField(users []*User) {
	for _, u := range users {
		if u.Portrait != nil {
			u.PortraitURL, _ = service.GetObjectStorageExecutor().ObjectKeyToURL(*u.Portrait, time.Hour)
		}
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

func (hd *userDBHD) GetByEmail(db *gorm.DB, email string) (*User, response.SError) {
	var user *User
	err := db.Where("email=?", email).First(&user).Error
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

type UserUpdatableFields struct {
	Name        string
	Email       string
	EmailActive *bool
	EmailBind   *bool
	Password    *Password
	Portrait    string
}

// UpdateUser update user field
func (hd *userDBHD) UpdateUser(db *gorm.DB, uid UID, updateFields *UserUpdatableFields) response.SError {
	updates := map[string]interface{}{}
	if updateFields.Name != "" {
		updates["name"] = updateFields.Name
	}
	if updateFields.Email != "" {
		updates["email"] = updateFields.Email
	}
	if updateFields.EmailActive != nil {
		updates["email_active"] = *updateFields.EmailActive
	}
	if updateFields.EmailBind != nil {
		updates["email_bind"] = *updateFields.EmailBind
	}
	if updateFields.Password != nil {
		updates["passport"] = updateFields.Password
	}
	if updateFields.Portrait != "" {
		updates["portrait"] = updateFields.Portrait
	}

	if len(updates) == 0 {
		return nil
	}
	err := db.Model(&User{}).Where("uid=?", uid).Updates(updates).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "update user fail")
	}
	return nil
}

func (hd *userDBHD) SearchUserByNameOrUID(db *gorm.DB, text string, pagination *Pagination) ([]*User, response.SError) {
	var users []*User
	offset := (pagination.Page - 1) * pagination.PageSize
	queryText := "%" + text + "%"
	err := db.Where("name Like ? or uid like ?", queryText, queryText).
		Offset(int(offset)).Limit(int(pagination.PageSize)).
		Find(&users).Error
	if err != nil {
		return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "search user fail")
	}

	postProcessUserField(users)
	return users, nil
}

func (hd *userDBHD) DeleteUserByID(db *gorm.DB, uid UID) response.SError {
	err := db.Where("uid=?", uid).Delete(&User{}).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "delete user fail")
	}
	return nil
}
