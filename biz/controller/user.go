package controller

import (
	"bytes"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/xid"
	"github.com/swordandtea/lets-habit-server/biz/config"
	"github.com/swordandtea/lets-habit-server/biz/dal"
	"github.com/swordandtea/lets-habit-server/biz/response"
	"github.com/swordandtea/lets-habit-server/biz/service"
	"github.com/swordandtea/lets-habit-server/util"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"sync"
	"text/template"
	"time"
)

type UserCtrl struct{}

var onceEmailActivate = &sync.Once{}
var onceEmailBind = &sync.Once{}
var onceEmailResetPassword = &sync.Once{}
var emailActivateTmpl *template.Template
var emailBindTmpl *template.Template
var emailResetPasswordTmpl *template.Template

// emailActivateTmplStr the email template used for user registering from email
const emailActivateTmplStr = `From: {{.From}}
To: {{.To}}
Subject: [lets-habits] 邮箱激活 (mail activate)
Content-Type: text/plain; charset=utf-8

欢迎加入lets-habits，点击下方链接以激活账户:
{{.ActiveLink}}

Welcome to join lets-habits, click the link below to activate your account:
{{.ActiveLink}}
`

// emailBindTmplStr the email template used for user to bind an email address
const emailBindTmplStr = `From: {{.From}}
To: {{.To}}
Subject: [lets-habits] 邮箱绑定 (mail activate)
Content-Type: text/plain; charset=utf-8

你正在为你的账户绑定邮箱，如果这是你的操作点击下方链接以完成绑定:
{{.BindLink}}
如果你没有操作绑定此邮箱，请忽略此邮件

You are currently binding email for your account, if it's your operation, click the link below to bind:
{{.BindLink}}
otherwise, please ignore this message
`

const emailResetPasswordTmplStr = `From: {{.From}}
To: {{.To}}
Subject: [lets-habits] 密码更新 (password update)
Content-Type: text/plain; charset=utf-8

你正在为你的账户更新密码，如果这是你的操作请填写下方验证码:
{{.VerifyCode}}
如果不是你的操作，请忽略此邮件

You are currently updating password for your account, if it's your operation, fill in the verify code below:
{{.VerifyCode}}
otherwise, please ignore this message
`

// emailActivateAllowedInterval the max time interval that allow a user to resend account activate email
const emailActivateAllowedInterval = time.Minute

// emailCodeExpireTime the token expired time for email activate and bind
const emailCodeExpireTime = time.Minute * 30

// userTokenExpireTime user auth token expire time: 7 day
const userTokenExpireTime = time.Hour * 24 * 7

type emailActivateTmplFiller struct {
	From       string
	To         string
	ActiveLink string
}

type emailBindTmplFiller struct {
	From     string
	To       string
	BindLink string
}

type emailResetPasswordTmplFiller struct {
	From       string
	To         string
	VerifyCode string
}

// GetEmailActivateTemplate lazy load email activate template
func GetEmailActivateTemplate() *template.Template {
	onceEmailActivate.Do(func() {
		emailActivateTmpl, _ = template.New("mail-activate-tmpl").Parse(emailActivateTmplStr)
	})
	return emailActivateTmpl
}

// GetEmailBindTemplate lazy load email bind template
func GetEmailBindTemplate() *template.Template {
	onceEmailBind.Do(func() {
		emailActivateTmpl, _ = template.New("mail-bind-tmpl").Parse(emailBindTmplStr)
	})
	return emailBindTmpl
}

func GetEmailResetPasswordTemplate() *template.Template {
	onceEmailResetPassword.Do(func() {
		emailResetPasswordTmpl, _ = template.New("mail-reset-password-tmpl").Parse(emailResetPasswordTmplStr)
	})
	return emailResetPasswordTmpl
}

