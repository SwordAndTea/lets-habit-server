package controller

import (
	"bytes"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/swordandtea/fhwh/biz/config"
	"github.com/swordandtea/fhwh/biz/dal"
	"github.com/swordandtea/fhwh/biz/response"
	"github.com/swordandtea/fhwh/biz/service"
	"github.com/swordandtea/fhwh/nullable"
	"gorm.io/gorm"
	"sync"
	"text/template"
	"time"
)

type UserCtrl struct{}

var once = &sync.Once{}
var emailActivateTmpl *template.Template

const emailActivateTmplStr = `From: {{.From}}
To: {{.To}}
Subject: [lets-habits] 邮箱激活 (mail activate)
Content-Type: text/plain; charset=utf-8

欢迎加入lets-habits，点击下方链接以激活账户:
{{.ActiveLink}}

welcome To join lets-habits, click the link below To activate your account:
{{.ActiveLink}}
`

const emailActivateAllowedInterval = time.Minute
const emailActivateCodeExpireTime = time.Minute * 10
const userTokenExpireTime = time.Hour * 72

type emailActivateTmplFiller struct {
	From       string
	To         string
	ActiveLink string
}

func GetEmailTemplate() *template.Template {
	once.Do(func() {
		emailActivateTmpl, _ = template.New("mail-activate-tmpl").Parse(emailActivateTmplStr)
	})
	return emailActivateTmpl
}

func (c *UserCtrl) sendActivateEmail(toMail string, uid string) response.SError {
	mailExecutor := service.GetMailExecutor()
	// prepare email message
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(emailActivateCodeExpireTime)),
		ID:        uid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(config.GlobalConfig.JWT.Cypher))
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "sign activate code fail")
	}

	data := &bytes.Buffer{}
	err = GetEmailTemplate().Execute(data, &emailActivateTmplFiller{
		From: mailExecutor.Sender(),
		To:   toMail,
		ActiveLink: fmt.Sprintf("%s?%s=%s", config.GlobalConfig.EmailService.ActivateURI,
			config.GlobalConfig.EmailService.ActivateParam, tokenStr),
	})
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "fill email activate template fail")
	}

	// send email
	err = mailExecutor.SendMail([]string{string(toMail)}, data.Bytes())
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "send email fail")
	}
	return nil
}

func (c *UserCtrl) EmailRegister(email string, password *dal.Password) response.SError {
	db := service.GetDBExecutor()
	user, sErr := dal.UserDBHD.GetByEmail(db, email)
	if sErr != nil {
		return sErr
	}
	if user != nil {
		return response.ErrorCode_UserNoPermission.New("email already registered")
	}

	uea, sErr := dal.UserEmailActivateDBHD.GetByEmail(db, email)
	if sErr != nil {
		return sErr
	}
	now := time.Now().UTC()
	if uea != nil { // means the email already registered but not activated yet
		if uea.SendAt.Add(emailActivateAllowedInterval).After(now) {
			return response.ErrorCode_UserNoPermission.New("the email is send within one minutes, try later")
		}

		sErr = WithDBTx(func(tx *gorm.DB) response.SError {
			sErr = dal.UserEmailActivateDBHD.UpdateSendTime(tx, uea.ID, now)
			if sErr != nil {
				return sErr
			}
			sErr = c.sendActivateEmail(email, string(uea.UID))
			if sErr != nil {
				return sErr
			}
			return nil
		})
		if sErr != nil {
			return sErr
		}
	} else { // means the email has not been registered
		uid := uuid.New().String()
		uea = &dal.UserEmailActivate{
			UID:       dal.UID(uid),
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
			return sErr
		}
	}
	return nil
}

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
