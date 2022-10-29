package controller

import (
	"bytes"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/swordandtea/fhwh/biz/config"
	"github.com/swordandtea/fhwh/biz/dal"
	"github.com/swordandtea/fhwh/biz/response"
	"github.com/swordandtea/fhwh/biz/service"
	"github.com/swordandtea/fhwh/nullable"
	"github.com/swordandtea/fhwh/util"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"sync"
	"text/template"
	"time"
)

type UserCtrl struct{}

var onceEmailActivate = &sync.Once{}
var onceEmailBind = &sync.Once{}
var emailActivateTmpl *template.Template
var emailBindTmpl *template.Template

// emailActivateTmplStr the email template used for user registering from email
const emailActivateTmplStr = `From: {{.From}}
To: {{.To}}
Subject: [lets-habits] 邮箱激活 (mail activate)
Content-Type: text/plain; charset=utf-8

欢迎加入lets-habits，点击下方链接以激活账户:
{{.ActiveLink}}

welcome To join lets-habits, click the link below to activate your account:
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

you are current binding email for account, if it's your operation, click the link below to bind:
{{.BindLink}}
otherwise, please ignore this message
`

// emailActivateAllowedInterval the max time interval that allow a user to resend account activate email
const emailActivateAllowedInterval = time.Minute

// emailCodeExpireTime the token expired time for email activate and bind
const emailCodeExpireTime = time.Minute * 10

// userTokenExpireTime user auth token expire time: 3 day
const userTokenExpireTime = time.Hour * 72

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

// sendActivateEmail send email activate email to targe email address
func (c *UserCtrl) sendActivateEmail(toMail string, uid dal.UID) response.SError {
	mailExecutor := service.GetMailExecutor()
	// prepare email message
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

// EmailRegister do email register, will send an email activate email to user
func (c *UserCtrl) EmailRegister(email string, password *dal.Password) (dal.UID, response.SError) {
	db := service.GetDBExecutor()
	user, sErr := dal.UserDBHD.GetByEmail(db, email)
	if sErr != nil {
		return "", sErr
	}
	if user != nil {
		return "", response.ErrorCode_UserNoPermission.New("email already registered")
	}

	uea, sErr := dal.UserEmailActivateDBHD.GetByEmail(db, email)
	if sErr != nil {
		return "", sErr
	}
	now := time.Now().UTC()
	var uid dal.UID
	if uea != nil { // means the email already registered but not activated yet
		if uea.SendAt.Add(emailActivateAllowedInterval).After(now) {
			return "", response.ErrorCode_UserNoPermission.New("the email is send within one minutes, try later")
		}

		sErr = WithDBTx(func(tx *gorm.DB) response.SError {
			sErr = dal.UserEmailActivateDBHD.UpdateSendTime(tx, uea.ID, now)
			if sErr != nil {
				return sErr
			}
			sErr = c.sendActivateEmail(email, uea.UID)
			if sErr != nil {
				return sErr
			}
			return nil
		})
		if sErr != nil {
			return "", sErr
		}
		uid = uea.UID
	} else { // means the email has not been registered
		uid = dal.UID(uuid.New().String())
		uea = &dal.UserEmailActivate{
			UID:       uid,
			Email:     email,
			Password:  password,
			SendAt:    now,
			Activated: false,
		}

		sErr = WithDBTx(func(tx *gorm.DB) response.SError {
			sErr = dal.UserEmailActivateDBHD.Add(tx, uea)
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
			return "", sErr
		}
	}
	return uid, nil
}

// EmailActivate confirm email activate, create a new user if activate success
func (c *UserCtrl) EmailActivate(activateCode string) (*dal.User, string /*user token*/, response.SError) {
	// verify activate code
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(activateCode, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GlobalConfig.JWT.Cypher), nil
	})
	if err != nil {
		return nil, "", response.ErrorCode_UserNoPermission.Wrap(err, "invalid activate code")
	}
	db := service.GetDBExecutor()
	uea, sErr := dal.UserEmailActivateDBHD.GetByUID(db, dal.UID(claims.ID))
	if sErr != nil {
		return nil, "", sErr
	}
	if uea == nil {
		return nil, "", response.ErrorCode_UserNoPermission.Wrap(err, "invalid activate code, no user found")
	}
	if uea.Activated {
		return nil, "", response.ErrorCode_UserNoPermission.Wrap(err, "email already activated")
	}

	// do activate
	var tokenStr string
	var user *dal.User
	sErr = WithDBTx(func(tx *gorm.DB) response.SError {
		sErr = dal.UserEmailActivateDBHD.SetActivated(tx, uea.ID)
		if sErr != nil {
			return sErr
		}

		user = &dal.User{
			UID:              uea.UID,
			Name:             nullable.NullString{},
			Email:            nullable.MakeNullString(uea.Email),
			Password:         uea.Password,
			Portrait:         nullable.NullString{},
			UserRegisterType: dal.UserRegisterTypeEmail,
		}

		sErr = dal.UserDBHD.Add(tx, user)
		if sErr != nil {
			return sErr
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(userTokenExpireTime)),
			ID:        uuid.New().String(),
		})

		tokenStr, err = token.SignedString([]byte(config.GlobalConfig.JWT.Cypher))
		if err != nil {
			return response.ErrroCode_InternalUnknownError.Wrap(err, "sign user token fail")
		}
		return nil
	})
	if sErr != nil {
		return nil, "", sErr
	}
	return user, tokenStr, nil
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

	if user.Email.Get() == email {
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

const PortraitSizeLimit = 10 * 1024 * 1024 // 10M

type UpdateUserBaseInfoFields struct {
	Name     string
	Portrait *multipart.FileHeader
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
		user.Name = nullable.MakeNullString(updateFields.Name)
	}

	var portraitData []byte

	if updateFields.Portrait != nil {
		if updateFields.Portrait.Size > PortraitSizeLimit {
			return nil, response.ErrorCode_InvalidParam.New("file size beyond limit")
		}
		fReader, err := updateFields.Portrait.Open()
		if err != nil {
			return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "open portrait file fail")
		}
		portraitData, err = io.ReadAll(fReader)
		if err != nil {
			return nil, response.ErrroCode_InternalUnknownError.Wrap(err, "read portrait data fail")
		}
		imageFormat := util.ParseRawImageFormat(portraitData)
		if imageFormat == util.ImgFormatUnknown {
			return nil, response.ErrorCode_InvalidParam.Wrap(err, "unsupported image format type")
		}
		updates.Portrait = fmt.Sprintf("portrait/%s.%s", uid, imageFormat)
		user.Portrait = nullable.MakeNullString(updates.Portrait)
		user.PortraitURL = service.GetObjectStorageExecutor().ObjectKeyToURL(updates.Portrait)
	}

	sErr = WithDBTx(func(tx *gorm.DB) response.SError {
		sErr = dal.UserDBHD.UpdateUser(tx, uid, updates)
		if sErr != nil {
			return sErr
		}
		if len(portraitData) > 0 {
			osExecutor := service.GetObjectStorageExecutor()
			err := osExecutor.PutObject(ctx, updates.Portrait, portraitData)
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