// sendActivateEmail send email activate email to targe email address
func (c *UserCtrl) sendActivateEmail(toMail string, uid dal.UID) response.SError {
	mailExecutor := service.GetMailExecutor()
	// prepare email message
	// generate activate token
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(emailCodeExpireTime)),
		ID:        string(uid),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(config.GlobalConfig.JWT.Cypher))
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "sign activate code fail")
	}

	data := &bytes.Buffer{}
	err = GetEmailActivateTemplate().Execute(data, &emailActivateTmplFiller{
		From: mailExecutor.Sender(),
		To:   toMail,
		ActiveLink: fmt.Sprintf("%s?%s=%s", config.GlobalConfig.EmailService.ActivateURI,
			config.GlobalConfig.EmailService.ActivateParam, tokenStr),
	})
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "fill email activate template fail")
	}

	// send email
	err = mailExecutor.SendMail([]string{toMail}, data.Bytes())
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "send email fail")
	}
	return nil
}

// sendEmailBindEmail send email bind email to target email address
func (c *UserCtrl) sendEmailBindEmail(toMail string, uid dal.UID) response.SError {
	mailExecutor := service.GetMailExecutor()
	// prepare email message
	claims := &jwt.RegisteredClaims{
		Subject:   toMail,
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(emailCodeExpireTime)),
		ID:        string(uid),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(config.GlobalConfig.JWT.Cypher))
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "sign bind code fail")
	}

	data := &bytes.Buffer{}
	err = GetEmailBindTemplate().Execute(data, &emailBindTmplFiller{
		From: mailExecutor.Sender(),
		To:   toMail,
		BindLink: fmt.Sprintf("%s?%s=%s", config.GlobalConfig.EmailService.BindURI,
			config.GlobalConfig.EmailService.BindParam, tokenStr),
	})
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "fill email bind template fail")
	}

	// send email
	err = mailExecutor.SendMail([]string{toMail}, data.Bytes())
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "send email fail")
	}
	return nil
}

func (c *UserCtrl) sendEmailResetPasswordEmail(toMail string, verifyCode string) response.SError {
	mailExecutor := service.GetMailExecutor()

	data := &bytes.Buffer{}
	err := GetEmailResetPasswordTemplate().Execute(data, &emailResetPasswordTmplFiller{
		From:       mailExecutor.Sender(),
		To:         toMail,
		VerifyCode: verifyCode,
	})
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "fill email reset password template fail")
	}

	err = mailExecutor.SendMail([]string{toMail}, data.Bytes())
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "send reset password email fail")
	}
	return nil
}

// EmailRegister do email register, will send an email activate email to user
func (c *UserCtrl) EmailRegister(email string, password *dal.Password) (*dal.User, response.SError) {
	db := service.GetDBExecutor()
	user, sErr := dal.UserDBHD.GetByEmail(db, email)
	if sErr != nil {
		return nil, sErr
	}
	if user != nil {
		return nil, response.ErrorCode_UserNoPermission.New("email already registered")
	}

	uid := dal.UID(xid.New().String())

	user = &dal.User{
		UID:              uid,
		Email:            &email,
		EmailActive:      false,
		Password:         password,
		UserRegisterType: dal.UserRegisterTypeEmail,
	}

	sErr = WithDBTx(db, func(tx *gorm.DB) response.SError {
		sErr = dal.UserDBHD.Add(tx, user)
		if sErr != nil {
			return sErr
		}

		sErr = c.sendActivateEmail(email, uid)
		if sErr != nil {
			return sErr
		}
		return nil
	})

	if sErr != nil {
		return nil, sErr
	}
	return user, nil
}

// CheckEmailActivated check whether user has activated account email
//func (c *UserCtrl) CheckEmailActivated(uid dal.UID) (bool, response.SError) {
//	db := service.GetDBExecutor()
//	user, sErr := dal.UserDBHD.GetByUID(db, uid)
//	if sErr != nil {
//		return false, sErr
//	}
//	if user == nil {
//		return false, response.ErrorCode_InvalidParam.New("invalid uid")
//	}
//	return user.EmailActive, nil
//}

