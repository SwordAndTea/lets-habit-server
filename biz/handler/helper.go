package handler

import (
	"encoding/base64"
	emailverify "github.com/AfterShip/email-verifier"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/golang-jwt/jwt/v4"
	"github.com/swordandtea/lets-habit-server/biz/config"
	"github.com/swordandtea/lets-habit-server/biz/dal"
	"github.com/swordandtea/lets-habit-server/biz/response"
	"time"
)

func BindAndValidateErr(err error) response.SError {
	return response.ErrorCode_InvalidParam.Wrap(err, "bind req fail")
}

var verifier = emailverify.NewVerifier().EnableSMTPCheck().EnableAutoUpdateDisposable()

// ValidateEmail validate whether an email is reachable
func ValidateEmail(e string) response.SError {
	ret, err := verifier.Verify(e)
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "verify email fail")
	}
	if !ret.Syntax.Valid {
		return response.ErrorCode_InvalidParam.New("invalid email syntax")
	}
	if !ret.SMTP.Deliverable && !ret.SMTP.CatchAll {
		return response.ErrorCode_InvalidParam.New("can not send to this email address")
	}

	return nil
}

// ValidatePassword validate whether a password is valid passport,
// currently the only limitation is that the length of a password should more than eight
// TODO: add character check
func ValidatePassword(p string) response.SError {
	if len(p) < 8 {
		return response.ErrorCode_InvalidParam.New("password length is less than eight")
	}
	return nil
}

const UserTokenExpireTime = time.Hour * 24 * 7 // one week

func GenerateUserToken(uid dal.UID) (string, response.SError) {
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(UserTokenExpireTime)),
		ID:        string(uid),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(config.GlobalConfig.JWT.Cypher))
	if err != nil {
		return "", response.ErrroCode_InternalUnknownError.Wrap(err, "generate user token fail")
	}
	return tokenStr, nil
}

func SetUserTokenCookie(rc *app.RequestContext, uid dal.UID) response.SError {
	userToken, sErr := GenerateUserToken(uid)
	if sErr != nil {
		return sErr
	}

	rc.SetCookie(
		UserTokenKey,
		userToken,
		int(UserTokenExpireTime/time.Second),
		"/",
		"",
		protocol.CookieSameSiteLaxMode,
		false,
		true,
	)
	return nil
}

func ExtractUserToken(token string) (dal.UID, response.SError) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GlobalConfig.JWT.Cypher), nil
	})

	if err != nil {
		return "", response.ErrorCode_UserAuthFail.Wrap(err, "verify user token fail")
	}

	if claims.ID == "" {
		return "", response.ErrorCode_UserAuthFail.Wrap(err, "invalid user token, no user id found")
	}
	return dal.UID(claims.ID), nil
}

func Base64ImgDecode(img string) ([]byte, response.SError) {
	if img == "" {
		return nil, nil
	}
	imageData, err := base64.StdEncoding.DecodeString(img)
	if err != nil {
		return nil, response.ErrorCode_InvalidParam.New("invalid image base64")
	}
	return imageData, nil
}
