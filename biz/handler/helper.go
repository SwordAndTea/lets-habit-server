package handler

import (
	emailverify "github.com/AfterShip/email-verifier"
	"github.com/swordandtea/fhwh/biz/response"
)

func BindAndValidateErr(err error) response.SError {
	return response.ErrorCode_InvalidParam.Wrap(err, "bind req fail")
}

var verifier = emailverify.NewVerifier().EnableSMTPCheck().EnableAutoUpdateDisposable()

func ValidateEmail(e string) response.SError {
	ret, err := verifier.Verify(e)
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "verify email fail")
	}
	if !ret.Syntax.Valid {
		return response.ErrorCode_InvalidParam.New("invalid email syntax")
	}
	if !ret.Disposable {
		return response.ErrorCode_InvalidParam.New("a disposable email is not allowed")
	}
	if !ret.SMTP.Deliverable && !ret.SMTP.CatchAll {
		return response.ErrorCode_InvalidParam.New("can not send to this email address")
	}

	return nil
}

func ValidatePassword(p string) response.SError {
	if len(p) < 8 {
		return response.ErrorCode_InvalidParam.New("password length is less than eight")
	}
	return nil
}