func (c *UserCtrl) ResendActivateEmail(uid dal.UID) response.SError {
	db := service.GetDBExecutor()
	user, sErr := dal.UserDBHD.GetByUID(db, uid)
	if sErr != nil {
		return sErr
	}

	if user == nil { // means the email has not been registered
		return response.ErrorCode_InvalidParam.New("user not registered")
	}

	if user.EmailActive { // email already activated
		return response.ErrorCode_UserNoPermission.New("email already activated")
	}

	if user.UserRegisterType != dal.UserRegisterTypeEmail {
		return response.ErrorCode_UserNoPermission.New("user not registered by email")
	}

	sErr = c.sendActivateEmail(*user.Email, user.UID)
	if sErr != nil {
		return sErr
	}
	return nil
}

// EmailActivate confirm email activate, create a new user if activate success
func (c *UserCtrl) EmailActivate(activateCode string) (*dal.User, response.SError) {
	// verify activate code
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(activateCode, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GlobalConfig.JWT.Cypher), nil
	})
	if err != nil {
		return nil, response.ErrorCode_UserNoPermission.Wrap(err, "invalid activate code")
	}
	db := service.GetDBExecutor()
	uid := dal.UID(claims.ID)
	user, sErr := dal.UserDBHD.GetByUID(db, uid)
	if sErr != nil {
		return nil, sErr
	}
	if user == nil {
		return nil, response.ErrorCode_InvalidParam.Wrap(err, "invalid activate code, no user found")
	}
	if user.EmailActive {
		return nil, response.ErrorCode_UserNoPermission.Wrap(err, "email already activated")
	}

	// do activate
	sErr = WithDBTx(db, func(tx *gorm.DB) response.SError {
		// mark user email activated
		return dal.UserDBHD.UpdateUser(tx, uid, &dal.UserUpdatableFields{EmailActive: util.LiteralValuePtr(true)})
	})
	if sErr != nil {
		return nil, sErr
	}
	user.EmailActive = true
	return user, nil
}

// StartEmailBinding begin email bind process, will send an email bind email to user
func (c *UserCtrl) StartEmailBinding(uid dal.UID, email string) response.SError {
	db := service.GetDBExecutor()
	user, sErr := dal.UserDBHD.GetByUID(db, uid)
	if sErr != nil {
		return sErr
	}

	if user == nil {
		return response.ErrorCode_InvalidParam.New("no user found")
	}

	if user.Email != nil && *user.Email == email {
		return response.ErrorCode_InvalidParam.New("same email with the email already bond")
	}

	sErr = c.sendEmailBindEmail(email, uid)
	if sErr != nil {
		return sErr
	}

	return nil
}

// ConfirmBindEmail confirm bind user email, update user email info
func (c *UserCtrl) ConfirmBindEmail(bindCode string) response.SError {
	// verify bind code
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(bindCode, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GlobalConfig.JWT.Cypher), nil
	})
	if err != nil {
		return response.ErrorCode_UserNoPermission.Wrap(err, "invalid activate code")
	}

	db := service.GetDBExecutor()
	sErr := dal.UserDBHD.UpdateUser(db, dal.UID(claims.ID), &dal.UserUpdatableFields{
		Email: claims.Subject,
	})
	if sErr != nil {
		return sErr
	}
	return nil
}

func (c *UserCtrl) LoginByEmail(email string, password *dal.Password) (*dal.User, response.SError) {
	db := service.GetDBExecutor()
	user, sErr := dal.UserDBHD.GetByEmail(db, email)
	if sErr != nil {
		return nil, sErr
	}
	if user == nil {
		return nil, response.ErrorCode_InvalidParam.New("email not registered")
	}
	if user.Password == nil {
		return nil, response.ErrorCode_InvalidParam.New("user has not set password yet")
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password.Data), []byte(password.Data))
	if err != nil {
		return nil, response.ErrorCode_InvalidParam.New("wrong password")
	}
	return user, nil
}

func (c *UserCtrl) GetUserByUID(uid dal.UID) (*dal.User, response.SError) {
	db := service.GetDBExecutor()
	user, sErr := dal.UserDBHD.GetByUID(db, uid)
	if sErr != nil {
		return nil, sErr
	}
	if user == nil {
		return nil, response.ErrorCode_InvalidParam.New("invalid uid, not found")
	}
	return user, nil
}

const PortraitSizeLimit = 10 * 1024 * 1024 // 10M

type UpdateUserBaseInfoFields struct {
	Name     string
	Portrait []byte
}

func (c *UserCtrl) UpdateUserBaseInfo(uid dal.UID, updateFields *UpdateUserBaseInfoFields) (*dal.User, response.SError) {
	ctx := context.Background()
	db := service.GetDBExecutor()
	user, sErr := dal.UserDBHD.GetByUID(db, uid)
	if sErr != nil {
		return nil, sErr
	}
	if user == nil {
		return nil, response.ErrorCode_InvalidParam.New("invalid uid, no user found")
	}

	updates := &dal.UserUpdatableFields{}
	if updateFields.Name != "" {
		updates.Name = updateFields.Name
		user.Name = &updateFields.Name
	}

	if len(updateFields.Portrait) > 0 {
		if len(updates.Portrait) > PortraitSizeLimit {
			return nil, response.ErrorCode_InvalidParam.New("file size beyond limit")
		}

		imageFormat := util.ParseRawImageFormat(updateFields.Portrait)
		if imageFormat == util.ImgFormatUnknown {
			return nil, response.ErrorCode_InvalidParam.New("unsupported image format type")
		}
		updates.Portrait = fmt.Sprintf("portrait/%s.%s", uid, imageFormat)
		user.Portrait = &updates.Portrait
		user.PortraitURL, _ = service.GetObjectStorageExecutor().ObjectKeyToURL(updates.Portrait, time.Hour)
	}

	sErr = WithDBTx(db, func(tx *gorm.DB) response.SError {
		sErr = dal.UserDBHD.UpdateUser(tx, uid, updates)
		if sErr != nil {
			return sErr
		}
		if len(updateFields.Portrait) > 0 {
			osExecutor := service.GetObjectStorageExecutor()
			err := osExecutor.PutObject(ctx, updates.Portrait, bytes.NewReader(updateFields.Portrait))
			if err != nil {
				return response.ErrroCode_InternalUnknownError.Wrap(err, "put portrait data fail")
			}
		}
		return nil
	})
	if sErr != nil {
		return nil, sErr
	}

	return user, nil
}

type SimplifiedUser struct {
	UID      dal.UID `json:"uid"`
	Name     *string `json:"name"`
	Portrait string  `json:"portrait"`
}

func (c *UserCtrl) SearchUserByNameOrUID(text string) ([]*SimplifiedUser, response.SError) {
	db := service.GetDBExecutor()
	users, sErr := dal.UserDBHD.SearchUserByNameOrUID(db, text, &dal.Pagination{
		Page:     1,
		PageSize: 10, // default only show ten people
	})
	if sErr != nil {
		return nil, sErr
	}

	simplifiedUsers := make([]*SimplifiedUser, 0, len(users))
	for _, user := range users {
		simplifiedUsers = append(simplifiedUsers, &SimplifiedUser{
			UID:      user.UID,
			Name:     user.Name,
			Portrait: user.PortraitURL,
		})
	}

	return simplifiedUsers, nil
}

func (c *UserCtrl) SendResetPasswordEmail(email string) response.SError {
	db := service.GetDBExecutor()

	user, sErr := dal.UserDBHD.GetByEmail(db, email)
	if sErr != nil {
		return sErr
	}
	if user == nil {
		return response.ErrorCode_InvalidParam.New("email not registered")
	}

	previousVerifyCodeRecord, sErr := dal.UserEmailVerifyCodeDBHD.GetByEmail(db, email)
	if sErr != nil {
		return sErr
	}
	now := time.Now().UTC()
	if previousVerifyCodeRecord != nil {
		if now.Sub(previousVerifyCodeRecord.SendAt) <= time.Minute {
			return response.ErrorCode_InvalidParam.New("operation too frequently, try later")
		}
	}

	verifyCode := util.GenerateVerifyCode(6)

	sErr = WithDBTx(db, func(tx *gorm.DB) response.SError {
		if previousVerifyCodeRecord != nil {
			sErr = dal.UserEmailVerifyCodeDBHD.DeleteByEmail(tx, email)
			if sErr != nil {
				return sErr
			}
		}

		sErr = dal.UserEmailVerifyCodeDBHD.Add(tx, &dal.UserEmailVerifyCode{
			Email:      email,
			VerifyCode: verifyCode,
			SendAt:     now,
		})
		if sErr != nil {
			return sErr
		}

		sErr = c.sendEmailResetPasswordEmail(email, verifyCode)
		if sErr != nil {
			return sErr
		}
		return nil
	})
	if sErr != nil {
		return sErr
	}
	return nil
}

func (c *UserCtrl) ResetPassword(email string, verifyCode string, newPassword string) response.SError {
	db := service.GetDBExecutor()
	user, sErr := dal.UserDBHD.GetByEmail(db, email)
	if sErr != nil {
		return sErr
	}
	if user == nil {
		return response.ErrorCode_InvalidParam.New("email not registered")
	}

	verifyCodeRecord, sErr := dal.UserEmailVerifyCodeDBHD.GetByEmail(db, email)
	if sErr != nil {
		return sErr
	}
	if verifyCodeRecord == nil {
		return response.ErrorCode_InvalidParam.New("this email currently has no available verify code")
	}

	if time.Now().UTC().Sub(verifyCodeRecord.SendAt) > time.Minute*10 {
		return response.ErrorCode_InvalidParam.New("verify code expired")
	}

	if verifyCodeRecord.VerifyCode != verifyCode {
		return response.ErrorCode_InvalidParam.New("verify code not right")
	}

	sErr = WithDBTx(db, func(tx *gorm.DB) response.SError {
		sErr = dal.UserDBHD.UpdateUser(tx, user.UID, &dal.UserUpdatableFields{
			Password: &dal.Password{
				Data:   newPassword,
				Hashed: false,
			},
		})
		if sErr != nil {
			return sErr
		}

		return dal.UserEmailVerifyCodeDBHD.DeleteByEmail(tx, email)
	})

	if sErr != nil {
		return sErr
	}

	return nil
}

func (c *UserCtrl) DeleteAccount(uid dal.UID) response.SError {
	db := service.GetDBExecutor()
	habits, _, sErr := dal.HabitDBHD.ListUserJoinedHabits(db, uid, nil)
	if sErr != nil {
		return sErr
	}

	sErr = WithDBTx(db, func(tx *gorm.DB) response.SError {
		var successor *dal.HabitGroup
		for _, habit := range habits {
			if habit.Owner == uid { // is habit owner
				// try to find a successor
				successor, sErr = dal.HabitGroupDBHD.GetByHabitIDAndExcludeUID(db, habit.ID, uid)
				if sErr != nil {
					return sErr
				}
				if successor == nil { // no successor means current use is the last one participate in this habit
					sErr = dal.HabitDBHD.DeleteByID(db, habit.ID)
				} else {
					sErr = dal.HabitDBHD.UpdateHabit(tx, habit.ID, &dal.HabitUpdatableFields{Owner: successor.UID})
				}
				if sErr != nil {
					return sErr
				}
			}
		}

		sErr = dal.HabitGroupDBHD.DeleteByUID(tx, uid)
		if sErr != nil {
			return sErr
		}

		sErr = dal.UserHabitConfigDBHD.DeleteByUID(tx, uid)
		if sErr != nil {
			return sErr
		}

		sErr = dal.HabitLogRecordDBHD.DeleteByUID(tx, uid)
		if sErr != nil {
			return sErr
		}

		sErr = dal.UnconfirmedHabitLogRecordDBHD.DeleteByUID(tx, uid)
		if sErr != nil {
			return sErr
		}

		sErr = dal.UserDBHD.DeleteUserByID(tx, uid)
		if sErr != nil {
			return sErr
		}
		return nil
	})

	if sErr != nil {
		return sErr
	}
	return nil
}
